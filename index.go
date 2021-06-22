package main
import (
 	"github.com/gin-gonic/gin"
 	"database/sql"
 	"gopkg.in/gorp.v1"
 	_ "github.com/lib/pq"
 	"strconv"
 	"log"
)

var dbmap = initDb()

func initDb() *gorp.DbMap {
 	db, err := sql.Open("postgres", "postgres://mafdginl:lm-XYdCZVRadbHZZY77Qw-uuQ-3AAm8N@batyr.db.elephantsql.com/mafdginl")
 	checkErr(err, "sql.Open failed")
 	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.AddTableWithName(Account{}, "Account").SetKeys(true, "account_number")
	dbmap.AddTableWithName(Customer{}, "Customer").SetKeys(true, "customer_number")
	return dbmap
}

func checkErr(err error, msg string) {
 	if err != nil {
 		log.Fatalln(msg, err)
 	}
}

func AccountInquiry(c *gin.Context) {
	var AccountDB Account
	var CustomerDB Customer
	paramAccNumber := c.Params.ByName("accountNumber")
 	accountNumber, _ := strconv.Atoi(paramAccNumber)
	accountError := dbmap.SelectOne(&AccountDB, "select * from account where account_number = $1", accountNumber)
    checkErr(accountError, "SelectOne data failed for account")
    log.Println("Account row:", AccountDB)

	customerError := dbmap.SelectOne(&CustomerDB, "select * from customer where customer_number = $1", AccountDB.CustomerNumber)
    checkErr(customerError, "SelectOne data failed for customer")
    log.Println("Customer row:", CustomerDB)

	if accountError == nil && customerError == nil {
		content := &AccountInquiryResponse{
 			CustomerName: CustomerDB.CustomerName,
 			AccountNumber: AccountDB.AccountNumber,
			Balance: AccountDB.Balance,
		}
 		c.JSON(200, content)
 	} else {
 		c.JSON(404, gin.H{"error": "Account Not Found"})
 	}
}

func TransferBalance(c *gin.Context) {
	var FromAccountDB Account
	var ToAccountDB Account
	var Req TransferRequest
	paramFromAccountNumber := c.Params.ByName("fromAccountNumber")
	fromAccountNumber, _ := strconv.Atoi(paramFromAccountNumber)
	c.Bind(&Req)
	log.Println(Req.ToAccountNumber == 0)

	if Req.ToAccountNumber <= 0 {
		c.JSON(400, gin.H{"error": "Invalid Source Account"})
		return
	}

	if Req.Amount <= 0 {
		c.JSON(400, gin.H{"error": "Invalid Amount"})
		return
	}

	fromAccountError := dbmap.SelectOne(&FromAccountDB, "select * from account where account_number = $1", fromAccountNumber)
    checkErr(fromAccountError, "SelectOne from account number failed")
    log.Println("From Account row:", FromAccountDB)

	if fromAccountError != nil {
		c.JSON(404, gin.H{"error": "Source Account Number Not Found"})
		return
	}

	toAccountError := dbmap.SelectOne(&ToAccountDB, "select * from account where account_number = $1", Req.ToAccountNumber)
    checkErr(toAccountError, "SelectOne benef account number failed")
    log.Println("Benef Account row:", ToAccountDB)

	if toAccountError != nil {
		c.JSON(404, gin.H{"error": "Beneficiary Account Number Not Found"})
		return
	}

	if Req.Amount > FromAccountDB.Balance {
		c.JSON(400, gin.H{"error": "Insufficient Balance"})
		return
	} else {
		FinalFromAcc := Account{
			AccountNumber: FromAccountDB.AccountNumber,
			CustomerNumber: FromAccountDB.CustomerNumber,
			Balance: FromAccountDB.Balance - Req.Amount,
		}

		FinalToAcc := Account{
			AccountNumber: ToAccountDB.AccountNumber,
			CustomerNumber: ToAccountDB.CustomerNumber,
			Balance: ToAccountDB.Balance + Req.Amount,
		}

		_, fromAccountError = dbmap.Update(&FinalFromAcc)
		if fromAccountError != nil {
			log.Println("Update Source Account Error Reason: ", fromAccountError)			
			c.JSON(400, gin.H{"error": "Debit Source Account Process Failed"})
			return
		} else {
			_, toAccountError = dbmap.Update(&FinalToAcc)
	
			if toAccountError == nil {
				c.Writer.WriteHeader(201)
			} else {
				log.Println("Update Beneficiary Account Error Reason: ", toAccountError)
				c.JSON(400, gin.H{"error": "Credit Source Account Process Failed"})
				return
			}
		}
	}
}


func index (c *gin.Context) {
    content := gin.H{"Hello": "Dito"}
    c.JSON(200, content)
}

type Customer struct {
 	CustomerNumber int64 `db:"customer_number" json:"customer_number"`
 	CustomerName string `db:"customer_name" json:"customer_name"`
}

type Account struct {
	AccountNumber int64 `db:"account_number" json:"account_number"`
	CustomerNumber int64 `db:"customer_number" json:"customer_number"`
	Balance int64 `db:"balance" json:"balance"`
}

type AccountInquiryResponse struct {
	AccountNumber int64 `json:"account_number"`
	CustomerName string `json:"customer_name"`
	Balance int64 `json:"balance"`
}

type TransferRequest struct {
	ToAccountNumber int64 `json:"to_account_number"`
	Amount int64 `json:"amount"`
}

func main() {
 	r := gin.Default()
	v1 := r.Group("account")
 	{
 		v1.GET("/:accountNumber", AccountInquiry)
 		v1.POST("/:fromAccountNumber/transfer", TransferBalance)
 	}
 	r.GET("/", index)
	r.Run(":3000")
}

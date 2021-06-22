FROM golang:alpine

ENV GIN_MODE=release
ENV PORT=3000

RUN apk add --no-cache git

RUN mkdir -p /services/demo-linkaja

WORKDIR /services/demo-linkaja/

RUN git clone https://github.com/gregoriusandito/demo-golang.git

WORKDIR /services/demo-linkaja/demo-golang

RUN go build

EXPOSE $PORT

ENTRYPOINT ["./demo-linkaja"]
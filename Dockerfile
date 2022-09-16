FROM golang:latest

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY .env ./

RUN go mod download

COPY ./ ./

RUN go build -o /main

CMD ["/main"]

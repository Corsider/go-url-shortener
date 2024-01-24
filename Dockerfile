FROM golang:latest

WORKDIR /app

COPY .env .
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
RUN go build cmd/api/main.go
CMD ["./main"]
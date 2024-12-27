## Build
FROM golang:1.23.2-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/main.go

ENTRYPOINT ["./main"]
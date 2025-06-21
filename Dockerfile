FROM golang:1.24.2 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o edu-bot ./cmd/main.go
EXPOSE 8080
CMD ["./edu-bot"]

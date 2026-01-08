FROM golang:1.25 AS builder

WORKDIR /app/microservices/messaging-service

COPY microservices/messaging-service/go.mod microservices/messaging-service/go.sum ./
COPY shared/ /app/shared/

RUN go mod download

COPY . /app

RUN go build -o /migrate-tool ./cmd/migrate/main.go

FROM debian:bookworm-slim
WORKDIR /root/
COPY --from=builder /migrate-tool .

CMD ["./migrate-tool"]

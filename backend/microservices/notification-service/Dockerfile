# Builder
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY shared/ ./shared/
COPY microservices/notification-service/ ./microservices/notification-service/
WORKDIR /app/microservices/notification-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

# Runtime
FROM gcr.io/distroless/static-debian12 AS production
WORKDIR /
COPY --from=builder /server /server
EXPOSE 50051
ENTRYPOINT ["/server"]

# Dev
FROM golang:1.25-alpine AS dev
WORKDIR /app
COPY shared/pkg/go.mod shared/pkg/go.sum ./shared/pkg/
COPY shared/proto/go.mod shared/proto/go.sum ./shared/proto/
COPY microservices/notification-service/go.mod microservices/notification-service/go.sum ./microservices/notification-service/
WORKDIR /app/microservices/notification-service
RUN go mod download
WORKDIR /app
COPY shared/ ./shared/
COPY microservices/notification-service/ ./microservices/notification-service/
WORKDIR /app/microservices/notification-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .
ENTRYPOINT ["/server"]

# Builder
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY shared/ ./shared/
COPY microservices/post-service/ ./microservices/post-service/
WORKDIR /app/microservices/post-service
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
COPY microservices/post-service/go.mod microservices/post-service/go.sum ./microservices/post-service/
WORKDIR /app/microservices/post-service
RUN go mod download
WORKDIR /app
COPY shared/ ./shared/
COPY microservices/post-service/ ./microservices/post-service/
WORKDIR /app/microservices/post-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .
ENTRYPOINT ["/server"]

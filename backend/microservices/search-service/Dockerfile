# Builder
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY shared/go.mod shared/go.sum ./shared/
COPY microservices/search-service/go.mod microservices/search-service/go.sum ./microservices/search-service/
WORKDIR /app/microservices/search-service
RUN go mod download
WORKDIR /app
COPY shared/ ./shared/
COPY microservices/search-service/ ./microservices/search-service/
WORKDIR /app/microservices/search-service
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
COPY microservices/search-service/go.mod microservices/search-service/go.sum ./microservices/search-service/
WORKDIR /app/microservices/search-service
RUN go mod download
WORKDIR /app
COPY shared/ ./shared/
COPY microservices/search-service/ ./microservices/search-service/
WORKDIR /app/microservices/search-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .
ENTRYPOINT ["/server"]

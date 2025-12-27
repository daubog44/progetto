# Builder
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY shared/ ./shared/
COPY microservices/gateway-service/ ./microservices/gateway-service/
WORKDIR /app/microservices/gateway-service
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
COPY shared/ ./shared/
COPY microservices/gateway-service/ ./microservices/gateway-service/
WORKDIR /app/microservices/gateway-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .
ENTRYPOINT ["/server"]

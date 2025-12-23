# Builder
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY shared/ /shared/
COPY microservices/cassandra-service/go.mod microservices/cassandra-service/go.sum ./
RUN go mod download
COPY microservices/cassandra-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

# Runtime
FROM gcr.io/distroless/static-debian12 AS production
WORKDIR /
COPY --from=builder /server /server
EXPOSE 8080
ENTRYPOINT ["/server"]

# Dev
FROM golang:1.25-alpine AS dev
WORKDIR /app
COPY shared/ /shared/
COPY microservices/cassandra-service/go.mod microservices/cassandra-service/go.sum ./
RUN go mod download
COPY microservices/cassandra-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .
ENTRYPOINT ["/server"]

# Builder
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY shared/ /shared/
COPY microservices/neo4j-service/go.mod microservices/neo4j-service/go.sum ./
RUN go mod download
COPY microservices/neo4j-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server main.go

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
COPY microservices/neo4j-service/go.mod microservices/neo4j-service/go.sum ./
RUN go mod download
COPY microservices/neo4j-service/ .
RUN go build -o /server main.go
ENTRYPOINT ["/server"]

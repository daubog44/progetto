#!/bin/bash

# scripts/add-service.sh

if [ -z "$1" ]; then
    echo "Usage: ./scripts/add-service.sh <service-name>"
    exit 1
fi

SERVICE_NAME=$1
SERVICE_PATH="microservices/$SERVICE_NAME"

if [ -d "$SERVICE_PATH" ]; then
    echo "Error: Service $SERVICE_NAME already exists at $SERVICE_PATH"
    exit 1
fi

echo "🚀 Creating microservice: $SERVICE_NAME..."

# 1. Create Directory
mkdir -p "$SERVICE_PATH"

# 2. Setup boilerplate and dependencies
cd "$SERVICE_PATH" || exit

# 2.1 Initialize Go module
go mod init "github.com/username/progetto/$SERVICE_NAME"
go mod edit -go=1.25
go mod edit -toolchain=go1.25.0
go mod edit -replace github.com/username/progetto/proto=../../shared/proto

# 2.2 Create Main.go Boilerplate FIRST
# This is crucial so go mod tidy can see the imports
cat <<EOF > "main.go"
package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
    
	datav1 "github.com/username/progetto/proto/gen/go/data/v1"
)

type server struct {
	datav1.UnimplementedDataServiceServer
}

func (s *server) GetData(ctx context.Context, req *datav1.GetDataRequest) (*datav1.GetDataResponse, error) {
	slog.Info("GetData called", "id", req.GetId())
	return &datav1.GetDataResponse{
		Data: "Response from $SERVICE_NAME for ID: " + req.GetId(),
	}, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting $SERVICE_NAME...")

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	datav1.RegisterDataServiceServer(s, &server{})
	reflection.Register(s)

	slog.Info("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		slog.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
EOF

# 2.3 Resolve dependencies
# Now go mod tidy will find the imports in main.go
go get google.golang.org/grpc@v1.77.0
go get google.golang.org/protobuf@v1.36.11
go get github.com/username/progetto/proto
go mod tidy

cd - > /dev/null

# 3. Create Dockerfile
cat <<EOF > "$SERVICE_PATH/Dockerfile"
# Builder
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY shared/ /shared/
COPY microservices/$SERVICE_NAME/go.mod microservices/$SERVICE_NAME/go.sum ./
RUN go mod download
COPY microservices/$SERVICE_NAME/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server main.go

# Runtime
FROM gcr.io/distroless/static-debian12 AS production
WORKDIR /
COPY --from=builder /server /server
EXPOSE 50051
ENTRYPOINT ["/server"]

# Dev
FROM golang:1.25-alpine AS dev
WORKDIR /app
COPY shared/ /shared/
COPY microservices/$SERVICE_NAME/go.mod microservices/$SERVICE_NAME/go.sum ./
RUN go mod download
COPY microservices/$SERVICE_NAME/ .
RUN go build -o /server main.go
ENTRYPOINT ["/server"]
EOF

# 5. Update go.work
if ! grep -q "./microservices/$SERVICE_NAME" go.work; then
    sed -i "/use (/a \	./microservices/$SERVICE_NAME" go.work
    echo "✅ Updated go.work"
fi

# 6. Update docker-compose.yml
if ! grep -q "$SERVICE_NAME:" docker-compose.yml; then
    # Insert before 'volumes:' to stay inside the 'services:' block
    sed -i "/^volumes:/i \  $SERVICE_NAME:\n    image: $SERVICE_NAME\n    build:\n      context: .\n      dockerfile: microservices/$SERVICE_NAME/Dockerfile\n      target: dev\n    container_name: $SERVICE_NAME\n    networks:\n      - microservices-net\n" docker-compose.yml
    echo "✅ Updated docker-compose.yml"
fi

# 7. Update Tiltfile
if ! grep -q "$SERVICE_NAME" Tiltfile; then
    cat <<EOF >> Tiltfile

# $SERVICE_NAME
docker_build(
    '$SERVICE_NAME',
    '.',
    dockerfile='microservices/$SERVICE_NAME/Dockerfile',
    live_update=[
        sync('./microservices/$SERVICE_NAME', '/app'),
        sync('./shared', '/shared'),
        run('go build -o /server main.go'),
        restart_container()
    ]
)
EOF
    echo "✅ Updated Tiltfile"
fi

echo "🎉 Service $SERVICE_NAME created successfully!"

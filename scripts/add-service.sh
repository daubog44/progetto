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
go mod edit -replace github.com/username/progetto/shared/pkg=../../shared/pkg

# 2.2 Create Main.go Boilerplate FIRST
# This is crucial so go mod tidy can see the imports
cat <<EOF > "main.go"
package main

import (
	"context"
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
    
	datav1 "github.com/username/progetto/proto/gen/go/data/v1"
	"github.com/username/progetto/shared/pkg/observability"
	"github.com/username/progetto/shared/pkg/watermillutil"
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
    // Initialize Observability
	cfg := observability.LoadConfigFromEnv()
	shutdown, err := observability.Init(context.Background(), cfg)
	if err != nil {
		slog.Error("failed to init observability", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown observability", "error", err)
		}
	}()

	slog.Info("Starting $SERVICE_NAME...", "config", cfg)

	// Metrics Server
	metricsPort := os.Getenv("PROMETHEUS_METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9091"
	}
	go func() {
		metricsHandler := watermillutil.GetMetricsHandler()
		slog.Info("Starting metrics server", "port", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, metricsHandler); err != nil {
			slog.Error("metrics server failed", "error", err)
		}
	}()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer(observability.GRPCServerOptions()...)
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
go get github.com/username/progetto/shared/pkg@v0.0.0-00010101000000-000000000000
go mod tidy

cd - > /dev/null

# 3. Create Dockerfile
cat <<EOF > "$SERVICE_PATH/Dockerfile"
# Builder
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY shared/ ./shared/
COPY microservices/$SERVICE_NAME/ ./microservices/$SERVICE_NAME/
WORKDIR /app/microservices/$SERVICE_NAME
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
COPY microservices/$SERVICE_NAME/ ./microservices/$SERVICE_NAME/
WORKDIR /app/microservices/$SERVICE_NAME
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .
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
    sed -i "/^volumes:/i \  $SERVICE_NAME:\n    image: $SERVICE_NAME\n    build:\n      context: .\n      dockerfile: microservices/$SERVICE_NAME/Dockerfile\n      target: dev\n    container_name: $SERVICE_NAME\n    environment:\n      - OTEL_SERVICE_NAME=$SERVICE_NAME\n      - OTEL_EXPORTER_OTLP_ENDPOINT=http://alloy:4317\n      - PROMETHEUS_METRICS_PORT=9091\n    networks:\n      - microservices-net\n" docker-compose.yml
    echo "✅ Updated docker-compose.yml"
fi

# 8. Update alloy.config
if ! grep -q "$SERVICE_NAME:9091" deploy/config/alloy.config; then
    # We want to insert into prometheus.scrape "watermill" { targets = [ ... ] }
    # We'll look for the "watermill" scrape block, then find the targets list, and insert before the closing bracket of the targets list
    # This is a bit fragile with sed, but let's try to match the specific structure we created.
    # Pattern: Look for 'prometheus.scrape "watermill"', then consume until 'targets = [', then look for ']' to insert before.
    # Simpler: Just rely on the last known entry or insert at top of list.
    
    # Let's insert before the closing bracket of the targets list inside watermill block.
    # We assume standard formatting.
    # using perl for multi-line match might be easier, or sed with context addresses.
    
    # We will search for 'prometheus.scrape "watermill"' and then the first ']' after that.
    # Note: This simple sed might break if there are other brackets.
    # Safe approach: Search for `prometheus.scrape "watermill"`, then subsequent lines until `]`. 
    # Actually, we can just append to the list if we can find the line `    { "__address__" = "messaging-service:9091", "service_name" = "messaging-service" },` and append after it.
    # But that line changes.
    
    # Robust-ish approach: Find `prometheus.scrape "watermill"`, then find the line containing `targets = [`, then find the matching `]`.
    # Let's try inserting after `targets = [` line.
    
    sed -i '/prometheus.scrape "watermill"/,/targets = \[/ {
        /targets = \[/a \    { "__address__" = "'"$SERVICE_NAME"':9091", "service_name" = "'"$SERVICE_NAME"'" },
    }' deploy/config/alloy.config
    
    echo "✅ Updated alloy.config"
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
        sync('./microservices/$SERVICE_NAME', '/app/microservices/$SERVICE_NAME'),
        sync('./shared', '/app/shared'),
        run('go build -o /server .'),
        restart_container()
    ]
)
EOF
    echo "✅ Updated Tiltfile"
fi

echo "🎉 Service $SERVICE_NAME created successfully!"

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
		Data: "Response from cultural-service for ID: " + req.GetId(),
	}, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting cultural-service...")

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

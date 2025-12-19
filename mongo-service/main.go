package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	datav1 "github.com/username/progetto/proto/gen/go/data/v1"
)

type server struct {
	datav1.UnimplementedDataServiceServer
	db *mongo.Client
}

func (s *server) GetData(ctx context.Context, req *datav1.GetDataRequest) (*datav1.GetDataResponse, error) {
	return &datav1.GetDataResponse{
		Data: "Sample data for ID: " + req.GetId(),
	}, nil
}

func (s *server) CreateUser(ctx context.Context, req *datav1.CreateUserRequest) (*datav1.CreateUserResponse, error) {
	collection := s.db.Database("progetto").Collection("users")

	user := bson.M{
		"firebase_uid": req.GetFirebaseUid(),
		"email":        req.GetEmail(),
		"display_name": req.GetDisplayName(),
		"created_at":   time.Now(),
	}

	res, err := collection.InsertOne(ctx, user)
	if err != nil {
		slog.Error("failed to insert user", "error", err)
		return nil, err
	}

	return &datav1.CreateUserResponse{
		Id:      res.InsertedID.(string), // Note: Mongo ObjectID by default, might need conversion
		Success: true,
	}, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mongoURI := os.Getenv("APP_MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}
	logger.Info("starting mongo-service", "uri", mongoURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		logger.Error("failed to connect to mongodb", "error", err)
		os.Exit(1)
	}
	// We don't disconnect here because the server will need it

	// Ping the primary
	if err := client.Ping(ctx, nil); err != nil {
		logger.Error("failed to ping mongodb", "error", err)
		os.Exit(1)
	}
	logger.Info("connected to MongoDB successfully")

	// gRPC Server setup
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	datav1.RegisterDataServiceServer(s, &server{db: client})
	reflection.Register(s) // For debugging with grpcurl

	logger.Info("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		logger.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}

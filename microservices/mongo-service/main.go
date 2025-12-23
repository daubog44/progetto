package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/username/progetto/mongo-service/internal/handler"
	"github.com/username/progetto/mongo-service/internal/repository"
	postv1 "github.com/username/progetto/proto/gen/go/post/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mongoURI := os.Getenv("APP_MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		logger.Error("failed to connect to mongodb", "error", err)
		os.Exit(1)
	}

	if err := client.Ping(ctx, nil); err != nil {
		logger.Error("failed to ping mongodb", "error", err)
		os.Exit(1)
	}

	db := client.Database("progetto")
	collection := db.Collection("posts")

	// Ensure Indexes
	_, _ = collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "author_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
	})

	repo := repository.NewMongoRepository(collection)
	postHandler := handler.NewPostHandler(repo)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()
	postv1.RegisterPostServiceServer(srv, postHandler)
	reflection.Register(srv)

	logger.Info("Post Service gRPC server listening on :50051")
	if err := srv.Serve(lis); err != nil {
		logger.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}

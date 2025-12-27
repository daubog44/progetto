package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"

	// "github.com/ThreeDotsLabs/watermill-opentelemetry/pkg/opentelemetry" REMOVED
	"github.com/username/progetto/post-service/internal/handler"
	"github.com/username/progetto/post-service/internal/repository"
	"github.com/username/progetto/post-service/internal/worker"
	postv1 "github.com/username/progetto/proto/gen/go/post/v1"
	"github.com/username/progetto/shared/pkg/observability"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	// Init Observability
	obsCfg := observability.LoadConfigFromEnv()
	shutdown, err := observability.Init(context.Background(), obsCfg)
	if err != nil {
		slog.Error("failed to init observability", "error", err)
	}
	defer func() {
		if shutdown != nil {
			shutdown(context.Background())
		}
	}()

	// Use context-aware logger initialized by observability
	logger = slog.Default()

	// Config
	mongoURI := os.Getenv("APP_MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}
	kafkaBrokers := os.Getenv("APP_KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:29092"
	}

	// 1. MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Mongo Options with Monitoring
	clientOpts := options.Client().ApplyURI(mongoURI)
	clientOpts.Monitor = otelmongo.NewMonitor()

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		logger.Error("failed to connect to mongodb", "error", err)
		os.Exit(1)
	}
	db := client.Database("progetto")

	// 2. Repositories
	postRepo := repository.NewMongoPostRepository(db)
	userRepo := repository.NewMongoUserRepository(db)

	// 3. Kafka Consumer (User Sync)
	kafkaSub, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       strings.Split(kafkaBrokers, ","),
			Unmarshaler:   kafka.DefaultMarshaler{},
			ConsumerGroup: "post_service_user_sync",
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		logger.Error("failed to create kafka subscriber", "error", err)
		os.Exit(1)
	}
	// OTel Subscriber Wrapper (Shared)
	subscriber := observability.NewTracingSubscriber(kafkaSub)

	userConsumer := worker.NewUserConsumer(userRepo)

	// Start consuming in background
	messages, err := subscriber.Subscribe(context.Background(), "user_created")
	if err != nil {
		logger.Error("failed to subscribe to topic", "error", err)
		os.Exit(1)
	}

	go func() {
		for msg := range messages {
			if err := userConsumer.Handle(msg); err != nil {
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()

	// 4. Publisher
	kafkaPub, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   strings.Split(kafkaBrokers, ","),
			Marshaler: kafka.DefaultMarshaler{},
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		logger.Error("failed to create kafka publisher", "error", err)
		os.Exit(1)
	}
	// OTel Publisher Wrapper (Shared)
	publisher := observability.NewTracingPublisher(kafkaPub)
	defer publisher.Close()

	// 5. gRPC Server
	postHandler := handler.NewPostHandler(postRepo, publisher)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer(observability.GRPCServerOptions()...)
	postv1.RegisterPostServiceServer(srv, postHandler)
	reflection.Register(srv)

	logger.Info("Post Service gRPC server listening on :50051")
	if err := srv.Serve(lis); err != nil {
		logger.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}

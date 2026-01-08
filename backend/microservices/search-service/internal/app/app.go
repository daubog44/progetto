package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	searchv1 "github.com/username/progetto/proto/gen/go/search/v1"
	"github.com/username/progetto/search-service/internal/api"
	"github.com/username/progetto/search-service/internal/config"
	"github.com/username/progetto/search-service/internal/events"
	"github.com/username/progetto/search-service/internal/handler"
	"github.com/username/progetto/search-service/internal/search"
	"github.com/username/progetto/shared/pkg/observability"
	"github.com/username/progetto/shared/pkg/watermillutil"
)

type App struct {
	Cfg              *config.Config
	Logger           *slog.Logger
	MeiliClient      *search.MeiliClient
	WatermillManager *events.WatermillManager
	GRPCServer       *grpc.Server
}

func New(cfg *config.Config) (*App, error) {
	logger := slog.Default()

	// Meilisearch Client
	meiliClient := search.NewMeiliClient(cfg.MeiliHost, cfg.MeiliKey)
	if err := meiliClient.EnsureIndex(context.Background()); err != nil {
		logger.Error("failed to ensure meilisearch index", "error", err)
		// Non-fatal for now, but worthy of a log
	}

	// 1. Create Publisher & Subscriber first (needed for Handler)
	publisher, err := watermillutil.NewKafkaPublisher(cfg.KafkaBrokers, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka publisher: %w", err)
	}

	// 2. Create Handler (needs Meili and Publisher)
	h := handler.NewNotificationHandler(meiliClient, publisher)

	// 3. Create Watermill Manager (creates Router, Subscriber, wires everything)
	// We pass the publisher we already created locally, or we let the manager create it?
	// The manager defined in previous step creates both. Let's adjust the Manager to be more flexible
	// OR, let's just make the App do the orchestration logic which is cleaner than a monolithic "Manager" constructor.

	// Let's rely on the events package helper we created, but we need to modify it to accept the publisher
	// since we need the publisher BEFORE the router to inject it into the handler.
	// Actually, looking at my previous write_to_file for watermill.go, it creates EVERYTHING.
	// This is a problem because Handler needs Publisher.

	// I will rewrite watermill.go logic here or fix watermill.go.
	// Let's assume I'll fix watermill.go to take the handler.
	// Wait, Handler needs Publisher. Manager creates Publisher.
	// Chicken and Egg.

	// Better Approach for App:
	// Init Publisher.
	// Init Handler.
	// Init Watermill Router (Configuring it).

	// Let's manually do it here or use a better factored events package.
	// I'll rewrite the events/watermill.go in the next step to be "Router definition" only.

	subscriber, err := watermillutil.NewKafkaSubscriber(cfg.KafkaBrokers, "search-service", logger)
	if err != nil {
		publisher.Close()
		return nil, fmt.Errorf("failed to create kafka subscriber: %w", err)
	}

	wmManager, err := events.NewRouter(logger, publisher, subscriber, h)
	if err != nil {
		publisher.Close()
		subscriber.Close()
		return nil, fmt.Errorf("failed to create watermill router: %w", err)
	}

	// gRPC Server
	grpcServer := grpc.NewServer(observability.GRPCServerOptions()...)
	searchv1.RegisterSearchServiceServer(grpcServer, api.NewServer(meiliClient))
	reflection.Register(grpcServer)

	return &App{
		Cfg:              cfg,
		Logger:           logger,
		MeiliClient:      meiliClient,
		WatermillManager: wmManager,
		GRPCServer:       grpcServer,
	}, nil
}

func (a *App) Run() error {
	// Start Watermill Router
	go func() {
		a.Logger.Info("Starting Watermill router")
		if err := a.WatermillManager.Router.Run(context.Background()); err != nil {
			a.Logger.Error("watermill router failed", "error", err)
		}
	}()

	// Start gRPC Server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen on :50051: %w", err)
	}

	a.Logger.Info("gRPC server listening on :50051")

	// Graceful Shutdown Channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		if err := a.GRPCServer.Serve(lis); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-stop:
		a.Logger.Info("Shutting down...")
		a.GRPCServer.GracefulStop()
		a.WatermillManager.Close()
	case err := <-errChan:
		return fmt.Errorf("grpc server error: %w", err)
	}

	return nil
}

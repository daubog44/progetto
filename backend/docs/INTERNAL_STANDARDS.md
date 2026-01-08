# Internal Microservice Standards

This document defines the standard architecture and implementation patterns for microservices in this project.

## Project Structure (Standard Layout)

Each microservice should strictly follow this directory structure:

- `main.go`: Entry point. Minimal logic. Handles config loading, dependency injection, and graceful shutdown.
- `internal/`
    - `config/`: Configuration logic (environment variables, etc.).
    - `handler/`: Business logic handlers. This captures what was previously named `worker`, `client`, or `service`.
    - `repository/`: Data access layer (Interfaces and Implementations).
    - `model/`: Domain models and DTOs.
    - `events/`: Event routing and publisher/subscriber logic.

## Component Design

### Handlers
Handlers should use a consistent struct injection pattern:

```go
type Handler struct {
    Repo      repository.SomeRepository
    Publisher message.Publisher
    Logger    *slog.Logger
}
```

### Dependency Injection
All dependencies (repositories, publishers, loggers) must be initialized in `main.go` and passed to handlers via constructors.

## Lifecycle & Graceful Shutdown

All services must use `signal.NotifyContext` to handle OS signals and ensure all background processes (HTTP servers, gRPC servers, Event Routers) shut down gracefully.

Example pattern in `main.go`:
```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

// Run servers/routers in goroutines or blocks that check ctx.Done()
go func() {
    if err := router.Run(ctx); err != nil {
        slog.Error("router failed", "error", err)
    }
}()

<-ctx.Done()
slog.Info("shutting down...")
// Close resources (DB connections, publishers, etc.)
```

## Configuration

Configuration must be managed in a dedicated `internal/config` package.
- Use a `Config` struct to hold all parameters.
- Implement a `Load()` function that reads from environment variables.
- **Strict Loading**: Fail fast (panic) if critical environment variables are missing.
- **Fail Early**: Environment variables should be validated during `Load()`, not later in the application lifecycle.

Example:
```go
func Load() *Config {
	return &Config{
		MongoURI:     mustGetEnv("APP_MONGO_URI"),
		KafkaBrokers: mustGetEnv("APP_KAFKA_BROKERS"),
	}
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("%s is required", key))
	}
	return v
}
```

## Consistency Rules
- **Configuration**: Always use `os.Getenv` or a dedicated `config` package. Fail fast (panic) for missing critical variables.
- **Logging**: Always use `log/slog`. Enrich logs with component/context information.
- **Errors**: Return meaningful errors. Avoid hiding errors unless intentionally dropped (e.g., malformed messages in workers).

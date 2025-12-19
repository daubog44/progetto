# Go Microservices Data Layer Project

A robust starting point for a Go-based microservices architecture focused on data persistence and observability. This project includes multiple data layer services, integrated hot-reload development, and a comprehensive observability stack.

## 🚀 Technologies

### Core
- **Language**: [Go (1.24)](https://go.dev/)
- **Orchestration**: [Docker Compose](https://docs.docker.com/compose/) & [Tilt](https://tilt.dev/)
- **Hot Reload**: [Air](https://github.com/air-verse/air)

### Data Stores
- **MongoDB**: Document database (port 27017)
- **Cassandra**: Wide-column store (port 9042)
- **Neo4j**: Graph database (port 7474/7687)
- **PostgreSQL**: Relational database (port 5432)
- **Meilisearch**: Search engine (port 7700)
- **MinIO**: S3 compatible object storage (port 9000/9001)

### Observability & Infrastructure
- **Prometheus**: Metrics collection
- **Grafana**: Visualization dashboard (port 3000, admin/admin)
- **Loki & Promtail**: Log aggregation
- **OpenTelemetry (OTEL)**: Distributed tracing and metrics collection
- **Kafka**: Message broker (KRaft mode)

## 📁 Project Structure

```text
├── cassandra-service/    # Cassandra integration service
├── mongo-service/        # MongoDB integration service
├── neo4j-service/        # Neo4j integration service
├── test-service/         # Integration tests using Testcontainers
├── infra/                # Configuration for Prometheus, Loki, etc.
├── proto/                # Shared Protobuf definitions
├── docker-compose.yml    # Infrastructure services
├── Tiltfile              # Development environment orchestration
├── Taskfile.yaml         # Useful commands (similar to Makefile)
└── go.work               # Go workspace configuration
```

## 🛠 Getting Started

### Prerequisites
- [Docker](https://www.docker.com/) & Docker Compose
- [Go 1.24+](https://go.dev/dl/)
- [Tilt](https://tilt.dev/install.html)
- [Task](https://taskfile.dev/install/) (optional, but recommended)

### 1. Initialize Infrastructure
Start the backend services (databases, Kafka, observability):
```bash
task up
# or: docker compose up -d
```

### 2. Launch Development Environment
Tilt provides a unified dashboard for all your Go services with **Hot Reload** enabled. Any change you make in the code will be instantly synced and recompiled inside the containers.
```bash
task dev
# or: tilt up
```
Open the Tilt UI (usually at http://localhost:10350) to see your services running.

### 3. Running Tests
We use [Testcontainers](https://golang.testcontainers.org/) for real integration testing.
```bash
task test
# or: go test -v ./test-service/...
```

## 👨‍💻 Development Workflow

### Adding a New Service
1. Create a new directory for the service.
2. Initialize the module: `go mod init github.com/username/progetto/your-service`.
3. Add the service to `go.work` in the root directory.
4. Create a `Dockerfile` using the multi-stage pattern (include a `dev` stage with `air`).
5. Add the service to `Tiltfile` using `docker_build`.

### Environment Variables
Each service uses a `.env` file and [koanf](https://github.com/knadh/koanf) for configuration. Environment variables should be prefixed with `APP_` (e.g., `APP_MONGO_URI` becomes `mongo.uri` in the code).

---
Created with ❤️ for high-performance data architectures.

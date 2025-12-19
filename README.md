# Go Microservices Data Layer Project

A robust starting point for a Go-based microservices architecture focused on data persistence. This project includes multiple data layer services with integrated hot-reload development via Tilt.

## 🚀 Technologies

### Core
- **Language**: [Go (1.24)](https://go.dev/)
- **Orchestration**: [Docker Compose](https://docs.docker.com/compose/) & [Tilt](https://tilt.dev/)

### Data Stores
- **MongoDB**: Document database (port 27017)
- **Cassandra**: Wide-column store (port 9042)
- **Neo4j**: Graph database (port 7474/7687)

## 📁 Project Structure

```text
├── cassandra-service/    # Cassandra integration service
├── mongo-service/        # MongoDB integration service
├── neo4j-service/        # Neo4j integration service
├── test-service/         # Integration tests using Testcontainers
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
- [Task](https://taskfile.dev/install/) (referred to as `go-task` here)

### 1. Initialize Infrastructure
Start the backend services:
```bash
go-task up
# or: docker compose up -d
```

### 2. Generate Protobuf Code
If you modify `.proto` files, regenerate the Go code:
```bash
go-task proto
# or: cd proto && buf generate
```

### 3. Launch Development Environment
Tilt provides a unified dashboard for all your Go services with **Hot Reload** enabled. Any change you make in the code will be instantly synced and recompiled inside the containers using Tilt's built-in `live_update`.
```bash
go-task dev
# or: tilt up
```
Open the Tilt UI (usually at http://localhost:10350) to see your services running.

### 4. Running Tests
We use [Testcontainers](https://golang.testcontainers.org/) for real integration testing.
```bash
go-task test
# or: go test -v ./test-service/...
```

## 👨‍💻 Development Workflow

### Adding a New Service
1. Create a new directory for the service.
2. Initialize the module: `go mod init github.com/username/progetto/your-service`.
3. Add the service to `go.work` in the root directory.
4. Create a `Dockerfile` using the multi-stage pattern.
5. Add the service to `Tiltfile` using `docker_build` with `live_update` and `restart_process()`.

### Environment Variables
Each service uses a `.env` file and [koanf](https://github.com/knadh/koanf) for configuration. Environment variables should be prefixed with `APP_` (e.g., `APP_MONGO_URI` becomes `mongo.uri` in the code).

---
Created with ❤️ for high-performance data architectures.

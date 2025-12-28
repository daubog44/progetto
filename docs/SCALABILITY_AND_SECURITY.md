# Scalability and Security: Production Readiness

This document outlines the architectural patterns used in the project, security measures implemented, and a roadmap (TODOs) for scaling the system to production.

## 1. Architectural Patterns

### API Gateway Pattern
We use a centralized **API Gateway** (`gateway-service`) as the single entry point for all client requests.
- **Benefits**: Centralized authentication, routing, rate limiting, and observability.
- **Protocol**: HTTP/REST and Server-Sent Events (SSE).

### Saga Pattern (Choreography)
For distributed transactions (e.g., User Registration), we use the **Saga Pattern** with *Choreography*.
1. **Auth Service** creates the user (Pending state).
2. **Auth Service** publishes `user_created` event to Kafka.
3. **Messaging Service** consumes `user_created` and creates a chat profile (Cassandra).
4. **Validation/Compensating Transaction**:
   - If Messaging Service fails, it publishes `user_creation_failed`.
   - **Auth Service** listens for failure and deletes the partial user (Compensation).

### Fan-Out Pattern (for Real-Time Notifications)
Currently, the API Gateway subscribes directly to Kafka to push notifications (like "User Created") to the frontend via **SSE (Server-Sent Events)**.
- **Current Limitation**: In a multi-instance gateway deployment, Kafka Consumer Groups distribute messages. If Instance A holds the user connection but Instance B receives the Kafka message, the user never gets the notification.
- **Solution (Fan-Out)**: See Scalability TODOs below.

### Resiliency Patterns
Implemented via a shared library (`shared/pkg/resiliency`):
- **Circuit Breaker**: Prevents cascading failures by stopping requests to failing services.
- **Retry with Exponential Backoff**: Automatically retries transient failures (e.g., network blips).

## 2. Security Practices

### JWT Authentication with Refresh Tokens
We implement a robust dual-token system:
- **Access Token (JWT)**: Short-lived (e.g., 15 min). Used for API access. Stateless verification by Gateway.
- **Refresh Token**: Long-lived (e.g., 7 days). Stored securely (Redis). Used to obtain new Access Tokens without re-login.
- **Revocation**: Refresh tokens can be revoked (deleted from Redis) to force logout.

### Password Hashing
- **Bcrypt**: User passwords are hashed using Bcrypt before storage in PostgreSQL. We never store plain-text passwords.

## 3. Production Scalability TODOs

### [CRITICAL] Real-Time State (SSE) Scaling
**The Problem**: Our current SSE implementation relies on the Gateway consuming Kafka messages directly using a persistent Consumer Group. This works for 1 Gateway instance. If we scale to 10 Gateways, Kafka will load-balance messages; a message destined for User A might arrive at Gateway 2, but User A is connected to Gateway 1.

**The Solution**:
1.  **Redis Pub/Sub**:
    -   When a backend service generates an event, it publishes it to a Kafka topic.
    -   A dedicated "Notification Service" consume Kafka and republishes to a **Redis Channel** (or directly from backend).
    -   **All Gateway instances** subscribe to the Redis channel. Redis broadcasts to ALL instances (Fan-Out).
    -   Each Gateway checks if User A is connected locally. If yes, send SSE. If no, ignore.
2.  **Alternative (Kafka only)**:
    -   Each Gateway instance generates a unique Consumer Group ID (e.g., `gateway-uuid`).
    -   Kafka broadcasts to ALL consumer groups.
    -   PRO: No Redis dependency for this. CON: Kafka consumer rebalancing storms if Gateways scale up/down frequently.

> [!NOTE]
> **Why not just Sticky Sessions?**
> Sticky Sessions (Session Affinity) ensure the **Client's HTTP connection** stays on a specific Gateway (e.g., Gateway 1). However, Kafka messages from the backend are load-balanced across *all* Gateway instances.
> If the specific Kafka message for User A arrives at Gateway 2, but User A is "stuck" to Gateway 1, Gateway 2 cannot deliver the message.
> Therefore, **Fan-Out (Broadcast)** is required so that *every* Gateway receives the event and checks if it holds the connection for that user.

### Database Scaling
- **PostgreSQL (Auth)**: Use Read Replicas/Connection Pooling (PgBouncer).
- **Cassandra (Messaging)**: Ensure proper Partition Keys (`user_id` vs `chat_id`) to distribute load evenly across nodes.
- **Connection Pooling (Strategy Update)**: 
    -   User requested **Unlimited Connections** (`MaxOpenConns=0`) to facilitate future sharding strategies.
    -   *Risk*: This delegates connection management entirely to the DB server. Ensure Postgres `max_connections` and OS file descriptors (`ulimit`) are high enough.

### Bottlenecks & Anti-Patterns to Fix
1.  **Schema Migration on Startup (Race Condition)**:
    -   *Current*: `messaging-service` creates Keyspace/Table on startup.
    -   *Problem*: If we scale to 5 replicas starting simultaneously, they race to create the table, potentially causing startup crases.
    -   *Fix*: Move schema creation to a separate K8s Job or migration tool (e.g., `golang-migrate`) that runs *before* the service starts.
2.  **Hardcoded Consistency**:
    -   *Current*: Cassandra uses `Quorum` hardcoded in code.
    -   *Fix*: Move this to configuration (`env` vars) to allow tuning Consistency vs Availability without recompiling.
- **Redis (Cache/Session)**: Use Redis Cluster mode.

### Operational Readiness
- [ ] **Dead Letter Queues (DLQ)**: Configure Kafka and Watermill to send permanently failed messages to a DLQ for manual inspection.
- [ ] **Graceful Shutdown**: Ensure all services handle SIGTERM to finish in-flight requests before stopping.
- [ ] **Rate Limiting**: Implement per-IP or per-User rate limiting at the Gateway level (Redis-backed).

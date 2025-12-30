# Architectural References & Case Studies

This document compiles key insights from major tech companies (Uber, Netflix, Discord) regarding architectures relevant to your project: microservices, gRPC, real-time notifications, and distributed transactions (Saga pattern).

## 1. Uber: Next-Gen Push Platform on gRPC

**Context:** Uber needed a reliable, high-scale system to send real-time updates (driver locations, trip status) to millions of mobile users. They transitioned from polling and SSE (Server-Sent Events) to a **gRPC-based bidirectional streaming** solution.

### Key Architectural Decisions:
- **Protocol:** Switched to **gRPC** over QUIC/HTTP3.
  - *Why?* To solve issues with "head-of-line" blocking and poor connectivity on mobile networks. gRPC streaming allows for instant acknowledgments and better connection management.
- **Payload efficiency:** Used **Protocol Buffers (Protobuf)**.
  - *Why?* Significant reduction in bandwidth (vital for mobile data) and faster parsing compared to JSON.
- **Reliability:** Implemented an "at-least-once" delivery guarantee with acknowledgments on the same stream, removing the need for separate ACK requests.
- **Components:**
  - **Netty**: High-performance asynchronous event-driven network application framework.
  - **ZooKeeper/Helix**: For cluster management and sharding of connections.
  - **Redis/Cassandra**: For buffering bursts and persistent storage of messages.

**Takeaway for your project:**
- For your **Notification Service**, considering gRPC streaming for internal service-to-service communication is standard.
- While your client-facing notifications might still use SSE (easier for web), ensuring a robust internal event structure (Protobuf) can save resources.

---

## 2. Netflix: Microservices & API Gateway

**Context:** Netflix operates one of the most complex microservices environments in the world. Their "frontend" is actually a massive orchestration of services accessed via an API Gateway.

### Key Architectural Decisions:
- **API Gateway (Zuul):**
  - Acts as the single entry point for all requests.
  - Handles **Dynamic Routing**, **Monitoring**, **Resiliency** (shedding load), and **Security** (authentication/SSL termination) at the edge.
  - Can run "groovyscripts" to dynamically change routing behavior without redeploying key services.
- **Hystrix (Resiliency):**
  - Implements the **Circuit Breaker** pattern. If a microservice fails or is slow, the gateway fails fast or returns a fallback response instead of cascading the error.
- **Conductor (Orchestration):**
  - A microservices orchestrator to manage workflows (Sagas) that span multiple services. The state of the workflow is maintained centrally, which is safer than pure choreography for complex business logic.

**Takeaway for your project:**
- Your **Gateway Service** is critical. Centralizing auth and routing logic there (like Zuul) keeps your backend services clean.
- Implementing "circuit breaking" logic (e.g., using Watermill or middleware) prevents one failing service from taking down the whole app.

---

## 3. Discord: Scaling Real-Time Notifications

**Context:** Discord is a real-time communication platform handling billions of messages daily. Their challenge is maintaining millions of persistent WebSocket connections.

### Key Architectural Decisions:
- **The Gateway (Elixir):**
  - Built their Gateway service in **Elixir** (running on the Erlang VM), known for handling massive concurrency with lightweight processes.
  - Each user connection is a separate lightweight process, which is fault-tolerant and isolated.
- **Fan-out Architecture:**
  - When a message is sent to a large server, the system "fans out" the notification to thousands of connected sessions efficiently.
  - They use **consistent hashing** to distribute "guild" (server) processes across nodes.
- **Compression:**
  - Heavy use of **Zstandard** compression for WebSocket traffic to reduce bandwidth usage.

**Takeaway for your project:**
- For your **Notification Service**, efficiency in handling idle connections is key.
- If you use Go (which also handles concurrency well with goroutines), ensure you are managing connection state efficiently and handling disconnects gracefully.

---

## 4. The Saga Pattern (Distributed Transactions)

**Context:** In microservices, you cannot use a single database transaction (ACID) across multiple services. The Saga pattern manages consistency.

### Common Implementations:
- **E-Commerce Order Flow (Classic Example):**
  1. **Order Service:** Creates pending order.
  2. **Payment Service:** Charges customer.
  3. **Inventory Service:** Reserves stock.
  4. **Shipping Service:** Creates label.
  - *Failure:* If Payment fails, the Saga triggers a **Compensating Transaction** to Cancel the Order.

### Approaches:
1. **Choreography (Decentralized):**
   - Services emit events ("OrderCreated"). Other services listen and react ("CreditReserved").
   - *Pros:* Simple to start, no central coordinator.
   - *Cons:* Hard to visualize the whole process; "cyclic dependencies" risk.
2. **Orchestration (Centralized):**
   - A centralized "Orchestrator" (like a State Machine) commands services to do work ("ProcessPayment", "UpdateInventory").
   - *Pros:* Clear flow visibility, easier to debug and manage complex failover logic.
   - *Example:* **Uber Cadence** or **Netflix Conductor**.

**Takeaway for your project:**
- You are already implementing Sagas (likely orchestration via your messaging service). Ensuring robust **Compensating Actions** (undo logic) for every step is the most critical part.
- Your "Poison Queue" strategy is a standard industry practice to handle "unprocessable" messages without blocking the entire pipeline.

## References
1. [Uber’s Next-Gen Push Platform on gRPC](https://www.uber.com/en-IT/blog/ubers-next-gen-push-platform-on-grpc/)
2. [Netflix Tech Blog: Optimizing the Netflix API](https://netflixtechblog.com/optimizing-the-netflix-api-5c9ac715cf19)
3. [Discord Engineering: How Discord Scales Elixir to 5,000,000 Concurrent Users](https://discord.com/blog/how-discord-scales-elixir-to-5-000-000-concurrent-users)
4. [Microservices.io: Sage Pattern](https://microservices.io/patterns/data/saga.html)

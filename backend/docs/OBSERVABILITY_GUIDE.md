# LGTM Observability Stack Guide

This guide provides a comprehensive overview of the observability stack implemented in this project, explaining how it works, how to configure it, and how to use it effectively.

## 1. Architecture Overview

We use the **Grafana LGT** stack (Loki, Grafana, Tempo) along with **Prometheus** for metrics and **Grafana Alloy** as the central telemetry collector.

### Components
-   **Grafana**: The visualization UI for all logs, metrics, and traces.
-   **Loki**: Log aggregation system.
-   **Tempo**: Distributed tracing backend.
-   **Alloy**: The central collector agent (OpenTelemetry, Prometheus, etc.). It's the "backbone" of the data flow.
-   **Pyroscope**: Continuous profiling. Used to analyze CPU, memory, and goroutine performance over time.
-   **K6**: Load testing tool used to generate realistic traffic and verify the stack works under pressure.

### Data Flow Diagram

```mermaid
graph TD
    subgraph Microservices
        App[Go Service] -->|OTLP| Alloy
        App -->|Logs| Alloy
        App -->|Metrics| Alloy
    subgraph "Collector & Agent Layer"
        Alloy[Grafana Alloy]
    end

    subgraph "Storage Layer"
        Alloy -->|Traces| Tempo[Grafana Tempo]
        Alloy -->|Metrics| Prometheus[Prometheus]
        Alloy -->|Logs| Loki[Grafana Loki]
        App -->|Profiling| Pyroscope[Grafana Pyroscope]
    end

    subgraph "Visualization & Alerting"
        Grafana --> Tempo
        Grafana --> Prometheus
        Grafana --> Loki
        Grafana --> Pyroscope
    end
```

---

## 2. Deep Dive: How Tracing Works

You might wonder: *How does the system know that an HTTP request to the Gateway, which calls the Auth Service via GRPC, which then publishes a Kafka event, are all part of the same "Trace"?*

The answer is **Context Propagation**.

### The Mechanism
1.  **Trace Context**: Every request is assigned a unique `TraceID`. This ID is carried in the Go `context.Context`.
2.  **Propagation**: When a service calls another (via HTTP, GRPC, or Messaging), it "injects" this `TraceID` into the transport metadata (Headers).
3.  **Extraction**: The receiving service "extracts" the `TraceID` from the metadata and creates a new Span that is a "child" of the previous one.

### In Your Code
We have implemented this automatically in `shared/pkg/observability`:

#### 1. Incoming HTTP Requests (Gateway)
The `observability.Middleware` intercepts every incoming HTTP request.
```go
// shared/pkg/observability/middleware.go
func Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // EXTRACTS trace context from HTTP headers (e.g., traceparent)
        // CREATES a new span for this request
        ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
        // ...
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

#### 2. GRPC Calls (Gateway -> Auth/Post)
We use `google.golang.org/grpc/otelgrpc`.
-   **Client**: When `gateway` calls `auth`, the interceptor INJECTS the context into GRPC metadata.
-   **Server**: When `auth` receives the call, the interceptor EXTRACTS the metadata and creates a child span.

#### 3. Database Calls (Auth -> Postgres, Post -> Mongo)
We use `otelgorm` and `otelmongo`.
-   **Gorm**: `db.WithContext(ctx).Create(...)` passes the active span to the driver, which records the SQL query as a child span.
-   **Mongo**: The `monitor` configured in `clientOpts` captures commands and creates spans linked to the context passed in `collection.InsertOne(ctx, ...)`.

#### 4. Messaging (Watermill/Kafka)
We created a custom wrapper in `shared/pkg/observability/watermill.go` because Watermill's native OTel library was missing.
-   **Publisher**:
    ```go
    // INJECTS trace context into Kafka Message Headers
    otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(msg.Metadata))
    ```
-   **Subscriber**:
    ```go
    // EXTRACTS trace context from Kafka Message Headers
    ctx := otel.GetTextMapPropagator().Extract(msg.Context(), propagation.MapCarrier(msg.Metadata))
    // Starts a new "consume" span linked to the publisher's span
    ```

---

## 3. Grafana Configuration Tutorial

### Access
-   **URL**: `http://localhost:3000`
-   **User**: `admin`
-   **Pass**: `admin`

### Data Sources
Alloy pushes data to these backends. You need to configure them in Grafana if not already provisioned.

1.  **Tempo (Traces)**
    -   Type: `Tempo`
    -   URL: `http://tempo:3200`
    -   *Deep Linking*: In settings, "Loki Search" -> URL `http://loki:3100`. This allows jumping from Trace -> Logs.

1.  **Prometheus (Metrics)**
    -   Type: `Prometheus`
    -   URL: `http://prometheus:9090`

3.  **Loki (Logs)**
    -   Type: `Loki`
    -   URL: `http://loki:3100`
    -   *Derived Fields*: Add a rule to extract `trace_id` from log lines to link Logs -> Traces.
        -   Regex: `trace_id=(\w+)`
        -   Query: `${__value.raw}`
        -   Link: Tempo

4.  **Pyroscope (Profiling)**
    -   Type: `Phlare` / `Pyroscope` (depending on Grafana version, use the Pyroscope icon)
    -   URL: `http://pyroscope:4040`


---

## 5. Advanced Tooling


### 5.2 Grafana Pyroscope (Profiling)
Profiling helps you understand *where* the CPU is spending time or *what* is eating memory line-by-line.
-   **Access**: http://localhost:4040
-   **In Go**: We use the Pyroscope Go agent in `shared/pkg/observability/init.go` (if configured).
-   **Use Case**: If the Auth service is slow, check the Flamegraph in Pyroscope to see which function is the bottleneck.

### 5.3 K6 (Validation)
k6 is not just for load testing; it's our "Traffic Generator" for testing the observability stack itself.
-   **Script**: Located in `test/k6/load.js`.
-   **Execution**: `docker compose run --rm k6`.
-   **Metric**: Look for `k6_http_req_duration` in Prometheus to see how the load test is performing.

---

## 6. How to Verify
1.  **Run Stack**: `tilt up` or `docker compose up -d`.
2.  **Generate Traffic**:
    ```bash
    # Run k6 load test
    docker compose run --rm k6
    ```
3.  **Explore**:
    -   Go to **Grafana > Explore**.
    -   Select **Tempo**.
    -   Query: `Search` -> Service Name: `gateway-service`.
    -   Click a trace.
    -   **Visual**: You will see the waterfall:
        -   `GET /posts` (Gateway)
        -   `grpc.ListenPosts` (PostService)
        -   `mongo.find` (MongoDB)

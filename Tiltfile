# Tiltfile

# --- Docker Compose ---
docker_compose('docker-compose.yml')

# --- Infrastructure ---
# --- Infrastructure ---
dc_resource('cassandra', labels=['Infrastructure'])
dc_resource('mongodb', labels=['Infrastructure'])
dc_resource('neo4j', labels=['Infrastructure'])
dc_resource('redis', labels=['Infrastructure'])
dc_resource('kafka', labels=['Infrastructure'])
dc_resource('postgres', labels=['Infrastructure'])

# --- Tooling (GUIs) ---
dc_resource('kafka-ui', labels=['Tooling'], links=[link('http://localhost:8080', 'Kafka UI')])
dc_resource('cloudbeaver', labels=['Tooling'], links=[link('http://localhost:8978', 'CloudBeaver')])
dc_resource('redis-commander', labels=['Tooling'], links=[link('http://localhost:8082', 'Redis Commander')])
dc_resource('mongo-express', labels=['Tooling'], links=[link('http://localhost:8081', 'Mongo Express')])

# Manual K6 Trigger
local_resource(
    'run-load-test',
    cmd='docker compose run --rm k6 run /scripts/load.js',
    labels=['Tooling'],
    trigger_mode=TRIGGER_MODE_MANUAL,
    auto_init=False,
)

# --- Observability Stack ---
dc_resource('minio', labels=['Observability'], links=[link('http://localhost:9001', 'MinIO Console')])
dc_resource('grafana', labels=['Observability'], links=[link('http://localhost:3000', 'Grafana')])
dc_resource('prometheus', labels=['Observability'], links=[link('http://localhost:9089', 'Prometheus')])
dc_resource('tempo', labels=['Observability']) # Headless
dc_resource('loki', labels=['Observability']) # Headless
dc_resource('pyroscope', labels=['Observability'], links=[link('http://localhost:4040', 'Pyroscope')])
dc_resource('alloy', labels=['Observability'])


# --- Microservices ---

# Cassandra Service
docker_build(
    'cassandra-service',
    '.',
    dockerfile='microservices/cassandra-service/Dockerfile',
    live_update=[
        sync('./microservices/cassandra-service', '/app'),
        sync('./shared', '/shared'),
        run('go build -o /server .'),
        restart_container()
    ]
)
dc_resource('cassandra-service', labels=['Microservices'])

# Neo4j Service
docker_build(
    'neo4j-service',
    '.',
    dockerfile='microservices/neo4j-service/Dockerfile',
    live_update=[
        sync('./microservices/neo4j-service', '/app'),
        sync('./shared', '/shared'),
        run('go build -o /server .'),
        restart_container()
    ]
)
dc_resource('neo4j-service', labels=['Microservices'])


# Auth Service
docker_build(
    'auth-service',
    '.',
    dockerfile='microservices/auth/Dockerfile',
    live_update=[
        sync('./microservices/auth', '/app'),
        sync('./shared', '/shared'),
        run('go build -o /server .'),
        restart_container()
    ]
)
dc_resource('auth-service', labels=['Microservices'])


# Gateway Service
docker_build(
    'gateway-service',
    '.',
    dockerfile='microservices/gateway-service/Dockerfile',
    live_update=[
        sync('./microservices/gateway-service', '/app'),
        sync('./shared', '/shared'),
        run('go build -o /server .'),
        restart_container()
    ]
)
dc_resource('gateway-service', labels=['Microservices'])


# Post Service
docker_build(
    'post-service',
    '.',
    dockerfile='microservices/post-service/Dockerfile',
    live_update=[
        sync('./microservices/post-service', '/app'),
        sync('./shared', '/shared'),
        run('go build -o /server .'),
        restart_container()
    ]
)
dc_resource('post-service', labels=['Microservices'])

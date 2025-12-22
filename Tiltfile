# Tiltfile

# --- Docker Compose ---
docker_compose('docker-compose.yml')

# --- Services ---

# Mongo Service
docker_build(
    'mongo-service',
    '.',
    dockerfile='microservices/mongo-service/Dockerfile',
    live_update=[
        sync('./microservices/mongo-service', '/app'),
        sync('./shared', '/shared'),
        run('go build -o /server main.go'),
        restart_container()
    ]
)

# Cassandra Service
docker_build(
    'cassandra-service',
    '.',
    dockerfile='microservices/cassandra-service/Dockerfile',
    live_update=[
        sync('./microservices/cassandra-service', '/app'),
        sync('./shared', '/shared'),
        run('go build -o /server main.go'),
        restart_container()
    ]
)

# Neo4j Service
docker_build(
    'neo4j-service',
    '.',
    dockerfile='microservices/neo4j-service/Dockerfile',
    live_update=[
        sync('./microservices/neo4j-service', '/app'),
        sync('./shared', '/shared'),
        run('go build -o /server main.go'),
        restart_container()
    ]
)

docker_build(
    'auth',
    '.',
    dockerfile='microservices/auth/Dockerfile',
    live_update=[
        sync('./microservices/auth', '/app'),
        sync('./shared', '/shared'),
        run('go build -o /server main.go'),
        restart_container()
    ]
)










# Tiltfile

# --- Docker Compose ---
docker_compose('docker-compose.yml')

# --- Services ---

# Mongo Service
docker_build(
    'mongo-service',
    '.',
    dockerfile='./mongo-service/Dockerfile',
    target='dev',
    live_update=[
        sync('./mongo-service', '/app'),
        sync('./shared', '/shared'),
        run('go build -o /server main.go'),
        restart_container()
    ]
)

# Cassandra Service
docker_build(
    'cassandra-service',
    '.',
    dockerfile='./cassandra-service/Dockerfile',
    target='dev',
    live_update=[
        sync('./cassandra-service', '/app'),
        sync('./shared', '/shared'),
        run('go build -o /server main.go'),
        restart_container()
    ]
)

# Neo4j Service
docker_build(
    'neo4j-service',
    '.',
    dockerfile='./neo4j-service/Dockerfile',
    target='dev',
    live_update=[
        sync('./neo4j-service', '/app'),
        sync('./shared', '/shared'),
        run('go build -o /server main.go'),
        restart_container()
    ]
)

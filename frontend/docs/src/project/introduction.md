# Introduzione al Progetto

## Scopo del Progetto

Il progetto mira a realizzare un'infrastruttura backend a microservizi robusta e scalabile, pensata per gestire elevati carichi di lavoro e flussi di dati complessi. Il focus principale è sull'affidabilità, la manutenibilità e l'osservabilità del sistema.

## Stack Tecnologico

### Backend

- **Linguaggio**: Go (Golang)
- **Comunicazione Sincrona**: gRPC / HTTP
- **Comunicazione Asincrona**: Kafka (con libreria Watermill)
- **Database**: MongoDB, Neo4j, Cassandra, Redis

### Frontend

- **Framework**: Vue 3
- **Build Tool**: Vite
- **Documentation**: VitePress

### Infrastruttura & DevOps

- **Containerizzazione**: Docker & Docker Compose
- **Orchestrazione Locale**: Tilt
- **Reverse Proxy**: Traefik (o Gateway dedicato)

## Caratteristiche Chiave

- **Saga Pattern**: Gestione transazioni distribuite.
- **CQRS**: Separazione tra comandi e query (dove applicabile).
- **Event Sourcing**: Persistenza basata sugli eventi.

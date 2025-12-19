# 🌐 Go Social Network: Modern Microservices Architecture

Benvenuti in **Go Social Network**, un progetto progettato per dimostrare e implementare un'architettura a microservizi allo stato dell'arte. L'obiettivo principale è costruire una piattaforma social scalabile, performante e manutenibile, utilizzando gli standard più moderni dell'ecosistema Go e del mondo cloud-native.

## 🎯 Visione del Progetto

Vogliamo realizzare un social network che sfrutti al massimo la separazione delle responsabilità (Separation of Concerns). Ogni modulo è un microservizio isolato che comunica tramite protocolli ultra-veloci come **gRPC**, garantendo prestazioni elevate e type-safety end-to-end.

### Punti di Forza:
- **Scalabilità Orizzontale**: Ogni servizio può scalare indipendentemente.
- **Polyglot Persistence**: Usiamo il database giusto per ogni compito (Mongo per documenti, Cassandra per time-series/feed, Neo4j per i grafi delle relazioni).
- **Developer Experience (DX)**: Sviluppo locale fluido grazie a **Tilt** con hot-reload istantaneo.
- **Standard Moderni**: Utilizzo di **Buf v2** per gRPC e **Testcontainers** per test d'integrazione affidabili.

---

## 📁 Architettura e Struttura

Il progetto è organizzato come un **monorepo** per facilitare la gestione del codice condiviso e dei contratti gRPC.

```text
├── microservices/
│   ├── mongo-service/     # Gestione profili e metadati (MongoDB)
│   ├── cassandra-service/ # Gestione Feed e attività (Cassandra)
│   ├── neo4j-service/     # Gestione relazioni tra utenti (Neo4j)
│   └── test-service/      # Integration tests using Testcontainers
├── shared/
│   └── proto/             # Contratti gRPC (Buf v2) e codice generato
├── tutorial/              # Documentazione interna e guide all'apprendimento
├── docker-compose.yml     # Infrastruttura di base (DB, Broker, etc.)
├── Tiltfile               # Orchestrazione dello sviluppo locale
└── Taskfile.yaml          # Automazione dei task comuni
```

---

## 🚀 Tecnologie Core

- **Language**: [Go (1.25)](https://go.dev/)
- **Communication**: [gRPC](https://grpc.io/) & [Protobuf](https://protobuf.dev/) gestiti via [Buf v2](https://buf.build/)
- **Orchestration**: [Docker Compose](https://docs.docker.com/compose/) & [Tilt](https://tilt.dev/)
- **Data Layer**: MongoDB, Cassandra, Neo4j
- **Testing**: [Testcontainers](https://golang.testcontainers.org/) per test reali in Go

---

## 🛠 Iniziare lo Sviluppo

### 1. Prerequisiti
Assicurati di avere installato:
- **Docker** & Docker Compose
- **Go 1.25+**
- **Tilt** (per l'ambiente di sviluppo)
- **Task** (esegui `go-task` invece di `make`)

### 2. Setup Rapido
Clona il repository e avvia l'infrastruttura di base (Database):
```bash
go-task up
```

### 3. Generazione API (Protobuf)
Utilizziamo **Buf v2** con **Managed Mode**. Per rigenerare il codice client/server gRPC:
```bash
go-task proto
```

### 4. Sviluppo con Tilt (Hot Reload)
Avvia l'intero ecosistema di microservizi in un unico comando. Tilt monitorerà i tuoi file e aggiornerà i container in tempo reale:
```bash
go-task dev
```
🔗 Accedi alla dashboard di Tilt su: [http://localhost:10350](http://localhost:10350)

---

## 👨‍💻 Workflow per i Contributori

### Aggiungere un Nuovo Microservizio
Per mantenere la qualità e la coerenza del progetto, segui questi step:
1. **Directory**: Crea una cartella in `services/`.
2. **Modulo Go**: Esegui `go mod init` all'interno.
3. **Workspace**: Aggiungi il nuovo path a `go.work`.
4. **Infra**: Aggiungi eventuali dipendenze (DB) a `docker-compose.yml`.
5. **Dockerfile**: Crea un Dockerfile multi-stage basato sugli esempi esistenti.
6. **Tilt**: Registra il servizio nel `Tiltfile` usando `docker_build` e `restart_container()`.

### Best Practices gRPC
Tutte le modifiche ai contratti devono avvenire in `shared/proto/`. Consulta la nostra [Guida Workflow gRPC](file:///home/daubog44/Scrivania/dev/progetto/tutorial/grpc-workflow.md) per i dettagli sull'uso di `Unimplemented...Server`.

---
Realizzato con ⚡ e ❤️ per il futuro dello sviluppo distribuito.

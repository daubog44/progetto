# Struttura del Progetto & Monorepo

Il progetto è organizzato come un **Monorepo**, contenente sia il codice backend che frontend. Questa struttura facilita la gestione delle dipendenze, la coerenza del versionamento e lo sviluppo full-stack.

## Root Directory

```
/
├── backend/            # Microservizi Go, Shared Libs, Configs
├── frontend/           # Vue 3 SPA, Design System, Docs
├── Taskfile.yaml       # Orchestrator per comandi di progetto (build, dev, test)
└── README.md           # Entry point documentazione generale
```

## Backend Structure (`/backend`)

Il backend segue i principi della **Clean Architecture** e organizza il codice in servizi distinti e librerie condivise.

```
/backend
├── microservices/          # Cartella contenente tutti i microservizi
│   ├── auth-service/       # Gestione autenticazione e JWT
│   ├── post-service/       # Gestione contenuti (Post, Commenti)
│   ├── social-service/     # Gestione grafo sociale (Neo4j)
│   ├── messaging-service/  # Chat e messaggistica
│   ├── search-service/     # Ricerca Full-Text (Meilisearch)
│   ├── notification/       # Notifiche e Email
│   └── gateway-service/    # API Gateway e SSE
├── shared/                 # Librerie condivise (Go Modules)
│   ├── pkg/                # Packages riutilizzabili (logger, database, middlewares)
│   └── pb/                 # Definizioni Protobuf e codice gRPC generato
├── deploy/                 # Configurazioni Kubernetes / Helm
├── scripts/                # Utility scripts (es. setup DB)
├── docker-compose.yml      # Definizione stack locale
├── Tiltfile                # Configurazione Tilt per sviluppo live Kubernetes/Docker
└── go.work                 # Go Workspace file per gestione multi-modulo
```

### Go Workspaces

Utilizziamo **Go Workspaces** (`go.work`) per permettere lo sviluppo simultaneo su più moduli (microservizi e libreria shared) senza dover pubblicare versioni intermedie.

### Standard Microservice Layout

Ogni microservizio segue una struttura interna rigorosa:

- `internal/app`: **Bootstrap**. Esegue il wiring delle dipendenze e avvia l'applicazione.
- `internal/api`: **Transport**. Implementazione dei server gRPC o HTTP.
- `internal/events`: **Async**. Configurazione di Watermill per Kafka (Router, Pub/Sub).
- `internal/handler`: **Logic**. Gestori di business logic chiamati da API o Eventi.
- `internal/repository`: **Data**. Accesso al database.
- `main.go`: **Entrypoint**. Minimalista, chiama `app.Run()`.

## Frontend Structure (`/frontend`)

Il frontend è una Single Page Application moderna.

```
/frontend
├── src/
│   ├── components/     # Componenti UI riutilizzabili (Shadcn)
│   ├── views/          # Pagine dell'applicazione
│   ├── stores/         # State management (Pinia)
│   ├── router/         # Definizioni rotte
│   └── api/            # Client HTTP/gRPC generati o manuali
├── docs/               # Questa documentazione (VitePress)
├── public/             # Assets statici
└── package.json        # Dipendenze Node.js
```

## Orchestrazione con Taskfile

Utilizziamo [Task](https://taskfile.dev) come task runner agnostico (alternativa moderna a Make).
Esempi di comandi:

- `task dev`: Avvia tutto lo stack in modalità sviluppo.
- `task build`: Compila tutti i servizi e il frontend.
- `task test`: Esegue suite di test backend e frontend.

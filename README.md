# 🌐 Vibely: Piattaforma Culturale Microservices

Benvenuti in **Vibely**, un ecosistema culturale completo dedicato a **libri, film, serie TV e musica**. Questo progetto implementa un'architettura a microservizi allo stato dell'arte per gestire community, recensioni, chat spoiler-safe e cataloghi multimediali.

## 🎯 Visione del Progetto

Vibely mira a creare uno spazio dove gli appassionati di cultura possono connettersi. Ogni area (Libri, Film, Musica) è supportata da una solida infrastruttura distribuita che garantisce prestazioni elevate e scalabilità.

### Caratteristiche Principali:
- **Ecosistema Culturale**: Schede dettagliate per ogni opera con community di genere.
- **Chat Spoiler-Safe**: Accesso alle conversazioni riservato a chi ha completato l'opera.
- **Archivio Personale**: Gestione dello stato di avanzamento (letture, visioni, ascolti).
- **Spazio Emergenti**: Vetrina per nuovi talenti, autori e cantanti.

### Punti di Forza Tecnici:
- **Resilienza**: Pattern avanzati (Circuit Breaker, Retry, Recovery) implementati via librerie condivise.
- **Osservabilità Completa**: Stack LGTM (Loki, Grafana, Tempo, Mimir-compatible) con OpenTelemetry (oTel) e Alloy.
- **Event-Driven**: Architettura reattiva basata su **Kafka** e **Watermill** (CQRS/Saga).
- **Scalabilità**: Ogni servizio scala indipendentemente (Stateless).
- **Polyglot Persistence**: Mongo, Postgres, Redis, Cassandra.
- **Type-Safety**: gRPC/Protobuf per comunicazione inter-servizio.

---

## 📁 Architettura e Struttura

Il progetto è organizzato come un **monorepo**.

```text
├── docs/                  # Documentazione
├── microservices/
│   ├── auth/              # Gestione Utenti (Postgres + Redis + Watermill)
│   ├── post-service/      # Gestione Post / Feed (MongoDB + Watermill)
│   ├── gateway-service/   # API Gateway (Huma + SSE)
│   └── messaging-service/ # Gestione Messaggi & Chat (Cassandra + Watermill)
├── shared/
│   ├── pkg/               # Librerie Condivise (grpcutil, watermillutil, observability, resiliency)
│   └── proto/             # Contratti gRPC (Buf v2)
├── deploy/                # Configurazioni (Grafana, Loki, Tempo, Alloy)
├── docker-compose.yml     # Infrastruttura (Kafka, DB, GUI Tools, LGTM Stack)
├── Tiltfile               # Orchestrazione dev locale e hot-reload
└── Taskfile.yaml          # Automazione
```

---

## 🛠 Iniziare lo Sviluppo

### 1. Requisiti
- Docker & Docker Compose
- Go 1.22+
- `task` (Taskfile)

### 2. Avvio Infrastruttura
Esegui l'intera stack (Microservizi + DB + Tools):
```bash
docker compose up -d
```
I dati persistenti verranno salvati nella cartella locale `./data/`.

### 3. Sviluppo con Tilt (Opzionale)
Se hai Tilt installato, per hot-reload:
```bash
tilt up
```
🔗 Dashboard Tilt: [http://localhost:10350](http://localhost:10350)

### 4. Database GUI Tools
Una volta avviato, accedi agli strumenti di gestione:
- **CloudBeaver** (Postgres/Cassandra): [http://localhost:8978](http://localhost:8978)
- **Redis Commander**: [http://localhost:8082](http://localhost:8082)
- **Mongo Express**: [http://localhost:8081](http://localhost:8081)


---

## 📚 Documentazione

La documentazione è stata unificata per semplicità:

- **[ARCHITECTURE.md](backend/docs/ARCHITECTURE.md)**: Visione d'insieme, stack tecnologico e workflow (Sequence Diagrams).
- **[DATA_MODELS.md](backend/docs/DATA_MODELS.md)**: Schemi database attuali (Postgres, Mongo, Redis).
- **[DEVELOPMENT.md](backend/docs/DEVELOPMENT.md)**: Guida operativa (Script, gRPC, Coding Standards).
- **[ARCHITECTURAL_CHALLENGES.md](backend/docs/architectural_challenges.md)**: Deep dive su sfide tecniche, notifiche Real-Time, migrazioni e ottimizzazioni.


---
Realizzato con ⚡ e ❤️ per Vibely.
```

//TODO: logica di caching
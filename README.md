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
- **Scalabilità Orizzontale**: Ogni servizio scala indipendentemente.
- **Polyglot Persistence**: Database specifico per ogni workload (Mongo, Cassandra, Neo4j).
- **Hot-Reload**: Sviluppo fluido con **Tilt**.
- **Type-Safety**: gRPC e **Buf v2** per contratti rigorosi.

---

## 📁 Architettura e Struttura

Il progetto è organizzato come un **monorepo**.

```text
├── docs/
│   ├── tech/              # Specifiche tecniche e visione (Vibely)
│   └── tutorial/          # Guide all'architettura e workflow
├── microservices/
│   ├── mongo-service/     # Gestione profili e metadati opere (MongoDB)
│   ├── cassandra-service/ # Feed, attività e messaggistica (Cassandra)
│   ├── neo4j-service/     # Relazioni social e grafi di interesse (Neo4j)
│   └── test-service/      # Integration tests
├── shared/
│   └── proto/             # Contratti gRPC (Buf v2)
├── docker-compose.yml     # Infrastruttura (Kafka, DB, etc.)
├── Tiltfile               # Orchestrazione dev locale
└── Taskfile.yaml          # Automazione
```

---

## 🛠 Iniziare lo Sviluppo

### 1. Setup Rapido
Esegui l'infrastruttura di base:
```bash
go-task up
```

### 2. Sviluppo con Tilt
Avvia i microservizi con hot-reload:
```bash
go-task dev
```
🔗 Dashboard Tilt: [http://localhost:10350](http://localhost:10350)

---

## 📚 Documentazione e Tutorial

Consulta le nostre guide dettagliate per comprendere il funzionamento interno:

- [Workflow gRPC & Buf v2](docs/tutorial/grpc-workflow.md) - Come generare codice dai contratti.
- [Architettura Event-Driven](docs/tutorial/event-driven-architecture.md) - Integrazione tra gRPC e Kafka.
- [Descrizione Piattaforma](docs/tech/description.md) - Visione e obiettivi di Vibely.

---
Realizzato con ⚡ e ❤️ per Vibely.
```

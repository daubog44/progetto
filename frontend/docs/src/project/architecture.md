# 🏛 Architettura del Sistema

Vibely è progettata come una piattaforma a **microservizi** scalabile, che utilizza un approccio **Hybrid Live** per garantire persistenza, performance e interazioni in tempo reale.

## 🎯 Obiettivi Architetturali

- **Scalabilità Orizzontale**: Ogni servizio è stateless e scalabile indipendentemente.
- **Resilienza**: Fallimenti isolati non abbattono l'intero sistema.
- **Event-Driven**: Disaccoppiamento tramite Kafka per operazioni asincrone.
- **Polyglot Persistence**: Utilizzo del database "giusto" per ogni tipo di dato.

---

## 🗺️ Diagramma del Sistema

Il seguente diagramma illustra le interazioni tra Client, Gateway, Microservizi e Data Layer.

```mermaid
flowchart TB
    subgraph Client_Layer ["Client Layer"]
        User(("User / App"))
    end

    subgraph Edge_Layer ["Edge Layer"]
        Gateway["Gateway Service<br/>(Huma/Go)"]
    end

    subgraph Service_Layer ["Microservices Layer"]
        Auth["Auth Service"]
        Post["Post Service"]
        Social["Social Service"]
        Messaging["Messaging Service"]
        Search["Search Service"]
        Notif["Notification Service"]
    end

    subgraph Data_Layer ["Data & State Layer"]
        PG[("Postgres<br/>(Auth DB)")]
        Mongo[("MongoDB<br/>(Post DB)")]
        Neo4j[("Neo4j<br/>(Social Graph)")]
        Cass[("Cassandra<br/>(Messages/Feed)")]
        Meili[("Meilisearch<br/>(Full Text)")]
        Redis[("Redis<br/>(Cache/PubSub)")]
        Kafka{"Kafka<br/>(Event Bus)"}
    end

    %% Client Interactions
    User <-->|HTTP/REST| Gateway
    User <-->|SSE (Events)| Gateway

    %% Gateway to Services (gRPC)
    Gateway <-->|gRPC| Auth
    Gateway <-->|gRPC| Post
    Gateway <-->|gRPC| Social
    Gateway <-->|gRPC| Messaging
    Gateway <-->|gRPC| Search

    %% Service to Data
    Auth <--> PG
    Auth <--> Redis
    Post <--> Mongo
    Social <--> Neo4j
    Messaging <--> Cass
    Search <--> Meili

    %% Event Driven (Writes)
    Auth -.->|"Pub: UserCreated"| Kafka
    Post -.->|"Pub: PostCreated"| Kafka
    Social -.->|"Pub: Followed"| Kafka
    Messaging -.->|"Pub: MessageSent"| Kafka

    %% Search Indexing (Async)
    Kafka == "Consumer Group" ==> Search

    %% Notification Flow (Reads)
    Kafka == "Consumer Group" ==> Notif
    Notif -- "Publish Targeted" --> Redis
    Redis -- "Sub (InstanceID)" --> Gateway
```

---

## 🏗 Stack Tecnologico

| Componente         | Tecnologia       | Responsabilità                                                        |
| :----------------- | :--------------- | :-------------------------------------------------------------------- |
| **Auth Service**   | Go + Postgres    | Gestione identità, sessioni (JWT) e profili base.                     |
| **Post Service**   | Go + MongoDB     | Gestione contenuti (Post) e commenti. Store flessibile per documenti. |
| **Social Service** | Go + Neo4j       | Gestione del grafo sociale (Follower/Following) e raccomandazioni.    |
| **Messaging**      | Go + Cassandra   | Chat privata e Feed Timeline. Ottimizzato per scritture veloci.       |
| **Search**         | Go + Meilisearch | Ricerca full-text su utenti e contenuti.                              |
| **Notification**   | Go + Redis/SSE   | Orchestrazione eventi e push notifiche real-time al frontend.         |
| **Event Bus**      | Kafka            | Comunicazione asincrona e consistenza eventuale tra servizi.          |

---

## 🧩 Modello di Comunicazione

Il sistema utilizza un approccio ibrido:

### 1. Sincrono (gRPC)

Utilizzato per operazioni "bloccanti" dove il client necessita di una risposta immediata o di dati consistenti.

- **Esempio**: Login, Lettura Profilo, Fetch dettaglio Post.
- **Protocollo**: gRPC interno tra Gateway e Microservizi.

### 2. Asincrono (Event-Driven)

Utilizzato per effetti collaterali, propagazione dati e notifiche.

- **Esempio**: "Utente Creato" -> Crea nodo grafo -> Indidicizza per ricerca -> Invia Email benvenuto.
- **Pattern**: Saga Orchestration / Choreography tramite Kafka.

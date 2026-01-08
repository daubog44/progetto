# 🔄 Data Flows & Operazioni

In Vibely, le operazioni sono divise in due categorie principali: **Sincrone** (User-facing, immediate) e **Asincrone** (System-facing, eventual consistency).

---

## 1. Registrazione Utente (Pattern Saga)

La registrazione è un processo distribuito che garantisce la coerenza tra i vari microservizi (SQL, Graph, Search).

**Trigger**: `POST /auth/register`

```mermaid
sequenceDiagram
    participant Client
    participant API as Gateway/Auth
    participant Kafka
    participant Social as Social (Neo4j)
    participant Search as Search (Meilisearch)
    participant Notif as Notification
    participant SSE as Gateway/SSE

    Note over Client, API: Fase Sincrona (Bloccante)
    Client->>API: Invia Dati Registrazione (Step 1)
    API->>API: Valida & Crea User su Postgres (Pending)
    API->>Kafka: Pubblica Evento `UserCreated`
    API-->>Client: 201 Created (Pending)

    Note over Kafka, Notif: Fase Asincrona (Parallela)
    par Parallel Processing
        Kafka->>Social: Consume `UserCreated`
        Social->>Social: Crea Nodo :Person
        Social->>Kafka: Pubblica `UserSyncedSocial`
    and
        Kafka->>Search: Consume `UserCreated`
        Search->>Search: Indicizza Documento
        Search->>Kafka: Pubblica `UserSyncedSearch`
    and
        Kafka->>Notif: Consume `UserCreated` (Aggregator)
        Notif->>Notif: Start Tracking
    end

    Note over Notif, SSE: Completamento (Saga End)
    Kafka->>Notif: Consume `UserSynced*`
    Notif->>Notif: Check All Synced
    Notif->>SSE: Publish `OnboardingCompleted` (Redis PubSub)
    SSE->>Client: Push Evento SSE (Step 2: Complete)
```

### Gestione Errori (Non implementata in MVP)

Se uno dei consumer fallisce e esaurisce i retry, si attiva una **Compensating Transaction** (es. `UserCreationFailed`) che notifica l'Auth Service per cancellare l'utente o si notifica l'utente dell'errore.

---

## 2. Gestione del Feed (La Timeline - Fan-out)

**Obiettivo:** Caricamento istantaneo della Home.
**Approccio:** Fan-out on Write (Pre-calcolo).

- **Fase Sincrona (Bloccante):** Salvataggio contenuto su MongoDB e Cache Preview.
- **Fase Asincrona (Non Bloccante):** Propagazione ID ai follower (Fan-out).

```mermaid
sequenceDiagram
    participant User as Utente (Writer)
    participant API
    participant Neo4j as Social (Graph)
    participant Cass as Cassandra (Timeline)
    participant Mongo as Mongo (Content)
    participant Redis as Redis (Cache)
    participant Worker as Async Worker

    Note over User, Redis: Fase Sincrona (Immediata)
    User->>API: 1. Pubblica Post
    API->>Mongo: 2. Salva Documento
    API->>Redis: 3. Cache Preview (TTL 24h)
    API->>Worker: 4. Enqueue Job "Fan-out" (Kafka/Redis)
    API-->>User: 201 Created (OK)

    Note over Worker, Cass: Fase Asincrona (Background)
    Worker->>Neo4j: 5. Get Follower IDs
    loop Per ogni Follower
        Worker->>Cass: 6. Insert Post_ID in Timeline
    end

    Note over User, Redis: Fase di Lettura (Reader)
    User->>API: 7. Get Feed Home
    API->>Cass: 8. Get Post IDs
    API->>Redis: 9. MGET (Multi-Get) Dati
    alt Cache Miss
        Redis-->>API: Null
        API->>Mongo: Fetch fallback
        API->>Redis: Set Cache
    end
    API-->>User: 10. JSON Feed
```

---

## 3. Gestione "Live" dei Commenti

**Obiettivo:** Nuovi commenti in real-time senza invalidare la cache.
**Tecnica:** Write-Through `LPUSHX` + SSE.

- **Fase Sincrona:** Persistenza su DB e aggiornamento cache Redis.
- **Fase Asincrona (Real-time):** Invio notifica via SSE ai client connessi.

```mermaid
sequenceDiagram
    participant User as Utente A
    participant API
    participant DB as MongoDB (Data)
    participant Redis as Redis (PubSub/List)
    participant SSE as SSE Server
    participant ClientB as Utente B

    Note over User, Redis: Fase Sincrona (Write)
    User->>API: 1. Invia Commento
    par Parallel Write
        API->>DB: 2. Persist Comment
        API->>Redis: 3. LPUSHX (Safe Update Cache)
        API->>Redis: 4. PUBLISH "NEW_COMMENT"
    end
    API-->>User: 201 Created (OK)

    Note over Redis, ClientB: Fase Asincrona (Real-time)
    Redis-->>SSE: 5. Evento PubSub Trigger
    SSE->>ClientB: 6. Push JSON (via SSE)
```

---

## 4. Gestione "Live" dei Mi Piace (Contatori)

**Obiettivo:** Aggiornamento veloce, evitare "reset a 1", throttling.
**Tecnica:** Redis Lua Script (`INCR_IF_EXISTS`) + SSE Throttling.

- **Fase Sincrona:** Aggiornamento contatore Redis (veloce) e Persistenza DB.
- **Fase Asincrona:** Throttling e invio evento aggregato via SSE.

```mermaid
sequenceDiagram
    participant User as Utente A
    participant API
    participant DB as MongoDB/Neo4j
    participant Redis
    participant SSE
    participant Viewer as Utente B

    Note over User, Redis: Fase Sincrona (Write)
    User->>API: Like Post
    par Parallel Write
        API->>DB: Persist Like
        API->>Redis: LUA Script (Incr IF Exists)
        API->>Redis: PUBLISH "NEW_LIKE"
    end
    API-->>User: 200 OK

    Note over SSE, Viewer: Fase Asincrona (Notify)
    Redis-->>SSE: Evento ricevuto
    SSE->>SSE: Aggrega (+1) (Throttling 2s)
    SSE->>Viewer: Send { likes: 505 }
```

---

## 5. Riepilogo Strategia Caching

| Tipo Dato     | Master DB   | Strategia Cache (Redis)           | Comando     | TTL |
| :------------ | :---------- | :-------------------------------- | :---------- | :-- |
| **Post Body** | MongoDB     | **Read-Through** (Leggo se manca) | `GET`/`SET` | 24h |
| **Feed**      | Cassandra   | **No Cache** (Già ottimizzato)    | -           | -   |
| **Commenti**  | MongoDB     | **Write-Update** (`LPUSHX`)       | `LPUSHX`    | 4h  |
| **Likes**     | Mongo/Neo4j | **Write-Update** (Lua Incr)       | `EVAL`      | 24h |

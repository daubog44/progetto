
# 🏛 Architettura Vibely

## 🎯 Visione & Obiettivi
Vibely è una piattaforma a microservizi per community culturali.
**Obiettivo**: Creare un sistema scalabile, resiliente e manutenibile per gestire Identità, Contenuti e Interazioni Sociali.

> 📘 **Approfondimento Tecnico**: Per i dettagli sulle sfide superate (es. orchestrazione notifiche, fallback, migrazioni), vedi **[ARCHITECTURAL_CHALLENGES.md](architectural_challenges.md)**.

---

## 🏗 Stack Tecnologico (Hybrid Live)

Affinato per garantire persistenza, performance e real-time.

| Componente | Tecnologia | Ruolo |
| :--- | :--- | :--- |
| **Source of Truth** | MongoDB | Document Store principale (Post, Commenti). |
| **Timeline Store** | Cassandra | Time-series ordinata solo per gli ID del feed (Fan-out). |
| **Cache & Pub/Sub** | Redis | Caching preview, conteggi write-through e messaggistica SSE. |
| **Real-Time** | SSE + Huma | Push "Live" verso il client (es. nuovi commenti/like). |
| **Media** | MinIO (S3) | Object Storage per immagini/video. |

---

## 🏎 Strategia Caching & Performance

Ecco la tabella di riferimento per la gestione dei dati in Redis:

| Tipo Dato | Dove vive (Master) | Strategia Redis (Cache) | Comando Critico | TTL |
| --- | --- | --- | --- | --- |
| **Post Body** | MongoDB | **Read-Through** (Leggo se manca) | `GET` / `SET` | **24 Ore** |
| **Feed (Home)** | Cassandra | **Non cachato** (Già ottimizzato) | N/A | N/A |
| **Commenti** | MongoDB | **Write-Update** (Aggiorno se c'è) | `LPUSHX` | **4 Ore** |
| **Likes (Count)** | MongoDB | **Write-Update** (Incr. se c'è) | `EVAL` (Lua) | **24 Ore** |
| **Likes (Stream)** | N/A | **Pub/Sub** (Effimero) | `PUBLISH` | Real-time |

---

## 🔄 Workflow Principali

### 1. Feed Timeline (Fan-out on Write)
Obiettivo: Caricamento istantaneo della Home. L'ID viene propagato ai follower in scrittura.

```mermaid
sequenceDiagram
    participant User as Utente (Writer)
    participant API
    participant Mongo as MongoDB (Content)
    participant Cass as Cassandra (Feed IDs)
    participant Redis as Redis (Cache)

    Note over User, Redis: Fase di Pubblicazione
    User->>API: 1. Pubblica Post
    API->>Mongo: 2. Salva Documento
    API->>Redis: 3. Salva Preview Cache (TTL 24h)
    API->>Cass: 4. Inserisce Post_ID a Follower (Fan-out)
    
    Note over User, Redis: Fase di Lettura
    User->>API: 5. Chiede Feed Home
    API->>Cass: 6. Get Post IDs
    API->>Redis: 7. MGET Dati Post
    alt Cache Miss
        Redis-->>API: Null
        API->>Mongo: Fetch Dati
        API->>Redis: Set Cache
    end
    API-->>User: 8. Feed JSON
```

### 2. Live Comments (Write-Through)
Obiettivo: Aggiornare la lista commenti in cache e notificare via SSE senza race conditions.

```mermaid
flowchart TD
    subgraph Write_Path ["Utente A commenta"]
    WA["Post Comment"] --> API_W["API Backend"]
    API_W -->|1. Persist| DB[("MongoDB")]
    API_W -->|2. Safe Update| R_Cache[("Redis List")]
    API_W -->|3. Publish| R_PubSub{"Redis Pub/Sub"}
    end

    subgraph Stream_Path ["Utente B guarda il post"]
    Client_B["App Utente B"] -->|"Connessione SSE"| API_S["API SSE Server"]
    API_S -.->|Subscribe| R_PubSub
    R_PubSub -->|"4. Event: NEW_COMMENT"| API_S
    API_S -->|"5. Push JSON"| Client_B
    end

    note["Redis Command: LPUSHX<br/>(Push solo se la lista esiste)"]
    note -.-> R_Cache
```

### 3. Live Likes (Throttling)
Obiettivo: Contatore "vivo" ma solido. Lua script per l'atomicità, Throttling per non saturare il client.

```mermaid
sequenceDiagram
    participant U as Utente A
    participant API
    participant DB as MongoDB
    participant Redis
    participant SSE
    participant V as Utente B

    U->>API: Like
    par Parallel & Atomic
        API->>DB: Insert Like
        API->>Redis: Lua Script (Incr IF exists)
        API->>Redis: Publish "NEW_LIKE"
    end
    
    rect rgb(240, 240, 240)
    Note over SSE, V: Throttling (2s buffer)
    Redis-->>SSE: Evento
    SSE->>SSE: Aggrega (+1)
    SSE->>V: Push Event {likes: 505}
    end
```

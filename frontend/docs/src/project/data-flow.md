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

    Note over Client, API: Fase Sincrona (Bloccante)
    Client->>API: Invia Dati Registrazione
    API->>API: Valida & Crea User su Postgres
    API->>Kafka: Pubblica Evento `UserCreated`
    API-->>Client: 201 Created (Immediato)

    Note over Kafka, Notif: Fase Asincrona (Eventual Consistency)
    par Parallel Processing
        Kafka->>Social: Consume `UserCreated`
        Social->>Social: Crea Nodo :Person
    and
        Kafka->>Search: Consume `UserCreated`
        Search->>Search: Indicizza Documento
    and
        Kafka->>Notif: Consume `UserCreated`
        Notif->>Notif: Invia Email Benvenuto
    end
```

### Gestione Errori (Non implementata in MVP)

Se uno dei consumer fallisce, si attiva una **Compensating Transaction** (es. `UserCreationFailed`) per cancellare l'utente o marchiarlo come "incompleto".

---

## 2. Pubblicazione Post (Fan-out on Write)

Per garantire una lettura veloce della Home Feed, il costo computazionale viene spostato in fase di scrittura.

**Trigger**: `POST /posts`

```mermaid
sequenceDiagram
    participant User as Utente A
    participant Post as Post Service
    participant Kafka
    participant Msg as Messaging/Feed
    participant Followers as Follower B, C..

    User->>Post: Crea Post
    Post->>Post: Salva su MongoDB
    Post->>Kafka: Pubblica `PostCreated`

    Kafka->>Msg: Consume `PostCreated`
    Msg->>Msg: Recupera Follower ID di A (da Neo4j/Cache)

    loop Per ogni Follower
        Msg->>Msg: Inserisce PostID in Timeline (Cassandra)
    end

    Note right of Msg: Ora il feed dei follower è aggiornato
```

---

## 3. Messaggistica Real-Time

Flusso ibrido Persistenza + Push Notification.

**Trigger**: Utente A invia messaggio a Utente B.

```mermaid
sequenceDiagram
    participant Sender
    participant Gateway
    participant MsgService
    participant Cassandra
    participant RedisPubSub
    participant Recipient

    Sender->>Gateway: POST /messages (gRPC)
    Gateway->>MsgService: Invia Messaggio

    par Persistenza
        MsgService->>Cassandra: INSERT into messages
    and Real-Time Delivery
        MsgService->>RedisPubSub: PUBLISH channel:user:B
    end

    RedisPubSub->>Gateway: Evento (SSE/WS)
    Gateway-->>Recipient: Push Messaggio JSON
```

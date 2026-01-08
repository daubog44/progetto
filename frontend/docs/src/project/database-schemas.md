# 🗄️ Schemi Database

Ogni microservizio gestisce il proprio database, garantendo l'isolamento dei dati e la possibilità di scegliere la tecnologia più adatta (Polyglot Persistence).

---

## 🔐 Auth Service (PostgreSQL)

Gestisce i dati anagrafici essenziali e le credenziali.

### Table: `users`

| Colonna        | Tipo            | Vincoli              | Descrizione                                           |
| :------------- | :-------------- | :------------------- | :---------------------------------------------------- |
| `id`           | `SERIAL` (uint) | **PK**               | Identificativo univoco interno.                       |
| `username`     | `VARCHAR`       | **UNIQUE**, Not Null | Nome utente pubblico.                                 |
| `email`        | `VARCHAR`       | **UNIQUE**, Not Null | Email per login e notifiche.                          |
| `password`     | `VARCHAR`       | Not Null             | Hash della password (Argon2/Bcrypt).                  |
| `role`         | `VARCHAR`       | Default `'user'`     | Ruolo per RBAC (`user`, `admin`).                     |
| `created_at`   | `TIMESTAMP`     |                      | Data registrazione.                                   |
| `updated_at`   | `TIMESTAMP`     |                      | Data ultima modifica.                                 |
| `last_seen_at` | `TIMESTAMP`     | Index                | Ultimo accesso persistente (aggiornato a intervalli). |
| `deleted_at`   | `TIMESTAMP`     | Index                | Supporto per Soft Delete.                             |

---

## 📝 Post Service (MongoDB)

Gestisce i contenuti generati dagli utenti e le interazioni testuali.

### Collection: `posts`

```json
{
  "_id": "ObjectId('653a...123')",
  "author_id": "uuid-string-123",
  "content": "Testo del post...",
  "media_urls": ["https://cdn.vibely/img1.jpg", "https://cdn.vibely/video.mp4"],
  "likes_count": 42,
  "comments_count": 5,
  "created_at": "ISODate('2023-10-27T...')"
}
```

_Indexes_:

- `author_id`: 1 (Per profilo utente)
- `created_at`: -1 (Per feed cronologico fallback)

### Collection: `comments`

Ottimizzata per letture veloci sotto i post.

```json
{
  "_id": "ObjectId('653b...456')",
  "post_id": "ObjectId('653a...123')",
  "author_id": "uuid-string-456",
  "content": "Bel post!",
  "created_at": "ISODate('2023-10-27T...')"
}
```

_Indexes_:

- `post_id`: 1, `created_at`: 1 (Per caricare i commenti di un post in ordine)

### Collection: `likes` (Opzionale / MVP)

Per tenere traccia di _chi_ ha messo like (per evitare doppi like e mostrare "Like attivi").

```json
{
  "_id": "ObjectId(...)",
  "post_id": "ObjectId(...)",
  "user_id": "uuid-string",
  "created_at": "ISODate(...)"
}
```

_Indexes_:

- `post_id`: 1, `user_id`: 1 (Unique) -> Garantisce un solo like per utente per post.

### Collection: `user_activity`

Dati operativi ad alta frequenza di scrittura (separati da Postgres per evitare locking/bloat).

```json
{
  "_id": "uuid-string-dell-utente", // L'ID del documento è l'ID Utente stesso (Lookup O(1))
  "last_seen_at": "ISODate(...)",
  "feed_cursor": "ISODate(...)", // Timestamp dell'ultimo post visto nel feed (per riprendere lo scroll)
  "preferences": {
    // Dati "caldi" di UI/UX
    "theme": "dark",
    "auto_play_videos": true,
    "notifications_muted": false
  }
}
```

---

## 🌐 Social Service (Neo4j)

Modella le relazioni sociali come un grafo.

### Node Label: `Person`

| Proprietà    | Tipo       | Descrizione                                    |
| :----------- | :--------- | :--------------------------------------------- |
| `id`         | `String`   | UUID dell'utente (Allineato con Auth Service). |
| `username`   | `String`   | Snapshot del username per display veloce.      |
| `email`      | `String`   | Email utente.                                  |
| `created_at` | `DateTime` | Data creazione nodo.                           |

### Node Label: `Post`

Node "ombra" per grafo sociale (contenuto principale su Mongo).

| Proprietà    | Tipo       | Descrizione                |
| :----------- | :--------- | :------------------------- |
| `id`         | `String`   | ObjectId (Mongo) del post. |
| `created_at` | `DateTime` | Data creazione.            |
| `author_id`  | `String`   | UUID dell'autore.          |

### Relationship: `FOLLOWS`

Rappresenta un utente che ne segue un altro.

`(:Person {id: "A"})-[:FOLLOWS {created_at: DateTime()}]->(:Person {id: "B"})`

### Relationship: `LIKED`

Rappresenta un utente che ha messo "Mi piace" a un post.
_Utile per raccomandazioni ("A chi piace questo post piace anche...")_

`(:Person {id: "A"})-[:LIKED {created_at: DateTime()}]->(:Post {id: "ObjectId..."})`

---

## 💬 Messaging & Feed Service (Cassandra)

Gestisce grandi volumi di dati time-series (chat e timeline).

### Keyspace: `vibely` (Replication Factor: 1 in dev)

### Table: `timeline` (Feed Fan-out)

Contiene la lista pre-calcolata dei post per ogni utente (Home Feed).
_Nota: Denormalizziamo alcuni dati (Preview) per evitare di dover interrogare MongoDB per ogni post nel feed (Read Amplification)._

```cql
CREATE TABLE timeline (
    user_id uuid,          -- L'utente che guarda il feed
    created_at timestamp,  -- Ordinamento
    post_id text,
    author_id text,
    author_username text,  -- Denormalizzato per evitare lookup
    content_preview text,  -- Primi 100 caratteri
    media_url_preview text,-- Prima immagine/video
    PRIMARY KEY (user_id, created_at)
) WITH CLUSTERING ORDER BY (created_at DESC);
```

#### 📌 Strategia Loading del Feed

- **Paginazione**: Usiamo la **Cursor-Based Pagination** basata su `created_at`.
  - _Sconsigliato:_ Offset pagination (`LIMIT 20 OFFSET 100`) -> Lento su grandi volumi.
  - _Consigliato:_ `WHERE created_at < :last_seen_timestamp LIMIT 20`. Molto veloce perché salta direttamente al punto giusto nell'indice SSTable.
- **Marker "Nuovi Post"**:
  - Invece di "Letto/Non Letto" per post, usiamo il timestamp `last_seen_at` dell'utente (persistito su Postgres).
  - Il Client memorizza localmente l'ultimo accesso al feed o lo richiede all'avvio.
  - Quando carica il feed, inserisce una barra "Hai visto tutto fino a qui" visiva se `post.created_at < local_last_seen`.
  - Questo approccio e **Zero-Write** sul database per la lettura.

### Table: `messages` (Chat)

Schema ottimizzato per recuperare la cronologia di una chat.

```cql
CREATE TABLE messages (
    chat_id uuid,
    bucket int,          -- Time bucket (es. mese) per partizionamento enormi chat
    message_id timeuuid, -- Garantisce unicità e ordinamento temporale
    author_id text,
    content text,
    created_at timestamp,
    PRIMARY KEY ((chat_id, bucket), message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);
```

---

## ⚡ Redis (Cache & Real-time)

Usato per caching, gestione presenze e contatori veloci.

### Data Structures

| Key Pattern               | Data Structure | Descrizione                                      | TTL                          |
| :------------------------ | :------------- | :----------------------------------------------- | :--------------------------- |
| `user_presence:{user_id}` | String (JSON)  | Stato (online/offline) e Ultimo Accesso.         | 1 settimana (Cache "Recent") |
| `post:{post_id}:likes`    | String (Int)   | Contatore temporaneo like (Write-back su Mongo). | Persistente/Sync             |
| `feed_cache:{user_id}`    | List / ZSet    | Cache "hot" degli ultimi 50 post della Home.     | ~1 ora                       |
| `session:{token}`         | String         | Blacklist token JWT (se logout esplicito).       | Expire Token                 |

---

## 🔍 Search Service (Meilisearch)

Search Engine ottimizzato per la UX (typo tolerance, velocità).

### Index: `users`

Documenti indicizzati per la ricerca globale.

```json
{
  "id": "uuid...",
  "username": "mario.rossi",
  "email": "mario@example.com"
}
```

**Settings**:

- **Searchable Attributes**: `["username", "email"]`
- **Typo Tolerance**: Enabled

# 🗄️ Schemi Database

Ogni microservizio gestisce il proprio database, garantendo l'isolamento dei dati e la possibilità di scegliere la tecnologia più adatta (Polyglot Persistence).

---

## 🔐 Auth Service (PostgreSQL)

Gestisce i dati anagrafici essenziali e le credenziali.

### Table: `users`

| Colonna      | Tipo            | Vincoli              | Descrizione                          |
| :----------- | :-------------- | :------------------- | :----------------------------------- |
| `id`         | `SERIAL` (uint) | **PK**               | Identificativo univoco interno.      |
| `username`   | `VARCHAR`       | **UNIQUE**, Not Null | Nome utente pubblico.                |
| `email`      | `VARCHAR`       | **UNIQUE**, Not Null | Email per login e notifiche.         |
| `password`   | `VARCHAR`       | Not Null             | Hash della password (Argon2/Bcrypt). |
| `role`       | `VARCHAR`       | Default `'user'`     | Ruolo per RBAC (`user`, `admin`).    |
| `created_at` | `TIMESTAMP`     |                      | Data registrazione.                  |
| `updated_at` | `TIMESTAMP`     |                      | Data ultima modifica.                |
| `deleted_at` | `TIMESTAMP`     | Index                | Supporto per Soft Delete.            |

---

## 📝 Post Service (MongoDB)

Gestisce i contenuti generati dagli utenti. MongoDB permette una struttura flessibile per meta-dati e allegati.

### Collection: `posts`

```json
{
  "_id": "ObjectId('...')",
  "author_id": "uuid-string",
  "content": "Testo del post...",
  "media_urls": ["https://cdn.vibely/img1.jpg", "https://cdn.vibely/video.mp4"],
  "likes_count": 42,
  "created_at": "ISODate('2023-10-27T...')"
}
```

### Collection: `comments` (Design)

_Nota: Schema di design per l'MVP, ottimizzato per letture veloci._

```json
{
  "_id": "ObjectId('...')",
  "post_id": "ObjectId('...')",
  "author_id": "uuid-string",
  "content": "Bel post!",
  "created_at": "ISODate('...')"
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

### Relationship: `FOLLOWS`

Rappresenta un utente che ne segue un altro.

`(:Person {id: "A"})-[:FOLLOWS {created_at: DateTime()}]->(:Person {id: "B"})`

---

## 💬 Messaging Service (Cassandra)

Gestisce grandi volumi di messaggi e dati time-series per i feed.

### Keyspace: `vibely` (Replication Factor: 1 in dev)

### Table: `users` (Denormalizzata)

Copia locale dati utente per join veloci in lettura messaggi.

```cql
CREATE TABLE users (
    user_id text PRIMARY KEY,
    email text,
    username text,
    created_at timestamp
);
```

### Table: `messages` (Design)

Schema ottimizzato per recuperare la cronologia di una chat.

```cql
CREATE TABLE messages (
    chat_id uuid,
    bucket int,          -- Time bucket (es. mese) per partizionamento
    message_id timeuuid, -- Garantisce unicità e ordinamento temporale
    author_id text,
    content text,
    created_at timestamp,
    PRIMARY KEY ((chat_id, bucket), message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);
```

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

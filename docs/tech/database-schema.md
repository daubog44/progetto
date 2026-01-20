# Schemi Database Vibely

Questa sezione definisce la struttura dei dati per l'ecosistema Vibely. Ogni database è scelto per ottimizzare uno specifico carico di lavoro (Workload).

---

## 1. Postgres (Auth Service)
**Ruolo**: Master of Record per l'identità sociale. Garantisce integrità e coerenza (ACID).

### Tabella: `users`
| Campo | Tipo | Descrizione |
| :--- | :--- | :--- |
| `id` | UUID (PK) | ID interno generato dal sistema. |
| `clerk_id` | VARCHAR(255) (Unique) | ID fornito dal provider di autenticazione (Clerk). |
| `email` | VARCHAR(255) (Unique) | Indirizzo email verificato. |
| `display_name` | VARCHAR(50) | Nome visualizzato nella piattaforma. |
| `avatar_url` | TEXT | Link all'immagine del profilo tramite Minio. |
| `created_at` | TIMESTAMP | Data di registrazione. |
| `updated_at` | TIMESTAMP | Ultima modifica del profilo. |

---

## 2. MongoDB (Cultural Service)
**Ruolo**: Gestione dei metadati flessibili per opere culturali e recensioni.

### Collezione: `metadata` (Opere)
```json
{
  "_id": "ObjectId",
  "type": "book | movie | music | tv_series",
  "title": "Il Signore degli Anelli",
  "author_director": "J.R.R. Tolkien",
  "genres": ["Fantasy", "Adventure"],
  "details": {
    "pages": 1178,
    "isbn": "123456789",
    "release_date": "1954-07-29"
  },
  "tags": ["epic", "ring", "frodo"]
}
```

### Collezione: `reviews`
```json
{
  "_id": "ObjectId",
  "user_id": "UUID",
  "content_id": "ObjectId",
  "rating": 5,
  "comment": "Un capolavoro assoluto.",
  "has_spoilers": true,
  "created_at": "ISO-Date",
  "stats": {
    "likes": 12,
    "comments_count": 1500
  }
}
```

### Collezione: `comment_buckets` (Bucket Pattern)
Invece di avere un documento per ogni commento (che creerebbe milioni di documenti e indici enormi), raggruppiamo i commenti in "bucket".
```json
{
  "_id": "ObjectId",
  "review_id": "ObjectId",
  "bucket_index": 0,
  "count": 50,
  "comments": [
    {
      "user_id": "UUID",
      "text": "Concordo pienamente!",
      "created_at": "ISO-Date"
    },
    ... (fino a 50 o 100 commenti)
  ]
}
```
*Vantaggi*: Caricamento di 50 commenti con una singola operazione di I/O, indici più piccoli, performance di paginazione costanti.

---

## 3. Neo4j (Social Graph)
**Ruolo**: Gestione delle relazioni complesse e dei grafi di interesse.

### Nodi e Relazioni
- **Node**: `User {id: UUID, display_name: string}`
- **Node**: `Genre {name: string}`
- **Relation**: `(:User)-[:FOLLOWS]->(:User)`
- **Relation**: `(:User)-[:INTERESTED_IN]->(:Genre)`
- **Relation**: `(:User)-[:READ|WATCHED|LISTENED]->(:Content {id: ObjectId})`

---

## 4. Cassandra (Feed & Messaging Service)
**Ruolo**: Persistenza di dati ad alto volume e serie temporali.

### Tabella: `feeds` (Activity Stream)
- **Partition Key**: `user_id`
- **Clustering Column**: `created_at` (DESC)
```sql
CREATE TABLE feeds (
    user_id uuid,
    created_at timeuuid,
    actor_id uuid,
    action_type text, -- 'review', 'follow', 'achievement'
    content_id text,
    payload text, -- JSON short summary
    PRIMARY KEY (user_id, created_at)
) WITH CLUSTERING ORDER BY (created_at DESC);
```

### Tabella: `messages` (Chat History)
- **Partition Key**: `conversation_id`
- **Clustering Column**: `message_id` (TimeUUID)
```sql
CREATE TABLE messages (
    conversation_id uuid,
    message_id timeuuid,
    sender_id uuid,
    content text,
    is_spoiler boolean,
    PRIMARY KEY (conversation_id, message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);
```

---

## 5. Redis (Real-time & Cache)
**Ruolo**: Latenza minima per stati volatili e notifiche push.

- **Presence**: `user:online:{user_id}` (key con TTL)
- **Real-time Chat**: `pubsub:chat:{conversation_id}`
- **Cache Metadati**: `content:meta:{content_id}`
- **Anti-Spoiler**: Set di utenti che hanno completato un'opera: `spoiler:safe:{content_id}`

---

## 6. Strategia per Migliaia di Commenti

Per gestire migliaia di commenti senza degradare le performance, Vibely adotta le seguenti tecniche:

### Bucket Pattern (MongoDB)
I commenti sono salvati in gruppi (es. 50-100 per documento). 
- **Lettura**: Con una singola query recuperiamo un'intera pagina di commenti.
- **Scrittura**: Usiamo `$push` con `$slice` per mantenere i bucket entro una dimensione ottimale.

### Paginazione via Cursore
Non usiamo `offset` (che diventerebbe lento con l'aumentare dei dati). Usiamo una paginazione basata su cursore:
1. Il client chiede i commenti per una recensione.
2. Il server restituisce il primo bucket e un `next_cursor` (il `bucket_index` successivo o l'ID dell'ultimo commento).
3. Il client usa il cursore per la richiesta successiva (Infinite Scroll).

### Denormalizzazione dei Contatori
Il numero totale dei commenti è salvato direttamente nel documento della Recensione (`stats.comments_count`). Questo valore viene aggiornato in modo asincrono tramite Kafka ad ogni nuovo commento, garantendo che il feed possa mostrare il contatore istantaneamente senza contare fisicamente i documenti.

### Virtualizzazione Frontend
Sul lato React/Mobile, utilizziamo la **Virtualizzazione delle Liste**. Solo i commenti visibili nel viewport vengono effettivamente renderizzati nel DOM, mantenendo la memoria fluida anche con 10.000+ righe caricate.

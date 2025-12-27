
# 🗄 Modelli Dati (Data Models)

Documentazione degli schema utilizzati dai servizi attivi.

---

## 1. Postgres (Auth Service)
**Ruolo**: Identity Provider, Master of Record per gli utenti.

### Tabella: `users`
| Campo | Tipo | Descrizione |
| :--- | :--- | :--- |
| `id` | UUID (PK) | Identificativo univoco. |
| `email` | VARCHAR | Univoco, indicizzato. |
| `password` | VARCHAR | Hash bcrypt. |
| `username` | VARCHAR | Univoco. |
| `role` | VARCHAR | `admin` o `user`. |
| `created_at` | TIMESTAMP | |

---

## 2. MongoDB (Post Service)
**Ruolo**: Contenuti generati dagli utenti (UGC).

### Collezione: `posts`
```json
{
  "_id": "ObjectId",
  "author_id": "string (UUID)",
  "content": "string",
  "media_urls": ["string"],
  "created_at": "ISO Date"
}
```

### Collezione: `users` (Replica di Lettura)
Copia locale minima degli utenti per evitare join distribuiti durante il rendering del feed.
```json
{
  "_id": "string (UUID)",
  "username": "string",
  "email": "string"
}
```

---

## 3. Redis (Auth Service / Gateway)
**Ruolo**: Stati volatili e Token.

- **Refresh Tokens**: `refresh:{token}` -> `user_id` (TTL: 6 mesi)

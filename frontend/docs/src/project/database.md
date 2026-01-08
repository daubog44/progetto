# Database e Schemi

In accordo con l'architettura "Polyglot Persistence" discussa nel riferimento _05-BasiDati-NoSQL_, utilizziamo tecnologie di storage specializzate.

## MongoDB (Document Store)

Utilizzato per la sua flessibilità schematica (**Schema-less**).

- **Teoria**: Gestisce dati semi-strutturati (BSON) e supporta query complesse tramite Aggregation Pipeline.
- **Uso**: Profili utente, Log applicativi (`users`, `notifications`, `logs`).
- **Cluster**: Configurabile come Replica Set per alta disponibilità.

## Neo4j (Graph Database)

Scelto per gestire dati altamente interconnessi (**Property Graph Model**).

- **Teoria**: I nodi e le relazioni sono "cittadini di prima classe". Risolve il problema del "traversal" che nei DB relazionali richiede JOIN costose.
- **Uso**: Social Graph (Amici, Follower).
- **Entità**: `User`, `Group`, `Post` collegati da relazioni `FOLLOWS`, `LIKED`.

## Cassandra (Wide-Column Store)

Database distribuito Peer-to-Peer ottimizzato per **Scritture Massive** e scalabilità lineare.

- **Teoria**: Architettura "Masterless" (Ring) che garantisce Availability e Partition Tolerance (AP nel CAP theorem).
- **Uso**: Timeline, Chat messaggi, Audit logs.
- **Modello Dati**: Definito dalle query (Query-driven modelling).

## Redis (Key-Value Store)

In-memory store per bassa latenza.

- **Uso**: Caching, Session Management, Pub/Sub.

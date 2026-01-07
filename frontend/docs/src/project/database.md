# Database e Schemi

Il progetto utilizza diversi database specializzati per ottimizzare le performance in base al tipo di dato.

## MongoDB

Utilizzato come documento store principale per dati non strutturati o semi-strutturati (es. profili utente, log applicativi).

- **Collezioni**: `users`, `notifications`, `logs`.
- **Vantaggi**: Schema-less, alta velocità di scrittura.

## Neo4j

Graph Database utilizzato per gestire le relazioni sociali e connessioni complesse tra entità.

- **Nodi**: `User`, `Group`, `Post`.
- **Relazioni**: `FOLLOWS`, `MEMBER_OF`, `LIKED`.
- **Vantaggi**: Performance ottimali per query su grafi (es. "amici degli amici").

## Cassandra

Column-family store per dati ad alta cardinalità e serie temporali (es. messaggi di chat, eventi audit).

- **Tabelle**: `messages_by_room`, `user_timeline`.
- **Vantaggi**: Scrittura veloce, scalabilità lineare.

## Redis

In-memory store utilizzato per:

- **Caching**: Risultati query frequenti.
- **Session Management**: Token sessione.
- **Rate Limiting**: Contatori richieste.
- **Pub/Sub**: Comunicazione real-time leggera.

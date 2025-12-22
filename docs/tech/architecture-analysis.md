Questa analisi identifica i potenziali punti di fallimento, i limiti di scalabilità e i colli di bottiglia dell'attuale architettura a microservizi di Vibely. Per le strategie adottate per mitigare questi rischi, consulta [Sicurezza, Performance e Resilienza](security-performance-resilience.md).

---

## 1. Data Layer & Polyglot Persistence
Nonostante l'uso del database corretto per ogni workload sia un punto di forza, introduce complessità:

### ⚠️ Consistenza Eventuale (Eventual Consistency)
Poiché usiamo Kafka per propagare i dati da Postgres (Auth) agli altri database (Mongo, Neo4j, Cassandra), esiste un ritardo intrinseco.
- **Rischio**: Un utente registra un account e prova subito a creare una recensione. Se `mongo-service` non ha ancora ricevuto l'evento di creazione utente, l'operazione gRPC potrebbe fallire.
- **Soluzione**: Implementare logiche di **retry** sul client o usare **Optimistic UI** che gestisce lo stato di "In attesa di sincronizzazione".

### ⚠️ Gestione delle Transazioni Distribuite
Non abbiamo un sistema di transazioni atomiche che copra più database.
- **Bottleneck**: Se un'operazione richiede la scrittura coordinata su Mongo e Neo4j (es. "Segui una Community e aggiorna il grafo"), un fallimento parziale può lasciare i dati inconsistenti.
- **Mitigazione**: Utilizzo del **Saga Pattern** con azioni di compensazione gestite da Watermill.

---

## 2. Microservices & Network
La comunicazione tra servizi è il cuore del sistema, ma può diventare un limite.

### ⚠️ Network Latency (Chiamate a sbalzo)
Le chiamate gRPC sincrone aggiungono latenza. Se il Gateway chiama il `cultural-service`, che a sua volta chiama l' `auth-service` per validare un permesso, la latenza si somma.
- **Collo di bottiglia**: L' `auth-service` può diventare il "single point of failure" o il principale rallentamento per ogni richiesta.
- **Soluzione**: Cache dei permessi/sessioni in **Redis** per ridurre le chiamate inter-servizio gRPC.

### ⚠️ Overhead di Serializzazione
Sotto carichi estremi (milioni di messaggi/sec), la serializzazione/deserializzazione Protobuf e la gestione dei buffer Kafka possono saturare la CPU dei microservizi Go.
- **Monitoraggio**: Necessario tracciamento costante tramite OpenTelemetry per identificare quale servizio ha l'overhead maggiore.

---

## 3. Monorepo & Shared Dependencies
### ⚠️ Il collo di bottiglia del file `shared/proto`
Tutti i microservizi dipendono dai file `.proto` generati in `shared/`.
- **Rischio**: Una modifica a un contratto Protobuf richiede la rigenerazione e il redeploy di potenzialmente tutti i servizi (effetto "Blast Radius").
- **Soluzione**: Utilizzo rigoroso del versioning dei pacchetti Protobuf (`v1`, `v2`) per garantire la compatibilità all'indietro.

---

## 4. Scalabilità degli Archivi (Data Storage)
### ⚠️ Neo4j e i Super-Nodi
Utenti "VIP" con milioni di follower creano dei "Super-Nodi" nel grafo Neo4j.
- **Bottleneck**: Query di attraversamento che coinvolgono questi nodi possono diventare estremamente lente.
- **Soluzione**: Sharding del grafo o utilizzo di cache specializzate per i follower di utenti popolari.

### ⚠️ Cassandra Tombstones
L'eliminazione massiva di messaggi o post in Cassandra crea "tombstones" che rallentano drasticamente le letture.
- **Rischio**: Operazioni di cancellazione frequenti degradano le performance del Feed.
- **Soluzione**: Design del data model che minimizzi le cancellazioni o utilizzi TTL (Time-To-Live) per la scadenza automatica.

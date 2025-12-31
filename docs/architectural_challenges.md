# Sfide Architetturali e di Codice Superate

Questo documento raccoglie le principali sfide tecniche affrontate durante lo sviluppo dell'architettura a microservizi, evidenziando le soluzioni adottate per problemi complessi di scalabilità, manutenibilità e osservabilità.

## 1. Architettura & Comunicazione

### Flusso di Notifiche Real-Time (Complex Event Processing)
Una delle sfide principali è stata orchestrare la notifica push verso il client in un ambiente distribuito dove l'utente può essere connesso a una qualsiasi istanza scalata dell'API Gateway.

*   **Problema**: Quando un evento accade (es. "post creato"), il servizio responsabile pubblica un evento su Kafka. Il `gateway-service` che mantiene la connessione SSE con l'utente potrebbe non essere lo stesso che ha originato la richiesta o non avere accesso diretto a Kafka.
*   **Soluzione**: Abbiamo implementato un pattern a doppio hop Pub/Sub:
    1.  **Kafka (Source of Truth)**: I servizi (es. `post-service`) pubblicano eventi di dominio su Kafka.
    2.  **Notification Service (Orchestrator)**: Consuma gli eventi da Kafka. Verifica la presenza dell'utente (online/offline) e identifica l'ID specifico dell'istanza del Gateway a cui l'utente è connesso (usando Redis).
    3.  **Redis Pub/Sub (Targeted Delivery)**: Il `notification-service` pubblica il messaggio su un canale Redis specifico per quell'istanza del Gateway (`gateway_events:<instance_id>`).
    4.  **API Gateway (SSE)**: Ogni istanza del Gateway sottoscrive solo al proprio canale Redis. Quando riceve un messaggio, lo inoltra via Server-Sent Events (SSE) allo specifico client connesso.

### Migrazioni del Database Separate
Per garantire l'idempotenza e il versionamento sicuro del database (in particolare Cassandra), abbiamo disaccoppiato la logica di migrazione dal ciclo di vita dell'applicazione.
*   **Implementazione**: Utilizziamo container dedicati (es. `messaging-migration`) che eseguono script di inizializzazione e migrazione all'avvio dell'infrastruttura, prima che i servizi dipendenti diventino "healthy". Questo evita race conditions durante lo scaling orizzontale dei servizi.

### Modelli di Dati Condivisi (Shared Library)
Per evitare duplicazione di codice e disallineamenti nei DTO e modelli di dominio trasversali.
*   **Soluzione**: Creazione di una `shared library` (`shared/pkg`) che centralizza:
    *   Definizioni dei modelli (es. `model/post.go`).
    *   Configurazioni di `watermill`, `grpc`, e `observability`.
    *   Gestione degli errori e middleware.

## 2. Codice & Affidabilità

### Error Handling & Tracing Uniforme
Gestire gli errori in modo coerente attraverso gRPC, HTTP e code asincrone.
*   **Approccio**:
    *   Mapping centralizzato degli errori gRPC verso codici HTTP.
    *   **LGTM Stack Integration**: Ogni errore tracciato porta con sé il `TraceID` e `SpanID`, permettendo di correlare un errore 500 nel frontend fino alla query database fallita nel microservizio di backend, visualizzando il tutto su Grafana (Tempo & Loki).

### Dockerfile Optimization & Caching
Ottimizzazione drastica dei tempi di build e pull delle immagini in CI/CD.
*   **Multi-Stage Build**: Separazione netta tra stage di `builder`, `dev` e `production` (Distroless).
*   **Module Caching**: Copia dei file `go.mod` e `go.sum` (inclusi quelli della cartella `shared`) e download delle dipendenze *prima* di copiare il codice sorgente. Questo permette a Docker di riutilizzare i layer di cache delle dipendenze se il codice cambia ma le librerie no.

### Dead Letter Queue (DLQ) & Polson Messages (TODO Integration)
Gestione dei messaggi che falliscono permanentemente l'elaborazione (poison messages) per evitare loop infiniti di retry.
*   **Stato**: Integrazione con Watermill Router configurata per redirigere i messaggi falliti dopo N tentativi su un topic `dead_letters`. La logica è presente in `watermillutil/router.go` ma richiede test approfonditi di recupero/analisi manuale.

## 3. Scalabilità & Gestione Stato

### Online/Offline Presence System
Monitorare lo stato di connessione degli utenti in tempo reale.
*   **Implementazione**: Uso di Redis con chiavi a scadenza (TTL).
    *   Il Gateway imposta `user_presence:<id>` con TTL breve.
    *   Un heartbeat periodico dal client/gateway rinnova il TTL.
    *   **TODO**: Persistenza su database non volatile (es. Postgres/Cassandra) per storicizzare le sessioni utente, dato che Redis è in-memory e volatile.

### Scripting & Automation
Gestione rapida dell'ecosistema di microservizi.
*   **Tools**: Script Bash (`add-service.sh`, `remove-service.sh`, `docker-cleanup.sh`) per:
    *   Scaffolding automatico di nuovi servizi con boilerplate standard (gRPC, OTel, Dockerfile, ecc.).
    *   Pulizia sicura di volumi e container orfani.

## 4. Osservabilità (The "LGTM" Stack)

Abbiamo adottato lo stack Grafana completo per una visibilità a 360°:
*   **Loki**: Log aggregation centralizzata.
*   **Grafana**: Dashboard unificate (es. `overview.json`, `services.json`).
*   **Tempo**: Distributed Tracing per visualizzare la latenza end-to-end tra microservizi (Kafka -> Handler -> Redis -> SSE).
*   **Mimir/Prometheus**: Metriche per CPU, Memoria e throughput (RPS).

## 5. Security

### Refresh Token & Validation
Gestione sicura delle sessioni stateless.
*   **Logica**: Implementazione della validazione JWT sia a livello di Middleware HTTP che nell'handshake iniziale SSE. La logica `jwtutil` condivisa assicura che un token scaduto venga rifiutato con un errore specifico, permettendo al frontend di tentare il refresh trasparentemente.

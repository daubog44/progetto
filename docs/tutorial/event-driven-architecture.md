# Architettura Ibrida: gRPC + Kafka (Watermill)

Questa guida esplora l'integrazione tra comunicazione sincrona (**gRPC**) e asincrona (**Event-Driven**) all'interno del nostro sistema di microservizi.

---

## 🏗️ La Filosofia: Sync per i Comandi, Async per gli Eventi

In un'architettura moderna, non esiste una taglia unica. Utilizziamo i due approcci in modo complementare:

### 1. gRPC (Sincrono - Request/Response)
**Quando si usa:** Per operazioni "critiche" dove l'utente si aspetta un risultato immediato.
- **Pro:** Latenza minima, tipizzazione forte (Protobuf), feedback istantaneo.
- **Contro:** Accoppiamento temporale (entrambi i servizi devono essere online).

### 2. Kafka + Watermill (Asincrono - Pub/Sub)
**Quando si usa:** Per propagare cambiamenti, gestire effetti collaterali e garantire la consistenza eventuale.
- **Pro:** Disaccoppiamento totale, resilienza (Kafka salva i messaggi se un servizio è giù), scalabilità.
- **Contro:** Consistenza eventuale (i dati potrebbero metterci qualche millisecondo ad aggiornarsi ovunque).

### 3. CQRS (Command Query Responsibility Segregation)
Il nostro sistema implementa naturalmente il pattern **CQRS**. Invece di avere un unico modello dati per lettura e scrittura, separiamo le responsabilità:
- **Command Side (Scrittura)**: Microservizi come `mongo-service` (per i metadati dei libri/film) o `auth-service` gestiscono le modifiche ai dati tramite **gRPC**. Emettono eventi su **Kafka** per notificare i cambiamenti.
- **Query Side (Lettura)**: Microservizi come `cassandra-service` (per il Feed delle recensioni) o un ipotetico `search-service` mantengono viste ottimizzate per la lettura, aggiornate asincronamente tramite gli eventi Kafka.

Questo ci permette di scalare le letture (Feed) indipendentemente dalle scritture (Recensioni) e di usare il Database più adatto a ogni scopo (es. Neo4j per i grafi di interesse, Cassandra per i feed storici).

---

## ⚙️ Lo Stack Tecnologico

### Kafka in KRaft Mode
Utilizziamo Kafka in **KRaft mode** (Kafka Raft Metadata mode). Questo significa che Kafka gestisce i propri metadati senza bisogno di Zookeeper, rendendo l'infrastruttura più semplice, veloce e moderna.

### gRPC & Protobuf (Buf v2)
Utilizziamo **gRPC** come protocollo di comunicazione principale per le operazioni sincrone.
- **Protobuf**: Definiamo i contratti in file `.proto` (es. `shared/proto/data/v1/data.proto`).
- **Buf v2**: Gestisce la linting, la generazione del codice e la compatibilità dei contratti. Usiamo il **Managed Mode** per garantire che i package Go siano sempre coerenti.

### Watermill: L'astrazione Go
[Watermill](https://watermill.io/) è una libreria Go che semplifica il lavoro con i database di messaggi. Funge da "gRPC per gli eventi":
- Consente di cambiare il backend (da Kafka a RabbitMQ o Redis) senza toccare la logica di business.
- Offre Middleware per: **Retry automatici**, **Logging**, **Metrics** e **Throttle control**.

---

### Polyglot Persistence: Quale DB per quale scopo?

Non esiste un "Database Principale" unico. Scegliamo la tecnologia in base al **carico di lavoro (Workload)**:

- **Auth/Identity (Postgres)**: Usiamo Postgres perché serve una coerenza (ACID) assoluta sulle credenziali. È il punto di partenza (Master) per ogni utente.
- **Cultura & Metadata (MongoDB)**: Per entità con strutture flessibili (es. un Libro può avere campi diversi da un Film, come numero di pagine vs durata), MongoDB è l'ideale. Gestisce bene anche la nidificazione dei generi e tags.
- **Social Graph & Interessi (Neo4j)**: Se vuoi sapere "chi ha gusti musicali simili" o "amici che hanno visto questo film", Neo4j vince. Una query di 3 livelli di profondità (gusti degli amici degli amici) che in SQL richiederebbe JOIN pesanti, in Neo4j è istantanea.
- **Feed & Timeline Culturali (Cassandra)**: Cassandra è perfetto per serie temporali (time-series). Quando apri il Feed delle attività degli amici, vuoi gli ultimi aggiornamenti. Cassandra scrive velocissimo e legge per "range di tempo" in modo imbattibile.
- **Messaggistica (Cassandra + Redis)**:
    - **Cassandra**: Per lo storico dei messaggi (persisenza a lungo termine, ordinata per tempo).
    - **Redis**: Per i messaggi in tempo reale (pub/sub), lo stato "online/offline" degli utenti e la cache dei metadati più letti per azzerare la latenza.

**Chi comanda?** L' `auth-service` è il **Master of Record** (Sorgente di Verità) per l'identità. Gli altri DB sono "Read Models" o "Viste specializzate" che vengono aggiornate asincronamente via Kafka quando i dati cambiano.

---

## 🔄 Workflow di Esempio

### 👤 Workflow: Creazione Utente
1.  **gRPC Call**: Il client invia i dati di registrazione al microservizio `auth-service`.
2.  **Azione Sincrona**: Il servizio valida i dati e scrive sul database.
3.  **Risposta gRPC**: L'utente riceve subito un "OK, account creato".
4.  **Evento Kafka**: `auth-service` pubblica un messaggio `UserRegistered` su Kafka tramite Watermill.
5.  **Effetti Asincroni**:
    - `mongo-service` riceve l'evento e crea il profilo base.
    - `email-service` invia la mail di benvenuto.

### � Workflow: Aggiunta Recensione
1.  **gRPC Call**: L'utente invia una recensione di un libro al `mongo-service`.
2.  **Azione Sincrona**: La recensione viene salvata nel DB e l'utente riceve "Recensione pubblicata con successo".
3.  **Evento Kafka**: `mongo-service` pubblica `ReviewCreated` su Kafka.
4.  **Effetti Asincroni**:
    - `cassandra-service` riceve l'evento e aggiorna i Feed di tutti gli amici dell'utente interessati a quel genere.
    - `neo4j-service` aggiorna i grafi di interesse dell'utente in base al voto dato al libro.

### 📜 Workflow: Lettura del Feed
1.  **gRPC Call**: Il client richiede il feed al `cassandra-service`.
2.  **Azione Sincrona**: Il servizio legge i dati pre-calcolati da Cassandra e li restituisce.
3.  *Nessun evento Kafka necessario qui*, a meno che non si voglia tracciare che un post è stato visualizzato (Analitiche asincrone).

### 🔍 Deep Dive: Registrazione Utente Full-Async

**Il Workflow Ibrido (Consigliato):**
1. **gRPC (Sincrono)**: Il Gateway chiama `auth-service`. Questo valida e scrive nel DB principale (Postgres/Mongo). Il Gateway aspetta la conferma e restituisce 200 all'utente.
2. **Kafka (Asincrono)**: Una volta confermato l'utente nel DB, **l'auth-service** pubblica gli eventi su Kafka per i "Data Services" (Feed, Grafi, Profile Optimization). Se questi falliscono, l'account c'è comunque e la Saga gestisce il resto.

---

## 🔔 Come notificare il Frontend?

Se un'operazione asincrona (es. generazione di un report o setup complesso) fallisce o finisce, come lo diciamo al browser?

1. **WebSockets (Push)**: Il database di notifiche ascolta Kafka e invia un messaggio tramite un server WebSocket (es. Centrifugo) direttamente al client.
2. **Long Polling (Pull)**: Il frontend, ricevuto il 200, interroga ogni 2 secondi un endpoint `GET /status/{job_id}` finché non è pronto.
3. **Optimistic UI con Recovery**: Il frontend mostra subito l'azione come fatta. Se riceve un errore via WebSocket/SSE, fa il "rollback" visivo e mostra un toast di errore.

---

## 🛡️ Gestione degli Errori e Consistenza

In un sistema distribuito, non possiamo usare transazioni del database (ACID) che coprono più microservizi. Ecco come risolviamo:

### 1. Il Problema del "Dual Write" (Transactional Outbox Pattern)
**Esempio:** Salvi l'utente nel DB ma il broker Kafka è giù proprio in quel secondo. Il messaggio è perso?
**Soluzione:** Non pubblicare direttamente su Kafka. Salva il messaggio in una tabella `Outbox` nello stesso database dell'utente, usando una transazione locale. Un processo separato legge dalla tabella e invia a Kafka.

### 2. Resilienza con Watermill (Retry & Poison Queue)
Watermill permette di configurare dei middleware per gestire i fallimenti dei consumer.

```go
// Esempio di configurazione Middleware in Go
router.AddMiddleware(
    middleware.Recoverer,
    middleware.Retry{
        MaxRetries:      3,
        InitialInterval: time.Millisecond * 100,
    }.Middleware,
    // Se fallisce dopo 3 retry, manda il messaggio in una "Poison Queue" (DLQ)
    middleware.PoisonQueue(publisher, "poison_topic").Middleware,
)
```

### 3. Saga Pattern (Compensazione)
Se il `cassandra-service` fallisce definitivamente nel creare il feed, deve avvisare il sistema. Non puoi "fare rollback" del DB Mongo dall'esterno. Invece, invii un evento `FeedCreationFailed` e il `mongo-service` (che lo ascolta) esegue una **Azione di Compensazione** (es. marca il profilo come incompleto o invia una notifica).

---

## 🎨 Strategie Frontend (UX)

Cosa vede l'utente se l'azione è asincrona?

### 1. Optimistic UI
Il frontend "assume" che tutto andrà bene. Se premo "Segui", il bottone diventa subito "Seguito". Se il backend fallisce asincronamente (via WebSocket), il frontend riporta lo stato a "Segui" con un errore.

### 2. Polling o WebSockets
Se la creazione del profilo è lenta:
- **Polling:** Il frontend chiede ogni 2 secondi `GET /profile`.
- **WebSockets/SSE:** Il server "spinge" la conferma al frontend appena l'evento Kafka viene consumato con successo.

---

## 📌 Regola d'Oro
Usa **gRPC** per le **Query** (letture) e per l'inizio di una **Command** (scrittura). Usa **Kafka** per tutto ciò che deve succedere **dopo** che il comando iniziale ha avuto successo. Se l'azione successiva è fondamentale, implementa il **Saga Pattern** per gestire i fallimenti con azioni di compensazione.

# 🛡️ Sicurezza, Performance e Resilienza

In Vibely, l'eccellenza tecnica non è un optional, ma la fondazione su cui poggia l'intera esperienza utente. Questo documento dettaglia le scelte architettoniche che garantiscono un sistema sicuro, veloce e robusto.

---

## 🔒 Sicurezza: "Secure by Design"

La nostra strategia di sicurezza si concentra sulla riduzione della superficie di attacco e sulla protezione dei dati in transito e a riposo.

### 1. Distroless Docker Images
Utilizziamo immagini **Distroless** (`gcr.io/distroless/static-debian12`) per il runtime dei nostri microservizi Go.
- **Perché**: Queste immagini contengono solo il binario dell'applicazione e le sue dipendenze minime. Non includono shell, package manager o altre utility Unix comuni.
- **Vantaggio**: Una vulnerabilità in una library non può essere sfruttata per ottenere una shell o scalare i privilegi, poiché gli strumenti necessari non esistono nell'immagine.

### 2. Comunicazione gRPC Protetta
Tutte le chiamate inter-service avvengono tramite **gRPC**.
- **Contratti Rigorosi**: gRPC utilizza Protobuf, che impone una struttura dati rigida, prevenendo attacchi basati su payload malformati (es. injection comuni nelle API REST).
- **Zero Trust**: (Work in Progress) Prevediamo l'implementazione di mTLS (Mutual TLS) per garantire che ogni servizio sia autenticato prima di poter parlare con un altro.

### 3. Servizi Esterni e Credentials
- **Identity Management**: Utilizziamo PostgreSQL come Master of Record per i dati sensibili, garantendo integrità referenziale.
- **Encryption**: I segreti sono gestiti tramite variabili d'ambiente (iniettate in produzione via Secret Manager) e non vengono mai salvati nel codice sorgente.

---

## 🚀 Performance: Efficienza allo Stato dell'Arte

Il sistema è ottimizzato per rispondere a milioni di richieste con latenze nell'ordine dei millisecondi.

### 1. Go (Golang) Runtime
Abbiamo scelto Go per la sua gestione efficiente della concorrenza (**Goroutines**) e il suo basso overhead di memoria.
- **Compilazione Nativa**: I binari sono compilati staticamente per Linux, garantendo che non ci siano overhead di interpretazione (come in Python o Node.js).
- **Serializzazione Protobuf**: Protobuf è significativamente più veloce e leggero del JSON, riducendo sia il carico sulla CPU che l'utilizzo della banda di rete.

### 2. Polyglot Persistence
Ogni microservizio usa il database più adatto al suo compito:
- **Cassandra**: Ottimizzata per scritture massive e letture veloci di timeline/feed.
- **Neo4j**: Interroga relazioni sociali complesse in millisecondi, dove un database SQL impiegherebbe secondi (o fallirebbe).
- **Redis**: Fornisce una cache L1 per i dati "hot" (es. token di sessione, metadati frequenti).

### 3. Scalabilità Orizzontale
Ogni servizio è stateless. Questo permette di scalare il numero di repliche istantaneamente in base al traffico rilevato da Kubernetes o strumenti di orchestrazione, senza perdita di stato.

---

## 🏗️ Resilienza: "Design for Failure"

In un sistema distribuito, i guasti sono inevitabili. Vibely è progettato per sopravvivere a guasti parziali.

### 1. Architettura Event-Driven (Kafka)
Utilizziamo **Kafka** per il disaccoppiamento dei servizi.
- **Resistenza ai Picchi**: Se un servizio è lento, Kafka funge da buffer, permettendo al sistema di processare i messaggi non appena le risorse tornano disponibili.
- **Ripristino**: In caso di crash di un consumatore, questo può ripartire dall'ultimo offset salvato senza perdere dati.

### 2. Graceful Degradation
Se il servizio di ricerca (Meilisearch) è temporaneamente offline, l'utente può comunque navigare nel proprio feed e chattare. Le funzionalità principali sono isolate per evitare il "effetto domino".

### 3. Osservabilità (Loki, Prometheus, Grafana)
Non possiamo riparare ciò che non vediamo.
- **Monitoring**: Dashboard in tempo reale monitorano CPU, memoria e latenza gRPC.
- **Alerting**: Notifiche immediate se i tassi di errore superano una certa soglia, permettendo interventi proattivi.

# Strategia di Gestione Errori e Affidabilità Distribuita

Garantire la resilienza in un sistema a microservizi come Vibely richiede un approccio a più livelli. Gli errori non sono eccezioni, ma eventi attesi del ciclo di vita. Questa guida consolida le strategie tecniche e i pattern architetturali per un sistema tollerante ai guasti.

---

## 1. Transazioni Distribuite e Transactional Outbox

### Il Problema del "Dual Write"
In microservizi come `auth-service`, spesso dobbiamo salvare dati su un database (Postgres) e contemporaneamente notificare altri servizi via Kafka.
Non puoi avere una transazione atomica tra un DB e un broker. Se il commit del DB ha successo ma l'invio a Kafka fallisce, il sistema diventa inconsistente.

### La Soluzione: Outbox Pattern
Usiamo il database come buffer temporaneo:
1. **Atomaticità Local**: Nello stesso blocco di transazione SQL, salviamo l'entità (es. Utente) e il messaggio dell'evento in una tabella `outbox`.
2. **Relay Worker**: Un processo separato legge la tabella `outbox` e pubblica su Kafka. Solo dopo la conferma di ricezione da parte di Kafka, il record viene marcato come `PROCESSED`.

**Vantaggi**: Garantisce la consegna "At-least-once" senza perdere eventi in caso di crash del publisher.

---

## 2. Saga Pattern (Gestione Fallimenti Parziali)

In flussi complessi (es. registrazione che aggiorna Postgres, Mongo e Cassandra), un fallimento a metà del flusso richiede una **Saga**.

- **Choreography-based Saga**: Ogni servizio ascolta eventi di successo/fallimento dai vicini.
- **Compensazione**: Se il `cassandra-service` fallisce l'inizializzazione del feed, emette `FeedInitializationFailed`. L'`auth-service` ascolta e "annulla" o marca come "da riparare" l'utente creato inizialmente.

---

## 3. Idempotenza (Prevenire la Duplicazione)

Poiché garantiamo la consegna dei messaggi "almeno una volta", è possibile ricevere lo stesso evento due volte.

### Strategia Backend (Idempotency Key)
- Ogni richiesta/evento porta una **Idempotency-Key** (UUID generato dal mittente).
- Il ricevente salva questa chiave in una tabella `processed_requests` o valida un vincolo `UNIQUE` sul database.
- Se la chiave esiste già, il backend restituisce il risultato memorizzato senza rieseguire la logica.

### Strategia Frontend
- Il frontend genera l'UUID e lo invia negli header gRPC.
- In caso di timeout, il frontend può riprovare in sicurezza usando la stessa chiave.

---

## 4. Errori in Kafka: DLQ e Poison Queues

Se un messaggio Kafka causa un errore sistematico nel consumer:
- **Retry con Backoff**: Usiamo il middleware di **Watermill** per 3-5 tentativi esponenziali.
- **Dead Letter Queue (DLQ)**: Se i retry falliscono, il messaggio viene spostato in un topic `dead_letters`. Questo evita di bloccare l'intero stream di Kafka per un singolo messaggio corrotto.

---

## 5. Outbox su Database NoSQL

Se il database primario non è relazionale:
- **MongoDB**: Usa le **Multi-Document Transactions** per l'outbox oppure i **Change Streams** (ascolto nativo delle modifiche alla collezione).
- **Cassandra**: Data l'assenza di transazioni multi-tabella, si preferisce l'**Event Sourcing** (scrivi prima su Kafka) o l'uso di **Lightweight Transactions (LWT)** per garantire l'ordine.

---

## 6. Recovery UI e Self-Healing

Il frontend deve gestire graficamente gli stati di errore asincrono rilevati via WebSocket:
- **Rollback Visivo**: Se un'azione "ottimistica" (es. Like) fallisce asincronamente, riporta lo stato UI alla versione precedente.
- **Pulsanti di Recovery**: Se il setup del profilo è incompleto, mostra un banner: *"Il tuo profilo è incompleto. [Completa Setup]"* che re-innesca il processo mancante.

---

## 7. Monitoraggio (Tracing & Alerting)

- **Distributed Tracing (OpenTelemetry)**: Utilizziamo un `trace_id` unico che segue il pacchetto attraverso gRPC, Kafka e i vari database.
- **Alerting**: Soglie di errore elevate nelle DLQ o nella tabella Outbox triggerano notifiche immediate al team dev.

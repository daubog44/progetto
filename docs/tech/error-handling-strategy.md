# Strategia di Gestione Errori e Affidabilità Distribuita

Garantire la resilienza in un sistema a microservizi come Vibely richiede un approccio a più livelli. Gli errori non sono eccezioni, ma eventi attesi del ciclo di vita. Questa guida consolida le strategie tecniche e i pattern architetturali per un sistema tollerante ai guasti.

---

## 1. Affidabilità dell'invio (Direct Push & Watermill)

In Vibely, abbiamo scelto di non usare l'Outbox Pattern (salvataggio su DB e polling successivo) per mantenere l'architettura snella e performante. Invece, i servizi pubblicano gli eventi direttamente su Kafka utilizzando **Watermill**.

### Come garantiamo la coerenza?
Poiché non esiste una transazione distribuita tra Database e Kafka, adottiamo queste precauzioni:
1. **Atomaticità Local**: Cerchiamo di rendere l'operazione di business (es. `PostReview`) il più possibile atomica.
2. **Watermill Publisher Retry**: Il publisher di Watermill è configurato con retry automatici in caso di brevi interruzioni di rete con Kafka.
3. **Eventual Consistency**: Accettiamo che, in rari casi catastrofici (crash del server tra la scrittura su DB e l'invio a Kafka), i dati debbano essere riconciliati manualmente o tramite script di audit.
4. **Idempotenza**: Fondamentale per gestire i retry infiniti senza duplicare i dati a valle (vedi sezione 3).

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

## 5. Strategia per Database NoSQL

Quando usiamo database come MongoDB o Cassandra, la strategia rimane la stessa: **Direct Push**.
- **MongoDB**: Usiamo le sessioni per garantire che la scrittura principale avvenga in modo atomico, seguita immediatamente dall'invio dell'evento.
- **Cassandra**: Utilizziamo le **Lightweight Transactions (LWT)** se l'ordine e la concorrenza sono critici prima di inviare l'evento a Kafka.

---

## 6. Recovery UI e Self-Healing

Il frontend deve gestire graficamente gli stati di errore asincrono rilevati via WebSocket:
- **Rollback Visivo**: Se un'azione "ottimistica" (es. Like) fallisce asincronamente, riporta lo stato UI alla versione precedente.
- **Pulsanti di Recovery**: Se il setup del profilo è incompleto, mostra un banner: *"Il tuo profilo è incompleto. [Completa Setup]"* che re-innesca il processo mancante.

---

## 7. Monitoraggio (Tracing & Alerting)

- **Distributed Tracing (OpenTelemetry)**: Utilizziamo un `trace_id` unico che segue il pacchetto attraverso gRPC, Kafka e i vari database.
- **Alerting**: Soglie di errore elevate nelle DLQ o nella tabella Outbox triggerano notifiche immediate al team dev.

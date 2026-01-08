# Error Handling

L'affidabilità del sistema è garantita da strategie robuste di gestione degli errori sincroni e asincroni.

## Errori Sincroni (gRPC/HTTP)

- **Intercettazione**: Middleware globali catturano panic ed errori non gestiti.
- **Codici di Stato**: Mappatura corretta dei codici di errore gRPC (es. `NotFound`, `InvalidArgument`) a status HTTP appropriati.
- **Circuit Breaker**: Protezione contro servizi degradati per evitare cascading failures.

## Errori Asincroni (Kafka/Saga)

La gestione degli errori nei flussi asincroni è critica per evitare data inconsistency.

### Strategia di Retry

- **Transient Errors**: (es. timeout DB) Il messaggio viene riaccodato con backoff esponenziale.
- **Permanent Errors**: (es. validazione fallita) Il messaggio viene scartato o inviato a una Poison Queue.

### Poison Queue & Dead Letter Queue (DLQ)

Messaggi che falliscono ripetutamente o che sono malformati vengono spostati in una DLQ dedicata per analisi manuale o rielaborazione successiva, impedendo il blocco della partizione Kafka.

### Compensazione (Saga)

Se una transazione distribuita fallisce a metà (es. Utente creato ma errore creazione Profilo), vengono emessi eventi di compensazione (es. `DeleteUser`) per riportare il sistema in uno stato consistente (Rollback).

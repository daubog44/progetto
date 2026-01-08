# Stack di Osservabilità

Il monitoraggio completo è implementato tramite lo stack Grafana/Prometheus/Tempo.

## Componenti

### Metriche (Prometheus & Alloy)

I microservizi espongono metriche Prometheus (latenza, throughput, errori). **Alloy** agisce come collector centralizzato.

### Dashboard (Grafana)

Grafici visuali per monitorare lo stato di salute del sistema.
_(Immagini Dashboard da inserire)_

### Tracing Distribuito (Jaeger/Tempo)

Ogni richiesta viene tracciata attraverso tutti i microservizi attraversati.

- **Trace ID**: Identificativo unico propagato via header HTTP e metadata Kafka.
- **Span**: Rappresenta una singola operazione (es. Query DB, Chiamata gRPC).

_(Esempio di Trace da inserire)_

## Log Centralizzati (Loki)

Log aggregati e ricercabili, correlati con Trace ID per un debugging rapido.

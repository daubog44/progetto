
# 🛠 Guida allo Sviluppo (Development Guide)

Standard per lo sviluppo, l'aggiunta di servizi e la gestione degli errori in Vibely.

---

## 🚀 1. Gestione Microservizi (Automazione)

Utilizza gli script in `scripts/` per gestire il ciclo di vita dei servizi:

### Aggiungere un nuovo servizio
```bash
./scripts/add-service.sh <nome-servizio>
```
Genera automaticamente:
- Struttura cartelle in `microservices/`
- Dockerfile multi-stage (Dev/Prod)
- Configurazione `go.work`, `docker-compose.yml`, `Tiltfile`
- Boilerplate gRPC + `slog`

### Rimuovere un servizio
```bash
./scripts/remove-service.sh <nome-servizio>
```

---

## 📡 2. Workflow gRPC & Buf

Tutti i contratti API sono in `shared/proto`.
Utilizziamo **Buf v2** per la generazione del codice.

### Modificare un'API
1. Modifica il file `.proto` in `shared/proto/<servizio>/v1/`.
2. Genera il codice Go:
   ```bash
   task proto
   ```
3. Implementa l'interfaccia nel server (ricorda di embeddare `Unimplemented...` per compatibilità).

---

## 🛡 3. Coding Standards & Error Handling

### Logging
> **REGOLA**: Usa SEMPRE `log/slog`.
> ❌ MAI usare `fmt.Println` o `log.Println`.

```go
logger.Error("operazione fallita", "error", err, "context_id", id)
```

### Gestione Errori gRPC
Usa i codici di stato appropriati:
- `codes.NotFound` per entità mancanti.
- `codes.InvalidArgument` per input non validi.
- `codes.Internal` per errori imprevisti (e loggali!).

### Resilienza
- **Database**: Usa sempre i bind mount locali per la persistenza (`./data/`).
- **Config**: Passa tutto via variabili d'ambiente in `docker-compose.yml`.

---

## 🔧 4. Utilizzo Librerie Condivise

Quando crei un nuovo servizio, usa i factory condivisi per ottenere resilienza e osservabilità "gratis":

### gRPC Server
```go
import "github.com/username/progetto/shared/pkg/grpcutil"

// Crea un server già configurato con OTel, Logging e Recovery
srv := grpcutil.NewServer()
```

### gRPC Client
```go
// Connessione resiliente con Circuit Breaker e Retry
conn, err := grpcutil.NewClient("target-service:50051", "target-service-breaker-name")
```

### Watermill (Kafka)
```go
import "github.com/username/progetto/shared/pkg/watermillutil"

// Publisher con Tracing
pub, err := watermillutil.NewKafkaPublisher(brokers, logger)

// Router con Recovery, Retry e Circuit Breaker
router, err := watermillutil.NewRouter(logger, "my-service-router")
```

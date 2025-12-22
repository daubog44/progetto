# 🛠 Aggiunta e Gestione Microservizi

In Vibely, l'aggiunta di un nuovo microservizio è un processo standardizzato per garantire coerenza nell'architettura, nel monitoraggio e nel deployment. Abbiamo automatizzato questo processo tramite script per ridurre il boilerplate e gli errori manuali.

---

## 🚀 Automazione: Script di Gestione

Abbiamo due script principali nella cartella `scripts/` per gestire il ciclo di vita dei microservizi.

### 1. Aggiungere un Servizio
Per creare un nuovo servizio Go con tutto il boilerplate necessario (main.go, Dockerfile, gRPC setup):

```bash
./scripts/add-service.sh <nome-servizio>
```

**Cosa fa lo script:**
- Crea la cartella in `microservices/<nome-servizio>`.
- Inizializza il modulo Go.
- Crea un `main.go` con server gRPC di base e supporto `reflection`.
- Crea un `Dockerfile` multi-stage (Production Distroless + Dev hot-reload).
- Registra il servizio in `go.work`.
- Aggiunge il servizio al `docker-compose.yml`.
- Configura il servizio in `Tiltfile` per lo sviluppo locale.

### 2. Rimuovere un Servizio
Per eliminare un servizio e tutti i suoi riferimenti nel progetto:

```bash
./scripts/remove-service.sh <nome-servizio>
```

**Cosa fa lo script:**
- Elimina la cartella del microservizio.
- Rimuove i riferimenti da `go.work`, `docker-compose.yml` e `Tiltfile`.

---

## 📖 Cosa c'è dentro il Boilerplate?

Il microservizio generato include alcune configurazioni standard:

### gRPC Server & Reflection
Il server viene configurato sulla porta `:50051`. Include `reflection.Register(s)`, che permette di usare strumenti come `grpcurl` o `Postman` per esplorare le API senza avere i file `.proto` a portata di mano.

### Structured Logging (`slog`)
Utilizziamo il pacchetto standard di Go `slog` configurato per emettere log in formato JSON. Questo è fondamentale per l'integrazione con **Grafana Loki**.

### Docker Multi-Stage
- **builder**: Compila il binario in modo statico.
- **production**: Utilizza un'immagine **Distroless** per la massima sicurezza.
- **dev**: Immagine ottimizzata per lo sviluppo locale con Tilt.

---

## 🛠 Passaggi Manuali Post-Creazione

Dopo aver usato lo script, dovrai:
1. Definire i tuoi messaggi e servizi in `shared/proto/`.
2. Eseguire `task proto` per generare il codice Go.
3. Implementare la logica di business nel `main.go` del tuo nuovo servizio, sostituendo le funzioni di esempio.
4. (Opzionale) Aggiungere variabili d'ambiente specifiche o dipendenze DB nel `docker-compose.yml`.

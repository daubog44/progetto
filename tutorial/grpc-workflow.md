# Workflow gRPC in Go: Guida Completa

Questa guida spiega come gestire le comunicazioni tra microservizi utilizzando gRPC nel progetto, coprendo dalla modifica del contratto (`.proto`) alla generazione del codice e all'implementazione di Server e Client.

## 1. Architettura del Workflow

Nel nostro progetto, i microservizi comunicano tramite gRPC seguendo questo schema:
1.  **Contratto**: Definito in `proto/`.
2.  **Generazione**: Il codice Go viene generato centralmente.
3.  **Server**: Un servizio (es. `mongo-service`) implementa le interfacce.
4.  **Client**: Altri servizi (es. `test-service`) consumano le API.

---

## 2. Modifica del file .proto

Tutte le definizioni dei servizi si trovano in `proto/data/v1/data.proto`.

### Esempio: Aggiungere un nuovo RPC
Se vuoi aggiungere un metodo per eliminare un utente:

```proto
service DataService {
  // ... altri metodi
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
}

message DeleteUserRequest {
  string id = 1;
}

message DeleteUserResponse {
  bool success = 1;
}
```

---

## 3. Installazione CLI (`buf`)

Il progetto utilizza **buf** per gestire i file protobuf in modo moderno.

### Come installarlo (Linux/macOS):
```bash
# Via Brew
brew install bufbuild/buf/buf

# O via binario diretto
BIN="/usr/local/bin" && \
VERSION="1.28.1" && \
curl -sSL \
  "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" \
  -o "${BIN}/buf" && \
chmod +x "${BIN}/buf"
```

---

## 4. Generazione Codice con Taskfile

Abbiamo automatizzato la generazione del codice tramite il `Taskfile.yaml` alla radice del progetto.

Ogni volta che modifichi un file `.proto`, esegui:
```bash
task proto
```
Questo comando esegue internamente `buf generate` nella cartella `proto/`, aggiornando il codice Go in `proto/gen/go/`.

---

## 5. Implementazione Server: Perché `Unimplemented...`?

Quando implementi il server (`mongo-service/main.go`), vedrai questo:

```go
type server struct {
    datav1.UnimplementedDataServiceServer // <--- Fondamentale
    db *mongo.Client
}
```

### Perché lo usiamo?
Si chiama **Forward Compatibility** (Compatibilità Futura).
*   **Sicurezza**: Se aggiungi un nuovo metodo al `.proto` e rigeneri il codice, il tuo server continuerà a compilare perché `Unimplemented...` fornisce un'implementazione di default che restituisce un errore "Unimplemented".
*   **Obbligatorio**: I plugin moderni di Go richiedono questo embedding per evitare che il codice si rompa quando l'interfaccia cresce.

---

## 6. Esempio Client

Ecco come un altro servizio può chiamare il `mongo-service`:

```go
// 1. Connessione al server
conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
defer conn.Close()

// 2. Creazione del client
client := datav1.NewDataServiceClient(conn)

// 3. Chiamata RPC
res, err := client.GetData(context.Background(), &datav1.GetDataRequest{Id: "123"})
```

---

## 7. Riassunto Comandi Rapidi

| Task | Comando |
| :--- | :--- |
| **Generare codice** | `task proto` |
| **Avviare servizi** | `task up` (Docker Compose) |
| **Sviluppo locale** | `task dev` (Tilt) |
| **Test** | `task test` |

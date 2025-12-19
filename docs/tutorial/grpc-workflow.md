# Workflow gRPC in Go: Guida Completa

Questa guida spiega come gestire le comunicazioni tra microservizi utilizzando gRPC nel progetto, coprendo dalla modifica del contratto (`.proto`) alla generazione del codice e all'implementazione di Server e Client.

## 1. Architettura del Workflow

Nel nostro progetto, i microservizi comunicano tramite gRPC seguendo questo schema:
1.  **Contratto**: Definito in `shared/proto/`.
2.  **Generazione**: Il codice Go viene generato centralmente.
3.  **Server**: Un servizio (es. `mongo-service`) implementa le interfacce.
4.  **Client**: Altri servizi (es. `test-service`) consumano le API.

---

## 2. Modifica del file .proto

Tutte le definizioni dei servizi si trovano in `shared/proto/data/v1/data.proto`.

### Esempio: Aggiungere un nuovo RPC
Se vuoi aggiungere un metodo per aggiungere una recensione a un libro:

```proto
service CulturalContentService {
  // ... altri metodi
  rpc AddBookReview(AddBookReviewRequest) returns (AddBookReviewResponse);
}

message AddBookReviewRequest {
  string user_id = 1;
  string book_id = 2;
  string content = 3;
  int32 rating = 4; // 1-5
  bool contains_spoilers = 5;
}

message AddBookReviewResponse {
  string review_id = 1;
  bool success = 2;
}
```

---

## 3. Installazione CLI (`buf`)

Il progetto utilizza **buf v2** per gestire i file protobuf in modo moderno e scalabile.

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

Abbiamo automatizzato la generazione del codice tramite il `Taskfile.yaml`. Con **Buf v2**, utilizziamo la **Managed Mode** per gestire automaticamente i package Go, semplificando la manutenzione.

Ogni volta che modifichi un file `.proto`, esegui:
```bash
task proto
```
Questo comando esegue `buf generate` nella cartella `shared/proto/`. Grazie ai **remote plugins**, la generazione è sempre coerente tra host e container. Il codice generato viene salvato in `shared/proto/gen/go/`.

---

## 5. Implementazione Server: Perché `Unimplemented...`?

Quando implementi il server (`mongo-service/main.go`), vedrai questo:

```go
type server struct {
    culturalv1.UnimplementedCulturalContentServiceServer // <--- Fondamentale
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
conn, _ := grpc.Dial("cultural-service:50051", grpc.WithInsecure())
defer conn.Close()

// 2. Creazione del client
client := culturalv1.NewCulturalContentServiceClient(conn)

// 3. Chiamata RPC
res, err := client.AddBookReview(context.Background(), &culturalv1.AddBookReviewRequest{
    UserId: "user_01",
    BookId: "book_99",
    Content: "Incredibile finale!",
    Rating: 5,
    ContainsSpoilers: true,
})
```

---

## 7. Riassunto Comandi Rapidi

| Task | Comando |
| :--- | :--- |
| **Generare codice** | `task proto` |
| **Avviare servizi** | `task up` (Docker Compose) |
| **Sviluppo locale** | `task dev` (Tilt) |
| **Test** | `task test` |

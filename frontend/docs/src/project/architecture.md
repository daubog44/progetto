# Architettura del Sistema

L'architettura è basata su microservizi che collaborano per fornire le funzionalità di business.

## Scalabilità

Ogni microservizio è stateless e può essere scalato orizzontalmente. L'uso di Kafka permette di gestire picchi di traffico ammortizzando il carico tramite code di messaggi.

## Gestione Sincrona vs Asincrona

### Sincrono (gRPC/HTTP)

Utilizzato per operazioni che richiedono risposta immediata (es. login, lettura dati profilo).

- **Vantaggi**: Bassa latenza, semplicità.
- **Svantaggi**: Accoppiamento temporale.

### Asincrono (Event-Driven)

Utilizzato per operazioni complesse, notifiche, e propagazione dati (es. registrazione utente, invio email).

- **Vantaggi**: Disaccoppiamento, resilienza, scalabilità.
- **Pattern**: Saga Orchestration.

## Data Flow Diagram

```mermaid
graph TD
    Client[Client Frontend]
    GW[API Gateway]
    Auth[Auth Service]
    User[User Service]
    Notif[Notification Service]
    Kafka{Kafka Broker}

    Client -->|HTTP Request| GW
    GW -->|gRPC| Auth
    GW -->|gRPC| User

    Auth -->|UserRegistered Event| Kafka
    User -->|UpdateProfile Event| Kafka

    Kafka -->|Consume| Notif
    Kafka -->|Consume| User

    subgraph Data Flow
    direction TB
    Auth -.-> DB[(Postgres/Mongo)]
    end
```

### Saga Pattern Flow

Esempio di flusso di registrazione utente (Saga):

```mermaid
sequenceDiagram
    participant C as Client
    participant GW as Gateway
    participant A as Auth Service
    participant K as Kafka
    participant O as Orchestrator
    participant U as User Service
    participant N as Notification Service

    C->>GW: POST /register
    GW->>A: Create Account
    A->>K: Publish UserCreated
    K->>O: Consume UserCreated
    O->>K: Send CreateProfile Command
    K->>U: Consume CreateProfile
    U-->>K: ProfileCreated Event
    K->>O: Consume ProfileCreated
    O->>K: Send SendWelcomeEmail Command
    K->>N: Consume SendWelcomeEmail
    N-->>K: EmailSent Event
```

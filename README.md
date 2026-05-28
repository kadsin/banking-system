# Banking System

I’ve spent the last few days reading about challenges in the banking system. It’s clear that these systems need high availability and strict consistency to satisfy users.

Common issues include lack of reconciliation, double-spending, and bottlenecks in balance adjustments. While imagining a digital bank that could serve an entire country, I looked into architectures like the Digital Twin. I realized that most systems actually rely on eventual consistency, which seems to work well for them.

Of course, a system like this needs many different components working together to ensure it runs correctly with minimal downtime.

## Architecture

### Overview

I have implemented a robust, scalable architecture designed to address the challenges of a country-scale banking system.

#### System Components

The architecture consists of three client-facing microservices situated behind an **API Gateway**, which serves as the single entry point for all client requests.

1. **Account Service:** Manages user registration and account creation.
2. **Balance Service:** Provides real-time balance inquiries and validation.
3. **Transaction (TX) Service:** Orchestrates fund transfers.

```mermaid
graph LR
    User([User]) --> Kong[API Gateway]
    Kong --> Acc[Account Service]
    Kong --> Bal[Balance Service]
    Kong --> TX[TX Service]
```

#### Transaction Flow & Consistency

When a user initiates a transfer, the **TX Service** first queries the **Balance Service** to validate sufficient funds. Once validated, it requests a balance adjustment and records the transaction in its local database using the **Transactional Outbox Pattern**. A CDC (**Debezium**) monitors the outbox table and streams events to **Kafka**.

From there, a **Core Worker** consumes these events for processing. If the transaction meets all business rules, it is committed, and a success message is published. If a failure occurs, a "Failed Transaction" event is broadcast to the system.

```mermaid
graph LR
    TX[TX Service] -->|1. Write| Outbox[(Outbox DB)]
    Outbox -->|2. Capture| CDC[Debezium]
    CDC -->|3. Stream| Kafka{Kafka}
    Kafka -->|4. Process| Core[Core Worker]
    Core -->|5. Commit| Ledger[(Main DB)]
```

### Addressing Edge Cases & Fault Tolerance

#### What if the Balance Service experiences a cache miss in Redis?

To ensure high performance, the **Balance Service** performs atomic adjustments within **Redis**. To handle potential cache misses, I have introduced a **Hydrator Service**.

The Hydrator maintains its own balance snapshot by consuming all transaction events from Kafka. If the Balance Service finds a key missing in Redis, it requests the state from the Hydrator, which repopulates the cache.

```mermaid
graph TD
    Bal[Balance Service] -- Miss --> Hyd[Hydrator]
    Kafka{Kafka} -.->|Continuous Sync| Hyd
    Hyd -- Repopulate --> Redis[(Balance Redis)]
```

#### How are duplicate transactions (Double Spending) prevented?

The TX Service implements **Idempotency**. Every transaction request must include a unique **Idempotency Key**. Upon receipt, the service stores this key in Redis with a `PENDING` status. It then listens to Kafka for the final status (`SUCCESS`/`FAILED`) to update the record. This ensures that if a user retries the same request, the system can provide the current status rather than re-processing the transfer.

```mermaid
sequenceDiagram
    Client->>TX Service: Request (Idempotency Key)
    TX Service->>Redis: Check/Set Key (Pending)
    Note right of Redis: If exists, return previous status
    TX Service->>Outbox: Save Transaction
    Kafka-->>TX Service: Update Status (Success/Fail)
```

#### What happens if a transaction is invalidated by the Core Worker?

In scenarios where the Core Worker rejects a transaction (e.g., due to a blocked account status), it publishes a "Transaction Failed" event to Kafka. A dedicated **Saga Worker** (reconciliation) listens for these events and initiates a **Compensating Transaction** (refund) through the Balance Service to revert any temporary holds or adjustments, ensuring the system returns to a consistent state.

```mermaid
graph LR
    Core[Core Worker] -->|Reject| Kafka{Kafka}
    Kafka -->|Consume| Saga[Saga Worker]
    Saga -->|Refund| Bal[Balance Service]
```

#### A user reports that their money was not transferred correctly (Observability).

This architecture uses **Kong** as the API Gateway to ensure that every request is assigned a unique **Request ID**. This ID is stored and propagated throughout the system, enabling effective log tracing and request tracking across services.
In addition, we use **Sentry** for real-time exception tracking and **Prometheus** for monitoring system health and consumer lag.

The system should log all events to provide full visibility into failures and help identify what happened during each incident.

### Notes

- To increase performance and reduce latency, we can implement an active-active (region-based) replication model for the core database; however, this will increase architectural complexity.
- To ensure high-performance retrieval of transaction history, I have applied the CQRS pattern, integrating an OLAP database for use by the TX service.
- Within this architecture, we can increase Kafka partition count and horizontally scale consumers/workers (and stateless services) to raise throughput and reduce processing latency. It just needs to add more detailed parameters to workers (e.g. regions).
- **The system should implement a circuit breaker mechanism to prevent cascading failures when critical components become unavailable.**

### Full System Diagram

```mermaid
graph TD
    classDef actor fill:#fdfefe;
    classDef gateway fill:#2e4053,color:#fff;
    classDef service fill:#2874a6,color:#fff;
    classDef database fill:#1e8449,color:#fff;
    classDef queue fill:#8e44ad,color:#fff;
    classDef cache fill:#cb4335,color:#fff;
    classDef worker fill:#d68910,color:#fff;
    classDef cdc fill:#17a589,color:#fff;

    User([User]):::actor --> GW[Gateway]:::gateway

    subgraph Services [Microservices Layer]
        direction TB
        TX[TX service]:::service
        Bal[Balance service]:::service
        Acc[Account service]:::service
    end

    GW -->|Request transfer| TX
    TX -->|Get transactions| GW
    Bal -->|Provide balance| Acc
    GW -->|Get details| Acc

    TX --- TXRedis[(TX Idempotency<br/>Redis)]:::cache
    TX -->|1. Validate & adjustment| Bal
    TX -->|2. Write event| Outbox[(Outbox DB)]:::database

    Bal -->|Atomic update| BalRedis[(Balance Redis)]:::cache
    Bal -.->|Cache missed<br>Warmup my cache| Hyd

    subgraph AsyncStreaming [Event Streaming]
        CDC((Debezium CDC)):::cdc
        Kafka{Kafka}:::queue
    end

    Outbox --> CDC
    CDC --> Kafka

    subgraph CoreSystem [Core Processing]
        Core[Core Worker]:::worker
        MainDB[(Main DB)]:::database
        AnalyticsDB[(OLAP)]:::database
    end

    Kafka -->|Read TXs<br>Kafka engine| AnalyticsDB
    AnalyticsDB -->|Get TX list| TX

    Kafka -->|Process TX| Core
    Core -->|Bulk add TXs| MainDB
    Core -->|Failed TX| Kafka

    subgraph Reliability [Reliability & Recovery]
        direction TB
        Rec[Saga Worker]:::worker
        Hyd[Hydrator Service]:::service
        HydStore[(Snapshot Store)]:::cache
    end

    Kafka -->|Consume failed TX| Rec
    Rec -->|Refund| Bal

    Kafka -->|Sync state| Hyd
    Hyd -->|Store snapshot| HydStore
    Hyd -.->|Warmup cache| BalRedis

    Acc --- AccDB[(Account DB)]:::database
```

## How To Run

### 1. Prepare the environment

Copy the example environment file and adjust values if needed.

```bash
cp .env.example .env
```

Example `.env`:

```env
APP_NAME="BankingSystem"
APP_ENV=dev
APP_DEBUG=true

DOC_AUTH_USERNAME=admin
DOC_AUTH_PASSWORD=123456

AUTH_JWT_SECRET=change_me
AUTH_ACCESS_TTL_MIN=60

TOPIC_TRANSACTIONS=prod.tx
TOPIC_FAILED=prod.failed

TREASURY_INITIAL_BALANCE=1000000
```

**Note**: `TREASURY_INITIAL_BALANCE` is the banking system’s initial treasury balance.

### 2. Install dependencies

```bash
make init
```

### 3. Run the application

```bash
make app
```

This command builds the application into `./build/app` and starts the HTTP server.

### 4. Run tests

Create a .env.testing file and run all tests:

```bash
cp .env.example .env
make test
```

Test command documentation:

```bash
make test:help
```

### 5. Swagger UI

After starting the application, the Swagger UI is available at:

```text
http://localhost:3000/docs
```

Swagger UI credentials are controlled with:

```env
DOC_AUTH_USERNAME=admin
DOC_AUTH_PASSWORD=123456
```

## Implementation

### Testing

This project needs increased test coverage, but in a real-world scenario, it would also require extensive integration and stress testing.

### Abstractions

| Interface                 | Purpose                                             |
| ------------------------- | --------------------------------------------------- |
| `AccountRepository`       | Account CRUD operations                             |
| `BalanceService`          | Balance queries and adjustments with cache handling |
| `TransferService`         | Fund transfer orchestration                         |
| `LedgerRepository`        | Redis-backed balance ledger                         |
| `OutboxRepository`        | Event outbox for CDC pattern                        |
| `OlapRepository`          | Transaction history queries (CQRS)                  |
| `HydratorService`         | Cache repopulation on misses                        |
| `TxIdempotencyRepository` | Idempotency key tracking                            |

### Core

| Implementation            | Location                                          | Details                                                   |
| ------------------------- | ------------------------------------------------- | --------------------------------------------------------- |
| `AccountService`          | `internal/service/account.go`                     | Creates accounts, validates balance, initiates transfers  |
| `BalanceService`          | `internal/service/balance.go`                     | Get/Adjust balance with automatic hydration on cache miss |
| `TransferService`         | `internal/service/transfer.go`                    | Validates funds, checks idempotency, writes outbox events |
| `Hydrator`                | `internal/service/hydrator.go`                    | Consumes Kafka events, maintains balance snapshots        |
| `LedgerRepository`        | `internal/datalayer/ledger_repository.go`         | Redis-backed atomic balance updates                       |
| `OutboxRepository`        | `internal/datalayer/outbox_repository.go`         | Persists events for CDC                                   |
| `OlapRepository`          | `internal/datalayer/olap_repository.go`           | Stores transactions for analytics                         |
| `TxIdempotencyRepository` | `internal/datalayer/tx_idempotency_repository.go` | Redis-backed idempotency key storage                      |
| `AccountRepository`       | `internal/datalayer/account_repository.go`        | Account persistence                                       |

### Transfer flow

```mermaid
graph TD
    subgraph Client["Client Request"]
        A["Client Initiates Transfer"]
    end

    subgraph TXService["TX Service Processing"]
        B["Check Idempotency<br/>(TxIdempotencyRepository)"]
        C["Validate Accounts<br/>(AccountRepository)"]
        D["Check Account Status<br/>(Not Blocked)"]
        E["Validate Balance<br/>(BalanceService)"]
        F["Create OutboxEvent<br/>(OutboxRepository)"]
        G["Adjust Balance<br/>From: -Amount<br/>To: +Amount"]
        H["Store Idempotency Key"]
    end

    subgraph Queue["Event Streaming"]
        I["Pull from Outbox<br>Publish to Queue<br/>(Topic: Transactions)"]
        J["Message Persisted<br/>with Offset"]
    end

    subgraph CoreWorker["Core Worker Processing"]
        K["Fetch batch<br>from queue"]
        L["Unmarshal Transaction"]
        M["Adjust MainDB<br/>From: -Amount<br/>To: +Amount"]
        N{Success?}
        O["Mark Complete<br/>(Status: Completed)"]
        P["Publish Failed Event<br/>(Topic: Failed)"]
    end

    subgraph MainDB["Persistence"]
        Q["BulkCreate Transactions<br/>(MainTransactionRepository)"]
        R["Commit queue offset<br/>(Kafka offset)"]
    end

    subgraph Saga["Reconciliation - On Failure"]
        S["Saga Worker<br/>(Topic: Failed)"]
        T["Unmarshal failed TX"]
        U["Refund both accounts<br/>From: +Amount<br/>To: -Amount"]
        V["Commit failed<br>queue Offset"]
    end

    subgraph Cache["Cache Handling"]
        W["On Balance Cache Miss"]
        X["Hydrator.Repopulate"]
        Y["Fetch Snapshots<br/>(HydratorRepository)"]
        Z["Replay queue events<br/>from last offset"]
        AA["Compute Balance"]
        AB["Update Ledger"]
        AC["Save New Snapshot"]
    end

    A -->|1| B
    B -->|2| C
    C -->|3| D
    D -->|4| E
    E -->|5| F
    F -->|6| G
    G -->|7| H
    H -->|8| I
    I -->|9| J

    J -->|10| K
    K -->|11| L
    L -->|12| M
    M -->|13| N
    N -->|Success| O
    N -->|Failure| P
    O -->|14| Q
    P -->|14b| Q
    Q -->|15| R

    P -->|16| S
    S -->|17| T
    T -->|18| U
    U -->|19| V

    E -.->|Cache Miss| W
    W -->|20| X
    X -->|21| Y
    Y -->|22| Z
    Z -->|23| AA
    AA -->|24| AB
    AB -->|25| AC
    AC -.->|Resume| E

    classDef client fill:#4A90E2,color:#fff
    classDef service fill:#7ED321,color:#fff
    classDef queue fill:#8e44ad,color:#fff
    classDef worker fill:#d68910,color:#fff
    classDef db fill:#1e8449,color:#fff
    classDef saga fill:#cb4335,color:#fff
    classDef cache fill:#F5A623,color:#fff
    classDef decision fill:#FFD700,color:#000

    class A client
    class B,C,D,E,F,G,H service
    class I,J queue
    class K,L,M,O worker
    class Q,R db
    class S,T,U,V saga
    class W,X,Y,Z,AA,AB,AC cache
    class N decision
```

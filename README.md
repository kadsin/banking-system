# Banking System

I’ve spent the last few days reading about challenges in the banking system. It’s clear that these systems need high availability and strong consistency to satisfy users.

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

#### A user reports that their money was not transferred correctly.

This architecture uses **Kong** as the Gateway so that all requests have a unique **Request ID**. This ID is stored and propagated through the system, allowing us to trace logs effectively. Additionally, we use **Sentry** for real-time exception tracking and **Prometheus** for monitoring system health and consumer lag.

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
    GW -->|Check balance| Bal
    GW -->|Get details| Acc

    TX --- TXRedis[(TX Idempotency<br/>Redis)]:::cache
    TX -->|1. Validate & adjustment| Bal
    TX -->|2. Write event| Outbox[(Outbox DB)]:::database

    Bal -->|Atomic update| BalRedis[(Balance Redis)]:::cache
    Bal -.->|Cache missed| Hyd

    subgraph AsyncStreaming [Event Streaming]
        CDC((Debezium CDC)):::cdc
        Kafka{Kafka}:::queue
    end

    Outbox --> CDC
    CDC --> Kafka

    subgraph CoreSystem [Core Processing]
        Core[Core Worker]:::worker
        MainDB[(Main DB)]:::database
    end

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

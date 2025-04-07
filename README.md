
---

# Wallet API Documentation

### 🚀 Build and Run

To build the project and spin up **3 services (Redis, API, and Postgres)** in detached mode, run:

```bash
docker-compose up --build -d
```

## 📌 Notes

- All API requests **require a `UserId` in the `Authorization` header**. (This is a simplified replacement for Auth Token)
- All `amount` and `balance` values are represented in **cents**.  
  For example: `$11.70` is shown as `1170` (data type: `int64`)

---

## 🚀 API Endpoints

### 💰 Deposit Money

**Endpoint:**  
`POST http://localhost:8080/api/deposit/v1`

**Headers:**
```
Authorization: 1  // UserId
```

**Request Body:**
```json
{
  "amount": 67
}
```

**Response:**
```json
{
  "balance": 167
}
```

---

### 🏧 Withdraw Money

**Endpoint:**  
`POST http://localhost:8080/api/withdraw/v1`

**Headers:**
```
Authorization: 1  // UserId
```

**Request Body:**
```json
{
  "amount": 67
}
```

**Response:**
```json
{
  "balance": 33
}
```

---

### 🔁 Transfer Money to Another User

**Endpoint:**  
`POST http://localhost:8080/api/transfer/v2`

**Headers:**
```
Authorization: 1  // UserId
```

**Request Body:**
```json
{
  "destination_user_id": 2,
  "amount": 1003
}
```

**Response:**
```json
{
  "success": true
}
```

---

### 📊 Check Wallet Balance

**Endpoint:**  
`GET http://localhost:8080/api/wallet/balance/v1`

**Headers:**
```
Authorization: 1  // UserId
```

**Response:**
```json
{
  "balance": 10000
}
```

---

### 📜 View Transaction History

**Endpoint:**  
`GET http://localhost:8080/api/transactions/v1`

**Headers:**
```
Authorization: 2  // UserId
```

**Query Parameters:**
- `type` (optional): `0 = All`, `1 = Deposit`, `2 = Withdraw`, `3 = Transfer` (default = `0`)
- `page` (default = `1`)
- `page_size` (default = `30`)

**Example Request:**
```
GET http://localhost:8080/api/transactions/v1?type=1&page=1&page_size=30
```

**Response:**
```json
{
  "balance": 10000
}
```

---

Let me know if you want to add OpenAPI (Swagger) support or diagrams for flow!

# Explain Decisions Made
# Highlight How Should Review
# Which Features Chose Not To Do in Submission

Use gzip for compressed response, speed up response

a
# TODO:
# 1. add indexing 

# TODO:
# 1. trace

Build Command
docker-compose up --build -d
docker volume prune -a
docker image prune -a

Added multiple workers for v2

snowflake id for transaction uuid

add indexing DestWalletId in Transaction for better history query

uses int64 for cents

throttling (currently using window sliding, better to use token bucket, or window sliding comparing last and first time, more flexible)
redis for distributed locking, multiple api instance checking same redis

grpc to wallet service



Caching Balance and Transaction History (Trade-off Between Performance and Consistency)
When users are eager to check a transfer, they might spam to refresh their balance and transaction history. Throttling might not be the best way to block users off on reasonable request per second, so I've added a redis layer instead of letting the spam requests entering the postgres. 
The most straightforward way of caching is when a request comes in, check if it exist in redis, if it doesn't, then we will write into the cache with expiry time (let's say 5 minutes) after retrieving it from database, so it will return data directly from the redis if it hasn't expired. But with this way, this might caused having stale data issue where users can only get the latest balance / history every 5 minutes. Certain events will update the balance / transaction history, so I decided to also work on the events on top of the previous cache implementation. 
After writing to the database, I prefer to evict the corresponding cache instead of updating corresponding cache directly. This ensures the next read will fetch fresh data from the DB and repopulate the cache. Updating corresponding cache might be unnecessary since not all users check their transactions or balance after performing events like transfer, withdraw or deposit. And evicting cache only writes to redis when users actually need the data.

---

## Transfer - Core Functions

The core transfer functionality involves four main steps:

1. **Row lock**: Lock the wallet balance of both User A and User B.
2. **Minus wallet balance from User A**: Deduct the transfer amount from User A's wallet.
3. **Add wallet balance to User B**: Credit the transfer amount to User B's wallet.
4. **Log transactions**: 
   - Add a transaction record for User A indicating the deduction.
   - Add a transaction record for User B indicating the addition.

## Version 1: Synchronous Transaction (Implemented)

### How it Works

In this version, all four steps are wrapped in a **single database transaction**:

- The user must wait for the transfer to complete before receiving a response. The transaction locks both User A and User B’s wallet balances, performs the transfer, and commits or rolls back based on the success of all steps.

### Pros
- **Straightforward and easy to implement**: Simple and easy-to-understand flow.
- **All-or-nothing**: If any step fails, the entire transaction rolls back, ensuring data consistency.

### Cons
- **Slower response time**: The system waits until the transfer is fully completed, causing delays, especially when there is lock contention.
- **Timeout risks**: High load may result in request timeouts or context timeouts due to database locking.

---

## Version 2: Asynchronous Transfer (Implemented, Retry Not Handled)

### How it Works

This version leverages a **job channel** to handle transfer requests asynchronously:

- Each transfer request is pushed into a **job channel** and then processed by a worker pool.
- The user receives an immediate response after the task is enqueued.
- The actual transfer happens in the background with no lock contention, as the transfer tasks are consumed by workers.

### Pros
- **Immediate feedback to users**: Users receive a response right after submitting their transfer request, without waiting for the transaction to complete.
- **Scales well under high concurrency**: No lock wait or timeout issues, making it highly scalable under high loads.

### Cons
- **Failure and retry handling**: Since the user does not wait for the transfer to complete, there is no immediate feedback on failures. Additional mechanisms for retry and failure handling are necessary to ensure reliability.

---

## Areas to Improve

### 1. Decouple Wallet into Microservice

- **Create a dedicated wallet service**: Separating the wallet functionality into a standalone microservice enhances scalability and allows for independent scaling of the wallet system.

### 2. Introduce gRPC API Gateway with Kafka Integration

- **gRPC API Gateway**: Expose gRPC endpoints via a **.proto** file to handle transfer requests.
- **Kafka**: Use Kafka to publish transfer requests into a topic. This allows for asynchronous processing with multiple consumers.

#### Why Use Kafka?
- **Message persistence**: Kafka messages are persistent and stored on disk.
- **High availability**: Kafka replicates messages across multiple nodes for fault tolerance.
- **Backpressure handling**: Kafka acts as a buffer to prevent overwhelming services, managing load spikes.
- **At-least-once delivery**: Kafka guarantees at-least-once delivery of messages. You can use `snowflake_id` to prevent duplicate processing of messages.

### 3. Create a Kafka Consumer like Transfer Service

- Implement a Kafka consumer that listens to transfer requests from the Kafka topic.
- Consumers (workers) process transfer jobs independently, ensuring parallel processing and scalability.

### 4. Add Retry Mechanism

- **Retry Mechanism**: If a transfer fails, push the job to a **retry topic** for reprocessing or a **Dead Letter Queue (DLQ)** for future inspection.

### 5. Implement Log Monitoring Service

- **Centralized logging**: Use a tool like **Grafana** to centralize logs and monitor key metrics such as CPU usage, memory usage, and response times across services.

### 6. Graceful Shutdown

- Implement **graceful shutdown** to ensure that the service can clean up resources properly (e.g., closing open database connections, stopping the worker pool) and avoid inconsistencies, especially in cases where the process exits halfway during a transfer.

---

## Testing

### How Long Spent on Tests

- Spent approximately **2-3 hours** writing tests for the `/pkg/model` directory, achieving **50% test coverage**.
- Didn't put more hours into it because I couldn't finish implementing all functions.

### Test Coverage

- The following test files were created:

```bash
go test -cover ./...
```

Test files:
- `/pkg/model/wallet_test.go`
- `/pkg/model/worker_test.go`
- `/pkg/utils/middlewares_test/throttle_test.go`
- `/pkg/model/transaction_test.go`

### Coverage

- **Total coverage**: 50% for the `/pkg/model` directory.

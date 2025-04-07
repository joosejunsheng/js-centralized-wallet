
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

```bash
curl -X POST http://localhost:8080/api/deposit/v1 \
  -H "Authorization: 1" \
  -H "Content-Type: application/json" \
  -d '{"amount": 67}'
```

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

```bash
curl -X POST http://localhost:8080/api/withdraw/v1 \
  -H "Authorization: 1" \
  -H "Content-Type: application/json" \
  -d '{"amount": 67}'
```

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

```bash
curl -X POST http://localhost:8080/api/transfer/v2 \
  -H "Authorization: 1" \
  -H "Content-Type: application/json" \
  -d '{"destination_user_id": 2, "amount": 1003}'
```

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

```bash
curl -X GET http://localhost:8080/api/wallet/balance/v1 \
  -H "Authorization: 1"
```

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

```bash
curl -X GET "http://localhost:8080/api/transactions/v1?type=1&page=1&page_size=30" \
  -H "Authorization: 2"
```

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

# Highlight How Should Review

Listen then Serve instead of http.ListenAndServe, implements middleware, more customizable
Uses ctx, lg := trace.Logger(ctx) whenever using log, to generate unique "trace path"

Use gzip for compressed response, speed up response

add indexing DestWalletId in Transaction for better history query

uses int64 for cents

throttling (currently using window sliding, better to use token bucket, or window sliding comparing last and first time, more flexible)
redis for distributed locking, multiple api instance checking same redis




## Caching Balance and Transaction History

In order to handle users' requests to frequently check their balance and transaction history, especially during transfer events, I decided to implement a caching mechanism. This approach addresses the potential issue of excessive load on the PostgreSQL database caused by frequent requests while ensuring a balance between performance and consistency.

### Caching Strategy

Users might spam refresh requests to check their balance and transaction history. While throttling can limit the number of requests per second, it isn't always the best solution for blocking users who make reasonable requests. Instead, I've introduced a **Redis** layer to cache data and reduce the load on the database.

### Cache Flow

1. **Request Handling**:
   - When a request comes in to fetch a user's balance or transaction history, the system first checks if the data exists in **Redis**.
   
2. **Cache Miss**:
   - If the data is not found in the cache (a "cache miss"), we retrieve the data from the PostgreSQL database and write it into Redis with an expiry time (e.g., 5 minutes).

3. **Cache Hit**:
   - If the data is found in the cache (a "cache hit") and it hasn't expired, we return the data directly from Redis without querying the database, resulting in faster response times.

### Trade-offs: Performance vs. Consistency

While caching improves performance, it can introduce **stale data** issues. For example:
- If the cache is set to expire every 5 minutes, users might not see the latest balance or transaction history until the cache is refreshed. This creates a trade-off between performance and data consistency.

## Handling Events and Cache Invalidation

Certain events, like **transfers**, **withdrawals**, or **deposits**, will update a user's balance or transaction history. To address this and ensure data consistency:

- After writing the updated data to the database, we **evict the corresponding cache** entry rather than directly updating it. This ensures that the next read operation will fetch fresh data from the database and repopulate the cache.
  
- **Why Eviction Over Update?**
  - Evicting the cache is preferred because not all users will immediately check their balance or transaction history after performing actions like transfers or deposits.
  - Eviction ensures that Redis only stores fresh data when it's actually needed, avoiding unnecessary writes to Redis and preventing stale cache reads.

By combining Redis caching with cache eviction after certain events, we achieve a good balance between performance (by reducing database load) and consistency (ensuring fresh data is fetched when necessary). This caching strategy ensures that the system can handle high traffic without sacrificing the accuracy of user data.


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

http://localhost:8080/api/transfer/v1

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

## Version 2: Asynchronous Transfer With Worker Pool (Implemented, Retry Not Handled)

http://localhost:8080/api/transfer/v2

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
- **At-least-once delivery**: Kafka guarantees at-least-once delivery of messages. We can use `snowflake_id` to prevent duplicate processing of messages.

### 3. Create a Kafka Consumer like Transfer Service

- Implement a Kafka consumer that listens to transfer requests from the Kafka topic.
- Consumers (workers) process transfer jobs independently, ensuring parallel processing and scalability.

### 4. Add Retry Mechanism

- **Retry Mechanism**: If a transfer fails, push the job to a **retry topic** for reprocessing or a **Dead Letter Queue (DLQ)** for future inspection.

### 5. Implement Log Monitoring Service

- **Centralized logging**: Use a tool like **Grafana** to centralize logs and monitor key metrics such as CPU usage, memory usage, and response times across services.

### 6. Graceful Shutdown

- Implement **graceful shutdown** to ensure that the service can clean up resources properly (e.g., closing open database connections, stopping the worker pool) and avoid inconsistencies, especially in cases where the process exits halfway during a transfer.

### 7. Implement Snowflake Id

I'm currently using an auto-increment integer for IDs. As the system scales to multiple servers or instances, I plan to implement the **Snowflake ID Generator**. This will provide globally unique identifiers with the following structure:

#### Snowflake ID Structure:

1. **Machine ID**: 
   - Identifies the machine or instance generating the ID.
   - Useful for distributed systems and ensuring unique ID generation across multiple machines.

2. **Sequence**: 
   - A counter that increments for each new ID generated within the same millisecond.
   - Prevents ID collisions when generating multiple IDs per millisecond.

3. **Timestamp**: 
   - The current time (usually in milliseconds or microseconds) when the ID is created.
   - Ensures that IDs are ordered chronologically.


---

## Testing

### How Long Spent on Tests

- Spent approximately **2-3 hours** writing tests for the `/pkg/model` directory, achieving **50% test coverage**.
- Didn't put more hours into it because couldn't finish implementing all functions.

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

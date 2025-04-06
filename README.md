
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
# Installation
# Highlight How Should Review
# Areas to be Improved
# How Long Spent on Test
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

4 steps in transfer
Row lock User A and User B
Minus wallet balance from User A
Add wallet balance to User B
Add 1 row to Transactions indicating User A minus
Add 1 row to Transactions indicating User B add

Version 1 (Implemented)
Wrap All 4 Steps in a Single Database Transaction

How it works:
User waits until their request to acquires lock to proceed with the transfer.
User get response when the transfer is fully completed.

Pros:
- Straightforward and easy to implement
- All-or-nothing: if any step fails, entire transaction rolls back

Cons:
- Slower response time due to lock contention
- May result in request timeouts or context timeouts under high load

Version 2 (Implemented, but did not handle retry if fails)
Push Transfer Requests to a Job Channel - Async

How it works:
Each incoming transfer request is pushed into a job channel.
A worker pool (with multiple workers) consumes the jobs from the channel.
The user receives an immediate response once the task is enqueued — actual processing happens in the background.

What to Expect:
User get response straight away as task pushed into job channel, task to be consumed in the background

Pros:
- Immediate feedback to users
- Scales well under high concurrency — no lock wait or timeout issues

Cons: 
- Requires additional retry and failure-handling mechanisms, since users won’t know instantly if a transfer fails


Other Solutions to Explore

Decouple Wallet Into Microservice
Create a dedicated service for Wallet, separating from the main service for better scalability.

Introduce gRPC API Gateway with Kafka Integration
Communicate with gRPC API Gateway with .proto files by exposing gRPC endpoints, publish transfer request into Kafka topics to be processed.

Create a Kafka Consumer like Transfer Service
Consumes transfer jobs from Kafka topic to process independently based on number of consumers in consumer group.

Add retry mechanism
If a transfer tails, push to other topics like retry topic to retry, or even a Dead Letter Queue for reprocessing.

Implement Log Monitoring Service like Granfana
Centralize Logs across services, also monitor metrics like CPU and memory usage.

Why use Kafka?
- Kafka message is persistent, on disk.
- Has multiple replica, high availability.
- Has good backpressure handling, acts like a buffer to not overwhelm service.
- Guarantees at-least-once delivery, use snowflake_id to verify, preventing duplication.













# User can deposit money into his/her wallet
# User can withdraw money from his/her wallet
# User can send money to another user
# User can check his/her wallet balance
# User can view his/her transaction history

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











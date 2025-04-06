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
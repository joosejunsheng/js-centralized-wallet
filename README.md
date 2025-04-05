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
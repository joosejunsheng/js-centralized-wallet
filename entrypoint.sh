#!/bin/sh
set -e

echo "Starting Postgres"

until pg_isready -h "$POSTGRES_HOST" -p 5432 -U "$POSTGRES_USER"; do
  echo "Waiting Postgres"
  sleep 2
done

echo "Postgres ready"

exec "$@"

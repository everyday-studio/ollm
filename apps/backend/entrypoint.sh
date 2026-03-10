#!/bin/sh
set -e

# Build the connection string dynamically from individual DB_* env vars.
# This ensures the correct host is used regardless of environment
# (e.g., DB_HOST=db in Docker, DB_HOST=localhost in local dev).
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="host=${DB_HOST:-localhost} port=${DB_PORT:-5432} user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME:-mydb} sslmode=${DB_SSLMODE:-disable}"

echo "Running database migrations..."
/app/goose -dir /app/migrations up

echo "Starting server..."
exec "$@"

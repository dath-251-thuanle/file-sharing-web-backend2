#!/bin/sh
set -e

echo "Waiting for database..."
until pg_isready -d "$DATABASE_URL"; do
  sleep 2
done

echo "Running migrations..."
migrate -path /app/migrations -database "$DATABASE_URL" up

echo "Starting app..."
exec ./main

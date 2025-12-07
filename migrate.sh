#!/bin/sh
set -e

echo "Starting migrations..."
echo "DATABASE_URL: $DATABASE_URL"

migrate -path=/migrations -database="$DATABASE_URL" up

echo "Migrations completed!"
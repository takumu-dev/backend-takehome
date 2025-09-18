#!/bin/bash

# Database migration script
set -e

# Load environment variables if .env file exists
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Set default values if not provided
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}
DB_USER=${DB_USER:-root}
DB_PASSWORD=${DB_PASSWORD:-}
DB_NAME=${DB_NAME:-blog_platform}

# Construct database URL
if [ -z "$DB_PASSWORD" ]; then
    DATABASE_URL="mysql://${DB_USER}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?parseTime=true"
else
    DATABASE_URL="mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?parseTime=true"
fi

MIGRATIONS_PATH="app/internal/infrastructure/database/migrations"

echo "Running database migrations..."
echo "Database URL: mysql://${DB_USER}:***@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}"

# Run migrations
/home/takumu/go/bin/migrate -path $MIGRATIONS_PATH -database "$DATABASE_URL" up

echo "Migrations completed successfully!"

#!/bin/sh

echo "Running schema migrations..."
go-migrate \
    -source file://migrations \
    -database postgres://"$DB_USERNAME":"$DB_PASSWORD"@"$DB_HOST":"$DB_PORT"/"$DB_NAME"?sslmode="$DB_SSLMODE" up

echo "Starting API..."
exec ./api
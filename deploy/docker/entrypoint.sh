#!/bin/sh
set -e

echo "Running database migrations..."
./migrate up

exec "$@"

#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$SCRIPT_DIR"

DATABASE_URL="${DATABASE_URL:-postgres://tyszzz:tyszzz@127.0.0.1:5433/tyszzz?sslmode=disable}"

echo "Running database migrations..."
./migrate \
  -path "$SCRIPT_DIR/migration" \
  -database "$DATABASE_URL" \
  up

echo "Database deploy completed."
#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$SCRIPT_DIR"

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <seed-version>"
    echo "Example: $0 v01"
    exit 1
fi

SEED_VERSION="$1"
SEED_DIR="$SCRIPT_DIR/seed/$SEED_VERSION"

if [ ! -d "$SEED_DIR" ]; then
    echo "Seed version not found: $SEED_VERSION"
    exit 1
fi

DATABASE_URL="${DATABASE_URL:-postgres://tyszzz:tyszzz@127.0.0.1:5433/tyszzz?sslmode=disable}"

echo "Running seed version: $SEED_VERSION"

found_seed=false

for seed_file in "$SEED_DIR"/*.sql; do
    [ -f "$seed_file" ] || continue
    found_seed=true

    echo "Applying seed: $seed_file"
    psql "$DATABASE_URL" \
        -v ON_ERROR_STOP=1 \
        -1 \
        -f "$seed_file"
done

if [ "$found_seed" = false ]; then
    echo "No SQL files found in: $SEED_DIR"
    exit 1
fi

echo "Seed $SEED_VERSION completed."
#!/bin/bash
# Wait for PostgreSQL to be ready before starting expenseOwl
MAX_RETRIES=30
RETRY_DELAY=2

echo "Waiting for PostgreSQL..."
for i in $(seq 1 $MAX_RETRIES); do
    if pg_isready -h localhost -p 5432 -U owluser -d expenseowldb -q 2>/dev/null; then
        echo "PostgreSQL is ready. Starting ExpenseOwl..."
        exec /opt/project/expenseOwl/expenseowl
    fi
    echo "  Attempt $i/$MAX_RETRIES: PostgreSQL not ready, retrying in ${RETRY_DELAY}s..."
    sleep $RETRY_DELAY
done

echo "ERROR: PostgreSQL did not become ready after $MAX_RETRIES attempts."
exit 1

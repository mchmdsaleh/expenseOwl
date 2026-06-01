#!/bin/bash

# Manual Development Setup for ExpenseOwl
# This script prepares the environment and builds the frontend.

set -e

echo "--- 1. Building Frontend ---"
if command -v npm &> /dev/null; then
    npm install
    cd frontend && npm run build
    cd ..
    echo "Frontend built successfully!"
else
    echo "Error: npm not found. Please install Node.js and npm."
    exit 1
fi

echo ""
echo "--- 2. Database Prerequisites ---"
echo "Ensure you have the following running:"
echo " - PostgreSQL (active)"
echo " - Redis (active on port 6379 or 6380)"
echo ""

echo "--- 3. Environment Configuration ---"
# Create a .env file if it doesn't exist
if [ ! -f .env ]; then
    cat <<EOF > .env
# PostgreSQL Config
STORAGE_TYPE=postgres
STORAGE_URL=localhost:5432/expenseowldb
STORAGE_USER=$(whoami)
STORAGE_PASS=your_db_password
STORAGE_SSL=disable

# Redis Config
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Security Config
JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || echo "change_me_to_something_secure")
JWT_EXPIRY_HOURS=24

# AI Feature Config (Optional)
OPENAI_API_KEY=your_openai_api_key
EOF
    echo ".env file created. Please update it with your actual database credentials."
else
    echo ".env file already exists."
fi

echo ""
echo "--- 4. How to Run ---"
echo "To start the application, run the following command:"
echo "source .env && go run ./cmd/expenseowl"
echo ""
echo "Note: If 'go' is not in your PATH, please install Go 1.23 or newer."

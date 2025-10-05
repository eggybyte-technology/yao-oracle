#!/bin/bash
# Run mock admin service for dashboard testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "🚀 Starting Mock Admin Service..."
echo ""

# Set environment
export PORT=8081

# Run the mock admin
go run cmd/mock-admin/main.go


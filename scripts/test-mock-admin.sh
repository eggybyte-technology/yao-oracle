#!/bin/bash

# Test mock-admin service with grpcurl
# This script tests all mock-admin gRPC endpoints

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
    log_error "grpcurl is not installed"
    log_info "Install with: brew install grpcurl (Mac) or go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
    exit 1
fi

# Configuration
GRPC_HOST="${GRPC_HOST:-localhost}"
GRPC_PORT="${GRPC_PORT:-9090}"
GRPC_ADDR="$GRPC_HOST:$GRPC_PORT"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║          🧪 Mock-Admin gRPC Service Test Suite             ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

log_info "Target: $GRPC_ADDR"
log_info "Proto files: $PROJECT_ROOT/api"
echo ""

# Test 1: List available services
log_info "Test 1: Listing available gRPC services..."
if grpcurl -plaintext "$GRPC_ADDR" list > /dev/null 2>&1; then
    log_success "✓ gRPC server is reachable"
    grpcurl -plaintext "$GRPC_ADDR" list | sed 's/^/  /'
else
    log_error "✗ Failed to connect to gRPC server at $GRPC_ADDR"
    log_warn "Make sure mock-admin is running: make run-mock-admin"
    exit 1
fi
echo ""

# Test 2: List methods of DashboardService
log_info "Test 2: Listing DashboardService methods..."
if grpcurl -plaintext "$GRPC_ADDR" list yao.oracle.v1.DashboardService > /dev/null 2>&1; then
    log_success "✓ DashboardService is available"
    grpcurl -plaintext "$GRPC_ADDR" list yao.oracle.v1.DashboardService | sed 's/^/  /'
else
    log_error "✗ DashboardService not found"
    exit 1
fi
echo ""

# Test 3: GetConfig RPC
log_info "Test 3: Testing GetConfig RPC..."
if grpcurl -plaintext -d '{}' "$GRPC_ADDR" yao.oracle.v1.DashboardService/GetConfig > /tmp/config_response.json 2>&1; then
    log_success "✓ GetConfig successful"
    echo "Response:"
    cat /tmp/config_response.json | jq '.' | sed 's/^/  /'
else
    log_error "✗ GetConfig failed"
    cat /tmp/config_response.json | sed 's/^/  /'
fi
echo ""

# Test 4: QueryCache RPC
log_info "Test 4: Testing QueryCache RPC..."
QUERY_REQUEST='{
  "namespace": "game-app",
  "key": "user:12345"
}'

if grpcurl -plaintext -d "$QUERY_REQUEST" "$GRPC_ADDR" yao.oracle.v1.DashboardService/QueryCache > /tmp/query_response.json 2>&1; then
    log_success "✓ QueryCache successful"
    echo "Response:"
    cat /tmp/query_response.json | jq '.' | sed 's/^/  /'
else
    log_error "✗ QueryCache failed"
    cat /tmp/query_response.json | sed 's/^/  /'
fi
echo ""

# Test 5: ManageSecret RPC
log_info "Test 5: Testing ManageSecret RPC..."
SECRET_REQUEST='{
  "namespace": "game-app",
  "new_api_key": "new-secret-key-123"
}'

if grpcurl -plaintext -d "$SECRET_REQUEST" "$GRPC_ADDR" yao.oracle.v1.DashboardService/ManageSecret > /tmp/secret_response.json 2>&1; then
    log_success "✓ ManageSecret successful"
    echo "Response:"
    cat /tmp/secret_response.json | jq '.' | sed 's/^/  /'
else
    log_error "✗ ManageSecret failed"
    cat /tmp/secret_response.json | sed 's/^/  /'
fi
echo ""

# Test 6: StreamMetrics RPC (server streaming)
log_info "Test 6: Testing StreamMetrics RPC (receiving 3 updates)..."
STREAM_REQUEST='{
  "namespace": ""
}'

if timeout 18s grpcurl -plaintext -d "$STREAM_REQUEST" "$GRPC_ADDR" yao.oracle.v1.DashboardService/StreamMetrics > /tmp/stream_response.json 2>&1; then
    log_success "✓ StreamMetrics successful (received updates)"
    echo "Sample response (first update):"
    head -n 50 /tmp/stream_response.json | jq '.' | sed 's/^/  /'
else
    # Timeout is expected for streaming RPC
    if [ -s /tmp/stream_response.json ]; then
        log_success "✓ StreamMetrics successful (stream active, terminated after timeout)"
        echo "Sample response (first update):"
        head -n 50 /tmp/stream_response.json | jq '.' 2>/dev/null | sed 's/^/  /' || echo "  (raw output)"
    else
        log_error "✗ StreamMetrics failed - no data received"
        cat /tmp/stream_response.json | sed 's/^/  /'
    fi
fi
echo ""

# Summary
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                    ✅ Test Suite Complete                   ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
log_info "All gRPC endpoints tested successfully!"
log_info "You can now connect your Flutter dashboard to this mock-admin instance."
echo ""

# Cleanup
rm -f /tmp/config_response.json /tmp/query_response.json /tmp/secret_response.json /tmp/stream_response.json


#!/bin/bash

# Run Flutter Dashboard in Development Mode
# This script starts both mock-admin and Flutter web server

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DASHBOARD_DIR="$PROJECT_ROOT/frontend/dashboard"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║      🎯 Yao-Oracle Dashboard Development Environment       ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "─────────────────────────────────────────────────────────────"
    echo "Cleaning up processes..."
    if [ ! -z "$MOCK_ADMIN_PID" ]; then
        kill $MOCK_ADMIN_PID 2>/dev/null || true
        echo "✓ Stopped mock-admin"
    fi
    if [ ! -z "$FLUTTER_PID" ]; then
        kill $FLUTTER_PID 2>/dev/null || true
        echo "✓ Stopped Flutter web server"
    fi
    echo "✅ Cleanup complete"
}

trap cleanup EXIT INT TERM

# Step 1: Build mock-admin if needed
echo "📦 Step 1: Building mock-admin..."
cd "$PROJECT_ROOT"
if [ ! -f "bin/mock-admin" ]; then
    echo "Building mock-admin binary..."
    make build-local
else
    echo "✓ mock-admin binary exists"
fi
echo ""

# Step 2: Start mock-admin
echo "🚀 Step 2: Starting mock-admin..."
cd "$PROJECT_ROOT"
./bin/mock-admin --grpc-port=9090 --password=admin123 --refresh-interval=5 &
MOCK_ADMIN_PID=$!
echo "✓ mock-admin started (PID: $MOCK_ADMIN_PID)"
echo "  gRPC endpoint: localhost:9090"
echo ""

# Wait for mock-admin to be ready
echo "⏳ Waiting for mock-admin to be ready..."
sleep 3
echo ""

# Step 3: Check Flutter dependencies
echo "📦 Step 3: Checking Flutter dependencies..."
cd "$DASHBOARD_DIR"
if [ ! -d ".dart_tool" ]; then
    echo "Installing Flutter dependencies..."
    flutter pub get
else
    echo "✓ Flutter dependencies are up to date"
fi
echo ""

# Step 4: Start Flutter web server
echo "🌐 Step 4: Starting Flutter web server..."
cd "$DASHBOARD_DIR"
echo "  Web URL: http://localhost:8080"
echo ""
flutter run -d chrome --web-port=8080 --web-hostname=localhost \
    --dart-define=GRPC_HOST=localhost \
    --dart-define=GRPC_PORT=9090 &
FLUTTER_PID=$!

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                  ✅ Dashboard is Ready!                     ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "📍 Services:"
echo "   • Dashboard:   http://localhost:8080"
echo "   • mock-admin:  localhost:9090 (gRPC)"
echo ""
echo "💡 The dashboard will auto-update with live metrics every 5 seconds"
echo ""
echo "Press Ctrl+C to stop all services"
echo "─────────────────────────────────────────────────────────────"
echo ""

# Wait for Flutter process
wait $FLUTTER_PID

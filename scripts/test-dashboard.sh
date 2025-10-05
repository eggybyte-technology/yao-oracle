#!/bin/bash
# Test dashboard with mock admin service
# One-click script to start both mock-admin (gRPC) and Flutter dashboard

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

source "${SCRIPT_DIR}/common.sh"

# Cleanup function
cleanup() {
    log_warn "Shutting down services..."
    
    if [ -n "${ADMIN_PID:-}" ]; then
        log_info "Stopping mock-admin (PID: $ADMIN_PID)..."
        kill $ADMIN_PID 2>/dev/null || true
    fi
    
    if [ -n "${DASHBOARD_PID:-}" ]; then
        log_info "Stopping Flutter dashboard (PID: $DASHBOARD_PID)..."
        kill $DASHBOARD_PID 2>/dev/null || true
    fi
    
    log_success "✅ Cleanup complete"
}

# Register cleanup on exit
trap cleanup EXIT INT TERM

main() {
    log_success "============================================================"
    log_success "🎯 Yao-Oracle Dashboard Testing Environment (Flutter + gRPC)"
    log_success "============================================================"
    echo ""
    
    # Step 1: Check dependencies
    log_step 1 "Checking dependencies..."
    check_command "go" "Please install Go 1.21+"
    check_command "flutter" "Please install Flutter SDK"
    check_command "protoc" "Please install protoc: brew install protobuf"
    log_success "✅ All dependencies found"
    echo ""
    
    # Step 2: Generate Dart gRPC code
    log_step 2 "Generating Dart gRPC code..."
    cd "$PROJECT_ROOT"
    
    # Check if protoc-gen-dart is installed
    if ! command -v protoc-gen-dart &> /dev/null; then
        log_warn "protoc-gen-dart not found. Installing..."
        flutter pub global activate protoc_plugin
        log_info "Make sure ~/.pub-cache/bin is in your PATH"
        export PATH="$PATH:$HOME/.pub-cache/bin"
    fi
    
    bash scripts/generate_dart_grpc.sh
    log_success "✅ Dart gRPC code generated"
    echo ""
    
    # Step 3: Check Flutter packages
    log_step 3 "Checking Flutter dependencies..."
    cd "$PROJECT_ROOT/frontend/dashboard"
    
    if [ ! -d ".dart_tool" ] || [ ! -f ".dart_tool/package_config.json" ]; then
        log_info "Installing Flutter packages (this may take a minute)..."
        flutter pub get
        log_success "✅ Dependencies installed"
    else
        log_success "✅ Dependencies already installed"
    fi
    echo ""
    
    # Step 4: Start mock admin gRPC service
    log_step 4 "Starting Mock Admin gRPC Service..."
    cd "$PROJECT_ROOT"
    
    # Check if port 9090 is available (gRPC port)
    if lsof -i :9090 > /dev/null 2>&1; then
        log_error "Port 9090 is already in use. Please stop the process using it:"
        lsof -i :9090
        exit 1
    fi
    
    # Start mock-admin in background
    go run cmd/mock-admin/main.go \
        --grpc-port=9090 \
        --password=admin123 \
        --refresh-interval=5 \
        > /tmp/yao-mock-admin.log 2>&1 &
    ADMIN_PID=$!
    
    # Wait for admin to be ready (gRPC doesn't have HTTP health check yet)
    log_info "Waiting for Mock Admin gRPC service to start..."
    sleep 3
    
    # Check if process is still running
    if ! kill -0 $ADMIN_PID 2>/dev/null; then
        log_error "Mock Admin failed to start. Check /tmp/yao-mock-admin.log for details"
        cat /tmp/yao-mock-admin.log
        exit 1
    fi
    
    log_success "✅ Mock Admin gRPC started (PID: $ADMIN_PID)"
    echo ""
    
    # Step 5: Start Flutter dashboard
    log_step 5 "Starting Flutter Dashboard..."
    cd "$PROJECT_ROOT/frontend/dashboard"
    
    # Check if port 8080 is available (Flutter web default)
    if lsof -i :8080 > /dev/null 2>&1; then
        log_warn "Port 8080 is in use. Flutter will pick another port."
    fi
    
    # Start Flutter web with gRPC configuration
    log_info "Starting Flutter web server with gRPC client..."
    log_info "gRPC endpoint: localhost:9090"
    
    # Run Flutter in foreground to see live output
    flutter run -d web-server --web-port 8080 --web-hostname 0.0.0.0 \
        --dart-define=GRPC_HOST=localhost \
        --dart-define=GRPC_PORT=9090 \
        2>&1 | while IFS= read -r line; do
            echo "$line" | tee -a /tmp/yao-dashboard-flutter.log
        done &
    DASHBOARD_PID=$!
    
    # Wait for dashboard to be ready
    log_info "Waiting for Flutter Dashboard to start (this may take 30-60 seconds)..."
    for i in {1..60}; do
        if curl -s http://localhost:8080 > /dev/null 2>&1; then
            log_success "✅ Flutter Dashboard started (PID: $DASHBOARD_PID)"
            break
        fi
        if [ $i -eq 60 ]; then
            log_error "Flutter Dashboard failed to start. Check /tmp/yao-dashboard-flutter.log for details"
            cat /tmp/yao-dashboard-flutter.log
            exit 1
        fi
        sleep 1
    done
    echo ""
    
    # Step 6: Display information
    log_success "============================================================"
    log_success "🎉 Testing Environment Ready!"
    log_success "============================================================"
    echo ""
    log_info "📊 Mock Admin Service (gRPC):"
    log_info "   gRPC Address: localhost:9090"
    log_info "   Log File:     /tmp/yao-mock-admin.log"
    log_info "   Password:     admin123"
    echo ""
    log_info "🎨 Flutter Dashboard:"
    log_info "   URL:       http://localhost:8080"
    log_info "   Log File:  /tmp/yao-dashboard-flutter.log"
    log_info "   Live Console: Watch the terminal for Flutter output"
    echo ""
    log_info "📝 Features to Test:"
    log_info "   ✓ Overview page   - Cluster metrics and component health"
    log_info "   ✓ Proxies page    - Instance list and QPS metrics"
    log_info "   ✓ Nodes page      - Memory usage and metrics"
    log_info "   ✓ Namespaces page - Business space configuration"
    log_info "   ✓ gRPC Stream     - Real-time updates (every 5s)"
    log_info "   ✓ Responsive      - Mobile, tablet, and desktop layouts"
    echo ""
    log_info "🔧 Debugging:"
    log_info "   View mock-admin logs:  tail -f /tmp/yao-mock-admin.log"
    log_info "   View Flutter logs:     tail -f /tmp/yao-dashboard-flutter.log"
    log_info "   Test gRPC endpoint:    grpcurl -plaintext localhost:9090 list"
    echo ""
    log_warn "Press Ctrl+C to stop both services"
    log_success "============================================================"
    echo ""
    
    # Display admin logs in real-time alongside Flutter output
    log_info "📺 Displaying Mock Admin Logs (streaming):"
    echo "------------------------------------------------------------"
    
    # Start tailing admin logs
    tail -f /tmp/yao-mock-admin.log &
    TAIL_ADMIN_PID=$!
    
    # Wait for user to stop (Ctrl+C will trigger cleanup)
    wait $DASHBOARD_PID $ADMIN_PID
    
    # Kill tail processes
    kill $TAIL_ADMIN_PID 2>/dev/null || true
}

main "$@"

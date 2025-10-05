#!/bin/bash
# scripts/proto-generate.sh
# Generates Go code from Protocol Buffers definitions

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! check_command "buf"; then
        log_error "buf is not installed. Install it with:"
        log_error "  go install github.com/bufbuild/buf/cmd/buf@latest"
        exit 1
    fi
    
    if ! check_command "protoc-gen-go"; then
        log_error "protoc-gen-go is not installed. Install it with:"
        log_error "  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
        exit 1
    fi
    
    if ! check_command "protoc-gen-go-grpc"; then
        log_error "protoc-gen-go-grpc is not installed. Install it with:"
        log_error "  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
        exit 1
    fi
    
    log_success "All dependencies are installed"
}

# generate generates the proto code
generate() {
    log_info "Generating Go code from proto files..."
    
    cd "${PROJECT_ROOT}"
    
    # Clean existing generated code
    if [ -d "pb" ]; then
        log_info "Cleaning existing generated code..."
        rm -rf pb
    fi
    
    # Generate code with buf (run from api directory)
    cd "${PROJECT_ROOT}/api"
    buf generate
    
    cd "${PROJECT_ROOT}"
    if [ -d "pb" ]; then
        log_success "Proto code generated successfully in pb/"
    else
        log_error "Proto code generation failed"
        exit 1
    fi
}

# lint runs buf lint
lint() {
    log_info "Linting proto files..."
    cd "${PROJECT_ROOT}"
    buf lint api
    log_success "Proto files are valid"
}

# main entry point
main() {
    log_info "Starting proto code generation"
    
    check_dependencies
    lint
    generate
    
    log_success "Proto code generation complete!"
}

main "$@"


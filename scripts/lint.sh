#!/bin/bash
# scripts/lint.sh
# Runs linters for yao-oracle project

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Configuration
FIX="${FIX:-false}"
STRICT="${STRICT:-false}"

# parse_args processes command-line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --fix)
                FIX=true
                shift
                ;;
            --strict)
                STRICT=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# show_help displays usage information
show_help() {
    cat << EOF
Usage: $0 [options]

Runs linters for yao-oracle project.

Options:
    --fix            Automatically fix issues where possible
    --strict         Use strict linting rules
    -h, --help       Display this help message

Examples:
    $0
    $0 --fix
    $0 --strict

EOF
}

# check_dependencies ensures required tools are installed
check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! check_command "go"; then
        log_error "Go is not installed. Install it from https://go.dev"
        exit 1
    fi
    
    log_success "All dependencies are installed"
}

# run_go_fmt runs go fmt
run_go_fmt() {
    log_info "Running go fmt..."
    
    cd "${PROJECT_ROOT}"
    
    if [[ "${FIX}" == "true" ]]; then
        log_info "Formatting code..."
        gofmt -w .
        log_success "Code formatted"
    else
        local fmt_result=$(gofmt -l .)
        if [ -n "$fmt_result" ]; then
            log_error "The following files need formatting:"
            echo "$fmt_result"
            log_info "Run '$0 --fix' to format automatically"
            exit 1
        fi
        log_success "Code formatting is correct"
    fi
}

# run_go_vet runs go vet
run_go_vet() {
    log_info "Running go vet..."
    
    cd "${PROJECT_ROOT}"
    
    if go vet ./...; then
        log_success "Go vet passed"
    else
        log_error "Go vet found issues"
        exit 1
    fi
}

# run_golangci_lint runs golangci-lint if available
run_golangci_lint() {
    if ! check_command "golangci-lint"; then
        log_warn "golangci-lint is not installed, skipping"
        log_info "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
        return 0
    fi
    
    log_info "Running golangci-lint..."
    
    cd "${PROJECT_ROOT}"
    
    local lint_args=("run" "./...")
    
    if [[ "${FIX}" == "true" ]]; then
        lint_args+=("--fix")
    fi
    
    if [[ "${STRICT}" == "true" ]]; then
        lint_args+=("--enable-all")
    fi
    
    if golangci-lint "${lint_args[@]}"; then
        log_success "golangci-lint passed"
    else
        log_error "golangci-lint found issues"
        if [[ "${FIX}" != "true" ]]; then
            log_info "Run '$0 --fix' to fix automatically"
        fi
        exit 1
    fi
}

# run_proto_lint runs buf lint on proto files
run_proto_lint() {
    if ! check_command "buf"; then
        log_warn "buf is not installed, skipping proto linting"
        return 0
    fi
    
    log_info "Running proto linting..."
    
    cd "${PROJECT_ROOT}"
    
    if [ ! -d "api" ]; then
        log_info "No api directory found, skipping proto linting"
        return 0
    fi
    
    if buf lint api; then
        log_success "Proto files are valid"
    else
        log_error "Proto linting failed"
        exit 1
    fi
}

# check_imports checks import organization
check_imports() {
    log_info "Checking import organization..."
    
    cd "${PROJECT_ROOT}"
    
    if ! check_command "goimports"; then
        log_warn "goimports is not installed, skipping"
        log_info "Install with: go install golang.org/x/tools/cmd/goimports@latest"
        return 0
    fi
    
    if [[ "${FIX}" == "true" ]]; then
        log_info "Organizing imports..."
        find . -name "*.go" -not -path "./pb/*" -not -path "./vendor/*" -exec goimports -w {} \;
        log_success "Imports organized"
    else
        local import_result=$(goimports -l .)
        if [ -n "$import_result" ]; then
            log_error "The following files have import issues:"
            echo "$import_result"
            log_info "Run '$0 --fix' to organize automatically"
            exit 1
        fi
        log_success "Import organization is correct"
    fi
}

# check_mod_tidy checks if go.mod is tidy
check_mod_tidy() {
    log_info "Checking go.mod tidiness..."
    
    cd "${PROJECT_ROOT}"
    
    if [[ "${FIX}" == "true" ]]; then
        log_info "Running go mod tidy..."
        go mod tidy
        log_success "go.mod is tidy"
    else
        # Create a temporary copy to check
        cp go.mod go.mod.bak
        cp go.sum go.sum.bak 2>/dev/null || true
        
        go mod tidy
        
        if ! diff -q go.mod go.mod.bak &> /dev/null; then
            log_error "go.mod is not tidy"
            log_info "Run '$0 --fix' or 'go mod tidy' to fix"
            mv go.mod.bak go.mod
            mv go.sum.bak go.sum 2>/dev/null || true
            exit 1
        fi
        
        rm go.mod.bak
        rm go.sum.bak 2>/dev/null || true
        
        log_success "go.mod is tidy"
    fi
}

# main is the entry point
main() {
    parse_args "$@"
    
    log_info "Starting linting"
    
    if [[ "${FIX}" == "true" ]]; then
        log_info "Auto-fix mode enabled"
    fi
    
    check_dependencies
    run_go_fmt
    run_go_vet
    check_imports
    check_mod_tidy
    run_golangci_lint
    run_proto_lint
    
    log_success "All linting checks passed!"
}

main "$@"


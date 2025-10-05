#!/bin/bash
# scripts/test.sh
# Runs tests for yao-oracle project

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Configuration
COVERAGE="${COVERAGE:-true}"
RACE="${RACE:-true}"
VERBOSE="${VERBOSE:-false}"
OUTPUT_DIR="${OUTPUT_DIR:-${PROJECT_ROOT}}"

# parse_args processes command-line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --no-coverage)
                COVERAGE=false
                shift
                ;;
            --no-race)
                RACE=false
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            --output-dir)
                OUTPUT_DIR="$2"
                shift 2
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

Runs tests for yao-oracle project.

Options:
    --no-coverage          Skip coverage report generation
    --no-race             Skip race detector
    -v, --verbose         Enable verbose test output
    --output-dir DIR      Directory for coverage reports (default: project root)
    -h, --help            Display this help message

Examples:
    $0
    $0 --verbose
    $0 --no-coverage --no-race

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

# run_tests runs the test suite
run_tests() {
    log_info "Running tests..."
    
    cd "${PROJECT_ROOT}"
    
    local test_args=()
    
    if [[ "${VERBOSE}" == "true" ]]; then
        test_args+=("-v")
    fi
    
    if [[ "${RACE}" == "true" ]]; then
        test_args+=("-race")
    fi
    
    if [[ "${COVERAGE}" == "true" ]]; then
        test_args+=("-coverprofile=${OUTPUT_DIR}/coverage.out")
        test_args+=("-covermode=atomic")
    fi
    
    # Run tests
    log_info "Test command: go test ${test_args[*]} ./..."
    
    if go test "${test_args[@]}" ./...; then
        log_success "All tests passed"
    else
        log_error "Tests failed"
        exit 1
    fi
}

# generate_coverage_report generates HTML coverage report
generate_coverage_report() {
    if [[ "${COVERAGE}" != "true" ]]; then
        return 0
    fi
    
    log_info "Generating coverage report..."
    
    cd "${PROJECT_ROOT}"
    
    if [ ! -f "${OUTPUT_DIR}/coverage.out" ]; then
        log_error "Coverage file not found: ${OUTPUT_DIR}/coverage.out"
        exit 1
    fi
    
    # Generate HTML report
    go tool cover -html="${OUTPUT_DIR}/coverage.out" -o "${OUTPUT_DIR}/coverage.html"
    
    # Calculate coverage percentage
    local coverage_percent=$(go tool cover -func="${OUTPUT_DIR}/coverage.out" | \
        grep total | \
        awk '{print $3}')
    
    log_success "Coverage report generated: ${OUTPUT_DIR}/coverage.html"
    log_info "Total coverage: ${coverage_percent}"
    
    # Check coverage threshold (optional)
    local threshold=60
    local coverage_num=$(echo "${coverage_percent}" | sed 's/%//')
    
    if (( $(echo "$coverage_num < $threshold" | bc -l) )); then
        log_warn "Coverage ${coverage_percent} is below threshold ${threshold}%"
    else
        log_success "Coverage ${coverage_percent} meets threshold ${threshold}%"
    fi
}

# run_lint runs linters (optional)
run_lint() {
    log_info "Running linters..."
    
    cd "${PROJECT_ROOT}"
    
    # Run go fmt
    log_info "Running go fmt..."
    local fmt_result=$(gofmt -l .)
    if [ -n "$fmt_result" ]; then
        log_error "The following files need formatting:"
        echo "$fmt_result"
        exit 1
    fi
    log_success "Code formatting is correct"
    
    # Run go vet
    log_info "Running go vet..."
    if go vet ./...; then
        log_success "Go vet passed"
    else
        log_error "Go vet found issues"
        exit 1
    fi
}

# show_summary displays test summary
show_summary() {
    echo ""
    log_info "Test Summary:"
    log_info "=============="
    
    if [[ "${COVERAGE}" == "true" ]]; then
        log_info "Coverage report: ${OUTPUT_DIR}/coverage.html"
        log_info "Coverage data: ${OUTPUT_DIR}/coverage.out"
    fi
    
    log_info ""
    log_info "To view coverage report, open: ${OUTPUT_DIR}/coverage.html"
}

# main is the entry point
main() {
    parse_args "$@"
    
    log_info "Starting test suite"
    
    check_dependencies
    run_tests
    generate_coverage_report
    run_lint
    show_summary
    
    log_success "Test suite complete!"
}

main "$@"


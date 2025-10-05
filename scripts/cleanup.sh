#!/bin/bash
# scripts/cleanup.sh
# Cleans build artifacts and generated files

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Configuration
DEEP_CLEAN="${DEEP_CLEAN:-false}"
DRY_RUN="${DRY_RUN:-false}"

# parse_args processes command-line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --deep)
                DEEP_CLEAN=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
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

Cleans build artifacts and generated files.

Options:
    --deep        Perform deep clean (includes vendor, cache)
    --dry-run     Show what would be deleted without deleting
    -h, --help    Display this help message

Examples:
    $0
    $0 --deep
    $0 --dry-run

EOF
}

# remove_item removes a file or directory
remove_item() {
    local item=$1
    local description=$2
    
    if [ ! -e "${item}" ]; then
        return 0
    fi
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "[DRY RUN] Would remove: ${description}"
        return 0
    fi
    
    log_info "Removing: ${description}"
    rm -rf "${item}"
    log_success "Removed: ${description}"
}

# clean_build_artifacts removes build outputs
clean_build_artifacts() {
    log_step "1/6" "Cleaning build artifacts"
    
    cd "${PROJECT_ROOT}"
    
    remove_item "bin/" "Binary directory (bin/)"
    remove_item "dist/" "Distribution directory (dist/)"
    remove_item "build/output/" "Build output directory"
}

# clean_generated_code removes generated code
clean_generated_code() {
    log_step "2/6" "Cleaning generated code"
    
    cd "${PROJECT_ROOT}"
    
    remove_item "pb/" "Generated Protocol Buffers code (pb/)"
}

# clean_test_artifacts removes test outputs
clean_test_artifacts() {
    log_step "3/6" "Cleaning test artifacts"
    
    cd "${PROJECT_ROOT}"
    
    remove_item "coverage.out" "Coverage data file"
    remove_item "coverage.html" "Coverage HTML report"
    remove_item "coverage.txt" "Coverage text report"
    
    # Find and remove test binaries
    local test_binaries=$(find . -name "*.test" -not -path "./vendor/*" 2>/dev/null || true)
    if [ -n "$test_binaries" ]; then
        for binary in $test_binaries; do
            remove_item "$binary" "Test binary: $binary"
        done
    fi
}

# clean_temporary_files removes temporary files
clean_temporary_files() {
    log_step "4/6" "Cleaning temporary files"
    
    cd "${PROJECT_ROOT}"
    
    remove_item "tmp/" "Temporary directory (tmp/)"
    remove_item "temp/" "Temporary directory (temp/)"
    remove_item "scratch/" "Scratch directory"
    remove_item "local/" "Local development directory"
    
    # Find and remove .log files
    local log_files=$(find . -name "*.log" -not -path "./vendor/*" 2>/dev/null || true)
    if [ -n "$log_files" ]; then
        for log in $log_files; do
            remove_item "$log" "Log file: $log"
        done
    fi
}

# clean_docker_artifacts removes Docker build artifacts
clean_docker_artifacts() {
    log_step "5/6" "Cleaning Docker artifacts"
    
    if ! check_command "docker"; then
        log_info "Docker not installed, skipping"
        return 0
    fi
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "[DRY RUN] Would clean Docker build cache"
        return 0
    fi
    
    log_info "Pruning Docker build cache..."
    docker builder prune -f &> /dev/null || true
    log_success "Docker build cache cleaned"
}

# deep_clean_extras removes additional items in deep clean mode
deep_clean_extras() {
    if [[ "${DEEP_CLEAN}" != "true" ]]; then
        return 0
    fi
    
    log_step "6/6" "Performing deep clean"
    
    cd "${PROJECT_ROOT}"
    
    remove_item "vendor/" "Vendor directory"
    
    # Clean Go cache
    if [[ "${DRY_RUN}" != "true" ]]; then
        log_info "Cleaning Go cache..."
        go clean -cache -modcache -testcache 2>/dev/null || true
        log_success "Go cache cleaned"
    else
        log_info "[DRY RUN] Would clean Go cache"
    fi
    
    # Clean Helm packages
    local helm_packages=$(find helm -name "*.tgz" 2>/dev/null || true)
    if [ -n "$helm_packages" ]; then
        for package in $helm_packages; do
            remove_item "$package" "Helm package: $package"
        done
    fi
}

# show_summary displays cleanup summary
show_summary() {
    echo ""
    log_info "Cleanup Summary:"
    log_info "==============="
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_warn "DRY RUN mode - no files were actually deleted"
    else
        log_success "Cleanup completed successfully"
    fi
    
    if [[ "${DEEP_CLEAN}" == "true" ]]; then
        log_info "Deep clean was performed"
    else
        log_info "Run with --deep for more thorough cleaning"
    fi
    
    # Show disk space saved (if not dry run)
    if [[ "${DRY_RUN}" != "true" ]]; then
        log_info ""
        log_info "Project directory size:"
        du -sh "${PROJECT_ROOT}" 2>/dev/null || true
    fi
}

# main is the entry point
main() {
    parse_args "$@"
    
    log_info "Starting cleanup"
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_warn "Running in DRY RUN mode - no files will be deleted"
    fi
    
    if [[ "${DEEP_CLEAN}" == "true" ]]; then
        log_warn "Deep clean mode enabled - this will remove vendor/ and caches"
    fi
    
    clean_build_artifacts
    clean_generated_code
    clean_test_artifacts
    clean_temporary_files
    clean_docker_artifacts
    deep_clean_extras
    show_summary
    
    log_success "Cleanup complete!"
}

main "$@"


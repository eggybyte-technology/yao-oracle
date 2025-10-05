#!/bin/bash
# scripts/build.sh
# Builds all yao-oracle services for multiple platforms
#
# Usage:
#   ./scripts/build.sh [options]
#
# Options:
#   --service SERVICE    Build only specified service (proxy|node|dashboard)
#   --os OS             Target operating system (default: linux)
#   --arch ARCH         Target architecture (default: amd64,arm64)
#   -v, --verbose       Enable verbose output
#   -h, --help          Display this help message
#
# Environment Variables:
#   BUILD_DIR           Output directory for binaries (default: ./bin)
#   GOOS               Target operating system
#   GOARCH             Target architecture

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Configuration with defaults
BUILD_DIR="${BUILD_DIR:-${PROJECT_ROOT}/bin}"
TARGET_OS="${GOOS:-linux}"
TARGET_ARCHS="${GOARCH:-amd64,arm64}"
SERVICES=("proxy" "node" "dashboard")
VERBOSE=false
VERSION="${VERSION:-$(get_version)}"

# parse_args processes command-line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --service)
                SERVICES=("$2")
                shift 2
                ;;
            --os)
                TARGET_OS="$2"
                shift 2
                ;;
            --arch)
                TARGET_ARCHS="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
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
    grep '^#' "$0" | grep -v '#!/bin/bash' | sed 's/^# //' | sed 's/^#//'
}

# check_prerequisites ensures required tools are installed
check_prerequisites() {
    log_info "Checking prerequisites"
    
    check_command "go" "Go compiler not found. Install from https://go.dev" || exit 1
    
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go version: ${go_version}"
    
    log_success "All prerequisites met"
}

# generate_proto generates protobuf code if needed
generate_proto() {
    if [[ ! -d "${PROJECT_ROOT}/pb" ]] || [[ -z "$(ls -A ${PROJECT_ROOT}/pb 2>/dev/null)" ]]; then
        log_info "Generating protobuf code"
        bash "${SCRIPT_DIR}/proto-generate.sh"
    else
        log_info "Protobuf code already generated, skipping"
    fi
}

# build_service builds a single service for specified architectures
build_service() {
    local service=$1
    local os=$2
    local arch=$3
    
    local output_dir="${BUILD_DIR}/${os}/${arch}"
    local output="${output_dir}/${service}"
    local source="./cmd/${service}"
    
    # Create output directory
    mkdir -p "${output_dir}"
    
    # Build command
    local build_cmd="CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build"
    build_cmd+=" -ldflags=\"-w -s -X main.Version=${VERSION}\""
    build_cmd+=" -o ${output} ${source}"
    
    if [[ "$VERBOSE" == "true" ]]; then
        build_cmd+=" -v"
    fi
    
    log_info "Building ${service} for ${os}/${arch}"
    
    cd "${PROJECT_ROOT}"
    if eval "${build_cmd}" 2>&1; then
        local size=$(du -h "${output}" | cut -f1)
        log_success "Built ${service} (${os}/${arch}) -> ${output} [${size}]"
        return 0
    else
        log_error "Failed to build ${service} for ${os}/${arch}"
        return 1
    fi
}

# build_all_services builds all services for all architectures
build_all_services() {
    local total_services=${#SERVICES[@]}
    local archs=(${TARGET_ARCHS//,/ })
    local total_archs=${#archs[@]}
    local total_builds=$((total_services * total_archs))
    local current=0
    
    log_info "Building ${total_services} service(s) for ${total_archs} architecture(s)"
    log_info "Total builds: ${total_builds}"
    
    for service in "${SERVICES[@]}"; do
        for arch in "${archs[@]}"; do
            current=$((current + 1))
            log_step "${current}/${total_builds}" "Building ${service} for ${TARGET_OS}/${arch}"
            
            if ! build_service "${service}" "${TARGET_OS}" "${arch}"; then
                log_error "Build failed, aborting"
                exit 1
            fi
        done
    done
}

# display_summary shows build summary
display_summary() {
    log_info "Build summary:"
    echo ""
    
    local archs=(${TARGET_ARCHS//,/ })
    for arch in "${archs[@]}"; do
        local output_dir="${BUILD_DIR}/${TARGET_OS}/${arch}"
        if [[ -d "${output_dir}" ]]; then
            echo "  ${TARGET_OS}/${arch}:"
            for service in "${SERVICES[@]}"; do
                local binary="${output_dir}/${service}"
                if [[ -f "${binary}" ]]; then
                    local size=$(du -h "${binary}" | cut -f1)
                    echo "    ✓ ${service}: ${size}"
                fi
            done
        fi
    done
    
    echo ""
}

# main is the entry point of the script
main() {
    parse_args "$@"
    
    log_info "Starting build process for yao-oracle"
    log_info "Version: ${VERSION}"
    log_info "Target OS: ${TARGET_OS}"
    log_info "Target architectures: ${TARGET_ARCHS}"
    log_info "Services: ${SERVICES[*]}"
    log_info "Output directory: ${BUILD_DIR}"
    
    # Check prerequisites
    check_prerequisites
    
    # Generate proto code if needed
    generate_proto
    
    # Build all services
    build_all_services
    
    # Display summary
    display_summary
    
    log_success "All builds completed successfully"
}

# Run main function
main "$@"


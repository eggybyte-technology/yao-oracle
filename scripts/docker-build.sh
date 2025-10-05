#!/bin/bash
# scripts/docker-build.sh
# Builds Docker images for all services using buildx
# Note: This script expects binaries to be pre-built by build.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Configuration
REGISTRY="${DOCKER_REGISTRY:-docker.io/eggybyte}"
VERSION="${VERSION:-$(get_version)}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"
PUSH="${PUSH:-false}"
SERVICES=("proxy" "node" "dashboard")
SKIP_BUILD="${SKIP_BUILD:-false}"

# parse_args processes command-line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --push)
                PUSH=true
                shift
                ;;
            --version)
                VERSION="$2"
                shift 2
                ;;
            --registry)
                REGISTRY="$2"
                shift 2
                ;;
            --platform)
                PLATFORMS="$2"
                shift 2
                ;;
            --service)
                SERVICES=("$2")
                shift 2
                ;;
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
}

# build_binaries builds Go binaries for all target platforms
build_binaries() {
    if [[ "${SKIP_BUILD}" == "true" ]]; then
        log_warn "Skipping binary build (--skip-build flag set)"
        return 0
    fi
    
    log_info "Building Go binaries for target platforms"
    
    # Extract unique architectures from PLATFORMS
    local archs=$(echo "${PLATFORMS}" | tr ',' '\n' | cut -d'/' -f2 | sort -u | tr '\n' ',' | sed 's/,$//')
    
    # Build binaries for all services and architectures
    local build_cmd="bash ${SCRIPT_DIR}/build.sh --os linux --arch ${archs}"
    
    if [[ "${#SERVICES[@]}" -eq 1 ]]; then
        build_cmd+=" --service ${SERVICES[0]}"
    fi
    
    if eval "${build_cmd}"; then
        log_success "Binary build completed"
    else
        log_error "Binary build failed"
        return 1
    fi
}

# verify_binaries checks that required binaries exist
verify_binaries() {
    log_info "Verifying binaries for Docker build"
    
    local archs=$(echo "${PLATFORMS}" | tr ',' '\n' | cut -d'/' -f2 | sort -u)
    local missing=false
    
    for service in "${SERVICES[@]}"; do
        for arch in ${archs}; do
            local binary="${PROJECT_ROOT}/bin/linux/${arch}/${service}"
            if [[ ! -f "${binary}" ]]; then
                log_error "Missing binary: ${binary}"
                missing=true
            else
                local size=$(du -h "${binary}" | cut -f1)
                log_info "Found ${service} (linux/${arch}): ${size}"
            fi
        done
    done
    
    if [[ "${missing}" == "true" ]]; then
        log_error "Some binaries are missing. Run with --skip-build=false or run 'make build' first"
        return 1
    fi
    
    log_success "All required binaries verified"
}

# setup_buildx ensures buildx builder exists
setup_buildx() {
    log_info "Setting up Docker buildx builder"
    
    if ! docker buildx inspect yao-oracle-builder >/dev/null 2>&1; then
        log_info "Creating new buildx builder"
        docker buildx create --name yao-oracle-builder --use
        docker buildx inspect --bootstrap
    else
        log_info "Using existing buildx builder"
        docker buildx use yao-oracle-builder
    fi
    
    log_success "Buildx builder ready"
}

# build_service builds a single service image
build_service() {
    local service=$1
    local dockerfile="build/${service}.Dockerfile"
    local image="${REGISTRY}/yao-oracle-${service}"
    
    log_info "Building ${service} for platforms: ${PLATFORMS}"
    
    local build_args=(
        "--platform" "${PLATFORMS}"
        "--file" "${dockerfile}"
        "--tag" "${image}:${VERSION}"
        "--tag" "${image}:latest"
    )
    
    # Add push or load flag
    if [[ "${PUSH}" == "true" ]]; then
        build_args+=("--push")
    else
        # For local builds, only use single platform
        build_args=()
        build_args+=(
            "--platform" "linux/amd64"
            "--file" "${dockerfile}"
            "--tag" "${image}:${VERSION}"
            "--tag" "${image}:latest"
            "--load"
        )
    fi
    
    # Build the image
    cd "${PROJECT_ROOT}"
    if docker buildx build "${build_args[@]}" .; then
        log_success "Built ${service}:${VERSION}"
    else
        log_error "Failed to build ${service}"
        return 1
    fi
}

# main is the entry point
main() {
    parse_args "$@"
    
    log_info "Starting Docker image build process"
    log_info "Registry: ${REGISTRY}"
    log_info "Version: ${VERSION}"
    log_info "Platforms: ${PLATFORMS}"
    log_info "Push: ${PUSH}"
    log_info "Services: ${SERVICES[*]}"
    
    # Step 1: Build Go binaries
    log_step "1/4" "Building Go binaries for target platforms"
    build_binaries || exit 1
    
    # Step 2: Verify binaries exist
    log_step "2/4" "Verifying binaries"
    verify_binaries || exit 1
    
    # Step 3: Setup buildx
    log_step "3/4" "Setting up Docker buildx"
    setup_buildx
    
    # Step 4: Build Docker images
    log_step "4/4" "Building Docker images"
    local total=${#SERVICES[@]}
    for i in "${!SERVICES[@]}"; do
        local service=${SERVICES[$i]}
        local num=$((i + 1))
        
        log_info "Building ${service} image (${num}/${total})"
        build_service "${service}"
    done
    
    log_success "All Docker images built successfully"
    
    if [[ "${PUSH}" == "true" ]]; then
        log_info "Images pushed to registry: ${REGISTRY}"
    else
        log_info "Images loaded to local Docker"
    fi
}

main "$@"


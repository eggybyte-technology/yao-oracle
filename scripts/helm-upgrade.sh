#!/bin/bash
# scripts/helm-upgrade.sh
# Upgrade yao-oracle Helm release
#
# Usage:
#   ./scripts/helm-upgrade.sh [options]
#
# Options:
#   --namespace NAMESPACE    Kubernetes namespace (default: yao-oracle)
#   --release RELEASE        Helm release name (default: yao-oracle)
#   --values FILE            Values file to use (default: values.yaml)
#   --dry-run                Perform a dry-run upgrade
#   --wait                   Wait for all pods to be ready
#   --force                  Force resource updates
#   -h, --help               Display this help message
#
# Environment Variables:
#   NAMESPACE                Kubernetes namespace
#   RELEASE_NAME             Helm release name
#   VALUES_FILE              Path to values file

set -euo pipefail

# Source common logging functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Check if common.sh exists
if [[ ! -f "${SCRIPT_DIR}/common.sh" ]]; then
    echo "ERROR: common.sh not found. Please ensure scripts/common.sh exists."
    exit 1
fi

source "${SCRIPT_DIR}/common.sh"

# Configuration with defaults
NAMESPACE="${NAMESPACE:-yao-oracle}"
RELEASE_NAME="${RELEASE_NAME:-yao-oracle}"
VALUES_FILE="${VALUES_FILE:-${PROJECT_ROOT}/helm/yao-oracle/values.yaml}"
DRY_RUN=false
WAIT=false
FORCE=false

# parse_args processes command-line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            --release)
                RELEASE_NAME="$2"
                shift 2
                ;;
            --values)
                VALUES_FILE="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --wait)
                WAIT=true
                shift
                ;;
            --force)
                FORCE=true
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
    grep '^#' "$0" | grep -v '#!/bin/bash' | sed 's/^# //'
}

# check_prerequisites verifies required tools are installed
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    check_command "helm" "Helm not found. Install from https://helm.sh" || exit 1
    check_command "kubectl" "kubectl not found. Install from https://kubernetes.io/docs/tasks/tools/" || exit 1
    
    # Check kubectl can connect to cluster
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# validate_release checks if release exists
validate_release() {
    if ! helm status "$RELEASE_NAME" -n "$NAMESPACE" &> /dev/null; then
        log_error "Release '$RELEASE_NAME' not found in namespace '$NAMESPACE'"
        log_error "Use helm-install.sh to install the chart first"
        exit 1
    fi
    log_info "Found existing release: $RELEASE_NAME"
}

# validate_values_file checks if values file exists
validate_values_file() {
    if [[ ! -f "$VALUES_FILE" ]]; then
        log_error "Values file not found: $VALUES_FILE"
        exit 1
    fi
    log_info "Using values file: $VALUES_FILE"
}

# show_diff shows the differences between current and new release
show_diff() {
    if [[ "$DRY_RUN" == "false" ]]; then
        log_info "Checking for differences..."
        helm diff upgrade "$RELEASE_NAME" "${PROJECT_ROOT}/helm/yao-oracle" \
            --namespace "$NAMESPACE" \
            --values "$VALUES_FILE" 2>/dev/null || true
    fi
}

# upgrade_chart upgrades the Helm release
upgrade_chart() {
    log_info "Upgrading Helm release..."
    log_info "Release: $RELEASE_NAME"
    log_info "Namespace: $NAMESPACE"
    
    local helm_cmd="helm upgrade $RELEASE_NAME ${PROJECT_ROOT}/helm/yao-oracle"
    helm_cmd+=" --namespace $NAMESPACE"
    helm_cmd+=" --values $VALUES_FILE"
    helm_cmd+=" --install"  # Install if not exists
    
    if [[ "$DRY_RUN" == "true" ]]; then
        helm_cmd+=" --dry-run --debug"
        log_info "Performing dry-run upgrade..."
    fi
    
    if [[ "$WAIT" == "true" ]]; then
        helm_cmd+=" --wait --timeout 5m"
        log_info "Waiting for all pods to be ready..."
    fi
    
    if [[ "$FORCE" == "true" ]]; then
        helm_cmd+=" --force"
        log_warn "Force flag enabled - resources will be recreated"
    fi
    
    if eval "$helm_cmd"; then
        log_success "Helm release upgraded successfully"
        return 0
    else
        log_error "Helm upgrade failed"
        return 1
    fi
}

# show_status displays the release status
show_status() {
    if [[ "$DRY_RUN" == "true" ]]; then
        return 0
    fi
    
    log_info "Release status:"
    helm status "$RELEASE_NAME" -n "$NAMESPACE"
    
    echo ""
    log_info "Deployed pods:"
    kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/instance=$RELEASE_NAME"
    
    echo ""
    log_info "Release history:"
    helm history "$RELEASE_NAME" -n "$NAMESPACE" --max 5
}

# main is the entry point of the script
main() {
    parse_args "$@"
    
    log_info "Starting Helm upgrade for yao-oracle"
    
    check_prerequisites
    validate_values_file
    
    if [[ "$DRY_RUN" == "false" ]]; then
        validate_release
        show_diff
    fi
    
    if upgrade_chart; then
        show_status
        log_success "Helm upgrade completed successfully"
        
        if [[ "$DRY_RUN" == "false" ]]; then
            echo ""
            log_info "To rollback this release if needed:"
            log_info "  helm rollback $RELEASE_NAME -n $NAMESPACE"
        fi
    else
        log_error "Helm upgrade failed"
        exit 1
    fi
}

# Run main function
main "$@"

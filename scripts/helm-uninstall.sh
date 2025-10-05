#!/bin/bash
# scripts/helm-uninstall.sh
# Uninstall yao-oracle Helm release from Kubernetes cluster
#
# Usage:
#   ./scripts/helm-uninstall.sh [options]
#
# Options:
#   --namespace NAMESPACE    Kubernetes namespace (default: yao-oracle)
#   --release RELEASE        Helm release name (default: yao-oracle)
#   --keep-namespace         Keep the namespace after uninstall
#   --dry-run                Show what would be deleted
#   -h, --help               Display this help message
#
# Environment Variables:
#   NAMESPACE                Kubernetes namespace
#   RELEASE_NAME             Helm release name

set -euo pipefail

# Source common logging functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Check if common.sh exists
if [[ ! -f "${SCRIPT_DIR}/common.sh" ]]; then
    echo "ERROR: common.sh not found. Please ensure scripts/common.sh exists."
    exit 1
fi

source "${SCRIPT_DIR}/common.sh"

# Configuration with defaults
NAMESPACE="${NAMESPACE:-yao-oracle}"
RELEASE_NAME="${RELEASE_NAME:-yao-oracle}"
KEEP_NAMESPACE=false
DRY_RUN=false

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
            --keep-namespace)
                KEEP_NAMESPACE=true
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
        log_warn "Release '$RELEASE_NAME' not found in namespace '$NAMESPACE'"
        log_warn "Nothing to uninstall"
        return 1
    fi
    return 0
}

# show_resources displays resources that will be deleted
show_resources() {
    log_info "Resources to be deleted:"
    echo ""
    
    kubectl get all -n "$NAMESPACE" -l "app.kubernetes.io/instance=$RELEASE_NAME" 2>/dev/null || true
    
    echo ""
    log_info "ConfigMaps and Secrets:"
    kubectl get configmap,secret -n "$NAMESPACE" -l "app.kubernetes.io/instance=$RELEASE_NAME" 2>/dev/null || true
    
    if [[ "$KEEP_NAMESPACE" == "false" ]]; then
        echo ""
        log_warn "Namespace '$NAMESPACE' will also be deleted"
    fi
}

# confirm_uninstall asks for user confirmation
confirm_uninstall() {
    if [[ "$DRY_RUN" == "true" ]]; then
        return 0
    fi
    
    echo ""
    log_warn "This action will delete the release '$RELEASE_NAME' and all its resources"
    
    if [[ "$KEEP_NAMESPACE" == "false" ]]; then
        log_warn "This will also delete the namespace '$NAMESPACE'"
    fi
    
    read -p "Are you sure you want to proceed? (yes/no): " -r
    echo
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        log_info "Uninstall cancelled by user"
        exit 0
    fi
}

# uninstall_release uninstalls the Helm release
uninstall_release() {
    log_info "Uninstalling Helm release..."
    log_info "Release: $RELEASE_NAME"
    log_info "Namespace: $NAMESPACE"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would execute: helm uninstall $RELEASE_NAME -n $NAMESPACE"
        return 0
    fi
    
    if helm uninstall "$RELEASE_NAME" -n "$NAMESPACE"; then
        log_success "Helm release uninstalled successfully"
        return 0
    else
        log_error "Helm uninstall failed"
        return 1
    fi
}

# delete_namespace deletes the namespace
delete_namespace() {
    if [[ "$KEEP_NAMESPACE" == "true" ]]; then
        log_info "Keeping namespace: $NAMESPACE"
        return 0
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would delete namespace: $NAMESPACE"
        return 0
    fi
    
    log_info "Deleting namespace: $NAMESPACE"
    
    if kubectl delete namespace "$NAMESPACE" --timeout=60s; then
        log_success "Namespace deleted successfully"
        return 0
    else
        log_error "Failed to delete namespace"
        return 1
    fi
}

# main is the entry point of the script
main() {
    parse_args "$@"
    
    log_info "Starting Helm uninstall for yao-oracle"
    
    check_prerequisites
    
    if ! validate_release; then
        exit 0
    fi
    
    show_resources
    confirm_uninstall
    
    if uninstall_release; then
        delete_namespace
        log_success "Helm uninstall completed successfully"
    else
        log_error "Helm uninstall failed"
        exit 1
    fi
}

# Run main function
main "$@"

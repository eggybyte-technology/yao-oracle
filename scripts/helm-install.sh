#!/bin/bash
# scripts/helm-install.sh
# Install yao-oracle Helm chart to Kubernetes cluster
#
# Usage:
#   ./scripts/helm-install.sh [options]
#
# Options:
#   --namespace NAMESPACE    Kubernetes namespace (default: yao-oracle)
#   --release RELEASE        Helm release name (default: yao-oracle)
#   --values FILE            Values file to use (default: values.yaml)
#   --dry-run                Perform a dry-run installation
#   --wait                   Wait for all pods to be ready
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
        log_error "Please check your kubeconfig and cluster connection"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# validate_values_file checks if values file exists
validate_values_file() {
    if [[ ! -f "$VALUES_FILE" ]]; then
        log_error "Values file not found: $VALUES_FILE"
        exit 1
    fi
    log_info "Using values file: $VALUES_FILE"
}

# create_namespace creates the namespace if it doesn't exist
create_namespace() {
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_info "Namespace '$NAMESPACE' already exists"
    else
        log_info "Creating namespace: $NAMESPACE"
        kubectl create namespace "$NAMESPACE"
        log_success "Namespace '$NAMESPACE' created"
    fi
}

# install_chart installs the Helm chart
install_chart() {
    log_info "Installing Helm chart..."
    log_info "Release: $RELEASE_NAME"
    log_info "Namespace: $NAMESPACE"
    
    local helm_cmd="helm install $RELEASE_NAME ${PROJECT_ROOT}/helm/yao-oracle"
    helm_cmd+=" --namespace $NAMESPACE"
    helm_cmd+=" --values $VALUES_FILE"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        helm_cmd+=" --dry-run --debug"
        log_info "Performing dry-run installation..."
    fi
    
    if [[ "$WAIT" == "true" ]]; then
        helm_cmd+=" --wait --timeout 5m"
        log_info "Waiting for all pods to be ready..."
    fi
    
    if eval "$helm_cmd"; then
        log_success "Helm chart installed successfully"
        return 0
    else
        log_error "Helm installation failed"
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
    log_info "Services:"
    kubectl get svc -n "$NAMESPACE" -l "app.kubernetes.io/instance=$RELEASE_NAME"
}

# main is the entry point of the script
main() {
    parse_args "$@"
    
    log_info "Starting Helm installation for yao-oracle"
    
    check_prerequisites
    validate_values_file
    
    if [[ "$DRY_RUN" == "false" ]]; then
        create_namespace
    fi
    
    if install_chart; then
        show_status
        log_success "Helm installation completed successfully"
        
        if [[ "$DRY_RUN" == "false" ]]; then
            echo ""
            log_info "To check the status of your release:"
            log_info "  helm status $RELEASE_NAME -n $NAMESPACE"
            echo ""
            log_info "To view the release notes:"
            log_info "  helm get notes $RELEASE_NAME -n $NAMESPACE"
            echo ""
            log_info "To run Helm tests:"
            log_info "  helm test $RELEASE_NAME -n $NAMESPACE"
        fi
    else
        log_error "Helm installation failed"
        exit 1
    fi
}

# Run main function
main "$@"

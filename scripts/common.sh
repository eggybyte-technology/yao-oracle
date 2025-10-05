#!/bin/bash
# scripts/common.sh
# Common functions for all scripts

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# log_info prints an info message
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# log_success prints a success message
log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# log_error prints an error message
log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# log_warn prints a warning message
log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# log_step prints a step message
log_step() {
    echo -e "${BLUE}[STEP $1]${NC} $2"
}

# check_command checks if a command exists
check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "$1 is not installed"
        return 1
    fi
    return 0
}

# get_version returns the version from git
get_version() {
    if git rev-parse --git-dir > /dev/null 2>&1; then
        git describe --tags --always --dirty 2>/dev/null || echo "dev"
    else
        echo "dev"
    fi
}


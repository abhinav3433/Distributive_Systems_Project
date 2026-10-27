#!/bin/bash
set -e

echo "=========================================================="
echo " Tearing Down Distributed E-Commerce Kubernetes Stack     "
echo "=========================================================="

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed or not in PATH."
    exit 1
fi

kubectl delete -k "$SCRIPT_DIR" --ignore-not-found=true

echo "Kubernetes stack deleted successfully."

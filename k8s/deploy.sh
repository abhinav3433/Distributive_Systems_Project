#!/bin/bash
set -e

echo "=========================================================="
echo " Deploying Distributed E-Commerce Platform to Kubernetes "
echo "=========================================================="

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 1. Check kubectl availability
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed or not in PATH."
    exit 1
fi

echo "Applying Kubernetes manifests via Kustomize..."
kubectl apply -k "$SCRIPT_DIR"

echo "Waiting for core databases to initialize..."
kubectl rollout status deployment/postgres-auth -n distributed-ecommerce --timeout=90s || true
kubectl rollout status deployment/postgres-product -n distributed-ecommerce --timeout=90s || true
kubectl rollout status deployment/postgres-order -n distributed-ecommerce --timeout=90s || true
kubectl rollout status deployment/postgres-payment -n distributed-ecommerce --timeout=90s || true
kubectl rollout status deployment/redis -n distributed-ecommerce --timeout=60s || true
kubectl rollout status deployment/rabbitmq -n distributed-ecommerce --timeout=90s || true

echo "Waiting for microservices to reach Ready status..."
kubectl rollout status deployment/auth-service -n distributed-ecommerce --timeout=90s || true
kubectl rollout status deployment/product-service -n distributed-ecommerce --timeout=90s || true
kubectl rollout status deployment/cart-service -n distributed-ecommerce --timeout=90s || true
kubectl rollout status deployment/order-service -n distributed-ecommerce --timeout=90s || true
kubectl rollout status deployment/payment-service -n distributed-ecommerce --timeout=90s || true
kubectl rollout status deployment/gateway -n distributed-ecommerce --timeout=90s || true
kubectl rollout status deployment/frontend -n distributed-ecommerce --timeout=90s || true

echo ""
echo "=========================================================="
echo " Deployment Complete! Pods status:"
echo "=========================================================="
kubectl get pods -n distributed-ecommerce

echo ""
echo "Access Points:"
echo "  - Frontend:    http://localhost:30000 (NodePort) or run: kubectl port-forward svc/frontend 3000:3000 -n distributed-ecommerce"
echo "  - API Gateway: http://localhost:30080 (NodePort) or run: kubectl port-forward svc/gateway 8080:8080 -n distributed-ecommerce"
echo "=========================================================="

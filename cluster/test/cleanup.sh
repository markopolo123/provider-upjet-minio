#!/usr/bin/env bash
set -aeuo pipefail

CLUSTER_NAME="provider-minio-testing"

echo "=== Cleaning up Kind cluster and resources ==="

# Stop any port forwards
echo "Stopping any background processes..."
pkill -f "kubectl port-forward.*minio" 2>/dev/null || true

# Check if cluster exists and delete it
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "Deleting Kind cluster: ${CLUSTER_NAME}"
    kind delete cluster --name="${CLUSTER_NAME}"
    echo "Cluster deleted successfully"
else
    echo "Cluster ${CLUSTER_NAME} not found"
fi

# Clean up any leftover contexts
kubectl config delete-context "kind-${CLUSTER_NAME}" 2>/dev/null || true

# Clean up any leftover kubeconfig entries
kubectl config unset clusters."kind-${CLUSTER_NAME}" 2>/dev/null || true
kubectl config unset users."kind-${CLUSTER_NAME}" 2>/dev/null || true

echo "=== Cleanup complete ==="
#!/usr/bin/env bash
set -aeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CLUSTER_NAME="provider-minio-testing"
MINIO_NAMESPACE="minio-system"

echo "=== Bootstrapping Kind cluster for Minio provider testing ==="

# Check if Kind is installed
if ! command -v kind &> /dev/null; then
    echo "Kind not found. Installing Kind..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        if command -v brew &> /dev/null; then
            brew install kind
        else
            echo "Please install Kind manually: https://kind.sigs.k8s.io/docs/user/quick-start/"
            exit 1
        fi
    else
        echo "Please install Kind manually: https://kind.sigs.k8s.io/docs/user/quick-start/"
        exit 1
    fi
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo "kubectl not found. Please install kubectl."
    exit 1
fi

# Check if helm is installed
if ! command -v helm &> /dev/null; then
    echo "helm not found. Please install helm."
    exit 1
fi

# Create Kind cluster
echo "Creating Kind cluster: ${CLUSTER_NAME}"
kind create cluster --config="${PROJECT_ROOT}/cluster/kind-config.yaml" --name="${CLUSTER_NAME}"

# Wait for cluster to be ready
echo "Waiting for cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=300s

# Install Crossplane
echo "Installing Crossplane..."
kubectl create namespace crossplane-system || true
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update
helm install crossplane crossplane-stable/crossplane \
    --namespace crossplane-system \
    --create-namespace \
    --wait

# Wait for Crossplane to be ready
echo "Waiting for Crossplane to be ready..."
kubectl wait --for=condition=Available deployment/crossplane --namespace=crossplane-system --timeout=300s

# Create upbound-system namespace for provider
echo "Creating upbound-system namespace..."
kubectl create namespace upbound-system || true

# Install Minio using simple deployment
echo "Installing Minio..."
kubectl create namespace ${MINIO_NAMESPACE} || true

# Deploy Minio using a simple deployment (not operator)
echo "Creating Minio deployment..."
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minio
  namespace: ${MINIO_NAMESPACE}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: minio
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
      - name: minio
        image: quay.io/minio/minio:RELEASE.2024-10-29T16-01-48Z
        command:
        - /bin/bash
        - -c
        args:
        - minio server /data --console-address :9001
        env:
        - name: MINIO_ACCESS_KEY
          value: "minioadmin"
        - name: MINIO_SECRET_KEY
          value: "minioadmin"
        ports:
        - containerPort: 9000
          name: api
        - containerPort: 9001
          name: console
        volumeMounts:
        - name: storage
          mountPath: /data
        readinessProbe:
          httpGet:
            path: /minio/health/ready
            port: 9000
          initialDelaySeconds: 10
          periodSeconds: 10
        livenessProbe:
          httpGet:
            path: /minio/health/live
            port: 9000
          initialDelaySeconds: 10
          periodSeconds: 10
      volumes:
      - name: storage
        emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: minio
  namespace: ${MINIO_NAMESPACE}
spec:
  selector:
    app: minio
  ports:
  - name: api
    port: 9000
    targetPort: 9000
  - name: console
    port: 9001
    targetPort: 9001
  type: ClusterIP
EOF

# Wait for Minio to be ready
echo "Waiting for Minio to be ready..."
kubectl wait --for=condition=Available deployment/minio --namespace=${MINIO_NAMESPACE} --timeout=300s

# Port forward Minio for testing (in background)
echo "Setting up port forwarding for Minio..."
kubectl port-forward --namespace=${MINIO_NAMESPACE} svc/minio 9000:9000 &
MINIO_PF_PID=$!

# Give port forward time to establish
sleep 5

# Test Minio connectivity
echo "Testing Minio connectivity..."
timeout 30 bash -c 'until curl -s http://localhost:9000/minio/health/ready; do sleep 1; done' || {
    echo "Warning: Could not verify Minio readiness, but continuing..."
}

# Create additional service alias for API access
echo "Creating Minio API service alias..."
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: minio-api
  namespace: ${MINIO_NAMESPACE}
spec:
  selector:
    app: minio
  ports:
  - name: api
    port: 9000
    targetPort: 9000
  type: ClusterIP
EOF

echo "=== Kind cluster bootstrap complete ==="
echo "Cluster name: ${CLUSTER_NAME}"
echo "Minio namespace: ${MINIO_NAMESPACE}"
echo "Minio credentials: minioadmin/minioadmin"
echo ""
echo "To use this cluster:"
echo "  kubectl config use-context kind-${CLUSTER_NAME}"
echo ""
echo "To access Minio:"
echo "  kubectl port-forward --namespace=${MINIO_NAMESPACE} svc/minio 9000:9000"
echo "  Then open: http://localhost:9000"
echo ""
echo "To install the provider:"
echo "  make local-deploy"
echo ""
echo "To run e2e tests:"
echo "  make e2e-minio"
echo ""
echo "To delete the cluster:"
echo "  kind delete cluster --name ${CLUSTER_NAME}"

# Kill the background port forward
kill $MINIO_PF_PID 2>/dev/null || true
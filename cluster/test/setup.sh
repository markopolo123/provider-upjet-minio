#!/usr/bin/env bash
set -aeuo pipefail

# Set default kubectl if not provided
KUBECTL=${KUBECTL:-kubectl}

echo "Running setup.sh"
echo "Setting up Minio provider credentials..."

# Default Minio credentials if not provided via environment
MINIO_SERVER=${MINIO_SERVER:-"minio-api.minio-system.svc.cluster.local:9000"}
MINIO_USER=${MINIO_USER:-"testuser"}
MINIO_PASSWORD=${MINIO_PASSWORD:-"testpassword123"}
MINIO_REGION=${MINIO_REGION:-"us-east-1"}

# Use provided cloud credentials or create default Minio credentials
if [ -n "${UPTEST_CLOUD_CREDENTIALS:-}" ]; then
    echo "Using provided UPTEST_CLOUD_CREDENTIALS..."
    ${KUBECTL} -n upbound-system create secret generic provider-secret --from-literal=credentials="${UPTEST_CLOUD_CREDENTIALS}" --dry-run=client -o yaml | ${KUBECTL} apply -f -
else
    echo "Creating default Minio credentials for local testing..."
    MINIO_CREDS=$(cat <<EOF
{
  "minio_server": "${MINIO_SERVER}",
  "minio_user": "${MINIO_USER}",
  "minio_password": "${MINIO_PASSWORD}",
  "minio_region": "${MINIO_REGION}"
}
EOF
)
    ${KUBECTL} -n upbound-system create secret generic provider-secret --from-literal=credentials="${MINIO_CREDS}" --dry-run=client -o yaml | ${KUBECTL} apply -f -
fi

echo "Waiting until provider is healthy..."
${KUBECTL} wait provider.pkg --all --for condition=Healthy --timeout 5m

echo "Waiting for all pods to come online..."
${KUBECTL} -n upbound-system wait --for=condition=Available deployment --all --timeout=5m

echo "Creating a default provider config..."
cat <<EOF | ${KUBECTL} apply -f -
apiVersion: template.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: provider-secret
      namespace: upbound-system
      key: credentials
EOF

# Wait for provider to be ready after configuration
echo "Waiting for provider to be ready with new configuration..."
${KUBECTL} wait provider.pkg --all --for condition=Healthy --timeout 5m
${KUBECTL} -n upbound-system wait --for=condition=Available deployment --all --timeout=5m

# Verify Minio connectivity if running in cluster
echo "Verifying Minio connectivity..."
if ${KUBECTL} get pods -n minio-system | grep -q minio; then
    echo "Minio pods found in cluster - good for testing"
else
    echo "Warning: No Minio pods found - external Minio server expected at ${MINIO_SERVER}"
fi

echo "Setup completed successfully!"

# MinIO Provider SSL Configuration Guide

This guide explains how to properly configure SSL/TLS connections for the MinIO Crossplane provider, including solutions for common ingress-related SSL issues.

## Overview

The MinIO provider supports both HTTP and HTTPS connections to MinIO servers. The key to proper configuration is understanding how the `minio_server`, `minio_ssl`, and `minio_insecure` parameters work together.

## Important: Server Format Requirements

The `minio_server` parameter **must not** include the protocol prefix (`http://` or `https://`). The provider will automatically construct the appropriate URL based on the `minio_ssl` setting.

### ✅ Correct formats:
- `localhost:9000`
- `minio.example.com:443`
- `minio.example.com` (uses default port based on SSL setting)

### ❌ Incorrect formats:
- `http://localhost:9000`
- `https://minio.example.com:443`

## Configuration Parameters

### Core SSL Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `minio_server` | string | required | Server endpoint in `host:port` or `host` format (no protocol prefix) |
| `minio_ssl` | string | `"false"` | Enable SSL/TLS connection (`"true"` or `"false"`) |
| `minio_insecure` | string | `"false"` | Skip SSL certificate verification (`"true"` or `"false"`) |

## Common Configuration Scenarios

### 1. Local Development (HTTP)

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: minio-local-secret
  namespace: upbound-system
stringData:
  credentials: |
    {
      "minio_server": "localhost:9000",
      "minio_user": "minioadmin",
      "minio_password": "minioadmin",
      "minio_region": "us-east-1",
      "minio_ssl": "false"
    }
```

### 2. Production HTTPS with Valid Certificate

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: minio-prod-secret
  namespace: upbound-system
stringData:
  credentials: |
    {
      "minio_server": "minio.example.com:443",
      "minio_user": "access-key",
      "minio_password": "secret-key",
      "minio_region": "us-east-1",
      "minio_ssl": "true",
      "minio_insecure": "false"
    }
```

### 3. HTTPS with Self-Signed Certificate

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: minio-selfsigned-secret
  namespace: upbound-system
stringData:
  credentials: |
    {
      "minio_server": "minio-internal.company.com:9000",
      "minio_user": "access-key",
      "minio_password": "secret-key",
      "minio_region": "us-east-1",
      "minio_ssl": "true",
      "minio_insecure": "true"
    }
```

### 4. Internal Cluster Service

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: minio-cluster-secret
  namespace: upbound-system
stringData:
  credentials: |
    {
      "minio_server": "minio.minio.svc.cluster.local",
      "minio_user": "access-key",
      "minio_password": "secret-key",
      "minio_region": "us-east-1",
      "minio_ssl": "false"
    }
```

## Troubleshooting Ingress SSL Issues

### Problem: "Client sent an HTTP request to an HTTPS server"

This error occurs when:
1. The ingress terminates SSL/TLS
2. The provider is configured with `minio_ssl: "false"`
3. The provider tries to make HTTP requests to an HTTPS endpoint

**Solution**: Set `minio_ssl: "true"` when connecting through HTTPS ingress:

```yaml
stringData:
  credentials: |
    {
      "minio_server": "minio-api.example.com",
      "minio_user": "access-key",
      "minio_password": "secret-key",
      "minio_region": "us-east-1",
      "minio_ssl": "true"
    }
```

### Problem: "Endpoint url cannot have fully qualified paths"

This error occurs when the server URL includes the protocol prefix.

**Solution**: Remove the protocol prefix from `minio_server`:

```yaml
# ❌ Wrong
"minio_server": "https://minio-api.example.com"

# ✅ Correct
"minio_server": "minio-api.example.com"
"minio_ssl": "true"
```

### Alternative: Use Internal Service

If you're having persistent issues with ingress SSL configuration, consider using the internal cluster service:

```yaml
stringData:
  credentials: |
    {
      "minio_server": "minio.minio.svc.cluster.local",
      "minio_ssl": "false"
    }
```

This bypasses ingress entirely and connects directly to the MinIO pods.

## Testing Your Configuration

### Environment Variables for Testing

You can test different configurations using environment variables:

```bash
# Test HTTPS configuration
export MINIO_SERVER="minio-api.example.com:443"
export MINIO_SSL="true"
export MINIO_INSECURE="false"
make e2e-minio-external

# Test internal service
export MINIO_SERVER="minio.minio.svc.cluster.local"
export MINIO_SSL="false"
make e2e-minio
```

### Validation

The provider performs connection validation during startup. Check the provider logs for validation results:

```bash
kubectl logs -n upbound-system deployment/provider-minio
```

## Migration Guide

If you're migrating from a configuration that included protocol prefixes:

1. **Remove protocol prefixes** from `minio_server`
2. **Add `minio_ssl`** parameter with appropriate value
3. **Add `minio_insecure`** if using self-signed certificates
4. **Test the connection** before deploying to production

### Example Migration

```yaml
# Before (❌ Old format)
stringData:
  credentials: |
    {
      "minio_server": "https://minio.example.com:9000",
      "minio_user": "access-key",
      "minio_password": "secret-key",
      "minio_region": "us-east-1"
    }

# After (✅ New format)
stringData:
  credentials: |
    {
      "minio_server": "minio.example.com:9000",
      "minio_user": "access-key",
      "minio_password": "secret-key",
      "minio_region": "us-east-1",
      "minio_ssl": "true",
      "minio_insecure": "false"
    }
```

## Best Practices

1. **Always use HTTPS in production** with valid certificates
2. **Use internal services** for cluster-to-cluster communication when possible
3. **Set `minio_insecure: "false"`** in production environments
4. **Test configurations** in development before deploying to production
5. **Monitor provider logs** for connection issues
6. **Use separate configs** for different environments (dev/staging/prod)
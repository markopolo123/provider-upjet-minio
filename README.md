# Provider MinIO

`provider-upjet-minio` is a [Crossplane](https://crossplane.io/) provider that is built using [Upjet](https://github.com/crossplane/upjet) code generation tools and exposes XRM-conformant managed resources for [MinIO](https://min.io/) object storage.

This provider enables you to manage MinIO resources directly from Kubernetes using Crossplane, providing a cloud-native way to provision and manage S3-compatible storage, IAM users, policies, and encryption keys.

## Features

- **10 MinIO Resources**: Complete coverage of core MinIO functionality
- **S3 Storage Management**: Buckets, objects, versioning, notifications, and policies
- **IAM Access Control**: Users, groups, policies, and service accounts
- **KMS Encryption**: Key management for server-side encryption
- **Kubernetes Native**: Full integration with Crossplane lifecycle management
- **Cross-Resource References**: Automatic dependency resolution between resources

## Supported Resources

### S3 Resources
- `Bucket` - S3-compatible buckets with ACL and lifecycle management
- `Object` - File objects with content, metadata, and versioning support  
- `BucketPolicy` - IAM policies attached to specific buckets
- `BucketVersioning` - Object versioning configuration for buckets
- `BucketNotification` - Event notifications for bucket operations

### IAM Resources  
- `User` - MinIO IAM users with credential management
- `Policy` - IAM policies defining permissions and access rules
- `Group` - User groups for organizing access permissions
- `ServiceAccount` - Service accounts for automated access

### KMS Resources
- `Key` - KMS encryption keys for server-side encryption

## Getting Started

### Prerequisites

- [Crossplane](https://crossplane.io/) v1.14+ installed in your Kubernetes cluster
- MinIO server instance (local or remote)
- `kubectl` configured to access your cluster

### Installation

Install the provider using the Crossplane CLI:

```bash
kubectl crossplane install provider xpkg.crossplane.io/markopolo123/provider-minio:latest
```

Or using declarative installation:

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-minio
spec:
  package: xpkg.crossplane.io/markopolo123/provider-minio:latest
```

### Configuration

1. **Create a Secret with MinIO credentials:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: provider-secret
  namespace: upbound-system
type: Opaque
stringData:
  credentials: |
    {
      "minio_server": "https://your-minio-endpoint:9000",
      "minio_user": "your-access-key",
      "minio_password": "your-secret-key", 
      "minio_region": "us-east-1"
    }
```

2. **Create a ProviderConfig:**

```yaml
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
```

The provider expects JSON credentials with:
- `minio_server`: Your MinIO server endpoint
- `minio_user`: Access key/username
- `minio_password`: Secret key/password  
- `minio_region`: AWS region (defaults to us-east-1)

Apply the secret first, then the ProviderConfig. Resources will automatically use the `default` ProviderConfig unless you specify otherwise.

## Usage Examples

### S3 Bucket

```yaml
apiVersion: s3.minio.crossplane.io/v1alpha1
kind: Bucket
metadata:
  name: my-storage-bucket
spec:
  forProvider:
    bucket: my-app-storage
    acl: private
    forceDestroy: true
  providerConfigRef:
    name: default
```

### IAM User

```yaml
apiVersion: iam.minio.crossplane.io/v1alpha1
kind: User
metadata:
  name: app-user
spec:
  forProvider:
    name: application-user
    forceDestroy: true
  providerConfigRef:
    name: default
```

### IAM Policy

```yaml
apiVersion: iam.minio.crossplane.io/v1alpha1
kind: Policy
metadata:
  name: s3-read-policy
spec:
  forProvider:
    name: s3-readonly-access
    policy: |
      {
        "Version": "2012-10-17",
        "Statement": [
          {
            "Effect": "Allow",
            "Action": ["s3:GetObject", "s3:ListBucket"],
            "Resource": ["arn:aws:s3:::my-app-storage", "arn:aws:s3:::my-app-storage/*"]
          }
        ]
      }
  providerConfigRef:
    name: default
```

### S3 Object

```yaml
apiVersion: s3.minio.crossplane.io/v1alpha1
kind: Object
metadata:
  name: config-file
spec:
  forProvider:
    bucketName: my-app-storage
    objectName: config/app.json
    content: '{"environment": "production"}'
    contentType: application/json
  providerConfigRef:
    name: default
```

### IAM Group

```yaml
apiVersion: iam.minio.crossplane.io/v1alpha1
kind: Group
metadata:
  name: developers
  annotations:
    crossplane.io/external-name: dev-team
spec:
  forProvider:
    forceDestroy: true
  providerConfigRef:
    name: default
```

### Bucket Versioning

```yaml
apiVersion: s3.minio.crossplane.io/v1alpha1
kind: BucketVersioning
metadata:
  name: bucket-versioning
  annotations:
    crossplane.io/external-name: my-app-storage
spec:
  forProvider:
    versioningConfiguration:
    - status: "Enabled"
      excludeFolders: false
  providerConfigRef:
    name: default
```

### Service Account

```yaml
apiVersion: iam.minio.crossplane.io/v1alpha1
kind: ServiceAccount
metadata:
  name: automated-service
spec:
  forProvider:
    targetUser: application-user
    description: "Service account for CI/CD automation"
  providerConfigRef:
    name: default
```

### KMS Key (Requires KMS-enabled MinIO)

```yaml
apiVersion: kms.minio.crossplane.io/v1alpha1
kind: Key
metadata:
  name: encryption-key
  annotations:
    crossplane.io/external-name: app-encryption-key
spec:
  forProvider: {}
  providerConfigRef:
    name: default
```

### Bucket Notification (Requires Queue Endpoint)

```yaml
apiVersion: s3.minio.crossplane.io/v1alpha1
kind: BucketNotification
metadata:
  name: bucket-events
  annotations:
    crossplane.io/external-name: my-app-storage
spec:
  forProvider:
    queue:
    - queueArn: "arn:minio:sqs::my-queue:webhook"
      events:
      - "s3:ObjectCreated:*"
      - "s3:ObjectRemoved:*"
      filterPrefix: "uploads/"
  providerConfigRef:
    name: default
```

## Resource Dependencies

Some resources depend on others existing first:

```
Bucket → BucketPolicy, BucketVersioning, BucketNotification, Object
User → ServiceAccount
```

## Cross-Resource References

The provider supports automatic reference resolution:

```yaml
apiVersion: s3.minio.crossplane.io/v1alpha1
kind: Object
metadata:
  name: referenced-object
spec:
  forProvider:
    bucketNameRef:
      name: my-storage-bucket  # References the Bucket resource
    objectName: data.json
    content: '{"key": "value"}'
  providerConfigRef:
    name: default
```

## Development

### Building the Provider

```bash
# Generate API types and CRDs
make generate

# Build provider binary and container
make build

# Run tests
make test

# Run e2e tests (requires MinIO instance)
make e2e
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make generate` to update generated code
6. Submit a pull request

### Local Development Setup

1. **Start local MinIO:**
```bash
docker run -p 9000:9000 -p 9001:9001 \
  --name minio \
  -e "MINIO_ROOT_USER=admin" \
  -e "MINIO_ROOT_PASSWORD=password123" \
  minio/minio server /data --console-address ":9001"
```

2. **Build and test provider:**
```bash
make build
make e2e MINIO_SERVER="localhost:9000" MINIO_USER="admin" MINIO_PASSWORD="password123"
```

## Configuration Reference

### Provider Configuration

| Field | Description | Required | Default |
|-------|-------------|----------|---------|
| `minio_server` | MinIO server endpoint (host:port) | Yes | - |
| `minio_user` | MinIO access key/username | Yes | - |
| `minio_password` | MinIO secret key/password | Yes | - |
| `minio_region` | MinIO region | No | us-east-1 |

### Common Annotations

| Annotation | Description | Example |
|------------|-------------|---------|
| `crossplane.io/external-name` | External resource identifier | `my-bucket-name` |
| `uptest.upbound.io/timeout` | Custom test timeout | `300s` |
| `uptest.upbound.io/conditions` | Custom ready conditions | `Ready,Synced` |

## Troubleshooting

### Common Issues

**Provider not healthy:**
```bash
kubectl get providers
kubectl describe provider provider-minio
```

**Resource stuck in creating:**
```bash
kubectl describe <resource-type> <resource-name>
kubectl logs -l pkg.crossplane.io/provider=provider-minio -n upbound-system
```

**Authentication errors:**
- Verify MinIO credentials in the secret
- Check MinIO server accessibility from cluster
- Ensure MinIO user has sufficient permissions

**KMS errors:**
- KMS resources require MinIO server with KMS configuration enabled
- Not supported in basic MinIO deployments

**Notification errors:**
- Bucket notifications require valid queue endpoints (SQS, webhook, etc.)
- ARNs must point to accessible message queues

## Compatibility

- **Crossplane**: v1.14+
- **Kubernetes**: v1.19+
- **MinIO Server**: Latest stable versions
- **Terraform Provider**: aminueza/minio v3.6.3

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Support

- [GitHub Issues](https://github.com/markopolo123/provider-upjet-minio/issues) - Bug reports and feature requests
- [Crossplane Slack](https://slack.crossplane.io/) - Community support
- [MinIO Documentation](https://docs.min.io/) - MinIO-specific questions

## Acknowledgments

- Built with [Upjet](https://github.com/crossplane/upjet) framework
- Based on [aminueza/terraform-provider-minio](https://github.com/aminueza/terraform-provider-minio)
- Follows [Crossplane](https://crossplane.io/) best practices
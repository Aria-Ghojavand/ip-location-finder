# Go App Helm Chart

This Helm chart deploys the Go IP Geolocation application to Kubernetes.

## Installation

1. **Install the chart:**
   ```bash
   helm install go-app ./helm/go-app
   ```

2. **Install with custom values:**
   ```bash
   helm install go-app ./helm/go-app -f my-values.yaml
   ```

3. **Install in a specific namespace:**
   ```bash
   helm install go-app ./helm/go-app --namespace production --create-namespace
   ```

## Configuration

### Key Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `2` |
| `image.repository` | Image repository | `go-app` |
| `image.tag` | Image tag | `latest` |
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `80` |
| `ingress.enabled` | Enable ingress | `true` |
| `ingress.hosts[0].host` | Hostname for ingress | `go-app.local` |
| `config.database.host` | Database host | `postgres-cluster-rw.production-db.svc.cluster.local` |
| `statefulSet.enabled` | Use StatefulSet instead of Deployment | `true` |
| `serviceMonitor.enabled` | Enable Prometheus ServiceMonitor | `true` |

### Database Configuration

Update the database configuration in `values.yaml`:

```yaml
config:
  database:
    host: "your-postgres-host"
    port: "5432"
    name: "your-db-name"
    user: "your-db-user"
```

### Secrets

Update base64 encoded secrets in `values.yaml`:

```yaml
secrets:
  ipstackApiKey: "base64-encoded-api-key"
  dbPassword: "base64-encoded-password"
```

To encode values:
```bash
echo -n "your-secret" | base64
```

## Usage Examples

### Deploy with custom database host:
```bash
helm install go-app ./helm/go-app \
  --set config.database.host="my-postgres.default.svc.cluster.local"
```

### Deploy with custom image:
```bash
helm install go-app ./helm/go-app \
  --set image.repository="myregistry/go-app" \
  --set image.tag="v1.0.0"
```

### Deploy as Deployment instead of StatefulSet:
```bash
helm install go-app ./helm/go-app \
  --set statefulSet.enabled=false
```

### Enable autoscaling:
```bash
helm install go-app ./helm/go-app \
  --set autoscaling.enabled=true \
  --set autoscaling.minReplicas=2 \
  --set autoscaling.maxReplicas=10
```

## Upgrading

```bash
helm upgrade go-app ./helm/go-app
```

## Uninstalling

```bash
helm uninstall go-app
```

## Testing

Run Helm tests:
```bash
helm test go-app
```

## Monitoring

The chart includes:
- Prometheus ServiceMonitor (when `serviceMonitor.enabled=true`)
- Health check endpoints
- Pod annotations for Prometheus scraping

## Features

- ✅ StatefulSet or Deployment support
- ✅ ConfigMap for configuration
- ✅ Secrets for sensitive data
- ✅ Service with metrics endpoint
- ✅ Ingress support
- ✅ Health checks (liveness, readiness, startup probes)
- ✅ Horizontal Pod Autoscaler support
- ✅ ServiceMonitor for Prometheus
- ✅ Helm tests
- ✅ ServiceAccount creation
- ✅ Resource limits and requests
- ✅ Pod annotations for monitoring
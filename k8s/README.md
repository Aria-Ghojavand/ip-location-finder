# Kubernetes Manifests for Go Application

This directory contains all the necessary Kubernetes manifests to deploy the Go application. PostgreSQL is expected to be running externally.

## Files Description

- `namespace.yaml` - Creates the go-app namespace
- `configmap.yaml` - Contains non-sensitive configuration data
- `secrets.yaml` - Contains sensitive data (API keys, passwords)
- `go-app-statefulset.yaml` - Go application StatefulSet
- `service.yaml` - Service for the application
- `ingress.yaml` - Ingress for external access
- `servicemonitor.yaml` - Prometheus ServiceMonitor for metrics

## Prerequisites

- PostgreSQL database should be running and accessible
- Update the database connection details in `configmap.yaml`

## Deployment Instructions

1. **Update database configuration** in `configmap.yaml`:
   - Set `DB_HOST` to your PostgreSQL server address
   - Update other database settings as needed

2. **Update secrets** - Before deploying, update the base64 encoded values in `secrets.yaml`:
   ```bash
   echo -n "your-actual-ipstack-api-key" | base64
   echo -n "your-actual-db-password" | base64
   ```

2. **Build and push Docker image**:
   ```bash
   docker build -t ariaghojavand/go-app:latest .
   docker push ariaghojavand/go-app:latest
   ```

3. **Update image reference** in `go-app-statefulset.yaml`:
   The image is already set to `ariaghojavand/go-app:latest`.

4. **Deploy in order**:
   ```bash
   kubectl apply -f namespace.yaml
   kubectl apply -f configmap.yaml
   kubectl apply -f secrets.yaml
   kubectl apply -f service.yaml
   kubectl apply -f go-app-statefulset.yaml
   kubectl apply -f ingress.yaml
   kubectl apply -f servicemonitor.yaml  # Only if Prometheus is installed
   ```

5. **Or deploy all at once**:
   ```bash
   kubectl apply -f k8s/
   ```

## Configuration Notes

- **Database**: Update database connection details in `configmap.yaml` to point to your existing PostgreSQL instance
- **Ingress**: Update the host in `ingress.yaml` to match your domain
- **Resources**: Adjust CPU/memory limits based on your requirements
- **Replicas**: The Go app is configured with 2 replicas for high availability
- **Health checks**: Configured for both readiness and liveness probes

## Accessing the Application

After deployment:
- Internal access: `http://go-app-service.go-app.svc.cluster.local`
- External access: Configure your DNS to point to the ingress controller IP
- Metrics: Available at `/metrics` endpoint for Prometheus scraping

## Monitoring

The application includes:
- Prometheus metrics exposed on `/metrics`
- ServiceMonitor for automatic Prometheus discovery
- Health check endpoints for Kubernetes probes
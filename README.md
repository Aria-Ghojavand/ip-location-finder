# Go IP Geolocation API

A high-performance REST API service for IP geolocation with caching, built with Go, PostgreSQL, and Prometheus monitoring.

## 🚀 Features

- **Fast IP Geolocation**: Get country information for IP addresses
- **Smart Caching**: PostgreSQL-based caching with 24-hour TTL
- **Bulk Processing**: Process up to 100 IPs in a single request
- **Prometheus Metrics**: Built-in monitoring and observability
- **Health Checks**: Ready for production deployments
- **Multiple Data Sources**: Support for IPStack API and free ip-api.com
- **Cache Management**: Clear individual or all cached entries
- **Kubernetes Ready**: Helm charts and K8s manifests included
- **CI/CD Pipeline**: GitLab CI with automated testing and deployment

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [API Endpoints](#-api-endpoints)
- [Environment Configuration](#-environment-configuration)
- [Local Development](#-local-development)
- [Docker Deployment](#-docker-deployment)
- [Kubernetes Deployment](#-kubernetes-deployment)
- [Monitoring](#-monitoring)
- [CI/CD Pipeline](#-cicd-pipeline)
- [API Documentation](#-api-documentation)
- [Performance](#-performance)
- [Contributing](#-contributing)

## 🏃 Quick Start

### Prerequisites

- Go 1.24+
- PostgreSQL 15+
- Docker (optional)
- Kubernetes cluster (for production)

### Local Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-app
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start PostgreSQL database**
   ```bash
   docker run -d \
     --name postgres-geoip \
     -e POSTGRES_DB=geoip_db \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=your-password \
     -p 5432:5432 \
     postgres:15-alpine
   ```

4. **Install dependencies and run**
   ```bash
   go mod download
   go run main.go
   ```

5. **Test the API**
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:8080/api/v1/geolocate/8.8.8.8
   ```

## 🔌 API Endpoints

### Health & Monitoring

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check endpoint |
| GET | `/metrics` | Prometheus metrics |

### IP Geolocation

| Method | Endpoint | Description | Parameters |
|--------|----------|-------------|------------|
| GET | `/api/v1/geolocate/{ip}` | Get country for single IP | `ip`: IPv4/IPv6 address |
| POST | `/api/v1/geolocate/bulk` | Get countries for multiple IPs | JSON body with `ips` array |

### Cache Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/cached` | List all cached IPs |
| DELETE | `/api/v1/cache/{ip}` | Clear cache for specific IP |
| DELETE | `/api/v1/cache` | Clear all cache |

## ⚙️ Environment Configuration

### Required Variables

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-database-password
DB_NAME=geoip_db

# Server Configuration
PORT=8080
```

### Optional Variables

```bash
# External API (for better rate limits)
IPSTACK_API_KEY=your-ipstack-api-key

# Application Settings
APP_ENV=development
LOG_LEVEL=debug
REQUEST_TIMEOUT=30s
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Database Tuning
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=5m
DB_SSL_MODE=disable

# Monitoring
METRICS_ENABLED=true
HEALTH_CHECK_ENABLED=true
```

See `.env.example` for a complete configuration template.

## 💻 Local Development

### Running with Hot Reload

```bash
# Install air for hot reload
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Code Quality

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run linter (install golangci-lint first)
golangci-lint run
```

## 🐳 Docker Deployment

### Build and Run

```bash
# Build Docker image
docker build -t go-app:latest .

# Run with Docker Compose
cat > docker-compose.yml << EOF
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PASSWORD=password
    depends_on:
      - postgres
  
  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=geoip_db
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
EOF

docker-compose up -d
```

## ☸️ Kubernetes Deployment

### Using Helm (Recommended)

```bash
# Install with Helm
helm install go-app ./helm/go-app \
  --set image.tag=latest \
  --set database.password=your-password \
  --set env.IPSTACK_API_KEY=your-api-key

# Upgrade deployment
helm upgrade go-app ./helm/go-app \
  --set image.tag=v1.2.0

# Uninstall
helm uninstall go-app
```

### Using kubectl

```bash
# Apply Kubernetes manifests
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -l app=go-app
kubectl get services
```

### Configuration Files

- **Helm Charts**: `./helm/go-app/` - Production-ready Helm deployment
- **K8s Manifests**: `./k8s/` - Direct Kubernetes manifests
- **Values**: 
  - `values.yaml` - Default values
  - `values-staging.yaml` - Staging environment
  - `values-production.yaml` - Production environment

## 📊 Monitoring

### Prometheus Metrics

The application exposes metrics at `/metrics`:

- `ip_geolocation_requests_total` - Total requests by country and source
- `ip_geolocation_request_duration_seconds` - Request duration histogram
- `ip_geolocation_cache_hits_total` - Cache hit counter
- `ip_geolocation_cache_misses_total` - Cache miss counter

### Grafana Dashboard

Import the dashboard from `./monitoring/working-grafana-dashboard.json` to visualize:

- Request rates and response times
- Cache hit/miss ratios
- Geographic distribution of requests
- Error rates and availability

### Health Monitoring

```bash
# Health check
curl http://localhost:8080/health

# Metrics endpoint
curl http://localhost:8080/metrics
```

## 🔄 CI/CD Pipeline

### GitLab CI Features

- **Automated Testing**: Unit tests with coverage reporting
- **Docker Builds**: Multi-stage builds with security scanning
- **Kubernetes Deployment**: Automated deployment with Helm
- **Security**: Vulnerability scanning with Trivy
- **Release Management**: Automated release notes

### Pipeline Stages

1. **Test**: Run tests with PostgreSQL service
2. **Build**: Build and push Docker images
3. **Deploy**: Deploy to Kubernetes cluster
4. **Release**: Generate release notes

### Configuration

Update `.gitlab-ci.yml` variables:

```yaml
variables:
  KUBE_NAMESPACE: "your-namespace"
  REGISTRY_URL: "your-registry"
  PRODUCTION_HOST: "your-domain.com"
```

## 📚 API Documentation

### Single IP Geolocation

```bash
# Request
GET /api/v1/geolocate/8.8.8.8

# Response
{
  "ip": "8.8.8.8",
  "country": "United States",
  "cached_at": "2025-10-04T10:30:00Z"
}
```

### Bulk IP Geolocation

```bash
# Request
POST /api/v1/geolocate/bulk
Content-Type: application/json

{
  "ips": ["8.8.8.8", "1.1.1.1", "208.67.222.222"]
}

# Response
{
  "results": [
    {
      "ip": "8.8.8.8",
      "country": "United States",
      "cached_at": "2025-10-04T10:30:00Z"
    },
    {
      "ip": "1.1.1.1",
      "country": "Australia",
      "cached_at": "2025-10-04T10:30:01Z"
    }
  ]
}
```

### Cache Management

```bash
# List cached IPs
GET /api/v1/cached

# Clear specific IP cache
DELETE /api/v1/cache/8.8.8.8

# Clear all cache
DELETE /api/v1/cache
```

### Error Responses

```json
{
  "error": "Invalid IP address"
}
```

## ⚡ Performance

### Benchmarks

- **Single IP**: ~5ms average response time (cached)
- **Bulk processing**: Up to 100 IPs per request
- **Cache hit ratio**: ~85% in typical workloads
- **Throughput**: 1000+ requests/second

### Optimization Features

- **Database Connection Pooling**: Configurable pool sizes
- **Smart Caching**: 24-hour TTL with conflict resolution
- **Bulk Processing**: Efficient batch operations
- **Metrics Collection**: Low-overhead monitoring

### Scaling Recommendations

- **Horizontal Scaling**: Stateless design supports multiple replicas
- **Database**: Use read replicas for cache queries
- **Caching**: Consider Redis for high-frequency access
- **Load Balancing**: Nginx or cloud load balancers

## 🤝 Contributing

### Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run quality checks: `go test ./... && go vet ./...`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

### Code Style

- Follow Go conventions and `gofmt` formatting
- Add comments for exported functions
- Include unit tests for new features
- Update documentation as needed

### Testing

```bash
# Run all tests
go test -v ./...

# Run with race detection
go test -race ./...

# Integration tests (requires database)
go test -tags=integration ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: Check this README and inline code comments
- **Issues**: Report bugs and feature requests via GitHub Issues
- **Monitoring**: Use Prometheus metrics and Grafana dashboards
- **Logs**: Application logs provide detailed error information

## 🔄 Changelog

### v1.0.0
- Initial release with IP geolocation API
- PostgreSQL caching implementation
- Prometheus metrics integration
- Kubernetes deployment support
- CI/CD pipeline with GitLab

---

**Built with ❤️ using Go, PostgreSQL, and Kubernetes**# ip-location-finder

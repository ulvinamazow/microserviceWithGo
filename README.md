# Go Microservice

A production-ready, observable, and resilient microservice built with Go.
It manages users and products, integrates with Kafka for asynchronous event processing, and exposes both REST and gRPC APIs.

---

## Features

- **Dual API support:** REST (Fiber) and gRPC
- **Authentication & Authorization:** JWT-based signup/login with bcrypt password hashing
- **Product Management:** Secure CRUD operations protected by auth middleware
- **Event-driven architecture:** Kafka producer/consumer for product-related events
- **Resilience mechanisms:**
  - Configurable retries with exponential backoff
  - Circuit breaker (CB) for database operations
- **Observability stack:**
  - OpenTelemetry tracing
  - Structured logging with Zap (trace-aware)
  - Prometheus metrics
- **Databases:** MongoDB (primary: users & products), PostgreSQL (planned / optional)
- **Containerization:** Docker & Docker Compose
- **Orchestration:** Kubernetes (Deployment, Service, HPA)
- **API Gateway:** Kong (DB-less mode)
- **CI/CD:** GitHub Actions (test → build → push → deploy)

---

## Tech Stack

| Category         | Technology                              |
|------------------|-----------------------------------------|
| Language         | Go 1.25+                                |
| HTTP Framework   | Fiber v2                                |
| gRPC             | google.golang.org/grpc                  |
| Databases        | MongoDB, PostgreSQL (optional)          |
| Messaging        | Apache Kafka (segmentio/kafka-go)       |
| Auth             | JWT (golang-jwt/jwt/v5), bcrypt         |
| Resilience       | sony/gobreaker, custom retry logic      |
| Observability    | OpenTelemetry, Zap, Prometheus          |
| Containerization | Docker, Docker Compose                  |
| Orchestration    | Kubernetes                              |
| API Gateway      | Kong (DB-less)                          |
| CI/CD            | GitHub Actions                          |

---

## Project Structure

```
.
├── main.go
├── go.mod / go.sum
├── Dockerfile
├── docker-compose.yml
├── config/
│   ├── config.yaml
│   ├── kong.yml
│   └── otel-collector-config.yaml
├── app/
│   ├── healthcheck/
│   ├── product/
│   └── user/
├── api/
│   ├── proto/
│   └── grpc/
├── pkg/
│   ├── auth/
│   ├── config/
│   ├── database/
│   ├── log/
│   ├── messaging/
│   ├── repository/
│   ├── resilience/
│   └── telemetry/
├── k8s/
│   ├── configmap.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   └── hpa.yaml
└── .github/workflows/
    └── ci-cd.yml
```

---

## Installation

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- (Optional) `protoc` for gRPC
- (Optional) Kubernetes cluster

### Local Development

**1. Clone repository**

```bash
git clone https://github.com/ulvinamazow/microserviceWithGo.git
cd microserviceWithGo
```

**2. Start dependencies**

```bash
docker compose up -d mongodb zookeeper kafka
```

**3. Configure application**

Edit `config/config.yaml` — default configuration works out-of-the-box.

**4. Run service**

```bash
go mod download
go run main.go
```

- REST API → `:8080`
- gRPC → `:50051`
- Health check → `http://localhost:8080/healthcheck`

### Docker Deployment

```bash
docker compose up --build
```

Services started:

- Go service
- MongoDB
- Kafka + Zookeeper
- Kong Gateway
- Jaeger
- Kafka UI

Gateway: `http://localhost:8000`

---

## API Endpoints

### Health & Metrics

| Method | Endpoint      | Description          |
|--------|---------------|----------------------|
| GET    | /healthcheck  | Service health status |
| GET    | /metrics      | Prometheus metrics   |

### Authentication

| Method | Endpoint      | Body                          | Response   |
|--------|---------------|-------------------------------|------------|
| POST   | /auth/signup  | `{username, email, password}` | User info  |
| POST   | /auth/login   | `{username, password}`        | JWT token  |

### Products

> Requires `Authorization: Bearer <token>`

| Method | Endpoint        | Description       |
|--------|-----------------|-------------------|
| GET    | /products       | List all products |
| GET    | /products/:id   | Get product       |
| POST   | /products       | Create product    |
| DELETE | /products/:id   | Delete product    |

---

## gRPC Usage

```bash
grpcurl -plaintext localhost:50051 list

grpcurl -plaintext -d '{"name":"book","price":19.99}' \
  localhost:50051 product.ProductService/CreateProduct

grpcurl -plaintext \
  localhost:50051 product.ProductService/ListProducts
```

---

## Configuration

Main config file: `config/config.yaml`

| Variable         | Description             |
|------------------|-------------------------|
| `port`           | HTTP port               |
| `mongodb.uri`    | MongoDB connection       |
| `kafka.brokers`  | Kafka brokers           |
| `kafka.topic`    | Event topic             |
| `auth.jwt_secret`| JWT secret              |
| `grpc.port`      | gRPC port               |


---

## Build & Run

```bash
# Build binary
CGO_ENABLED=0 go build -o server .

# Run tests
go test ./...

# Docker image
docker build -t go-microservice:latest .
```

---

## Kubernetes Deployment

```bash
# Create secret (optional)
kubectl create secret generic app-secrets \
  --from-literal=jwt-secret=your-secret

# Apply manifests
kubectl apply -f k8s/
```

Includes:

- Deployment (replicas: 2)
- Service
- ConfigMap
- HPA (2–10 pods, CPU 60%)

---

## Contributing

```bash
git checkout -b feature/amazing-feature
git commit -m "Add amazing feature"
git push origin feature/amazing-feature
```

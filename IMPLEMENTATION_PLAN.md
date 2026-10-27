# Distributed Systems E-Commerce Platform - Implementation Plan

## Project Roadmap & Objective

The goal of this project is to build and maintain a production-grade distributed microservices e-commerce platform demonstrating core distributed systems concepts:
- Service isolation with database-per-service
- Hybrid synchronous protocol: REST on edge / gRPC internally
- Decoupled asynchronous messaging with RabbitMQ AMQP 0-9-1
- Scalable deployment via Docker Compose and Kubernetes

---

## Technical Stack

- **Frontend**: React 18 (Vite), JavaScript, HTML5, CSS3 (Obsidian/Slate modern UI)
- **API Gateway**: Go, Gin Framework, gRPC Client Pool, REST-to-gRPC translation
- **Backend Services**: Go, Gin Framework (HTTP) & gRPC (Protobuf v3)
- **Service-to-Service Communication**: gRPC (`productv1.ProductServiceClient` with HTTP fallback)
- **Storage**:
  - PostgreSQL 15 (Auth, Product, Order, Payment databases)
  - Redis 7 (Cart Service RAM storage)
- **Messaging**: RabbitMQ 3 (`order_exchange` topic exchange, `order.created` routing key)
- **Infrastructure**: Docker, Docker Compose, Kubernetes manifests & Kustomize

---

## Directory Structure

```
Distributed_ecommerce/
├── ARCHITECTURE.md
├── IMPLEMENTATION_PLAN.md
├── IMPLEMENTATION_STATUS.md
├── ACCEPTANCE_TEST.md
├── README.md
├── docker-compose.yml
├── .env.example
├── go.work
├── scripts/
│   ├── init.sql
│   └── gen-proto.sh
├── proto/
│   ├── auth/v1/auth.proto
│   ├── product/v1/product.proto
│   ├── cart/v1/cart.proto
│   ├── order/v1/order.proto
│   └── payment/v1/payment.proto
├── k8s/
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secrets.yaml
│   ├── pvc.yaml
│   ├── kustomization.yaml
│   ├── deploy.sh
│   ├── teardown.sh
│   ├── databases/
│   └── services/
├── frontend/
│   ├── Dockerfile
│   ├── nginx.conf
│   └── src/
│       ├── App.jsx
│       └── index.css
└── services/
    ├── api-gateway/
    │   ├── grpcclient/
    │   ├── handler/
    │   └── middleware/
    ├── auth-service/
    │   ├── grpc/
    │   ├── handler/
    │   ├── service/
    │   └── repository/
    ├── product-service/
    │   ├── grpc/
    │   ├── handler/
    │   ├── service/
    │   └── repository/
    ├── cart-service/
    │   ├── grpc/
    │   ├── handler/
    │   ├── service/
    │   └── repository/
    ├── order-service/
    │   ├── grpc/
    │   ├── handler/
    │   ├── service/
    │   └── repository/
    └── payment-service/
        ├── grpc/
        ├── handler/
        ├── service/
        └── repository/
```

---

## Detailed Service Component Specifications

Each Go service implements an architecture with:
1. `main.go`: Entry point, concurrent Gin HTTP engine and gRPC server startup, graceful shutdown listeners.
2. `config`: Environment variable loading (ports, DB strings, AMQP credentials, gRPC addresses).
3. `grpc/server.go`: Protobuf service implementation, standard `grpc_health_v1.RegisterHealthServer`, status error codes.
4. `handler layer`: Gin request binding, dual route mounting (`/` status endpoint, `/api/*` aliases), HTTP status codes.
5. `service layer`: Core business logic, token generation, payload validation, inter-service gRPC calls.
6. `repository layer`: Data persistence using PostgreSQL SQL queries or Redis commands.
7. `health endpoint`: Dual probes: HTTP `GET /health` and gRPC `grpc.health.v1.Health/Check`.

---

## Phased Implementation Roadmap

- [x] **Phase 1**: Workspace Inspection, Architectural Specs, & Project Directory Skeleton setup.
- [x] **Phase 2**: Service Data Layer & SQL Schema Initialization (`scripts/init.sql`).
- [x] **Phase 3**: Backend Microservice Implementations (Auth, Product, Cart, Order, Payment).
- [x] **Phase 4**: API Gateway Routing & JWT Middleware Integration.
- [x] **Phase 5**: RabbitMQ Event Publisher & Consumer Integration.
- [x] **Phase 6**: Modern React Frontend Development & Visual Polish.
- [x] **Phase 7**: Docker Compose Orchestration & Verification.
- [x] **Phase 8**: Kubernetes Production Manifests & Kustomize Deployment (`k8s/`).
- [x] **Phase 9**: gRPC Protocol Buffers & Dual-Routing Implementation:
  - [x] Protocol Buffer definitions under `proto/*/v1/*.proto`
  - [x] Code generation script `scripts/gen-proto.sh`
  - [x] gRPC servers on dedicated ports (`9091`-`9095`) with Google health check
  - [x] Order Service $\rightarrow$ Product Service gRPC client with fallback
  - [x] API Gateway gRPC client pool and REST-to-gRPC translation
  - [x] Fix 404 direct-access issue via root `GET /` endpoints and route aliases
  - [x] Dual-protocol Dockerfiles, Docker Compose, and Kubernetes Service updates
  - [x] Complete test suite and comprehensive documentation

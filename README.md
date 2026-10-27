# Distributed Systems E-Commerce Platform

[![Architecture](https://img.shields.io/badge/Architecture-Microservices-blue.svg)](#architecture)
[![Language](https://img.shields.io/badge/Language-Go%201.21%20%7C%20JavaScript-00ADD8.svg)](https://golang.org)
[![Broker](https://img.shields.io/badge/Broker-RabbitMQ%203-FF6600.svg)](https://www.rabbitmq.com)
[![Storage](https://img.shields.io/badge/Databases-PostgreSQL%2015%20%7C%20Redis%207-336791.svg)](https://www.postgresql.org)
[![Orchestration](https://img.shields.io/badge/Orchestration-Docker%20Compose%20%7C%20Kubernetes-2496ED.svg)](https://kubernetes.io)
[![UI](https://img.shields.io/badge/Frontend-React%2018%20%2B%20Vite-61DAFB.svg)](https://react.dev)

A production-grade, containerized distributed e-commerce microservices platform built with Go, PostgreSQL, Redis, RabbitMQ, and React. Engineered to demonstrate real-world distributed systems principles: single responsibility microservices, database-per-service isolation, asynchronous event-driven choreography, idempotent payment settlement, unified API gateway reverse-proxying, and multi-environment orchestration across Docker Compose and Kubernetes.

---

## Table of Contents

1. [Architectural Overview](#architectural-overview)
2. [Component & Port Matrix](#component--port-matrix)
3. [Repository Structure](#repository-structure)
4. [Prerequisites](#prerequisites)
5. [Quick Start: Docker Compose](#quick-start-docker-compose)
6. [Deployment: Kubernetes (k8s)](#deployment-kubernetes-k8s)
7. [gRPC Communication Layer](#grpc-communication-layer)
8. [Direct Testing & 404 Prevention](#direct-testing--404-prevention)
9. [RabbitMQ Asynchronous Settlement Flow](#rabbitmq-asynchronous-settlement-flow)
10. [Complete REST API Reference](#complete-rest-api-reference)
11. [Frontend Design & User Interface](#frontend-design--user-interface)
12. [Database Management with pgAdmin 4](#database-management-with-pgadmin-4)
13. [Local Development in GoLand / VS Code](#local-development-in-goland--vs-code)
14. [Testing & Verification](#testing--verification)

---

## Architectural Overview

The platform decomposes e-commerce capabilities into 5 isolated backend microservices, a central API Gateway, a React single-page application, 4 dedicated PostgreSQL databases, a Redis session store, and a durable RabbitMQ message broker. It uses a **hybrid communication model**:
- **Client $\rightarrow$ API Gateway**: HTTP/REST and JSON over port `8080`.
- **API Gateway $\rightarrow$ Microservices**: Binary gRPC over dedicated internal ports (`9091`-`9095`).
- **Order Service $\rightarrow$ Product Service**: Direct inter-service gRPC over port `9092` (with automated HTTP fallback).
- **Order Service $\rightarrow$ Payment Service**: Decoupled asynchronous event-driven choreography via RabbitMQ AMQP 0-9-1 (`order.created`).

```mermaid
flowchart TD
    User["Client Browser"]
    Frontend["React SPA (Nginx)<br/>Port: 3000"]
    Gateway["API Gateway (Go / Gin)<br/>HTTP: 8080"]

    User -->|HTTP / React UI| Frontend
    User -->|REST API calls| Gateway

    subgraph Core Services ["Microservices Layer (Dual Protocol: HTTP + gRPC)"]
        Auth["Auth Service<br/>HTTP: 8081 | gRPC: 9091"]
        Product["Product Service<br/>HTTP: 8082 | gRPC: 9092"]
        Cart["Cart Service<br/>HTTP: 8083 | gRPC: 9093"]
        Order["Order Service<br/>HTTP: 8084 | gRPC: 9094"]
        Payment["Payment Service<br/>HTTP: 8085 | gRPC: 9095"]
    end

    subgraph Storage ["Databases & In-Memory Stores"]
        PG_Auth[("PostgreSQL<br/>auth_db<br/>Port: 5432")]
        PG_Prod[("PostgreSQL<br/>product_db<br/>Port: 5433")]
        Redis_DB[("Redis 7<br/>cart:{user_id}<br/>Port: 6379")]
        PG_Ord[("PostgreSQL<br/>order_db<br/>Port: 5434")]
        PG_Pay[("PostgreSQL<br/>payment_db<br/>Port: 5435")]
    end

    subgraph Messaging ["AMQP 0-9-1 Event Bus"]
        RMQ{{"RabbitMQ 3<br/>Exchange: order_exchange (topic)<br/>Routing Key: order.created"}}
    end

    %% Gateway to Microservice gRPC Translation
    Gateway -->|gRPC :9091| Auth
    Gateway -->|gRPC :9092| Product
    Gateway -->|gRPC :9093| Cart
    Gateway -->|gRPC :9094| Order
    Gateway -->|gRPC :9095| Payment

    %% Synchronous inter-service gRPC
    Order -->|gRPC :9092 Product Validation| Product

    %% Persistence links
    Auth -->|SQL| PG_Auth
    Product -->|SQL| PG_Prod
    Cart -->|Hash Ops| Redis_DB
    Order -->|ACID Transaction| PG_Ord
    Payment -->|SQL| PG_Pay

    %% Asynchronous Event links
    Order -.->|Publish 'order.created'| RMQ
    RMQ -.->|Consume 'order.created'| Payment
```

### Key Distributed Systems Patterns Implemented

- **Hybrid Communication Architecture**: Combines web-friendly REST APIs for external frontend clients with binary, multiplexed, low-latency gRPC RPCs for internal east-west service-to-service communication.
- **Database-per-Service**: Each microservice owns its datastore exclusively. Cross-service database queries are prohibited; services communicate solely via gRPC/REST APIs or asynchronous message events.
- **Event-Driven Choreography**: Order placement and payment settlement are decoupled via RabbitMQ. When an order is persisted to `order_db`, the Order Service asynchronously emits an `OrderCreated` event to the durable `order_exchange`.
- **At-Least-Once Delivery & Idempotency**: RabbitMQ messages use persistent delivery mode (delivery mode 2) and durable queues (`payment_order_created_queue`). The Payment Service processes events asynchronously, generating unique transaction identifiers.
- **Resilient Connection Lifecycle & Fallbacks**: All services connecting to RabbitMQ implement non-blocking reconnection loops. Inter-service gRPC calls (Order $\rightarrow$ Product) include automated fallback to HTTP REST endpoints.
- **Dual Routing & 404 Prevention**: All microservices expose a root `GET /` status endpoint and dual-mount routes (`/` and `/api/...`), allowing seamless testing directly or via the API Gateway.

---

## Component & Port Matrix

| Service | Technology | HTTP Port | gRPC Port | Host Port (Compose) | Storage / Dependency | Health Check |
|:---|:---|:---:|:---:|:---:|:---|:---|
| **API Gateway** | Go 1.21, Gin, gRPC Pool | `8080` | Client | `8080` | Downstream Microservices | `GET /health` |
| **Auth Service** | Go 1.21, Gin, gRPC, bcrypt, JWT | `8081` | `9091` | - | PostgreSQL (`auth_db`) | `GET /health`, gRPC Health |
| **Product Service** | Go 1.21, Gin, gRPC | `8082` | `9092` | - | PostgreSQL (`product_db`) | `GET /health`, gRPC Health |
| **Cart Service** | Go 1.21, Gin, gRPC, go-redis | `8083` | `9093` | - | Redis 7 | `GET /health`, gRPC Health |
| **Order Service** | Go 1.21, Gin, gRPC, amqp091 | `8084` | `9094` | - | PostgreSQL (`order_db`), RabbitMQ | `GET /health`, gRPC Health |
| **Payment Service** | Go 1.21, Gin, gRPC, amqp091 | `8085` | `9095` | - | PostgreSQL (`payment_db`), RabbitMQ | `GET /health`, gRPC Health |
| **React Frontend** | React 18, Vite, Nginx | `3000` | - | `3000` | API Gateway (`:8080`) | `GET /` |
| **PostgreSQL (Auth)** | PostgreSQL 15 Alpine | `5432` | - | `5432` | Volume: `postgres_auth_data` | `pg_isready` |
| **PostgreSQL (Product)** | PostgreSQL 15 Alpine | `5432` | - | `5433` | Volume: `postgres_product_data` | `pg_isready` |
| **PostgreSQL (Order)** | PostgreSQL 15 Alpine | `5432` | - | `5434` | Volume: `postgres_order_data` | `pg_isready` |
| **PostgreSQL (Payment)** | PostgreSQL 15 Alpine | `5432` | - | `5435` | Volume: `postgres_payment_data` | `pg_isready` |
| **Redis** | Redis 7 Alpine | `6379` | - | `6379` (Internal) | Volume: `redis_data` | `redis-cli ping` |
| **RabbitMQ** | RabbitMQ 3 Management | `5672` | - | `15672` (Web UI) | Volume: `rabbitmq_data` | `rabbitmq-diagnostics` |

---

## Repository Structure

```
Distributed_ecommerce/
├── docker-compose.yml              # 13-container root Compose stack
├── .env.example                    # Environment variable & port documentation
├── go.work                         # Go multi-module workspace
├── IMPLEMENTATION_STATUS.md        # Architectural readiness audit
├── ACCEPTANCE_TEST.md              # End-to-end acceptance test verification suite
├── ARCHITECTURE.md                 # Technical architecture specifications
├── pgadmin_servers.json            # One-click pgAdmin 4 connection profile
├── scripts/
│   ├── init.sql                    # Initial SQL schema
│   └── gen-proto.sh                # Protocol Buffer compiler script
├── proto/                          # Protocol Buffer definitions
│   ├── auth/v1/auth.proto
│   ├── product/v1/product.proto
│   ├── cart/v1/cart.proto
│   ├── order/v1/order.proto
│   └── payment/v1/payment.proto
├── k8s/                            # Production Kubernetes deployment
│   ├── namespace.yaml              # 'distributed-ecommerce' namespace
│   ├── configmap.yaml              # Environment variables & init SQL scripts
│   ├── secrets.yaml                # Base64 JWT and database credentials
│   ├── pvc.yaml                    # PersistentVolumeClaims for DBs and RabbitMQ
│   ├── kustomization.yaml          # Kustomize manifest bundle
│   ├── deploy.sh                   # Automated rollout script with wait conditions
│   ├── teardown.sh                 # Graceful cluster teardown script
│   ├── databases/                  # Stateful database deployments & services
│   │   ├── postgres-auth.yaml
│   │   ├── postgres-product.yaml
│   │   ├── postgres-order.yaml
│   │   ├── postgres-payment.yaml
│   │   ├── redis.yaml
│   │   └── rabbitmq.yaml
│   └── services/                   # Stateless microservice deployments & services
│       ├── auth-service.yaml
│       ├── product-service.yaml
│       ├── cart-service.yaml
│       ├── order-service.yaml
│       ├── payment-service.yaml
│       ├── gateway.yaml            # NodePort 30080
│       └── frontend.yaml           # NodePort 30000
├── frontend/                       # Modern React 18 / Vite SPA
│   ├── Dockerfile                  # Multi-stage: Node build -> Nginx Alpine
│   ├── nginx.conf                  # Single-page application router configuration
│   └── src/
│       ├── App.jsx                 # 8 full pages & state management
│       └── index.css               # Clean Obsidian/Slate design system
└── services/                       # Go Microservices
    ├── api-gateway/                # Reverse proxy, gRPC client pool & CORS router
    ├── auth-service/               # HTTP (:8081) & gRPC (:9091)
    ├── product-service/            # HTTP (:8082) & gRPC (:9092)
    ├── cart-service/               # HTTP (:8083) & gRPC (:9093)
    ├── order-service/              # HTTP (:8084) & gRPC (:9094)
    └── payment-service/            # HTTP (:8085) & gRPC (:9095)
```

---

## Prerequisites

- **Docker & Docker Compose**: Docker Desktop 4.20+ with Docker Compose v2 enabled.
- **Go**: Version 1.21+ (required only if developing or testing Go binaries locally).
- **Node.js**: Version 18+ & npm (required only if running Vite dev server locally).
- **Kubernetes CLI (`kubectl`)**: Version 1.26+ (required for Kubernetes deployment). Local clusters such as Minikube, Kind, or Docker Desktop Kubernetes are fully supported.
- **pgAdmin 4**: (Optional) For visual inspection of PostgreSQL database tables.

---

## Quick Start: Docker Compose

Docker Compose builds all Go services and the React frontend using multi-stage Dockerfiles and initializes all persistent storage and message queues.

### 1. Build and Start All 13 Containers

```bash
docker compose up --build -d
```

### 2. Verify Container Health

```bash
docker compose ps
```

All 13 containers should display a status of `Up` or `Up (healthy)`:

```
NAME               IMAGE                                   STATUS
auth-service       distributed_ecommerce-auth-service      Up
cart-service       distributed_ecommerce-cart-service      Up
frontend           distributed_ecommerce-frontend          Up (0.0.0.0:3000->3000/tcp)
gateway            distributed_ecommerce-gateway           Up (0.0.0.0:8080->8080/tcp)
order-service      distributed_ecommerce-order-service     Up
payment-service    distributed_ecommerce-payment-service   Up
postgres-auth      postgres:15-alpine                      Up (healthy)
postgres-order     postgres:15-alpine                      Up (healthy)
postgres-payment   postgres:15-alpine                      Up (healthy)
postgres-product   postgres:15-alpine                      Up (healthy)
product-service    distributed_ecommerce-product-service   Up
rabbitmq           rabbitmq:3-management-alpine            Up (healthy) (0.0.0.0:15672->15672/tcp)
redis              redis:7-alpine                          Up (healthy)
```

### 3. Access Platform Interfaces

- **Web Application**: [`http://localhost:3000`](http://localhost:3000)
- **API Gateway**: [`http://localhost:8080/health`](http://localhost:8080/health)
- **RabbitMQ Dashboard**: [`http://localhost:15672`](http://localhost:15672) (User: `guest` / Password: `guest`)

### 4. Inspect Live Logs

To observe asynchronous events flowing from the Order Service to the Payment Service:

```bash
docker compose logs -f order-service payment-service
```

### 5. Teardown

```bash
# Stop containers without removing database volumes
docker compose down

# Stop containers and remove all persistent volumes (clean slate)
docker compose down -v
```

---

## Deployment: Kubernetes (k8s)

The repository provides complete declarative Kubernetes manifests using Kustomize under `k8s/`.

### 1. Pre-build Container Images

If using Docker Desktop Kubernetes or Minikube, ensure images are built locally:

```bash
# Standard Docker Desktop Kubernetes:
docker compose build

# If using Minikube:
eval $(minikube docker-env)
docker compose build
```

### 2. Automated Deployment

Run the automated deployment script, which applies manifests and monitors rollout readiness:

```bash
chmod +x k8s/deploy.sh k8s/teardown.sh
./k8s/deploy.sh
```

Alternatively, apply using `kubectl` directly:

```bash
kubectl apply -k k8s/
```

### 3. Check Pod Status

```bash
kubectl get pods -n distributed-ecommerce
```

```
NAME                               READY   STATUS    RESTARTS   AGE
auth-service-58d799b685-2k48j      1/1     Running   0          45s
cart-service-75b48bc69-qpvxz       1/1     Running   0          45s
frontend-7d488c9cf7-4gmt9          1/1     Running   0          45s
gateway-59d89dbd99-5f2z9           1/1     Running   0          45s
order-service-67756f7bf-m2c87      1/1     Running   0          45s
payment-service-5b7975cfb8-w8xkl   1/1     Running   0          45s
postgres-auth-5645bbdd64-s6wzp     1/1     Running   0          45s
postgres-order-7c74b94f6f-hsq78    1/1     Running   0          45s
postgres-payment-97cf5ff9b-8v42j   1/1     Running   0          45s
postgres-product-6d744b79b5-t28zp  1/1     Running   0          45s
rabbitmq-75d8d4d7f5-2z4kp          1/1     Running   0          45s
redis-767dcb7db5-9wlm7             1/1     Running   0          45s
```

### 4. Accessing Services in Kubernetes

The Gateway and Frontend are exposed via Kubernetes `NodePort` services:

- **Frontend**: [`http://localhost:30000`](http://localhost:30000) (NodePort `30000`)
- **API Gateway**: [`http://localhost:30080`](http://localhost:30080) (NodePort `30080`)

Alternatively, use standard port-forwarding:

```bash
# Forward Frontend to 3000
kubectl port-forward svc/frontend 3000:3000 -n distributed-ecommerce

# Forward Gateway to 8080
kubectl port-forward svc/gateway 8080:8080 -n distributed-ecommerce
```

### 5. Kubernetes Teardown

```bash
./k8s/teardown.sh
```

---

## gRPC Communication Layer

The platform leverages gRPC over Protocol Buffers (proto3) for high-throughput, strongly typed, low-latency communication between services.

### 1. Protocol Buffers & Generation

All service contracts are defined in `proto/`:
- `proto/auth/v1/auth.proto`: `Register`, `Login`, `ValidateToken`, `GetUser`.
- `proto/product/v1/product.proto`: `ListProducts`, `GetProduct`, `CreateProduct`, `UpdateProduct`, `DeleteProduct`.
- `proto/cart/v1/cart.proto`: `GetCart`, `AddItem`, `UpdateItem`, `RemoveItem`, `ClearCart`.
- `proto/order/v1/order.proto`: `CreateOrder`, `GetOrder`, `ListOrders`.
- `proto/payment/v1/payment.proto`: `ProcessPayment`, `GetPayment`.

To regenerate gRPC Go code from the `.proto` files, run:
```bash
./scripts/gen-proto.sh
```

### 2. Standard gRPC Health Probes

All microservices register the official `grpc_health_v1.HealthServer` alongside their domain RPC services. Kubernetes and monitoring tools can probe standard gRPC health checks via:
```bash
grpc-health-probe -addr=localhost:9091
```

### 3. API Gateway Translation & Status Code Mapping

The API Gateway transparently translates incoming HTTP/REST requests to binary gRPC calls using an in-memory client connection pool (`services/api-gateway/grpcclient/client.go`). Errors returned by downstream gRPC services are mapped to proper HTTP status codes:

| gRPC Status Code | HTTP Status Code | Behavior |
|:---|:---:|:---|
| `codes.InvalidArgument` | `400 Bad Request` | Missing or malformed parameters |
| `codes.Unauthenticated` | `401 Unauthorized` | Invalid or expired JWT |
| `codes.PermissionDenied` | `403 Forbidden` | Insufficient access rights |
| `codes.NotFound` | `404 Not Found` | Entity (product, order, etc.) does not exist |
| `codes.AlreadyExists` | `409 Conflict` | Duplicate email during registration |
| `codes.Unavailable` | `503 Service Unavailable` | Downstream service down or network partition |
| `codes.Internal` | `500 Internal Server Error` | Unhandled database or system exception |

### 4. Service-to-Service gRPC Call (Order $\rightarrow$ Product)

During checkout, the Order Service must validate item availability and pricing against the Product Service. This is executed over gRPC:
- Client: `productv1.NewProductServiceClient(conn)` targeting `product-service:9092`.
- Fallback: If the gRPC call encounters a transport error, the Order Service automatically fails over to the HTTP REST endpoint (`http://product-service:8082/products/:id`), ensuring maximum availability.

---

## Direct Testing & 404 Prevention

### Root Cause of the "404 page not found" Issue

When developers test individual microservices directly (e.g. browsing `http://localhost:8081/` or issuing requests directly to a microservice's port rather than through the Gateway), Gin previously responded with:
```
404 page not found
```
This was caused by two factors:
1. Microservices did not define a root `GET /` route, so hitting the base URL in a browser resulted in a 404.
2. Route definitions inside microservices did not include `/api/...` path aliases. If an API call including the `/api/` prefix was sent directly to a microservice instead of the API Gateway, Gin could not match the path.

### Resolution Implemented

1. **Root Status JSON**: Every microservice now implements a `GET /` endpoint that returns a JSON status payload:
   ```json
   {
     "service": "auth-service",
     "status": "UP",
     "http_port": "8081",
     "grpc_port": "9091"
   }
   ```
2. **Dual-Route Mounting**: All services now mount duplicate route aliases (e.g., both `/register` AND `/api/auth/register`), enabling tests to pass cleanly whether routed through the Gateway or executed directly against the microservice.

---

## RabbitMQ Asynchronous Settlement Flow

When a user checks out, payment settlement is executed asynchronously without blocking the user response.

```mermaid
sequenceDiagram
    autonumber
    actor Client as React Client
    participant GW as API Gateway (:8080)
    participant ORD as Order Service (:8084)
    participant DB_O as PostgreSQL (order_db)
    participant RMQ as RabbitMQ (order_exchange)
    participant PAY as Payment Service (:8085)
    participant DB_P as PostgreSQL (payment_db)

    Client->>GW: POST /api/orders (Authorization: Bearer <token>)
    GW->>ORD: Forward /orders
    ORD->>ORD: Validate cart & fetch product prices
    ORD->>DB_O: BEGIN TRANSACTION<br/>INSERT INTO orders<br/>INSERT INTO order_items<br/>COMMIT
    ORD->>RMQ: Publish "OrderCreated" (Routing Key: order.created)
    ORD-->>GW: HTTP 201 Created (Order ID: 1, Status: PENDING)
    GW-->>Client: HTTP 201 Created (Order Placed)

    Note over RMQ,PAY: Asynchronous Event Processing
    RMQ->>PAY: Deliver "OrderCreated" event
    PAY->>PAY: Parse event & simulate settlement processing
    PAY->>DB_P: INSERT INTO payments (status='SUCCESS', txn_id='txn_sim_...')
    PAY->>RMQ: Acknowledge message (ack)

    Client->>GW: GET /api/payments/1
    GW->>PAY: Forward /payments/1
    PAY->>DB_P: SELECT * FROM payments WHERE order_id = 1
    PAY-->>GW: HTTP 200 OK (status: SUCCESS, txn_id: txn_sim_...)
    GW-->>Client: Final Payment Settlement Displayed
```

### Event Payload Schema

```json
{
  "event_type": "OrderCreated",
  "order_id": 1,
  "user_id": 1,
  "amount": 1299.99
}
```

### Observed Execution Log Trace

```log
order-service    | [ORDER] Order created (ID: 1, UserID: 1, Total: $1299.99)
order-service    | [ORDER] Publishing OrderCreated for Order #1 (Amount: $1299.99)
order-service    | [RABBITMQ] OrderCreated published (RoutingKey: order.created, OrderID: 1)
payment-service  | [PAYMENT] OrderCreated received
payment-service  | [PAYMENT] Processing payment for Order #1 (User: 1, Amount: $1299.99)
payment-service  | [PAYMENT] Payment SUCCESS for Order #1 (PaymentID: 1, TxnID: txn_sim_1788860869_ec7441c24cf3407e)
```

---

## Complete REST API Reference

All requests must be directed to the **API Gateway** on `http://localhost:8080`.

### 1. Health Endpoints

| Endpoint | Method | Response Sample | Description |
|:---|:---:|:---|:---|
| `/health` | `GET` | `{"service":"api-gateway","status":"UP"}` | API Gateway status |
| `/api/auth/health` | `GET` | `{"service":"auth-service","status":"UP"}` | Auth Service health |
| `/api/products/health` | `GET` | `{"service":"product-service","status":"UP"}` | Product Service health |
| `/api/cart/health` | `GET` | `{"service":"cart-service","status":"UP"}` | Cart Service health |
| `/api/orders/health` | `GET` | `{"service":"order-service","status":"UP"}` | Order Service health |
| `/api/payments/health` | `GET` | `{"service":"payment-service","status":"UP"}` | Payment Service health |

---

### 2. Authentication API (`/api/auth/*`)

#### Register User
`POST /api/auth/register`

```json
// Request Body
{
  "name": "Alex Mercer",
  "email": "alex.mercer@example.com",
  "password": "SecurePassword123!"
}

// Response: HTTP 201 Created
{
  "id": 1,
  "name": "Alex Mercer",
  "email": "alex.mercer@example.com",
  "created_at": "2026-10-13T10:15:30Z"
}
```

#### User Login
`POST /api/auth/login`

```json
// Request Body
{
  "email": "alex.mercer@example.com",
  "password": "SecurePassword123!"
}

// Response: HTTP 200 OK
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "name": "Alex Mercer",
    "email": "alex.mercer@example.com"
  }
}
```

#### Get Current User Profile
`GET /api/auth/me`  
*Header: `Authorization: Bearer <token>`*

---

### 3. Product Catalog API (`/api/products/*`)

#### List All Products
`GET /api/products`

```json
// Response: HTTP 200 OK
[
  {
    "id": 1,
    "name": "ProBook Laptop 15",
    "description": "High-performance laptop with 32GB RAM and 1TB SSD",
    "price": 1299.99,
    "stock": 45,
    "category": "Electronics",
    "image_url": "https://images.unsplash.com/photo-1517336714731-489689fd1ca8?auto=format&fit=crop&w=800&q=80"
  }
]
```

#### Get Product Details
`GET /api/products/:id`

---

### 4. Shopping Cart API (`/api/cart/*`)

*All Cart endpoints require `Authorization: Bearer <token>`.*

#### Retrieve Cart
`GET /api/cart`

```json
// Response: HTTP 200 OK
{
  "items": [
    {
      "product_id": 1,
      "name": "ProBook Laptop 15",
      "price": 1299.99,
      "quantity": 2,
      "subtotal": 2599.98
    }
  ],
  "total": 2599.98
}
```

#### Add Item to Cart
`POST /api/cart/items`

```json
// Request Body
{
  "product_id": 1,
  "quantity": 1
}
```

#### Update Item Quantity
`PUT /api/cart/items/:productId`

```json
// Request Body
{
  "quantity": 3
}
```

#### Remove Single Item
`DELETE /api/cart/items/:productId`

#### Clear Entire Cart
`DELETE /api/cart`

---

### 5. Order Management API (`/api/orders/*`)

*All Order endpoints require `Authorization: Bearer <token>`.*

#### Place Order
`POST /api/orders`

```json
// Request Body
{
  "items": [
    {
      "product_id": 1,
      "quantity": 1
    }
  ]
}

// Response: HTTP 201 Created
{
  "id": 1,
  "user_id": 1,
  "total_amount": 1299.99,
  "status": "PENDING",
  "items": [
    {
      "id": 1,
      "order_id": 1,
      "product_id": 1,
      "product_name": "ProBook Laptop 15",
      "quantity": 1,
      "unit_price": 1299.99
    }
  ],
  "created_at": "2026-10-13T10:20:00Z"
}
```

#### List User Orders
`GET /api/orders`

#### Get Order Details
`GET /api/orders/:id`

---

### 6. Payment Settlement API (`/api/payments/*`)

#### Get Payment Status by Order ID
`GET /api/payments/:orderId`

```json
// Response: HTTP 200 OK
{
  "id": 1,
  "order_id": 1,
  "user_id": 1,
  "amount": 1299.99,
  "status": "SUCCESS",
  "payment_method": "RABBITMQ_ASYNC_SIMULATED",
  "transaction_id": "txn_sim_1788860869_ec7441c24cf3407e",
  "created_at": "2026-10-13T10:20:01Z"
}
```

---

## Frontend Design & User Interface

The React frontend has been engineered using an obsidian/slate design palette, avoiding common generic AI-generated aesthetic tropes (such as neon purple gradients, huge pill buttons, excessive whitespace, and card-in-card nesting).

### Key Frontend Features

1. **Product Catalog**: High-density grid with category filter tabs, live search input, stock level badges, and direct cart add actions.
2. **Product Details View**: Comprehensive technical specs, breadcrumb navigation, and quantity counters.
3. **Shopping Cart Sidebar**: Sticky order summary with calculated tax, discount inputs, and real-time total computation.
4. **Checkout Page**: Structured 2-column layout verifying line items against order totals.
5. **Orders & Settlement History**: Status badges (`PENDING`, `SUCCESS`) and interactive order drawer.
6. **Order Details & Settlement Timeline**: Live settlement lifecycle timeline visualizing:
   - Order Placed & Persisted (PostgreSQL)
   - AMQP Message Dispatched (RabbitMQ)
   - Asynchronous Payment Captured & Cleared
   - Transaction Reference & Timestamp

---

## Database Management with pgAdmin 4

The PostgreSQL databases map to individual host ports under Docker Compose to permit simultaneous external access without collisions:

| Database | Service Name | Host Port | Internal Port | Database Name | Username | Password |
|:---|:---|:---:|:---:|:---|:---:|:---:|
| Auth DB | `postgres-auth` | `5432` | `5432` | `auth_db` | `postgres` | `password` |
| Product DB | `postgres-product` | `5433` | `5432` | `product_db` | `postgres` | `password` |
| Order DB | `postgres-order` | `5434` | `5432` | `order_db` | `postgres` | `password` |
| Payment DB | `postgres-payment` | `5435` | `5432` | `payment_db` | `postgres` | `password` |

### One-Click Connection via `pgadmin_servers.json`

A ready-to-import configuration file is provided in the repository root: `pgadmin_servers.json`.

In pgAdmin 4:
1. Open **Tools** > **Import/Export Servers**.
2. Select **Import** and choose `Distributed_ecommerce/pgadmin_servers.json`.
3. All 4 databases will automatically populate under the server group `"Distributed E-Commerce"`.

---

## Local Development in GoLand / VS Code

A multi-module Go workspace file (`go.work`) is configured in the repository root, linking all 6 Go modules:

```go
go 1.21

use (
    ./services/api-gateway
    ./services/auth-service
    ./services/cart-service
    ./services/order-service
    ./services/payment-service
    ./services/product-service
)
```

### Running Databases via Docker and Services Locally

1. Start only infrastructure containers:
   ```bash
   docker compose up -d postgres-auth postgres-product postgres-order postgres-payment redis rabbitmq
   ```
2. Open the project in GoLand or VS Code. GoLand automatically detects `go.work`.
3. Run or debug any microservice directly. For instance, to run `order-service`:
   ```bash
   cd services/order-service
   export DB_HOST=localhost DB_PORT=5434 DB_USER=postgres DB_PASSWORD=password DB_NAME=order_db
   export PRODUCT_SERVICE_URL=http://localhost:8082
   export RABBITMQ_URL=amqp://guest:guest@localhost:5672/
   export PORT=8084
   go run main.go
   ```

---

## Testing & Verification

### 1. Running Backend & gRPC Unit Tests

All 6 Go microservices and gRPC servers include unit and integration tests:

```bash
# Run tests across all services in the Go workspace
go test ./services/api-gateway/... \
        ./services/auth-service/... \
        ./services/product-service/... \
        ./services/cart-service/... \
        ./services/order-service/... \
        ./services/payment-service/...
```

To run only the API Gateway in-memory gRPC integration test suite:
```bash
go test -v ./services/api-gateway/handler/ -run TestGatewayGRPCIntegration
```

### 2. Building the Frontend

Verify that the React production bundle compiles cleanly:

```bash
cd frontend
npm install
npm run build
```

### 3. End-to-End Acceptance Verification

The complete platform has undergone full acceptance testing across all 22 scenarios, verified against live container instances. Full test execution logs and validation evidence are documented in [`ACCEPTANCE_TEST.md`](ACCEPTANCE_TEST.md).

---

## License

This project is developed for educational and academic distributed systems demonstrations. Distributed under the MIT License.

# Distributed Systems E-Commerce Platform - Implementation Plan

## Project Roadmap & Objective

The goal of this project is to build a demonstrable distributed microservices e-commerce platform (~70%+ functionality) for an academic submission. 

---

## Technical Stack

- **Frontend**: React (Vite), JavaScript, HTML5, CSS3
- **API Gateway**: Go, Gin Framework, HTTP Reverse Proxy
- **Backend Services**: Go, Gin Framework, REST APIs
- **Storage**:
  - PostgreSQL 15 (Auth, Product, Order, Payment databases)
  - Redis 7 (Cart Service RAM storage)
- **Messaging**: RabbitMQ 3 (Asynchronous `OrderCreated` processing)
- **Infrastructure**: Docker & Docker Compose

---

## Directory Structure

```
Distributed_ecommerce/
├── ARCHITECTURE.md
├── IMPLEMENTATION_PLAN.md
├── docker-compose.yml
├── scripts/
│   └── init.sql
├── frontend/
│   ├── index.html
│   ├── package.json
│   ├── vite.config.js
│   └── src/
│       ├── App.jsx
│       ├── main.jsx
│       ├── index.css
│       ├── components/
│       ├── pages/
│       └── services/
└── services/
    ├── api-gateway/
    │   ├── main.go
    │   ├── config/
    │   │   └── config.go
    │   ├── handler/
    │   │   └── proxy.go
    │   ├── middleware/
    │   │   └── auth.go
    │   └── service/
    ├── auth-service/
    │   ├── main.go
    │   ├── config/
    │   │   └── config.go
    │   ├── handler/
    │   │   └── auth_handler.go
    │   ├── service/
    │   │   └── auth_service.go
    │   └── repository/
    │       └── user_repository.go
    ├── product-service/
    │   ├── main.go
    │   ├── config/
    │   │   └── config.go
    │   ├── handler/
    │   │   └── product_handler.go
    │   ├── service/
    │   │   └── product_service.go
    │   └── repository/
    │       └── product_repository.go
    ├── cart-service/
    │   ├── main.go
    │   ├── config/
    │   │   └── config.go
    │   ├── handler/
    │   │   └── cart_handler.go
    │   ├── service/
    │   │   └── cart_service.go
    │   └── repository/
    │       └── cart_repository.go
    ├── order-service/
    │   ├── main.go
    │   ├── config/
    │   │   └── config.go
    │   ├── handler/
    │   │   └── order_handler.go
    │   ├── service/
    │   │   └── order_service.go
    │   └── repository/
    │       └── order_repository.go
    └── payment-service/
        ├── main.go
        ├── config/
        └── config.go
        ├── handler/
        │   └── payment_handler.go
        ├── service/
        │   └── payment_service.go
        └── repository/
            └── payment_repository.go
```

---

## Detailed Service Component Specifications

Each Go service implements an architecture with:
1. `main.go`: Entry point, Gin engine instantiation, route registration, graceful startup.
2. `config`: Environment variable loading (ports, DB strings, secret keys).
3. `handler layer`: Gin request binding, HTTP status codes, JSON responses.
4. `service layer`: Core business logic, token generation, payload validation, event publication/consumption.
5. `repository layer`: Data persistence using PostgreSQL SQL queries or Redis commands.
6. `health endpoint`: `/health` returning `{"status": "UP", "service": "<service-name>"}`.

---

## Phased Implementation Roadmap

- [x] **Phase 1**: Workspace Inspection, Architectural Specs, & Project Directory Skeleton setup.
- [ ] **Phase 2**: Service Data Layer & SQL Schema Initialization (`scripts/init.sql`).
- [ ] **Phase 3**: Backend Microservice Implementations (Auth, Product, Cart, Order, Payment).
- [ ] **Phase 4**: API Gateway Routing & JWT Middleware Integration.
- [ ] **Phase 5**: RabbitMQ Event Publisher & Consumer Integration.
- [ ] **Phase 6**: Modern React Frontend Development.
- [ ] **Phase 7**: Docker Compose Orchestration & Verification.

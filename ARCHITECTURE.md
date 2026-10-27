# Distributed Systems E-Commerce Platform Architecture

This document describes the high-level design, component boundaries, database schemas, messaging flows, and communication protocols (REST, gRPC, and AMQP) for the Distributed E-Commerce Platform.

---

## 1. High-Level Hybrid Architecture

The platform employs a **hybrid communication architecture**:
1. **Frontend-to-Gateway (REST / HTTP)**: The React single-page application interacts with the Go API Gateway exclusively via REST/JSON.
2. **Gateway-to-Microservices (gRPC / Protobuf)**: The API Gateway translates external HTTP requests into high-performance, strongly typed binary gRPC RPCs across dedicated backend gRPC ports.
3. **Internal Service-to-Service (gRPC with HTTP Fallback)**: Synchronous inter-service dependencies (such as Order Service validating product availability and prices against Product Service) communicate over gRPC (`productv1.ProductServiceClient`) with automatic fallback to HTTP.
4. **Asynchronous Order-to-Payment (AMQP 0-9-1)**: Asynchronous decoupled processing is handled by RabbitMQ, where `OrderCreated` events are published to a durable topic exchange and consumed by the Payment Service worker.

```
                          ┌───────────────────────────┐
                          │   React Frontend (Vite)   │
                          │   (HTTP / Client UI :3000)│
                          └─────────────┬─────────────┘
                                        │ REST / JSON (Port 8080)
                                        v
                          ┌───────────────────────────┐
                          │      Go API Gateway       │
                          │  - REST to gRPC Transl.   │
                          │  - JWT Auth Middleware    │
                          │  - CORS & Health Checks   │
                          └──────┬───┬───┬───┬────────┘
                                 │   │   │   │
          ┌──────────────────────┘   │   │   └──────────────────────┐
          │ gRPC (:9091)             │   │                          │ gRPC (:9093)
          v                          │   │ gRPC (:9092)             v
  ┌───────────────┐                  │   v                  ┌─────────────────────┐
  │  Auth/User    │                  │ ┌───────────────┐    │    Cart Service     │
  │    Service    │                  │ │ Product       │    │  (Redis Store)      │
  │  (HTTP: 8081) │                  │ │ Service       │    │  (HTTP: 8083)       │
  └───────┬───────┘                  │ │ (HTTP: 8082)  │    └──────────┬──────────┘
          │                          │ └───────┬───────┘               │
          v                          │         ▲                       v
  ┌───────────────┐                  │         │ gRPC (:9092)   ┌─────────────────────┐
  │ PostgreSQL    │                  │         │ (Inter-svc)    │    Redis Cache      │
  │  (auth_db)    │                  │         │                │ (Port 6379)         │
  └───────────────┘                  │ (Port 9094)              └─────────────────────┘
                                     v         │
                           ┌───────────────────┴───┐
                           │     Order Service     │
                           │     (HTTP: 8084)      │
                           └─────────┬─────────────┘
                                     │
                 ┌───────────────────┴───────────────────┐
                 │ PostgreSQL (order_db)                 │
                 └───────────────────────────────────────┘
                                     │
                                     │ Publish "order.created"
                                     v
                           ┌───────────────────┐
                           │     RabbitMQ      │
                           │ Exchange: orders  │
                           └─────────┬─────────┘
                                     │
                                     │ Consume "order.created"
                                     v gRPC (:9095) / HTTP (:8085)
                           ┌───────────────────┐
                           │  Payment Service  │
                           └─────────┬─────────┘
                                     │
                 ┌───────────────────┴───────────────────┐
                 │ PostgreSQL (payment_db)               │
                 └───────────────────────────────────────┘
```

---

## 2. Microservice Responsibilities & Tech Stack

| Component | Protocols Exposed | Technology Stack | Responsibilities |
| :--- | :--- | :--- | :--- |
| **Frontend** | HTTP (`:3000`) | React 18, Vite, JavaScript, CSS | Responsive UI for browsing, cart, checkout, orders, auth. |
| **API Gateway** | REST (`:8080`) | Go 1.21, Gin, gRPC Client Pool | Translates REST requests to internal gRPC calls, handles JWT auth, CORS, unified health checks. |
| **Auth Service** | HTTP (`:8081`), gRPC (`:9091`) | Go 1.21, Gin, gRPC, PostgreSQL, bcrypt | User signup, credential hashing, login, JWT token issuance, profile lookup. |
| **Product Service** | HTTP (`:8082`), gRPC (`:9092`) | Go 1.21, Gin, gRPC, PostgreSQL | Product catalog CRUD, stock queries, pricing verification for Order Service. |
| **Cart Service** | HTTP (`:8083`), gRPC (`:9093`) | Go 1.21, Gin, gRPC, Redis 7 | Fast ephemeral cart operations (add, remove, update item quantity, clear cart). |
| **Order Service** | HTTP (`:8084`), gRPC (`:9094`) | Go 1.21, Gin, gRPC, PostgreSQL, RabbitMQ | Atomic order creation, inter-service gRPC product validation, publishing `OrderCreated` events. |
| **Payment Service**| HTTP (`:8085`), gRPC (`:9095`) | Go 1.21, Gin, gRPC, PostgreSQL, RabbitMQ | Subscribes to `OrderCreated` events, processes mock settlement, payment status queries. |

---

## 3. Protocol Buffers & gRPC Architecture

All gRPC interfaces are defined under `proto/` using Protocol Buffers v3 and compiled via `scripts/gen-proto.sh`:

### 3.1 Package Hierarchy
- `proto/auth/v1/auth.proto` (`auth.v1.AuthService`)
  - `Register(RegisterRequest) -> RegisterResponse`
  - `Login(LoginRequest) -> LoginResponse`
  - `ValidateToken(ValidateTokenRequest) -> ValidateTokenResponse`
  - `GetUser(GetUserRequest) -> GetUserResponse`
- `proto/product/v1/product.proto` (`product.v1.ProductService`)
  - `ListProducts(ListProductsRequest) -> ListProductsResponse`
  - `GetProduct(GetProductRequest) -> Product`
  - `CreateProduct(CreateProductRequest) -> Product`
  - `UpdateProduct(UpdateProductRequest) -> Product`
  - `DeleteProduct(DeleteProductRequest) -> DeleteProductResponse`
- `proto/cart/v1/cart.proto` (`cart.v1.CartService`)
  - `GetCart(GetCartRequest) -> Cart`
  - `AddItem(AddItemRequest) -> Cart`
  - `UpdateItem(UpdateItemRequest) -> Cart`
  - `RemoveItem(RemoveItemRequest) -> Cart`
  - `ClearCart(ClearCartRequest) -> ClearCartResponse`
- `proto/order/v1/order.proto` (`order.v1.OrderService`)
  - `CreateOrder(CreateOrderRequest) -> Order`
  - `GetOrder(GetOrderRequest) -> Order`
  - `ListOrders(ListOrdersRequest) -> ListOrdersResponse`
- `proto/payment/v1/payment.proto` (`payment.v1.PaymentService`)
  - `ProcessPayment(ProcessPaymentRequest) -> Payment`
  - `GetPayment(GetPaymentRequest) -> Payment`

### 3.2 Standard gRPC Health Checking
Every microservice registers the official Google gRPC Health checking protocol (`grpc.health.v1.Health/Check`) alongside standard Gin HTTP health probes (`GET /health`), providing dual liveness verification.

### 3.3 Error Mapping
The API Gateway maps gRPC status codes into corresponding HTTP response codes:
- `codes.InvalidArgument` $\rightarrow$ `400 Bad Request`
- `codes.Unauthenticated` $\rightarrow$ `401 Unauthorized`
- `codes.PermissionDenied` $\rightarrow$ `403 Forbidden`
- `codes.NotFound` $\rightarrow$ `404 Not Found`
- `codes.AlreadyExists` $\rightarrow$ `409 Conflict`
- `codes.Unavailable` $\rightarrow$ `503 Service Unavailable`
- `codes.Internal` $\rightarrow$ `500 Internal Server Error`

---

## 4. Dual Routing & 404 Prevention

To ensure seamless behavior whether requests arrive from the API Gateway or developers test individual microservices directly via browser or curl:
1. **Root Status Route**: Every microservice registers a `GET /` handler returning its service name, status (`UP`), and available HTTP/gRPC ports.
2. **Dual Path Aliases**: Every microservice registers routes both with and without prefixes (e.g., `POST /register` AND `POST /auth/register` AND `POST /api/auth/register`), preventing accidental `404 page not found` errors during direct testing.

---

## 5. Database Schemas

### 5.1 Auth Service (`auth_db` - PostgreSQL)
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'customer',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### 5.2 Product Service (`product_db` - PostgreSQL)
```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    stock_quantity INT NOT NULL DEFAULT 0,
    category VARCHAR(100),
    image_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### 5.3 Cart Service (`cart_db` - Redis)
Data is stored using Redis Hash data structures:
- Key format: `cart:{user_id}`
- Field: `product_id`
- Value: JSON payload e.g. `{"product_id": 1, "name": "Laptop", "price": 999.99, "quantity": 2}`

### 5.4 Order Service (`order_db` - PostgreSQL)
```sql
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    total_amount NUMERIC(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- PENDING, PAID, FAILED, CANCELLED
    shipping_address TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(id) ON DELETE CASCADE,
    product_id INT NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    quantity INT NOT NULL
);
```

### 5.5 Payment Service (`payment_db` - PostgreSQL)
```sql
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    order_id INT UNIQUE NOT NULL,
    user_id INT NOT NULL,
    amount NUMERIC(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL, -- SUCCESS, FAILED
    payment_method VARCHAR(50) DEFAULT 'CREDIT_CARD',
    transaction_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

---

## 6. Messaging & Asynchronous Event Architecture

### RabbitMQ Topology
- **Exchange Name**: `order_exchange` (Topic Exchange)
- **Routing Key**: `order.created`
- **Queue Name**: `payment_order_created_queue`

### Message Payload (`OrderCreatedEvent`)
```json
{
  "event_id": "evt_987654321",
  "event_type": "OrderCreated",
  "timestamp": "2026-10-13T14:12:31Z",
  "data": {
    "order_id": 101,
    "user_id": 42,
    "total_amount": 1999.98,
    "items": [
      {
        "product_id": 1,
        "quantity": 2,
        "price": 999.99
      }
    ]
  }
}
```

### Workflow Execution Sequence
1. User clicks **"Checkout"** on Frontend.
2. Frontend sends `POST /api/orders` request to **API Gateway**.
3. API Gateway converts request to `order.v1.OrderService/CreateOrder` gRPC call over port `9094`.
4. Order Service validates product availability & price against **Product Service** via `product.v1.ProductService/GetProduct` over port `9092`.
5. Order Service writes Order (`status = PENDING`) & Order Items to `order_db` inside an ACID transaction.
6. Order Service publishes `OrderCreated` message to RabbitMQ exchange `order_exchange`.
7. Order Service responds to Gateway over gRPC; Gateway responds `201 Created` with JSON payload to Frontend.
8. **Payment Service** worker receives message from `payment_order_created_queue`.
9. Payment Service processes simulated settlement, creates payment record in `payment_db` with `SUCCESS`, and updates order status.

---

## 7. Security & Authentication Flow

1. User registers/logs in via `POST /api/auth/login`.
2. API Gateway invokes `auth.v1.AuthService/Login` via gRPC.
3. Auth Service checks credentials against `users` table with `bcrypt.CompareHashAndPassword`.
4. Auth Service generates a signed JWT token containing claims: `user_id`, `email`, `role`, and expiration timestamp (`exp`).
5. Token is returned to client.
6. Client attaches header `Authorization: Bearer <jwt_token>` for subsequent protected requests.
7. API Gateway intercepts requests, validates JWT signature and expiration, extracts `user_id`, and injects `X-User-ID` into outgoing gRPC metadata context.

---

## 8. Port & Network Allocation Matrix

| Service | HTTP Port | gRPC Port | Kubernetes ClusterIP | Host Port (Compose) |
|:---|:---:|:---:|:---:|:---:|
| **API Gateway** | `8080` | Client | `api-gateway:8080` | `8080` (NodePort 30080) |
| **Auth Service** | `8081` | `9091` | `auth-service:8081`, `9091` | Internal |
| **Product Service** | `8082` | `9092` | `product-service:8082`, `9092` | Internal |
| **Cart Service** | `8083` | `9093` | `cart-service:8083`, `9093` | Internal |
| **Order Service** | `8084` | `9094` | `order-service:8084`, `9094` | Internal |
| **Payment Service** | `8085` | `9095` | `payment-service:8085`, `9095` | Internal |
| **Frontend** | `3000` | - | `frontend:3000` | `3000` (NodePort 30000) |
| **PostgreSQL (Auth)** | `5432` | - | `postgres-auth:5432` | `5432` |
| **PostgreSQL (Product)** | `5432` | - | `postgres-product:5432` | `5433` |
| **PostgreSQL (Order)** | `5432` | - | `postgres-order:5432` | `5434` |
| **PostgreSQL (Payment)** | `5432` | - | `postgres-payment:5432` | `5435` |
| **Redis** | `6379` | - | `redis:6379` | Internal |
| **RabbitMQ AMQP / UI** | `5672` / `15672` | - | `rabbitmq:5672`, `15672` | `5672` / `15672` |

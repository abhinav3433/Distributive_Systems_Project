# Distributed Systems E-Commerce Platform Architecture

This document describes the high-level design, component boundaries, database schemas, messaging flows, and communication protocols for the Distributed E-Commerce Platform.

---

## 1. High-Level Architecture

```
                          ┌───────────────────────────┐
                          │   React Frontend (Vite)   │
                          │   (HTTP / Client UI)      │
                          └─────────────┬─────────────┘
                                        │ REST / JSON (Port 8080)
                                        v
                          ┌───────────────────────────┐
                          │      Go API Gateway       │
                          │   - Routing & Reverse Proxy│
                          │   - JWT Auth Middleware   │
                          │   - CORS / Logging        │
                          └──────┬───┬───┬───┬────────┘
                                 │   │   │   │
          ┌──────────────────────┘   │   │   └──────────────────────┐
          │ (Port 8081)              │   │                          │ (Port 8083)
          v                          │   │ (Port 8082)              v
  ┌───────────────┐                  │   v                  ┌─────────────────────┐
  │  Auth/User    │                  │ ┌───────────────┐    │    Cart Service     │
  │    Service    │                  │ │ Product       │    │  (Redis Store)      │
  └───────┬───────┘                  │ │ Service       │    └──────────┬──────────┘
          │                          │ └───────┬───────┘               │
          v                          │         │                       v
  ┌───────────────┐                  │         v            ┌─────────────────────┐
  │ PostgreSQL    │                  │ ┌───────────────┐    │    Redis Cache      │
  │  (auth_db)    │                  │ │ PostgreSQL    │    │ (Port 6379)         │
  └───────────────┘                  │ │ (product_db)  │    └─────────────────────┘
                                     │ └───────────────┘
                                     │ (Port 8084)
                                     v
                           ┌───────────────────┐
                           │   Order Service   │
                           └─────────┬─────────┘
                                     │
                 ┌───────────────────┴───────────────────┐
                 │ PostgreSQL (order_db)                 │
                 └───────────────────────────────────────┘
                                     │
                                     │ Publish "OrderCreated"
                                     v
                           ┌───────────────────┐
                           │     RabbitMQ      │
                           │ Exchange: orders  │
                           └─────────┬─────────┘
                                     │
                                     │ Consume "OrderCreated"
                                     v (Port 8085)
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

| Component | Technology Stack | Responsibilities |
| :--- | :--- | :--- |
| **Frontend** | React 18, Vite, JavaScript, CSS | User Interface for catalog browsing, user authentication, cart management, checkout, and order history. |
| **API Gateway** | Go, Gin Framework | Central entry point, route dispatching, reverse proxying, CORS management, and JWT verification middleware. |
| **Auth Service** | Go, Gin, PostgreSQL, bcrypt | User signup, credential hashing, login, JWT token generation, user profile lookup. |
| **Product Service** | Go, Gin, PostgreSQL | Catalog management, product listings, category filtering, inventory management. |
| **Cart Service** | Go, Gin, Redis | Fast ephemeral shopping cart storage (add item, remove item, update quantities, clear cart). |
| **Order Service** | Go, Gin, PostgreSQL, RabbitMQ | Order creation, line item storage, state management, event publication (`OrderCreated`). |
| **Payment Service**| Go, Gin, PostgreSQL, RabbitMQ | Subscribes to `OrderCreated` events, processes payment transactions (mocked), records payment results, and updates payment state. |

---

## 3. Database Schemas

### 3.1 Auth Service (`auth_db` - PostgreSQL)
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

### 3.2 Product Service (`product_db` - PostgreSQL)
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

### 3.3 Cart Service (`cart_db` - Redis)
Data is stored using Redis Hash data structures:
- Key format: `cart:{user_id}`
- Field: `product_id`
- Value: JSON payload e.g. `{"product_id": 1, "name": "Laptop", "price": 999.99, "quantity": 2}`

### 3.4 Order Service (`order_db` - PostgreSQL)
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

### 3.5 Payment Service (`payment_db` - PostgreSQL)
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

## 4. Messaging & Asynchronous Event Architecture

### RabbitMQ Topology
- **Exchange Name**: `ecom_events` (Direct / Topic Exchange)
- **Routing Key**: `order.created`
- **Queue Name**: `payment_order_created_queue`

### Message Payload (`OrderCreatedEvent`)
```json
{
  "event_id": "evt_987654321",
  "event_type": "OrderCreated",
  "timestamp": "2026-09-08T14:12:31Z",
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
2. Frontend sends `POST /api/v1/orders` request to **API Gateway**.
3. API Gateway forwards to **Order Service**.
4. Order Service verifies cart items, writes Order (`status = PENDING`) & Order Items to `order_db`.
5. Order Service publishes `OrderCreated` message to RabbitMQ exchange `ecom_events`.
6. Order Service responds `201 Created` with Order ID to Gateway & Frontend.
7. **Payment Service** worker receives message from `payment_order_created_queue`.
8. Payment Service runs payment authorization logic, inserts record into `payment_db`, and updates Order status via internal HTTP/event callbacks.

---

## 5. Security & Authentication Flow

1. User registers/logs in via `POST /api/v1/auth/login`.
2. Auth Service checks credentials against `users` table with `bcrypt.CompareHashAndPassword`.
3. Auth Service generates a signed JWT token containing claims: `user_id`, `email`, `role`, and expiration timestamp (`exp`).
4. Token returned to client in response header/body.
5. Client attaches header `Authorization: Bearer <jwt_token>` for subsequent protected API requests.
6. API Gateway intercepts requests, validates JWT signature & expiration, extracts `user_id` from claims, and injects `X-User-ID` header into downstream requests.

---

## 6. Port & Network Allocation

- `8080`: API Gateway (Host public port)
- `8081`: Auth Service (Internal / Gateway target)
- `8082`: Product Service (Internal / Gateway target)
- `8083`: Cart Service (Internal / Gateway target)
- `8084`: Order Service (Internal / Gateway target)
- `8085`: Payment Service (Internal / Gateway target)
- `3000`: Frontend React Dev Server
- `5432`: PostgreSQL
- `6379`: Redis
- `5672` / `15672`: RabbitMQ AMQP / Management Web Dashboard

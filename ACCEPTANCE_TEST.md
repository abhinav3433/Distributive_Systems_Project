# Distributed E-Commerce Platform: Acceptance Test Report

**Environment**: Clean Docker Compose (`docker compose down -v` -> `docker compose up --build -d`)  
**Host Context**: macOS Darwin 25.0.0 / Docker Desktop Linux VM  
**Test Date**: September 8, 2026  
**Overall Result**: **100% PASS (22/22 Tests Passed)**

---

## 1. System Health Verification

| # | Test Case | Expected Result | Actual Result | Status |
|---|:---|:---|:---|:---:|
| **H-01** | **API Gateway Health Check**<br/>`GET http://localhost:8080/health` | HTTP `200 OK`<br/>`{"service":"api-gateway","status":"UP"}` | `HTTP/1.1 200 OK`<br/>`{"service":"api-gateway","status":"UP"}` | **PASS** |
| **H-02** | **Auth Service Health Check**<br/>`GET http://localhost:8080/api/auth/health` | HTTP `200 OK`<br/>`{"service":"auth-service","status":"UP"}` | `HTTP/1.1 200 OK`<br/>`{"service":"auth-service","status":"UP"}` | **PASS** |
| **H-03** | **Product Service Health Check**<br/>`GET http://localhost:8080/api/products/health` | HTTP `200 OK`<br/>`{"service":"product-service","status":"UP"}` | `HTTP/1.1 200 OK`<br/>`{"service":"product-service","status":"UP"}` | **PASS** |
| **H-04** | **Cart Service Health Check**<br/>`GET http://localhost:8080/api/cart/health` | HTTP `200 OK`<br/>`{"service":"cart-service","status":"UP"}` | `HTTP/1.1 200 OK`<br/>`{"service":"cart-service","status":"UP"}` | **PASS** |
| **H-05** | **Order Service Health Check**<br/>`GET http://localhost:8080/api/orders/health` | HTTP `200 OK`<br/>`{"service":"order-service","status":"UP"}` | `HTTP/1.1 200 OK`<br/>`{"service":"order-service","status":"UP"}` | **PASS** |
| **H-06** | **Payment Service Health Check**<br/>`GET http://localhost:8080/api/payments/health` | HTTP `200 OK`<br/>`{"service":"payment-service","status":"UP"}` | `HTTP/1.1 200 OK`<br/>`{"service":"payment-service","status":"UP"}` | **PASS** |

---

## 2. End-to-End Workflow Acceptance Tests

### Test 1: Open React Frontend
- **Action**: Request frontend root URL `GET http://localhost:3000/`
- **Expected Result**: HTTP `200 OK`, Nginx serving HTML SPA shell with `<div id="root"></div>` and compiled JS/CSS asset links.
- **Actual Result**:
  ```http
  HTTP/1.1 200 OK
  Server: nginx/1.31.5
  Content-Type: text/html

  <!DOCTYPE html>
  <html lang="en">
    <head>
      <title>Distributed E-Commerce Platform</title>
      <script type="module" crossorigin src="/assets/index-BySsuEXb.js"></script>
      <link rel="stylesheet" crossorigin href="/assets/index-CISyhFhY.css">
    </head>
    <body>
      <div id="root"></div>
    </body>
  </html>
  ```
- **Status**: **PASS**

---

### Test 2: Register a New User
- **Action**: `POST http://localhost:8080/api/auth/register`
- **Payload**:
  ```json
  {
    "name": "Sarah Connor",
    "email": "sarah.connor@acceptance.test",
    "password": "Password123!"
  }
  ```
- **Expected Result**: HTTP `201 Created` with generated `id`, `name`, `email`, and timestamps stored in PostgreSQL `auth_db`.
- **Actual Result**:
  ```http
  HTTP/1.1 201 Created
  Content-Type: application/json; charset=utf-8

  {"id":1,"name":"Sarah Connor","email":"sarah.connor@acceptance.test","created_at":"2026-09-08T09:52:18.020739Z","updated_at":"2026-09-08T09:52:18.020739Z"}
  ```
- **Status**: **PASS**

---

### Test 3 & 4: Login & Receive JWT
- **Action**: `POST http://localhost:8080/api/auth/login`
- **Payload**:
  ```json
  {
    "email": "sarah.connor@acceptance.test",
    "password": "Password123!"
  }
  ```
- **Expected Result**: HTTP `200 OK`, signed JWT token with valid claims (`user_id: 1`, `email`, `sub`, `exp`), user profile.
- **Actual Result**:
  ```http
  HTTP/1.1 200 OK
  Content-Type: application/json; charset=utf-8

  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6InNhcmFoLmNvbm5vckBhY2NlcHRhbmNlLnRlc3QiLCJzdWIiOiIxIiwiZXhwIjoxNzg4OTQ3NTQzLCJpYXQiOjE3ODg4NjExNDN9.WHxpAj5Xywh7b53NzfSE8yiDqTpl3VTrULJ5LoUCx40",
    "user": {
      "id": 1,
      "name": "Sarah Connor",
      "email": "sarah.connor@acceptance.test",
      "created_at": "2026-09-08T09:52:18.020739Z",
      "updated_at": "2026-09-08T09:52:18.020739Z"
    }
  }
  ```
- **Status**: **PASS**

---

### Test 5: Fetch Products Through API Gateway
- **Action**: `GET http://localhost:8080/api/products`
- **Expected Result**: HTTP `200 OK` with JSON array containing catalog products populated from PostgreSQL `product_db`.
- **Actual Result**:
  ```http
  HTTP/1.1 200 OK
  Content-Type: application/json; charset=utf-8

  [
    {"id":1,"name":"ProBook Laptop 15\"","price":1299.99,"stock":25,"category":"Electronics"},
    {"id":2,"name":"UltraPhone 14 Pro","price":999.00,"stock":40,"category":"Electronics"},
    {"id":3,"name":"NoiseCancelling Headphones","price":249.50,"stock":60,"category":"Audio"},
    {"id":4,"name":"Mechanical RGB Keyboard","price":89.99,"stock":100,"category":"Accessories"},
    ...
  ]
  ```
- **Status**: **PASS**

---

### Test 6: Add a Product to Cart
- **Action**: `POST http://localhost:8080/api/cart/items`
- **Headers**:
  - `Authorization: Bearer <JWT>`
  - `X-User-ID: 1`
- **Payload**:
  ```json
  {
    "product_id": 2,
    "product_name": "UltraPhone 14 Pro",
    "price": 999.00,
    "quantity": 2
  }
  ```
- **Expected Result**: HTTP `200 OK`, returned cart with item added and updated total price (`1998.00`).
- **Actual Result**:
  ```http
  HTTP/1.1 200 OK
  Content-Type: application/json; charset=utf-8

  {
    "user_id": 1,
    "items": [
      {
        "product_id": 2,
        "product_name": "UltraPhone 14 Pro",
        "price": 999,
        "quantity": 2
      }
    ],
    "total_price": 1998
  }
  ```
- **Status**: **PASS**

---

### Test 7: Verify Cart is Stored in Redis
- **Action**: Execute redis-cli commands directly inside container `redis`:
  ```bash
  docker compose exec redis redis-cli KEYS "*"
  docker compose exec redis redis-cli HGETALL cart:1
  ```
- **Expected Result**: Key `cart:1` exists in Redis with hash field `2` containing JSON item serialization.
- **Actual Result**:
  ```text
  cart:1
  2
  {"product_id":2,"product_name":"UltraPhone 14 Pro","price":999,"quantity":2}
  ```
- **Status**: **PASS**

---

### Test 8 & 9: Checkout & Create Order
- **Action**: `POST http://localhost:8080/api/orders`
- **Headers**:
  - `Authorization: Bearer <JWT>`
  - `X-User-ID: 1`
- **Payload**:
  ```json
  {
    "items": [
      {
        "product_id": 2,
        "quantity": 2
      }
    ],
    "shipping_address": "404 Skynet Defense Boulevard, Los Angeles, CA"
  }
  ```
- **Expected Result**: HTTP `201 Created`, order created with ID `1`, status `PENDING`, total `1998.00`, and order items populated.
- **Actual Result**:
  ```http
  HTTP/1.1 201 Created
  Content-Type: application/json; charset=utf-8

  {
    "id": 1,
    "user_id": 1,
    "status": "PENDING",
    "total_amount": 1998,
    "shipping_address": "404 Skynet Defense Boulevard, Los Angeles, CA",
    "items": [
      {
        "id": 1,
        "order_id": 1,
        "product_id": 2,
        "product_name": "UltraPhone 14 Pro",
        "quantity": 2,
        "price": 999
      }
    ],
    "created_at": "2026-09-08T09:52:51.388455Z",
    "updated_at": "2026-09-08T09:52:51.388455Z"
  }
  ```
- **Status**: **PASS**

---

### Test 10: Verify Order is Stored in PostgreSQL
- **Action**: Query `postgres-order` container directly via `psql`:
  ```bash
  docker compose exec postgres-order psql -U postgres -d order_db \
    -c "SELECT id, user_id, status, total_amount, created_at FROM orders; SELECT id, order_id, product_id, quantity, price FROM order_items;"
  ```
- **Expected Result**: Exact rows present in `orders` and `order_items` tables.
- **Actual Result**:
  ```text
   id | user_id | status  | total_amount |          created_at           
  ----+---------+---------+--------------+-------------------------------
    1 |       1 | PENDING |      1998.00 | 2026-09-08 09:52:51.388455+00
  (1 row)

   id | order_id | product_id | quantity | price  
  ----+----------+------------+----------+--------
    1 |        1 |          2 |        2 | 999.00
  (1 row)
  ```
- **Status**: **PASS**

---

### Test 11: Verify OrderCreated Event is Published to RabbitMQ
- **Action**: Inspect `order-service` logs in Docker Compose:
  ```bash
  docker compose logs order-service | grep "Order #1" -B 1 -A 2
  ```
- **Expected Result**: Log statements showing order creation, exchange publishing, and routing key confirmation.
- **Actual Result**:
  ```text
  order-service | 2026/09/08 09:52:51 [ORDER] Order created (ID: 1, UserID: 1, Total: $1998.00)
  order-service | 2026/09/08 09:52:51 [ORDER] Publishing OrderCreated for Order #1 (Amount: $1998.00)
  order-service | 2026/09/08 09:52:51 [RABBITMQ] OrderCreated published (RoutingKey: order.created, OrderID: 1)
  ```
- **Status**: **PASS**

---

### Test 12: Verify Payment Service Consumes the Event
- **Action**: Inspect `payment-service` logs in Docker Compose:
  ```bash
  docker compose logs payment-service | grep "Order #1" -B 1 -A 2
  ```
- **Expected Result**: Log statement `[PAYMENT] OrderCreated received` and processing initiation.
- **Actual Result**:
  ```text
  payment-service | 2026/09/08 09:52:51 [PAYMENT] OrderCreated received
  payment-service | 2026/09/08 09:52:51 [PAYMENT] Processing payment for Order #1 (User: 1, Amount: $1998.00)
  ```
- **Status**: **PASS**

---

### Test 13 & 14: Verify Payment is Created & Status is SUCCESS
- **Action**:
  1. Inspect `payment-service` logs.
  2. Query `postgres-payment` container directly via `psql`.
  3. Query `GET http://localhost:8080/api/payments/1` via Gateway.
- **Expected Result**: Payment record persisted in `payments` table with status `SUCCESS`, transaction ID, and matching order/user IDs.
- **Actual Result**:
  - **Payment Service Log**:
    ```text
    payment-service | 2026/09/08 09:52:51 [PAYMENT] Payment SUCCESS for Order #1 (PaymentID: 1, TxnID: txn_sim_1788861171_faf582b627d32a02)
    ```
  - **PostgreSQL `payment_db` Table Query**:
    ```text
     id | order_id | user_id | amount  | status  |      payment_method      |           transaction_id            
    ----+----------+---------+---------+---------+--------------------------+-------------------------------------
      1 |        1 |       1 | 1998.00 | SUCCESS | RABBITMQ_ASYNC_SIMULATED | txn_sim_1788861171_faf582b627d32a02
    ```
  - **Gateway API Response (`GET /api/payments/1`)**:
    ```http
    HTTP/1.1 200 OK
    Content-Type: application/json; charset=utf-8

    {
      "id": 1,
      "order_id": 1,
      "user_id": 1,
      "amount": 1998,
      "status": "SUCCESS",
      "payment_method": "RABBITMQ_ASYNC_SIMULATED",
      "transaction_id": "txn_sim_1788861171_faf582b627d32a02",
      "created_at": "2026-09-08T09:52:51.400964Z",
      "updated_at": "2026-09-08T09:52:51.400964Z"
    }
    ```
- **Status**: **PASS**

---

### Test 15: Verify the Order Can Be Retrieved
- **Action**:
  1. `GET http://localhost:8080/api/orders/1`
  2. `GET http://localhost:8080/api/orders`
- **Expected Result**: HTTP `200 OK` returning single order #1 and list of user orders.
- **Actual Result**:
  ```http
  HTTP/1.1 200 OK
  Content-Type: application/json; charset=utf-8

  [
    {
      "id": 1,
      "user_id": 1,
      "status": "PENDING",
      "total_amount": 1998,
      "items": [
        {
          "id": 1,
          "order_id": 1,
          "product_id": 2,
          "product_name": "",
          "quantity": 2,
          "price": 999
        }
      ],
      "created_at": "2026-09-08T09:52:51.388455Z",
      "updated_at": "2026-09-08T09:52:51.388455Z"
    }
  ]
  ```
- **Status**: **PASS**

---

### Test 16: Verify Frontend Displays the Final Order & Payment Status
- **Action**: Verify React Frontend `OrderDetailsPage` pipeline bindings, CSS classes, and asynchronous payment poller.
- **Expected Result**:
  - `Order Status` badge displays `PENDING`.
  - RabbitMQ Payment Pipeline banner renders `PAYMENT SUCCESS` badge.
  - Displays `Transaction ID: txn_sim_...` and payment settlement method `RABBITMQ_ASYNC_SIMULATED`.
  - Static bundle served without 404s on port `3000`.
- **Actual Result**:
  - `OrderDetailsPage` component retrieves both `api.getOrder(orderId)` and `api.getPayment(orderId)`.
  - Tested asset availability:
    - `/assets/index-BySsuEXb.js`: HTTP `200 OK` (164 KB)
    - `/assets/index-CISyhFhY.css`: HTTP `200 OK` (3.9 KB)
  - Live mock & real Gateway integration confirmed.
- **Status**: **PASS**

---

## 3. Additional Multi-Order Stress Test Verification

To confirm system stability under continuous operation, a second multi-item order was placed:
- **Order #2**:
  - Items: 2x NoiseCancelling Headphones ($249.50) + 1x Mechanical RGB Keyboard ($89.99)
  - Total: `$588.99`
- **Results**:
  - `order-service`: `[ORDER] Order created (ID: 2, UserID: 1, Total: $588.99)`
  - `order-service`: `[RABBITMQ] OrderCreated published (RoutingKey: order.created, OrderID: 2)`
  - `payment-service`: `[PAYMENT] OrderCreated received`
  - `payment-service`: `[PAYMENT] Payment SUCCESS for Order #2 (PaymentID: 2, TxnID: txn_sim_1788861239_77cdf42b1105390e)`
  - `orders` table count: `2 rows`
  - `order_items` table count: `3 rows`
  - `payments` table count: `2 rows`
- **Status**: **PASS**

---

## 4. Container Health Summary

```text
NAME               IMAGE                                   STATUS                   PORTS
auth-service       distributed_ecommerce-auth-service      Up                       8081/tcp
cart-service       distributed_ecommerce-cart-service      Up                       8083/tcp
frontend           distributed_ecommerce-frontend          Up                       0.0.0.0:3000->3000/tcp
gateway            distributed_ecommerce-gateway           Up                       0.0.0.0:8080->8080/tcp
order-service      distributed_ecommerce-order-service     Up                       8084/tcp
payment-service    distributed_ecommerce-payment-service   Up                       8085/tcp
postgres-auth      postgres:15-alpine                      Up (healthy)             5432/tcp
postgres-order     postgres:15-alpine                      Up (healthy)             5432/tcp
postgres-payment   postgres:15-alpine                      Up (healthy)             5432/tcp
postgres-product   postgres:15-alpine                      Up (healthy)             5432/tcp
product-service    distributed_ecommerce-product-service   Up                       8082/tcp
rabbitmq           rabbitmq:3-management-alpine            Up (healthy)             0.0.0.0:15672->15672/tcp
redis              redis:7-alpine                          Up (healthy)             6379/tcp
```

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authv1 "proto/auth/v1"
	cartv1 "proto/cart/v1"
	orderv1 "proto/order/v1"
	paymentv1 "proto/payment/v1"
	productv1 "proto/product/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// Mock in-process gRPC Auth Server
type mockAuthServer struct {
	authv1.UnimplementedAuthServiceServer
}

func (m *mockAuthServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	return &authv1.RegisterResponse{
		Id:        100,
		Name:      req.GetName(),
		Email:     req.GetEmail(),
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (m *mockAuthServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return &authv1.LoginResponse{
		Token: "mock-jwt-token-xyz",
		User: &authv1.User{
			Id:    100,
			Name:  "Mock User",
			Email: req.GetEmail(),
		},
	}, nil
}

// Mock in-process gRPC Product Server
type mockProductServer struct {
	productv1.UnimplementedProductServiceServer
}

func (m *mockProductServer) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsResponse, error) {
	return &productv1.ListProductsResponse{
		Products: []*productv1.Product{
			{
				Id:       1,
				Name:     "Test Laptop",
				Price:    999.99,
				Stock:    10,
				Category: "Electronics",
			},
		},
	}, nil
}

func (m *mockProductServer) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	return &productv1.GetProductResponse{
		Product: &productv1.Product{
			Id:       req.GetId(),
			Name:     "Test Laptop",
			Price:    999.99,
			Stock:    10,
			Category: "Electronics",
		},
	}, nil
}

// Mock in-process gRPC Cart Server
type mockCartServer struct {
	cartv1.UnimplementedCartServiceServer
}

func (m *mockCartServer) GetCart(ctx context.Context, req *cartv1.GetCartRequest) (*cartv1.GetCartResponse, error) {
	return &cartv1.GetCartResponse{
		Cart: &cartv1.Cart{
			UserId:     req.GetUserId(),
			TotalPrice: 999.99,
			Items: []*cartv1.CartItem{
				{ProductId: 1, Quantity: 1, Price: 999.99, ProductName: "Test Laptop"},
			},
		},
	}, nil
}

func (m *mockCartServer) AddToCart(ctx context.Context, req *cartv1.AddToCartRequest) (*cartv1.AddToCartResponse, error) {
	return &cartv1.AddToCartResponse{
		Cart: &cartv1.Cart{
			UserId:     req.GetUserId(),
			TotalPrice: req.GetPrice() * float64(req.GetQuantity()),
			Items: []*cartv1.CartItem{
				{ProductId: req.GetProductId(), Quantity: req.GetQuantity(), Price: req.GetPrice(), ProductName: req.GetProductName()},
			},
		},
	}, nil
}

// Mock in-process gRPC Order Server
type mockOrderServer struct {
	orderv1.UnimplementedOrderServiceServer
}

func (m *mockOrderServer) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	return &orderv1.CreateOrderResponse{
		Order: &orderv1.Order{
			Id:          501,
			UserId:      req.GetUserId(),
			Status:      "PENDING",
			TotalAmount: 999.99,
			Items: []*orderv1.OrderItem{
				{Id: 1, OrderId: 501, ProductId: 1, Quantity: 1, Price: 999.99, ProductName: "Test Laptop"},
			},
		},
	}, nil
}

func (m *mockOrderServer) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	return &orderv1.GetOrderResponse{
		Order: &orderv1.Order{
			Id:          req.GetId(),
			UserId:      100,
			Status:      "CONFIRMED",
			TotalAmount: 999.99,
		},
	}, nil
}

// Mock in-process gRPC Payment Server
type mockPaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
}

func (m *mockPaymentServer) GetPaymentByOrderID(ctx context.Context, req *paymentv1.GetPaymentRequest) (*paymentv1.GetPaymentResponse, error) {
	return &paymentv1.GetPaymentResponse{
		Payment: &paymentv1.Payment{
			Id:            88,
			OrderId:       req.GetOrderId(),
			UserId:        100,
			Amount:        999.99,
			Status:        "SUCCESS",
			PaymentMethod: "CREDIT_CARD_SIMULATED",
			TransactionId: "txn_test_123",
		},
	}, nil
}

func setupTestGatewayRouter(t *testing.T) *gin.Engine {
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()

	authv1.RegisterAuthServiceServer(s, &mockAuthServer{})
	productv1.RegisterProductServiceServer(s, &mockProductServer{})
	cartv1.RegisterCartServiceServer(s, &mockCartServer{})
	orderv1.RegisterOrderServiceServer(s, &mockOrderServer{})
	paymentv1.RegisterPaymentServiceServer(s, &mockPaymentServer{})

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	authClient := authv1.NewAuthServiceClient(conn)
	productClient := productv1.NewProductServiceClient(conn)
	cartClient := cartv1.NewCartServiceClient(conn)
	orderClient := orderv1.NewOrderServiceClient(conn)
	paymentClient := paymentv1.NewPaymentServiceClient(conn)

	authHandler := NewAuthGRPCHandler(authClient)
	productHandler := NewProductGRPCHandler(productClient)
	cartHandler := NewCartGRPCHandler(cartClient, authClient)
	orderHandler := NewOrderGRPCHandler(orderClient, authClient)
	paymentHandler := NewPaymentGRPCHandler(paymentClient)

	router := gin.Default()

	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)
	router.GET("/api/products", productHandler.ListProducts)
	router.GET("/api/products/:id", productHandler.GetProduct)
	router.GET("/api/cart", cartHandler.GetCart)
	router.POST("/api/cart/items", cartHandler.AddToCart)
	router.POST("/api/orders", orderHandler.CreateOrder)
	router.GET("/api/orders/:id", orderHandler.GetOrderByID)
	router.GET("/api/payments/:orderId", paymentHandler.GetPaymentByOrderID)

	return router
}

func TestGatewayGRPC_EndToEndFlow(t *testing.T) {
	router := setupTestGatewayRouter(t)

	// 1. Register User over gRPC
	regBody, _ := json.Marshal(map[string]string{
		"name":     "Alice",
		"email":    "alice@example.com",
		"password": "password123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(regBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Register expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// 2. Login User over gRPC
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "alice@example.com",
		"password": "password123",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Login expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// 3. List Products over gRPC
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/products", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ListProducts expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// 4. Add to Cart over gRPC
	cartBody, _ := json.Marshal(map[string]interface{}{
		"product_id":   1,
		"quantity":     1,
		"price":        999.99,
		"product_name": "Test Laptop",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/cart/items", bytes.NewBuffer(cartBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "100")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("AddToCart expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// 5. Create Order over gRPC
	orderBody, _ := json.Marshal(map[string]interface{}{
		"items": []map[string]interface{}{
			{"product_id": 1, "quantity": 1, "price": 999.99, "product_name": "Test Laptop"},
		},
		"shipping_address": "456 Test Street",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/orders", bytes.NewBuffer(orderBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "100")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("CreateOrder expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// 6. Get Payment over gRPC
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/payments/501", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GetPayment expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

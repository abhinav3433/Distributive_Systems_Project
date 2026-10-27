package grpc

import (
	"context"
	"testing"

	"order-service/messaging"
	"order-service/repository"
	"order-service/service"
	orderv1 "proto/order/v1"
)

func setupTestOrderGRPC() (*OrderGRPCServer, service.OrderService) {
	repo := repository.NewMockOrderRepository()
	pub := messaging.NewMockEventPublisher()
	svc := service.NewOrderService(repo, pub, "")
	return NewOrderGRPCServer(svc), svc
}

func TestOrderGRPC_CreateAndGet(t *testing.T) {
	server, _ := setupTestOrderGRPC()
	ctx := context.Background()

	// Create Order
	req := &orderv1.CreateOrderRequest{
		UserId: 10,
		Items: []*orderv1.OrderItemInput{
			{
				ProductId:   1,
				Quantity:    2,
				Price:       50.0,
				ProductName: "Keyboard",
			},
			{
				ProductId:   2,
				Quantity:    1,
				Price:       25.0,
				ProductName: "Mouse",
			},
		},
		ShippingAddress: "123 Distributed Way",
	}

	createResp, err := server.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	ord := createResp.GetOrder()
	if ord.GetId() <= 0 {
		t.Errorf("expected positive order id, got %d", ord.GetId())
	}
	if ord.GetTotalAmount() != 125.0 {
		t.Errorf("expected total 125.0, got %f", ord.GetTotalAmount())
	}
	if len(ord.GetItems()) != 2 {
		t.Errorf("expected 2 items, got %d", len(ord.GetItems()))
	}

	// Get Order By ID
	getResp, err := server.GetOrder(ctx, &orderv1.GetOrderRequest{Id: ord.GetId()})
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	if getResp.GetOrder().GetTotalAmount() != 125.0 {
		t.Errorf("expected total 125.0, got %f", getResp.GetOrder().GetTotalAmount())
	}

	// List Orders By User
	listResp, err := server.ListOrders(ctx, &orderv1.ListOrdersRequest{UserId: 10})
	if err != nil {
		t.Fatalf("ListOrders failed: %v", err)
	}
	if len(listResp.GetOrders()) != 1 {
		t.Errorf("expected 1 order, got %d", len(listResp.GetOrders()))
	}
}

func TestOrderGRPC_CreateWithStructuredAddress(t *testing.T) {
	server, _ := setupTestOrderGRPC()
	ctx := context.Background()

	req := &orderv1.CreateOrderRequest{
		UserId: 25,
		Items: []*orderv1.OrderItemInput{
			{
				ProductId:   5,
				Quantity:    1,
				Price:       49.99,
				ProductName: "Wireless Ergonomic Mouse",
			},
		},
		Address: &orderv1.ShippingAddress{
			RecipientName: "Jane Smith",
			PhoneNumber:   "+1 555-432-1098",
			AddressLine1:  "500 Cloud Parkway",
			AddressLine2:  "Floor 12",
			City:          "Seattle",
			State:         "WA",
			PostalCode:    "98101",
			Country:       "United States",
		},
	}

	createResp, err := server.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("CreateOrder with Address failed: %v", err)
	}

	ord := createResp.GetOrder()
	if ord.GetAddress() == nil {
		t.Fatalf("expected attached proto Address, got nil")
	}
	if ord.GetAddress().GetRecipientName() != "Jane Smith" {
		t.Errorf("expected recipient Jane Smith, got %s", ord.GetAddress().GetRecipientName())
	}
	if ord.GetAddress().GetCity() != "Seattle" {
		t.Errorf("expected city Seattle, got %s", ord.GetAddress().GetCity())
	}
	if ord.GetShippingAddress() == "" {
		t.Errorf("expected formatted shipping summary string, got empty")
	}

	// Verify GetOrder returns structured address
	getResp, err := server.GetOrder(ctx, &orderv1.GetOrderRequest{Id: ord.GetId()})
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	if getResp.GetOrder().GetAddress() == nil {
		t.Fatalf("expected GetOrder to return Address, got nil")
	}
	if getResp.GetOrder().GetAddress().GetPostalCode() != "98101" {
		t.Errorf("expected postal code 98101, got %s", getResp.GetOrder().GetAddress().GetPostalCode())
	}
}

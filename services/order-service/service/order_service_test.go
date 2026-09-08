package service

import (
	"context"
	"errors"
	"testing"

	"order-service/messaging"
	"order-service/model"
	"order-service/repository"
)

func setupTestOrderService() (OrderService, *messaging.MockEventPublisher) {
	repo := repository.NewMockOrderRepository()
	mockPublisher := messaging.NewMockEventPublisher()
	return NewOrderService(repo, mockPublisher, ""), mockPublisher
}

func TestCreateOrderAndCalculateTotal(t *testing.T) {
	svc, mockPublisher := setupTestOrderService()
	ctx := context.Background()
	userID := 10

	req := model.CreateOrderRequest{
		Items: []model.OrderItemRequest{
			{
				ProductID:   1,
				ProductName: "ProBook Laptop 15\"",
				Price:       1299.99,
				Quantity:    2,
			},
			{
				ProductID:   4,
				ProductName: "Mechanical RGB Keyboard",
				Price:       89.99,
				Quantity:    1,
			},
		},
		ShippingAddress: "123 Distributed Way, Cloud City",
	}

	order, err := svc.CreateOrder(ctx, userID, req)
	if err != nil {
		t.Fatalf("Expected no error creating order, got: %v", err)
	}

	if order.ID == 0 {
		t.Errorf("Expected non-zero order ID")
	}
	if order.UserID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, order.UserID)
	}
	if order.Status != model.StatusPending {
		t.Errorf("Expected initial status PENDING, got %s", order.Status)
	}

	expectedTotal := (1299.99 * 2) + (89.99 * 1)
	if order.TotalAmount != expectedTotal {
		t.Errorf("Expected total amount %f, got %f", expectedTotal, order.TotalAmount)
	}

	if len(order.Items) != 2 {
		t.Fatalf("Expected 2 order items, got %d", len(order.Items))
	}

	// Verify event was published
	if len(mockPublisher.PublishedEvents) != 1 {
		t.Fatalf("Expected 1 published event, got %d", len(mockPublisher.PublishedEvents))
	}
	event := mockPublisher.PublishedEvents[0]
	if event.OrderID != order.ID {
		t.Errorf("Expected event OrderID %d, got %d", order.ID, event.OrderID)
	}
	if event.Amount != expectedTotal {
		t.Errorf("Expected event amount %f, got %f", expectedTotal, event.Amount)
	}
}

func TestCreateOrderValidation(t *testing.T) {
	svc, _ := setupTestOrderService()
	ctx := context.Background()

	// 1. Invalid User ID
	_, err := svc.CreateOrder(ctx, 0, model.CreateOrderRequest{
		Items: []model.OrderItemRequest{{ProductID: 1, Quantity: 1, Price: 10}},
	})
	if err == nil {
		t.Errorf("Expected error for zero user ID, got nil")
	}

	// 2. Empty items
	_, err = svc.CreateOrder(ctx, 1, model.CreateOrderRequest{
		Items: []model.OrderItemRequest{},
	})
	if err == nil {
		t.Errorf("Expected error for empty items, got nil")
	}

	// 3. Negative quantity
	_, err = svc.CreateOrder(ctx, 1, model.CreateOrderRequest{
		Items: []model.OrderItemRequest{{ProductID: 1, Quantity: -1, Price: 10}},
	})
	if err == nil {
		t.Errorf("Expected error for negative quantity, got nil")
	}

	// 4. Missing price when product service unavailable
	_, err = svc.CreateOrder(ctx, 1, model.CreateOrderRequest{
		Items: []model.OrderItemRequest{{ProductID: 1, Quantity: 1, Price: 0}},
	})
	if err == nil {
		t.Errorf("Expected error for zero price when product service is offline, got nil")
	}
}

func TestGetOrderByID(t *testing.T) {
	svc, _ := setupTestOrderService()
	ctx := context.Background()

	order, err := svc.CreateOrder(ctx, 5, model.CreateOrderRequest{
		Items: []model.OrderItemRequest{{ProductID: 2, Quantity: 3, Price: 49.99}},
	})
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Fetch created order
	fetched, err := svc.GetOrderByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("Expected no error fetching order, got: %v", err)
	}
	if fetched.ID != order.ID {
		t.Errorf("Expected order ID %d, got %d", order.ID, fetched.ID)
	}
	if len(fetched.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(fetched.Items))
	}

	// Fetch non-existent order
	_, err = svc.GetOrderByID(ctx, 9999)
	if err == nil {
		t.Errorf("Expected error fetching non-existent order, got nil")
	}
	if !errors.Is(err, repository.ErrOrderNotFound) {
		t.Errorf("Expected ErrOrderNotFound, got: %v", err)
	}
}

func TestGetOrdersByUserID(t *testing.T) {
	svc, _ := setupTestOrderService()
	ctx := context.Background()
	targetUser := 15

	// Create 2 orders for targetUser
	_, _ = svc.CreateOrder(ctx, targetUser, model.CreateOrderRequest{
		Items: []model.OrderItemRequest{{ProductID: 1, Quantity: 1, Price: 100}},
	})
	_, _ = svc.CreateOrder(ctx, targetUser, model.CreateOrderRequest{
		Items: []model.OrderItemRequest{{ProductID: 2, Quantity: 2, Price: 50}},
	})

	// Create 1 order for a different user
	_, _ = svc.CreateOrder(ctx, 99, model.CreateOrderRequest{
		Items: []model.OrderItemRequest{{ProductID: 3, Quantity: 1, Price: 200}},
	})

	orders, err := svc.GetOrdersByUserID(ctx, targetUser)
	if err != nil {
		t.Fatalf("Expected no error getting user orders, got: %v", err)
	}
	if len(orders) != 2 {
		t.Errorf("Expected 2 orders for user %d, got %d", targetUser, len(orders))
	}
}

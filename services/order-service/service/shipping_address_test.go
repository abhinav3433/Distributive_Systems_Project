package service

import (
	"context"
	"strings"
	"testing"

	"order-service/messaging"
	"order-service/model"
	"order-service/repository"
)

func setupTestOrderServiceWithRepo() (OrderService, *repository.MockOrderRepository, *messaging.MockEventPublisher) {
	repo := repository.NewMockOrderRepository()
	mockPublisher := messaging.NewMockEventPublisher()
	svc := NewOrderService(repo, mockPublisher, "")
	return svc, repo, mockPublisher
}

func validTestAddress() *model.ShippingAddress {
	return &model.ShippingAddress{
		RecipientName: "Alice Wonderland",
		PhoneNumber:   "+1 (555) 987-6543",
		AddressLine1:  "742 Evergreen Terrace",
		AddressLine2:  "Apt 2B",
		City:          "Springfield",
		State:         "OR",
		PostalCode:    "97477",
		Country:       "United States",
	}
}

// 1. Successful order creation with valid shipping address
func TestCreateOrderWithShippingAddress(t *testing.T) {
	svc, _, mockPublisher := setupTestOrderServiceWithRepo()
	ctx := context.Background()
	userID := 42

	req := model.CreateOrderRequest{
		Items: []model.OrderItemRequest{
			{ProductID: 1, ProductName: "ProBook Laptop 15\"", Price: 1299.99, Quantity: 1},
		},
		Address: validTestAddress(),
	}

	order, err := svc.CreateOrder(ctx, userID, req)
	if err != nil {
		t.Fatalf("Failed to create order with address: %v", err)
	}

	if order.Address == nil {
		t.Fatalf("Expected order to have attached ShippingAddress, got nil")
	}
	if order.Address.RecipientName != "Alice Wonderland" {
		t.Errorf("Expected recipient Alice Wonderland, got %s", order.Address.RecipientName)
	}
	if order.Address.OrderID != order.ID {
		t.Errorf("Expected address OrderID %d to match order ID %d", order.Address.OrderID, order.ID)
	}
	if order.ShippingAddress == "" {
		t.Errorf("Expected formatted ShippingAddress summary string, got empty")
	}

	// Verify RabbitMQ event published only after successful commit
	if len(mockPublisher.PublishedEvents) != 1 {
		t.Fatalf("Expected 1 published event, got %d", len(mockPublisher.PublishedEvents))
	}
	if mockPublisher.PublishedEvents[0].OrderID != order.ID {
		t.Errorf("Expected published event OrderID %d, got %d", order.ID, mockPublisher.PublishedEvents[0].OrderID)
	}
}

// 2. Comprehensive validation rules on address fields
func TestShippingAddressValidationRules(t *testing.T) {
	svc, _, _ := setupTestOrderServiceWithRepo()
	ctx := context.Background()

	testCases := []struct {
		name      string
		modifyFn  func(a *model.ShippingAddress)
		shouldErr bool
	}{
		{
			name: "Missing recipient name",
			modifyFn: func(a *model.ShippingAddress) {
				a.RecipientName = "   "
			},
			shouldErr: true,
		},
		{
			name: "Recipient name exceeds 255 chars",
			modifyFn: func(a *model.ShippingAddress) {
				a.RecipientName = strings.Repeat("A", 256)
			},
			shouldErr: true,
		},
		{
			name: "Missing phone number",
			modifyFn: func(a *model.ShippingAddress) {
				a.PhoneNumber = ""
			},
			shouldErr: true,
		},
		{
			name: "Phone number exceeds 50 chars",
			modifyFn: func(a *model.ShippingAddress) {
				a.PhoneNumber = strings.Repeat("1", 51)
			},
			shouldErr: true,
		},
		{
			name: "Missing address line 1",
			modifyFn: func(a *model.ShippingAddress) {
				a.AddressLine1 = ""
			},
			shouldErr: true,
		},
		{
			name: "Address line 1 exceeds 255 chars",
			modifyFn: func(a *model.ShippingAddress) {
				a.AddressLine1 = strings.Repeat("A", 256)
			},
			shouldErr: true,
		},
		{
			name: "Address line 2 exceeds 255 chars",
			modifyFn: func(a *model.ShippingAddress) {
				a.AddressLine2 = strings.Repeat("B", 256)
			},
			shouldErr: true,
		},
		{
			name: "Missing city",
			modifyFn: func(a *model.ShippingAddress) {
				a.City = "  "
			},
			shouldErr: true,
		},
		{
			name: "City exceeds 100 chars",
			modifyFn: func(a *model.ShippingAddress) {
				a.City = strings.Repeat("C", 101)
			},
			shouldErr: true,
		},
		{
			name: "Missing state",
			modifyFn: func(a *model.ShippingAddress) {
				a.State = ""
			},
			shouldErr: true,
		},
		{
			name: "State exceeds 100 chars",
			modifyFn: func(a *model.ShippingAddress) {
				a.State = strings.Repeat("S", 101)
			},
			shouldErr: true,
		},
		{
			name: "Missing postal code",
			modifyFn: func(a *model.ShippingAddress) {
				a.PostalCode = ""
			},
			shouldErr: true,
		},
		{
			name: "Postal code exceeds 50 chars",
			modifyFn: func(a *model.ShippingAddress) {
				a.PostalCode = strings.Repeat("9", 51)
			},
			shouldErr: true,
		},
		{
			name: "Missing country",
			modifyFn: func(a *model.ShippingAddress) {
				a.Country = " "
			},
			shouldErr: true,
		},
		{
			name: "Country exceeds 100 chars",
			modifyFn: func(a *model.ShippingAddress) {
				a.Country = strings.Repeat("C", 101)
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			addr := validTestAddress()
			tc.modifyFn(addr)

			_, err := svc.CreateOrder(ctx, 10, model.CreateOrderRequest{
				Items:   []model.OrderItemRequest{{ProductID: 1, Price: 50, Quantity: 1}},
				Address: addr,
			})

			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for case '%s', got nil", tc.name)
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Expected success for case '%s', got error: %v", tc.name, err)
			}
		})
	}
}

// 3. Transaction rollback when shipping address insert fails (no event published, no partial state)
func TestShippingAddressDatabaseFailureRollback(t *testing.T) {
	svc, repo, mockPublisher := setupTestOrderServiceWithRepo()
	ctx := context.Background()

	// Configure mock repo to simulate failure when saving shipping address
	repo.FailOnAddress = true

	req := model.CreateOrderRequest{
		Items:   []model.OrderItemRequest{{ProductID: 1, Price: 100, Quantity: 1}},
		Address: validTestAddress(),
	}

	_, err := svc.CreateOrder(ctx, 10, req)
	if err == nil {
		t.Fatalf("Expected order creation to fail when address insert fails, got nil")
	}

	// Verify no OrderCreated event was published to RabbitMQ
	if len(mockPublisher.PublishedEvents) != 0 {
		t.Errorf("Expected 0 published events on transaction rollback, got %d", len(mockPublisher.PublishedEvents))
	}

	// Verify no order was saved in database
	orders, _ := repo.GetOrdersByUserID(ctx, 10)
	if len(orders) != 0 {
		t.Errorf("Expected 0 orders for user 10 after rollback, got %d", len(orders))
	}
}

// 4. Retrieve order with its associated shipping address
func TestRetrieveOrderWithShippingAddress(t *testing.T) {
	svc, _, _ := setupTestOrderServiceWithRepo()
	ctx := context.Background()
	userID := 88

	created, err := svc.CreateOrder(ctx, userID, model.CreateOrderRequest{
		Items:   []model.OrderItemRequest{{ProductID: 2, Price: 49.99, Quantity: 2}},
		Address: validTestAddress(),
	})
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Fetch by ID
	fetched, err := svc.GetOrderByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to fetch order: %v", err)
	}
	if fetched.Address == nil {
		t.Fatalf("Expected fetched order to have Address, got nil")
	}
	if fetched.Address.RecipientName != "Alice Wonderland" {
		t.Errorf("Expected recipient Alice Wonderland, got %s", fetched.Address.RecipientName)
	}
	if fetched.Address.City != "Springfield" {
		t.Errorf("Expected city Springfield, got %s", fetched.Address.City)
	}

	// Fetch by User ID
	userOrders, err := svc.GetOrdersByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to fetch user orders: %v", err)
	}
	if len(userOrders) != 1 {
		t.Fatalf("Expected 1 order, got %d", len(userOrders))
	}
	if userOrders[0].Address == nil {
		t.Fatalf("Expected user order to have Address, got nil")
	}
	if userOrders[0].Address.PostalCode != "97477" {
		t.Errorf("Expected postal code 97477, got %s", userOrders[0].Address.PostalCode)
	}
}

// 5. Behavior when existing historical order has no shipping address
func TestHistoricalOrderWithoutShippingAddress(t *testing.T) {
	svc, repo, _ := setupTestOrderServiceWithRepo()
	ctx := context.Background()

	// Simulate historical order with nil Address and empty string
	historicalOrder := &model.Order{
		UserID:      99,
		Status:      model.StatusConfirmed,
		TotalAmount: 199.99,
	}
	items := []model.OrderItem{
		{ProductID: 1, Quantity: 1, Price: 199.99},
	}
	if err := repo.CreateOrderWithItems(ctx, historicalOrder, items); err != nil {
		t.Fatalf("Failed to insert historical order: %v", err)
	}

	// Fetch historical order by ID - must not crash or fail
	fetched, err := svc.GetOrderByID(ctx, historicalOrder.ID)
	if err != nil {
		t.Fatalf("Expected success fetching historical order, got: %v", err)
	}
	if fetched.Address != nil {
		t.Errorf("Expected nil Address for historical order, got %+v", fetched.Address)
	}

	// Fetch historical orders by User ID - must not crash or fail
	userOrders, err := svc.GetOrdersByUserID(ctx, 99)
	if err != nil {
		t.Fatalf("Expected success fetching user orders with historical record, got: %v", err)
	}
	if len(userOrders) != 1 {
		t.Fatalf("Expected 1 order, got %d", len(userOrders))
	}
	if userOrders[0].Address != nil {
		t.Errorf("Expected nil Address for historical order in list, got %+v", userOrders[0].Address)
	}
}

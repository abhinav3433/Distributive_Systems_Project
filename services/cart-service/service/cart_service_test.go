package service

import (
	"context"
	"errors"
	"testing"

	"cart-service/model"
	"cart-service/repository"
)

func setupTestCartService() CartService {
	repo := repository.NewMockCartRepository()
	return NewCartService(repo)
}

func TestAddToCartAndGetCart(t *testing.T) {
	svc := setupTestCartService()
	ctx := context.Background()
	userID := 42

	// 1. Initially empty cart
	cart, err := svc.GetUserCart(ctx, userID)
	if err != nil {
		t.Fatalf("Expected no error fetching initial cart, got: %v", err)
	}
	if len(cart.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(cart.Items))
	}

	// 2. Add product 101 to cart
	req := model.AddToCartRequest{
		ProductID:   101,
		ProductName: "Gaming Laptop",
		Price:       1200.00,
		Quantity:    2,
	}

	cart, err = svc.AddToCart(ctx, userID, req)
	if err != nil {
		t.Fatalf("Expected no error adding item to cart, got: %v", err)
	}

	if len(cart.Items) != 1 {
		t.Fatalf("Expected 1 item in cart, got %d", len(cart.Items))
	}
	if cart.Items[0].ProductID != 101 {
		t.Errorf("Expected product ID 101, got %d", cart.Items[0].ProductID)
	}
	if cart.Items[0].Quantity != 2 {
		t.Errorf("Expected quantity 2, got %d", cart.Items[0].Quantity)
	}
	if cart.TotalPrice != 2400.00 {
		t.Errorf("Expected total price 2400.00, got %f", cart.TotalPrice)
	}
}

func TestUpdateItemQuantity(t *testing.T) {
	svc := setupTestCartService()
	ctx := context.Background()
	userID := 42

	// Add item first
	_, _ = svc.AddToCart(ctx, userID, model.AddToCartRequest{
		ProductID:   202,
		ProductName: "Wireless Mouse",
		Price:       25.00,
		Quantity:    1,
	})

	// Update quantity to 5
	cart, err := svc.UpdateItemQuantity(ctx, userID, 202, model.UpdateItemQuantityRequest{
		Quantity: 5,
	})
	if err != nil {
		t.Fatalf("Expected no error updating quantity, got: %v", err)
	}

	if cart.Items[0].Quantity != 5 {
		t.Errorf("Expected updated quantity 5, got %d", cart.Items[0].Quantity)
	}
	if cart.TotalPrice != 125.00 {
		t.Errorf("Expected total price 125.00, got %f", cart.TotalPrice)
	}

	// Update non-existent item
	_, err = svc.UpdateItemQuantity(ctx, userID, 999, model.UpdateItemQuantityRequest{
		Quantity: 2,
	})
	if err == nil {
		t.Errorf("Expected error updating non-existent item, got nil")
	}
	if !errors.Is(err, repository.ErrItemNotFound) {
		t.Errorf("Expected ErrItemNotFound, got: %v", err)
	}
}

func TestRemoveFromCart(t *testing.T) {
	svc := setupTestCartService()
	ctx := context.Background()
	userID := 42

	_, _ = svc.AddToCart(ctx, userID, model.AddToCartRequest{
		ProductID:   303,
		ProductName: "Keyboard",
		Price:       75.00,
		Quantity:    1,
	})

	// Remove product 303
	cart, err := svc.RemoveFromCart(ctx, userID, 303)
	if err != nil {
		t.Fatalf("Expected no error removing item, got: %v", err)
	}

	if len(cart.Items) != 0 {
		t.Errorf("Expected 0 items after removal, got %d", len(cart.Items))
	}
}

func TestClearUserCart(t *testing.T) {
	svc := setupTestCartService()
	ctx := context.Background()
	userID := 42

	_, _ = svc.AddToCart(ctx, userID, model.AddToCartRequest{
		ProductID: 1, Quantity: 1, Price: 10.0,
	})
	_, _ = svc.AddToCart(ctx, userID, model.AddToCartRequest{
		ProductID: 2, Quantity: 2, Price: 20.0,
	})

	// Clear cart
	err := svc.ClearUserCart(ctx, userID)
	if err != nil {
		t.Fatalf("Expected no error clearing cart, got: %v", err)
	}

	cart, err := svc.GetUserCart(ctx, userID)
	if err != nil {
		t.Fatalf("Expected no error fetching cart, got: %v", err)
	}
	if len(cart.Items) != 0 {
		t.Errorf("Expected empty cart, got %d items", len(cart.Items))
	}
}

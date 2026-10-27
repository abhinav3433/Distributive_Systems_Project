package grpc

import (
	"context"
	"testing"

	"cart-service/repository"
	"cart-service/service"
	cartv1 "proto/cart/v1"
)

func setupTestCartGRPC() (*CartGRPCServer, service.CartService) {
	repo := repository.NewMockCartRepository()
	svc := service.NewCartService(repo)
	return NewCartGRPCServer(svc), svc
}

func TestCartGRPC_Lifecycle(t *testing.T) {
	server, _ := setupTestCartGRPC()
	ctx := context.Background()
	userID := int64(42)

	// 1. Initial Cart (Empty)
	cartResp, err := server.GetCart(ctx, &cartv1.GetCartRequest{UserId: userID})
	if err != nil {
		t.Fatalf("GetCart failed: %v", err)
	}
	if len(cartResp.GetCart().GetItems()) != 0 {
		t.Errorf("expected empty cart, got %d items", len(cartResp.GetCart().GetItems()))
	}

	// 2. Add Item 1
	addResp, err := server.AddToCart(ctx, &cartv1.AddToCartRequest{
		UserId:      userID,
		ProductId:   1,
		Quantity:    2,
		ProductName: "ProBook Laptop",
		Price:       1299.99,
	})
	if err != nil {
		t.Fatalf("AddToCart failed: %v", err)
	}
	if len(addResp.GetCart().GetItems()) != 1 {
		t.Fatalf("expected 1 item in cart, got %d", len(addResp.GetCart().GetItems()))
	}
	if addResp.GetCart().GetTotalPrice() != 2599.98 {
		t.Errorf("expected total 2599.98, got %f", addResp.GetCart().GetTotalPrice())
	}

	// 3. Update Item Quantity
	updResp, err := server.UpdateCartItem(ctx, &cartv1.UpdateCartItemRequest{
		UserId:    userID,
		ProductId: 1,
		Quantity:  3,
	})
	if err != nil {
		t.Fatalf("UpdateCartItem failed: %v", err)
	}
	if updResp.GetCart().GetItems()[0].GetQuantity() != 3 {
		t.Errorf("expected quantity 3, got %d", updResp.GetCart().GetItems()[0].GetQuantity())
	}

	// 4. Remove Item
	remResp, err := server.RemoveFromCart(ctx, &cartv1.RemoveFromCartRequest{
		UserId:    userID,
		ProductId: 1,
	})
	if err != nil {
		t.Fatalf("RemoveFromCart failed: %v", err)
	}
	if len(remResp.GetCart().GetItems()) != 0 {
		t.Errorf("expected empty cart after removal, got %d items", len(remResp.GetCart().GetItems()))
	}

	// 5. Clear Cart
	clearResp, err := server.ClearCart(ctx, &cartv1.ClearCartRequest{UserId: userID})
	if err != nil {
		t.Fatalf("ClearCart failed: %v", err)
	}
	if !clearResp.GetSuccess() {
		t.Error("expected success = true on ClearCart")
	}
}

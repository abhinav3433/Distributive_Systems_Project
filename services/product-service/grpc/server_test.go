package grpc

import (
	"context"
	"testing"

	"product-service/repository"
	"product-service/service"
	productv1 "proto/product/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupTestProductGRPC() (*ProductGRPCServer, service.ProductService) {
	repo := repository.NewMockProductRepository()
	svc := service.NewProductService(repo)
	return NewProductGRPCServer(svc), svc
}

func TestProductGRPC_ListProducts(t *testing.T) {
	server, _ := setupTestProductGRPC()

	resp, err := server.ListProducts(context.Background(), &productv1.ListProductsRequest{})
	if err != nil {
		t.Fatalf("ListProducts failed: %v", err)
	}
	if len(resp.GetProducts()) == 0 {
		t.Fatal("expected seeded products, got 0")
	}

	// Filter by Category
	respFilter, err := server.ListProducts(context.Background(), &productv1.ListProductsRequest{
		Category: "Electronics",
	})
	if err != nil {
		t.Fatalf("ListProducts with category filter failed: %v", err)
	}
	for _, p := range respFilter.GetProducts() {
		if p.GetCategory() != "Electronics" {
			t.Errorf("expected category Electronics, got %s", p.GetCategory())
		}
	}
}

func TestProductGRPC_GetProduct_Success_And_NotFound(t *testing.T) {
	server, _ := setupTestProductGRPC()

	resp, err := server.GetProduct(context.Background(), &productv1.GetProductRequest{Id: 1})
	if err != nil {
		t.Fatalf("GetProduct(1) failed: %v", err)
	}
	if resp.GetProduct().GetId() != 1 {
		t.Errorf("expected product ID 1, got %d", resp.GetProduct().GetId())
	}

	// Not Found
	_, err = server.GetProduct(context.Background(), &productv1.GetProductRequest{Id: 99999})
	if err == nil {
		t.Fatal("expected error on non-existent product, got nil")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
}

func TestProductGRPC_Create_Update_Delete(t *testing.T) {
	server, _ := setupTestProductGRPC()

	// Create
	createResp, err := server.CreateProduct(context.Background(), &productv1.CreateProductRequest{
		Name:        "New Gadget",
		Description: "A brand new gadget",
		Price:       199.99,
		Stock:       50,
		Category:    "Electronics",
	})
	if err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}
	newID := createResp.GetProduct().GetId()
	if newID <= 0 {
		t.Errorf("expected positive ID, got %d", newID)
	}

	// Update
	updateResp, err := server.UpdateProduct(context.Background(), &productv1.UpdateProductRequest{
		Id:          newID,
		Name:        "Updated Gadget",
		Description: "Updated description",
		Price:       249.99,
		Stock:       75,
		Category:    "Electronics",
	})
	if err != nil {
		t.Fatalf("UpdateProduct failed: %v", err)
	}
	if updateResp.GetProduct().GetName() != "Updated Gadget" {
		t.Errorf("expected updated name, got %s", updateResp.GetProduct().GetName())
	}

	// Delete
	delResp, err := server.DeleteProduct(context.Background(), &productv1.DeleteProductRequest{Id: newID})
	if err != nil {
		t.Fatalf("DeleteProduct failed: %v", err)
	}
	if !delResp.GetSuccess() {
		t.Error("expected success = true on delete")
	}
}

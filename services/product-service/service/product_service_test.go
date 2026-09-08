package service

import (
	"errors"
	"testing"

	"product-service/model"
	"product-service/repository"
)

func setupTestProductService() ProductService {
	repo := repository.NewMockProductRepository()
	return NewProductService(repo)
}

func TestListProducts(t *testing.T) {
	svc := setupTestProductService()

	products, err := svc.ListProducts()
	if err != nil {
		t.Fatalf("Expected no error on ListProducts, got: %v", err)
	}

	if len(products) < 10 {
		t.Errorf("Expected at least 10 seed products, got %d", len(products))
	}
}

func TestCreateProduct(t *testing.T) {
	svc := setupTestProductService()

	// 1. Test valid product creation
	req := model.CreateProductRequest{
		Name:        "Smart 4K TV 55\"",
		Description: "Ultra HD Smart OLED Television",
		Price:       699.99,
		Stock:       15,
		Category:    "Electronics",
	}

	p, err := svc.CreateProduct(req)
	if err != nil {
		t.Fatalf("Expected no error creating product, got: %v", err)
	}

	if p.ID == 0 {
		t.Errorf("Expected non-zero product ID")
	}
	if p.Name != "Smart 4K TV 55\"" {
		t.Errorf("Expected product name 'Smart 4K TV 55\"', got '%s'", p.Name)
	}

	// 2. Test invalid product creation (empty name)
	invalidReq := model.CreateProductRequest{
		Name:     "",
		Price:    100.0,
		Category: "Electronics",
	}
	_, err = svc.CreateProduct(invalidReq)
	if err == nil {
		t.Errorf("Expected error for empty product name, got nil")
	}

	// 3. Test invalid price
	invalidPriceReq := model.CreateProductRequest{
		Name:     "Test Product",
		Price:    -10.0,
		Category: "Electronics",
	}
	_, err = svc.CreateProduct(invalidPriceReq)
	if err == nil {
		t.Errorf("Expected error for negative price, got nil")
	}
}

func TestGetProduct(t *testing.T) {
	svc := setupTestProductService()

	// Get seed product 1
	p, err := svc.GetProduct(1)
	if err != nil {
		t.Fatalf("Expected no error getting product ID 1, got: %v", err)
	}
	if p.ID != 1 {
		t.Errorf("Expected product ID 1, got %d", p.ID)
	}

	// Get non-existent product
	_, err = svc.GetProduct(9999)
	if err == nil {
		t.Errorf("Expected error for non-existent product, got nil")
	}
	if !errors.Is(err, repository.ErrProductNotFound) {
		t.Errorf("Expected ErrProductNotFound, got: %v", err)
	}
}

func TestUpdateProduct(t *testing.T) {
	svc := setupTestProductService()

	updateReq := model.UpdateProductRequest{
		Name:  "ProBook Laptop 15\" (2026 Edition)",
		Price: 1399.99,
		Stock: 30,
	}

	updated, err := svc.UpdateProduct(1, updateReq)
	if err != nil {
		t.Fatalf("Expected no error updating product, got: %v", err)
	}
	if updated.Name != "ProBook Laptop 15\" (2026 Edition)" {
		t.Errorf("Expected updated name, got '%s'", updated.Name)
	}
	if updated.Price != 1399.99 {
		t.Errorf("Expected updated price 1399.99, got %f", updated.Price)
	}

	// Update non-existent product
	_, err = svc.UpdateProduct(9999, updateReq)
	if err == nil {
		t.Errorf("Expected error updating non-existent product, got nil")
	}
}

func TestDeleteProduct(t *testing.T) {
	svc := setupTestProductService()

	// Delete product 1
	err := svc.DeleteProduct(1)
	if err != nil {
		t.Fatalf("Expected no error deleting product ID 1, got: %v", err)
	}

	// Verify it's deleted
	_, err = svc.GetProduct(1)
	if err == nil {
		t.Errorf("Expected error fetching deleted product, got nil")
	}

	// Delete non-existent product
	err = svc.DeleteProduct(9999)
	if err == nil {
		t.Errorf("Expected error deleting non-existent product, got nil")
	}
}

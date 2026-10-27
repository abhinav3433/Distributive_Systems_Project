package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"api-gateway/grpcclient"
	productv1 "proto/product/v1"

	"github.com/gin-gonic/gin"
)

type ProductGRPCHandler struct {
	client productv1.ProductServiceClient
}

func NewProductGRPCHandler(client productv1.ProductServiceClient) *ProductGRPCHandler {
	return &ProductGRPCHandler{client: client}
}

func formatProductJSON(p *productv1.Product) gin.H {
	if p == nil {
		return gin.H{}
	}
	return gin.H{
		"id":          p.GetId(),
		"name":        p.GetName(),
		"description": p.GetDescription(),
		"price":       p.GetPrice(),
		"stock":       p.GetStock(),
		"category":    p.GetCategory(),
		"created_at":  p.GetCreatedAt(),
		"updated_at":  p.GetUpdatedAt(),
	}
}

func (h *ProductGRPCHandler) ListProducts(c *gin.Context) {
	category := c.Query("category")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.ListProducts(ctx, &productv1.ListProductsRequest{
		Category: category,
	})
	if err != nil {
		grpcclient.HandleGRPCError(c, err, "failed to list products")
		return
	}

	var result []gin.H
	for _, p := range resp.GetProducts() {
		result = append(result, formatProductJSON(p))
	}
	if result == nil {
		result = []gin.H{}
	}

	c.JSON(http.StatusOK, result)
}

func (h *ProductGRPCHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetProduct(ctx, &productv1.GetProductRequest{Id: id})
	if err != nil {
		grpcclient.HandleGRPCError(c, err, "failed to get product")
		return
	}

	c.JSON(http.StatusOK, formatProductJSON(resp.GetProduct()))
}

type CreateProductInput struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int32   `json:"stock" binding:"gte=0"`
	Category    string  `json:"category" binding:"required"`
}

func (h *ProductGRPCHandler) CreateProduct(c *gin.Context) {
	var input CreateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.CreateProduct(ctx, &productv1.CreateProductRequest{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
		Category:    input.Category,
	})
	if err != nil {
		grpcclient.HandleGRPCError(c, err, "failed to create product")
		return
	}

	c.JSON(http.StatusCreated, formatProductJSON(resp.GetProduct()))
}

func (h *ProductGRPCHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var input CreateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.UpdateProduct(ctx, &productv1.UpdateProductRequest{
		Id:          id,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
		Category:    input.Category,
	})
	if err != nil {
		grpcclient.HandleGRPCError(c, err, "failed to update product")
		return
	}

	c.JSON(http.StatusOK, formatProductJSON(resp.GetProduct()))
}

func (h *ProductGRPCHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err = h.client.DeleteProduct(ctx, &productv1.DeleteProductRequest{Id: id})
	if err != nil {
		grpcclient.HandleGRPCError(c, err, "failed to delete product")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

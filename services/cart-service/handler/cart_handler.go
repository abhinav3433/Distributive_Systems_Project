package handler

import (
	"errors"
	"net/http"
	"strconv"

	"cart-service/model"
	"cart-service/repository"
	"cart-service/service"

	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	cartSvc service.CartService
}

func NewCartHandler(svc service.CartService) *CartHandler {
	return &CartHandler{cartSvc: svc}
}

func getUserID(c *gin.Context) int {
	// Try header X-User-ID
	if uidStr := c.GetHeader("X-User-ID"); uidStr != "" {
		if uid, err := strconv.Atoi(uidStr); err == nil && uid > 0 {
			return uid
		}
	}
	// Try query parameter user_id
	if uidStr := c.Query("user_id"); uidStr != "" {
		if uid, err := strconv.Atoi(uidStr); err == nil && uid > 0 {
			return uid
		}
	}
	// Default to demo user ID 1
	return 1
}

func (h *CartHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "UP",
		"service": "cart-service",
	})
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID := getUserID(c)
	cart, err := h.cartSvc.GetUserCart(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cart: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	userID := getUserID(c)
	var req model.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	cart, err := h.cartSvc.AddToCart(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) UpdateItemQuantity(c *gin.Context) {
	userID := getUserID(c)
	productIDParam := c.Param("productId")
	productID, err := strconv.Atoi(productIDParam)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req model.UpdateItemQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	cart, err := h.cartSvc.UpdateItemQuantity(c.Request.Context(), userID, productID, req)
	if err != nil {
		if errors.Is(err, repository.ErrItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found in cart"})
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item quantity: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	userID := getUserID(c)
	productIDParam := c.Param("productId")
	productID, err := strconv.Atoi(productIDParam)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	cart, err := h.cartSvc.RemoveFromCart(c.Request.Context(), userID, productID)
	if err != nil {
		if errors.Is(err, repository.ErrItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found in cart"})
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item from cart: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) ClearCart(c *gin.Context) {
	userID := getUserID(c)
	err := h.cartSvc.ClearUserCart(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart cleared successfully",
		"user_id": userID,
	})
}

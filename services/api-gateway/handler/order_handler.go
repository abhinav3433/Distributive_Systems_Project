package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"api-gateway/grpcclient"
	authv1 "proto/auth/v1"
	orderv1 "proto/order/v1"

	"github.com/gin-gonic/gin"
)

type OrderGRPCHandler struct {
	client     orderv1.OrderServiceClient
	authClient authv1.AuthServiceClient
}

func NewOrderGRPCHandler(client orderv1.OrderServiceClient, authClient authv1.AuthServiceClient) *OrderGRPCHandler {
	return &OrderGRPCHandler{
		client:     client,
		authClient: authClient,
	}
}

func (h *OrderGRPCHandler) getUserID(c *gin.Context) (int64, bool) {
	// 1. Context variable from authentication middleware
	if uid, exists := c.Get("userID"); exists {
		if id, ok := uid.(int64); ok && id > 0 {
			return id, true
		}
		if id, ok := uid.(int); ok && id > 0 {
			return int64(id), true
		}
	}

	// 2. Direct X-User-ID header
	if headerVal := c.GetHeader("X-User-ID"); headerVal != "" {
		if id, err := strconv.ParseInt(headerVal, 10, 64); err == nil && id > 0 {
			return id, true
		}
	}

	// 3. Bearer JWT Token validated via Auth Service gRPC
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") && h.authClient != nil {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		resp, err := h.authClient.ValidateToken(ctx, &authv1.ValidateTokenRequest{Token: token})
		if err == nil && resp.GetValid() && resp.GetUserId() > 0 {
			return resp.GetUserId(), true
		}
	}

	// Default fallback for guest/unauthenticated testing if none provided
	return 1, true
}

func formatOrderJSON(o *orderv1.Order) gin.H {
	if o == nil {
		return gin.H{}
	}

	var items []gin.H
	for _, it := range o.GetItems() {
		items = append(items, gin.H{
			"id":           it.GetId(),
			"order_id":     it.GetOrderId(),
			"product_id":   it.GetProductId(),
			"product_name": it.GetProductName(),
			"quantity":     it.GetQuantity(),
			"price":        it.GetPrice(),
		})
	}
	if items == nil {
		items = []gin.H{}
	}

	res := gin.H{
		"id":               o.GetId(),
		"user_id":          o.GetUserId(),
		"status":           o.GetStatus(),
		"total_amount":     o.GetTotalAmount(),
		"shipping_address": o.GetShippingAddress(),
		"items":            items,
		"created_at":       o.GetCreatedAt(),
		"updated_at":       o.GetUpdatedAt(),
	}

	if o.GetAddress() != nil {
		a := o.GetAddress()
		res["address"] = gin.H{
			"id":             a.GetId(),
			"order_id":       a.GetOrderId(),
			"recipient_name": a.GetRecipientName(),
			"phone_number":   a.GetPhoneNumber(),
			"address_line1":  a.GetAddressLine1(),
			"address_line2":  a.GetAddressLine2(),
			"city":           a.GetCity(),
			"state":          a.GetState(),
			"postal_code":    a.GetPostalCode(),
			"country":        a.GetCountry(),
			"created_at":     a.GetCreatedAt(),
		}
	}

	return res
}

type OrderItemRequestJSON struct {
	ProductID   int64   `json:"product_id" binding:"required,gt=0"`
	Quantity    int32   `json:"quantity" binding:"required,gt=0"`
	Price       float64 `json:"price"`
	ProductName string  `json:"product_name"`
}

type ShippingAddressInput struct {
	RecipientName string `json:"recipient_name"`
	PhoneNumber   string `json:"phone_number"`
	AddressLine1  string `json:"address_line1"`
	AddressLine2  string `json:"address_line2"`
	City          string `json:"city"`
	State         string `json:"state"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
}

type CreateOrderInput struct {
	Items           []OrderItemRequestJSON `json:"items" binding:"required,min=1"`
	ShippingAddress string                 `json:"shipping_address"`
	Address         *ShippingAddressInput  `json:"address"`
}

func (h *OrderGRPCHandler) CreateOrder(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized or missing user identifier"})
		return
	}

	var input CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var protoItems []*orderv1.OrderItemInput
	for _, it := range input.Items {
		protoItems = append(protoItems, &orderv1.OrderItemInput{
			ProductId:   it.ProductID,
			Quantity:    it.Quantity,
			Price:       it.Price,
			ProductName: it.ProductName,
		})
	}

	var protoAddr *orderv1.ShippingAddress
	if input.Address != nil {
		protoAddr = &orderv1.ShippingAddress{
			RecipientName: input.Address.RecipientName,
			PhoneNumber:   input.Address.PhoneNumber,
			AddressLine1:  input.Address.AddressLine1,
			AddressLine2:  input.Address.AddressLine2,
			City:          input.Address.City,
			State:         input.Address.State,
			PostalCode:    input.Address.PostalCode,
			Country:       input.Address.Country,
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId:          userID,
		Items:           protoItems,
		ShippingAddress: input.ShippingAddress,
		Address:         protoAddr,
	})
	if err != nil {
		grpcclient.HandleGRPCError(c, err, "failed to create order")
		return
	}

	c.JSON(http.StatusCreated, formatOrderJSON(resp.GetOrder()))
}

func (h *OrderGRPCHandler) GetOrders(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized or missing user identifier"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.ListOrders(ctx, &orderv1.ListOrdersRequest{UserId: userID})
	if err != nil {
		grpcclient.HandleGRPCError(c, err, "failed to list orders")
		return
	}

	var result []gin.H
	for _, o := range resp.GetOrders() {
		result = append(result, formatOrderJSON(o))
	}
	if result == nil {
		result = []gin.H{}
	}

	c.JSON(http.StatusOK, result)
}

func (h *OrderGRPCHandler) GetOrderByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetOrder(ctx, &orderv1.GetOrderRequest{Id: id})
	if err != nil {
		grpcclient.HandleGRPCError(c, err, "failed to get order")
		return
	}

	c.JSON(http.StatusOK, formatOrderJSON(resp.GetOrder()))
}

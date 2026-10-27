package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"api-gateway/grpcclient"
	paymentv1 "proto/payment/v1"

	"github.com/gin-gonic/gin"
)

type PaymentGRPCHandler struct {
	client paymentv1.PaymentServiceClient
}

func NewPaymentGRPCHandler(client paymentv1.PaymentServiceClient) *PaymentGRPCHandler {
	return &PaymentGRPCHandler{client: client}
}

func formatPaymentJSON(p *paymentv1.Payment) gin.H {
	if p == nil {
		return gin.H{}
	}
	return gin.H{
		"id":             p.GetId(),
		"order_id":       p.GetOrderId(),
		"user_id":        p.GetUserId(),
		"amount":         p.GetAmount(),
		"status":         p.GetStatus(),
		"payment_method": p.GetPaymentMethod(),
		"transaction_id": p.GetTransactionId(),
		"created_at":     p.GetCreatedAt(),
		"updated_at":     p.GetUpdatedAt(),
	}
}

func (h *PaymentGRPCHandler) GetPaymentByOrderID(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetPaymentByOrderID(ctx, &paymentv1.GetPaymentRequest{OrderId: orderID})
	if err != nil {
		grpcclient.HandleGRPCError(c, err, "failed to get payment details")
		return
	}

	c.JSON(http.StatusOK, formatPaymentJSON(resp.GetPayment()))
}

type ProcessPaymentInput struct {
	OrderID       int64   `json:"order_id" binding:"required,gt=0"`
	UserID        int64   `json:"user_id" binding:"required,gt=0"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod string  `json:"payment_method"`
}

func (h *PaymentGRPCHandler) ProcessPayment(c *gin.Context) {
	var input ProcessPaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.ProcessPayment(ctx, &paymentv1.ProcessPaymentRequest{
		OrderId:       input.OrderID,
		UserId:        input.UserID,
		Amount:        input.Amount,
		PaymentMethod: input.PaymentMethod,
	})
	if err != nil {
		grpcclient.HandleGRPCError(c, err, "failed to process payment")
		return
	}

	c.JSON(http.StatusOK, formatPaymentJSON(resp.GetPayment()))
}

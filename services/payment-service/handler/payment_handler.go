package handler

import (
	"errors"
	"net/http"
	"strconv"

	"payment-service/model"
	"payment-service/repository"
	"payment-service/service"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentSvc service.PaymentService
}

func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentSvc: svc}
}

func (h *PaymentHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "UP",
		"service": "payment-service",
	})
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	var req model.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	if req.UserID <= 0 {
		if uidStr := c.GetHeader("X-User-ID"); uidStr != "" {
			if uid, err := strconv.Atoi(uidStr); err == nil && uid > 0 {
				req.UserID = uid
			}
		}
	}
	if req.UserID <= 0 {
		req.UserID = 1
	}

	payment, err := h.paymentSvc.ProcessPayment(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment processing failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

func (h *PaymentHandler) GetPaymentByOrderID(c *gin.Context) {
	orderIDParam := c.Param("orderId")
	orderID, err := strconv.Atoi(orderIDParam)
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	payment, err := h.paymentSvc.GetPaymentByOrderID(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payment record not found for order"})
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payment: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

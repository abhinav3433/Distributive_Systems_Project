package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"api-gateway/grpcclient"
	authv1 "proto/auth/v1"

	"github.com/gin-gonic/gin"
)

type AuthGRPCHandler struct {
	client authv1.AuthServiceClient
}

func NewAuthGRPCHandler(client authv1.AuthServiceClient) *AuthGRPCHandler {
	return &AuthGRPCHandler{client: client}
}

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *AuthGRPCHandler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.Register(ctx, &authv1.RegisterRequest{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		grpcclient.HandleGRPCError(c, err, "failed to register user")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":         resp.GetId(),
			"name":       resp.GetName(),
			"email":      resp.GetEmail(),
			"created_at": resp.GetCreatedAt(),
		},
	})
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthGRPCHandler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.Login(ctx, &authv1.LoginRequest{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		grpcclient.HandleGRPCError(c, err, "login failed")
		return
	}

	u := resp.GetUser()
	c.JSON(http.StatusOK, gin.H{
		"token": resp.GetToken(),
		"user": gin.H{
			"id":         u.GetId(),
			"name":       u.GetName(),
			"email":      u.GetEmail(),
			"created_at": u.GetCreatedAt(),
			"updated_at": u.GetUpdatedAt(),
		},
	})
}

func (h *AuthGRPCHandler) GetMe(c *gin.Context) {
	// 1. Check Authorization Bearer header or X-User-ID header
	userIDStr := c.GetHeader("X-User-ID")
	authHeader := c.GetHeader("Authorization")

	var userID int64
	if userIDStr != "" {
		if parsed, err := strconv.ParseInt(userIDStr, 10, 64); err == nil && parsed > 0 {
			userID = parsed
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if userID == 0 && authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		valResp, err := h.client.ValidateToken(ctx, &authv1.ValidateTokenRequest{Token: token})
		if err == nil && valResp.GetValid() {
			userID = valResp.GetUserId()
		}
	}

	if userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized or missing credentials"})
		return
	}

	resp, err := h.client.GetMe(ctx, &authv1.GetMeRequest{UserId: userID})
	if err != nil {
		grpcclient.HandleGRPCError(c, err, "failed to get current user")
		return
	}

	u := resp.GetUser()
	c.JSON(http.StatusOK, gin.H{
		"id":         u.GetId(),
		"name":       u.GetName(),
		"email":      u.GetEmail(),
		"created_at": u.GetCreatedAt(),
		"updated_at": u.GetUpdatedAt(),
	})
}

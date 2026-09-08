package service

import (
	"testing"

	"auth-service/model"
	"auth-service/repository"
)

func setupTestAuthService() AuthService {
	mockRepo := repository.NewMockUserRepository()
	return NewAuthService(mockRepo, "test-secret-key")
}

func TestRegister(t *testing.T) {
	svc := setupTestAuthService()

	// 1. Test successful registration
	req := model.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	user, err := svc.Register(req)
	if err != nil {
		t.Fatalf("Expected no error on registration, got: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
	if user.Email != "john@example.com" {
		t.Errorf("Expected email john@example.com, got %s", user.Email)
	}
	if user.Name != "John Doe" {
		t.Errorf("Expected name John Doe, got %s", user.Name)
	}

	// 2. Test duplicate email registration
	_, err = svc.Register(req)
	if err == nil {
		t.Errorf("Expected error for duplicate email registration, got nil")
	}

	// 3. Test invalid input (empty password)
	invalidReq := model.RegisterRequest{
		Name:     "Jane",
		Email:    "jane@example.com",
		Password: "123",
	}
	_, err = svc.Register(invalidReq)
	if err == nil {
		t.Errorf("Expected error for short password, got nil")
	}
}

func TestLogin(t *testing.T) {
	svc := setupTestAuthService()

	// Register user first
	regReq := model.RegisterRequest{
		Name:     "Alice Smith",
		Email:    "alice@example.com",
		Password: "secretpassword",
	}
	_, err := svc.Register(regReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	// 1. Test successful login
	loginReq := model.LoginRequest{
		Email:    "alice@example.com",
		Password: "secretpassword",
	}
	resp, err := svc.Login(loginReq)
	if err != nil {
		t.Fatalf("Expected successful login, got: %v", err)
	}
	if resp.Token == "" {
		t.Errorf("Expected non-empty JWT token")
	}
	if resp.User.Email != "alice@example.com" {
		t.Errorf("Expected user email alice@example.com, got %s", resp.User.Email)
	}

	// 2. Test incorrect password
	wrongPassReq := model.LoginRequest{
		Email:    "alice@example.com",
		Password: "wrongpassword",
	}
	_, err = svc.Login(wrongPassReq)
	if err == nil {
		t.Errorf("Expected error for wrong password, got nil")
	}

	// 3. Test non-existent email
	noUserReq := model.LoginRequest{
		Email:    "nobody@example.com",
		Password: "secretpassword",
	}
	_, err = svc.Login(noUserReq)
	if err == nil {
		t.Errorf("Expected error for non-existent user, got nil")
	}
}

func TestValidateToken(t *testing.T) {
	svc := setupTestAuthService()

	regReq := model.RegisterRequest{
		Name:     "Bob",
		Email:    "bob@example.com",
		Password: "password123",
	}
	_, err := svc.Register(regReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	loginResp, err := svc.Login(model.LoginRequest{
		Email:    "bob@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// 1. Validate valid token
	claims, err := svc.ValidateToken(loginResp.Token)
	if err != nil {
		t.Fatalf("Expected valid token validation, got error: %v", err)
	}
	if claims.Email != "bob@example.com" {
		t.Errorf("Expected claim email bob@example.com, got %s", claims.Email)
	}

	// 2. Validate corrupted token
	_, err = svc.ValidateToken("invalid.token.string")
	if err == nil {
		t.Errorf("Expected error for invalid token string, got nil")
	}
}

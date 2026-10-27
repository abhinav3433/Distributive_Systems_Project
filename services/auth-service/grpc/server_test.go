package grpc

import (
	"context"
	"testing"

	"auth-service/repository"
	"auth-service/service"
	authv1 "proto/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupTestAuthGRPC() (*AuthGRPCServer, service.AuthService) {
	repo := repository.NewMockUserRepository()
	authSvc := service.NewAuthService(repo, "test-secret")
	return NewAuthGRPCServer(authSvc), authSvc
}

func TestAuthGRPC_Register_Success(t *testing.T) {
	server, _ := setupTestAuthGRPC()

	req := &authv1.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := server.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.GetEmail() != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", resp.GetEmail())
	}
	if resp.GetId() <= 0 {
		t.Errorf("expected positive user id, got %d", resp.GetId())
	}
}

func TestAuthGRPC_Register_DuplicateEmail(t *testing.T) {
	server, _ := setupTestAuthGRPC()

	req := &authv1.RegisterRequest{
		Name:     "User One",
		Email:    "duplicate@example.com",
		Password: "password123",
	}

	_, err := server.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	_, err = server.Register(context.Background(), req)
	if err == nil {
		t.Fatal("expected error on duplicate register, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.AlreadyExists {
		t.Errorf("expected AlreadyExists status code, got %v", st.Code())
	}
}

func TestAuthGRPC_Login_Success_And_GetMe(t *testing.T) {
	server, _ := setupTestAuthGRPC()

	// Register
	_, err := server.Register(context.Background(), &authv1.RegisterRequest{
		Name:     "Login User",
		Email:    "login@example.com",
		Password: "secretpassword",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Login
	loginResp, err := server.Login(context.Background(), &authv1.LoginRequest{
		Email:    "login@example.com",
		Password: "secretpassword",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginResp.GetToken() == "" {
		t.Error("expected non-empty JWT token")
	}

	// ValidateToken
	valResp, err := server.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{
		Token: loginResp.GetToken(),
	})
	if err != nil {
		t.Fatalf("validate token failed: %v", err)
	}
	if !valResp.GetValid() {
		t.Error("expected token to be valid")
	}

	// GetMe
	meResp, err := server.GetMe(context.Background(), &authv1.GetMeRequest{
		UserId: valResp.GetUserId(),
	})
	if err != nil {
		t.Fatalf("GetMe failed: %v", err)
	}
	if meResp.GetUser().GetEmail() != "login@example.com" {
		t.Errorf("expected email login@example.com, got %s", meResp.GetUser().GetEmail())
	}
}

func TestAuthGRPC_Login_InvalidPassword(t *testing.T) {
	server, _ := setupTestAuthGRPC()

	_, _ = server.Register(context.Background(), &authv1.RegisterRequest{
		Name:     "User Two",
		Email:    "user2@example.com",
		Password: "password123",
	})

	_, err := server.Login(context.Background(), &authv1.LoginRequest{
		Email:    "user2@example.com",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error on invalid password, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated status code, got %v", st.Code())
	}
}

func TestAuthGRPC_GetMe_NotFound(t *testing.T) {
	server, _ := setupTestAuthGRPC()

	_, err := server.GetMe(context.Background(), &authv1.GetMeRequest{
		UserId: 99999,
	})
	if err == nil {
		t.Fatal("expected error on non-existent user, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Errorf("expected NotFound status code, got %v", st.Code())
	}
}

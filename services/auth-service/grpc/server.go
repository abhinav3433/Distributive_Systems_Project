package grpc

import (
	"context"
	"errors"
	"time"

	"auth-service/model"
	"auth-service/repository"
	"auth-service/service"
	authv1 "proto/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthGRPCServer struct {
	authv1.UnimplementedAuthServiceServer
	authSvc service.AuthService
}

func NewAuthGRPCServer(authSvc service.AuthService) *AuthGRPCServer {
	return &AuthGRPCServer{
		authSvc: authSvc,
	}
}

func (s *AuthGRPCServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	user, err := s.authSvc.Register(model.RegisterRequest{
		Name:     req.GetName(),
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to register user: "+err.Error())
	}

	return &authv1.RegisterResponse{
		Id:        int64(user.ID),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *AuthGRPCServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	loginResp, err := s.authSvc.Login(model.LoginRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, "login failed: "+err.Error())
	}

	return &authv1.LoginResponse{
		Token: loginResp.Token,
		User: &authv1.User{
			Id:        int64(loginResp.User.ID),
			Name:      loginResp.User.Name,
			Email:     loginResp.User.Email,
			CreatedAt: loginResp.User.CreatedAt.Format(time.RFC3339),
			UpdatedAt: loginResp.User.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (s *AuthGRPCServer) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	if req == nil || req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	claims, err := s.authSvc.ValidateToken(req.GetToken())
	if err != nil {
		return &authv1.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	return &authv1.ValidateTokenResponse{
		Valid:  true,
		UserId: int64(claims.UserID),
		Email:  claims.Email,
	}, nil
}

func (s *AuthGRPCServer) GetMe(ctx context.Context, req *authv1.GetMeRequest) (*authv1.GetMeResponse, error) {
	if req == nil || req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}

	user, err := s.authSvc.GetUserByID(int(req.GetUserId()))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user: "+err.Error())
	}

	return &authv1.GetMeResponse{
		User: &authv1.User{
			Id:        int64(user.ID),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

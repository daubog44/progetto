package handler

import (
	"context"

	"log/slog"

	"github.com/username/progetto/auth/internal/service"
	"github.com/username/progetto/auth/internal/validator"
	authv1 "github.com/username/progetto/proto/gen/go/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	service service.AuthService
	logger  *slog.Logger
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  slog.Default().With("component", "auth_handler"),
	}
}

func (h *AuthHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if err := validator.ValidateRegister(req.Email, req.Password, req.Username); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := h.service.Register(ctx, req.Email, req.Password, req.Username)
	if err != nil {
		h.logger.Error("failed to register user", "error", err, "email", req.Email)
		return nil, status.Errorf(codes.Internal, "failed to register: %v", err)
	}

	return &authv1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	accessToken, refreshToken, expiresIn, err := h.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			h.logger.Warn("invalid login attempt", "email", req.Email)
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		h.logger.Error("failed to login", "error", err, "email", req.Email)
		return nil, status.Errorf(codes.Internal, "failed to login: %v", err)
	}

	return &authv1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (h *AuthHandler) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	accessToken, refreshToken, expiresIn, err := h.service.Refresh(ctx, req.RefreshToken)
	if err != nil {
		if err == service.ErrInvalidToken {
			h.logger.Warn("invalid refresh token attempt")
			return nil, status.Error(codes.Unauthenticated, "invalid or expired refresh token")
		}
		h.logger.Error("failed to refresh token", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to refresh token: %v", err)
	}

	return &authv1.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

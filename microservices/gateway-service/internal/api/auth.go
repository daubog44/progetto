package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	authv1 "github.com/username/progetto/proto/gen/go/auth/v1"
)

type RegisterInput struct {
	Body struct {
		Username string `json:"username" doc:"Username"`
		Email    string `json:"email" doc:"Email address"`
		Password string `json:"password" doc:"Password"`
	}
}

type RegisterOutput struct {
	Body struct {
		UserID       string `json:"user_id"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
}

type LoginInput struct {
	Body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
}

type LoginOutput struct {
	Body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
}

type RefreshInput struct {
	Body struct {
		RefreshToken string `json:"refresh_token"`
	}
}

type RefreshOutput struct {
	Body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
}

func RegisterAuthRoutes(api huma.API, client authv1.AuthServiceClient, logger *slog.Logger) {
	huma.Register(api, huma.Operation{
		OperationID: "register",
		Method:      http.MethodPost,
		Path:        "/auth/register",
		Summary:     "Register a new user",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, input *RegisterInput) (*RegisterOutput, error) {
		resp, err := client.Register(ctx, &authv1.RegisterRequest{
			Username: input.Body.Username,
			Email:    input.Body.Email,
			Password: input.Body.Password,
		})
		if err != nil {
			logger.ErrorContext(ctx, "register failed", "error", err)
			return nil, MapGRPCError(err)
		}
		out := &RegisterOutput{}
		out.Body.UserID = resp.UserId
		out.Body.AccessToken = resp.AccessToken
		out.Body.RefreshToken = resp.RefreshToken
		out.Body.ExpiresIn = resp.ExpiresIn
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Login user",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
		resp, err := client.Login(ctx, &authv1.LoginRequest{
			Email:    input.Body.Email,
			Password: input.Body.Password,
		})
		if err != nil {
			logger.ErrorContext(ctx, "login failed", "error", err, "email", input.Body.Email)
			return nil, MapGRPCError(err)
		}
		out := &LoginOutput{}
		out.Body.AccessToken = resp.AccessToken
		out.Body.RefreshToken = resp.RefreshToken
		out.Body.ExpiresIn = resp.ExpiresIn
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "refresh",
		Method:      http.MethodPost,
		Path:        "/auth/refresh",
		Summary:     "Refresh access token",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, input *RefreshInput) (*RefreshOutput, error) {
		resp, err := client.Refresh(ctx, &authv1.RefreshRequest{
			RefreshToken: input.Body.RefreshToken,
		})
		if err != nil {
			logger.ErrorContext(ctx, "refresh failed", "error", err)
			return nil, MapGRPCError(err)
		}
		out := &RefreshOutput{}
		out.Body.AccessToken = resp.AccessToken
		out.Body.RefreshToken = resp.RefreshToken
		out.Body.ExpiresIn = resp.ExpiresIn
		return out, nil
	})
}

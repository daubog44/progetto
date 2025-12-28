package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lithammer/shortuuid/v3"
	"github.com/username/progetto/auth/internal/model"
	"github.com/username/progetto/auth/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthService interface {
	Register(ctx context.Context, email, password, username string) (string, error)
	Login(ctx context.Context, email, password string) (string, string, int64, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, int64, error)
	CompensateUserCreation(ctx context.Context, userID string) error
}

type authService struct {
	userRepo        repository.UserRepository
	tokenRepo       repository.TokenRepository
	publisher       message.Publisher
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	publisher message.Publisher,
	jwtSecret string,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		publisher:       publisher,
		jwtSecret:       []byte(jwtSecret),
		accessTokenTTL:  15 * time.Minute,
		refreshTokenTTL: 6 * 30 * 24 * time.Hour, // ~6 months
	}
}

// Register creates a user and publishes an event
func (s *authService) Register(ctx context.Context, email, password, username string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Dynamic Username
	if username == "" {
		username = "user_" + shortuuid.New()
	}

	user := &model.User{
		Email:    email,
		Password: string(hashedPassword),
		Username: username,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return "", err
	}

	// Publish Event
	eventPayload := map[string]interface{}{
		"user_id":  fmt.Sprintf("%d", user.ID),
		"email":    user.Email,
		"username": user.Username,
	}
	payloadBytes, _ := json.Marshal(eventPayload)
	msg := message.NewMessage(watermill.NewUUID(), payloadBytes)
	msg.SetContext(ctx)

	if err := s.publisher.Publish("user_created", msg); err != nil {
		// Log error but don't fail registration? Ideally we should use outbox pattern, but for now simple publish
		return fmt.Sprintf("%d", user.ID), fmt.Errorf("failed to publish event: %w", err)
	}

	return fmt.Sprintf("%d", user.ID), nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, int64, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", 0, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", 0, ErrInvalidCredentials
	}

	return s.generateTokens(ctx, user.ID)
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (string, string, int64, error) {
	userIDStr, err := s.tokenRepo.GetUserIDByRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", 0, ErrInvalidToken
	}

	userID, _ := strconv.Atoi(userIDStr)

	// Rotate: Delete old
	if err := s.tokenRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		// If fail to delete, might be race condition or redis down.
		// Proceeding might be risky if we want strict rotation.
		return "", "", 0, err
	}

	return s.generateTokens(ctx, uint(userID))
}

func (s *authService) generateTokens(ctx context.Context, userID uint) (string, string, int64, error) {
	// Fetch User to get Role
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", "", 0, err
	}

	// Access Token (JWT)
	claims := jwt.MapClaims{
		"sub":  fmt.Sprintf("%d", userID),
		"role": user.Role,
		"exp":  time.Now().Add(s.accessTokenTTL).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", 0, err
	}

	// Refresh Token (Opaque)
	b := make([]byte, 32)
	rand.Read(b)
	refreshToken := base64.URLEncoding.EncodeToString(b)

	// Save Refresh Token
	err = s.tokenRepo.SetRefreshToken(ctx, refreshToken, fmt.Sprintf("%d", userID), s.refreshTokenTTL)
	if err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshToken, int64(s.accessTokenTTL.Seconds()), nil
}

func (s *authService) CompensateUserCreation(ctx context.Context, userID string) error {
	id, err := strconv.Atoi(userID)
	if err != nil {
		return err
	}
	return s.userRepo.Delete(ctx, uint(id))
}

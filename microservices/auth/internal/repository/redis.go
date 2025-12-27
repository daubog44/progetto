package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenRepository interface {
	SetRefreshToken(ctx context.Context, token string, userID string, duration time.Duration) error
	GetUserIDByRefreshToken(ctx context.Context, token string) (string, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

type redisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) TokenRepository {
	return &redisRepository{client: client}
}

func (r *redisRepository) SetRefreshToken(ctx context.Context, token string, userID string, duration time.Duration) error {
	return r.client.Set(ctx, "refresh:"+token, userID, duration).Err()
}

func (r *redisRepository) GetUserIDByRefreshToken(ctx context.Context, token string) (string, error) {
	return r.client.Get(ctx, "refresh:"+token).Result()
}

func (r *redisRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	return r.client.Del(ctx, "refresh:"+token).Err()
}

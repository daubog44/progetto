package presencestore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/username/progetto/shared/pkg/presence"
)

type Store struct {
	rdb *redis.Client
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

// GetUserPresence returns the user's presence record.
func (s *Store) GetUserPresence(ctx context.Context, userID string) (*presence.UserPresence, error) {
	key := fmt.Sprintf("user_presence:%s", userID)
	data, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // No record
	}
	if err != nil {
		return nil, err
	}

	var p presence.UserPresence
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetUserGateway returns the instance ID only if the user is "online".
func (s *Store) GetUserGateway(ctx context.Context, userID string) (string, error) {
	p, err := s.GetUserPresence(ctx, userID)
	if err != nil {
		return "", err
	}
	if p != nil && p.Status == "online" {
		return p.InstanceID, nil
	}
	return "", nil
}

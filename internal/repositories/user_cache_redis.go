package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Belixk/CommerceTwo/internal/entity"
	"github.com/redis/go-redis/v9"
)

type userCache struct {
	client *redis.Client
}

func NewUserCache(client *redis.Client) *userCache {
	return &userCache{client: client}
}

func (c *userCache) Get(ctx context.Context, key string) (*entity.User, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var user entity.User

	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *userCache) Set(ctx context.Context, key string, user *entity.User, ttl time.Duration) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

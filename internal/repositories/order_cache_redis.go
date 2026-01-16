package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Belixk/CommerceTwo/internal/entity"
	"github.com/redis/go-redis/v9"
)

type orderCache struct {
	client *redis.Client
}

// NewOrderCache создает новый экземпляр кеша для заказов
func NewOrderCache(client *redis.Client) *orderCache {
	return &orderCache{client: client}
}

func (c *orderCache) Get(ctx context.Context, key string) (*entity.Order, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err // Если ключа нет, вернется redis.Nil
	}

	var order entity.Order
	if err := json.Unmarshal([]byte(val), &order); err != nil {
		return nil, err
	}

	return &order, nil
}

func (c *orderCache) Set(ctx context.Context, key string, order *entity.Order, ttl time.Duration) error {
	// Если передали nil, это может быть сигналом к удалению или инвалидации
	if order == nil {
		return c.client.Del(ctx, key).Err()
	}

	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

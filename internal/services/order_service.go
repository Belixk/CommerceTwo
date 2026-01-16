package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Belixk/CommerceTwo/internal/entity"
	"github.com/Belixk/CommerceTwo/internal/repositories"
)

var ErrOrderNil = errors.New("order not must be a nil")

type Cache interface {
	Get(ctx context.Context, key string) (*entity.Order, error)
	Set(ctx context.Context, key string, order *entity.Order, ttl time.Duration) error
}

type OrderService struct {
	repo  repositories.OrderRepository
	cache Cache
}

func NewOrderService(repo repositories.OrderRepository, cache Cache) *OrderService {
	return &OrderService{
		repo:  repo,
		cache: cache,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	if order == nil {
		return nil, ErrOrderNil
	}

	if len(order.Items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	var total int64
	for _, item := range order.Items {
		total += item.Price * int64(item.Quantity)
	}
	order.Total = total

	return s.repo.CreateOrder(ctx, order)
}

func (s *OrderService) GetOrderByID(ctx context.Context, id int64) (*entity.Order, error) {
	key := fmt.Sprintf("order:%d", id)

	if order, err := s.cache.Get(ctx, key); err == nil && order != nil {
		return order, nil
	}

	order, err := s.repo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, key, order, 15*time.Minute)

	return order, nil
}

func (s *OrderService) GetOrderbyUserID(ctx context.Context, user_id int64) (*entity.Order, error) {
	key := fmt.Sprintf("order:%d", user_id)

	if order, err := s.cache.Get(ctx, key); err == nil && order != nil {
		return order, nil
	}

	order, err := s.repo.GetOrderByUserID(ctx, user_id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, key, order, 15*time.Minute)

	return order, nil
}

package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Belixk/CommerceTwo/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOrderRepo struct{ mock.Mock }

func (m *MockOrderRepo) CreateOrder(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	args := m.Called(ctx, order)
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepo) GetOrderByID(ctx context.Context, id int64) (*entity.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepo) GetOrderByUserID(ctx context.Context, user_id int64) (*entity.Order, error) {
	args := m.Called(ctx, user_id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepo) UpdateOrder(ctx context.Context, order *entity.Order) error {
	return m.Called(ctx, order).Error(0)
}

func (m *MockOrderRepo) DeleteOrderByID(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

type MockOrderCache struct{ mock.Mock }

func (m *MockOrderCache) Get(ctx context.Context, key string) (*entity.Order, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderCache) Set(ctx context.Context, key string, order *entity.Order, ttl time.Duration) error {
	args := m.Called(ctx, key, order, ttl)
	return args.Error(0)
}

func TestOrderService_CreateOrder(t *testing.T) {
	repo := new(MockOrderRepo)
	cache := new(MockOrderCache)
	service := NewOrderService(repo, cache)

	t.Run("should calculate total correctly", func(t *testing.T) {
		order := &entity.Order{
			Items: []entity.OrderItem{
				{Price: 100, Quantity: 2},
				{Price: 50, Quantity: 3},
			},
		}

		repo.On("CreateOrder", mock.Anything, mock.MatchedBy(func(o *entity.Order) bool {
			return o.Total == 350
		})).Return(order, nil)

		res, err := service.CreateOrder(context.Background(), order)

		assert.NoError(t, err)
		assert.Equal(t, int64(350), res.Total)
		repo.AssertExpectations(t)
	})

	t.Run("should return if no items", func(t *testing.T) {
		order := &entity.Order{Items: []entity.OrderItem{}}

		res, err := service.CreateOrder(context.Background(), order)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "order must be have at least one item", err.Error())
	})
}

func TestOrderService_GetOrderByID(t *testing.T) {
	repo := new(MockOrderRepo)
	cache := new(MockOrderCache)
	service := NewOrderService(repo, cache)
	ctx := context.Background()
	orderID := int64(1)
	cacheKey := "order:1"

	t.Run("cache hit - should not call repo", func(t *testing.T) {
		expectedOrder := &entity.Order{ID: orderID, Total: 500}

		// Настраиваем кеш: он должен вернуть заказ
		cache.On("Get", ctx, cacheKey).Return(expectedOrder, nil)

		res, err := service.GetOrderByID(ctx, orderID)

		assert.NoError(t, err)
		assert.Equal(t, expectedOrder, res)
		// Проверяем, что CreateOrder НЕ вызывался (так как данные из кеша)
		repo.AssertNotCalled(t, "GetOrderByID", mock.Anything, mock.Anything)
	})

	t.Run("cache miss - should call repo and set cache", func(t *testing.T) {
		// Очищаем ожидания от предыдущего теста
		cache.ExpectedCalls = nil

		expectedOrder := &entity.Order{ID: orderID, Total: 500}

		// 1. Кеш возвращает ошибку или nil
		cache.On("Get", ctx, cacheKey).Return(nil, errors.New("not found"))
		// 2. Репозиторий возвращает данные
		repo.On("GetOrderByID", ctx, orderID).Return(expectedOrder, nil)
		// 3. Сервис должен сохранить данные в кеш
		cache.On("Set", ctx, cacheKey, expectedOrder, mock.Anything).Return(nil)

		res, err := service.GetOrderByID(ctx, orderID)

		assert.NoError(t, err)
		assert.Equal(t, expectedOrder, res)
		repo.AssertExpectations(t)
		cache.AssertExpectations(t)
	})
}

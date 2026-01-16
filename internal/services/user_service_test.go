package services

import (
	"context"
	"testing"
	"time"

	"github.com/Belixk/CommerceTwo/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	return nil, nil
}
func (m *MockUserRepo) Update(ctx context.Context, user *entity.User) error { return nil }
func (m *MockUserRepo) Delete(ctx context.Context, id int64) error          { return nil }

type MockHasher struct{ mock.Mock }

func (m *MockHasher) Hash(p string) (string, error) {
	args := m.Called(p)
	return args.String(0), args.Error(1)
}
func (m *MockHasher) Compare(h, p string) error { return nil }

type MockCache struct{ mock.Mock }

func (m *MockCache) Get(ctx context.Context, k string) (*entity.User, error) {
	args := m.Called(ctx, k)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, k string, u *entity.User, t time.Duration) error {
	return nil
}

func TestUserService_CreateUser(t *testing.T) {
	repo := new(MockUserRepo)
	hasher := new(MockHasher)
	cache := new(MockCache)
	service := NewUserService(repo, cache, hasher)

	ctx := context.Background()

	t.Run("short password error", func(t *testing.T) {
		user := &entity.User{Email: "test@test.com"}
		res, err := service.CreateUser(ctx, user, "123")

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "password is too short", err.Error())
	})

	t.Run("succes create", func(t *testing.T) {
		user := &entity.User{FirstName: "Maxim", Email: "maxim@test.com"}
		password := "123456"
		hashedPassword := "hashed_123456"

		hasher.On("Hash", password).Return(hashedPassword, nil)
		repo.On("Create", ctx, mock.Anything).Return(user, nil)

		res, err := service.CreateUser(ctx, user, password)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, hashedPassword, user.PasswordHash)

		repo.AssertExpectations(t)
		hasher.AssertExpectations(t)
	})
}

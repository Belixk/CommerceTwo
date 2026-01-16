package services

import (
	"context"
	"testing"

	"github.com/Belixk/CommerceTwo/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	return nil, nil
}
func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error { return nil }
func (m *MockUserRepository) Delete(ctx context.Context, id int64) error          { return nil }

func TestCreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, nil)

	user := &entity.User{
		FirstName: "Maxim",
		LastName:  "Antonov",
		Email:     "test@test.com",
		Age:       19,
	}

	mockRepo.On("Create", mock.Anything, user).Return(user, nil)
	result, err := service.CreateUser(context.Background(), user, "password123")

	assert.NoError(t, err)
	assert.Equal(t, "Maxim", result.FirstName)
	mockRepo.AssertExpectations(t)
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, nil)

	// Создаем юзера с плохим Email (без собачки)
	user := &entity.User{
		FirstName: "Maxim",
		LastName:  "Antonov",
		Email:     "invalid-email",
		Age:       19,
	}

	// Вызываем метод
	result, err := service.CreateUser(context.Background(), user, "password123")

	// ПРОВЕРЯЕМ:
	// 1. Должна быть ошибка
	assert.Error(t, err)
	// 2. Ошибка должна быть именно про Email (мы её в entity прописали)
	assert.Equal(t, entity.ErrInvalidEmail, err)
	// 3. Результат должен быть nil
	assert.Nil(t, result)

	// ПРОВЕРЯЕМ: репозиторий НЕ должен был вызываться,
	// так как валидация зафейлилась раньше.
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

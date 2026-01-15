package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Belixk/CommerceTwo/internal/entity"
	"github.com/Belixk/CommerceTwo/internal/repositories"
)

type UserCache interface {
	Get(ctx context.Context, key string) (*entity.User, error)
	Set(ctx context.Context, key string, user *entity.User, ttl time.Duration) error
}

type UserService struct {
	repo  repositories.UserRepository
	cache UserCache
}

func NewUserService(repo repositories.UserRepository, cache UserCache) *UserService {
	return &UserService{
		repo:  repo,
		cache: cache,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *entity.User, password string) (*entity.User, error) {
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, user)
}

func (s *UserService) GetUserById(ctx context.Context, id int64) (*entity.User, error) {
	key := fmt.Sprintf("user:%d", id)

	if user, err := s.cache.Get(ctx, key); err == nil && user != nil {
		return user, nil
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, key, user, 15*time.Minute)

	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	key := fmt.Sprintf("user:%s", email)

	if user, err := s.cache.Get(ctx, key); err == nil && user != nil {
		return user, nil
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, key, user, time.Minute*15)
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *entity.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return err
	}

	key := fmt.Sprintf("user:%d", user.ID)
	_ = s.cache.Set(ctx, key, user, 15*time.Minute)

	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	_ = s.cache.Set(ctx, fmt.Sprintf("user:%d", id), nil, 0)
	return s.repo.Delete(ctx, id)
}

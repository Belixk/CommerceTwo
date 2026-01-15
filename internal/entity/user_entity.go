package entity

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidFirstName = errors.New("invalid first name")
	ErrInvalidLastName  = errors.New("invalid last name")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrInvalidAge       = errors.New("invalid age")
)

type User struct {
	ID           int64     `json:"id" db:"id"`
	FirstName    string    `json:"first_name" db:"first_name" binding:"required"`
	LastName     string    `json:"last_name" db:"last_name" binding:"required"`
	Email        string    `json:"email" db:"email" binding:"required,email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Age          int       `json:"age" db:"age"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

func (u *User) Validate() error {
	u.FirstName = strings.TrimSpace(u.FirstName)
	u.LastName = strings.TrimSpace(u.LastName)
	u.Email = strings.TrimSpace(u.Email)

	if u.FirstName == "" {
		return ErrInvalidFirstName
	}
	if u.LastName == "" {
		return ErrInvalidLastName
	}
	if !strings.Contains(u.Email, "@") {
		return ErrInvalidEmail
	}
	if u.Age < 0 || u.Age > 125 {
		return ErrInvalidAge
	}

	return nil
}

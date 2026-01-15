package entity

import "time"

type Order struct {
	ID int64 `json:"id" db:"id"`
	UserID int64 `json:"user_id" db:"user_id"`
	Items []OrderItem `json:"items" db:"items"`
	Amount float64 `json:"amount" db:"amount"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
	ID int64 `json:"id" db:"id"`
	OrderID int64 `json:"order_id" db:"order_id"`
	Name string `json:"name" db:"name"`
	Amount float64 `json:"amount" db:"amount"`
}

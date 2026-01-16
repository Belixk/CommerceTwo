package entity

import "time"

type Order struct {
	ID        int64       `json:"id" db:"id"`
	UserID    int64       `json:"user_id" db:"user_id"`
	Items     []OrderItem `json:"items" db:"-"`
	Total     int64       `json:"total" db:"total"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
	ID       int64  `json:"id" db:"id"`
	OrderID  int64  `json:"order_id" db:"order_id"`
	Name     string `json:"name" db:"name"`
	Quantity int    `json:"quantity" db:"quantity"`
	Price    int64  `json:"price" db:"price"`
}

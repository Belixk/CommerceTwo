package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Belixk/CommerceTwo/internal/entity"

	"github.com/jmoiron/sqlx"
)

var ErrOrderNotFound = errors.New("order with this id not found")

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *entity.Order) (*entity.Order, error)
	GetOrderByID(ctx context.Context, id int64) (*entity.Order, error)
	GetOrderByUserID(ctx context.Context, id int64) (*entity.Order, error)
	UpdateOrder(ctx context.Context, order *entity.Order) error
	DeleteOrderByID(ctx context.Context, id int64) error
}

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	query := `
		INSERT INTO order (items, amount, created_at, updated_at)
		VALUES($1, $2, NOW(), NOW())
		RETURNING id, user_id, created_at, updated_at
	`
	err := r.db.QueryRowxContext(
		ctx,
		query,
		order.ID,
		order.UserID,
		order.Items,
		order.Amount,
		order.CreatedAt,
		order.UpdatedAt,
	).StructScan(order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *orderRepository) GetOrderByID(ctx context.Context, id int64) (*entity.Order, error) {
	var order *entity.Order
	query := `
		SELECT id, user_id, items, amount, created_at, updated_at
		FROM order
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, order, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	return order, nil
}

func (r *orderRepository) GetOrderByUserID(ctx context.Context, user_id int64) (*entity.Order, error) {
	var order *entity.Order
	query := `
		SELECT id, user_id, items, amount, created_at, updated_at
		FROM order
		WHERE user_id = $1
	`
	err := r.db.GetContext(ctx, order, query, user_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	return order, nil
}

func (r *orderRepository) UpdateOrder(ctx context.Context, order *entity.Order) error {
	query := `
		UPDATE order
		SET id = :id, user_id = :user_id, items = :items, amount = :amount , updated_at = Now()
		WHERE id = :id
	`
	result, err := r.db.NamedExecContext(ctx, query, order)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return err
	}

	return nil
}

func (r *orderRepository) DeleteOrderByID(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return err
	}

	return nil
}

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
	// Запускаем транзакция
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // кидаем в отложеное срабатывает откат бд, если вдруг что-то пойдёт не так
	// Готовим сами товары и сохраняем
	queryOrder := `
		INSERT INTO orders (user_id, total, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRowxContext(ctx, queryOrder, order.UserID, order.Total).StructScan(order)
	if err != nil {
		return nil, err
	}

	// Теперь готовим сами предметы в заказе
	queryItem := `
		INSERT INTO order_items (order_id, name, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	// Проходимся в цикле, чтобы привязать товар к order.ID
	for i := range order.Items {
		order.Items[i].OrderID = order.ID

		err = tx.QueryRowxContext(
			ctx, queryItem,
			order.Items[i].OrderID,
			order.Items[i].Name,
			order.Items[i].Quantity,
			order.Items[i].Price,
		).Scan(&order.Items[i].ID)
		if err != nil {
			return nil, err
		}
	}

	// Если всё успешно прошло публикуем изменение в бд
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return order, nil
}

func (r *orderRepository) GetOrderByID(ctx context.Context, id int64) (*entity.Order, error) {
	// Берём сам заказ
	var order entity.Order

	queryOrder := `
		SELECT id, user_id, total, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, &order, queryOrder, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	// После берём сами предметы в заказе
	var items []entity.OrderItem

	queryItems := `SELECT id, order_id, name, quantity, price FROM order_items WHERE order_id = $1`
	err = r.db.SelectContext(ctx, &items, queryItems, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	order.Items = items
	return &order, nil
}

func (r *orderRepository) GetOrderByUserID(ctx context.Context, user_id int64) (*entity.Order, error) {
	// Сначала берём сам заказ
	var order entity.Order

	queryOrder := `
		SELECT id, user_id, total, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 1
	`
	err := r.db.GetContext(ctx, &order, queryOrder, user_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	// Теперь берём предметы в заказе
	var items []entity.OrderItem

	queryItem := `SELECT id, order_id, name, quantity, price FROM order_items WHERE order_id = $1`
	err = r.db.SelectContext(ctx, &items, queryItem, order.ID)
	if err != nil {
		return nil, err
	}

	order.Items = items
	return &order, nil
}

func (r *orderRepository) UpdateOrder(ctx context.Context, order *entity.Order) error {
	query := `
		UPDATE orders
		SET total = $1, updated_at = NOW()
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, order.Total, order.ID)
	if err != nil {
		return err
	}
	// Проверка на изменение строк, если 0, возвращаем, что не найден заказ
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrOrderNotFound
	}

	return nil
}

func (r *orderRepository) DeleteOrderByID(ctx context.Context, id int64) error {
	// начинаем транзакцию
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback() // Откатываем бд, если вдруг произошла ошибка

	// Сначала удаляем предметы в заказе
	_, err = tx.ExecContext(ctx, "DELETE FROM order_items WHERE order_id = $1", id)
	if err != nil {
		return err
	}

	// Теперь удаляем сам заказ
	result, err := tx.ExecContext(ctx, "DELETE FROM orders WHERE id = $1", id)
	if err != nil {
		return err
	}

	// Проверка на изменение строк, если 0, возвращаем, что не найден заказ
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrOrderNotFound
	}

	// Публикуем изменение в бд
	return tx.Commit()
}

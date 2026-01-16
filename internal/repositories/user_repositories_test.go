package repositories

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewUserRepository(sqlxDB)

	t.Run("success", func(t *testing.T) {
		id := int64(1)
		rows := sqlmock.NewRows([]string{"id", "first_name", "email"}).AddRow(id, "Maxim", "max@test.com")

		mock.ExpectQuery("SELECT (.+) FROM users WHERE id = \\$1").WithArgs(id).WillReturnRows(rows)

		user, err := repo.GetByID(context.Background(), id)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "Maxim", user.FirstName)
	})

	t.Run("not found", func(t *testing.T) {
		id := int64(404)
		mock.ExpectQuery("SELECT (.+) FROM users WHERE id = \\$1").WithArgs(id).WillReturnError(sql.ErrNoRows)
		user, err := repo.GetByID(context.Background(), id)
		
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrUserNotFound, err)
	})
}
package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"gobankapi/internal/models"

	"github.com/lib/pq"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// Создание пользователя
func (r *UserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (email, username, password_hash)
			  VALUES ($1, $2, $3)
			  RETURNING id, created_at`

	err := r.DB.QueryRow(query, user.Email, user.Username, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("пользователь с таким email или username уже существует")
			}
		}
	}

	return err
}

// Поиск по email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, created_at
		FROM users
		WHERE email = $1
	`
	user := &models.User{}
	err := r.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // пользователь не найден — это не ошибка
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

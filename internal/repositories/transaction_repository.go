package repositories

import (
	"database/sql"
	"gobankapi/internal/models"
)

type TransactionRepository struct {
	DB *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{DB: db}
}

func (r *TransactionRepository) Log(tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (from_account_id, to_account_id, amount, type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.DB.QueryRow(query, tx.FromAccountID, tx.ToAccountID, tx.Amount, tx.Type).
		Scan(&tx.ID, &tx.CreatedAt)
}

func (r *TransactionRepository) GetMonthlySummary(userID int) (float64, float64, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN amount < 0 THEN amount ELSE 0 END), 0) AS expenses
		FROM (
			SELECT 
				t.amount, 
				a.user_id
			FROM transactions t
			JOIN accounts a ON 
				t.from_account_id = a.id OR t.to_account_id = a.id
			WHERE DATE_TRUNC('month', t.created_at) = DATE_TRUNC('month', CURRENT_DATE)
		) AS subquery
		WHERE user_id = $1
	`

	var income, expenses float64
	err := r.DB.QueryRow(query, userID).Scan(&income, &expenses)
	return income, expenses, err
}

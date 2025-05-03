package repositories

import (
	"database/sql"
	"gobankapi/internal/models"
)

type CreditRepository struct {
	DB *sql.DB
}

func NewCreditRepository(db *sql.DB) *CreditRepository {
	return &CreditRepository{DB: db}
}

func (r *CreditRepository) Create(credit *models.Credit) error {
	query := `
		INSERT INTO credits (user_id, account_id, amount, term_months, annual_rate, monthly_payment)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return r.DB.QueryRow(query,
		credit.UserID,
		credit.AccountID,
		credit.Amount,
		credit.TermMonths,
		credit.AnnualRate,
		credit.MonthlyPayment,
	).Scan(&credit.ID, &credit.CreatedAt)
}

func (r *CreditRepository) GetActiveCreditLoad(userID int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(c.amount), 0)
		FROM credits c
		WHERE c.user_id = $1
		  AND EXISTS (
			  SELECT 1 FROM payment_schedules ps
			  WHERE ps.credit_id = c.id AND ps.paid = false
		  )
	`

	var total float64
	err := r.DB.QueryRow(query, userID).Scan(&total)
	return total, err
}

package repositories

import (
	"database/sql"
	"gobankapi/internal/models"
	"time"
)

type PaymentScheduleRepository struct {
	DB *sql.DB
}

func NewPaymentScheduleRepository(db *sql.DB) *PaymentScheduleRepository {
	return &PaymentScheduleRepository{DB: db}
}

func (r *PaymentScheduleRepository) Create(schedule *models.PaymentSchedule) error {
	query := `
		INSERT INTO payment_schedules (credit_id, due_date, amount, paid, penalty)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return r.DB.QueryRow(query,
		schedule.CreditID,
		schedule.DueDate,
		schedule.Amount,
		schedule.Paid,
		schedule.Penalty,
	).Scan(&schedule.ID)
}

func (r *PaymentScheduleRepository) FindByCreditID(creditID int) ([]*models.PaymentSchedule, error) {
	query := `
		SELECT id, credit_id, due_date, amount, paid, paid_at, penalty
		FROM payment_schedules
		WHERE credit_id = $1
		ORDER BY due_date ASC
	`
	rows, err := r.DB.Query(query, creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.PaymentSchedule
	for rows.Next() {
		var p models.PaymentSchedule
		err := rows.Scan(&p.ID, &p.CreditID, &p.DueDate, &p.Amount, &p.Paid, &p.PaidAt, &p.Penalty)
		if err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, nil
}

func (r *PaymentScheduleRepository) GetScheduledPayments(accountID int, until time.Time) (float64, error) {
	query := `
		SELECT COALESCE(SUM(ps.amount + ps.penalty), 0)
		FROM payment_schedules ps
		JOIN credits c ON ps.credit_id = c.id
		WHERE c.account_id = $1
		  AND ps.paid = false
		  AND ps.due_date <= $2
	`

	var total float64
	err := r.DB.QueryRow(query, accountID, until).Scan(&total)
	return total, err
}

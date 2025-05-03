package models

import "time"

type Credit struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	AccountID      int       `json:"account_id"`
	Amount         float64   `json:"amount"`
	TermMonths     int       `json:"term_months"`
	AnnualRate     float64   `json:"annual_rate"`
	MonthlyPayment float64   `json:"monthly_payment"`
	CreatedAt      time.Time `json:"created_at"`
}

type PaymentSchedule struct {
	ID       int        `json:"id"`
	CreditID int        `json:"credit_id"`
	DueDate  time.Time  `json:"due_date"`
	Amount   float64    `json:"amount"`
	Paid     bool       `json:"paid"`
	PaidAt   *time.Time `json:"paid_at,omitempty"`
	Penalty  float64    `json:"penalty"`
}

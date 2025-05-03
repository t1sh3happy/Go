package models

import "time"

type Transaction struct {
	ID            int       `json:"id"`
	FromAccountID *int      `json:"from_account_id,omitempty"`
	ToAccountID   *int      `json:"to_account_id,omitempty"`
	Amount        float64   `json:"amount"`
	Type          string    `json:"type"` // "deposit", "withdraw", "transfer"
	CreatedAt     time.Time `json:"created_at"`
}

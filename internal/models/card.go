package models

import "time"

type Card struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	AccountID int       `json:"account_id"`
	NumberPGP string    `json:"-"` // не выводим напрямую
	ExpiryPGP string    `json:"-"`
	CVVHash   string    `json:"-"`
	HMAC      string    `json:"hmac"`
	CreatedAt time.Time `json:"created_at"`
}

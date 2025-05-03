package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // json:"-" означает, что PasswordHash не попадёт в JSON-ответы.
	CreatedAt    time.Time `json:"created_at"`
}

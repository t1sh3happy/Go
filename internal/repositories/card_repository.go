package repositories

import (
	"database/sql"
	"gobankapi/internal/models"
)

type CardRepository struct {
	DB *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{DB: db}
}

func (r *CardRepository) Create(card *models.Card) error {
	query := `
		INSERT INTO cards (user_id, account_id, number_pgp, expiry_pgp, cvv_hash, hmac)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return r.DB.QueryRow(query,
		card.UserID,
		card.AccountID,
		card.NumberPGP,
		card.ExpiryPGP,
		card.CVVHash,
		card.HMAC,
	).Scan(&card.ID, &card.CreatedAt)
}

func (r *CardRepository) FindByUserID(userID int) ([]*models.Card, error) {
	query := `
		SELECT id, user_id, account_id, number_pgp, expiry_pgp, hmac, created_at
		FROM cards
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		var card models.Card
		err := rows.Scan(&card.ID, &card.UserID, &card.AccountID, &card.NumberPGP, &card.ExpiryPGP, &card.HMAC, &card.CreatedAt)
		if err != nil {
			return nil, err
		}
		cards = append(cards, &card)
	}
	return cards, nil
}

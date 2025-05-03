package repositories

import (
	"database/sql"
	"errors"
	"gobankapi/internal/models"
)

type AccountRepository struct {
	DB *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{DB: db}
}

func (r *AccountRepository) Create(account *models.Account) error {
	query := `
		INSERT INTO accounts (user_id, number, balance)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.DB.QueryRow(query, account.UserID, account.Number, account.Balance).
		Scan(&account.ID, &account.CreatedAt)
	return err
}

func (r *AccountRepository) FindByUserID(userID int) ([]*models.Account, error) {
	query := `
		SELECT id, user_id, number, balance, created_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*models.Account
	for rows.Next() {
		var acc models.Account
		err := rows.Scan(&acc.ID, &acc.UserID, &acc.Number, &acc.Balance, &acc.CreatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, &acc)
	}

	return accounts, nil
}

func (r *AccountRepository) UpdateBalance(accountID, userID int, amount float64) error {
	query := `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2 AND user_id = $3
	`
	result, err := r.DB.Exec(query, amount, accountID, userID)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *AccountRepository) Transfer(fromID, toID, userID int, amount float64) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Проверка баланса отправителя
	var balance float64
	err = tx.QueryRow(`SELECT balance FROM accounts WHERE id = $1 AND user_id = $2`, fromID, userID).Scan(&balance)
	if err != nil {
		return err
	}
	if balance < amount {
		return errors.New("insufficient funds")
	}

	// Списание и зачисление
	_, err = tx.Exec(`UPDATE accounts SET balance = balance - $1 WHERE id = $2`, amount, fromID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE id = $2`, amount, toID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *AccountRepository) GetBalance(accountID int, userID int) (float64, error) {
	query := `SELECT balance FROM accounts WHERE id = $1 AND user_id = $2`
	var balance float64
	err := r.DB.QueryRow(query, accountID, userID).Scan(&balance)
	return balance, err
}

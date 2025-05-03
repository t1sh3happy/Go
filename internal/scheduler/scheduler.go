package scheduler

import (
	"database/sql"
	"log"
	"time"
)

// Старт автоматического шедулера
func StartScheduler(db *sql.DB, intervalHours int) {
	log.Printf("Шедулер запущен. Интервал: %d часов.\n", intervalHours)

	ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
	defer ticker.Stop()

	// Первый запуск сразу
	Run(db)

	for {
		select {
		case <-ticker.C:
			Run(db)
		}
	}
}

// Одноразовая обработка платежей
func Run(db *sql.DB) {
	log.Println("Запуск обработки просроченных платежей...")

	query := `
		SELECT ps.id, ps.credit_id, ps.amount, c.account_id
		FROM payment_schedules ps
		JOIN credits c ON ps.credit_id = c.id
		WHERE ps.paid = false AND ps.due_date < CURRENT_DATE
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Println("Ошибка запроса платежей:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			paymentID int
			creditID  int
			amount    float64
			accountID int
			balance   float64
		)

		err := rows.Scan(&paymentID, &creditID, &amount, &accountID)
		if err != nil {
			log.Println("Ошибка сканирования:", err)
			continue
		}

		err = db.QueryRow("SELECT balance FROM accounts WHERE id = $1", accountID).Scan(&balance)
		if err != nil {
			log.Println("Ошибка получения баланса:", err)
			continue
		}

		if balance >= amount {
			tx, err := db.Begin()
			if err != nil {
				log.Println("Ошибка начала транзакции:", err)
				continue
			}

			_, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, accountID)
			if err == nil {
				_, err = tx.Exec(`
					UPDATE payment_schedules
					SET paid = true, paid_at = $1
					WHERE id = $2
				`, time.Now(), paymentID)
			}

			if err != nil {
				log.Println("Ошибка транзакции списания:", err)
				tx.Rollback()
			} else {
				tx.Commit()
				log.Printf("Платёж #%d успешно списан\n", paymentID)
			}
		} else {
			penalty := amount * 0.10
			_, err := db.Exec(`
				UPDATE payment_schedules
				SET penalty = penalty + $1
				WHERE id = $2
			`, penalty, paymentID)

			if err != nil {
				log.Println("Ошибка начисления штрафа:", err)
			} else {
				log.Printf("Недостаточно средств. Платёж #%d: начислен штраф %.2f\n", paymentID, penalty)
			}
		}
	}

	log.Println("Шедулер завершил обработку.")
}

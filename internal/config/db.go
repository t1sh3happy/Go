package config

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("postgres", AppConfig.DB_DSN)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("БД недоступна: %v", err)
	}

	log.Println("Подключение к БД успешно")
}

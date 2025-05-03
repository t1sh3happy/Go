package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	JWTSecret string
	DB_DSN    string
	SMTPHost  string
	SMTPPort  string
	SMTPUser  string
	SMTPPass  string
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env не найден — будут использованы переменные окружения")
	}

	AppConfig = &Config{
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "defaultsecret"),
		DB_DSN:    getEnv("DB_DSN", ""),
		SMTPHost:  getEnv("SMTP_HOST", ""),
		SMTPPort:  getEnv("SMTP_PORT", ""),
		SMTPUser:  getEnv("SMTP_USER", ""),
		SMTPPass:  getEnv("SMTP_PASS", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

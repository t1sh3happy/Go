package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/go-mail/mail/v2"
)

type Mailer struct {
	dialer *mail.Dialer
	from   string
}

func NewMailer() *Mailer {
	host := os.Getenv("SMTP_HOST")
	port := getEnvAsInt("SMTP_PORT", 587)
	from := os.Getenv("SMTP_USER")
	login := os.Getenv("SMTP_LOGIN")
	pass := os.Getenv("SMTP_PASS")

	d := mail.NewDialer(host, port, login, pass)
	d.TLSConfig = &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	return &Mailer{
		dialer: d,
		from:   from,
	}
}

func (m *Mailer) SendPaymentConfirmation(to string, amount float64) error {
	content := fmt.Sprintf(`
		<h1>Спасибо за оплату!</h1>
		<p>Сумма: <strong>%.2f RUB</strong></p>
		<small>Это автоматическое уведомление</small>
	`, amount)

	msg := mail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", "Платеж успешно проведен")
	msg.SetBody("text/html", content)

	if err := m.dialer.DialAndSend(msg); err != nil {
		log.Printf("SMTP error: %v", err)
		return fmt.Errorf("email sending failed")
	}

	log.Printf("Email sent to %s", to)
	return nil
}

// вспомогательная функция
func getEnvAsInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		var num int
		fmt.Sscanf(val, "%d", &num)
		return num
	}
	return defaultVal
}

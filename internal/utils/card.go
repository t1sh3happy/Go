package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Генерация валидного номера карты (по алгоритму Луна)
func GenerateCardNumber() string {
	base := make([]int, 15)
	for i := 0; i < 15; i++ {
		base[i] = rand.Intn(10)
	}
	// Алгоритм Луна
	sum := 0
	for i := 0; i < 15; i++ {
		n := base[14-i]
		if i%2 == 0 {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
	}
	checkDigit := (10 - (sum % 10)) % 10
	base = append(base, checkDigit)

	cardNumber := ""
	for _, digit := range base {
		cardNumber += fmt.Sprint(digit)
	}
	return cardNumber
}

func GenerateExpiryDate() string {
	expiry := time.Now().AddDate(3, 0, 0) // +3 года
	return expiry.Format("01/06")         // MM/YY
}

func GenerateCVV() string {
	return fmt.Sprintf("%03d", rand.Intn(1000))
}

func HashCVV(cvv string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	return string(hash), err
}

func ComputeHMAC(data string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

package utils

import (
	"math"
)

// Расчёт аннуитетного ежемесячного платежа
func CalculateAnnuity(amount float64, annualRate float64, months int) float64 {
	if annualRate == 0 {
		return amount / float64(months)
	}
	monthlyRate := annualRate / 12 / 100
	payment := amount * (monthlyRate * math.Pow(1+monthlyRate, float64(months))) /
		(math.Pow(1+monthlyRate, float64(months)) - 1)
	return math.Round(payment*100) / 100
}

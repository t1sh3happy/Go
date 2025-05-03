package handlers

import (
	"encoding/json"
	"gobankapi/internal/middleware"
	"gobankapi/internal/models"
	"gobankapi/internal/repositories"
	"gobankapi/internal/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type CreditHandler struct {
	CreditRepo   *repositories.CreditRepository
	ScheduleRepo *repositories.PaymentScheduleRepository
}

func NewCreditHandler(c *repositories.CreditRepository, s *repositories.PaymentScheduleRepository) *CreditHandler {
	return &CreditHandler{
		CreditRepo:   c,
		ScheduleRepo: s,
	}
}

type CreateCreditRequest struct {
	AccountID  int     `json:"account_id"`
	Amount     float64 `json:"amount"`
	TermMonths int     `json:"term_months"`
	AnnualRate float64 `json:"annual_rate"`
}

func (h *CreditHandler) CreateCredit(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := strconv.Atoi(userIDStr)

	var req CreateCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	monthlyPayment := utils.CalculateAnnuity(req.Amount, req.AnnualRate, req.TermMonths)

	credit := &models.Credit{
		UserID:         userID,
		AccountID:      req.AccountID,
		Amount:         req.Amount,
		TermMonths:     req.TermMonths,
		AnnualRate:     req.AnnualRate,
		MonthlyPayment: monthlyPayment,
	}

	err := h.CreditRepo.Create(credit)
	if err != nil {
		http.Error(w, "Could not create credit", http.StatusInternalServerError)
		return
	}

	// Сгенерируем график платежей
	for i := 1; i <= req.TermMonths; i++ {
		schedule := &models.PaymentSchedule{
			CreditID: credit.ID,
			DueDate:  time.Now().AddDate(0, i, 0), // каждый месяц
			Amount:   monthlyPayment,
			Paid:     false,
			Penalty:  0,
		}
		_ = h.ScheduleRepo.Create(schedule) // упрощённо без обработки ошибки
	}

	resp := map[string]interface{}{
		"credit_id":       credit.ID,
		"monthly_payment": monthlyPayment,
		"created_at":      credit.CreatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *CreditHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creditID, err := strconv.Atoi(vars["creditId"])
	if err != nil {
		http.Error(w, "Invalid credit ID", http.StatusBadRequest)
		return
	}

	schedule, err := h.ScheduleRepo.FindByCreditID(creditID)
	if err != nil {
		http.Error(w, "Could not fetch schedule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

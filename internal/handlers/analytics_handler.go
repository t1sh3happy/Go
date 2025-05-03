package handlers

import (
	"encoding/json"
	"gobankapi/internal/middleware"
	"gobankapi/internal/repositories"
	"net/http"
	"strconv"
)

type AnalyticsHandler struct {
	TransactionRepo *repositories.TransactionRepository
	CreditRepo      *repositories.CreditRepository
}

func NewAnalyticsHandler(txRepo *repositories.TransactionRepository, creditRepo *repositories.CreditRepository) *AnalyticsHandler {
	return &AnalyticsHandler{
		TransactionRepo: txRepo,
		CreditRepo:      creditRepo,
	}
}

func (h *AnalyticsHandler) GetCreditLoad(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := strconv.Atoi(userIDStr)

	total, err := h.CreditRepo.GetActiveCreditLoad(userID)
	if err != nil {
		http.Error(w, "Could not fetch credit load", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"active_credit_load": total,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

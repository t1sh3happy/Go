package handlers

import (
	"encoding/json"
	"gobankapi/internal/middleware"
	"gobankapi/internal/models"
	"gobankapi/internal/repositories"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type AccountHandler struct {
	AccountRepo     *repositories.AccountRepository
	TransactionRepo *repositories.TransactionRepository
	ScheduleRepo    *repositories.PaymentScheduleRepository
}

func NewAccountHandler(
	accRepo *repositories.AccountRepository,
	txRepo *repositories.TransactionRepository,
	schedRepo *repositories.PaymentScheduleRepository,
) *AccountHandler {
	return &AccountHandler{
		AccountRepo:     accRepo,
		TransactionRepo: txRepo,
		ScheduleRepo:    schedRepo,
	}
}

// POST /accounts
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := strconv.Atoi(userIDStr)

	account := &models.Account{
		UserID:  userID,
		Number:  generateAccountNumber(userID),
		Balance: 0,
	}

	err := h.AccountRepo.Create(account)
	if err != nil {
		http.Error(w, "Could not create account", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

// GET /accounts
func (h *AccountHandler) GetUserAccounts(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := strconv.Atoi(userIDStr)

	accounts, err := h.AccountRepo.FindByUserID(userID)
	if err != nil {
		http.Error(w, "Could not fetch accounts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

// Генерация простого номера счёта (можно заменить на UUID, IBAN и т.д.)
func generateAccountNumber(userID int) string {
	return "ACC" + strconv.Itoa(userID) + time.Now().Format("20060102150405")
}

type BalanceRequest struct {
	AccountID int     `json:"account_id"`
	Amount    float64 `json:"amount"`
}

// POST /accounts/deposit
func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := strconv.Atoi(userIDStr)

	var req BalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.AccountRepo.UpdateBalance(req.AccountID, userID, req.Amount)
	if err != nil {
		http.Error(w, "Deposit failed", http.StatusInternalServerError)
		return
	}

	_ = h.TransactionRepo.Log(&models.Transaction{
		ToAccountID: &req.AccountID,
		Amount:      req.Amount,
		Type:        "deposit",
	})

	w.Write([]byte(`{"status":"ok","action":"deposit"}`))
}

// POST /accounts/withdraw
func (h *AccountHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := strconv.Atoi(userIDStr)

	var req BalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Отрицательное значение для списания
	err := h.AccountRepo.UpdateBalance(req.AccountID, userID, -req.Amount)
	if err != nil {
		http.Error(w, "Withdraw failed", http.StatusInternalServerError)
		return
	}

	_ = h.TransactionRepo.Log(&models.Transaction{
		FromAccountID: &req.AccountID,
		Amount:        -req.Amount,
		Type:          "withdraw",
	})

	w.Write([]byte(`{"status":"ok","action":"withdraw"}`))
}

type TransferRequest struct {
	FromAccountID int     `json:"from_account_id"`
	ToAccountID   int     `json:"to_account_id"`
	Amount        float64 `json:"amount"`
}

func (h *AccountHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := strconv.Atoi(userIDStr)

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.AccountRepo.Transfer(req.FromAccountID, req.ToAccountID, userID, req.Amount)
	if err != nil {
		http.Error(w, "Transfer failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// списание со счёта отправителя
	_ = h.TransactionRepo.Log(&models.Transaction{
		FromAccountID: &req.FromAccountID,
		Amount:        -req.Amount,
		Type:          "transfer",
	})

	// пополнение счёта получателя
	_ = h.TransactionRepo.Log(&models.Transaction{
		ToAccountID: &req.ToAccountID,
		Amount:      req.Amount,
		Type:        "transfer",
	})

	w.Write([]byte(`{"status":"ok","action":"transfer"}`))
}

func (h *AccountHandler) PredictBalance(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := strconv.Atoi(userIDStr)

	vars := mux.Vars(r)
	accountID, err := strconv.Atoi(vars["accountId"])
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 365 {
		http.Error(w, "Invalid days (1-365)", http.StatusBadRequest)
		return
	}

	// Текущий баланс
	currentBalance, err := h.AccountRepo.GetBalance(accountID, userID)
	if err != nil {
		http.Error(w, "Could not fetch balance", http.StatusInternalServerError)
		return
	}

	// Платежи в течение N дней
	until := time.Now().AddDate(0, 0, days)
	scheduled, err := h.ScheduleRepo.GetScheduledPayments(accountID, until)
	if err != nil {
		http.Error(w, "Could not fetch payments", http.StatusInternalServerError)
		return
	}

	expectedBalance := currentBalance - scheduled

	resp := map[string]interface{}{
		"account_id":               accountID,
		"current_balance":          currentBalance,
		"expected_balance":         expectedBalance,
		"total_scheduled_payments": scheduled,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

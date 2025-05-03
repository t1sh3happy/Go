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
)

type CardHandler struct {
	CardRepo *repositories.CardRepository
}

func NewCardHandler(repo *repositories.CardRepository) *CardHandler {
	return &CardHandler{CardRepo: repo}
}

type CreateCardRequest struct {
	AccountID int `json:"account_id"`
}

// POST /cards
func (h *CardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := strconv.Atoi(userIDStr)

	var req CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AccountID == 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cardNumber := utils.GenerateCardNumber()
	expiry := utils.GenerateExpiryDate()
	cvv := utils.GenerateCVV()

	cvvHash, err := utils.HashCVV(cvv)
	if err != nil {
		http.Error(w, "CVV hash error", http.StatusInternalServerError)
		return
	}

	hmac := utils.ComputeHMAC(cardNumber, []byte("super-secret-hmac-key")) // TODO: заменить на config

	card := &models.Card{
		UserID:    userID,
		AccountID: req.AccountID,
		NumberPGP: cardNumber, // временно без шифрования
		ExpiryPGP: expiry,
		CVVHash:   cvvHash,
		HMAC:      hmac,
	}

	err = h.CardRepo.Create(card)
	if err != nil {
		http.Error(w, "Could not create card", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"card_id":   card.ID,
		"number":    cardNumber,
		"expiry":    expiry,
		"cvv":       cvv,
		"createdAt": card.CreatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *CardHandler) GetUserCards(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := strconv.Atoi(userIDStr)

	cards, err := h.CardRepo.FindByUserID(userID)
	if err != nil {
		http.Error(w, "Could not fetch cards", http.StatusInternalServerError)
		return
	}

	// Ответ без зашифрованных полей
	type CardResponse struct {
		ID        int    `json:"id"`
		AccountID int    `json:"account_id"`
		Number    string `json:"number"`
		Expiry    string `json:"expiry"`
		HMAC      string `json:"hmac"`
		CreatedAt string `json:"created_at"`
	}

	var resp []CardResponse
	for _, c := range cards {
		resp = append(resp, CardResponse{
			ID:        c.ID,
			AccountID: c.AccountID,
			Number:    c.NumberPGP, // пока без расшифровки
			Expiry:    c.ExpiryPGP, // пока без PGP
			HMAC:      c.HMAC,
			CreatedAt: c.CreatedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

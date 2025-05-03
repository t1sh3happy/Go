package router

import (
	"encoding/json"
	"gobankapi/internal/config"
	"gobankapi/internal/handlers"
	"gobankapi/internal/middleware"
	"gobankapi/internal/repositories"
	"gobankapi/internal/services"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()
	mailer := services.NewMailer()

	userRepo := repositories.NewUserRepository(config.DB)
	userHandler := handlers.NewUserHandler(userRepo)

	// --- Публичные маршруты ---
	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}).Methods("GET")

	r.HandleFunc("/register", userHandler.Register).Methods("POST")
	r.HandleFunc("/login", userHandler.Login).Methods("POST")

	r.HandleFunc("/register-form", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "register.html"))
	}).Methods("GET")

	// --- Защищённые маршруты ---
	authRouter := r.PathPrefix("/api").Subrouter()
	authRouter.Use(middleware.AuthMiddleware)

	authRouter.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"userID": userID,
		})
	}).Methods("GET")

	// --- Маршрут для логин-формы ---
	r.HandleFunc("/login-form", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "login.html"))
	}).Methods("GET")

	// --- Маршрут для формы проверки токена ---
	r.HandleFunc("/me-form", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "me.html"))
	}).Methods("GET")

	// --- Маршрут для аккаунтов, транзакций и прогноза платежей + страницы для проверки ---
	accountRepo := repositories.NewAccountRepository(config.DB)
	transactionRepo := repositories.NewTransactionRepository(config.DB)
	scheduleRepo := repositories.NewPaymentScheduleRepository(config.DB)

	accountHandler := handlers.NewAccountHandler(accountRepo, transactionRepo, scheduleRepo)

	authRouter.HandleFunc("/accounts", accountHandler.CreateAccount).Methods("POST")
	authRouter.HandleFunc("/accounts", accountHandler.GetUserAccounts).Methods("GET")

	r.HandleFunc("/accounts-form", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "accounts.html"))
	}).Methods("GET")

	// --- Маршрут для пополнения и списания + страница проверки ---
	authRouter.HandleFunc("/accounts/deposit", accountHandler.Deposit).Methods("POST")
	authRouter.HandleFunc("/accounts/withdraw", accountHandler.Withdraw).Methods("POST")

	r.HandleFunc("/accounts-balance", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "accounts-balance.html"))
	}).Methods("GET")

	// --- Маршрут для перевода между счетами + страница проверки ---
	authRouter.HandleFunc("/transfer", accountHandler.Transfer).Methods("POST")

	r.HandleFunc("/transfer-form", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "transfer.html"))
	}).Methods("GET")

	// --- Маршрут для создания карт + страница проверки ---
	cardRepo := repositories.NewCardRepository(config.DB)
	cardHandler := handlers.NewCardHandler(cardRepo)

	authRouter.HandleFunc("/cards", cardHandler.CreateCard).Methods("POST")

	r.HandleFunc("/cards-form", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "cards.html"))
	}).Methods("GET")

	// --- Маршрут для получения списка карт + страница проверки ---
	authRouter.HandleFunc("/cards", cardHandler.GetUserCards).Methods("GET")

	r.HandleFunc("/cards-view", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "cards-view.html"))
	}).Methods("GET")

	// --- Блок и маршрут по кредитам + страница проверки ---
	creditRepo := repositories.NewCreditRepository(config.DB)
	creditHandler := handlers.NewCreditHandler(creditRepo, scheduleRepo)

	authRouter.HandleFunc("/credits", creditHandler.CreateCredit).Methods("POST")

	r.HandleFunc("/credits-form", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "credits.html"))
	}).Methods("GET")

	// --- Маршрут для графика платежей + страница проверки ---
	authRouter.HandleFunc("/credits/{creditId}/schedule", creditHandler.GetSchedule).Methods("GET")

	r.HandleFunc("/schedule-form", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "schedule-view.html"))
	}).Methods("GET")

	// --- Маршрут для аналитики по месяцам и кредитам + страницы проверки ---
	analyticsHandler := handlers.NewAnalyticsHandler(transactionRepo, creditRepo)
	authRouter.HandleFunc("/analytics/credit-load", analyticsHandler.GetCreditLoad).Methods("GET")

	r.HandleFunc("/analytics-monthly", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "analytics-monthly.html"))
	}).Methods("GET")

	r.HandleFunc("/analytics-credit", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "analytics-credit.html"))
	}).Methods("GET")

	// --- Маршрут для прогноза баланса ---
	authRouter.HandleFunc("/accounts/{accountId}/predict", accountHandler.PredictBalance).Methods("GET")

	r.HandleFunc("/predict-form", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "predict-balance.html"))
	}).Methods("GET")

	// --- Проверка SMTP ---
	authRouter.HandleFunc("/test-email", func(w http.ResponseWriter, r *http.Request) {
		err := mailer.SendPaymentConfirmation("your@email.com", 149.90)
		if err != nil {
			http.Error(w, "Ошибка отправки письма: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Письмо отправлено"))
	}).Methods("GET")

	r.HandleFunc("/test-email-form", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "test-email.html"))
	}).Methods("GET")

	// --- Получение ключевой ставки ---
	authRouter.HandleFunc("/test-rate", func(w http.ResponseWriter, r *http.Request) {
		rate, err := services.GetCentralBankRate()
		if err != nil {
			http.Error(w, "Ошибка получения ставки: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]float64{
			"rate": rate,
		})
	}).Methods("GET")

	r.HandleFunc("/test-rate-form", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "test-rate.html"))
	}).Methods("GET")

	return r
}

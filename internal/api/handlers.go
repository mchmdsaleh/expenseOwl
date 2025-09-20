package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/tanq16/expenseowl/internal/auth"
	"github.com/tanq16/expenseowl/internal/storage"
	"github.com/tanq16/expenseowl/internal/web"
)

// Handler holds the storage interface
type Handler struct {
	storage storage.Storage
	auth    *auth.Manager
}

// NewHandler creates a new API handler
func NewHandler(s storage.Storage, authManager *auth.Manager) *Handler {
	return &Handler{
		storage: s,
		auth:    authManager,
	}
}

// RequireWebAuth ensures a valid session exists for browser navigations.
func (h *Handler) RequireWebAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.auth == nil {
			next(w, r)
			return
		}
		token := h.auth.ExtractToken(r)
		if token == "" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if _, err := h.auth.ValidateToken(token); err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		_ = h.auth.Renew(token)
		next(w, r)
	}
}

// RequireAPIAuth ensures API calls originate from an authenticated session.
func (h *Handler) RequireAPIAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.auth == nil {
			next(w, r)
			return
		}
		token := h.auth.ExtractToken(r)
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
			return
		}
		if _, err := h.auth.ValidateToken(token); err != nil {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "session expired"})
			return
		}
		_ = h.auth.Renew(token)
		next(w, r)
	}
}

// ErrorResponse is a generic JSON error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeJSON is a helper to write JSON responses
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

// ------------------------------------------------------------
// Config Handlers
// ------------------------------------------------------------

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	config, err := h.storage.GetConfig()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to get config"})
		log.Printf("API ERROR: Failed to get config: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, config)
}

func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	categories, err := h.storage.GetCategories()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to get categories"})
		log.Printf("API ERROR: Failed to get categories: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

func (h *Handler) UpdateCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	var categories []string
	if err := json.NewDecoder(r.Body).Decode(&categories); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	var sanitizedCategories []string
	for _, category := range categories {
		sanitized, err := storage.ValidateCategory(category)
		if err != nil {
			log.Printf("API ERROR: Invalid category provided: %v\n", err)
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Invalid category '%s': %v", category, err)})
			return
		}
		sanitizedCategories = append(sanitizedCategories, sanitized)
	}
	if err := h.storage.UpdateCategories(sanitizedCategories); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to update categories"})
		log.Printf("API ERROR: Failed to update categories: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) GetCurrency(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	currency, err := h.storage.GetCurrency()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to get currency"})
		log.Printf("API ERROR: Failed to get currency: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, currency)
}

func (h *Handler) UpdateCurrency(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	var currency string
	if err := json.NewDecoder(r.Body).Decode(&currency); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	if err := h.storage.UpdateCurrency(currency); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		log.Printf("API ERROR: Failed to update currency: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) GetStartDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	startDate, err := h.storage.GetStartDate()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to get start date"})
		log.Printf("API ERROR: Failed to get start date: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, startDate)
}

func (h *Handler) UpdateStartDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	var startDate int
	if err := json.NewDecoder(r.Body).Decode(&startDate); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	if err := h.storage.UpdateStartDate(startDate); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		log.Printf("API ERROR: Failed to update start date: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// ------------------------------------------------------------
// Expense Handlers
// ------------------------------------------------------------

func (h *Handler) AddExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	var expense storage.Expense
	if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	if err := expense.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if expense.Date.IsZero() {
		expense.Date = time.Now()
	}
	if err := h.storage.AddExpense(expense); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to save expense"})
		log.Printf("API ERROR: Failed to save expense: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, expense)
}

func (h *Handler) GetExpenses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	expenses, err := h.storage.GetAllExpenses()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve expenses"})
		log.Printf("API ERROR: Failed to retrieve expenses: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, expenses)
}

func (h *Handler) EditExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID parameter is required"})
		return
	}
	var expense storage.Expense
	if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	if err := expense.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := h.storage.UpdateExpense(id, expense); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to edit expense"})
		log.Printf("API ERROR: Failed to edit expense: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, expense)
}

func (h *Handler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID parameter is required"})
		return
	}
	if err := h.storage.RemoveExpense(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete expense"})
		log.Printf("API ERROR: Failed to delete expense: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) DeleteMultipleExpenses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	var payload struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	if err := h.storage.RemoveMultipleExpenses(payload.IDs); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete multiple expenses"})
		log.Printf("API ERROR: Failed to delete multiple expenses: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// ------------------------------------------------------------
// Recurring Expense Handlers
// ------------------------------------------------------------

func (h *Handler) AddRecurringExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	var re storage.RecurringExpense
	if err := json.NewDecoder(r.Body).Decode(&re); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	if err := re.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := h.storage.AddRecurringExpense(re); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to add recurring expense"})
		log.Printf("API ERROR: Failed to add recurring expense: %v\n", err)
		return
	}
	writeJSON(w, http.StatusCreated, re)
}

func (h *Handler) GetRecurringExpenses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	res, err := h.storage.GetRecurringExpenses()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to get recurring expenses"})
		log.Printf("API ERROR: Failed to get recurring expenses: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) UpdateRecurringExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID parameter is required"})
		return
	}
	updateAll, _ := strconv.ParseBool(r.URL.Query().Get("updateAll"))

	var re storage.RecurringExpense
	if err := json.NewDecoder(r.Body).Decode(&re); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	if err := re.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := h.storage.UpdateRecurringExpense(id, re, updateAll); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to update recurring expense"})
		log.Printf("API ERROR: Failed to update recurring expense: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) DeleteRecurringExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID parameter is required"})
		return
	}
	removeAll, _ := strconv.ParseBool(r.URL.Query().Get("removeAll"))

	if err := h.storage.RemoveRecurringExpense(id, removeAll); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete recurring expense"})
		log.Printf("API ERROR: Failed to delete recurring expense: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// ------------------------------------------------------------
// Static and UI Handlers
// ------------------------------------------------------------

func (h *Handler) ServeLoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	if h.auth != nil {
		token := h.auth.ExtractToken(r)
		if token != "" {
			if _, err := h.auth.ValidateToken(token); err == nil {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
	}
	w.Header().Set("Content-Type", "text/html")
	if err := web.ServeTemplate(w, "login.html"); err != nil {
		http.Error(w, "Failed to serve template", http.StatusInternalServerError)
	}
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid form data"})
		return
	}
	if h.auth == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "authentication not configured"})
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	token, err := h.auth.Authenticate(username, password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Invalid username or password"})
		return
	}
	h.auth.SetSessionCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]string{"redirect": "/"})
}

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	if h.auth == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "authentication not configured"})
		return
	}
	h.auth.ClearSessionCookie(w, r)
	writeJSON(w, http.StatusOK, map[string]string{"status": "signed out"})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.HandleLogin(w, r)
		return
	}
	h.ServeLoginPage(w, r)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.HandleLogout(w, r)
}

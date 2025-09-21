package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tanq16/expenseowl/internal/storage"
)

func (h *Handler) CreateExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}

	var expense storage.Expense
	if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "X-User-ID header is required"})
		return
	}
	if _, err := uuid.Parse(userID); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid user identifier"})
		return
	}

	if err := expense.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if expense.Date.IsZero() {
		expense.Date = time.Now()
	}

	expense.UserID = userID
	if err := h.storage.AddExpense(userID, expense); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to save expense"})
		return
	}

	writeJSON(w, http.StatusCreated, expense)
}

func Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-API-Key")
		if token == "" {
			// try to get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					token = parts[1]
				}
			}
		}

		if os.Getenv("API_KEY") != "" && token != os.Getenv("API_KEY") {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Invalid or missing API key"})
			return
		}
		next(w, r)
	}
}

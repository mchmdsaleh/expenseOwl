package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/tanq16/expenseowl/internal/auth"
	"github.com/tanq16/expenseowl/internal/storage"
	"github.com/tanq16/expenseowl/internal/user"
	"github.com/tanq16/expenseowl/internal/web"
)

// Handler coordinates API requests.
type Handler struct {
	storage storage.Storage
	users   *user.Service
	auth    *auth.JWTManager
}

// NewHandler creates a new API handler.
func NewHandler(s storage.Storage, userService *user.Service, authManager *auth.JWTManager) *Handler {
	return &Handler{
		storage: s,
		users:   userService,
		auth:    authManager,
	}
}

// RequireAPIAuth ensures API calls originate from an authenticated session.
func (h *Handler) RequireAPIAuth(next http.HandlerFunc) http.HandlerFunc {
	if h.auth == nil {
		return next
	}
	return h.auth.RequireWithRefresh(next)
}

// RequireAdmin enforces admin-only access on top of authenticated sessions.
func (h *Handler) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return h.RequireAPIAuth(func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := h.userFromRequest(r)
		if err != nil {
			unauthorized(w)
			return
		}
		if userCtx.Role != user.RoleAdmin {
			forbidden(w)
			return
		}
		next(w, r)
	})
}

func (h *Handler) userFromRequest(r *http.Request) (auth.UserContext, error) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == "" {
		return auth.UserContext{}, auth.ErrUnauthorized
	}
	return user, nil
}

func unauthorized(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
}

func forbidden(w http.ResponseWriter) {
	writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "admin privileges required"})
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	config, err := h.storage.GetConfig(userCtx.ID)
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	categories, err := h.storage.GetCategories(userCtx.ID)
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
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
	if err := h.storage.UpdateCategories(userCtx.ID, sanitizedCategories); err != nil {
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	currency, err := h.storage.GetCurrency(userCtx.ID)
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	var currency string
	if err := json.NewDecoder(r.Body).Decode(&currency); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	if err := h.storage.UpdateCurrency(userCtx.ID, currency); err != nil {
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	startDate, err := h.storage.GetStartDate(userCtx.ID)
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	var startDate int
	if err := json.NewDecoder(r.Body).Decode(&startDate); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	if err := h.storage.UpdateStartDate(userCtx.ID, startDate); err != nil {
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
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
	expense.UserID = userCtx.ID
	if err := h.storage.AddExpense(userCtx.ID, expense); err != nil {
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	expenses, err := h.storage.GetAllExpenses(userCtx.ID)
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
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
	if err := h.storage.UpdateExpense(userCtx.ID, id, expense); err != nil {
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID parameter is required"})
		return
	}
	if err := h.storage.RemoveExpense(userCtx.ID, id); err != nil {
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	var payload struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	if err := h.storage.RemoveMultipleExpenses(userCtx.ID, payload.IDs); err != nil {
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
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
	re.UserID = userCtx.ID
	if err := h.storage.AddRecurringExpense(userCtx.ID, re); err != nil {
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	res, err := h.storage.GetRecurringExpenses(userCtx.ID)
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
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
	re.UserID = userCtx.ID
	if err := h.storage.UpdateRecurringExpense(userCtx.ID, id, re, updateAll); err != nil {
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
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID parameter is required"})
		return
	}
	removeAll, _ := strconv.ParseBool(r.URL.Query().Get("removeAll"))

	if err := h.storage.RemoveRecurringExpense(userCtx.ID, id, removeAll); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete recurring expense"})
		log.Printf("API ERROR: Failed to delete recurring expense: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// ------------------------------------------------------------
// Authentication Handlers
// ------------------------------------------------------------

type userResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Role      string    `json:"role"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

func toUserResponse(u *user.User) userResponse {
	return userResponse{
		ID:        u.ID.String(),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Role:      u.Role,
	}
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	if h.users == nil || h.auth == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "authentication not configured"})
		return
	}
	var payload struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	ctx := r.Context()
	usr, err := h.users.Register(ctx, user.CreateParams{
		Email:     payload.Email,
		Password:  payload.Password,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
	})
	if err != nil {
		switch err {
		case user.ErrEmailTaken:
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "email already registered"})
		default:
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to create user"})
			log.Printf("API ERROR: Signup failed: %v\n", err)
		}
		return
	}
	if err := h.storage.EnsureUserDefaults(usr.ID.String()); err != nil {
		log.Printf("API ERROR: Failed to provision defaults for user %s: %v\n", usr.ID, err)
	}
	token, err := h.auth.Generate(ctx, usr.ID.String(), usr.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to create session"})
		log.Printf("API ERROR: Failed to generate token: %v\n", err)
		return
	}
	writeJSON(w, http.StatusCreated, authResponse{Token: token, User: toUserResponse(usr)})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	if h.users == nil || h.auth == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "authentication not configured"})
		return
	}
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	ctx := r.Context()
	usr, err := h.users.Authenticate(ctx, payload.Email, payload.Password)
	if err != nil {
		switch err {
		case user.ErrInvalidPassword, user.ErrInvalidArguments:
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
		default:
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to authenticate"})
			log.Printf("API ERROR: Login failed: %v\n", err)
		}
		return
	}
	if err := h.storage.EnsureUserDefaults(usr.ID.String()); err != nil {
		log.Printf("API ERROR: Failed to ensure defaults for user %s: %v\n", usr.ID, err)
	}
	token, err := h.auth.Generate(ctx, usr.ID.String(), usr.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to create session"})
		log.Printf("API ERROR: Failed to generate token: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, authResponse{Token: token, User: toUserResponse(usr)})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	token := auth.ExtractToken(r)
	if token == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "missing token"})
		return
	}
	if err := h.auth.Revoke(r.Context(), token); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to revoke session"})
		log.Printf("API ERROR: Failed to revoke token: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "signed out"})
}

func (h *Handler) Session(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	uid, parseErr := uuid.Parse(userCtx.ID)
	if parseErr != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "invalid user identifier"})
		return
	}
	usr, err := h.users.Get(r.Context(), uid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch user"})
		log.Printf("API ERROR: Session lookup failed: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(usr))
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	uid, parseErr := uuid.Parse(userCtx.ID)
	if parseErr != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "invalid user identifier"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		usr, err := h.users.Get(r.Context(), uid)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch user"})
			log.Printf("API ERROR: Profile fetch failed: %v\n", err)
			return
		}
		writeJSON(w, http.StatusOK, toUserResponse(usr))
	case http.MethodPatch:
		var payload struct {
			Email     string `json:"email"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			return
		}
		updated, err := h.users.UpdateProfile(r.Context(), uid, user.UpdateProfileParams{
			Email:     payload.Email,
			FirstName: payload.FirstName,
			LastName:  payload.LastName,
		})
		if err != nil {
			switch err {
			case user.ErrInvalidArguments:
				writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid input"})
			case user.ErrEmailTaken:
				writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "email already registered"})
			case user.ErrUserNotFound:
				writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "user not found"})
			default:
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to update profile"})
				log.Printf("API ERROR: Profile update failed: %v\n", err)
			}
			return
		}
		writeJSON(w, http.StatusOK, toUserResponse(updated))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
	}
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	uid, parseErr := uuid.Parse(userCtx.ID)
	if parseErr != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "invalid user identifier"})
		return
	}
	var payload struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if err := h.users.UpdatePassword(r.Context(), uid, payload.CurrentPassword, payload.NewPassword); err != nil {
		switch err {
		case user.ErrInvalidPassword:
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
		case user.ErrInvalidArguments:
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid inputs"})
		default:
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to update password"})
			log.Printf("API ERROR: Update password failed: %v\n", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "password updated"})
}

// ------------------------------------------------------------
// Admin Handlers
// ------------------------------------------------------------

func (h *Handler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	if h.users == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "user service not configured"})
		return
	}
	users, err := h.users.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to list users"})
		log.Printf("API ERROR: Admin list users failed: %v\n", err)
		return
	}
	responses := make([]userResponse, 0, len(users))
	for i := range users {
		u := users[i]
		responses = append(responses, toUserResponse(&u))
	}
	writeJSON(w, http.StatusOK, responses)
}

func (h *Handler) AdminUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	if h.users == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "user service not configured"})
		return
	}
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	var payload struct {
		ID   string `json:"id"`
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if payload.ID == "" || payload.Role == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "id and role are required"})
		return
	}
	uid, err := uuid.Parse(payload.ID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}
	if userCtx.ID == payload.ID && payload.Role != userCtx.Role {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "cannot change your own role"})
		return
	}
	if err := h.users.UpdateRole(r.Context(), uid, payload.Role); err != nil {
		switch err {
		case user.ErrInvalidRole:
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid role"})
		case user.ErrUserNotFound:
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "user not found"})
		default:
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to update role"})
			log.Printf("API ERROR: Admin update role failed: %v\n", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "role updated"})
}

// ServeSPA continues to deliver the frontend bundle.
func (h *Handler) ServeSPA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	web.ServeSPA(w, r)
}

package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tanq16/expenseowl/internal/integrations/telegram"
)

// TelegramLinks handles CRUD-style requests for Telegram ingest credentials.
func (h *Handler) TelegramLinks(w http.ResponseWriter, r *http.Request) {
	if h.telegram == nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "telegram integration not configured"})
		return
	}
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	userID, err := uuid.Parse(userCtx.ID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid user identifier"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		links, err := h.telegram.List(r.Context(), userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to list telegram links"})
			return
		}
		var payload []map[string]any
		for _, link := range links {
			payload = append(payload, telegram.SanitizeLink(link, false))
		}
		writeJSON(w, http.StatusOK, payload)

	case http.MethodPost:
		var body struct {
			Label string `json:"label"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
			return
		}
		link, err := h.telegram.IssueLink(r.Context(), userID, body.Label)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, telegram.SanitizeLink(*link, true))

	case http.MethodDelete:
		linkIDStr := r.URL.Query().Get("id")
		if linkIDStr == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "link id is required"})
			return
		}
		linkID, err := uuid.Parse(linkIDStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid link id"})
			return
		}
		if err := h.telegram.Revoke(r.Context(), userID, linkID); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, telegram.ErrNotFound) {
				status = http.StatusNotFound
			}
			writeJSON(w, status, ErrorResponse{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
	}
}

// TelegramCompleteLink finalises the chat ↔ user pairing using a one-time code.
func (h *Handler) TelegramCompleteLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	if h.telegram == nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "telegram integration not configured"})
		return
	}

	var body struct {
		Code     string `json:"code"`
		ChatID   int64  `json:"chatId"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	link, err := h.telegram.CompleteLink(r.Context(), body.Code, body.ChatID, body.Username)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, telegram.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, telegram.ErrAlreadyLinked):
			status = http.StatusConflict
		}
		writeJSON(w, status, ErrorResponse{Error: err.Error()})
		return
	}

	userRecord, err := h.users.Get(r.Context(), link.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to load linked user"})
		return
	}

	payload := telegram.SanitizeLink(*link, false)
	if userRecord != nil {
		payload["user"] = map[string]any{
			"id":        userRecord.ID,
			"email":     userRecord.Email,
			"firstName": userRecord.FirstName,
			"lastName":  userRecord.LastName,
			"name":      strings.TrimSpace(userRecord.FirstName + " " + userRecord.LastName),
		}
	}

	writeJSON(w, http.StatusOK, payload)
}

// TelegramResolve identifies the ExpenseOwl user linked to a Telegram chat.
func (h *Handler) TelegramResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	if h.telegram == nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "telegram integration not configured"})
		return
	}

	var body struct {
		ChatID int64 `json:"chatId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	link, err := h.telegram.ResolveChat(r.Context(), body.ChatID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, telegram.ErrNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, telegram.SanitizeLink(*link, true))
}

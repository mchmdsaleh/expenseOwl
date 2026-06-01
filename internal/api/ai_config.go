package api

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/tanq16/expenseowl/internal/storage"
)

func (h *Handler) GetAIContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	context, err := h.storage.GetAIContext(userCtx.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to get AI context"})
		log.Printf("API ERROR: Failed to get AI context: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"context": context})
}

func (h *Handler) UpdateAIContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	var payload struct {
		Context string `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	if err := h.storage.UpdateAIContext(userCtx.ID, payload.Context); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to update AI context"})
		log.Printf("API ERROR: Failed to update AI context: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) GetAIConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	config, err := h.storage.GetAIConfig(userCtx.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to get AI config"})
		log.Printf("API ERROR: Failed to get AI config: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, config)
}

func (h *Handler) UpdateAIConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	var payload storage.AIConfig
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	if err := h.storage.UpdateAIConfig(userCtx.ID, payload); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to update AI config"})
		log.Printf("API ERROR: Failed to update AI config: %v\n", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

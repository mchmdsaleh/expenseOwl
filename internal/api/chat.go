package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/tanq16/expenseowl/internal/ai"
)

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}

	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Failed to parse form"})
		return
	}

	messagesJSON := r.FormValue("messages")
	var messages []openai.ChatCompletionMessage
	if err := json.Unmarshal([]byte(messagesJSON), &messages); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid messages format"})
		return
	}

	// Handle File Attachments
	files := r.MultipartForm.File["files"]
	hasCSVAttachment := false
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			continue
		}

		fileName := strings.ToLower(fileHeader.Filename)
		if strings.HasSuffix(fileName, ".csv") {
			hasCSVAttachment = true
			// Attach CSV content as text
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Attached CSV File (%s):\n%s", fileHeader.Filename, string(content)),
			})
		} else if isImage(fileName) {
			// Attach Image as Base64 for Vision
			base64Img := base64.StdEncoding.EncodeToString(content)
			mimeType := http.DetectContentType(content)
			messages = append(messages, openai.ChatCompletionMessage{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: fmt.Sprintf("data:%s;base64,%s", mimeType, base64Img),
						},
					},
					{
						Type:    openai.ChatMessagePartTypeText,
						Text:    fmt.Sprintf("Attached Image File: %s", fileHeader.Filename),
					},
				},
			})
		}
	}

	if hasCSVAttachment {
		now := time.Now()
		messages = append(messages, openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleUser,
			Content: fmt.Sprintf(
				"Date parsing rules for this CSV: today is %s (local date %s). If a bank row embeds MMDD without a year (example `TRANSAKSI DEBIT TGL: 0526`), interpret it as month/day in the current year %d and output ISO date. Do not default to year 2023 unless the source explicitly contains 2023.",
				now.UTC().Format(time.RFC3339),
				now.Format("2006-01-02"),
				now.Year(),
			),
		})
	}

	// Fetch User AI Context and Config
	aiContext, _ := h.storage.GetAIContext(userCtx.ID)
	aiConfig, _ := h.storage.GetAIConfig(userCtx.ID)

	// Call AI Agent
	agent := ai.NewAgent(aiConfig)
	resp, err := agent.Chat(r.Context(), messages, aiContext)
	if err != nil {
		log.Printf("CHAT ERROR: AI call failed: %v\n", err)
		writeJSON(w, http.StatusBadGateway, ErrorResponse{Error: "AI failed to respond"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func isImage(fileName string) bool {
	extensions := []string{".jpg", ".jpeg", ".png", ".webp"}
	for _, ext := range extensions {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}
	return false
}

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

	// Call AI Agent in a loop to handle auto-executable tools (like data fetching)
	agent := ai.NewAgent(aiConfig)
	
	// Maximum 3 turns to prevent infinite loops
	for i := 0; i < 3; i++ {
		resp, err := agent.Chat(r.Context(), messages, aiContext)
		if err != nil {
			log.Printf("CHAT ERROR: AI call failed: %v\n", err)
			writeJSON(w, http.StatusBadGateway, ErrorResponse{Error: "AI failed to respond"})
			return
		}

		if len(resp.Choices) == 0 {
			writeJSON(w, http.StatusOK, resp)
			return
		}

		msg := resp.Choices[0].Message
		// If there are tool calls, check if we can auto-execute them
		if len(msg.ToolCalls) > 0 {
			messages = append(messages, msg)
			allAutoExecuted := true
			
			for _, tc := range msg.ToolCalls {
				if tc.Function.Name == "get_financial_report" {
					reportData, err := h.handleGetFinancialReport(r, userCtx.ID, tc.Function.Arguments)
					if err != nil {
						reportData = fmt.Sprintf("Error fetching report: %v", err)
					}
					messages = append(messages, openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleTool,
						Content:    reportData,
						ToolCallID: tc.ID,
					})
				} else {
					// This tool requires user confirmation or is not auto-executable
					allAutoExecuted = false
				}
			}

			if allAutoExecuted {
				// Continue the loop to let the AI process the tool results
				continue
			}
		}

		// If we reach here, it's either a plain text response or contains tools that need UI confirmation
		writeJSON(w, http.StatusOK, resp)
		return
	}

	writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "AI reached max auto-execution turns"})
}

func (h *Handler) handleGetFinancialReport(r *http.Request, userID string, argsStr string) (string, error) {
	var args struct {
		ReportType string `json:"report_type"`
	}
	json.Unmarshal([]byte(argsStr), &args)

	now := time.Now()
	// Fetch budget summaries for current month
	summaries, err := h.storage.GetBudgetSummaries(userID, now)
	if err != nil {
		return "", err
	}

	// Fetch all expenses to calculate actual spending
	expenses, err := h.storage.GetAllExpenses(userID)
	if err != nil {
		return "", err
	}

	manager, _ := h.encryptionManagerFromRequest(r)
	
	// Get start of current month config
	config, _ := h.storage.GetConfig(userID)
	startDate := 1
	if config != nil {
		startDate = config.StartDate
	}
	
	monthStart := time.Date(now.Year(), now.Month(), startDate, 0, 0, 0, 0, now.Location())
	if now.Day() < startDate {
		monthStart = monthStart.AddDate(0, -1, 0)
	}
	monthEnd := monthStart.AddDate(0, 1, 0)

	totals := make(map[string]float64)
	for i := range expenses {
		if manager != nil {
			decryptExpense(manager, &expenses[i])
		}
		
		if expenses[i].Amount < 0 && (expenses[i].Date.After(monthStart) || expenses[i].Date.Equal(monthStart)) && expenses[i].Date.Before(monthEnd) {
			cat := expenses[i].Category
			if cat == "" {
				cat = "Uncategorized"
			}
			totals[cat] += MathAbs(expenses[i].Amount)
		}
	}

	// Format data for AI consumption
	type AISummary struct {
		Category string  `json:"category"`
		Budget   float64 `json:"budget"`
		Spent    float64 `json:"spent"`
		Status   string  `json:"status"`
	}
	var list []AISummary
	for _, s := range summaries {
		spent := totals[s.Budget.Category]
		status := "under_budget"
		if spent > s.EffectiveAmount {
			status = "over_budget"
		} else if spent > s.EffectiveAmount*0.8 {
			status = "approaching_limit"
		}
		list = append(list, AISummary{
			Category: s.Budget.Category,
			Budget:   s.EffectiveAmount,
			Spent:    spent,
			Status:   status,
		})
	}

	raw, _ := json.Marshal(list)
	return string(raw), nil
}

func MathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
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

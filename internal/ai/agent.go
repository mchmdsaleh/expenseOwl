package ai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/tanq16/expenseowl/internal/storage"
)

const baseSystemPrompt = `You are ExpenseOwl AI, a professional financial assistant. 
Your goal is to help users manage their expenses and budgets.

You have access to tools to:
1. Retrieve financial reports (budgets vs actual spending) for insights.
2. Preview parsed data from bank mutations or receipts.
3. Add budgets.
4. Add expenses.
5. Fetch transaction history.

When asked about spending trends, current month status, or insights, ALWAYS call 'get_financial_report' first to get real data before answering.

When a user uploads a CSV or image, apply the following CUSTOM RULES provided by the user:
---
%s
---

If no custom rules are provided or they are empty, use your best judgment for professional financial categorization.

ALWAYS use the 'preview_parsed_data' tool when you have processed a file (CSV or Image) so the user can verify the data before it is saved.
`

// Agent handles chat interactions with various AI providers using Function Calling.
type Agent struct {
	config *storage.AIConfig
}

// NewAgent creates a new AI Chat Agent.
func NewAgent(config *storage.AIConfig) *Agent {
	return &Agent{
		config: config,
	}
}

// Chat handles a single chat turn, routing to the correct provider.
func (a *Agent) Chat(ctx context.Context, messages []openai.ChatCompletionMessage, userContext string) (openai.ChatCompletionResponse, error) {
	if a.config == nil || a.config.APIKey == "" {
		return openai.ChatCompletionResponse{}, fmt.Errorf("AI not configured for this user. Please set your API Key in Settings.")
	}

	systemMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf(baseSystemPrompt, userContext),
	}
	fullMessages := append([]openai.ChatCompletionMessage{systemMsg}, messages...)

	// Currently, we use the OpenAI-compatible SDK for most providers as they offer compatible endpoints.
	// Anthropic and Google (Gemini) also have OpenAI-compatible adapters or can be used via shims.
	// For simplicity and broad compatibility (Kimi, Qwen, DeepSeek, OpenAI), we use NewClientWithConfig.
	
	clientConfig := openai.DefaultConfig(a.config.APIKey)
	if a.config.BaseURL != "" {
		clientConfig.BaseURL = a.config.BaseURL
	}
	
	// Handle special cases for providers if needed (e.g. Anthropic/Google specific headers)
	// For now, we assume the user provides an OpenAI-compatible BaseURL for non-OpenAI providers.
	
	client := openai.NewClientWithConfig(clientConfig)
	
	model := a.config.Model
	if model == "" {
		model = openai.GPT4o // Fallback
	}

	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: fullMessages,
		Tools:    getTools(),
	}

	return client.CreateChatCompletion(ctx, req)
}

func getTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_financial_report",
				Description: "Retrieve current budget summaries, spending trends, and top categories. Use this to provide financial insights, warnings, or answers about current spending.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"report_type": {
							Type: jsonschema.String,
							Enum: []string{"current_month", "weekly_summary"},
						},
					},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "preview_parsed_data",
				Description: "Preview a list of expenses parsed from a CSV or Image. This MUST be called after processing a file.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"expenses": {
							Type: jsonschema.Array,
							Items: &jsonschema.Definition{
								Type: jsonschema.Object,
								Properties: map[string]jsonschema.Definition{
									"name":     {Type: jsonschema.String},
									"category": {Type: jsonschema.String},
									"amount":   {Type: jsonschema.Number},
									"date":     {Type: jsonschema.String, Description: "ISO 8601 format"},
									"currency": {Type: jsonschema.String},
								},
								Required: []string{"name", "category", "amount", "date"},
							},
						},
					},
					Required: []string{"expenses"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "add_budget",
				Description: "Request to add a new budget category.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"category": {Type: jsonschema.String},
						"amount":   {Type: jsonschema.Number},
						"period":   {Type: jsonschema.String, Enum: []string{"monthly"}},
					},
					Required: []string{"category", "amount"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "add_expense",
				Description: "Request to add a single expense.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"name":     {Type: jsonschema.String},
						"category": {Type: jsonschema.String},
						"amount":   {Type: jsonschema.Number},
						"date":     {Type: jsonschema.String, Description: "ISO 8601 format"},
					},
					Required: []string{"name", "category", "amount"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "fetch_transactions",
				Description: "Fetch transaction history to answer user queries.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"category":   {Type: jsonschema.String},
						"start_date": {Type: jsonschema.String, Description: "ISO 8601"},
						"end_date":   {Type: jsonschema.String, Description: "ISO 8601"},
					},
				},
			},
		},
	}
}

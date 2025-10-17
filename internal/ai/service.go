package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

const systemPrompt = `Kamu adalah asisten ekstraksi expense. Dari teks Telegram (Indonesia/Inggris), keluarkan hanya JSON sesuai aturan ini:

Format input

Satu atau banyak transaksi dalam satu pesan.

Transaksi dipisah dengan ; atau baris baru.

Opsi tags di mana pun di akhir pesan: t: tag1, tag2 → jika ada, masukkan ke field tags (array string). Jika tidak ada, jangan buat field tags.

Output

Jika ada >1 transaksi → keluarkan array berisi objek-objek transaksi.

Jika hanya 1 transaksi → boleh objek tunggal.

Tanpa teks lain, tanpa backticks.

Skema objek transaksi

name (string) — nama/merchant.

category (string, salah satu dari): food_drinks, transport, fuel, shopping, bills_utilities, entertainment, health_fitness, groceries, personal_care, software_subscription, misc.

amount (number) — angka bersih (positif).

currency (string, ISO-4217 huruf kecil; default idr bila tidak disebut; deteksi usd/sgd/eur/jpy, atau simbol $ € ¥).

date (string, ISO-8601 UTC diakhiri Z).

tags (array string) — hanya tampil bila ada t: di input.

Kategori (heuristik, non-limitatif)

bakso, ciomy/siomay, jco, kopi, ayam, sushi → food_drinks

gojek, grab, ojek, taksi, krl, tol/toll, parkir → transport

bensin, pertalite, pertamax, shell → fuel

pln/listrik, pdam/air, internet, telkomsel, indihome → bills_utilities

netflix, spotify, youtube premium, miitel → software_subscription (atau entertainment bila lebih tepat)

guardian, watsons, sabun, shampoo → personal_care (atau groceries sesuai konteks)

tokopedia, shopee, ace, ikea → shopping
Jika tak jelas → misc.

Aturan amount (Indonesia)

rb / ribu / k → ×1_000

jt / juta / million / m → ×1_000_000

Pisah ribuan . dan desimal , boleh; normalisasi ke angka murni.

Tanggal

Jika tak disebut: pakai tanggal dari variabel yang diberikan di prompt (lihat bagian User).

Convert ke ISO-8601 UTC (akhiri Z).

Jangan membuat tanggal masa depan acak.

Patuhi

Kembalikan hanya JSON (objek atau array).

Jangan isi tags bila tidak ada t:.

Jangan menambah field lain.`

const userPromptTemplate = `Gunakan teks berikut untuk diekstrak. Jika ada beberapa transaksi dipisah ;, keluarkan array objek (satu objek per transaksi).
Jika ada t: ... di akhir pesan, buat tags (array). Jika tidak ada t:, jangan sertakan tags sama sekali.

Teks: "%s"

Tanggal default untuk date (pakai waktu kirim Telegram, sudah dalam UTC detik epoch):
DefaultDateUTC: "%s"

jika ada lebih dari satu transaksi, WAJIB keluarkan array JSON.

Kembalikan hanya JSON sesuai skema.`

// Expense represents a parsed expense structure returned by the LLM.
type Expense struct {
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Amount   float64  `json:"amount"`
	Currency string   `json:"currency"`
	Date     string   `json:"date"`
	Tags     []string `json:"tags,omitempty"`
}

// Parser handles calls to the OpenAI API for expense extraction.
type Parser struct {
	client      *openai.Client
	model       string
	timeout     time.Duration
	maxRetries  int
	enabled     bool
	rateLimiter <-chan time.Time
}

// NewExpenseParser creates a new expense parser. If apiKey is empty, the parser is disabled (nil is returned).
func NewExpenseParser(apiKey string) (*Parser, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, nil
	}
	client := openai.NewClient(apiKey)
	return &Parser{
		client:     client,
		model:      openai.GPT4oMini,
		timeout:    45 * time.Second,
		maxRetries: 2,
		enabled:    true,
	}, nil
}

// IsEnabled reports whether the parser is configured with an API key.
func (p *Parser) IsEnabled() bool {
	return p != nil && p.enabled && p.client != nil
}

// Parse attempts to extract expenses from the provided free-form text.
func (p *Parser) Parse(ctx context.Context, text, defaultDate string) ([]Expense, error) {
	if !p.IsEnabled() {
		return nil, errors.New("expense parser not configured")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("text is required")
	}
	defaultDate = strings.TrimSpace(defaultDate)
	if defaultDate == "" {
		defaultDate = time.Now().UTC().Format(time.RFC3339)
	}

	userPrompt := fmt.Sprintf(userPromptTemplate, escapeQuotes(text), escapeQuotes(defaultDate))
	req := openai.ChatCompletionRequest{
		Model:       p.model,
		Temperature: 0,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		},
	}

	var lastErr error
	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		requestCtx, cancel := context.WithTimeout(ctx, p.timeout)
		resp, err := p.client.CreateChatCompletion(requestCtx, req)
		cancel()
		if err != nil {
			lastErr = err
			continue
		}
		if len(resp.Choices) == 0 {
			lastErr = errors.New("no choices returned from OpenAI")
			continue
		}
		content := strings.TrimSpace(resp.Choices[0].Message.Content)
		expenses, err := parseExpensesJSON(content)
		if err != nil {
			lastErr = err
			continue
		}
		return expenses, nil
	}
	if lastErr == nil {
		lastErr = errors.New("failed to parse expenses")
	}
	return nil, lastErr
}

func escapeQuotes(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}

func parseExpensesJSON(raw string) ([]Expense, error) {
	raw = sanitizeModelOutput(raw)
	if raw == "" {
		return nil, errors.New("model returned empty response")
	}

	if expenses, err := decodeExpenses(raw); err == nil {
		return expenses, nil
	}

	if segment := extractJSONSegment(raw); segment != "" && segment != raw {
		return parseExpensesJSON(segment)
	}

	return nil, fmt.Errorf("failed to parse JSON response: %s", raw)
}

func sanitizeModelOutput(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "`")
	if strings.HasPrefix(strings.ToLower(raw), "json") {
		raw = strings.TrimSpace(raw[4:])
	}
	return raw
}

func decodeExpenses(raw string) ([]Expense, error) {
	var (
		arr     []Expense
		single  Expense
		wrapped struct {
			Expenses []Expense `json:"expenses"`
			Items    []Expense `json:"items"`
		}
	)

	if strings.HasPrefix(raw, "[") {
		if err := json.Unmarshal([]byte(raw), &arr); err == nil {
			return arr, nil
		}
	}

	if err := json.Unmarshal([]byte(raw), &single); err == nil {
		return []Expense{single}, nil
	}

	if err := json.Unmarshal([]byte(raw), &wrapped); err == nil {
		switch {
		case len(wrapped.Expenses) > 0:
			return wrapped.Expenses, nil
		case len(wrapped.Items) > 0:
			return wrapped.Items, nil
		}
	}

	joined := fmt.Sprintf("[%s]", strings.ReplaceAll(raw, "}{", "},{"))
	if err := json.Unmarshal([]byte(joined), &arr); err == nil && len(arr) > 0 {
		return arr, nil
	}

	return nil, errors.New("unable to decode expenses payload")
}

func extractJSONSegment(raw string) string {
	find := func(open, close rune) string {
		start := strings.IndexRune(raw, open)
		if start == -1 {
			return ""
		}
		level := 0
		for i := start; i < len(raw); i++ {
			switch raw[i] {
			case byte(open):
				level++
			case byte(close):
				level--
				if level == 0 {
					return raw[start : i+1]
				}
			}
		}
		return ""
	}

	if segment := find('[', ']'); segment != "" {
		return segment
	}
	return find('{', '}')
}

package telegram

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service coordinates Telegram link persistence.
type Service struct {
	db *sql.DB
}

// Link represents a Telegram chat connection for expense ingestion.
type Link struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	ChatID         sql.NullInt64
	Label          string
	LinkCode       sql.NullString
	IngestToken    string
	TelegramHandle sql.NullString
	CreatedAt      time.Time
	LinkedAt       sql.NullTime
	RevokedAt      sql.NullTime
	LastSeenAt     sql.NullTime
}

var (
	// ErrNotFound is returned when no Telegram link matches the lookup criteria.
	ErrNotFound = errors.New("telegram link not found")
	// ErrAlreadyLinked indicates the chat has already been paired with a user.
	ErrAlreadyLinked = errors.New("telegram chat already linked")
	// ErrTokenInvalid indicates the provided ingest token is not recognised.
	ErrTokenInvalid = errors.New("invalid ingest token")
)

// NewService wires a Service with the given database connection.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// IssueLink creates a new pending Telegram link for the given user and returns the
// link alongside the generated link code and ingest token (plain value).
func (s *Service) IssueLink(ctx context.Context, userID uuid.UUID, label string) (*Link, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("userID is required")
	}
	label = strings.TrimSpace(label)
	if label == "" {
		label = "Default"
	}
	if len(label) > 100 {
		label = label[:100]
	}

	id := uuid.New()
	code, err := randomCode(8)
	if err != nil {
		return nil, err
	}
	token, err := randomToken(24)
	if err != nil {
		return nil, err
	}

	var link Link
	query := `
        INSERT INTO telegram_links (id, user_id, label, link_code, ingest_token)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, user_id, chat_id, label, link_code, ingest_token, telegram_username, created_at, linked_at, revoked_at, last_seen_at
    `
	err = s.db.QueryRowContext(ctx, query, id, userID, label, code, token).Scan(
		&link.ID,
		&link.UserID,
		&link.ChatID,
		&link.Label,
		&link.LinkCode,
		&link.IngestToken,
		&link.TelegramHandle,
		&link.CreatedAt,
		&link.LinkedAt,
		&link.RevokedAt,
		&link.LastSeenAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, fmt.Errorf("label already in use for this user")
		}
		return nil, fmt.Errorf("failed to create telegram link: %w", err)
	}
	// Ensure the plain ingest token value is preserved for the caller.
	link.IngestToken = token
	// The link code is present in link.LinkCode (sql.NullString) with the plain code.
	link.LinkCode.String = code
	link.LinkCode.Valid = true
	return &link, nil
}

// CompleteLink assigns a Telegram chat to a pending link using the one-time code.
func (s *Service) CompleteLink(ctx context.Context, code string, chatID int64, username string) (*Link, error) {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == "" {
		return nil, fmt.Errorf("link code is required")
	}
	username = strings.TrimSpace(username)

	query := `
        UPDATE telegram_links
        SET chat_id = $1,
            telegram_username = NULLIF($2, ''),
            linked_at = NOW(),
            link_code = NULL,
            last_seen_at = NOW()
        WHERE link_code ILIKE $3
          AND revoked_at IS NULL
          AND chat_id IS NULL
        RETURNING id, user_id, chat_id, label, link_code, ingest_token, telegram_username, created_at, linked_at, revoked_at, last_seen_at
    `
	var link Link
	err := s.db.QueryRowContext(ctx, query, chatID, username, code).Scan(
		&link.ID,
		&link.UserID,
		&link.ChatID,
		&link.Label,
		&link.LinkCode,
		&link.IngestToken,
		&link.TelegramHandle,
		&link.CreatedAt,
		&link.LinkedAt,
		&link.RevokedAt,
		&link.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, ErrAlreadyLinked
		}
		return nil, fmt.Errorf("failed to complete telegram link: %w", err)
	}
	return &link, nil
}

// ResolveChat returns the active link associated with the chat id.
func (s *Service) ResolveChat(ctx context.Context, chatID int64) (*Link, error) {
	query := `
        UPDATE telegram_links
        SET last_seen_at = NOW()
        WHERE id = (
            SELECT id FROM telegram_links
            WHERE chat_id = $1
              AND revoked_at IS NULL
            LIMIT 1
        )
        RETURNING id, user_id, chat_id, label, link_code, ingest_token, telegram_username, created_at, linked_at, revoked_at, last_seen_at
    `
	var link Link
	err := s.db.QueryRowContext(ctx, query, chatID).Scan(
		&link.ID,
		&link.UserID,
		&link.ChatID,
		&link.Label,
		&link.LinkCode,
		&link.IngestToken,
		&link.TelegramHandle,
		&link.CreatedAt,
		&link.LinkedAt,
		&link.RevokedAt,
		&link.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to resolve telegram chat: %w", err)
	}
	return &link, nil
}

// ResolveToken looks up the link given an ingest token.
func (s *Service) ResolveToken(ctx context.Context, token string) (*Link, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrTokenInvalid
	}
	query := `
        SELECT id, user_id, chat_id, label, link_code, ingest_token, telegram_username, created_at, linked_at, revoked_at, last_seen_at
        FROM telegram_links
        WHERE ingest_token = $1
          AND revoked_at IS NULL
        LIMIT 1
    `
	var link Link
	err := s.db.QueryRowContext(ctx, query, token).Scan(
		&link.ID,
		&link.UserID,
		&link.ChatID,
		&link.Label,
		&link.LinkCode,
		&link.IngestToken,
		&link.TelegramHandle,
		&link.CreatedAt,
		&link.LinkedAt,
		&link.RevokedAt,
		&link.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenInvalid
		}
		return nil, fmt.Errorf("failed to resolve ingest token: %w", err)
	}
	return &link, nil
}

// List returns all Telegram links for the given user ordered by creation date.
func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Link, error) {
	query := `
        SELECT id, user_id, chat_id, label, link_code, ingest_token, telegram_username, created_at, linked_at, revoked_at, last_seen_at
        FROM telegram_links
        WHERE user_id = $1
        ORDER BY created_at DESC
    `
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list telegram links: %w", err)
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var link Link
		if err := rows.Scan(
			&link.ID,
			&link.UserID,
			&link.ChatID,
			&link.Label,
			&link.LinkCode,
			&link.IngestToken,
			&link.TelegramHandle,
			&link.CreatedAt,
			&link.LinkedAt,
			&link.RevokedAt,
			&link.LastSeenAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan telegram link: %w", err)
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

// Revoke marks a link as revoked so it can no longer ingest data.
func (s *Service) Revoke(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) error {
	query := `
        UPDATE telegram_links
        SET revoked_at = NOW()
        WHERE user_id = $1
          AND id = $2
          AND revoked_at IS NULL
    `
	res, err := s.db.ExecContext(ctx, query, userID, linkID)
	if err != nil {
		return fmt.Errorf("failed to revoke telegram link: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to confirm telegram link revoke: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func randomCode(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}
	for i := 0; i < length; i++ {
		bytes[i] = alphabet[int(bytes[i])%len(alphabet)]
	}
	return string(bytes), nil
}

func randomToken(byteLength int) (string, error) {
	buf := make([]byte, byteLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// SanitizeLink prepares a link for JSON responses by blanking secret fields when needed.
func SanitizeLink(link Link, includeSecrets bool) map[string]any {
	payload := map[string]any{
		"id":             link.ID,
		"userId":         link.UserID,
		"label":          link.Label,
		"chatId":         nullInt64ToPtr(link.ChatID),
		"telegramHandle": nullStringToPtr(link.TelegramHandle),
		"createdAt":      link.CreatedAt,
	}
	if link.LinkedAt.Valid {
		payload["linkedAt"] = link.LinkedAt.Time
	}
	if link.RevokedAt.Valid {
		payload["revokedAt"] = link.RevokedAt.Time
	}
	if link.LastSeenAt.Valid {
		payload["lastSeenAt"] = link.LastSeenAt.Time
	}
	if includeSecrets {
		if link.LinkCode.Valid {
			payload["linkCode"] = link.LinkCode.String
		}
		payload["ingestToken"] = link.IngestToken
	}
	return payload
}

func nullInt64ToPtr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}

func nullStringToPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	v := s.String
	return &v
}

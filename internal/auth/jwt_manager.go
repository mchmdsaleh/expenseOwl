package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
)

type contextKey string

const userIDContextKey contextKey = "expenseowl_user_id"

// JWTManager issues and validates JWT access tokens stored alongside Redis sessions.
type JWTManager struct {
	secret []byte
	expiry time.Duration
	redis  *redis.Client
}

// NewJWTManager constructs a new manager.
func NewJWTManager(secret string, expiry time.Duration, redisClient *redis.Client) (*JWTManager, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, fmt.Errorf("JWT secret cannot be empty")
	}
	if expiry <= 0 {
		expiry = 24 * time.Hour
	}
	if redisClient == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	return &JWTManager{
		secret: []byte(secret),
		expiry: expiry,
		redis:  redisClient,
	}, nil
}

// Generate issues a signed token for the provided user ID.
func (m *JWTManager) Generate(ctx context.Context, userID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("userID cannot be empty")
	}
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", err
	}
	if err := m.redis.Set(ctx, signed, userID, m.expiry).Err(); err != nil {
		return "", fmt.Errorf("failed to persist session: %w", err)
	}
	return signed, nil
}

// Validate parses and verifies the token, returning the associated user ID.
func (m *JWTManager) Validate(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", ErrUnauthorized
	}
	parsed, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return "", ErrUnauthorized
	}
	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok || !parsed.Valid {
		return "", ErrUnauthorized
	}
	userID := claims.Subject
	stored, err := m.redis.Get(ctx, token).Result()
	if err != nil || stored == "" {
		return "", ErrUnauthorized
	}
	if stored != userID {
		return "", ErrUnauthorized
	}
	return userID, nil
}

// Revoke removes a token from Redis.
func (m *JWTManager) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := m.redis.Del(ctx, token).Err(); err != nil && err != redis.Nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	return nil
}

// ExtractToken obtains a bearer token from the request.
func ExtractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// WithUserID adds the user ID to the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

// UserIDFromContext retrieves the authenticated user ID.
func UserIDFromContext(ctx context.Context) (string, bool) {
	value := ctx.Value(userIDContextKey)
	if value == nil {
		return "", false
	}
	uid, ok := value.(string)
	return uid, ok && uid != ""
}

// Require wraps an http.HandlerFunc enforcing auth and injecting the user ID.
func (m *JWTManager) Require(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := ExtractToken(r)
		userID, err := m.Validate(r.Context(), token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r.WithContext(WithUserID(r.Context(), userID)))
	}
}

// RequireWithRefresh optionally refreshes the token TTL on successful requests.
func (m *JWTManager) RequireWithRefresh(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := ExtractToken(r)
		userID, err := m.Validate(r.Context(), token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_ = m.redis.Expire(r.Context(), token, m.expiry).Err()
		next(w, r.WithContext(WithUserID(r.Context(), userID)))
	}
}

package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUnauthorized       = errors.New("unauthorized")
)

type session struct {
	username string
	expires  time.Time
}

// Manager handles authentication and session lifecycle.
type Manager struct {
	username        string
	password        string
	sessionDuration time.Duration
	cookieName      string
	cookieDomain    string
	cookieSecure    bool

	mu       sync.RWMutex
	sessions map[string]session
}

// Options allows customizing manager behaviour.
type Options struct {
	CookieName      string
	CookieDomain    string
	CookieSecure    bool
	SessionDuration time.Duration
}

// NewManager constructs a Manager with the provided credentials and options.
func NewManager(username, password string, opts Options) (*Manager, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}
	if opts.CookieName == "" {
		opts.CookieName = "expenseowl_session"
	}
	if opts.SessionDuration <= 0 {
		opts.SessionDuration = 24 * time.Hour
	}
	return &Manager{
		username:        username,
		password:        password,
		sessionDuration: opts.SessionDuration,
		cookieName:      opts.CookieName,
		cookieDomain:    opts.CookieDomain,
		cookieSecure:    opts.CookieSecure,
		sessions:        make(map[string]session),
	}, nil
}

// NewManagerFromEnv builds a Manager using environment variables.
//
// Required:
//   - APP_USERNAME
//   - APP_PASSWORD
//
// Optional:
//   - SESSION_COOKIE_NAME (default: expenseowl_session)
//   - SESSION_COOKIE_DOMAIN
//   - SESSION_COOKIE_SECURE ("true"/"1")
//   - SESSION_DURATION_HOURS (integer, default: 24)
func NewManagerFromEnv() (*Manager, error) {
	username := os.Getenv("APP_USERNAME")
	password := os.Getenv("APP_PASSWORD")

	cookieName := os.Getenv("SESSION_COOKIE_NAME")
	cookieDomain := os.Getenv("SESSION_COOKIE_DOMAIN")
	cookieSecure := parseBoolEnv(os.Getenv("SESSION_COOKIE_SECURE"))
	sessionDuration := parseDurationHours(os.Getenv("SESSION_DURATION_HOURS"))

	return NewManager(username, password, Options{
		CookieName:      cookieName,
		CookieDomain:    cookieDomain,
		CookieSecure:    cookieSecure,
		SessionDuration: sessionDuration,
	})
}

// Authenticate validates provided credentials and returns a session token.
func (m *Manager) Authenticate(username, password string) (string, error) {
	if m == nil {
		return "", ErrUnauthorized
	}
	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(username)), []byte(m.username)) != 1 {
		return "", ErrInvalidCredentials
	}
	if subtle.ConstantTimeCompare([]byte(password), []byte(m.password)) != 1 {
		return "", ErrInvalidCredentials
	}

	token, err := generateToken()
	if err != nil {
		return "", err
	}

	m.mu.Lock()
	m.sessions[token] = session{
		username: m.username,
		expires:  time.Now().Add(m.sessionDuration),
	}
	m.mu.Unlock()

	return token, nil
}

// ValidateToken checks if a session token is valid and not expired.
func (m *Manager) ValidateToken(token string) (string, error) {
	if m == nil {
		return "", ErrUnauthorized
	}
	m.mu.RLock()
	session, ok := m.sessions[token]
	m.mu.RUnlock()
	if !ok {
		return "", ErrUnauthorized
	}
	if time.Now().After(session.expires) {
		m.mu.Lock()
		delete(m.sessions, token)
		m.mu.Unlock()
		return "", ErrUnauthorized
	}
	return session.username, nil
}

// Renew extends the lifetime of a session token, if valid.
func (m *Manager) Renew(token string) error {
	if m == nil {
		return ErrUnauthorized
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[token]
	if !ok {
		return ErrUnauthorized
	}
	if time.Now().After(session.expires) {
		delete(m.sessions, token)
		return ErrUnauthorized
	}
	session.expires = time.Now().Add(m.sessionDuration)
	m.sessions[token] = session
	return nil
}

// Logout removes a session token.
func (m *Manager) Logout(token string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	delete(m.sessions, token)
	m.mu.Unlock()
}

// SetSessionCookie writes the session cookie to the response.
func (m *Manager) SetSessionCookie(w http.ResponseWriter, token string) {
	if m == nil {
		return
	}
	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(m.sessionDuration),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   m.cookieSecure,
	}
	if m.cookieDomain != "" {
		cookie.Domain = m.cookieDomain
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the authentication cookie and session.
func (m *Manager) ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
	if m == nil {
		return
	}
	token := m.ExtractToken(r)
	if token != "" {
		m.Logout(token)
	}
	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   m.cookieSecure,
	}
	if m.cookieDomain != "" {
		cookie.Domain = m.cookieDomain
	}
	http.SetCookie(w, cookie)
}

// ExtractToken retrieves the session cookie token from the request.
func (m *Manager) ExtractToken(r *http.Request) string {
	if m == nil {
		return ""
	}
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func generateToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(tokenBytes), nil
}

func parseBoolEnv(val string) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func parseDurationHours(val string) time.Duration {
	if val == "" {
		return 24 * time.Hour
	}
	hours, err := strconv.Atoi(val)
	if err != nil || hours <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(hours) * time.Hour
}

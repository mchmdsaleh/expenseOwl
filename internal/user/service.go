package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	errEmailTaken       = errors.New("email already registered")
	errInvalidPassword  = errors.New("invalid credentials")
	errUserNotFound     = errors.New("user not found")
	errInvalidArguments = errors.New("invalid arguments")
	errInvalidRole      = errors.New("invalid role")
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// User represents an ExpenseOwl account holder.
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Role         string
}

// Repository handles persistence for users.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a user repository backed by PostgreSQL.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Service provides high-level user operations.
type Service struct {
	repo *Repository
}

// NewService constructs a Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateParams contains registration inputs.
type CreateParams struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Role      string
}

func sanitizeName(value string) string {
	return strings.TrimSpace(value)
}

func normalizeRole(role string) (string, error) {
	role = strings.TrimSpace(strings.ToLower(role))
	if role == "" {
		return RoleUser, nil
	}
	switch role {
	case RoleAdmin, RoleUser:
		return role, nil
	default:
		return "", errInvalidRole
	}
}

func (s *Service) Register(ctx context.Context, params CreateParams) (*User, error) {
	email := strings.TrimSpace(strings.ToLower(params.Email))
	password := strings.TrimSpace(params.Password)
	if email == "" || password == "" {
		return nil, errInvalidArguments
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	role, err := normalizeRole(params.Role)
	if err != nil {
		return nil, err
	}
	user := &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		FirstName:    sanitizeName(params.FirstName),
		LastName:     sanitizeName(params.LastName),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Role:         role,
	}
	if user.FirstName == "" {
		user.FirstName = "Expense"
	}
	if user.LastName == "" {
		user.LastName = "Owl"
	}
	if err := s.repo.create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Authenticate validates credentials and returns the user.
func (s *Service) Authenticate(ctx context.Context, email, password string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, errInvalidArguments
	}
	user, err := s.repo.findByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, errInvalidPassword
	}
	return user, nil
}

// UpdatePassword replaces the stored credentials after verifying the old password.
func (s *Service) UpdatePassword(ctx context.Context, userID uuid.UUID, current, newPassword string) error {
	if newPassword == "" {
		return errInvalidArguments
	}
	user, err := s.repo.findByID(ctx, userID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(current)) != nil {
		return errInvalidPassword
	}
	if len(newPassword) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return s.repo.updatePassword(ctx, userID, string(hash))
}

// List returns all users ordered by creation time.
func (s *Service) List(ctx context.Context) ([]User, error) {
	return s.repo.list(ctx)
}

// UpdateRole updates a user's role.
func (s *Service) UpdateRole(ctx context.Context, userID uuid.UUID, role string) error {
	normalized, err := normalizeRole(role)
	if err != nil {
		return err
	}
	return s.repo.updateRole(ctx, userID, normalized)
}

func (r *Repository) create(ctx context.Context, user *User) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO users (id, email, password_hash, first_name, last_name, created_at, updated_at, role)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.CreatedAt, user.UpdatedAt, user.Role)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return errEmailTaken
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *Repository) findByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, `
        SELECT id, email, password_hash, first_name, last_name, created_at, updated_at, role
        FROM users
        WHERE email = $1
    `, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errInvalidPassword
		}
		return nil, fmt.Errorf("failed to lookup user: %w", err)
	}
	return &user, nil
}

func (r *Repository) findByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, `
        SELECT id, email, password_hash, first_name, last_name, created_at, updated_at, role
        FROM users
        WHERE id = $1
    `, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errUserNotFound
		}
		return nil, fmt.Errorf("failed to lookup user: %w", err)
	}
	return &user, nil
}

func (r *Repository) updatePassword(ctx context.Context, id uuid.UUID, newHash string) error {
	res, err := r.db.ExecContext(ctx, `
        UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3
    `, newHash, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read update result: %w", err)
	}
	if affected == 0 {
		return errUserNotFound
	}
	return nil
}

func (r *Repository) list(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, email, first_name, last_name, created_at, updated_at, role
        FROM users
        ORDER BY created_at
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt, &user.Role); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *Repository) updateRole(ctx context.Context, id uuid.UUID, role string) error {
	res, err := r.db.ExecContext(ctx, `
        UPDATE users SET role = $1, updated_at = $2 WHERE id = $3
    `, role, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read update result: %w", err)
	}
	if rows == 0 {
		return errUserNotFound
	}
	return nil
}

// Get retrieves a user by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.findByID(ctx, id)
}

// Repo exposes underlying repository for advanced operations.
func (s *Service) Repo() *Repository {
	return s.repo
}

// Errors exported for API checks.
var (
	ErrEmailTaken       = errEmailTaken
	ErrInvalidPassword  = errInvalidPassword
	ErrUserNotFound     = errUserNotFound
	ErrInvalidArguments = errInvalidArguments
	ErrInvalidRole      = errInvalidRole
)

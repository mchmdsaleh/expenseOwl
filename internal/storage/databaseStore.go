package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/tanq16/expenseowl/internal/encryption"
)

// databaseStore implements the Storage interface for PostgreSQL.
type databaseStore struct {
	db *sql.DB
}

// SQL queries as constants for reusability and clarity.
const (
	createUsersTableSQL = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(320) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    role VARCHAR(20) NOT NULL DEFAULT 'user'
);
`

	ensureUserRoleColumnSQL = `
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';
`

	createUserSettingsTableSQL = `
CREATE TABLE IF NOT EXISTS user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    categories TEXT NOT NULL,
    currency VARCHAR(255) NOT NULL,
    start_date INTEGER NOT NULL,
    end_of_month BOOLEAN NOT NULL DEFAULT FALSE
);
`

	ensureUserSettingsEndOfMonthColumnSQL = `
ALTER TABLE user_settings
    ADD COLUMN IF NOT EXISTS end_of_month BOOLEAN NOT NULL DEFAULT FALSE;
`

	createExpensesTableSQL = `
CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recurring_id UUID,
    blob TEXT NOT NULL
);
`

	createBudgetsTableSQL = `
CREATE TABLE IF NOT EXISTS budgets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(255) NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    period VARCHAR(20) NOT NULL DEFAULT 'monthly',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, category, period)
);
`

	createBudgetOverridesTableSQL = `
CREATE TABLE IF NOT EXISTS budget_overrides (
    id UUID PRIMARY KEY,
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (budget_id, period_start)
);
`

	createBudgetAdjustmentsTableSQL = `
CREATE TABLE IF NOT EXISTS budget_adjustments (
    id UUID PRIMARY KEY,
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (budget_id, period_start)
);
`

	ensureExpensesBlobColumnSQL = `
ALTER TABLE expenses
    ADD COLUMN IF NOT EXISTS blob TEXT;
`

	createRecurringExpensesTableSQL = `
CREATE TABLE IF NOT EXISTS recurring_expenses (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    amount NUMERIC(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    category VARCHAR(255) NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    interval VARCHAR(50) NOT NULL,
    occurrences INTEGER NOT NULL,
    tags TEXT,
    blob TEXT
);
`

	ensureRecurringBlobColumnSQL = `
ALTER TABLE recurring_expenses
    ADD COLUMN IF NOT EXISTS blob TEXT;
`

	createTelegramLinksTableSQL = `
CREATE TABLE IF NOT EXISTS telegram_links (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chat_id BIGINT,
    label VARCHAR(100) NOT NULL,
    link_code VARCHAR(64),
    ingest_token VARCHAR(128) NOT NULL,
    telegram_username VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    linked_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    last_seen_at TIMESTAMPTZ
);
`

	createTelegramLinksLabelIndexSQL = `
CREATE UNIQUE INDEX IF NOT EXISTS idx_telegram_links_user_label_active
    ON telegram_links (user_id, lower(label))
    WHERE revoked_at IS NULL;
`

	createTelegramLinksChatIndexSQL = `
CREATE UNIQUE INDEX IF NOT EXISTS idx_telegram_links_chat_id_active
    ON telegram_links (chat_id)
    WHERE chat_id IS NOT NULL AND revoked_at IS NULL;
`
)

func InitializePostgresStore(baseConfig SystemConfig) (Storage, error) {
	dbURL := makeDBURL(baseConfig)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %v", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %v", err)
	}
	log.Println("Connected to PostgreSQL database")

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create database tables: %v", err)
	}
	return &databaseStore{db: db}, nil
}

func makeDBURL(baseConfig SystemConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s?sslmode=%s", baseConfig.StorageUser, baseConfig.StoragePass, baseConfig.StorageURL, baseConfig.StorageSSL)
}

func createTables(db *sql.DB) error {
	queries := []string{
		createUsersTableSQL,
		ensureUserRoleColumnSQL,
		createUserSettingsTableSQL,
		ensureUserSettingsEndOfMonthColumnSQL,
		createExpensesTableSQL,
		createBudgetsTableSQL,
		createBudgetOverridesTableSQL,
		createBudgetAdjustmentsTableSQL,
		ensureExpensesBlobColumnSQL,
		createRecurringExpensesTableSQL,
		ensureRecurringBlobColumnSQL,
		createTelegramLinksTableSQL,
		createTelegramLinksLabelIndexSQL,
		createTelegramLinksChatIndexSQL,
	}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	if err := ensureExpenseBlobSchema(db); err != nil {
		return err
	}
	return ensureRecurringUserColumn(db)
}

func ensureExpenseBlobSchema(db *sql.DB) error {
	if _, err := db.Exec(`
        ALTER TABLE expenses
            DROP COLUMN IF EXISTS name,
            DROP COLUMN IF EXISTS category,
            DROP COLUMN IF EXISTS amount,
            DROP COLUMN IF EXISTS currency,
            DROP COLUMN IF EXISTS date,
            DROP COLUMN IF EXISTS tags
    `); err != nil {
		return err
	}
	if _, err := db.Exec(`ALTER TABLE expenses ALTER COLUMN blob SET NOT NULL`); err != nil {
		return err
	}
	return nil
}

func ensureRecurringUserColumn(db *sql.DB) error {
	var tableExists bool
	if err := db.QueryRow(`
        SELECT EXISTS (
            SELECT 1
            FROM information_schema.tables
            WHERE table_name = 'recurring_expenses'
        )
    `).Scan(&tableExists); err != nil {
		return err
	}
	if !tableExists {
		return nil
	}

	var columnExists bool
	if err := db.QueryRow(`
        SELECT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_name = 'recurring_expenses' AND column_name = 'user_id'
        )
    `).Scan(&columnExists); err != nil {
		return err
	}
	if !columnExists {
		if _, err := db.Exec(`ALTER TABLE recurring_expenses ADD COLUMN user_id UUID`); err != nil {
			return err
		}
	}

	if _, err := db.Exec(`
        UPDATE recurring_expenses re
        SET user_id = src.user_id
        FROM (
            SELECT DISTINCT recurring_id, user_id
            FROM expenses
            WHERE recurring_id IS NOT NULL AND user_id IS NOT NULL
        ) src
        WHERE re.user_id IS NULL AND re.id = src.recurring_id
    `); err != nil {
		return err
	}

	var userCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&userCount); err != nil {
		return err
	}
	if userCount == 1 {
		if _, err := db.Exec(`
            UPDATE recurring_expenses re
            SET user_id = u.id
            FROM (
                SELECT id FROM users LIMIT 1
            ) u
            WHERE re.user_id IS NULL
        `); err != nil {
			return err
		}
	}

	return nil
}

func (s *databaseStore) Close() error {
	return s.db.Close()
}

func (s *databaseStore) EnsureUserDefaults(userID string) error {
	if userID == "" {
		return errors.New("userID is required")
	}
	defaults := Config{}
	defaults.SetBaseConfig()
	categoriesJSON, err := json.Marshal(defaults.Categories)
	if err != nil {
		return fmt.Errorf("failed to marshal default categories: %v", err)
	}
	_, err = s.db.Exec(`
        INSERT INTO user_settings (user_id, categories, currency, start_date, end_of_month)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (user_id) DO NOTHING
    `, userID, string(categoriesJSON), defaults.Currency, defaults.StartDate, defaults.EndOfMonth)
	return err
}

func (s *databaseStore) GetConfig(userID string) (*Config, error) {
	if err := s.EnsureUserDefaults(userID); err != nil {
		return nil, err
	}
	var categoriesStr, currency string
	var startDate int
	var endOfMonth bool
	err := s.db.QueryRow(`
        SELECT categories, currency, start_date, end_of_month
        FROM user_settings
        WHERE user_id = $1
    `, userID).Scan(&categoriesStr, &currency, &startDate, &endOfMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to load user config: %v", err)
	}
	var config Config
	config.Currency = currency
	config.StartDate = startDate
	config.EndOfMonth = endOfMonth
	if err := json.Unmarshal([]byte(categoriesStr), &config.Categories); err != nil {
		return nil, fmt.Errorf("failed to unmarshal categories: %v", err)
	}
	recurring, err := s.GetRecurringExpenses(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load recurring expenses: %v", err)
	}
	config.RecurringExpenses = recurring
	return &config, nil
}

func (s *databaseStore) GetCategories(userID string) ([]string, error) {
	cfg, err := s.GetConfig(userID)
	if err != nil {
		return nil, err
	}
	return cfg.Categories, nil
}

func (s *databaseStore) UpdateCategories(userID string, categories []string) error {
	if err := s.EnsureUserDefaults(userID); err != nil {
		return err
	}
	categoriesJSON, err := json.Marshal(categories)
	if err != nil {
		return fmt.Errorf("failed to marshal categories: %v", err)
	}
	_, err = s.db.Exec(`
        UPDATE user_settings SET categories = $1 WHERE user_id = $2
    `, string(categoriesJSON), userID)
	return err
}

func (s *databaseStore) GetCurrency(userID string) (string, error) {
	if err := s.EnsureUserDefaults(userID); err != nil {
		return "", err
	}
	var currency string
	err := s.db.QueryRow(`SELECT currency FROM user_settings WHERE user_id = $1`, userID).Scan(&currency)
	if err != nil {
		return "", fmt.Errorf("failed to load currency: %v", err)
	}
	return currency, nil
}

func (s *databaseStore) UpdateCurrency(userID string, currency string) error {
	if !slices.Contains(SupportedCurrencies, currency) {
		return fmt.Errorf("invalid currency: %s", currency)
	}
	if err := s.EnsureUserDefaults(userID); err != nil {
		return err
	}
	_, err := s.db.Exec(`UPDATE user_settings SET currency = $1 WHERE user_id = $2`, currency, userID)
	return err
}

func (s *databaseStore) GetStartDate(userID string) (int, error) {
	if err := s.EnsureUserDefaults(userID); err != nil {
		return 0, err
	}
	var startDate int
	err := s.db.QueryRow(`SELECT start_date FROM user_settings WHERE user_id = $1`, userID).Scan(&startDate)
	if err != nil {
		return 0, fmt.Errorf("failed to load start date: %v", err)
	}
	return startDate, nil
}

func (s *databaseStore) UpdateStartDate(userID string, startDate int) error {
	if startDate < 1 || startDate > 31 {
		return fmt.Errorf("invalid start date: %d", startDate)
	}
	if err := s.EnsureUserDefaults(userID); err != nil {
		return err
	}
	_, err := s.db.Exec(`UPDATE user_settings SET start_date = $1 WHERE user_id = $2`, startDate, userID)
	return err
}

func (s *databaseStore) GetEndOfMonth(userID string) (bool, error) {
	if err := s.EnsureUserDefaults(userID); err != nil {
		return false, err
	}
	var endOfMonth bool
	err := s.db.QueryRow(`SELECT end_of_month FROM user_settings WHERE user_id = $1`, userID).Scan(&endOfMonth)
	if err != nil {
		return false, fmt.Errorf("failed to load end-of-month preference: %v", err)
	}
	return endOfMonth, nil
}

func (s *databaseStore) UpdateEndOfMonth(userID string, endOfMonth bool) error {
	if err := s.EnsureUserDefaults(userID); err != nil {
		return err
	}
	_, err := s.db.Exec(`UPDATE user_settings SET end_of_month = $1 WHERE user_id = $2`, endOfMonth, userID)
	return err
}

func (s *databaseStore) GetBudgets(userID string) ([]Budget, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, category, amount::float8, currency, period, created_at, updated_at
		FROM budgets
		WHERE user_id = $1
		ORDER BY category
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query budgets: %v", err)
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		budget, err := scanBudget(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget: %v", err)
		}
		budgets = append(budgets, budget)
	}
	return budgets, nil
}

func (s *databaseStore) AddBudget(userID string, budget Budget) (Budget, error) {
	if userID == "" {
		return Budget{}, errors.New("userID is required")
	}
	if err := s.EnsureUserDefaults(userID); err != nil {
		return Budget{}, err
	}
	if budget.Category == "" {
		return Budget{}, fmt.Errorf("category is required")
	}
	sanitized, err := ValidateCategory(budget.Category)
	if err != nil {
		return Budget{}, err
	}
	budget.Category = sanitized
	if budget.Amount <= 0 {
		return Budget{}, fmt.Errorf("budget amount must be greater than 0")
	}
	budget.Amount = math.Round(budget.Amount*100) / 100
	if budget.Currency == "" {
		currency, err := s.GetCurrency(userID)
		if err != nil {
			return Budget{}, err
		}
		budget.Currency = currency
	}
	period := budget.Period
	if period == "" {
		period = BudgetPeriodMonthly
	}
	period = strings.ToLower(period)
	if period != BudgetPeriodMonthly {
		return Budget{}, fmt.Errorf("unsupported budget period: %s", budget.Period)
	}
	budget.Period = period
	if budget.ID == "" {
		budget.ID = uuid.New().String()
	}
	budget.UserID = userID
	err = s.db.QueryRow(`
		INSERT INTO budgets (id, user_id, category, amount, currency, period)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`, budget.ID, userID, budget.Category, budget.Amount, budget.Currency, budget.Period).Scan(&budget.CreatedAt, &budget.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				return Budget{}, fmt.Errorf("budget for category '%s' already exists", budget.Category)
			}
		}
		return Budget{}, fmt.Errorf("failed to insert budget: %v", err)
	}
	return budget, nil
}

func (s *databaseStore) UpdateBudget(userID, id string, budget Budget) (Budget, error) {
	if id == "" {
		return Budget{}, fmt.Errorf("budget ID is required")
	}
	if userID == "" {
		return Budget{}, errors.New("userID is required")
	}
	if err := s.EnsureUserDefaults(userID); err != nil {
		return Budget{}, err
	}
	if budget.Category == "" {
		return Budget{}, fmt.Errorf("category is required")
	}
	sanitized, err := ValidateCategory(budget.Category)
	if err != nil {
		return Budget{}, err
	}
	budget.Category = sanitized
	if budget.Amount <= 0 {
		return Budget{}, fmt.Errorf("budget amount must be greater than 0")
	}
	budget.Amount = math.Round(budget.Amount*100) / 100
	if budget.Currency == "" {
		currency, err := s.GetCurrency(userID)
		if err != nil {
			return Budget{}, err
		}
		budget.Currency = currency
	}
	period := budget.Period
	if period == "" {
		period = BudgetPeriodMonthly
	}
	period = strings.ToLower(period)
	if period != BudgetPeriodMonthly {
		return Budget{}, fmt.Errorf("unsupported budget period: %s", budget.Period)
	}
	budget.Period = period
	budget.ID = id
	budget.UserID = userID
	err = s.db.QueryRow(`
		UPDATE budgets
		SET category = $1, amount = $2, currency = $3, period = $4, updated_at = NOW()
		WHERE id = $5 AND user_id = $6
		RETURNING created_at, updated_at
	`, budget.Category, budget.Amount, budget.Currency, budget.Period, id, userID).Scan(&budget.CreatedAt, &budget.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Budget{}, fmt.Errorf("budget not found")
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				return Budget{}, fmt.Errorf("budget for category '%s' already exists", budget.Category)
			}
		}
		return Budget{}, fmt.Errorf("failed to update budget: %v", err)
	}
	return budget, nil
}

func (s *databaseStore) RemoveBudget(userID, id string) error {
	if id == "" {
		return fmt.Errorf("budget ID is required")
	}
	res, err := s.db.Exec(`DELETE FROM budgets WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete budget: %v", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read delete result: %v", err)
	}
	if count == 0 {
		return fmt.Errorf("budget not found")
	}
	return nil
}

func (s *databaseStore) GetBudgetSummaries(userID string, month time.Time) ([]BudgetSummary, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	periodStart := normalizeMonth(month)
	rows, err := s.db.Query(`
		SELECT
			b.id,
			b.user_id,
			b.category,
			b.amount::float8,
			b.currency,
			b.period,
			b.created_at,
			b.updated_at,
			o.id,
			o.amount::float8,
			a.id,
			a.amount::float8
		FROM budgets b
		LEFT JOIN budget_overrides o ON o.budget_id = b.id AND o.period_start = $2
		LEFT JOIN budget_adjustments a ON a.budget_id = b.id AND a.period_start = $2
		WHERE b.user_id = $1
		ORDER BY b.category
	`, userID, periodStart)
	if err != nil {
		return nil, fmt.Errorf("failed to query budget summaries: %v", err)
	}
	defer rows.Close()

	var summaries []BudgetSummary
	for rows.Next() {
		var (
			budget           Budget
			overrideID       sql.NullString
			overrideAmount   sql.NullFloat64
			adjustmentID     sql.NullString
			adjustmentAmount sql.NullFloat64
		)
		if err := rows.Scan(
			&budget.ID,
			&budget.UserID,
			&budget.Category,
			&budget.Amount,
			&budget.Currency,
			&budget.Period,
			&budget.CreatedAt,
			&budget.UpdatedAt,
			&overrideID,
			&overrideAmount,
			&adjustmentID,
			&adjustmentAmount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan budget summary: %v", err)
		}
		summary := BudgetSummary{
			Budget:      budget,
			PeriodStart: periodStart,
		}
		base := budget.Amount
		if overrideID.Valid {
			summary.OverrideID = overrideID.String
			if overrideAmount.Valid {
				value := overrideAmount.Float64
				summary.OverrideAmount = &value
				base = value
			}
		}
		if adjustmentID.Valid {
			summary.AdjustmentID = adjustmentID.String
		}
		if adjustmentAmount.Valid {
			summary.AdjustmentAmount = adjustmentAmount.Float64
		}
		summary.EffectiveAmount = base + summary.AdjustmentAmount
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func (s *databaseStore) UpsertBudgetOverride(userID, budgetID string, month time.Time, amount float64) (BudgetOverride, error) {
	if amount <= 0 {
		return BudgetOverride{}, fmt.Errorf("override amount must be greater than 0")
	}
	if err := s.ensureBudgetOwnership(userID, budgetID); err != nil {
		return BudgetOverride{}, err
	}
	periodStart := normalizeMonth(month)
	var override BudgetOverride
	insertID := uuid.New().String()
	if err := s.db.QueryRow(`
		INSERT INTO budget_overrides (id, budget_id, user_id, period_start, amount)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (budget_id, period_start)
		DO UPDATE SET amount = EXCLUDED.amount, user_id = EXCLUDED.user_id, updated_at = NOW()
		RETURNING id, budget_id, user_id, period_start, amount::float8, created_at, updated_at
	`, insertID, budgetID, userID, periodStart, amount).Scan(
		&override.ID,
		&override.BudgetID,
		&override.UserID,
		&override.PeriodStart,
		&override.Amount,
		&override.CreatedAt,
		&override.UpdatedAt,
	); err != nil {
		return BudgetOverride{}, fmt.Errorf("failed to upsert budget override: %v", err)
	}
	return override, nil
}

func (s *databaseStore) DeleteBudgetOverride(userID, overrideID string) error {
	if overrideID == "" {
		return fmt.Errorf("override ID is required")
	}
	res, err := s.db.Exec(`DELETE FROM budget_overrides WHERE id = $1 AND user_id = $2`, overrideID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete budget override: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read delete result: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("override not found")
	}
	return nil
}

func (s *databaseStore) UpsertBudgetAdjustment(userID, budgetID string, month time.Time, amount float64) (BudgetAdjustment, error) {
	if amount == 0 {
		return BudgetAdjustment{}, fmt.Errorf("adjustment amount cannot be zero")
	}
	if err := s.ensureBudgetOwnership(userID, budgetID); err != nil {
		return BudgetAdjustment{}, err
	}
	periodStart := normalizeMonth(month)
	var adjustment BudgetAdjustment
	insertID := uuid.New().String()
	if err := s.db.QueryRow(`
		INSERT INTO budget_adjustments (id, budget_id, user_id, period_start, amount)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (budget_id, period_start)
		DO UPDATE SET amount = EXCLUDED.amount, user_id = EXCLUDED.user_id, updated_at = NOW()
		RETURNING id, budget_id, user_id, period_start, amount::float8, created_at, updated_at
	`, insertID, budgetID, userID, periodStart, amount).Scan(
		&adjustment.ID,
		&adjustment.BudgetID,
		&adjustment.UserID,
		&adjustment.PeriodStart,
		&adjustment.Amount,
		&adjustment.CreatedAt,
		&adjustment.UpdatedAt,
	); err != nil {
		return BudgetAdjustment{}, fmt.Errorf("failed to upsert budget adjustment: %v", err)
	}
	return adjustment, nil
}

func (s *databaseStore) DeleteBudgetAdjustment(userID, adjustmentID string) error {
	if adjustmentID == "" {
		return fmt.Errorf("adjustment ID is required")
	}
	res, err := s.db.Exec(`DELETE FROM budget_adjustments WHERE id = $1 AND user_id = $2`, adjustmentID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete budget adjustment: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read delete result: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("adjustment not found")
	}
	return nil
}

func scanBudget(scanner interface{ Scan(...any) error }) (Budget, error) {
	var budget Budget
	err := scanner.Scan(&budget.ID, &budget.UserID, &budget.Category, &budget.Amount, &budget.Currency, &budget.Period, &budget.CreatedAt, &budget.UpdatedAt)
	if err != nil {
		return Budget{}, err
	}
	return budget, nil
}

func scanExpense(scanner interface{ Scan(...any) error }) (Expense, error) {
	var expense Expense
	var recurringID sql.NullString
	var userID string
	var blob sql.NullString
	err := scanner.Scan(&expense.ID, &userID, &recurringID, &blob)
	if err != nil {
		return Expense{}, err
	}
	expense.UserID = userID
	if recurringID.Valid {
		expense.RecurringID = recurringID.String
	}
	if blob.Valid {
		expense.Blob = blob.String
	}
	return expense, nil
}

func (s *databaseStore) GetAllExpenses(userID string) ([]Expense, error) {
	rows, err := s.db.Query(`
        SELECT id, user_id, recurring_id, blob
        FROM expenses
        WHERE user_id = $1
        ORDER BY id DESC
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query expenses: %v", err)
	}
	defer rows.Close()

	var expenses []Expense
	for rows.Next() {
		expense, err := scanExpense(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense: %v", err)
		}
		expenses = append(expenses, expense)
	}
	return expenses, nil
}

func (s *databaseStore) GetExpense(userID, id string) (Expense, error) {
	expense, err := scanExpense(s.db.QueryRow(`
        SELECT id, user_id, recurring_id, blob
        FROM expenses
        WHERE user_id = $1 AND id = $2
    `, userID, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return Expense{}, fmt.Errorf("expense with ID %s not found", id)
		}
		return Expense{}, fmt.Errorf("failed to get expense: %v", err)
	}
	return expense, nil
}

func (s *databaseStore) AddExpense(userID string, expense Expense) error {
	if userID == "" {
		return errors.New("userID is required")
	}
	if expense.ID == "" {
		expense.ID = uuid.New().String()
	}
	expense.UserID = userID
	if expense.Blob == "" {
		payload := expense
		payload.Blob = ""
		raw, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to serialize expense: %v", err)
		}
		expense.Blob = string(raw)
	}
	_, err := s.db.Exec(`
        INSERT INTO expenses (id, user_id, recurring_id, blob)
        VALUES ($1, $2, $3, $4)
    `, expense.ID, userID, nullString(expense.RecurringID), expense.Blob)
	return err
}

func (s *databaseStore) UpdateExpense(userID, id string, expense Expense) error {
	if expense.Blob == "" {
		payload := expense
		payload.Blob = ""
		raw, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to serialize expense: %v", err)
		}
		expense.Blob = string(raw)
	}
	res, err := s.db.Exec(`
        UPDATE expenses
        SET blob = $1, recurring_id = $2
        WHERE id = $3 AND user_id = $4
    `, expense.Blob, nullString(expense.RecurringID), id, userID)
	if err != nil {
		return fmt.Errorf("failed to update expense: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read update result: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("expense with ID %s not found", id)
	}
	return nil
}

func (s *databaseStore) RemoveExpense(userID, id string) error {
	res, err := s.db.Exec(`DELETE FROM expenses WHERE user_id = $1 AND id = $2`, userID, id)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read delete result: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("expense with ID %s not found", id)
	}
	return nil
}

func (s *databaseStore) AddMultipleExpenses(userID string, expenses []Expense) error {
	if len(expenses) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(pq.CopyIn("expenses", "id", "user_id", "recurring_id", "blob"))
	if err != nil {
		return fmt.Errorf("failed to prepare bulk insert: %v", err)
	}
	defer stmt.Close()

	for _, exp := range expenses {
		if exp.ID == "" {
			exp.ID = uuid.New().String()
		}
		exp.UserID = userID
		if exp.Blob == "" {
			payload := exp
			payload.Blob = ""
			raw, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("failed to serialize expense: %v", err)
			}
			exp.Blob = string(raw)
		}
		if _, err := stmt.Exec(exp.ID, userID, nullString(exp.RecurringID), exp.Blob); err != nil {
			return fmt.Errorf("failed to insert expense: %v", err)
		}
	}
	if _, err := stmt.Exec(); err != nil {
		return fmt.Errorf("failed to finalize bulk insert: %v", err)
	}
	return tx.Commit()
}

func (s *databaseStore) RemoveMultipleExpenses(userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.db.Exec(`
        DELETE FROM expenses WHERE user_id = $1 AND id = ANY($2)
    `, userID, pq.Array(ids))
	return err
}

func (s *databaseStore) GetRecurringExpenses(userID string) ([]RecurringExpense, error) {
	rows, err := s.db.Query(`
        SELECT id, user_id, name, amount, currency, category, start_date, interval, occurrences, tags, blob
        FROM recurring_expenses
        WHERE user_id = $1
        ORDER BY start_date DESC
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query recurring expenses: %v", err)
	}
	defer rows.Close()

	var results []RecurringExpense
	for rows.Next() {
		var rec RecurringExpense
		var tagsStr sql.NullString
		var blob sql.NullString
		err := rows.Scan(&rec.ID, &rec.UserID, &rec.Name, &rec.Amount, &rec.Currency, &rec.Category, &rec.StartDate, &rec.Interval, &rec.Occurrences, &tagsStr, &blob)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recurring expense: %v", err)
		}
		if tagsStr.Valid && tagsStr.String != "" {
			if err := json.Unmarshal([]byte(tagsStr.String), &rec.Tags); err != nil {
				return nil, fmt.Errorf("failed to parse tags for recurring expense %s: %v", rec.ID, err)
			}
		}
		if blob.Valid {
			rec.Blob = blob.String
		}
		results = append(results, rec)
	}
	return results, nil
}

func (s *databaseStore) GetRecurringExpense(userID, id string) (RecurringExpense, error) {
	var rec RecurringExpense
	var tagsStr sql.NullString
	var blob sql.NullString
	err := s.db.QueryRow(`
        SELECT id, user_id, name, amount, currency, category, start_date, interval, occurrences, tags, blob
        FROM recurring_expenses
        WHERE user_id = $1 AND id = $2
    `, userID, id).Scan(&rec.ID, &rec.UserID, &rec.Name, &rec.Amount, &rec.Currency, &rec.Category, &rec.StartDate, &rec.Interval, &rec.Occurrences, &tagsStr, &blob)
	if err != nil {
		if err == sql.ErrNoRows {
			return RecurringExpense{}, fmt.Errorf("recurring expense with ID %s not found", id)
		}
		return RecurringExpense{}, fmt.Errorf("failed to get recurring expense: %v", err)
	}
	if tagsStr.Valid && tagsStr.String != "" {
		if err := json.Unmarshal([]byte(tagsStr.String), &rec.Tags); err != nil {
			return RecurringExpense{}, fmt.Errorf("failed to parse tags: %v", err)
		}
	}
	if blob.Valid {
		rec.Blob = blob.String
	}
	return rec, nil
}

func (s *databaseStore) AddRecurringExpense(userID string, recurringExpense RecurringExpense, enc *encryption.Manager) error {
	if userID == "" {
		return errors.New("userID is required")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	if recurringExpense.ID == "" {
		recurringExpense.ID = uuid.New().String()
	}
	recurringExpense.UserID = userID
	if recurringExpense.Currency == "" {
		currency, err := s.GetCurrency(userID)
		if err != nil {
			return err
		}
		recurringExpense.Currency = currency
	}
	tagsJSON, err := json.Marshal(recurringExpense.Tags)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
        INSERT INTO recurring_expenses (id, user_id, name, amount, currency, category, start_date, interval, occurrences, tags, blob)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `, recurringExpense.ID, userID, recurringExpense.Name, recurringExpense.Amount, recurringExpense.Currency, recurringExpense.Category, recurringExpense.StartDate, recurringExpense.Interval, recurringExpense.Occurrences, string(tagsJSON), nullString(recurringExpense.Blob))
	if err != nil {
		return fmt.Errorf("failed to insert recurring expense: %v", err)
	}

	expensesToAdd := generateExpensesFromRecurring(userID, recurringExpense, false)
	if err := bulkInsertExpenses(tx, expensesToAdd, enc); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *databaseStore) UpdateRecurringExpense(userID, id string, recurringExpense RecurringExpense, updateAll bool, enc *encryption.Manager) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	recurringExpense.ID = id
	recurringExpense.UserID = userID
	if recurringExpense.Currency == "" {
		currency, err := s.GetCurrency(userID)
		if err != nil {
			return err
		}
		recurringExpense.Currency = currency
	}
	tagsJSON, err := json.Marshal(recurringExpense.Tags)
	if err != nil {
		return err
	}
	res, err := tx.Exec(`
        UPDATE recurring_expenses
        SET name = $1, amount = $2, currency = $3, category = $4, start_date = $5, interval = $6, occurrences = $7, tags = $8, blob = $9
        WHERE id = $10 AND user_id = $11
    `, recurringExpense.Name, recurringExpense.Amount, recurringExpense.Currency, recurringExpense.Category, recurringExpense.StartDate, recurringExpense.Interval, recurringExpense.Occurrences, string(tagsJSON), nullString(recurringExpense.Blob), id, userID)
	if err != nil {
		return fmt.Errorf("failed to update recurring expense: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read update result: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("recurring expense with ID %s not found", id)
	}

	if updateAll {
		if _, err := tx.Exec(`DELETE FROM expenses WHERE user_id = $1 AND recurring_id = $2`, userID, id); err != nil {
			return fmt.Errorf("failed to delete existing expenses: %v", err)
		}
	} else {
		if _, err := tx.Exec(`DELETE FROM expenses WHERE user_id = $1 AND recurring_id = $2 AND date > $3`, userID, id, time.Now()); err != nil {
			return fmt.Errorf("failed to delete future expenses: %v", err)
		}
	}

	expensesToAdd := generateExpensesFromRecurring(userID, recurringExpense, !updateAll)
	if err := bulkInsertExpenses(tx, expensesToAdd, enc); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *databaseStore) RemoveRecurringExpense(userID, id string, removeAll bool) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(`DELETE FROM recurring_expenses WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete recurring expense: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read delete result: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("recurring expense with ID %s not found", id)
	}

	if removeAll {
		if _, err := tx.Exec(`DELETE FROM expenses WHERE user_id = $1 AND recurring_id = $2`, userID, id); err != nil {
			return fmt.Errorf("failed to delete related expenses: %v", err)
		}
	} else {
		if _, err := tx.Exec(`DELETE FROM expenses WHERE user_id = $1 AND recurring_id = $2 AND date > $3`, userID, id, time.Now()); err != nil {
			return fmt.Errorf("failed to delete future expenses: %v", err)
		}
	}
	return tx.Commit()
}

func bulkInsertExpenses(tx *sql.Tx, expenses []Expense, enc *encryption.Manager) error {
	if len(expenses) == 0 {
		return nil
	}
	stmt, err := tx.Prepare(pq.CopyIn("expenses", "id", "user_id", "recurring_id", "blob"))
	if err != nil {
		return fmt.Errorf("failed to prepare expense bulk insert: %v", err)
	}
	defer stmt.Close()

	for _, exp := range expenses {
		if exp.Blob == "" {
			// Build payload without nested blob, then encrypt or store plaintext
			payload := exp
			payload.Blob = ""
			if enc != nil {
				blob, encErr := enc.Encrypt(payload)
				if encErr != nil {
					return fmt.Errorf("failed to encrypt generated expense: %v", encErr)
				}
				exp.Blob = blob
			} else {
				raw, mErr := json.Marshal(payload)
				if mErr != nil {
					return fmt.Errorf("failed to serialize expense: %v", mErr)
				}
				exp.Blob = string(raw)
			}
		}
		if _, err := stmt.Exec(exp.ID, exp.UserID, nullString(exp.RecurringID), exp.Blob); err != nil {
			return fmt.Errorf("failed to copy expense: %v", err)
		}
	}
	if _, err := stmt.Exec(); err != nil {
		return fmt.Errorf("failed to finalize expense batch: %v", err)
	}
	return nil
}

func generateExpensesFromRecurring(userID string, recExp RecurringExpense, fromToday bool) []Expense {
	var expenses []Expense
	currentDate := recExp.StartDate
	today := time.Now()
	occurrencesToGenerate := recExp.Occurrences

	if fromToday {
		for currentDate.Before(today) && (recExp.Occurrences == 0 || occurrencesToGenerate > 0) {
			switch recExp.Interval {
			case "daily":
				currentDate = currentDate.AddDate(0, 0, 1)
			case "weekly":
				currentDate = currentDate.AddDate(0, 0, 7)
			case "monthly":
				currentDate = currentDate.AddDate(0, 1, 0)
			case "yearly":
				currentDate = currentDate.AddDate(1, 0, 0)
			default:
				return expenses
			}
			if recExp.Occurrences > 0 {
				occurrencesToGenerate--
			}
		}
	}

	count := occurrencesToGenerate
	if recExp.Occurrences == 0 {
		count = 200
	}

	for i := 0; recExp.Occurrences == 0 && i < count || (recExp.Occurrences > 0 && occurrencesToGenerate > 0); i++ {
		exp := Expense{
			ID:          uuid.New().String(),
			UserID:      userID,
			RecurringID: recExp.ID,
			Name:        recExp.Name,
			Category:    recExp.Category,
			Amount:      recExp.Amount,
			Currency:    recExp.Currency,
			Date:        currentDate,
			Tags:        recExp.Tags,
		}
		expenses = append(expenses, exp)
		switch recExp.Interval {
		case "daily":
			currentDate = currentDate.AddDate(0, 0, 1)
		case "weekly":
			currentDate = currentDate.AddDate(0, 0, 7)
		case "monthly":
			currentDate = currentDate.AddDate(0, 1, 0)
		case "yearly":
			currentDate = currentDate.AddDate(1, 0, 0)
		default:
			return expenses
		}
		if recExp.Occurrences > 0 {
			occurrencesToGenerate--
		}
	}
	return expenses
}

func (s *databaseStore) ensureBudgetOwnership(userID, budgetID string) error {
	if userID == "" || budgetID == "" {
		return fmt.Errorf("userID and budgetID are required")
	}
	var owner string
	if err := s.db.QueryRow(`SELECT user_id FROM budgets WHERE id = $1`, budgetID).Scan(&owner); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("budget not found")
		}
		return fmt.Errorf("failed to verify budget ownership: %v", err)
	}
	if owner != userID {
		return fmt.Errorf("budget not found")
	}
	return nil
}

func normalizeMonth(t time.Time) time.Time {
	if t.IsZero() {
		t = time.Now()
	}
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
}

func nullString(val string) interface{} {
	if val == "" {
		return nil
	}
	return val
}

// DB exposes the underlying sql.DB for repositories that need direct access (e.g. users).
func (s *databaseStore) DB() *sql.DB {
	return s.db
}

package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

	createUserSettingsTableSQL = `
CREATE TABLE IF NOT EXISTS user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    categories TEXT NOT NULL,
    currency VARCHAR(255) NOT NULL,
    start_date INTEGER NOT NULL
);
`

	createExpensesTableSQL = `
CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recurring_id UUID,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(255) NOT NULL,
    amount NUMERIC(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    tags TEXT
);
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
    tags TEXT
);
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
	queries := []string{createUsersTableSQL, createUserSettingsTableSQL, createExpensesTableSQL, createRecurringExpensesTableSQL}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
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
        INSERT INTO user_settings (user_id, categories, currency, start_date)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id) DO NOTHING
    `, userID, string(categoriesJSON), defaults.Currency, defaults.StartDate)
	return err
}

func (s *databaseStore) GetConfig(userID string) (*Config, error) {
	if err := s.EnsureUserDefaults(userID); err != nil {
		return nil, err
	}
	var categoriesStr, currency string
	var startDate int
	err := s.db.QueryRow(`
        SELECT categories, currency, start_date
        FROM user_settings
        WHERE user_id = $1
    `, userID).Scan(&categoriesStr, &currency, &startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to load user config: %v", err)
	}
	var config Config
	config.Currency = currency
	config.StartDate = startDate
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

func scanExpense(scanner interface{ Scan(...any) error }) (Expense, error) {
	var expense Expense
	var tagsStr sql.NullString
	var recurringID sql.NullString
	var userID string
	err := scanner.Scan(&expense.ID, &userID, &recurringID, &expense.Name, &expense.Category, &expense.Amount, &expense.Currency, &expense.Date, &tagsStr)
	if err != nil {
		return Expense{}, err
	}
	expense.UserID = userID
	if recurringID.Valid {
		expense.RecurringID = recurringID.String
	}
	if tagsStr.Valid && tagsStr.String != "" {
		if err := json.Unmarshal([]byte(tagsStr.String), &expense.Tags); err != nil {
			return Expense{}, fmt.Errorf("failed to parse tags for expense %s: %v", expense.ID, err)
		}
	}
	return expense, nil
}

func (s *databaseStore) GetAllExpenses(userID string) ([]Expense, error) {
	rows, err := s.db.Query(`
        SELECT id, user_id, recurring_id, name, category, amount, currency, date, tags
        FROM expenses
        WHERE user_id = $1
        ORDER BY date DESC
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
        SELECT id, user_id, recurring_id, name, category, amount, currency, date, tags
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
	if expense.Currency == "" {
		currency, err := s.GetCurrency(userID)
		if err != nil {
			return err
		}
		expense.Currency = currency
	}
	if expense.Date.IsZero() {
		expense.Date = time.Now()
	}
	tagsJSON, err := json.Marshal(expense.Tags)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
        INSERT INTO expenses (id, user_id, recurring_id, name, category, amount, currency, date, tags)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `, expense.ID, userID, nullString(expense.RecurringID), expense.Name, expense.Category, expense.Amount, expense.Currency, expense.Date, string(tagsJSON))
	return err
}

func (s *databaseStore) UpdateExpense(userID, id string, expense Expense) error {
	tagsJSON, err := json.Marshal(expense.Tags)
	if err != nil {
		return err
	}
	if expense.Currency == "" {
		currency, err := s.GetCurrency(userID)
		if err != nil {
			return err
		}
		expense.Currency = currency
	}
	res, err := s.db.Exec(`
        UPDATE expenses
        SET name = $1, category = $2, amount = $3, currency = $4, date = $5, tags = $6, recurring_id = $7
        WHERE id = $8 AND user_id = $9
    `, expense.Name, expense.Category, expense.Amount, expense.Currency, expense.Date, string(tagsJSON), nullString(expense.RecurringID), id, userID)
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

	stmt, err := tx.Prepare(pq.CopyIn("expenses", "id", "user_id", "recurring_id", "name", "category", "amount", "currency", "date", "tags"))
	if err != nil {
		return fmt.Errorf("failed to prepare bulk insert: %v", err)
	}
	defer stmt.Close()

	for _, exp := range expenses {
		if exp.ID == "" {
			exp.ID = uuid.New().String()
		}
		exp.UserID = userID
		if exp.Currency == "" {
			currency, err := s.GetCurrency(userID)
			if err != nil {
				return err
			}
			exp.Currency = currency
		}
		tagsJSON, err := json.Marshal(exp.Tags)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(exp.ID, userID, nullString(exp.RecurringID), exp.Name, exp.Category, exp.Amount, exp.Currency, exp.Date, string(tagsJSON)); err != nil {
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
        SELECT id, user_id, name, amount, currency, category, start_date, interval, occurrences, tags
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
		err := rows.Scan(&rec.ID, &rec.UserID, &rec.Name, &rec.Amount, &rec.Currency, &rec.Category, &rec.StartDate, &rec.Interval, &rec.Occurrences, &tagsStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recurring expense: %v", err)
		}
		if tagsStr.Valid && tagsStr.String != "" {
			if err := json.Unmarshal([]byte(tagsStr.String), &rec.Tags); err != nil {
				return nil, fmt.Errorf("failed to parse tags for recurring expense %s: %v", rec.ID, err)
			}
		}
		results = append(results, rec)
	}
	return results, nil
}

func (s *databaseStore) GetRecurringExpense(userID, id string) (RecurringExpense, error) {
	var rec RecurringExpense
	var tagsStr sql.NullString
	err := s.db.QueryRow(`
        SELECT id, user_id, name, amount, currency, category, start_date, interval, occurrences, tags
        FROM recurring_expenses
        WHERE user_id = $1 AND id = $2
    `, userID, id).Scan(&rec.ID, &rec.UserID, &rec.Name, &rec.Amount, &rec.Currency, &rec.Category, &rec.StartDate, &rec.Interval, &rec.Occurrences, &tagsStr)
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
	return rec, nil
}

func (s *databaseStore) AddRecurringExpense(userID string, recurringExpense RecurringExpense) error {
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
        INSERT INTO recurring_expenses (id, user_id, name, amount, currency, category, start_date, interval, occurrences, tags)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `, recurringExpense.ID, userID, recurringExpense.Name, recurringExpense.Amount, recurringExpense.Currency, recurringExpense.Category, recurringExpense.StartDate, recurringExpense.Interval, recurringExpense.Occurrences, string(tagsJSON))
	if err != nil {
		return fmt.Errorf("failed to insert recurring expense: %v", err)
	}

	expensesToAdd := generateExpensesFromRecurring(userID, recurringExpense, false)
	if err := bulkInsertExpenses(tx, expensesToAdd); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *databaseStore) UpdateRecurringExpense(userID, id string, recurringExpense RecurringExpense, updateAll bool) error {
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
        SET name = $1, amount = $2, currency = $3, category = $4, start_date = $5, interval = $6, occurrences = $7, tags = $8
        WHERE id = $9 AND user_id = $10
    `, recurringExpense.Name, recurringExpense.Amount, recurringExpense.Currency, recurringExpense.Category, recurringExpense.StartDate, recurringExpense.Interval, recurringExpense.Occurrences, string(tagsJSON), id, userID)
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
	if err := bulkInsertExpenses(tx, expensesToAdd); err != nil {
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

func bulkInsertExpenses(tx *sql.Tx, expenses []Expense) error {
	if len(expenses) == 0 {
		return nil
	}
	stmt, err := tx.Prepare(pq.CopyIn("expenses", "id", "user_id", "recurring_id", "name", "category", "amount", "currency", "date", "tags"))
	if err != nil {
		return fmt.Errorf("failed to prepare expense bulk insert: %v", err)
	}
	defer stmt.Close()

	for _, exp := range expenses {
		tagsJSON, err := json.Marshal(exp.Tags)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(exp.ID, exp.UserID, nullString(exp.RecurringID), exp.Name, exp.Category, exp.Amount, exp.Currency, exp.Date, string(tagsJSON)); err != nil {
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

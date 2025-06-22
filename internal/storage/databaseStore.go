package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"slices"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// databaseStore implements Storage interface - for PostgreSQL database storage
type databaseStore struct {
	db *sql.DB
}

// initializes the PostgreSQL storage backend
func InitializePostgresStore(baseConfig SystemConfig) (*databaseStore, error) {
	db, err := sql.Open("postgres", baseConfig.StorageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %v", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %v", err)
	}
	log.Println("Connected to PostgreSQL database")
	createTables(db)
	return &databaseStore{db: db}, nil
}

// primitive methods

func createTables(db *sql.DB) {
	createExpensesTable(db)
	createConfigTable(db)
	createRecurringExpensesTable(db)
}

func createExpensesTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS expenses (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		date DATE NOT NULL,
		amount DECIMAL(10, 2) NOT NULL,
		category VARCHAR(255) NOT NULL,
		tags TEXT NOT NULL,
		impact VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
	`
	db.Exec(query)
}

func createConfigTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS config (
		id VARCHAR(255) PRIMARY KEY DEFAULT 'default',
		categories TEXT NOT NULL,
		currency VARCHAR(255) NOT NULL,
		start_date INTEGER NOT NULL,
		tags TEXT NOT NULL,
		default_tags TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
	`
	db.Exec(query)
}

func createRecurringExpensesTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS recurring_expenses (
		expense_id VARCHAR(255) PRIMARY KEY,
		expense_name VARCHAR(255) NOT NULL,
		expense_date DATE NOT NULL,
		expense_amount DECIMAL(10, 2) NOT NULL,
		expense_category VARCHAR(255) NOT NULL,
		expense_tags TEXT NOT NULL,
		expense_impact VARCHAR(255) NOT NULL,
		date DATE NOT NULL,
		interval VARCHAR(255) NOT NULL,
		occurrences INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
	`
	db.Exec(query)
}

func (s *databaseStore) saveConfig(config *Config) error {
	categoriesJSON, _ := json.Marshal(config.Categories)
	tagsJSON, _ := json.Marshal(config.Tags)
	defaultTagsJSON, _ := json.Marshal(config.DefaultTags)
	query := `
		INSERT INTO config (id, categories, currency, start_date, tags, default_tags, updated_at)
		VALUES ('default', $1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			categories = $1,
			currency = $2,
			start_date = $3,
			tags = $4,
			default_tags = $5,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := s.db.Exec(query, string(categoriesJSON), config.Currency, config.StartDate, string(tagsJSON), string(defaultTagsJSON))
	return err
}

// ------------------------------------------------------------
// DatabaseStore interface methods
// ------------------------------------------------------------

func (s *databaseStore) Close() error {
	return s.db.Close()
}

func (s *databaseStore) GetConfig() (*Config, error) {
	query := `SELECT categories, currency, start_date, tags, default_tags FROM config WHERE id = 'default'`
	var categoriesStr, currency, tagsStr, defaultTagsStr string
	var startDate int
	err := s.db.QueryRow(query).Scan(&categoriesStr, &currency, &startDate, &tagsStr, &defaultTagsStr)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create default config if it doesn't exist
			config := &Config{}
			config.SetBaseConfig()
			return config, s.saveConfig(config)
		}
		return nil, fmt.Errorf("failed to get config: %v", err)
	}
	var categories, tags, defaultTags []string
	if err := json.Unmarshal([]byte(categoriesStr), &categories); err != nil {
		return nil, fmt.Errorf("failed to parse categories: %v", err)
	}
	if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
		return nil, fmt.Errorf("failed to parse tags: %v", err)
	}
	if err := json.Unmarshal([]byte(defaultTagsStr), &defaultTags); err != nil {
		return nil, fmt.Errorf("failed to parse default tags: %v", err)
	}
	return &Config{
		Categories:  categories,
		Currency:    currency,
		StartDate:   startDate,
		Tags:        tags,
		DefaultTags: defaultTags,
	}, nil
}

func (s *databaseStore) GetCategories() ([]string, error) {
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}
	return config.Categories, nil
}

func (s *databaseStore) UpdateCategories(categories []string) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}
	config.Categories = categories
	return s.saveConfig(config)
}

func (s *databaseStore) GetCurrency() (string, error) {
	config, err := s.GetConfig()
	if err != nil {
		return "", err
	}
	return config.Currency, nil
}

func (s *databaseStore) UpdateCurrency(currency string) error {
	if !slices.Contains(supportedCurrencies, currency) {
		return fmt.Errorf("invalid currency: %s", currency)
	}
	config, err := s.GetConfig()
	if err != nil {
		return err
	}
	config.Currency = currency
	return s.saveConfig(config)
}

func (s *databaseStore) GetStartDate() (int, error) {
	config, err := s.GetConfig()
	if err != nil {
		return 0, err
	}
	return config.StartDate, nil
}

func (s *databaseStore) UpdateStartDate(startDate int) error {
	if startDate < 0 || startDate > 31 {
		return fmt.Errorf("invalid start date: %d", startDate)
	}
	config, err := s.GetConfig()
	if err != nil {
		return err
	}
	config.StartDate = startDate
	return s.saveConfig(config)
}

func (s *databaseStore) GetTags() ([]string, error) {
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}
	return config.Tags, nil
}

func (s *databaseStore) UpdateTags(tags []string) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}
	config.Tags = tags
	return s.saveConfig(config)
}

func (s *databaseStore) GetDefaultTags() ([]string, error) {
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}
	return config.DefaultTags, nil
}

func (s *databaseStore) UpdateDefaultTags(tags []string) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}
	config.DefaultTags = tags
	return s.saveConfig(config)
}

func (s *databaseStore) GetAllExpenses() ([]*Expense, error) {
	query := `SELECT id, name, date, amount, category, tags, impact FROM expenses ORDER BY date DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query expenses: %v", err)
	}
	defer rows.Close()
	var expenses []*Expense
	for rows.Next() {
		var expense Expense
		var tagsStr string
		err := rows.Scan(&expense.ID, &expense.Name, &expense.Date, &expense.Amount, &expense.Category, &tagsStr, &expense.Impact)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense: %v", err)
		}
		if err := json.Unmarshal([]byte(tagsStr), &expense.Tags); err != nil {
			return nil, fmt.Errorf("failed to parse tags: %v", err)
		}
		expenses = append(expenses, &expense)
	}
	return expenses, nil
}

func (s *databaseStore) GetExpense(id string) (*Expense, error) {
	query := `SELECT id, name, date, amount, category, tags, impact FROM expenses WHERE id = $1`
	var expense Expense
	var tagsStr string
	err := s.db.QueryRow(query, id).Scan(&expense.ID, &expense.Name, &expense.Date, &expense.Amount, &expense.Category, &tagsStr, &expense.Impact)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("expense with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get expense: %v", err)
	}
	if err := json.Unmarshal([]byte(tagsStr), &expense.Tags); err != nil {
		return nil, fmt.Errorf("failed to parse tags: %v", err)
	}
	return &expense, nil
}

func (s *databaseStore) AddExpense(expense *Expense) error {
	if expense.ID == "" {
		expense.ID = uuid.New().String()
	}
	tagsJSON, _ := json.Marshal(expense.Tags)
	query := `
		INSERT INTO expenses (id, name, date, amount, category, tags, impact, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err := s.db.Exec(query, expense.ID, expense.Name, expense.Date, expense.Amount, expense.Category, string(tagsJSON), expense.Impact)
	return err
}

func (s *databaseStore) RemoveExpense(id string) error {
	query := `DELETE FROM expenses WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("expense with ID %s not found", id)
	}
	return nil
}

func (s *databaseStore) UpdateExpense(id string, expense *Expense) error {
	tagsJSON, _ := json.Marshal(expense.Tags)
	query := `
		UPDATE expenses 
		SET name = $1, date = $2, amount = $3, category = $4, tags = $5, impact = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
	`
	result, err := s.db.Exec(query, expense.Name, expense.Date, expense.Amount, expense.Category, string(tagsJSON), expense.Impact, id)
	if err != nil {
		return fmt.Errorf("failed to update expense: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("expense with ID %s not found", id)
	}
	return nil
}

func (s *databaseStore) GetRecurringExpenses() ([]RecurringExpense, error) {
	query := `SELECT expense_id, expense_name, expense_date, expense_amount, expense_category, expense_tags, expense_impact, date, interval, occurrences FROM recurring_expenses`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query recurring expenses: %v", err)
	}
	defer rows.Close()
	var recurringExpenses []RecurringExpense
	for rows.Next() {
		var re RecurringExpense
		var tagsStr string
		err := rows.Scan(&re.Expense.ID, &re.Expense.Name, &re.Expense.Date, &re.Expense.Amount, &re.Expense.Category, &tagsStr, &re.Expense.Impact, &re.Date, &re.Interval, &re.Occurrences)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recurring expense: %v", err)
		}
		if err := json.Unmarshal([]byte(tagsStr), &re.Expense.Tags); err != nil {
			return nil, fmt.Errorf("failed to parse tags: %v", err)
		}
		recurringExpenses = append(recurringExpenses, re)
	}
	return recurringExpenses, nil
}

func (s *databaseStore) GetRecurringExpense(id string) (*RecurringExpense, error) {
	query := `SELECT expense_id, expense_name, expense_date, expense_amount, expense_category, expense_tags, expense_impact, date, interval, occurrences FROM recurring_expenses WHERE expense_id = $1`
	var re RecurringExpense
	var tagsStr string
	err := s.db.QueryRow(query, id).Scan(&re.Expense.ID, &re.Expense.Name, &re.Expense.Date, &re.Expense.Amount, &re.Expense.Category, &tagsStr, &re.Expense.Impact, &re.Date, &re.Interval, &re.Occurrences)
	if err != nil {
		return nil, fmt.Errorf("failed to query recurring expense: %v", err)
	}
	if err := json.Unmarshal([]byte(tagsStr), &re.Expense.Tags); err != nil {
		return nil, fmt.Errorf("failed to parse tags: %v", err)
	}
	return &re, nil
}

func (s *databaseStore) AddRecurringExpense(recurringExpense *RecurringExpense) error {
	if recurringExpense.ID == "" {
		recurringExpense.ID = uuid.New().String()
	}
	tagsJSON, _ := json.Marshal(recurringExpense.Expense.Tags)
	query := `
		INSERT INTO recurring_expenses (expense_id, expense_name, expense_date, expense_amount, expense_category, expense_tags, expense_impact, date, interval, occurrences)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := s.db.Exec(query,
		recurringExpense.ID,
		recurringExpense.Expense.Name,
		recurringExpense.Expense.Date,
		recurringExpense.Expense.Amount,
		recurringExpense.Expense.Category,
		string(tagsJSON),
		recurringExpense.Expense.Impact,
		recurringExpense.Date,
		recurringExpense.Interval,
		recurringExpense.Occurrences)
	return err
}

func (s *databaseStore) RemoveRecurringExpense(id string) error {
	query := `DELETE FROM recurring_expenses WHERE expense_id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete recurring expense: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("recurring expense with ID %s not found", id)
	}
	return nil
}

func (s *databaseStore) UpdateRecurringExpense(id string, recurringExpense *RecurringExpense) error {
	tagsJSON, _ := json.Marshal(recurringExpense.Expense.Tags)
	query := `
		UPDATE recurring_expenses 
		SET expense_name = $1, expense_date = $2, expense_amount = $3, expense_category = $4, expense_tags = $5, expense_impact = $6, date = $7, interval = $8, occurrences = $9
		WHERE expense_id = $10
	`
	result, err := s.db.Exec(query,
		recurringExpense.Expense.Name,
		recurringExpense.Expense.Date,
		recurringExpense.Expense.Amount,
		recurringExpense.Expense.Category,
		string(tagsJSON),
		recurringExpense.Expense.Impact,
		recurringExpense.Date,
		recurringExpense.Interval,
		recurringExpense.Occurrences,
		id)
	if err != nil {
		return fmt.Errorf("failed to update recurring expense: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("recurring expense with ID %s not found", id)
	}
	return nil
}

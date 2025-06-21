package storage

import (
	"database/sql"
	"fmt"
	"os"
)

// Storage interface for all storage types
type Storage interface {
	Initialize() error
	Close() error
	GetConfig() (*Config, error)

	GetCategories() ([]string, error)
	UpdateCategories(categories []string) error

	GetCurrency() (string, error)
	UpdateCurrency(currency string) error
	GetStartDate() (int, error)
	UpdateStartDate(startDate int) error

	GetDefaultTags() ([]string, error)
	UpdateDefaultTags(tags []string) error
	GetTags() ([]string, error)
	UpdateTags(tags []string) error

	GetAllExpenses() ([]*Expense, error)
	GetExpense(id string) (*Expense, error)
	AddExpense(expense *Expense) error
	RemoveExpense(id string) error
	UpdateExpense(expense *Expense) error
}

// Initialize initializes the storage
func InitializeStorage() (Storage, error) {
	baseConfig := SystemConfig{}
	baseConfig.SetStorageConfig()
	switch baseConfig.StorageType {
	case BackendTypeJSON:
		return InitializeJsonStore()
	case BackendTypeMySQL:
		return InitializeMySqlStore()
	case BackendTypePostgres:
		return InitializePostgresStore()
	case BackendTypeSQLite:
		return InitializeSqliteStore()
	}
	return nil, fmt.Errorf("invalid data store: %s", baseConfig.StorageType)
}

// mySqlStore implements Storage interface - for MySQL database storage
type mySqlStore struct {
	db *sql.DB
}

// initializes the MySQL storage backend
func InitializeMySqlStore() (*mySqlStore, error) {
	db, err := sql.Open("mysql", os.Getenv("MYSQL_URL"))
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL database: %v", err)
	}
	return &mySqlStore{db: db}, nil
}

// postgresStore implements Storage interface - for PostgreSQL database storage
type postgresStore struct {
	db *sql.DB
}

// initializes the PostgreSQL storage backend
func InitializePostgresStore() (*postgresStore, error) {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %v", err)
	}
	return &postgresStore{db: db}, nil
}

// sqliteStore implements Storage interface - for SQLite database storage
type sqliteStore struct {
	db *sql.DB
}

// initializes the SQLite storage backend
func InitializeSqliteStore() (*sqliteStore, error) {
	db, err := sql.Open("sqlite3", os.Getenv("SQLITE_URL"))
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %v", err)
	}
	return &sqliteStore{db: db}, nil
}

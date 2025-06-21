package storage

import (
	"errors"
	"os"
	"time"
)

type BackendType string

const (
	BackendTypeJSON     BackendType = "json"
	BackendTypeSQLite   BackendType = "sqlite"
	BackendTypePostgres BackendType = "postgres"
	BackendTypeMySQL    BackendType = "mysql"
)

type ExpenseImpact string

const (
	EIGain ExpenseImpact = "gain"
	EILoss ExpenseImpact = "loss"
)

type Config struct {
	Categories  []string
	Currency    string
	StartDate   int
	Tags        []string
	DefaultTags []string
}

type SystemConfig struct {
	StorageURL  string
	StorageType BackendType
	StorageUser string
	StoragePass string
}

type Expense struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Tags     []string      `json:"tags"`
	Impact   ExpenseImpact `json:"impact"`
	Category string        `json:"category"`
	Amount   float64       `json:"amount"`
	Date     time.Time     `json:"date"`
}

var defaultCategories = []string{
	"Food",
	"Groceries",
	"Travel",
	"Rent",
	"Utilities",
	"Entertainment",
	"Healthcare",
	"Shopping",
	"Miscellaneous",
	"Income",
}

var supportedCurrencies = []string{
	"usd", // US Dollar
	"eur", // Euro
	"gbp", // British Pound
	"jpy", // Japanese Yen
	"cny", // Chinese Yuan
	"krw", // Korean Won
	"inr", // Indian Rupee
	"rub", // Russian Ruble
	"brl", // Brazilian Real
	"zar", // South African Rand
	"aed", // UAE Dirham
	"aud", // Australian Dollar
	"cad", // Canadian Dollar
	"chf", // Swiss Franc
	"hkd", // Hong Kong Dollar
	"sgd", // Singapore Dollar
	"thb", // Thai Baht
	"try", // Turkish Lira
	"mxn", // Mexican Peso
	"php", // Philippine Peso
	"pln", // Polish Złoty
	"sek", // Swedish Krona
	"nzd", // New Zealand Dollar
	"dkk", // Danish Krone
	"idr", // Indonesian Rupiah
	"ils", // Israeli New Shekel
	"vnd", // Vietnamese Dong
	"myr", // Malaysian Ringgit
}

func (e *Expense) Validate() error {
	if e.Name == "" {
		return errors.New("expense name is required")
	}
	if e.Category == "" {
		return errors.New("category is required")
	}
	return nil
}

func (c *Config) SetBaseConfig() {
	c.Categories = defaultCategories
	c.Currency = "usd"
	c.StartDate = 1
	c.Tags = []string{}
	c.DefaultTags = []string{}
}

func (c *SystemConfig) SetStorageConfig() {
	c.StorageType = backendTypeFromEnv(os.Getenv("STORAGE_TYPE"))
	c.StorageURL = backendURLFromEnv(os.Getenv("STORAGE_TYPE"))
	c.StorageUser = os.Getenv("STORAGE_USER")
	c.StoragePass = os.Getenv("STORAGE_PASS")
}

func backendTypeFromEnv(env string) BackendType {
	switch env {
	case "json":
		return BackendTypeJSON
	case "sqlite":
		return BackendTypeSQLite
	case "postgres":
		return BackendTypePostgres
	case "mysql":
		return BackendTypeMySQL
	}
	return BackendTypeJSON
}

func backendURLFromEnv(env string) string {
	switch env {
	case "json":
		return "data/expenses.json"
	case "sqlite":
		return "data/expenses.sqlite"
	case "postgres":
		return "" // not sure about default url
	case "mysql":
		return "" // not sure about default url
	}
	return "data/expenses.json"
}

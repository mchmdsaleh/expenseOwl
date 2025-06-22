package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

// JSONStore implementats Storage interface - for JSON file storage
type jsonStore struct {
	configPath string
	filePath   string
	mu         sync.RWMutex
}

type expensesFileData struct {
	Expenses []*Expense `json:"expenses"`
}

// initializes the JSON storage backend
func InitializeJsonStore(baseConfig SystemConfig) (*jsonStore, error) {
	configPath := filepath.Join(baseConfig.StorageURL, "config.json")
	filePath := filepath.Join(baseConfig.StorageURL, "expenses.json")
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}

	// create expenses file if it doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		initialData := expensesFileData{Expenses: []*Expense{}}
		data, err := json.Marshal(initialData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal initial data: %v", err)
		}
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to create storage file: %v", err)
		}
		log.Println("Created expense storage file")
	} else {
		log.Println("Found existing expense storage file")
	}

	// create config file if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		initialConfig := Config{}
		initialConfig.SetBaseConfig()
		data, err := json.Marshal(initialConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal initial config: %v", err)
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to create config file: %v", err)
		}
		log.Println("Created expense storage config")
	} else {
		log.Println("Found existing expense storage config")
	}

	return &jsonStore{
		configPath: configPath,
		filePath:   filePath,
	}, nil
}

// primitive methods

func (s *jsonStore) readExpensesFile(path string) (*expensesFileData, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data expensesFileData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}
	log.Println("Read expenses file")
	return &data, nil
}

func (s *jsonStore) writeExpensesFile(path string, data *expensesFileData) error {
	content, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	log.Println("Wrote expenses file")
	return os.WriteFile(path, content, 0644)
}

func (s *jsonStore) readConfigFile(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data Config
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}
	log.Println("Read config file")
	return &data, nil
}

func (s *jsonStore) writeConfigFile(path string, data *Config) error {
	content, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	log.Println("Wrote config file")
	return os.WriteFile(path, content, 0644)
}

// ------------------------------------------------------------
// JSONStore interface methods
// ------------------------------------------------------------

func (s *jsonStore) Close() error {
	return nil
}

func (s *jsonStore) GetConfig() (*Config, error) {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	return data, nil
}

func (s *jsonStore) GetCategories() ([]string, error) {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	return data.Categories, nil
}

func (s *jsonStore) UpdateCategories(categories []string) error {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	data.Categories = categories
	return s.writeConfigFile(s.configPath, data)
}

func (s *jsonStore) GetCurrency() (string, error) {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %v", err)
	}
	return data.Currency, nil
}

func (s *jsonStore) UpdateCurrency(currency string) error {
	if !slices.Contains(supportedCurrencies, currency) {
		return fmt.Errorf("invalid currency: %s", currency)
	}
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	data.Currency = currency
	return s.writeConfigFile(s.configPath, data)
}

func (s *jsonStore) GetStartDate() (int, error) {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read config file: %v", err)
	}
	return data.StartDate, nil
}

func (s *jsonStore) UpdateStartDate(startDate int) error {
	if startDate < 0 || startDate > 31 {
		return fmt.Errorf("invalid start date: %d", startDate)
	}
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	data.StartDate = startDate
	return s.writeConfigFile(s.configPath, data)
}

func (s *jsonStore) GetTags() ([]string, error) {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	return data.Tags, nil
}

func (s *jsonStore) UpdateTags(tags []string) error {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	data.Tags = tags
	return s.writeConfigFile(s.configPath, data)
}

func (s *jsonStore) GetDefaultTags() ([]string, error) {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	return data.DefaultTags, nil
}

func (s *jsonStore) UpdateDefaultTags(tags []string) error {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	data.DefaultTags = tags
	return s.writeConfigFile(s.configPath, data)
}

func (s *jsonStore) GetRecurringExpenses() ([]RecurringExpense, error) {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	return data.RecurringExpenses, nil
}

func (s *jsonStore) GetRecurringExpense(id string) (*RecurringExpense, error) {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	for i, r := range data.RecurringExpenses {
		if r.ID == id {
			return &data.RecurringExpenses[i], nil
		}
	}
	return nil, fmt.Errorf("recurring expense with ID %s not found", id)
}

func (s *jsonStore) AddRecurringExpense(recurringExpense *RecurringExpense) error {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	if recurringExpense.ID == "" {
		recurringExpense.ID = uuid.New().String()
	}
	data.RecurringExpenses = append(data.RecurringExpenses, *recurringExpense)
	return s.writeConfigFile(s.configPath, data)
}

func (s *jsonStore) RemoveRecurringExpense(id string) error {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	data.RecurringExpenses = slices.DeleteFunc(data.RecurringExpenses, func(r RecurringExpense) bool {
		return r.ID == id
	})
	return s.writeConfigFile(s.configPath, data)
}

func (s *jsonStore) UpdateRecurringExpense(id string, recurringExpense *RecurringExpense) error {
	data, err := s.readConfigFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	for i, r := range data.RecurringExpenses {
		if r.ID == id {
			data.RecurringExpenses[i] = *recurringExpense
			return s.writeConfigFile(s.configPath, data)
		}
	}
	return fmt.Errorf("recurring expense with ID %s not found", id)
}

func (s *jsonStore) GetAllExpenses() ([]*Expense, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := s.readExpensesFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage file: %v", err)
	}
	return data.Expenses, nil
}

func (s *jsonStore) GetExpense(id string) (*Expense, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := s.readExpensesFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage file: %v", err)
	}
	for i, exp := range data.Expenses {
		if exp.ID == id {
			log.Printf("Retrieved expense with ID %s\n", id)
			return data.Expenses[i], nil
		}
	}
	return nil, fmt.Errorf("expense with ID %s not found", id)
}

func (s *jsonStore) AddExpense(expense *Expense) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := s.readExpensesFile(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to read storage file: %v", err)
	}
	if expense.ID == "" {
		expense.ID = uuid.New().String()
	}
	if expense.Date.IsZero() {
		expense.Date = time.Now()
	}
	data.Expenses = append(data.Expenses, expense)
	log.Printf("Added expense with ID %s\n", expense.ID)
	return s.writeExpensesFile(s.filePath, data)
}

func (s *jsonStore) RemoveExpense(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := s.readExpensesFile(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to read storage file: %v", err)
	}
	found := false
	newExpenses := make([]*Expense, 0, len(data.Expenses)-1)
	for _, exp := range data.Expenses {
		if exp.ID != id {
			newExpenses = append(newExpenses, exp)
		} else {
			found = true
		}
	}
	if !found {
		log.Printf("Expense with ID %s not found\n", id)
		return fmt.Errorf("expense with ID %s not found", id)
	}
	log.Printf("Deleted expense with ID %s\n", id)
	data.Expenses = newExpenses
	return s.writeExpensesFile(s.filePath, data)
}

func (s *jsonStore) UpdateExpense(id string, expense *Expense) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := s.readExpensesFile(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to read storage file: %v", err)
	}
	found := false
	for i, exp := range data.Expenses {
		if exp.ID == id {
			data.Expenses[i] = expense
			found = true
			break
		}
	}
	if !found {
		log.Printf("expense with ID %s not found\n", id)
		return fmt.Errorf("expense with ID %s not found", id)
	}
	log.Printf("Edited expense with ID %s\n", id)
	return s.writeExpensesFile(s.filePath, data)
}

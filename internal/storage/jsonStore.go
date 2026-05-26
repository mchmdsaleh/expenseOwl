package storage

import (
	"fmt"
	"time"

	"github.com/tanq16/expenseowl/internal/encryption"
)

// jsonStore is intentionally unimplemented in multi-user mode.
type jsonStore struct{}

func InitializeJsonStore(baseConfig SystemConfig) (*jsonStore, error) {
	return nil, fmt.Errorf("json storage backend is not supported in multi-user mode; please configure PostgreSQL")
}

func (s *jsonStore) Close() error { return nil }

func (s *jsonStore) EnsureUserDefaults(userID string) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) GetConfig(userID string) (*Config, error) {
	return nil, fmt.Errorf("json backend not available")
}
func (s *jsonStore) GetCategories(userID string) ([]string, error) {
	return nil, fmt.Errorf("json backend not available")
}
func (s *jsonStore) UpdateCategories(userID string, categories []string) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) GetCurrency(userID string) (string, error) {
	return "", fmt.Errorf("json backend not available")
}
func (s *jsonStore) UpdateCurrency(userID, currency string) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) GetStartDate(userID string) (int, error) {
	return 0, fmt.Errorf("json backend not available")
}
func (s *jsonStore) UpdateStartDate(userID string, startDate int) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) GetEndOfMonth(userID string) (bool, error) {
	return false, fmt.Errorf("json backend not available")
}
func (s *jsonStore) UpdateEndOfMonth(userID string, endOfMonth bool) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) GetBudgets(userID string) ([]Budget, error) {
	return nil, fmt.Errorf("json backend not available")
}
func (s *jsonStore) AddBudget(userID string, budget Budget) (Budget, error) {
	return Budget{}, fmt.Errorf("json backend not available")
}
func (s *jsonStore) UpdateBudget(userID, id string, budget Budget) (Budget, error) {
	return Budget{}, fmt.Errorf("json backend not available")
}
func (s *jsonStore) RemoveBudget(userID, id string) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) GetBudgetSummaries(userID string, month time.Time) ([]BudgetSummary, error) {
	return nil, fmt.Errorf("json backend not available")
}
func (s *jsonStore) UpsertBudgetOverride(userID, budgetID string, month time.Time, amount float64) (BudgetOverride, error) {
	return BudgetOverride{}, fmt.Errorf("json backend not available")
}
func (s *jsonStore) DeleteBudgetOverride(userID, overrideID string) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) UpsertBudgetAdjustment(userID, budgetID string, month time.Time, amount float64) (BudgetAdjustment, error) {
	return BudgetAdjustment{}, fmt.Errorf("json backend not available")
}
func (s *jsonStore) DeleteBudgetAdjustment(userID, adjustmentID string) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) GetRecurringExpenses(userID string) ([]RecurringExpense, error) {
	return nil, fmt.Errorf("json backend not available")
}
func (s *jsonStore) GetRecurringExpense(userID, id string) (RecurringExpense, error) {
	return RecurringExpense{}, fmt.Errorf("json backend not available")
}
func (s *jsonStore) AddRecurringExpense(userID string, recurringExpense RecurringExpense, enc *encryption.Manager) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) RemoveRecurringExpense(userID, id string, removeAll bool) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) UpdateRecurringExpense(userID, id string, recurringExpense RecurringExpense, updateAll bool, enc *encryption.Manager) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) GetAllExpenses(userID string) ([]Expense, error) {
	return nil, fmt.Errorf("json backend not available")
}
func (s *jsonStore) GetExpense(userID, id string) (Expense, error) {
	return Expense{}, fmt.Errorf("json backend not available")
}
func (s *jsonStore) AddExpense(userID string, expense Expense) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) RemoveExpense(userID, id string) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) AddMultipleExpenses(userID string, expenses []Expense) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) RemoveMultipleExpenses(userID string, ids []string) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) UpdateExpense(userID, id string, expense Expense) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) GetAIContext(userID string) (string, error) {
	return "", fmt.Errorf("json backend not available")
}
func (s *jsonStore) UpdateAIContext(userID, context string) error {
	return fmt.Errorf("json backend not available")
}
func (s *jsonStore) GetAIConfig(userID string) (*AIConfig, error) {
	return nil, fmt.Errorf("json backend not available")
}
func (s *jsonStore) UpdateAIConfig(userID string, config AIConfig) error {
	return fmt.Errorf("json backend not available")
}

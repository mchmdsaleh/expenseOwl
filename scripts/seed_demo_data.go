package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/tanq16/expenseowl/internal/storage"
	"github.com/tanq16/expenseowl/internal/user"
)

// DemoTransaction represents a template transaction for demo data
type DemoTransaction struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	DayOfMonth int   `json:"dayOfMonth"`
	Tags     []string `json:"tags"`
}

// DemoBudget represents a template budget for demo data
type DemoBudget struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
}

// DemoData holds credentials and transaction template for easy reuse
type DemoData struct {
	Credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		UserID   string `json:"userID"`
	} `json:"credentials"`
	Budgets      []DemoBudget      `json:"budgets"`
	Transactions []DemoTransaction `json:"transactions"`
	GeneratedAt  time.Time         `json:"generatedAt"`
	Note         string            `json:"note"`
}

func main() {
	// Demo account credentials
	demoEmail := "demo@expenseowl.com"
	demoPassword := "Demo123456!"

	// Initialize storage
	store, err := storage.InitializeStorage()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Get database connection
	dbProvider, ok := store.(interface{ DB() *sql.DB })
	if !ok {
		log.Fatalf("PostgreSQL storage is required")
	}
	db := dbProvider.DB()

	ctx := context.Background()

	// Create user service
	userService := user.NewService(user.NewRepository(db))

	// Check if demo user already exists
	existingUser, err := userService.GetByEmail(ctx, demoEmail)
	var demoUserID string
	
	if err == nil && existingUser != nil {
		log.Printf("Demo user already exists: %s (ID: %s)\n", existingUser.Email, existingUser.ID)
		demoUserID = existingUser.ID.String()
		// Clear existing data
		clearUserData(db, demoUserID)
	} else {
		// Create new demo user
		params := user.CreateParams{
			Email:     demoEmail,
			Password:  demoPassword,
			FirstName: "Demo",
			LastName:  "User",
			Role:      user.RoleUser,
		}
		createdUser, err := userService.Register(ctx, params)
		if err != nil {
			log.Fatalf("Failed to create demo user: %v", err)
		}
		demoUserID = createdUser.ID.String()
		log.Printf("✓ Demo user created: %s (ID: %s)\n", createdUser.Email, createdUser.ID)
	}

	// Ensure user defaults
	if err := store.EnsureUserDefaults(demoUserID); err != nil {
		log.Fatalf("Failed to ensure user defaults: %v", err)
	}

	// Update user settings with demo categories
	demoCategories := []string{
		"Food & Dining",
		"Transportation",
		"Shopping",
		"Entertainment",
		"Utilities",
		"Healthcare",
		"Education",
		"Personal Care",
	}
	if err := store.UpdateCategories(demoUserID, demoCategories); err != nil {
		log.Fatalf("Failed to update categories: %v", err)
	}
	log.Println("✓ Categories configured")

	// Set currency
	if err := store.UpdateCurrency(demoUserID, "idr"); err != nil {
		log.Fatalf("Failed to update currency: %v", err)
	}
	log.Println("✓ Currency set to IDR")

	// Add demo budgets (amounts in IDR)
	demoBudgets := []DemoBudget{
		{Category: "Food & Dining", Amount: 9000000},      // ~$600
		{Category: "Transportation", Amount: 3000000},     // ~$200
		{Category: "Shopping", Amount: 4500000},           // ~$300
		{Category: "Entertainment", Amount: 2250000},      // ~$150
		{Category: "Utilities", Amount: 3750000},          // ~$250
		{Category: "Healthcare", Amount: 1500000},         // ~$100
		{Category: "Education", Amount: 3000000},          // ~$200
		{Category: "Personal Care", Amount: 1500000},      // ~$100
	}

	for _, b := range demoBudgets {
		budget := storage.Budget{
			ID:        uuid.New().String(),
			UserID:    demoUserID,
			Category:  b.Category,
			Amount:    b.Amount,
			Currency:  "idr",
			Period:    storage.BudgetPeriodMonthly,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err := store.AddBudget(demoUserID, budget)
		if err != nil {
			log.Printf("Warning: Failed to add budget for %s: %v\n", b.Category, err)
		}
	}
	log.Printf("✓ %d budgets created\n", len(demoBudgets))

	// Add demo transactions for the current month
	demoTransactions := []DemoTransaction{
		// Food & Dining
		{Name: "Breakfast at Cafe", Amount: 187500, Category: "Food & Dining", DayOfMonth: 1, Tags: []string{"breakfast"}},
		{Name: "Grocery Shopping", Amount: 1800000, Category: "Food & Dining", DayOfMonth: 2, Tags: []string{"groceries"}},
		{Name: "Lunch with friends", Amount: 525000, Category: "Food & Dining", DayOfMonth: 5, Tags: []string{"lunch"}},
		{Name: "Pizza Night", Amount: 427500, Category: "Food & Dining", DayOfMonth: 8, Tags: []string{"dinner"}},
		{Name: "Grocery Shopping", Amount: 1425000, Category: "Food & Dining", DayOfMonth: 10, Tags: []string{"groceries"}},
		{Name: "Coffee with colleague", Amount: 120000, Category: "Food & Dining", DayOfMonth: 12, Tags: []string{"coffee"}},
		{Name: "Dinner at restaurant", Amount: 975000, Category: "Food & Dining", DayOfMonth: 15, Tags: []string{"dinner"}},
		{Name: "Grocery Shopping", Amount: 1650000, Category: "Food & Dining", DayOfMonth: 18, Tags: []string{"groceries"}},
		{Name: "Brunch", Amount: 630000, Category: "Food & Dining", DayOfMonth: 22, Tags: []string{"brunch"}},
		{Name: "Fast food", Amount: 232500, Category: "Food & Dining", DayOfMonth: 25, Tags: []string{"fast-food"}},

		// Transportation
		{Name: "Uber ride", Amount: 270000, Category: "Transportation", DayOfMonth: 3, Tags: []string{"uber"}},
		{Name: "Gas", Amount: 975000, Category: "Transportation", DayOfMonth: 7, Tags: []string{"fuel"}},
		{Name: "Bus pass - Monthly", Amount: 1200000, Category: "Transportation", DayOfMonth: 1, Tags: []string{"transit"}},
		{Name: "Taxi ride", Amount: 330000, Category: "Transportation", DayOfMonth: 14, Tags: []string{"taxi"}},

		// Shopping
		{Name: "Clothes", Amount: 1275000, Category: "Shopping", DayOfMonth: 4, Tags: []string{"fashion"}},
		{Name: "Books", Amount: 675000, Category: "Shopping", DayOfMonth: 9, Tags: []string{"books"}},
		{Name: "Electronics store", Amount: 1800000, Category: "Shopping", DayOfMonth: 16, Tags: []string{"tech"}},
		{Name: "Home items", Amount: 900000, Category: "Shopping", DayOfMonth: 20, Tags: []string{"home"}},

		// Entertainment
		{Name: "Movie tickets", Amount: 450000, Category: "Entertainment", DayOfMonth: 6, Tags: []string{"movies"}},
		{Name: "Streaming subscription", Amount: 239850, Category: "Entertainment", DayOfMonth: 1, Tags: []string{"subscription"}},
		{Name: "Concert ticket", Amount: 1125000, Category: "Entertainment", DayOfMonth: 19, Tags: []string{"concerts"}},

		// Utilities
		{Name: "Electricity bill", Amount: 1800000, Category: "Utilities", DayOfMonth: 5, Tags: []string{"bills"}},
		{Name: "Water bill", Amount: 675000, Category: "Utilities", DayOfMonth: 10, Tags: []string{"bills"}},
		{Name: "Internet", Amount: 1275000, Category: "Utilities", DayOfMonth: 1, Tags: []string{"subscription"}},

		// Healthcare
		{Name: "Pharmacy", Amount: 525000, Category: "Healthcare", DayOfMonth: 11, Tags: []string{"medication"}},
		{Name: "Doctor visit", Amount: 2250000, Category: "Healthcare", DayOfMonth: 17, Tags: []string{"medical"}},

		// Education
		{Name: "Online course", Amount: 1485000, Category: "Education", DayOfMonth: 2, Tags: []string{"learning"}},
		{Name: "Textbook", Amount: 1275000, Category: "Education", DayOfMonth: 13, Tags: []string{"books"}},

		// Personal Care
		{Name: "Haircut", Amount: 525000, Category: "Personal Care", DayOfMonth: 21, Tags: []string{"salon"}},
		{Name: "Gym membership", Amount: 900000, Category: "Personal Care", DayOfMonth: 1, Tags: []string{"fitness"}},
	}

	// Get current month start
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	for _, trans := range demoTransactions {
		// Create date with the specified day of month
		expenseDate := time.Date(currentMonth.Year(), currentMonth.Month(), trans.DayOfMonth, 
			12, 0, 0, 0, time.UTC)
		
		expense := storage.Expense{
			ID:       uuid.New().String(),
			UserID:   demoUserID,
			Name:     trans.Name,
			Amount:   trans.Amount,
			Category: trans.Category,
			Currency: "idr",
			Date:     expenseDate,
			Tags:     trans.Tags,
		}
		
		if err := store.AddExpense(demoUserID, expense); err != nil {
			log.Printf("Warning: Failed to add expense: %v\n", err)
		}
	}
	log.Printf("✓ %d demo transactions created for %s\n", len(demoTransactions), currentMonth.Format("January 2006"))

	// Prepare demo data file
	demoData := DemoData{
		Budgets:      demoBudgets,
		Transactions: demoTransactions,
		GeneratedAt:  time.Now(),
		Note: `This file contains the demo data template. To use for the next month:
1. Update the "generatedAt" field to the new month
2. The "dayOfMonth" field in transactions will be used as-is
3. Simply re-run the seed script and it will use these dates for the new month`,
	}

	demoData.Credentials.Email = demoEmail
	demoData.Credentials.Password = demoPassword
	demoData.Credentials.UserID = demoUserID

	// Save demo data to file
	demoDataJSON, err := json.MarshalIndent(demoData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal demo data: %v", err)
	}

	demoDataFile := "scripts/demo_data.json"
	if err := os.WriteFile(demoDataFile, demoDataJSON, 0644); err != nil {
		log.Fatalf("Failed to write demo data file: %v", err)
	}
	log.Printf("✓ Demo data saved to %s\n", demoDataFile)

	// Save credentials to separate secure file
	credentialsData := map[string]string{
		"email":  demoEmail,
		"password": demoPassword,
		"userID": demoUserID,
	}
	credentialsJSON, err := json.MarshalIndent(credentialsData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal credentials: %v", err)
	}

	credentialsFile := "scripts/demo_credentials.json"
	if err := os.WriteFile(credentialsFile, credentialsJSON, 0600); err != nil {
		log.Fatalf("Failed to write credentials file: %v", err)
	}
	log.Printf("✓ Demo credentials saved to %s (with restricted permissions)\n", credentialsFile)

	separator := strings.Repeat("=", 60)
	fmt.Println("\n" + separator)
	fmt.Println("DEMO DATA SETUP COMPLETE")
	fmt.Println(separator)
	fmt.Printf("Email:    %s\n", demoEmail)
	fmt.Printf("Password: %s\n", demoPassword)
	fmt.Printf("User ID:  %s\n", demoUserID)
	fmt.Println("\nCredentials saved to: scripts/demo_credentials.json")
	fmt.Println("Transaction template saved to: scripts/demo_data.json")
	fmt.Println(separator + "\n")
}

// clearUserData removes all existing data for a user (expenses and budgets)
func clearUserData(db *sql.DB, userID string) {
	ctx := context.Background()

	// Delete expenses
	_, err := db.ExecContext(ctx, "DELETE FROM expenses WHERE user_id = $1::uuid", userID)
	if err != nil {
		log.Printf("Warning: Failed to clear expenses: %v\n", err)
	}

	// Delete budget overrides and adjustments
	_, err = db.ExecContext(ctx, "DELETE FROM budget_overrides WHERE user_id = $1::uuid", userID)
	if err != nil {
		log.Printf("Warning: Failed to clear budget overrides: %v\n", err)
	}
	_, err = db.ExecContext(ctx, "DELETE FROM budget_adjustments WHERE user_id = $1::uuid", userID)
	if err != nil {
		log.Printf("Warning: Failed to clear budget adjustments: %v\n", err)
	}

	// Delete budgets
	_, err = db.ExecContext(ctx, "DELETE FROM budgets WHERE user_id = $1::uuid", userID)
	if err != nil {
		log.Printf("Warning: Failed to clear budgets: %v\n", err)
	}

	log.Println("✓ Existing data cleared")
}

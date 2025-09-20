package main

import (
	"log"
	"net/http"

	"github.com/tanq16/expenseowl/internal/api"
	"github.com/tanq16/expenseowl/internal/auth"
	"github.com/tanq16/expenseowl/internal/storage"
	"github.com/tanq16/expenseowl/internal/web"
)

var version = "dev"

func runServer() {
	storage, err := storage.InitializeStorage()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()
	authManager, err := auth.NewManagerFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize authentication: %v", err)
	}
	handler := api.NewHandler(storage, authManager)

	// Version Handler
	http.HandleFunc("/version", handler.RequireAPIAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(version))
	}))

	http.HandleFunc("/login", handler.Login)
	http.HandleFunc("/logout", handler.Logout)

	// Static assets for SPA
	http.HandleFunc("/assets/", web.ServeAsset)
	http.HandleFunc("/manifest.json", web.ServeAsset)
	http.HandleFunc("/sw.js", web.ServeAsset)
	http.HandleFunc("/logo.png", web.ServeAsset)
	http.HandleFunc("/favicon.ico", web.ServeAsset)
	http.HandleFunc("/fa.min.css", web.ServeAsset)
	http.HandleFunc("/webfonts/", web.ServeAsset)
	http.HandleFunc("/pwa/", web.ServeAsset)

	// SPA entry (requires authentication)
	http.HandleFunc("/", handler.RequireWebAuth(web.ServeSPA))

	// Config
	http.HandleFunc("/config", handler.RequireAPIAuth(handler.GetConfig))
	http.HandleFunc("/categories", handler.RequireAPIAuth(handler.GetCategories))
	http.HandleFunc("/categories/edit", handler.RequireAPIAuth(handler.UpdateCategories))
	http.HandleFunc("/currency", handler.RequireAPIAuth(handler.GetCurrency))
	http.HandleFunc("/currency/edit", handler.RequireAPIAuth(handler.UpdateCurrency))
	http.HandleFunc("/startdate", handler.RequireAPIAuth(handler.GetStartDate))
	http.HandleFunc("/startdate/edit", handler.RequireAPIAuth(handler.UpdateStartDate))
	// http.HandleFunc("/tags", handler.GetTags)
	// http.HandleFunc("/tags/edit", handler.UpdateTags)

	// Expenses
	http.HandleFunc("/expense", handler.RequireAPIAuth(handler.AddExpense))                     // PUT for add
	http.HandleFunc("/expenses", handler.RequireAPIAuth(handler.GetExpenses))                   // GET all
	http.HandleFunc("/expense/edit", handler.RequireAPIAuth(handler.EditExpense))               // PUT for edit
	http.HandleFunc("/expense/delete", handler.RequireAPIAuth(handler.DeleteExpense))           // DELETE for single
	http.HandleFunc("/expenses/delete", handler.RequireAPIAuth(handler.DeleteMultipleExpenses)) // DELETE for multiple

	// Recurring Expenses
	http.HandleFunc("/recurring-expense", handler.RequireAPIAuth(handler.AddRecurringExpense))           // PUT for add
	http.HandleFunc("/recurring-expenses", handler.RequireAPIAuth(handler.GetRecurringExpenses))         // GET all
	http.HandleFunc("/recurring-expense/edit", handler.RequireAPIAuth(handler.UpdateRecurringExpense))   // PUT for edit
	http.HandleFunc("/recurring-expense/delete", handler.RequireAPIAuth(handler.DeleteRecurringExpense)) // DELETE

	// Import/Export
	http.HandleFunc("/export/csv", handler.RequireAPIAuth(handler.ExportCSV))
	http.HandleFunc("/import/csv", handler.RequireAPIAuth(handler.ImportCSV))
	http.HandleFunc("/import/csvold", handler.RequireAPIAuth(handler.ImportOldCSV))

	// External API
	http.HandleFunc("/api/v1/expenses", api.Authenticate(handler.CreateExpenseHandler))

	log.Println("Starting server on port 9080...")
	if err := http.ListenAndServe(":9080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func main() {
	runServer()
}

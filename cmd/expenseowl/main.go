package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tanq16/expenseowl/internal/api"
	"github.com/tanq16/expenseowl/internal/auth"
	"github.com/tanq16/expenseowl/internal/storage"
	"github.com/tanq16/expenseowl/internal/user"
	"github.com/tanq16/expenseowl/internal/web"
)

var version = "dev"

func runServer() {
	store, err := storage.InitializeStorage()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	dbProvider, ok := store.(interface{ DB() *sql.DB })
	if !ok {
		log.Fatalf("PostgreSQL storage is required for multi-user mode")
	}
	userService := user.NewService(user.NewRepository(dbProvider.DB()))

	redisClient := newRedisClient()
	defer redisClient.Close()

	jwtManager := newJWTManager(redisClient)

	handler := api.NewHandler(store, userService, jwtManager)

	mux := http.NewServeMux()

	// Version endpoint (protected)
	mux.HandleFunc("/version", handler.RequireAPIAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(version))
	}))

	// Authentication routes
	mux.HandleFunc("/api/v1/user/signup", handler.Signup)
	mux.HandleFunc("/api/v1/user/login", handler.Login)
	mux.HandleFunc("/api/v1/user/logout", handler.Logout)
	mux.HandleFunc("/api/v1/session", handler.RequireAPIAuth(handler.Session))
	mux.HandleFunc("/api/v1/user/update_password", handler.RequireAPIAuth(handler.UpdatePassword))

	// Static assets for SPA
	mux.HandleFunc("/assets/", web.ServeAsset)
	mux.HandleFunc("/manifest.json", web.ServeAsset)
	mux.HandleFunc("/sw.js", web.ServeAsset)
	mux.HandleFunc("/logo.png", web.ServeAsset)
	mux.HandleFunc("/favicon.ico", web.ServeAsset)
	mux.HandleFunc("/fa.min.css", web.ServeAsset)
	mux.HandleFunc("/webfonts/", web.ServeAsset)
	mux.HandleFunc("/pwa/", web.ServeAsset)

	// SPA entry point
	mux.HandleFunc("/", handler.ServeSPA)

	// Config
	mux.HandleFunc("/config", handler.RequireAPIAuth(handler.GetConfig))
	mux.HandleFunc("/categories", handler.RequireAPIAuth(handler.GetCategories))
	mux.HandleFunc("/categories/edit", handler.RequireAPIAuth(handler.UpdateCategories))
	mux.HandleFunc("/currency", handler.RequireAPIAuth(handler.GetCurrency))
	mux.HandleFunc("/currency/edit", handler.RequireAPIAuth(handler.UpdateCurrency))
	mux.HandleFunc("/startdate", handler.RequireAPIAuth(handler.GetStartDate))
	mux.HandleFunc("/startdate/edit", handler.RequireAPIAuth(handler.UpdateStartDate))

	// Expenses
	mux.HandleFunc("/expense", handler.RequireAPIAuth(handler.AddExpense))
	mux.HandleFunc("/expenses", handler.RequireAPIAuth(handler.GetExpenses))
	mux.HandleFunc("/expense/edit", handler.RequireAPIAuth(handler.EditExpense))
	mux.HandleFunc("/expense/delete", handler.RequireAPIAuth(handler.DeleteExpense))
	mux.HandleFunc("/expenses/delete", handler.RequireAPIAuth(handler.DeleteMultipleExpenses))

	// Recurring Expenses
	mux.HandleFunc("/recurring-expense", handler.RequireAPIAuth(handler.AddRecurringExpense))
	mux.HandleFunc("/recurring-expenses", handler.RequireAPIAuth(handler.GetRecurringExpenses))
	mux.HandleFunc("/recurring-expense/edit", handler.RequireAPIAuth(handler.UpdateRecurringExpense))
	mux.HandleFunc("/recurring-expense/delete", handler.RequireAPIAuth(handler.DeleteRecurringExpense))

	// Import/Export
	mux.HandleFunc("/export/csv", handler.RequireAPIAuth(handler.ExportCSV))
	mux.HandleFunc("/import/csv", handler.RequireAPIAuth(handler.ImportCSV))
	mux.HandleFunc("/import/csvold", handler.RequireAPIAuth(handler.ImportOldCSV))

	// External API
	mux.HandleFunc("/api/v1/expenses", api.Authenticate(handler.CreateExpenseHandler))

	server := &http.Server{
		Addr:    ":9080",
		Handler: mux,
	}

	log.Println("Starting server on port 9080...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func newRedisClient() *redis.Client {
	addr := fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", "localhost"), getEnv("REDIS_PORT", "6380"))
	db, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	return client
}

func newJWTManager(client *redis.Client) *auth.JWTManager {
	secret := os.Getenv("JWT_SECRET")
	expiryHours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil || expiryHours <= 0 {
		expiryHours = 24
	}
	manager, err := auth.NewJWTManager(secret, time.Duration(expiryHours)*time.Hour, client)
	if err != nil {
		log.Fatalf("Failed to initialize JWT manager: %v", err)
	}
	return manager
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	runServer()
}

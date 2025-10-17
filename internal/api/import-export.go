package api

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tanq16/expenseowl/internal/storage"
)

// ExportCSV streams a CSV export of expenses, optionally filtered by period.
func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}

	manager, err := h.encryptionManagerFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	expenses, err := h.storage.GetAllExpenses(userCtx.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve expenses"})
		log.Printf("API ERROR: Failed to retrieve expenses for CSV export: %v\n", err)
		return
	}

	if manager != nil {
		for i := range expenses {
			if err := decryptExpense(manager, &expenses[i]); err != nil {
				log.Printf("API ERROR: Failed to decrypt expense %s for CSV export: %v\n", expenses[i].ID, err)
			}
		}
	}

	expenses, err = h.filterExpensesForExport(r, userCtx.ID, expenses)
	if err != nil {
		switch err.(type) {
		case validationError:
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=expenses.csv")
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	headers := []string{"ID", "Name", "Category", "Amount", "Date", "Tags"}
	if err := writer.Write(headers); err != nil {
		log.Printf("API ERROR: Failed to write CSV header: %v\n", err)
		return
	}

	// Write records
	for _, expense := range expenses {
		record := []string{
			expense.ID,
			expense.Name,
			expense.Category,
			strconv.FormatFloat(expense.Amount, 'f', 2, 64),
			expense.Date.Format(time.RFC3339),
			strings.Join(expense.Tags, ","),
		}
		if err := writer.Write(record); err != nil {
			log.Printf("API ERROR: Failed to write CSV record for expense ID %s: %v\n", expense.ID, err)
			continue
		}
	}
	log.Println("HTTP: Exported expenses to CSV")
}

type validationError struct {
	message string
}

func (e validationError) Error() string {
	return e.message
}

func (h *Handler) filterExpensesForExport(r *http.Request, userID string, expenses []storage.Expense) ([]storage.Expense, error) {
	period := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("period")))
	if period == "" {
		period = "month"
	}

	now := time.Now().In(time.Local)
	loc := now.Location()

	var (
		start       time.Time
		end         time.Time
		applyFilter bool
	)

	switch period {
	case "all":
		applyFilter = false
	case "today":
		start, end = dayBounds(now, loc, 0)
		applyFilter = true
	case "yesterday":
		start, end = dayBounds(now, loc, -1)
		applyFilter = true
	case "week":
		start, end = weekBounds(now, loc)
		applyFilter = true
	case "range":
		startParam := strings.TrimSpace(r.URL.Query().Get("start"))
		endParam := strings.TrimSpace(r.URL.Query().Get("end"))
		if startParam == "" || endParam == "" {
			return nil, validationError{message: "start and end parameters are required for range exports"}
		}
		var err error
		start, end, err = customRangeBounds(startParam, endParam, loc)
		if err != nil {
			return nil, validationError{message: err.Error()}
		}
		applyFilter = true
	case "month":
		fallthrough
	default:
		if period != "month" {
			return nil, validationError{message: "invalid period parameter"}
		}
		startDateValue, err := h.storage.GetStartDate(userID)
		if err != nil {
			log.Printf("API ERROR: Failed to resolve start date for CSV export: %v\n", err)
			return nil, fmt.Errorf("failed to resolve budgeting start date")
		}
		endOfMonthValue, err := h.storage.GetEndOfMonth(userID)
		if err != nil {
			log.Printf("API ERROR: Failed to resolve end-of-month preference for CSV export: %v\n", err)
			return nil, fmt.Errorf("failed to resolve end-of-month preference")
		}
		start, end = monthBounds(now, loc, startDateValue, endOfMonthValue)
		applyFilter = true
	}

	if !applyFilter {
		return expenses, nil
	}

	filtered := make([]storage.Expense, 0, len(expenses))
	for _, expense := range expenses {
		if expense.Date.IsZero() {
			continue
		}
		if expense.Date.Before(start) || expense.Date.After(end) {
			continue
		}
		filtered = append(filtered, expense)
	}
	return filtered, nil
}

func dayBounds(base time.Time, loc *time.Location, offset int) (time.Time, time.Time) {
	local := base.In(loc).AddDate(0, 0, offset)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	end := start.Add(24*time.Hour - time.Nanosecond)
	return start, end
}

func weekBounds(base time.Time, loc *time.Location) (time.Time, time.Time) {
	local := base.In(loc)
	weekday := int(local.Weekday())
	mondayIndex := (weekday + 6) % 7
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	start = start.AddDate(0, 0, -mondayIndex)
	end := start.AddDate(0, 0, 6).Add(24*time.Hour - time.Nanosecond)
	return start, end
}

func customRangeBounds(startStr, endStr string, loc *time.Location) (time.Time, time.Time, error) {
	const layout = "2006-01-02"
	startDate, err := time.ParseInLocation(layout, startStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date; use YYYY-MM-DD")
	}
	endDate, err := time.ParseInLocation(layout, endStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date; use YYYY-MM-DD")
	}
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, loc).Add(24*time.Hour - time.Nanosecond)
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("start date must be before end date")
	}
	return start, end, nil
}

func monthBounds(base time.Time, loc *time.Location, startDate int, endOfMonth bool) (time.Time, time.Time) {
	local := base.In(loc)

	if endOfMonth {
		start := time.Date(local.Year(), local.Month(), 0, 0, 0, 0, 0, loc)
		end := time.Date(local.Year(), local.Month()+1, 0, 0, 0, 0, 0, loc)
		end = end.AddDate(0, 0, -1).Add(24*time.Hour - time.Nanosecond)
		return start, end
	}

	if startDate <= 1 {
		start := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc)
		end := time.Date(local.Year(), local.Month()+1, 1, 0, 0, 0, 0, loc).Add(-time.Nanosecond)
		return start, end
	}

	thisMonthStart := startDate
	currentMonthDays := daysInMonth(local.Year(), local.Month(), loc)
	if thisMonthStart > currentMonthDays {
		thisMonthStart = currentMonthDays
	}

	prevYear, prevMonth := local.Year(), local.Month()-1
	if prevMonth < time.January {
		prevMonth = time.December
		prevYear--
	}
	prevMonthStart := startDate
	prevMonthDays := daysInMonth(prevYear, prevMonth, loc)
	if prevMonthStart > prevMonthDays {
		prevMonthStart = prevMonthDays
	}

	if local.Day() < thisMonthStart {
		start := time.Date(prevYear, prevMonth, prevMonthStart, 0, 0, 0, 0, loc)
		end := time.Date(local.Year(), local.Month(), thisMonthStart, 0, 0, 0, 0, loc).Add(-time.Nanosecond)
		return start, end
	}

	nextYear, nextMonth := local.Year(), local.Month()+1
	if nextMonth > time.December {
		nextMonth = time.January
		nextYear++
	}
	nextMonthStart := startDate
	nextMonthDays := daysInMonth(nextYear, nextMonth, loc)
	if nextMonthStart > nextMonthDays {
		nextMonthStart = nextMonthDays
	}

	start := time.Date(local.Year(), local.Month(), thisMonthStart, 0, 0, 0, 0, loc)
	end := time.Date(nextYear, nextMonth, nextMonthStart, 0, 0, 0, 0, loc).Add(-time.Nanosecond)
	return start, end
}

func daysInMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

// imports expenses from CSV
func (h *Handler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	manager, err := h.encryptionManagerFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max file size
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Could not parse multipart form"})
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Error retrieving the file"})
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Failed to read CSV file"})
		return
	}
	if len(records) < 2 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "CSV file must have a header and at least one data row"})
		return
	}

	header := records[0]
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[strings.ToLower(strings.TrimSpace(col))] = i
	}
	// Check for mandatory columns
	requiredCols := []string{"name", "category", "amount", "date"}
	for _, col := range requiredCols {
		if _, ok := colMap[col]; !ok {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Missing required column: %s", col)})
			return
		}
	}
	// Get optional column indices
	idIdx, idExists := colMap["id"]
	tagsIdx, tagsExists := colMap["tags"]
	currencyIdx, currencyExists := colMap["currency"]

	currentCategories, err := h.storage.GetCategories(userCtx.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Could not retrieve current categories"})
		return
	}
	categorySet := make(map[string]bool)
	for _, cat := range currentCategories {
		categorySet[strings.ToLower(cat)] = true
	}
	var newCategories []string
	var importedCount, skippedCount int
	// TODO: might be worth setting default currency when we have currency updation behavior
	currencyVal, err := h.storage.GetCurrency(userCtx.ID)
	if err != nil {
		log.Printf("Error: Could not retrieve currency, shutting down import: %v\n", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Could not retrieve currency"})
		return
	}

	for i, record := range records[1:] {
		if len(record) != len(header) {
			log.Printf("Warning: Skipping row %d due to incorrect column count\n", i+2)
			skippedCount++
			continue
		}

		// Check if expense exists by ID, if provided - without doing a clash resolution
		if idExists {
			id := record[idIdx]
			if _, err := h.storage.GetExpense(userCtx.ID, id); err == nil {
				log.Printf("Info: Skipping row %d because expense with ID '%s' already exists\n", i+2, id)
				skippedCount++
				continue
			}
		}

		// Check for currency field, if provided - default is retrieved
		localCurrency := currencyVal
		if currencyExists {
			currency := record[currencyIdx]
			if !slices.Contains(storage.SupportedCurrencies, currency) {
				log.Printf("Warning: Skipping row %d due to invalid currency: %s\n", i+2, currency)
				skippedCount++
				continue
			}
			localCurrency = strings.TrimSpace(currency)
		}

		amount, err := strconv.ParseFloat(record[colMap["amount"]], 64)
		if err != nil {
			log.Printf("Warning: Skipping row %d due to invalid amount: %s\n", i+2, record[colMap["amount"]])
			skippedCount++
			continue
		}
		date, err := parseDate(record[colMap["date"]])
		if err != nil {
			log.Printf("Warning: Skipping row %d due to invalid date: %v\n", i+2, err)
			skippedCount++
			continue
		}
		category := strings.TrimSpace(record[colMap["category"]])
		if _, ok := categorySet[strings.ToLower(category)]; !ok {
			newCategories = append(newCategories, category)
			categorySet[strings.ToLower(category)] = true // Add to set to handle duplicates in the same file
		}
		var tags []string
		if tagsExists {
			tagsStr := record[tagsIdx]
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
				for i := range tags {
					tags[i] = strings.TrimSpace(tags[i])
				}
			}
		}

		expense := storage.Expense{
			Name:     strings.TrimSpace(record[colMap["name"]]),
			Category: category,
			Amount:   amount,
			Currency: localCurrency,
			Date:     date,
			Tags:     tags,
		}
		if err := expense.Validate(); err != nil {
			log.Printf("Warning: Skipping row %d due to validation error: %v\n", i+2, err)
			skippedCount++
			continue
		}
		if err := ensureExpenseBlob(manager, &expense); err != nil {
			log.Printf("Warning: Skipping row %d due to encryption error: %v\n", i+2, err)
			skippedCount++
			continue
		}
		if err := h.storage.AddExpense(userCtx.ID, expense); err != nil {
			log.Printf("Error: Could not add expense from row %d: %v\n", i+2, err)
			skippedCount++
			continue
		}
		importedCount++
		time.Sleep(10 * time.Millisecond) // Throttle to reduce storage overhead
	}

	if len(newCategories) > 0 {
		if err := h.storage.UpdateCategories(userCtx.ID, append(currentCategories, newCategories...)); err != nil {
			log.Printf("Warning: Failed to add new categories to config: %v\n", err)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":          "success",
		"total_processed": len(records) - 1,
		"imported":        importedCount,
		"skipped":         skippedCount,
		"new_categories":  newCategories,
	})
	log.Printf("HTTP: Imported %d expenses from CSV file. Skipped %d records.", importedCount, skippedCount)
}

// handles importing from ExpenseOwl < v4.0
// TODO: remove this in the future
func (h *Handler) ImportOldCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}
	userCtx, err := h.userFromRequest(r)
	if err != nil {
		unauthorized(w)
		return
	}
	manager, err := h.encryptionManagerFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max file size
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Could not parse multipart form"})
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Error retrieving the file"})
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Failed to read CSV file"})
		return
	}
	if len(records) < 2 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "CSV file must have a header and at least one data row"})
		return
	}

	header := records[0]
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[strings.ToLower(strings.TrimSpace(col))] = i
	}
	requiredCols := []string{"name", "category", "amount", "date"}
	for _, col := range requiredCols {
		if _, ok := colMap[col]; !ok {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Missing required column: %s", col)})
			return
		}
	}

	currentCategories, err := h.storage.GetCategories(userCtx.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Could not retrieve current categories"})
		return
	}
	categorySet := make(map[string]bool)
	for _, cat := range currentCategories {
		categorySet[strings.ToLower(cat)] = true
	}
	var newCategories []string
	var importedCount, skippedCount int

	for i, record := range records[1:] {
		if len(record) != len(header) {
			log.Printf("Warning: Skipping row %d due to incorrect column count\n", i+2)
			skippedCount++
			continue
		}
		amount, err := strconv.ParseFloat(record[colMap["amount"]], 64)
		if err != nil {
			log.Printf("Warning: Skipping row %d due to invalid amount: %s\n", i+2, record[colMap["amount"]])
			skippedCount++
			continue
		}
		date, err := parseDate(record[colMap["date"]])
		if err != nil {
			log.Printf("Warning: Skipping row %d due to invalid date: %v\n", i+2, err)
			skippedCount++
			continue
		}
		category := strings.TrimSpace(record[colMap["category"]])
		if _, ok := categorySet[strings.ToLower(category)]; !ok {
			newCategories = append(newCategories, category)
			categorySet[strings.ToLower(category)] = true // Add to set to handle duplicates in the same file
		}

		// switches sign for new expenseowl
		amountUpdated := amount
		if category != "Income" {
			amountUpdated = amount * -1
		}
		expense := storage.Expense{
			Name:     strings.TrimSpace(record[colMap["name"]]),
			Category: category,
			Amount:   amountUpdated,
			Date:     date,
		}
		if err := expense.Validate(); err != nil {
			log.Printf("Warning: Skipping row %d due to validation error: %v\n", i+2, err)
			skippedCount++
			continue
		}
		if err := ensureExpenseBlob(manager, &expense); err != nil {
			log.Printf("Warning: Skipping row %d due to encryption error: %v\n", i+2, err)
			skippedCount++
			continue
		}
		if err := h.storage.AddExpense(userCtx.ID, expense); err != nil {
			log.Printf("Error: Could not add expense from row %d: %v\n", i+2, err)
			skippedCount++
			continue
		}
		importedCount++
		time.Sleep(10 * time.Millisecond)
	}

	if len(newCategories) > 0 {
		if err := h.storage.UpdateCategories(userCtx.ID, append(currentCategories, newCategories...)); err != nil {
			log.Printf("Warning: Failed to add new categories to config: %v\n", err)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":          "success",
		"total_processed": len(records) - 1,
		"imported":        importedCount,
		"skipped":         skippedCount,
		"new_categories":  newCategories,
	})
	log.Printf("HTTP: Imported %d expenses from CSV file. Skipped %d records.", importedCount, skippedCount)
}

func parseDate(dateStr string) (time.Time, error) {
	dateFormats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006-1-2",
		"2006/01/02",
		"2006/1/2",
	}
	for _, format := range dateFormats {
		if d, err := time.Parse(format, dateStr); err == nil {
			return d.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

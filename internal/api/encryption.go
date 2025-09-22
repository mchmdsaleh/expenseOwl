package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"

    "github.com/tanq16/expenseowl/internal/encryption"
    "github.com/tanq16/expenseowl/internal/storage"
)

const encryptionHeader = "X-Encryption-Key"

func (h *Handler) encryptionManagerFromRequest(r *http.Request) (*encryption.Manager, error) {
	cipher := strings.TrimSpace(r.Header.Get(encryptionHeader))
	if cipher == "" {
		return nil, nil
	}
	manager, err := encryption.NewManagerFromCipher(cipher)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key: %w", err)
	}
	return manager, nil
}

func decryptExpense(manager *encryption.Manager, expense *storage.Expense) error {
    if expense == nil || expense.Blob == "" {
        return nil
    }
    // If no manager is provided, attempt to parse plaintext JSON blobs.
    if manager == nil {
        var plain storage.Expense
        if err := json.Unmarshal([]byte(expense.Blob), &plain); err != nil {
            return fmt.Errorf("encrypted expense provided without %s header", encryptionHeader)
        }
        if plain.ID == "" {
            plain.ID = expense.ID
        }
        if plain.UserID == "" {
            plain.UserID = expense.UserID
        }
        plain.Blob = expense.Blob
        *expense = plain
        return nil
    }
    var decrypted storage.Expense
    if err := manager.Decrypt(expense.Blob, &decrypted); err != nil {
        // Fallback: try to parse as plaintext JSON if decryption fails.
        var plain storage.Expense
        if jsonErr := json.Unmarshal([]byte(expense.Blob), &plain); jsonErr != nil {
            return fmt.Errorf("failed to decrypt expense: %w", err)
        }
        if plain.ID == "" {
            plain.ID = expense.ID
        }
        if plain.UserID == "" {
            plain.UserID = expense.UserID
        }
        plain.Blob = expense.Blob
        *expense = plain
        return nil
    }
    if decrypted.ID == "" {
        decrypted.ID = expense.ID
    }
    if decrypted.UserID == "" {
        decrypted.UserID = expense.UserID
    }
    decrypted.Blob = expense.Blob
    *expense = decrypted
    return nil
}

func ensureExpenseBlob(manager *encryption.Manager, expense *storage.Expense) error {
    if expense == nil || expense.Blob != "" {
        return nil
    }
    payload := *expense
    payload.Blob = ""
    if manager != nil {
        blob, err := manager.Encrypt(payload)
        if err != nil {
            return fmt.Errorf("failed to encrypt expense: %w", err)
        }
        expense.Blob = blob
        return nil
    }
    // No encryption: store plaintext JSON blob
    raw, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to serialize expense: %w", err)
    }
    expense.Blob = string(raw)
    return nil
}

func decryptRecurring(manager *encryption.Manager, recurring *storage.RecurringExpense) error {
    if recurring == nil || recurring.Blob == "" {
        return nil
    }
    if manager == nil {
        var plain storage.RecurringExpense
        if err := json.Unmarshal([]byte(recurring.Blob), &plain); err != nil {
            return fmt.Errorf("encrypted recurring expense provided without %s header", encryptionHeader)
        }
        if plain.ID == "" {
            plain.ID = recurring.ID
        }
        if plain.UserID == "" {
            plain.UserID = recurring.UserID
        }
        plain.Blob = recurring.Blob
        *recurring = plain
        return nil
    }
    var decrypted storage.RecurringExpense
    if err := manager.Decrypt(recurring.Blob, &decrypted); err != nil {
        // Fallback to plaintext JSON if decryption fails
        var plain storage.RecurringExpense
        if jsonErr := json.Unmarshal([]byte(recurring.Blob), &plain); jsonErr != nil {
            return fmt.Errorf("failed to decrypt recurring expense: %w", err)
        }
        if plain.ID == "" {
            plain.ID = recurring.ID
        }
        if plain.UserID == "" {
            plain.UserID = recurring.UserID
        }
        plain.Blob = recurring.Blob
        *recurring = plain
        return nil
    }
    if decrypted.ID == "" {
        decrypted.ID = recurring.ID
    }
    if decrypted.UserID == "" {
        decrypted.UserID = recurring.UserID
    }
    decrypted.Blob = recurring.Blob
    *recurring = decrypted
    return nil
}

func ensureRecurringBlob(manager *encryption.Manager, recurring *storage.RecurringExpense) error {
    if recurring == nil || recurring.Blob != "" {
        return nil
    }
    payload := *recurring
    payload.Blob = ""
    if manager != nil {
        blob, err := manager.Encrypt(payload)
        if err != nil {
            return fmt.Errorf("failed to encrypt recurring expense: %w", err)
        }
        recurring.Blob = blob
        return nil
    }
    // No encryption: store plaintext JSON blob
    raw, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to serialize recurring expense: %w", err)
    }
    recurring.Blob = string(raw)
    return nil
}

package api

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"

    "github.com/tanq16/expenseowl/internal/encryption"
    "github.com/tanq16/expenseowl/internal/storage"
)

const encryptionHeader = "X-Encryption-Key"

// serverEncryptionManager returns a manager based on a server-side key.
// Set EXTERNAL_ENCRYPTION_KEY to enable automatic at-rest encryption for
// inbound integrations (e.g., n8n) without requiring the client key.
func serverEncryptionManager() (*encryption.Manager, error) {
    key := strings.TrimSpace(os.Getenv("EXTERNAL_ENCRYPTION_KEY"))
    if key == "" {
        return nil, nil
    }
    return encryption.NewManagerFromCipher(key)
}

// deriveExternalCipher returns a hex(SHA-256(EXTERNAL_ENCRYPTION_KEY + ":" + userID)).
// Returns empty string if the base key is not configured or userID is empty.
func deriveExternalCipher(userID string) string {
    base := strings.TrimSpace(os.Getenv("EXTERNAL_ENCRYPTION_KEY"))
    uid := strings.TrimSpace(userID)
    if base == "" || uid == "" {
        return ""
    }
    h := sha256.Sum256([]byte(base + ":" + uid))
    return hex.EncodeToString(h[:])
}

// externalManagerForUser returns an encryption manager derived per-user when configured.
func externalManagerForUser(userID string) (*encryption.Manager, error) {
    cipher := deriveExternalCipher(userID)
    if cipher == "" {
        return nil, nil
    }
    return encryption.NewManagerFromCipher(cipher)
}

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
    // Try in order: user manager, per-user derived manager, server fallback, plaintext JSON
    var decrypted storage.Expense
    if manager != nil {
        if err := manager.Decrypt(expense.Blob, &decrypted); err == nil {
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
    }
    if ext, _ := externalManagerForUser(expense.UserID); ext != nil {
        if err := ext.Decrypt(expense.Blob, &decrypted); err == nil {
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
    }
    if srv, _ := serverEncryptionManager(); srv != nil {
        if err := srv.Decrypt(expense.Blob, &decrypted); err == nil {
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
    }
    // Fallback: parse as plaintext JSON
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

func ensureExpenseBlob(manager *encryption.Manager, expense *storage.Expense) error {
    if expense == nil || expense.Blob != "" {
        return nil
    }
    payload := *expense
    payload.Blob = ""
    // Prefer client-supplied key, else per-user derived key, else server key, else plaintext
    if manager != nil {
        blob, err := manager.Encrypt(payload)
        if err != nil {
            return fmt.Errorf("failed to encrypt expense: %w", err)
        }
        expense.Blob = blob
        return nil
    }
    if ext, _ := externalManagerForUser(expense.UserID); ext != nil {
        blob, err := ext.Encrypt(payload)
        if err != nil {
            return fmt.Errorf("failed to encrypt expense: %w", err)
        }
        expense.Blob = blob
        return nil
    }
    if srv, _ := serverEncryptionManager(); srv != nil {
        blob, err := srv.Encrypt(payload)
        if err != nil {
            return fmt.Errorf("failed to encrypt expense: %w", err)
        }
        expense.Blob = blob
        return nil
    }
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
    // Try: user manager, per-user derived manager, server manager, plaintext
    var decrypted storage.RecurringExpense
    if manager != nil {
        if err := manager.Decrypt(recurring.Blob, &decrypted); err == nil {
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
    }
    if ext, _ := externalManagerForUser(recurring.UserID); ext != nil {
        if err := ext.Decrypt(recurring.Blob, &decrypted); err == nil {
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
    }
    if srv, _ := serverEncryptionManager(); srv != nil {
        if err := srv.Decrypt(recurring.Blob, &decrypted); err == nil {
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
    }
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
    if ext, _ := externalManagerForUser(recurring.UserID); ext != nil {
        blob, err := ext.Encrypt(payload)
        if err != nil {
            return fmt.Errorf("failed to encrypt recurring expense: %w", err)
        }
        recurring.Blob = blob
        return nil
    }
    if srv, _ := serverEncryptionManager(); srv != nil {
        blob, err := srv.Encrypt(payload)
        if err != nil {
            return fmt.Errorf("failed to encrypt recurring expense: %w", err)
        }
        recurring.Blob = blob
        return nil
    }
    raw, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to serialize recurring expense: %w", err)
    }
    recurring.Blob = string(raw)
    return nil
}

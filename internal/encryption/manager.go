package encryption

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"strings"

	jose "gopkg.in/square/go-jose.v2"
)

var (
	ErrMissingCipher = errors.New("encryption key missing")
)

type Manager struct {
	key []byte
}

func NewManagerFromCipher(cipher string) (*Manager, error) {
	cipher = strings.TrimSpace(cipher)
	if cipher == "" {
		return nil, ErrMissingCipher
	}
	sum := sha256.Sum256([]byte(cipher))
	key := make([]byte, len(sum))
	copy(key, sum[:])
	return &Manager{key: key}, nil
}

func (m *Manager) Encrypt(v interface{}) (string, error) {
	if m == nil || len(m.key) == 0 {
		return "", ErrMissingCipher
	}
	payload, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	recipient := jose.Recipient{
		Algorithm: jose.A256KW,
		Key:       m.key,
	}
	enc, err := jose.NewEncrypter(jose.A128CBC_HS256, recipient, (&jose.EncrypterOptions{}).WithContentType("json"))
	if err != nil {
		return "", err
	}
	object, err := enc.Encrypt(payload)
	if err != nil {
		return "", err
	}
	return object.CompactSerialize()
}

func (m *Manager) Decrypt(blob string, v interface{}) error {
	if m == nil || len(m.key) == 0 {
		return ErrMissingCipher
	}
	blob = strings.TrimSpace(blob)
	if blob == "" {
		return errors.New("ciphertext payload is empty")
	}
	object, err := jose.ParseEncrypted(blob)
	if err != nil {
		return err
	}
	plaintext, err := object.Decrypt(m.key)
	if err != nil {
		return err
	}
	return json.Unmarshal(plaintext, v)
}

func (m *Manager) Key() []byte {
	if m == nil {
		return nil
	}
	dup := make([]byte, len(m.key))
	copy(dup, m.key)
	return dup
}

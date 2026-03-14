package vault

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/pbkdf2"

	"github.com/npc-live/openvault/internal/config"
	"github.com/npc-live/openvault/internal/crypto"
	"github.com/npc-live/openvault/internal/keychain"
	"github.com/npc-live/openvault/internal/store"
)

// Vault is the high-level orchestrator that ties together the store,
// keychain, and crypto layers.
type Vault struct {
	st  store.Store
	key []byte
}

// Init creates a new vault: generates an encryption key, stores it in the
// keychain, creates the DB directory and opens the store.
func Init(dbPath string, kc keychain.Keychain) error {
	// Ensure directory exists.
	dir := config.DefaultDir()
	if err := os.MkdirAll(dir, config.DirPerm); err != nil {
		return fmt.Errorf("create vault dir: %w", err)
	}

	// Generate encryption key.
	key, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	// Store key in keychain.
	if err := kc.SetKey(config.ServiceName, key); err != nil {
		return fmt.Errorf("store key in keychain: %w", err)
	}

	// Create the store (creates the DB file).
	st, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	return st.Close()
}

// Open opens an existing vault.
func Open(dbPath string, kc keychain.Keychain) (*Vault, error) {
	key, err := kc.GetKey(config.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("get encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, errors.New("invalid key length in keychain")
	}

	st, err := store.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	return &Vault{st: st, key: key}, nil
}

// DeriveKey derives a 32-byte AES key from password + secretKey using PBKDF2-SHA256.
func DeriveKey(password, secretKey []byte) []byte {
	return pbkdf2.Key(password, secretKey, 100000, 32, sha256.New)
}

// ReEncrypt decrypts all secrets with the current key and re-encrypts them with newKey.
// The in-memory key is updated to newKey on success.
func (v *Vault) ReEncrypt(newKey []byte) error {
	entries, err := v.st.ListEntries()
	if err != nil {
		return fmt.Errorf("list entries: %w", err)
	}
	for _, e := range entries {
		plain, err := crypto.Decrypt(v.key, e.Value)
		if err != nil {
			return fmt.Errorf("decrypt %q: %w", e.Key, err)
		}
		enc, err := crypto.Encrypt(newKey, plain)
		if err != nil {
			return fmt.Errorf("re-encrypt %q: %w", e.Key, err)
		}
		if err := v.st.SetRaw(e.Key, enc, e.UpdatedAt); err != nil {
			return fmt.Errorf("store %q: %w", e.Key, err)
		}
	}
	copy(v.key, newKey)
	return nil
}

// RawSet stores pre-encrypted bytes directly (used during sync pull).
func (v *Vault) RawSet(name string, enc []byte, ts int64) error {
	return v.st.SetRaw(name, enc, ts)
}

// ListEntries returns all encrypted entries with timestamps (used during sync push).
func (v *Vault) ListEntries() ([]store.Entry, error) {
	return v.st.ListEntries()
}

// Set encrypts and stores a secret.
func (v *Vault) Set(name string, value []byte) error {
	enc, err := crypto.Encrypt(v.key, value)
	if err != nil {
		return fmt.Errorf("encrypt %q: %w", name, err)
	}
	return v.st.Set(name, enc)
}

// Get retrieves and decrypts a secret.
func (v *Vault) Get(name string) ([]byte, error) {
	enc, err := v.st.Get(name)
	if err != nil {
		return nil, err
	}
	plain, err := crypto.Decrypt(v.key, enc)
	if err != nil {
		return nil, fmt.Errorf("decrypt %q: %w", name, err)
	}
	return plain, nil
}

// Delete removes a secret.
func (v *Vault) Delete(name string) error {
	return v.st.Delete(name)
}

// List returns all secret names.
func (v *Vault) List() ([]string, error) {
	return v.st.List()
}

// Env returns all secrets as KEY=value pairs suitable for environment injection.
func (v *Vault) Env() (map[string]string, error) {
	keys, err := v.st.List()
	if err != nil {
		return nil, err
	}
	env := make(map[string]string, len(keys))
	for _, k := range keys {
		val, err := v.Get(k)
		if err != nil {
			return nil, fmt.Errorf("get %q: %w", k, err)
		}
		env[k] = string(val)
	}
	return env, nil
}

// Close zeroes the in-memory key and closes the store.
func (v *Vault) Close() error {
	for i := range v.key {
		v.key[i] = 0
	}
	return v.st.Close()
}

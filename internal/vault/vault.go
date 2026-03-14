package vault

import (
	"errors"
	"fmt"
	"os"

	"github.com/qing/openvault/internal/config"
	"github.com/qing/openvault/internal/crypto"
	"github.com/qing/openvault/internal/keychain"
	"github.com/qing/openvault/internal/store"
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

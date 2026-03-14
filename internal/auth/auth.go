// Package auth manages the cloud sync JWT in the system keychain.
package auth

import (
	"errors"

	"github.com/npc-live/openvault/internal/keychain"
)

const tokenService = "openvault-token"

// ErrNotLoggedIn is returned when no JWT is stored.
var ErrNotLoggedIn = errors.New("not logged in — run `openvault login` first")

// GetToken retrieves the stored JWT. Returns ErrNotLoggedIn if absent.
func GetToken(kc keychain.Keychain) (string, error) {
	raw, err := kc.GetKey(tokenService)
	if err != nil {
		return "", ErrNotLoggedIn
	}
	return string(raw), nil
}

// SetToken stores a JWT in the keychain.
func SetToken(kc keychain.Keychain, token string) error {
	return kc.SetKey(tokenService, []byte(token))
}

// DeleteToken removes the JWT from the keychain (best-effort).
func DeleteToken(kc keychain.Keychain) {
	// SetKey with empty overwrites; on Darwin the -U flag handles upsert.
	// For a true delete we rely on the OS-specific implementation.
	// We use a zero-length sentinel and callers check via GetToken.
	_ = kc.SetKey(tokenService, []byte{})
}

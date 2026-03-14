// Package remote provides an HTTP client for the OpenVault cloud Worker API.
package remote

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/npc-live/openvault/internal/config"
	"github.com/npc-live/openvault/internal/store"
)

// Client talks to the Cloudflare Worker.
type Client struct {
	base  string
	token string
	http  *http.Client
}

// New creates a Client. token may be empty for unauthenticated calls.
func New(token string) *Client {
	return &Client{
		base:  config.APIBaseURL,
		token: token,
		http:  &http.Client{},
	}
}

type authResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

// Register creates an unverified account. The server sends a verification email.
// Returns the message from the server (not a token yet).
func (c *Client) Register(email, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	resp, err := c.http.Post(c.base+"/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server %d: %s", resp.StatusCode, raw)
	}
	var ar authResponse
	_ = json.Unmarshal(raw, &ar)
	return ar.Message, nil
}

// VerifyEmail confirms the 6-digit code and returns the JWT.
func (c *Client) VerifyEmail(email, code string) (string, error) {
	body, _ := json.Marshal(map[string]string{"email": email, "code": code})
	resp, err := c.http.Post(c.base+"/verify-email", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server %d: %s", resp.StatusCode, raw)
	}
	var ar authResponse
	if err := json.Unmarshal(raw, &ar); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	return ar.Token, nil
}

// Login authenticates and returns a JWT.
func (c *Client) Login(email, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	resp, err := c.http.Post(c.base+"/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server %d: %s", resp.StatusCode, raw)
	}
	var ar authResponse
	if err := json.Unmarshal(raw, &ar); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	return ar.Token, nil
}

// Logout revokes the current JWT on the server.
func (c *Client) Logout() error {
	req, _ := http.NewRequest(http.MethodPost, c.base+"/logout", nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

// ForgotPassword triggers a password reset email.
func (c *Client) ForgotPassword(email string) (string, error) {
	body, _ := json.Marshal(map[string]string{"email": email})
	resp, err := c.http.Post(c.base+"/forgot-password", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server %d: %s", resp.StatusCode, raw)
	}
	var ar authResponse
	_ = json.Unmarshal(raw, &ar)
	return ar.Message, nil
}

// ResetPassword sets a new password using the emailed code.
func (c *Client) ResetPassword(email, code, newPassword string) (string, error) {
	body, _ := json.Marshal(map[string]string{"email": email, "code": code, "new_password": newPassword})
	resp, err := c.http.Post(c.base+"/reset-password", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server %d: %s", resp.StatusCode, raw)
	}
	var ar authResponse
	_ = json.Unmarshal(raw, &ar)
	return ar.Message, nil
}

// RemoteEntry is the wire format for a secret.
type RemoteEntry struct {
	KeyName        string `json:"key_name"`
	EncryptedValue string `json:"encrypted_value"` // hex
	UpdatedAt      int64  `json:"updated_at"`
}

// GetSecrets fetches all remote secrets.
func (c *Client) GetSecrets() ([]RemoteEntry, error) {
	req, _ := http.NewRequest(http.MethodGet, c.base+"/secrets", nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server %d: %s", resp.StatusCode, msg)
	}
	var entries []RemoteEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return entries, nil
}

// PutSecrets bulk-upserts local entries to the server.
func (c *Client) PutSecrets(entries []store.Entry) error {
	wire := make([]RemoteEntry, len(entries))
	for i, e := range entries {
		wire[i] = RemoteEntry{
			KeyName:        e.Key,
			EncryptedValue: hex.EncodeToString(e.Value),
			UpdatedAt:      e.UpdatedAt,
		}
	}
	body, _ := json.Marshal(wire)
	req, _ := http.NewRequest(http.MethodPut, c.base+"/secrets", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http put: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server %d: %s", resp.StatusCode, msg)
	}
	return nil
}

// DeleteSecret removes a single secret from the server.
func (c *Client) DeleteSecret(keyName string) error {
	req, _ := http.NewRequest(http.MethodDelete, c.base+"/secrets/"+keyName, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http delete: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server %d: %s", resp.StatusCode, msg)
	}
	return nil
}

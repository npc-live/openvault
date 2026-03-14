package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/npc-live/openvault/internal/auth"
	"github.com/npc-live/openvault/internal/config"
	"github.com/npc-live/openvault/internal/input"
	"github.com/npc-live/openvault/internal/keychain"
	"github.com/npc-live/openvault/internal/remote"
	"github.com/npc-live/openvault/internal/vault"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Create a cloud account and enable E2EE sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(config.DefaultDBPath()); os.IsNotExist(err) {
			return fmt.Errorf("vault not found — run `openvault init` first")
		}

		fmt.Print("Email: ")
		var email string
		if _, err := fmt.Scanln(&email); err != nil {
			return fmt.Errorf("read email: %w", err)
		}

		password, err := input.ReadSecret("Password: ")
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
		defer func() {
			for i := range password {
				password[i] = 0
			}
		}()

		confirm, err := input.ReadSecret("Confirm password: ")
		if err != nil {
			return fmt.Errorf("read confirm: %w", err)
		}
		if string(password) != string(confirm) {
			for i := range confirm {
				confirm[i] = 0
			}
			return fmt.Errorf("passwords do not match")
		}
		for i := range confirm {
			confirm[i] = 0
		}

		// ── Step 1: server registration + email verification ──────────────────
		// Do ALL server calls BEFORE touching the local vault.
		// This prevents partial state if network/server fails.

		rc := remote.New("")
		msg, err := rc.Register(email, string(password))
		if err != nil {
			return fmt.Errorf("register: %w", err)
		}
		fmt.Println(msg)

		fmt.Print("Verification code: ")
		var code string
		if _, err := fmt.Scanln(&code); err != nil {
			return fmt.Errorf("read code: %w", err)
		}
		code = strings.TrimSpace(code)

		token, err := rc.VerifyEmail(email, code)
		if err != nil {
			return fmt.Errorf("verify email: %w", err)
		}

		// ── Step 2: re-encrypt vault now that we have a confirmed account ─────

		secretKey := make([]byte, 32)
		if _, err := rand.Read(secretKey); err != nil {
			return fmt.Errorf("generate secret key: %w", err)
		}
		vaultKey := vault.DeriveKey(password, secretKey)

		kc := keychain.New()
		v, err := vault.Open(config.DefaultDBPath(), kc)
		if err != nil {
			return fmt.Errorf("open vault: %w", err)
		}
		defer v.Close()

		if err := v.ReEncrypt(vaultKey); err != nil {
			return fmt.Errorf("re-encrypt vault: %w", err)
		}
		if err := kc.SetKey(config.ServiceName, vaultKey); err != nil {
			return fmt.Errorf("update keychain: %w", err)
		}

		if err := auth.SetToken(kc, token); err != nil {
			return fmt.Errorf("store token: %w", err)
		}

		// ── Step 3: write secret key to config dir (not home root) ────────────
		skPath := config.SecretKeyPath()
		if err := os.WriteFile(skPath, []byte(hex.EncodeToString(secretKey)+"\n"), 0600); err != nil {
			return fmt.Errorf("write secret key: %w", err)
		}

		// ── Step 4: push local entries ────────────────────────────────────────
		entries, err := v.ListEntries()
		if err != nil {
			return fmt.Errorf("list entries: %w", err)
		}
		if len(entries) > 0 {
			if err := remote.New(token).PutSecrets(entries); err != nil {
				return fmt.Errorf("push secrets: %w", err)
			}
		}

		fmt.Printf("\nRegistered successfully!\n")
		fmt.Printf("Secret key saved to: %s\n", skPath)
		fmt.Printf("IMPORTANT: Back up this file — you need it to log in on a new device.\n")
		if len(entries) > 0 {
			fmt.Printf("Pushed %d secret(s) to the cloud.\n", len(entries))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}

package cmd

import (
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

var secretKeyFile string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to cloud sync on this device",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Prompt email.
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

		kc := keychain.New()

		// If --secret-key provided, derive and update the local vault key.
		if secretKeyFile != "" {
			raw, err := os.ReadFile(secretKeyFile)
			if err != nil {
				return fmt.Errorf("read secret key file: %w", err)
			}
			hexStr := strings.TrimSpace(string(raw))
			secretKey, err := hex.DecodeString(hexStr)
			if err != nil {
				return fmt.Errorf("decode secret key: %w", err)
			}
			vaultKey := vault.DeriveKey(password, secretKey)
			if err := kc.SetKey(config.ServiceName, vaultKey); err != nil {
				return fmt.Errorf("update keychain: %w", err)
			}
			fmt.Println("Vault key derived and stored.")
		}

		// Authenticate → JWT.
		rc := remote.New("")
		token, err := rc.Login(email, string(password))
		if err != nil {
			return fmt.Errorf("login: %w", err)
		}

		if err := auth.SetToken(kc, token); err != nil {
			return fmt.Errorf("store token: %w", err)
		}

		fmt.Println("Logged in. Run `openvault sync` to pull your secrets.")
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&secretKeyFile, "secret-key", "", "Path to secret key file (required on new devices)")
	rootCmd.AddCommand(loginCmd)
}

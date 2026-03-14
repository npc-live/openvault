package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/qing/openvault/internal/config"
	"github.com/qing/openvault/internal/keychain"
	"github.com/qing/openvault/internal/vault"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new vault",
	Long:  "Creates an encrypted vault and stores the master key in your system keychain.",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := config.DefaultDBPath()

		// Check if already initialized.
		if _, err := os.Stat(dbPath); err == nil {
			fmt.Fprintln(os.Stderr, "Vault already initialized at", dbPath)
			return nil
		}

		kc := keychain.New()
		if err := vault.Init(dbPath, kc); err != nil {
			return fmt.Errorf("init vault: %w", err)
		}

		fmt.Println("Vault initialized at", dbPath)
		fmt.Println()
		fmt.Println("To enable automatic secret injection into your shell, add the following")
		fmt.Println("to your shell configuration file:")
		fmt.Println()

		self, _ := os.Executable()
		if self == "" {
			self = "openvault"
		}

		fmt.Println("  # zsh (~/.zshrc)")
		fmt.Printf("  eval \"$(%s shell-init --shell zsh)\"\n", self)
		fmt.Println()
		fmt.Println("  # bash (~/.bashrc or ~/.bash_profile)")
		fmt.Printf("  eval \"$(%s shell-init --shell bash)\"\n", self)
		fmt.Println()
		fmt.Println("Then restart your shell or run: source ~/.zshrc")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

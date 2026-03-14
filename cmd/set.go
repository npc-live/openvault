package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/npc-live/openvault/internal/config"
	"github.com/npc-live/openvault/internal/input"
	"github.com/npc-live/openvault/internal/keychain"
	"github.com/npc-live/openvault/internal/vault"
)

var setCmd = &cobra.Command{
	Use:   "set <KEY>",
	Short: "Set a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		secret, err := input.ReadSecret(fmt.Sprintf("Enter value for %s: ", name))
		if err != nil {
			return err
		}
		if len(secret) == 0 {
			return fmt.Errorf("value cannot be empty")
		}

		kc := keychain.New()
		v, err := vault.Open(config.DefaultDBPath(), kc)
		if err != nil {
			return fmt.Errorf("open vault: %w", err)
		}
		defer v.Close()

		if err := v.Set(name, secret); err != nil {
			return err
		}
		fmt.Printf("Secret %q stored.\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}

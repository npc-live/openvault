package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/qing/openvault/internal/config"
	"github.com/qing/openvault/internal/keychain"
	"github.com/qing/openvault/internal/store"
	"github.com/qing/openvault/internal/vault"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <KEY>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a secret",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		kc := keychain.New()
		v, err := vault.Open(config.DefaultDBPath(), kc)
		if err != nil {
			return fmt.Errorf("open vault: %w", err)
		}
		defer v.Close()

		if err := v.Delete(name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return fmt.Errorf("secret %q not found", name)
			}
			return err
		}
		fmt.Printf("Secret %q deleted.\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

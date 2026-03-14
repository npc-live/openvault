package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/npc-live/openvault/internal/auth"
	"github.com/npc-live/openvault/internal/config"
	"github.com/npc-live/openvault/internal/keychain"
	"github.com/npc-live/openvault/internal/remote"
	"github.com/npc-live/openvault/internal/store"
	"github.com/npc-live/openvault/internal/vault"
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

		// If logged in, also delete from server (best-effort).
		if token, err := auth.GetToken(kc); err == nil && len(token) > 0 {
			rc := remote.New(token)
			if rerr := rc.DeleteSecret(name); rerr != nil {
				fmt.Printf("Warning: could not delete from server: %v\n", rerr)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

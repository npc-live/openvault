package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/npc-live/openvault/internal/config"
	"github.com/npc-live/openvault/internal/keychain"
	"github.com/npc-live/openvault/internal/vault"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all secret names",
	RunE: func(cmd *cobra.Command, args []string) error {
		kc := keychain.New()
		v, err := vault.Open(config.DefaultDBPath(), kc)
		if err != nil {
			return fmt.Errorf("open vault: %w", err)
		}
		defer v.Close()

		keys, err := v.List()
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			fmt.Println("(no secrets stored)")
			return nil
		}
		for _, k := range keys {
			fmt.Println(k)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

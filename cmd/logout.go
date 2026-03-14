package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/npc-live/openvault/internal/auth"
	"github.com/npc-live/openvault/internal/keychain"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear cloud sync credentials from this device",
	RunE: func(cmd *cobra.Command, args []string) error {
		kc := keychain.New()
		auth.DeleteToken(kc)
		fmt.Println("Logged out.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

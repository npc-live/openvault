package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/npc-live/openvault/internal/auth"
	"github.com/npc-live/openvault/internal/keychain"
	"github.com/npc-live/openvault/internal/remote"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear cloud sync credentials from this device",
	RunE: func(cmd *cobra.Command, args []string) error {
		kc := keychain.New()

		// Revoke the JWT on the server (best-effort — don't fail if offline).
		if token, err := auth.GetToken(kc); err == nil && len(token) > 0 {
			if rerr := remote.New(token).Logout(); rerr != nil {
				fmt.Printf("Warning: could not revoke token on server: %v\n", rerr)
			}
		}

		auth.DeleteToken(kc)
		fmt.Println("Logged out.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

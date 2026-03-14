package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/qing/openvault/internal/config"
	"github.com/qing/openvault/internal/keychain"
	"github.com/qing/openvault/internal/vault"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print all secrets as export statements",
	Long: `Prints all secrets as shell export statements, suitable for eval.
Example usage in shell:
  eval "$(openvault env)"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		kc := keychain.New()
		v, err := vault.Open(config.DefaultDBPath(), kc)
		if err != nil {
			return fmt.Errorf("open vault: %w", err)
		}
		defer v.Close()

		env, err := v.Env()
		if err != nil {
			return err
		}
		for k, val := range env {
			fmt.Printf("export %s=%q\n", k, val)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
}

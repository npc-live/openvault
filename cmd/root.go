package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "openvault",
	Short: "Encrypted local secret storage with automatic shell injection",
	Long: `OpenVault stores your API keys and secrets encrypted on disk,
injecting them into your shell automatically — no more .env files or
hardcoded credentials.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

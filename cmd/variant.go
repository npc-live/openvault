package cmd

import (
	"fmt"
	"strconv"

	"github.com/npc-live/openvault/internal/config"
	"github.com/npc-live/openvault/internal/input"
	"github.com/npc-live/openvault/internal/keychain"
	"github.com/npc-live/openvault/internal/vault"
	"github.com/spf13/cobra"
)

var variantCmd = &cobra.Command{
	Use:   "variant",
	Short: "Manage multiple values for a secret",
}

var variantAddCmd = &cobra.Command{
	Use:   "add <KEY>",
	Short: "Add a new variant value for a secret",
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

		idx, err := v.VariantAdd(name, secret)
		if err != nil {
			return err
		}
		fmt.Printf("Added variant %d for %q.\n", idx, name)
		return nil
	},
}

var variantUseCmd = &cobra.Command{
	Use:   "use <KEY> <INDEX>",
	Short: "Switch the active variant for a secret",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		idx, err := strconv.Atoi(args[1])
		if err != nil || idx < 1 {
			return fmt.Errorf("index must be a positive integer")
		}

		kc := keychain.New()
		v, err := vault.Open(config.DefaultDBPath(), kc)
		if err != nil {
			return fmt.Errorf("open vault: %w", err)
		}
		defer v.Close()

		if err := v.VariantUse(name, idx); err != nil {
			return err
		}
		fmt.Printf("Switched %q to variant %d.\n", name, idx)
		return nil
	},
}

var variantListCmd = &cobra.Command{
	Use:     "list <KEY>",
	Aliases: []string{"ls"},
	Short:   "List all variants for a secret",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		kc := keychain.New()
		v, err := vault.Open(config.DefaultDBPath(), kc)
		if err != nil {
			return fmt.Errorf("open vault: %w", err)
		}
		defer v.Close()

		entries, err := v.VariantList(name)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Printf("No variants for %q. Use `openvault variant add %s` to add one.\n", name, name)
			return nil
		}

		for _, e := range entries {
			marker := " "
			if e.Active {
				marker = "*"
			}
			fmt.Printf("[%s] %d  %s\n", marker, e.Index, maskValue(string(e.Value)))
		}
		return nil
	},
}

var variantRmCmd = &cobra.Command{
	Use:   "rm <KEY> <INDEX>",
	Short: "Remove a variant for a secret",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		idx, err := strconv.Atoi(args[1])
		if err != nil || idx < 1 {
			return fmt.Errorf("index must be a positive integer")
		}

		kc := keychain.New()
		v, err := vault.Open(config.DefaultDBPath(), kc)
		if err != nil {
			return fmt.Errorf("open vault: %w", err)
		}
		defer v.Close()

		if err := v.VariantRemove(name, idx); err != nil {
			return err
		}
		fmt.Printf("Removed variant %d from %q.\n", idx, name)
		return nil
	},
}

func maskValue(s string) string {
	if len(s) <= 8 {
		return "••••••••"
	}
	return s[:4] + "••••" + s[len(s)-4:]
}

func init() {
	variantCmd.AddCommand(variantAddCmd)
	variantCmd.AddCommand(variantUseCmd)
	variantCmd.AddCommand(variantListCmd)
	variantCmd.AddCommand(variantRmCmd)
	rootCmd.AddCommand(variantCmd)
}

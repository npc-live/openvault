package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/npc-live/openvault/internal/input"
	"github.com/npc-live/openvault/internal/remote"
)

var forgotPasswordCmd = &cobra.Command{
	Use:   "forgot-password",
	Short: "Reset your cloud account password via email",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("Email: ")
		var email string
		if _, err := fmt.Scanln(&email); err != nil {
			return fmt.Errorf("read email: %w", err)
		}

		rc := remote.New("")
		msg, err := rc.ForgotPassword(email)
		if err != nil {
			return err
		}
		fmt.Println(msg)

		fmt.Print("Reset code: ")
		var code string
		if _, err := fmt.Scanln(&code); err != nil {
			return fmt.Errorf("read code: %w", err)
		}
		code = strings.TrimSpace(code)

		newPassword, err := input.ReadSecret("New password: ")
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
		confirm, err := input.ReadSecret("Confirm new password: ")
		if err != nil {
			return fmt.Errorf("read confirm: %w", err)
		}
		if string(newPassword) != string(confirm) {
			return fmt.Errorf("passwords do not match")
		}

		result, err := rc.ResetPassword(email, code, string(newPassword))
		if err != nil {
			return fmt.Errorf("reset password: %w", err)
		}
		fmt.Println(result)
		fmt.Println("Run `openvault login` to authenticate with your new password.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(forgotPasswordCmd)
}

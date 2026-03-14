package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var shellFlag string

var shellInitCmd = &cobra.Command{
	Use:   "shell-init",
	Short: "Print shell integration code for automatic secret injection",
	Long: `Prints shell hook code that automatically injects secrets before
each command. Add the eval line to your shell config for seamless injection.

  # zsh (~/.zshrc)
  eval "$(openvault shell-init --shell zsh)"

  # bash (~/.bashrc)
  eval "$(openvault shell-init --shell bash)"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		self, err := os.Executable()
		if err != nil {
			self = "openvault"
		}

		switch shellFlag {
		case "zsh":
			fmt.Printf(`_openvault_preexec() {
  eval "$(%s env 2>/dev/null)"
}
autoload -Uz add-zsh-hook
add-zsh-hook preexec _openvault_preexec
`, self)
		case "bash":
			fmt.Printf(`_openvault_inject() {
  eval "$(%s env 2>/dev/null)"
}
if [[ -z "$PROMPT_COMMAND" ]]; then
  PROMPT_COMMAND="_openvault_inject"
else
  PROMPT_COMMAND="_openvault_inject;$PROMPT_COMMAND"
fi
`, self)
		default:
			return fmt.Errorf("unsupported shell %q; supported: zsh, bash", shellFlag)
		}
		return nil
	},
}

func init() {
	shellInitCmd.Flags().StringVar(&shellFlag, "shell", "", "Shell type: zsh or bash (required)")
	shellInitCmd.MarkFlagRequired("shell")
	rootCmd.AddCommand(shellInitCmd)
}

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/qing/openvault/internal/config"
	"github.com/qing/openvault/internal/keychain"
	"github.com/qing/openvault/internal/vault"
)

var runCmd = &cobra.Command{
	Use:   "run <command> [args...]",
	Short: "Run a command with secrets injected as environment variables",
	Long: `Runs the given command with all vault secrets injected as environment
variables. Uses syscall.Exec to replace the current process, so signals
are transparently forwarded.

Example:
  openvault run npm run dev
  openvault run docker push myimage`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		kc := keychain.New()
		v, err := vault.Open(config.DefaultDBPath(), kc)
		if err != nil {
			return fmt.Errorf("open vault: %w", err)
		}

		env, err := v.Env()
		v.Close() // close before exec
		if err != nil {
			return err
		}

		// Resolve the binary.
		binary, err := exec.LookPath(args[0])
		if err != nil {
			return fmt.Errorf("command not found: %s", args[0])
		}

		// Build environment: inherit current env, then overlay vault secrets.
		environ := os.Environ()
		for k, val := range env {
			environ = append(environ, k+"="+val)
		}

		// Replace this process with the target command (signals forwarded).
		return syscall.Exec(binary, args, environ)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

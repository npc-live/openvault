package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/npc-live/openvault/internal/config"
	"github.com/npc-live/openvault/internal/keychain"
	"github.com/npc-live/openvault/internal/vault"
	"github.com/spf13/cobra"
)

var resolveExport bool

var resolveCmd = &cobra.Command{
	Use:   "resolve <file.json>",
	Short: "Resolve vault refs in a JSON file",
	Long: `Reads a JSON file, resolves all "*_ref" fields by looking up their
values in the vault, and outputs the result.

Fields ending in "_ref" are treated as vault key references. The resolved
output replaces "_ref" keys with their base names containing the plaintext
value. Non-ref fields are passed through unchanged.

Example JSON (data/api-creds.json):
  {
    "wallet_address": "0xABCD",
    "secrets_backend": "openvault",
    "poly_api_key_ref": "clawfirm/poly-api-key",
    "poly_secret_ref":  "clawfirm/poly-secret"
  }

Output (openvault resolve data/api-creds.json):
  {
    "wallet_address": "0xABCD",
    "poly_api_key": "sk-xxx",
    "poly_secret":  "yyy"
  }

With --export:
  export POLY_API_KEY=sk-xxx
  export POLY_SECRET=yyy`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parse JSON: %w", err)
		}

		// Collect which vault keys we need to resolve.
		type refEntry struct {
			outputKey string
			vaultKey  string
		}
		var refs []refEntry
		for k, v := range raw {
			if strings.HasSuffix(k, "_ref") {
				vaultKey, ok := v.(string)
				if !ok {
					return fmt.Errorf("field %q must be a string vault key reference", k)
				}
				refs = append(refs, refEntry{
					outputKey: strings.TrimSuffix(k, "_ref"),
					vaultKey:  vaultKey,
				})
			}
		}

		if len(refs) == 0 {
			return fmt.Errorf("no *_ref fields found in %s", args[0])
		}

		kc := keychain.New()
		v, err := vault.Open(config.DefaultDBPath(), kc)
		if err != nil {
			return fmt.Errorf("open vault: %w", err)
		}
		defer v.Close()

		resolved := make(map[string]string, len(refs))
		for _, ref := range refs {
			val, err := v.Get(ref.vaultKey)
			if err != nil {
				return fmt.Errorf("resolve %q → vault key %q: %w", ref.outputKey, ref.vaultKey, err)
			}
			resolved[ref.outputKey] = string(val)
		}

		if resolveExport {
			for k, val := range resolved {
				fmt.Printf("export %s=%s\n", strings.ToUpper(k), shellQuote(val))
			}
			return nil
		}

		// Build output JSON: pass-through non-ref fields, add resolved fields.
		out := make(map[string]any, len(raw))
		for k, val := range raw {
			if !strings.HasSuffix(k, "_ref") {
				out[k] = val
			}
		}
		for k, val := range resolved {
			out[k] = val
		}

		enc, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal output: %w", err)
		}
		fmt.Println(string(enc))
		return nil
	},
}

// shellQuote wraps a value in single quotes, escaping any single quotes inside.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func init() {
	resolveCmd.Flags().BoolVar(&resolveExport, "export", false, "Output as shell export statements instead of JSON")
	rootCmd.AddCommand(resolveCmd)
}

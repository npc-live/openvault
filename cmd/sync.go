package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/npc-live/openvault/internal/auth"
	"github.com/npc-live/openvault/internal/config"
	"github.com/npc-live/openvault/internal/keychain"
	"github.com/npc-live/openvault/internal/remote"
	"github.com/npc-live/openvault/internal/vault"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync secrets with the cloud (pull + merge + push)",
	RunE: func(cmd *cobra.Command, args []string) error {
		kc := keychain.New()

		token, err := auth.GetToken(kc)
		if err != nil {
			return err // ErrNotLoggedIn
		}
		if len(token) == 0 {
			return auth.ErrNotLoggedIn
		}

		v, err := vault.Open(config.DefaultDBPath(), kc)
		if err != nil {
			return fmt.Errorf("open vault: %w", err)
		}
		defer v.Close()

		rc := remote.New(token)

		// Pull remote entries.
		remoteEntries, err := rc.GetSecrets()
		if err != nil {
			return fmt.Errorf("pull from server: %w", err)
		}

		// Get local entries (encrypted bytes + timestamps).
		localEntries, err := v.ListEntries()
		if err != nil {
			return fmt.Errorf("list local entries: %w", err)
		}

		// Build local index: key → {value, updatedAt}.
		type localEntry struct {
			value     []byte
			updatedAt int64
		}
		localIndex := make(map[string]localEntry, len(localEntries))
		for _, e := range localEntries {
			localIndex[e.Key] = localEntry{value: e.Value, updatedAt: e.UpdatedAt}
		}

		// Merge: remote wins when its timestamp is strictly newer.
		pulled := 0
		for _, re := range remoteEntries {
			encBytes, err := hex.DecodeString(re.EncryptedValue)
			if err != nil {
				return fmt.Errorf("decode remote entry %q: %w", re.KeyName, err)
			}
			local, exists := localIndex[re.KeyName]
			if !exists || re.UpdatedAt > local.updatedAt {
				if err := v.RawSet(re.KeyName, encBytes, re.UpdatedAt); err != nil {
					return fmt.Errorf("store %q: %w", re.KeyName, err)
				}
				pulled++
			}
		}

		// Reload local entries after merge (includes newly pulled ones).
		mergedEntries, err := v.ListEntries()
		if err != nil {
			return fmt.Errorf("list merged entries: %w", err)
		}

		// Push all merged entries to server.
		pushed := 0
		if err := rc.PutSecrets(mergedEntries); err != nil {
			return fmt.Errorf("push to server: %w", err)
		}
		pushed = len(mergedEntries)

		fmt.Printf("↓ %d pulled, ↑ %d pushed\n", pulled, pushed)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

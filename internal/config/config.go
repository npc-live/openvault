package config

import (
	"os"
	"path/filepath"
)

const (
	ServiceName = "openvault"
	DBFileName  = "vault.db"
	DirPerm     = 0700
	FilePerm    = 0600
	APIBaseURL  = "https://openvault.mrwill84.workers.dev"
)

func DefaultDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, ServiceName)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/tmp", ServiceName)
	}
	return filepath.Join(home, ".config", ServiceName)
}

func DefaultDBPath() string {
	return filepath.Join(DefaultDir(), DBFileName)
}

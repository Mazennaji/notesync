package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Version        string `json:"version"`
	VaultPath      string `json:"vaultPath"`
	NotionParentID string `json:"notionParentId"`
	SyncMode       string `json:"syncMode"`
}

func Validate(c Config) error {
	if c.Version == "" {
		return errors.New("version is required")
	}
	if c.VaultPath == "" {
		return errors.New("vaultPath is required")
	}
	info, err := os.Stat(c.VaultPath)
	if err != nil {
		return fmt.Errorf("vaultPath does not exist: %s", c.VaultPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("vaultPath is not a directory: %s", c.VaultPath)
	}
	if c.SyncMode != "manual" && c.SyncMode != "watch" {
		return fmt.Errorf("invalid syncMode: %q (want manual or watch)", c.SyncMode)
	}
	return nil
}

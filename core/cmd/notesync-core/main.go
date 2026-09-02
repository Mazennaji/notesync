package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Mazennaji/notesync/core/internal/config"
	"github.com/Mazennaji/notesync/core/internal/ipc"
	"github.com/Mazennaji/notesync/core/internal/obsidian"
	"github.com/Mazennaji/notesync/core/internal/storage"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		fail(logger, "failed to read stdin: "+err.Error())
		return
	}

	var req ipc.Request
	if err := json.Unmarshal(in, &req); err != nil {
		fail(logger, "invalid request json: "+err.Error())
		return
	}

	logger.Info("received command", "command", req.Command)

	resp := dispatch(req, logger)

	if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
		logger.Error("failed to encode response", "error", err)
	}
}

func dispatch(req ipc.Request, logger *slog.Logger) ipc.Response {
	switch req.Command {
	case "ping":
		return ipc.Response{OK: true, Data: map[string]string{"message": "pong"}}

	case "status":
		cfg, err := decodeConfig(req.Config)
		if err != nil {
			return ipc.Response{OK: false, Error: "bad config: " + err.Error()}
		}
		store, err := storage.Open(filepath.Join(cfg.VaultPath, ".notesync", "state.db"))
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		defer store.Close()
		count, err := store.CountNotes()
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		return ipc.Response{OK: true, Data: map[string]any{
			"vaultPath": cfg.VaultPath,
			"syncMode":  cfg.SyncMode,
			"notes":     count,
		}}

	case "config.validate":
		cfg, err := decodeConfig(req.Config)
		if err != nil {
			return ipc.Response{OK: false, Error: "bad config: " + err.Error()}
		}
		if err := config.Validate(cfg); err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		return ipc.Response{OK: true, Data: map[string]bool{"valid": true}}

	case "db.init":
		cfg, err := decodeConfig(req.Config)
		if err != nil {
			return ipc.Response{OK: false, Error: "bad config: " + err.Error()}
		}
		if err := config.Validate(cfg); err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		dbPath := filepath.Join(cfg.VaultPath, ".notesync", "state.db")
		store, err := storage.Open(dbPath)
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		defer store.Close()
		logger.Info("state db ready", "path", dbPath)
		return ipc.Response{OK: true, Data: map[string]string{"dbPath": dbPath}}

	case "vault.scan":
		cfg, err := decodeConfig(req.Config)
		if err != nil {
			return ipc.Response{OK: false, Error: "bad config: " + err.Error()}
		}
		if err := config.Validate(cfg); err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		paths, err := obsidian.Discover(cfg.VaultPath)
		if err != nil {
			return ipc.Response{OK: false, Error: "scan failed: " + err.Error()}
		}
		store, err := storage.Open(filepath.Join(cfg.VaultPath, ".notesync", "state.db"))
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		defer store.Close()
		added, err := store.UpsertNotes(paths)
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		logger.Info("vault scanned", "found", len(paths), "added", added)
		return ipc.Response{OK: true, Data: map[string]int{
			"found": len(paths), "added": added}}

	default:
		return ipc.Response{OK: false, Error: "unknown command: " + req.Command}
	}
}

func decodeConfig(raw map[string]any) (config.Config, error) {
	var cfg config.Config
	b, err := json.Marshal(raw)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func fail(l *slog.Logger, msg string) {
	l.Error(msg)
	_ = json.NewEncoder(os.Stdout).Encode(ipc.Response{OK: false, Error: msg})
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Mazennaji/notesync/core/internal/auth"
	"github.com/Mazennaji/notesync/core/internal/config"
	"github.com/Mazennaji/notesync/core/internal/ipc"
	"github.com/Mazennaji/notesync/core/internal/markdown"
	"github.com/Mazennaji/notesync/core/internal/notion"
	"github.com/Mazennaji/notesync/core/internal/obsidian"
	"github.com/Mazennaji/notesync/core/internal/storage"
	"github.com/Mazennaji/notesync/core/internal/sync"
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

	case "auth.store":
		token, _ := req.Args["token"].(string)
		name, err := auth.Store(token)
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		logger.Info("notion authenticated", "integration", name)
		return ipc.Response{OK: true, Data: map[string]string{"integration": name}}

	case "auth.status":
		if _, err := auth.Get(); err != nil {
			return ipc.Response{OK: true, Data: map[string]bool{"authenticated": false}}
		}
		return ipc.Response{OK: true, Data: map[string]bool{"authenticated": true}}

	case "auth.logout":
		if err := auth.Delete(); err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		return ipc.Response{OK: true, Data: map[string]bool{"loggedOut": true}}

	case "notion.pages":
		cfg, err := decodeConfig(req.Config)
		if err != nil {
			return ipc.Response{OK: false, Error: "bad config: " + err.Error()}
		}
		client, err := notion.New()
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		pages, err := client.SearchPages()
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}

		store, err := storage.Open(filepath.Join(cfg.VaultPath, ".notesync", "state.db"))
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		defer store.Close()

		linked := 0
		for _, pg := range pages {
			ok, err := store.LinkNotionPage(pg.Title, pg.ID)
			if err != nil {
				return ipc.Response{OK: false, Error: err.Error()}
			}
			if ok {
				linked++
			}
		}
		logger.Info("notion pages discovered", "found", len(pages), "linked", linked)
		return ipc.Response{OK: true, Data: map[string]int{
			"found": len(pages), "linked": linked,
		}}

	case "config.setParent":
		if _, err := decodeConfig(req.Config); err != nil {
			return ipc.Response{OK: false, Error: "bad config: " + err.Error()}
		}
		parentID, _ := req.Args["parentId"].(string)
		if parentID == "" {
			return ipc.Response{OK: false, Error: "parentId is required"}
		}
		client, err := notion.New()
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		if err := client.CheckPage(parentID); err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		return ipc.Response{OK: true, Data: map[string]string{"parentId": parentID}}

	case "notion.create":
		cfg, err := decodeConfig(req.Config)
		if err != nil {
			return ipc.Response{OK: false, Error: "bad config: " + err.Error()}
		}
		if cfg.NotionParentID == "" {
			return ipc.Response{OK: false, Error: "no parent page set (run `notesync parent`)"}
		}
		client, err := notion.New()
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		store, err := storage.Open(filepath.Join(cfg.VaultPath, ".notesync", "state.db"))
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		defer store.Close()

		unlinked, err := store.UnlinkedNotes()
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}

		created := 0
		for _, n := range unlinked {
			pageID, err := client.CreatePage(cfg.NotionParentID, n.Title)
			if err != nil {
				return ipc.Response{OK: false, Error: fmt.Sprintf("creating %q: %v", n.Title, err)}
			}
			if err := store.SetNotionPageID(n.ID, pageID); err != nil {
				return ipc.Response{OK: false, Error: err.Error()}
			}
			created++
		}
		logger.Info("notion pages created", "created", created)
		return ipc.Response{OK: true, Data: map[string]int{"created": created}}

	case "notion.push":
		cfg, err := decodeConfig(req.Config)
		if err != nil {
			return ipc.Response{OK: false, Error: "bad config: " + err.Error()}
		}
		client, err := notion.New()
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		store, err := storage.Open(filepath.Join(cfg.VaultPath, ".notesync", "state.db"))
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		defer store.Close()

		linked, err := store.LinkedNotes()
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}

		pushed, skipped := 0, 0
		for _, n := range linked {
			raw, err := os.ReadFile(filepath.Join(cfg.VaultPath, filepath.FromSlash(n.LocalPath)))
			if err != nil {
				return ipc.Response{OK: false, Error: fmt.Sprintf("reading %s: %v", n.LocalPath, err)}
			}
			localHash := sync.Hash(string(raw))

			remoteMD, err := client.FetchMarkdown(n.NotionPageID)
			if err != nil {
				return ipc.Response{OK: false, Error: fmt.Sprintf("fetching %s: %v", n.Title, err)}
			}
			remoteHash := sync.Hash(remoteMD)

			lastSynced, _ := store.LastSyncedHash(n.ID)
			d := sync.Classify(n.LocalPath, localHash, remoteHash, lastSynced)

			if d.Action != sync.Push && d.Action != sync.NewLocal {
				skipped++
				continue
			}

			blocks, err := markdown.Parse(raw)
			if err != nil {
				_ = store.LogHistory(n.ID, "push", "local_to_remote", "error", localHash, remoteHash, err.Error())
				return ipc.Response{OK: false, Error: fmt.Sprintf("parsing %s: %v", n.LocalPath, err)}
			}
			if err := client.UpdatePageContent(n.NotionPageID, blocks); err != nil {
				_ = store.LogHistory(n.ID, "push", "local_to_remote", "error", localHash, remoteHash, err.Error())
				return ipc.Response{OK: false, Error: fmt.Sprintf("pushing %s: %v", n.LocalPath, err)}
			}

			if err := store.RecordSync(n.ID, localHash, localHash, localHash, "synced"); err != nil {
				return ipc.Response{OK: false, Error: err.Error()}
			}
			_ = store.LogHistory(n.ID, "push", "local_to_remote", "success", localHash, localHash, "")
			pushed++
		}

		logger.Info("push complete", "pushed", pushed, "skipped", skipped)
		return ipc.Response{OK: true, Data: map[string]int{"pushed": pushed, "skipped": skipped}}

	case "sync.diff":
		cfg, err := decodeConfig(req.Config)
		if err != nil {
			return ipc.Response{OK: false, Error: "bad config: " + err.Error()}
		}
		client, err := notion.New()
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		store, err := storage.Open(filepath.Join(cfg.VaultPath, ".notesync", "state.db"))
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		defer store.Close()

		linked, err := store.LinkedNotes()
		if err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}

		var decisions []map[string]string
		counts := map[string]int{}
		for _, n := range linked {
			raw, err := os.ReadFile(filepath.Join(cfg.VaultPath, filepath.FromSlash(n.LocalPath)))
			if err != nil {
				return ipc.Response{OK: false, Error: fmt.Sprintf("reading %s: %v", n.LocalPath, err)}
			}
			localHash := sync.Hash(string(raw))

			remoteMD, err := client.FetchMarkdown(n.NotionPageID)
			if err != nil {
				return ipc.Response{OK: false, Error: fmt.Sprintf("fetching %s: %v", n.Title, err)}
			}
			remoteHash := sync.Hash(remoteMD)

			lastSynced, err := store.LastSyncedHash(n.ID)
			if err != nil {
				return ipc.Response{OK: false, Error: err.Error()}
			}

			d := sync.Classify(n.LocalPath, localHash, remoteHash, lastSynced)
			counts[string(d.Action)]++
			decisions = append(decisions, map[string]string{
				"note":   d.NotePath,
				"action": string(d.Action),
			})
		}

		return ipc.Response{OK: true, Data: map[string]any{
			"decisions": decisions,
			"counts":    counts,
		}}

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

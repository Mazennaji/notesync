package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"

	"github.com/Mazennaji/notesync/core/internal/config"
	"github.com/Mazennaji/notesync/core/internal/ipc"
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
		return ipc.Response{OK: true, Data: map[string]any{"status": "not yet implemented"}}

	case "config.validate":
		cfg, err := decodeConfig(req.Config)
		if err != nil {
			return ipc.Response{OK: false, Error: "bad config: " + err.Error()}
		}
		if err := config.Validate(cfg); err != nil {
			return ipc.Response{OK: false, Error: err.Error()}
		}
		return ipc.Response{OK: true, Data: map[string]bool{"valid": true}}

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

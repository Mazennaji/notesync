package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"

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

	var resp ipc.Response
	switch req.Command {
	case "ping":
		resp = ipc.Response{OK: true, Data: map[string]string{"message": "pong"}}
	case "status":
		resp = ipc.Response{OK: true, Data: map[string]any{"status": "not yet implemented"}}
	default:
		resp = ipc.Response{OK: false, Error: "unknown command: " + req.Command}
	}

	json.NewEncoder(os.Stdout).Encode(resp)
}

func fail(l *slog.Logger, msg string) {
	l.Error(msg)
	json.NewEncoder(os.Stdout).Encode(ipc.Response{OK: false, Error: msg})
}

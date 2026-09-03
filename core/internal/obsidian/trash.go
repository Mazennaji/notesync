package obsidian

import (
	"os"
	"path/filepath"
	"time"
)

func Trash(vaultPath, relPath string) error {
	src := filepath.Join(vaultPath, filepath.FromSlash(relPath))
	trashDir := filepath.Join(vaultPath, ".notesync", "trash")
	if err := os.MkdirAll(trashDir, 0o755); err != nil {
		return err
	}
	stamp := time.Now().Format("20060102-150405")
	dst := filepath.Join(trashDir, stamp+"-"+filepath.Base(relPath))
	return os.Rename(src, dst)
}

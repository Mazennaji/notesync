package obsidian

import (
	"io/fs"
	"path/filepath"
	"strings"
)

var skipDirs = map[string]bool{
	".notesync": true,
	".obsidian": true,
	".git":      true,
	".trash":    true,
}

func Discover(vaultPath string) ([]string, error) {
	var notes []string

	err := filepath.WalkDir(vaultPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			rel, err := filepath.Rel(vaultPath, path)
			if err != nil {
				return err
			}
			notes = append(notes, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return notes, nil
}

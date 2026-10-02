package vaultresolver

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrFileNotFound = errors.New("file does not exist")
	ErrVaultNotFound = errors.New("Vault not found")
)

type Vault struct {
	Name string
	Path string
}

// FindForFile searches upward from the Markdown file's directory
// and returns the nearest ancestor containing a .obsidian directory.
//
// Example:
//
//   D:\Projects\Thesis\Chapter1\test.md
//
// If:
//
//   D:\Projects\Thesis\.obsidian
//
// exists, the returned Vault is:
//
//   Name = "Thesis"
//   Path = "D:\Projects\Thesis"
//
// The search stops at the filesystem root.
func FindForFile(filePath string) (Vault, error) {
	if strings.TrimSpace(filePath) == "" {
		return Vault{}, errors.New("file path must not be empty")
	}

	absolute, err := filepath.Abs(filePath)
	if err != nil {
		return Vault{}, fmt.Errorf("make absolute file path: %w", err)
	}

	absolute = filepath.Clean(absolute)

	info, err := os.Stat(absolute)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Vault{}, fmt.Errorf("%w: %s", ErrFileNotFound, absolute)
		}
		return Vault{}, fmt.Errorf("stat file: %w", err)
	}

	if info.IsDir() {
		return Vault{}, fmt.Errorf("path is a directory: %s", absolute)
	}

	dir := filepath.Dir(absolute)

	for {
		obsidianDir := filepath.Join(dir, ".obsidian")

		info, err := os.Stat(obsidianDir)
		if err == nil && info.IsDir() {
			name := filepath.Base(dir)

			if strings.TrimSpace(name) == "" {
				return Vault{}, fmt.Errorf("invalid Vault name: %s", dir)
			}

			return Vault{
				Name: name,
				Path: dir,
			}, nil
		}

		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return Vault{}, fmt.Errorf("inspect %s: %w", obsidianDir, err)
		}

		parent := filepath.Dir(dir)

		// Reached filesystem root.
		if parent == dir {
			break
		}

		dir = parent
	}

	return Vault{}, fmt.Errorf("%w for file: %s", ErrVaultNotFound, absolute)
}

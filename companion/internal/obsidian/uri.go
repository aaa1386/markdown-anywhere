package obsidian

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zhangcongke/markdown-anywhere/companion/internal/pathutil"
	
)

var ErrEmptyVaultName = errors.New("Vault name must not be empty")

// BuildOpenURI builds the URI understood by Obsidian's open handler.
func BuildOpenURI(vaultName, vaultRoot, targetPath string) (string, error) {
	if strings.TrimSpace(vaultName) == "" {
		return "", ErrEmptyVaultName
	}

	relative, err := pathutil.RelativeTo(vaultRoot, targetPath)
	if err != nil {
		return "", fmt.Errorf("make Obsidian file path: %w", err)
	}

	// Keep Unicode characters and '/' unchanged.
	// Encode spaces as %20 instead of '+'.
	fileValue := strings.ReplaceAll(relative, " ", "%20")

	return "obsidian://open?vault=" + vaultName + "&file=" + fileValue, nil
}

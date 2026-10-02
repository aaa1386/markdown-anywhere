package obsidian

import (
	"errors"
	"fmt"
	"net/url"
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

	// Obsidian 1.13.7 has trouble resolving a Persian Vault name when
	// it is percent-encoded by url.Values.Encode() and spaces become '+'.
	// Keep the Vault name in its original Unicode form, while still
	// URL-encoding the file path.
	fileValue := url.QueryEscape(relative)

	return "obsidian://open?vault=" + vaultName + "&file=" + fileValue, nil
}

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

	query := url.Values{}
	query.Set("vault", vaultName)
	query.Set("file", relative)
	return "obsidian://open?" + query.Encode(), nil
}

package pathutil

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

const HashPrefixLength = 12

var ErrEmptyPath = errors.New("path must not be empty")

// Absolute returns a cleaned absolute path without resolving symlinks.
func Absolute(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", ErrEmptyPath
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("make path absolute: %w", err)
	}
	return filepath.Clean(abs), nil
}

// NormalizeAbsolute returns the stable path identity used for hashing and
// comparing symlink targets. It intentionally does not resolve symlinks: a
// missing target must retain the same identity it had before removal.
func NormalizeAbsolute(path string) (string, error) {
	abs, err := Absolute(path)
	if err != nil {
		return "", err
	}

	identity := filepath.ToSlash(abs)
	if runtime.GOOS == "windows" {
		identity = strings.ToLower(identity)
	}
	return identity, nil
}

// IsWithin reports whether path is lexically inside root or equal to root.
// Lexical comparison is intentional: a symlink placed inside a Vault must
// remain addressable as a Vault-relative path for Obsidian.
func IsWithin(path, root string) (bool, error) {
	pathAbs, err := Absolute(path)
	if err != nil {
		return false, err
	}
	rootAbs, err := Absolute(root)
	if err != nil {
		return false, err
	}

	if !strings.EqualFold(filepath.VolumeName(pathAbs), filepath.VolumeName(rootAbs)) {
		return false, nil
	}

	rel, err := filepath.Rel(rootAbs, pathAbs)
	if err != nil {
		return false, fmt.Errorf("compare paths: %w", err)
	}
	if rel == "." {
		return true, nil
	}
	if filepath.IsAbs(rel) || rel == ".." {
		return false, nil
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator)), nil
}

// RelativeTo returns a Vault-relative, slash-separated path.
func RelativeTo(root, path string) (string, error) {
	rootAbs, err := Absolute(root)
	if err != nil {
		return "", err
	}
	pathAbs, err := Absolute(path)
	if err != nil {
		return "", err
	}

	within, err := IsWithin(pathAbs, rootAbs)
	if err != nil {
		return "", err
	}
	if !within {
		return "", fmt.Errorf("path %q is outside root %q", pathAbs, rootAbs)
	}

	rel, err := filepath.Rel(rootAbs, pathAbs)
	if err != nil {
		return "", fmt.Errorf("make relative path: %w", err)
	}
	return filepath.ToSlash(rel), nil
}

// HashPrefix returns the short hash used as the external-link filename prefix.
func HashPrefix(path string) (string, error) {
	identity, err := NormalizeAbsolute(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(identity))
	return hex.EncodeToString(sum[:])[:HashPrefixLength], nil
}

// LinkName returns the deterministic mirror filename for a source file.
func LinkName(source string) (string, error) {
	abs, err := Absolute(source)
	if err != nil {
		return "", err
	}
	prefix, err := HashPrefix(abs)
	if err != nil {
		return "", err
	}
	base := filepath.Base(abs)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "", fmt.Errorf("source path has no filename: %q", source)
	}
	return prefix + "-" + base, nil
}

// IsManagedLinkName recognizes the naming convention used by this project.
func IsManagedLinkName(name string) bool {
	if len(name) <= HashPrefixLength+1 || name[HashPrefixLength] != '-' {
		return false
	}
	for _, ch := range name[:HashPrefixLength] {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			return false
		}
	}
	return true
}

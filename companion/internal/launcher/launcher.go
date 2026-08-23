package launcher

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var (
	ErrEmptyURI   = errors.New("Obsidian URI must not be empty")
	ErrInvalidURI = errors.New("invalid Obsidian URI")
)

// ValidateURI limits the launcher to the URI shape produced by the Obsidian
// package. It prevents the Windows shell from receiving an arbitrary protocol.
func ValidateURI(uri string) error {
	if strings.TrimSpace(uri) == "" {
		return ErrEmptyURI
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("parse Obsidian URI: %w", err)
	}
	if parsed.Scheme != "obsidian" || parsed.Host != "open" {
		return fmt.Errorf("%w: expected obsidian://open", ErrInvalidURI)
	}
	query := parsed.Query()
	if query.Get("vault") == "" || query.Get("file") == "" {
		return fmt.Errorf("%w: vault and file are required", ErrInvalidURI)
	}
	return nil
}

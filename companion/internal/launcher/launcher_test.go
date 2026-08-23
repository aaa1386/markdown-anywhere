package launcher

import (
	"errors"
	"testing"
)

func TestValidateURIAcceptsObsidianOpenURI(t *testing.T) {
	uri := "obsidian://open?vault=Workbench&file=_external-open%2Fnote.md"
	if err := ValidateURI(uri); err != nil {
		t.Fatal(err)
	}
}

func TestValidateURIRejectsOtherProtocols(t *testing.T) {
	if err := ValidateURI("https://example.com"); !errors.Is(err, ErrInvalidURI) {
		t.Fatalf("expected invalid URI error, got %v", err)
	}
}

func TestValidateURIRejectsMissingFile(t *testing.T) {
	if err := ValidateURI("obsidian://open?vault=Workbench"); !errors.Is(err, ErrInvalidURI) {
		t.Fatalf("expected invalid URI error, got %v", err)
	}
}

func TestValidateURIRejectsEmptyURI(t *testing.T) {
	if err := ValidateURI(" "); !errors.Is(err, ErrEmptyURI) {
		t.Fatalf("expected empty URI error, got %v", err)
	}
}

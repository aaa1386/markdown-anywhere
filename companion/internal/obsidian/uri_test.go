package obsidian

import (
	"net/url"
	"path/filepath"
	"testing"
)

func TestBuildOpenURI(t *testing.T) {
	vault := t.TempDir()
	target := filepath.Join(vault, "子目录", "中文 note.md")
	uri, err := BuildOpenURI("我的 Vault", vault, target)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Scheme != "obsidian" || parsed.Host != "open" {
		t.Fatalf("unexpected URI: %s", uri)
	}
	query := parsed.Query()
	if query.Get("vault") != "我的 Vault" {
		t.Fatalf("unexpected vault query: %q", query.Get("vault"))
	}
	if query.Get("file") != "子目录/中文 note.md" {
		t.Fatalf("unexpected file query: %q", query.Get("file"))
	}
}

func TestBuildOpenURIRejectsExternalPath(t *testing.T) {
	_, err := BuildOpenURI("Vault", t.TempDir(), filepath.Join(t.TempDir(), "note.md"))
	if err == nil {
		t.Fatal("expected external path to be rejected")
	}
}

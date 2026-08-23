package symlink

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhangcongke/markdown-anywhere/companion/internal/obsidian"
	"github.com/zhangcongke/markdown-anywhere/companion/internal/pathutil"
)

func TestEnsureKeepsVaultFileUnlinked(t *testing.T) {
	vault := t.TempDir()
	source := filepath.Join(vault, "inside.md")
	writeFile(t, source)

	manager, err := New(vault, "_external-open")
	if err != nil {
		t.Fatal(err)
	}
	link, err := manager.Ensure(source)
	if err != nil {
		t.Fatal(err)
	}
	if link.External || link.Path != source {
		t.Fatalf("expected original Vault path, got %+v", link)
	}
}

func TestEnsureCreatesAndReusesExternalLink(t *testing.T) {
	vault := t.TempDir()
	externalDir := t.TempDir()
	source := filepath.Join(externalDir, "中文 note.md")
	writeFile(t, source)

	manager, err := New(vault, "_external-open")
	if err != nil {
		t.Fatal(err)
	}
	first, err := manager.Ensure(source)
	if err != nil {
		t.Fatal(err)
	}
	if !first.External || first.Existing {
		t.Fatalf("expected new external link, got %+v", first)
	}
	second, err := manager.Ensure(source)
	if err != nil {
		t.Fatal(err)
	}
	if second.Path != first.Path || !second.Existing {
		t.Fatalf("expected existing link to be reused: first=%+v second=%+v", first, second)
	}

	info, err := os.Lstat(first.Path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink at %s", first.Path)
	}
}

func TestExternalLinkIsAddressableFromObsidian(t *testing.T) {
	vault := t.TempDir()
	externalDir := t.TempDir()
	source := filepath.Join(externalDir, "outside.md")
	writeFile(t, source)

	manager, err := New(vault, "_external-open")
	if err != nil {
		t.Fatal(err)
	}
	link, err := manager.Ensure(source)
	if err != nil {
		t.Fatal(err)
	}
	uri, err := obsidian.BuildOpenURI("Vault", vault, link.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(uri, "file=_external-open%2F") {
		t.Fatalf("expected Vault-relative mirror path in URI, got %s", uri)
	}
}

func TestListAndCleanupInvalidLinks(t *testing.T) {
	vault := t.TempDir()
	externalDir := t.TempDir()
	source := filepath.Join(externalDir, "note.md")
	writeFile(t, source)
	manager, err := New(vault, "_external-open")
	if err != nil {
		t.Fatal(err)
	}
	link, err := manager.Ensure(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}

	links, err := manager.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].Valid {
		t.Fatalf("expected one invalid link, got %+v", links)
	}
	removed, err := manager.CleanupInvalid()
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 1 || removed[0] != link.Path {
		t.Fatalf("unexpected removed links: %+v", removed)
	}
	if _, err := os.Lstat(link.Path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected link to be removed, got %v", err)
	}
}

func TestEnsureRejectsCollision(t *testing.T) {
	vault := t.TempDir()
	externalDir := t.TempDir()
	source := filepath.Join(externalDir, "note.md")
	writeFile(t, source)
	manager, err := New(vault, "_external-open")
	if err != nil {
		t.Fatal(err)
	}
	name, err := filepath.Abs(source)
	if err != nil {
		t.Fatal(err)
	}
	linkName, err := linkNameForTest(name)
	if err != nil {
		t.Fatal(err)
	}
	collision := filepath.Join(manager.MirrorRoot, linkName)
	if err := os.MkdirAll(manager.MirrorRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, collision)
	if _, err := manager.Ensure(source); !errors.Is(err, ErrLinkCollision) {
		t.Fatalf("expected collision error, got %v", err)
	}
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("# note\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func linkNameForTest(source string) (string, error) {
	return pathutil.LinkName(source)
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrDefaultWhenFileIsMissing(t *testing.T) {
	cfg, err := LoadOrDefault(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != CurrentVersion || cfg.MirrorFolder != DefaultMirror || cfg.CleanupDays != DefaultCleanup {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	want := Config{
		Version:      CurrentVersion,
		DefaultVault: Vault{Name: "Workbench", Path: `C:\Vault`},
		MirrorFolder: "_external-open",
		CleanupDays:  14,
	}
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadOrDefault(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestValidateRejectsMirrorOutsideVault(t *testing.T) {
	cfg := Default()
	cfg.MirrorFolder = filepath.Join("..", "outside")
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid mirror folder")
	}
}

func TestLoadRejectsMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadOrDefault(path); err == nil {
		t.Fatal("expected malformed JSON error")
	}
}

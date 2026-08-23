package management

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zhangcongke/markdown-anywhere/companion/internal/config"
	"github.com/zhangcongke/markdown-anywhere/companion/internal/registry"
	"github.com/zhangcongke/markdown-anywhere/companion/internal/symlink"
)

type fakeRegistry struct {
	status registry.AssociationStatus
}

func (r fakeRegistry) Status() (registry.AssociationStatus, error) {
	return r.status, nil
}

func TestStatusReportsConfigAndRegistration(t *testing.T) {
	vault := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := config.Default()
	cfg.DefaultVault = config.Vault{Name: "Workbench", Path: vault}
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatal(err)
	}

	service := Service{
		Version:    "0.1.0-dev",
		ConfigPath: configPath,
		Registry: fakeRegistry{status: registry.AssociationStatus{
			Registered: true,
			Command:    `"C:\\MarkownAnywhere.exe" "%1"`,
		}},
	}
	report, err := service.Status()
	if err != nil {
		t.Fatal(err)
	}
	if !report.ConfigExists || !report.Registered || report.MirrorFolder != config.DefaultMirror {
		t.Fatalf("unexpected status report: %+v", report)
	}
}

func TestLinksAndCleanupReportManagedLinks(t *testing.T) {
	vault := t.TempDir()
	externalDir := t.TempDir()
	source := filepath.Join(externalDir, "outside.md")
	if err := os.WriteFile(source, []byte("# outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := config.Default()
	cfg.DefaultVault = config.Vault{Name: "Workbench", Path: vault}
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatal(err)
	}
	manager, err := symlink.New(vault, cfg.MirrorFolder)
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

	service := Service{ConfigPath: configPath}
	report, err := service.Links()
	if err != nil {
		t.Fatal(err)
	}
	if report.Active != 0 || report.Invalid != 1 || len(report.Links) != 1 || report.Links[0].Path != link.Path {
		t.Fatalf("unexpected links report: %+v", report)
	}
	cleanup, err := service.Cleanup()
	if err != nil {
		t.Fatal(err)
	}
	if cleanup.Count != 1 || len(cleanup.Removed) != 1 || cleanup.Removed[0] != link.Path {
		t.Fatalf("unexpected cleanup report: %+v", cleanup)
	}
}

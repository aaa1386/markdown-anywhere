package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/zhangcongke/markdown-anywhere/companion/internal/config"
	"github.com/zhangcongke/markdown-anywhere/companion/internal/launcher"
	"github.com/zhangcongke/markdown-anywhere/companion/internal/management"
	"github.com/zhangcongke/markdown-anywhere/companion/internal/obsidian"
	"github.com/zhangcongke/markdown-anywhere/companion/internal/registry"
	"github.com/zhangcongke/markdown-anywhere/companion/internal/symlink"
	"github.com/zhangcongke/markdown-anywhere/companion/internal/vaultresolver"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintln(os.Stderr, "usage: markown-anywhere <file> | --register | --unregister | --status --json | --links --json | --cleanup | --version")
		os.Exit(2)
	}

	if os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}

	if os.Args[1] == "--register" {
		executable, err := os.Executable()
		if err != nil {
			fail(err)
		}

		associationRegistry, err := registry.New()
		if err != nil {
			fail(err)
		}

		if err := associationRegistry.Register(executable); err != nil {
			fail(err)
		}

		fmt.Println("registered")
		return
	}

	if os.Args[1] == "--unregister" {
		associationRegistry, err := registry.New()
		if err != nil {
			fail(err)
		}

		if err := associationRegistry.Unregister(); err != nil {
			fail(err)
		}

		fmt.Println("unregistered")
		return
	}

	if len(os.Args) == 3 && os.Args[1] == "--status" && os.Args[2] == "--json" {
		report, err := managementService().Status()
		if err != nil {
			fail(err)
		}

		writeJSON(report)
		return
	}

	if len(os.Args) == 3 && os.Args[1] == "--links" && os.Args[2] == "--json" {
		report, err := managementService().Links()
		if err != nil {
			fail(err)
		}

		writeJSON(report)
		return
	}

	if len(os.Args) == 2 && os.Args[1] == "--cleanup" {
		report, err := managementService().Cleanup()
		if err != nil {
			fail(err)
		}

		writeJSON(report)
		return
	}

	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: markown-anywhere <file> | --register | --unregister | --status --json | --links --json | --cleanup | --version")
		os.Exit(2)
	}

	configPath, err := config.DefaultPath()
	if err != nil {
		fail(err)
	}

	cfg, err := config.LoadOrDefault(configPath)
	if err != nil {
		fail(err)
	}

	filePath := os.Args[1]

	// ------------------------------------------------------------
	// Determine the Vault from the Markdown file path.
	//
	// We first search upward from the file location for a
	// directory containing .obsidian.
	//
	// If no Vault is found, we fall back to the configured
	// DefaultVault, preserving the original behavior.
	// ------------------------------------------------------------

	vault, err := vaultresolver.FindForFile(filePath)

	if err != nil {
		if !errors.Is(err, vaultresolver.ErrVaultNotFound) {
			fail(err)
		}

		// No .obsidian directory was found.
		// Preserve the original 0.1.5 behavior by using DefaultVault.
		if cfg.DefaultVault.Name == "" || cfg.DefaultVault.Path == "" {
			fail(errors.New("default Vault is not configured"))
		}

		vault.Name = cfg.DefaultVault.Name
		vault.Path = cfg.DefaultVault.Path
	}

	if vault.Name == "" || vault.Path == "" {
		fail(errors.New("resolved Vault is not configured"))
	}

	// ------------------------------------------------------------
	// Create / reuse the external-file symlink inside the
	// resolved Vault.
	// ------------------------------------------------------------

	manager, err := symlink.New(
		vault.Path,
		cfg.MirrorFolder,
	)
	if err != nil {
		fail(err)
	}

	link, err := manager.Ensure(filePath)
	if err != nil {
		fail(err)
	}

	// ------------------------------------------------------------
	// Open the link in the resolved Vault.
	// ------------------------------------------------------------

	uri, err := obsidian.BuildOpenURI(
		vault.Name,
		vault.Path,
		link.Path,
	)
	if err != nil {
		fail(err)
	}

	if err := launcher.Open(uri); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

func managementService() management.Service {
	configPath, err := config.DefaultPath()
	if err != nil {
		fail(err)
	}

	associationRegistry, err := registry.New()
	if err != nil {
		fail(err)
	}

	return management.Service{
		Version:    version,
		ConfigPath: configPath,
		Registry:   associationRegistry,
	}
}

func writeJSON(value any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(value); err != nil {
		fail(err)
	}
}

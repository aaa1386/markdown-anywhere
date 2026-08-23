package management

import (
	"errors"
	"fmt"
	"os"

	"github.com/zhangcongke/markdown-anywhere/companion/internal/config"
	"github.com/zhangcongke/markdown-anywhere/companion/internal/registry"
	"github.com/zhangcongke/markdown-anywhere/companion/internal/symlink"
)

var ErrVaultNotConfigured = errors.New("default Vault is not configured")

type RegistryStatusReader interface {
	Status() (registry.AssociationStatus, error)
}

type Service struct {
	Version    string
	ConfigPath string
	Registry   RegistryStatusReader
}

type StatusReport struct {
	Version           string       `json:"version"`
	ConfigPath        string       `json:"configPath"`
	ConfigExists      bool         `json:"configExists"`
	DefaultVault      config.Vault `json:"defaultVault"`
	MirrorFolder      string       `json:"mirrorFolder"`
	CleanupDays       int          `json:"cleanupDays"`
	Registered        bool         `json:"registered"`
	RegisteredCommand string       `json:"registeredCommand,omitempty"`
}

type LinksReport struct {
	DefaultVault config.Vault       `json:"defaultVault"`
	MirrorFolder string             `json:"mirrorFolder"`
	MirrorPath   string             `json:"mirrorPath"`
	Active       int                `json:"active"`
	Invalid      int                `json:"invalid"`
	Links        []symlink.LinkInfo `json:"links"`
}

type CleanupReport struct {
	Removed []string `json:"removed"`
	Count   int      `json:"count"`
}

func (s Service) Status() (StatusReport, error) {
	cfg, err := config.LoadOrDefault(s.ConfigPath)
	if err != nil {
		return StatusReport{}, err
	}
	configExists, err := fileExists(s.ConfigPath)
	if err != nil {
		return StatusReport{}, err
	}
	if s.Registry == nil {
		return StatusReport{}, errors.New("registry status reader is not configured")
	}
	association, err := s.Registry.Status()
	if err != nil {
		return StatusReport{}, err
	}
	return StatusReport{
		Version:           s.Version,
		ConfigPath:        s.ConfigPath,
		ConfigExists:      configExists,
		DefaultVault:      cfg.DefaultVault,
		MirrorFolder:      cfg.MirrorFolder,
		CleanupDays:       cfg.CleanupDays,
		Registered:        association.Registered,
		RegisteredCommand: association.Command,
	}, nil
}

func (s Service) Links() (LinksReport, error) {
	cfg, err := s.loadConfiguredVault()
	if err != nil {
		return LinksReport{}, err
	}
	manager, err := symlink.New(cfg.DefaultVault.Path, cfg.MirrorFolder)
	if err != nil {
		return LinksReport{}, err
	}
	links, err := manager.List()
	if err != nil {
		return LinksReport{}, err
	}
	report := LinksReport{
		DefaultVault: cfg.DefaultVault,
		MirrorFolder: cfg.MirrorFolder,
		MirrorPath:   manager.MirrorRoot,
		Links:        links,
	}
	for _, link := range links {
		if link.Valid {
			report.Active++
		} else {
			report.Invalid++
		}
	}
	return report, nil
}

func (s Service) Cleanup() (CleanupReport, error) {
	cfg, err := s.loadConfiguredVault()
	if err != nil {
		return CleanupReport{}, err
	}
	manager, err := symlink.New(cfg.DefaultVault.Path, cfg.MirrorFolder)
	if err != nil {
		return CleanupReport{}, err
	}
	removed, err := manager.CleanupInvalid()
	if err != nil {
		return CleanupReport{}, err
	}
	if removed == nil {
		removed = []string{}
	}
	return CleanupReport{Removed: removed, Count: len(removed)}, nil
}

func (s Service) loadConfiguredVault() (config.Config, error) {
	cfg, err := config.LoadOrDefault(s.ConfigPath)
	if err != nil {
		return config.Config{}, err
	}
	if cfg.DefaultVault.Name == "" || cfg.DefaultVault.Path == "" {
		return config.Config{}, ErrVaultNotConfigured
	}
	return cfg, nil
}

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect config file: %w", err)
	}
	if info.IsDir() {
		return false, fmt.Errorf("config path is a directory: %s", path)
	}
	return true, nil
}

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	CurrentVersion   = 1
	AppDirectoryName = "MarkownAnywhere"
	ConfigFileName   = "config.json"
	DefaultMirror    = "_external-open"
	DefaultCleanup   = 7
)

type Vault struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Config struct {
	Version      int    `json:"version"`
	DefaultVault Vault  `json:"defaultVault"`
	MirrorFolder string `json:"mirrorFolder"`
	CleanupDays  int    `json:"cleanupDays"`
}

func Default() Config {
	return Config{
		Version:      CurrentVersion,
		MirrorFolder: DefaultMirror,
		CleanupDays:  DefaultCleanup,
	}
}

func PathFromAppData(appData string) (string, error) {
	if strings.TrimSpace(appData) == "" {
		return "", errors.New("app data directory must not be empty")
	}
	return filepath.Join(appData, AppDirectoryName, ConfigFileName), nil
}

func DefaultPath() (string, error) {
	appData, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find user config directory: %w", err)
	}
	return PathFromAppData(appData)
}

func LoadOrDefault(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (c Config) Validate() error {
	if c.Version != CurrentVersion {
		return fmt.Errorf("unsupported config version %d", c.Version)
	}
	if c.MirrorFolder == "" {
		return errors.New("mirror folder must not be empty")
	}
	if filepath.IsAbs(c.MirrorFolder) {
		return errors.New("mirror folder must be relative to the Vault")
	}
	cleanMirror := filepath.Clean(c.MirrorFolder)
	if cleanMirror == "." || cleanMirror == ".." || strings.HasPrefix(cleanMirror, ".."+string(filepath.Separator)) {
		return errors.New("mirror folder must stay inside the Vault")
	}
	if c.CleanupDays < 0 {
		return errors.New("cleanup days must not be negative")
	}
	return nil
}

func (c *Config) applyDefaults() {
	if c.Version == 0 {
		c.Version = CurrentVersion
	}
	if c.MirrorFolder == "" {
		c.MirrorFolder = DefaultMirror
	}
	if c.CleanupDays == 0 {
		c.CleanupDays = DefaultCleanup
	}
}

package registry

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	ProgID             = "MarkownAnywhere.Markdown"
	DisplayName        = "Markown Anywhere Markdown"
	MarkdownExtension  = ".md"
	MarkdownExtensions = ".markdown"
	classesRoot        = `Software\Classes`
	backupRoot         = `Software\MarkownAnywhere\RegistryBackup`
)

var ErrEmptyExecutable = errors.New("executable path must not be empty")

type Value struct {
	Exists    bool
	KeyExists bool
	Data      string
}

type AssociationStatus struct {
	Registered bool
	Command    string
}

// Store is the small registry surface needed by the association manager.
// The production implementation is backed by HKCU; tests use an in-memory
// implementation so registration behavior is verified without changing the
// user's registry.
type Store interface {
	Read(keyPath, valueName string) (Value, error)
	Set(keyPath, valueName, value string) error
	DeleteValue(keyPath, valueName string) error
	DeleteTree(keyPath string) error
}

type Registry struct {
	store Store
}

type association struct {
	id        string
	keyPath   string
	value     string
	valueName string
}

var associations = []association{
	{id: "extension-md", keyPath: joinKey(classesRoot, MarkdownExtension), valueName: "", value: ProgID},
	{id: "extension-markdown", keyPath: joinKey(classesRoot, MarkdownExtensions), valueName: "", value: ProgID},
	{id: "progid-command", keyPath: joinKey(classesRoot, ProgID, "shell", "open", "command"), valueName: ""},
	{id: "progid-icon", keyPath: joinKey(classesRoot, ProgID, "DefaultIcon"), valueName: ""},
	{id: "progid", keyPath: joinKey(classesRoot, ProgID), valueName: "", value: DisplayName},
}

type savedValue struct {
	keyExists   bool
	valueExists bool
	value       string
}

func New() (Registry, error) {
	store, err := newStore()
	if err != nil {
		return Registry{}, err
	}
	return Registry{store: store}, nil
}

func NewWithStore(store Store) Registry {
	return Registry{store: store}
}

func (r Registry) Register(executablePath string) error {
	if strings.TrimSpace(executablePath) == "" {
		return ErrEmptyExecutable
	}
	executablePath, err := filepath.Abs(executablePath)
	if err != nil {
		return fmt.Errorf("normalize executable path: %w", err)
	}

	for _, item := range associations {
		if err := r.saveBackup(item); err != nil {
			return err
		}
	}

	command := commandFor(executablePath)
	icon := executablePath + ",-1"
	for _, item := range associations {
		value := item.value
		switch item.id {
		case "progid-command":
			value = command
		case "progid-icon":
			value = icon
		}
		if err := r.store.Set(item.keyPath, item.valueName, value); err != nil {
			return fmt.Errorf("set registry value %s: %w", item.keyPath, err)
		}
	}
	return nil
}

func (r Registry) Unregister() error {
	foundBackup := false
	for _, item := range associations {
		saved, found, err := r.loadBackup(item)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		foundBackup = true
		if saved.valueExists {
			if err := r.store.Set(item.keyPath, item.valueName, saved.value); err != nil {
				return fmt.Errorf("restore registry value %s: %w", item.keyPath, err)
			}
		} else if err := r.store.DeleteValue(item.keyPath, item.valueName); err != nil {
			return fmt.Errorf("remove registry value %s: %w", item.keyPath, err)
		}
		if !saved.keyExists {
			if err := r.store.DeleteTree(item.keyPath); err != nil {
				return fmt.Errorf("remove created registry key %s: %w", item.keyPath, err)
			}
		}
	}
	if !foundBackup {
		return nil
	}

	if err := r.store.DeleteTree(backupRoot); err != nil {
		return fmt.Errorf("remove registry backup: %w", err)
	}
	return nil
}

func (r Registry) Status() (AssociationStatus, error) {
	md, err := r.store.Read(joinKey(classesRoot, MarkdownExtension), "")
	if err != nil {
		return AssociationStatus{}, fmt.Errorf("read %s association: %w", MarkdownExtension, err)
	}
	markdown, err := r.store.Read(joinKey(classesRoot, MarkdownExtensions), "")
	if err != nil {
		return AssociationStatus{}, fmt.Errorf("read %s association: %w", MarkdownExtensions, err)
	}
	command, err := r.store.Read(joinKey(classesRoot, ProgID, "shell", "open", "command"), "")
	if err != nil {
		return AssociationStatus{}, fmt.Errorf("read ProgID command: %w", err)
	}
	return AssociationStatus{
		Registered: md.Exists && md.Data == ProgID && markdown.Exists && markdown.Data == ProgID && command.Exists,
		Command:    command.Data,
	}, nil
}

func (r Registry) saveBackup(item association) error {
	backupKey := joinKey(backupRoot, item.id)
	marker, err := r.store.Read(backupKey, "keyExists")
	if err != nil {
		return fmt.Errorf("read registry backup %s: %w", item.id, err)
	}
	if marker.Exists {
		return nil
	}

	current, err := r.store.Read(item.keyPath, item.valueName)
	if err != nil {
		return fmt.Errorf("read existing registry value %s: %w", item.keyPath, err)
	}
	if err := r.store.Set(backupKey, "keyExists", boolString(current.KeyExists)); err != nil {
		return fmt.Errorf("save registry backup %s: %w", item.id, err)
	}
	if err := r.store.Set(backupKey, "valueExists", boolString(current.Exists)); err != nil {
		return fmt.Errorf("save registry backup %s: %w", item.id, err)
	}
	if current.Exists {
		if err := r.store.Set(backupKey, "value", current.Data); err != nil {
			return fmt.Errorf("save registry backup value %s: %w", item.id, err)
		}
	}
	return nil
}

func (r Registry) loadBackup(item association) (savedValue, bool, error) {
	backupKey := joinKey(backupRoot, item.id)
	keyMarker, err := r.store.Read(backupKey, "keyExists")
	if err != nil {
		return savedValue{}, false, fmt.Errorf("read registry backup %s: %w", item.id, err)
	}
	if !keyMarker.Exists {
		return savedValue{}, false, nil
	}
	keyExists, err := parseBool(keyMarker.Data)
	if err != nil {
		return savedValue{}, false, fmt.Errorf("decode registry backup %s: %w", item.id, err)
	}

	valueMarker, err := r.store.Read(backupKey, "valueExists")
	if err != nil || !valueMarker.Exists {
		if err == nil {
			err = errors.New("missing valueExists marker")
		}
		return savedValue{}, false, fmt.Errorf("read registry backup value marker %s: %w", item.id, err)
	}
	valueExists, err := parseBool(valueMarker.Data)
	if err != nil {
		return savedValue{}, false, fmt.Errorf("decode registry backup value marker %s: %w", item.id, err)
	}

	saved := savedValue{keyExists: keyExists, valueExists: valueExists}
	if valueExists {
		value, err := r.store.Read(backupKey, "value")
		if err != nil || !value.Exists {
			if err == nil {
				err = errors.New("missing value")
			}
			return savedValue{}, false, fmt.Errorf("read registry backup value %s: %w", item.id, err)
		}
		saved.value = value.Data
	}
	return saved, true, nil
}

func commandFor(executablePath string) string {
	return fmt.Sprintf(`"%s" "%%1"`, executablePath)
}

func boolString(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func parseBool(value string) (bool, error) {
	switch value {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, fmt.Errorf("invalid boolean marker %q", value)
	}
}

func joinKey(parts ...string) string {
	return strings.Join(parts, `\`)
}

package registry

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

type fakeStore struct {
	keys map[string]map[string]string
}

func newFakeStore() *fakeStore {
	return &fakeStore{keys: make(map[string]map[string]string)}
}

func (s *fakeStore) Read(keyPath, valueName string) (Value, error) {
	values, exists := s.keys[keyPath]
	if !exists {
		return Value{}, nil
	}
	value, valueExists := values[valueName]
	return Value{Exists: valueExists, KeyExists: true, Data: value}, nil
}

func (s *fakeStore) Set(keyPath, valueName, value string) error {
	if s.keys[keyPath] == nil {
		s.keys[keyPath] = make(map[string]string)
	}
	s.keys[keyPath][valueName] = value
	return nil
}

func (s *fakeStore) DeleteValue(keyPath, valueName string) error {
	values, exists := s.keys[keyPath]
	if !exists {
		return nil
	}
	delete(values, valueName)
	return nil
}

func (s *fakeStore) DeleteTree(keyPath string) error {
	for key := range s.keys {
		if key == keyPath || strings.HasPrefix(key, keyPath+`\`) {
			delete(s.keys, key)
		}
	}
	return nil
}

func (s *fakeStore) put(keyPath, valueName, value string) {
	_ = s.Set(keyPath, valueName, value)
}

func (s *fakeStore) get(keyPath, valueName string) (string, bool) {
	value, ok := s.keys[keyPath][valueName]
	return value, ok
}

func TestRegisterAndUnregisterRestoresPreviousValues(t *testing.T) {
	store := newFakeStore()
	store.put(joinKey(classesRoot, MarkdownExtension), "", "Obsidian.Markdown")
	store.put(joinKey(classesRoot, ProgID), "", "Previous owner")
	store.put(joinKey(classesRoot, ProgID, "shell", "open", "command"), "", `"C:\Old.exe" "%1"`)
	registry := NewWithStore(store)

	if err := registry.Register(`C:\Tools\MarkownAnywhere.exe`); err != nil {
		t.Fatal(err)
	}
	status, err := registry.Status()
	if err != nil {
		t.Fatal(err)
	}
	if !status.Registered || status.Command != `"C:\Tools\MarkownAnywhere.exe" "%1"` {
		t.Fatalf("unexpected registered status: %+v", status)
	}
	assertValue(t, store, joinKey(classesRoot, MarkdownExtension), "", ProgID)
	assertValue(t, store, joinKey(classesRoot, MarkdownExtensions), "", ProgID)
	assertValue(t, store, joinKey(classesRoot, ProgID), "", DisplayName)
	assertValue(t, store, joinKey(classesRoot, ProgID, "shell", "open", "command"), "", `"C:\Tools\MarkownAnywhere.exe" "%1"`)
	assertValue(t, store, joinKey(classesRoot, ProgID, "DefaultIcon"), "", `C:\Tools\MarkownAnywhere.exe,-1`)

	if err := registry.Register(`C:\Tools\NewLocation.exe`); err != nil {
		t.Fatal(err)
	}
	assertValue(t, store, joinKey(classesRoot, ProgID, "shell", "open", "command"), "", `"C:\Tools\NewLocation.exe" "%1"`)

	if err := registry.Unregister(); err != nil {
		t.Fatal(err)
	}
	status, err = registry.Status()
	if err != nil {
		t.Fatal(err)
	}
	if status.Registered {
		t.Fatalf("expected unregistered status: %+v", status)
	}
	assertValue(t, store, joinKey(classesRoot, MarkdownExtension), "", "Obsidian.Markdown")
	assertMissing(t, store, joinKey(classesRoot, MarkdownExtensions), "")
	assertValue(t, store, joinKey(classesRoot, ProgID), "", "Previous owner")
	assertValue(t, store, joinKey(classesRoot, ProgID, "shell", "open", "command"), "", `"C:\Old.exe" "%1"`)
	assertMissing(t, store, joinKey(classesRoot, ProgID, "DefaultIcon"), "")
	if _, exists := store.keys[backupRoot]; exists {
		t.Fatalf("backup root still exists: %#v", store.keys)
	}
}

func TestUnregisterRemovesOnlyOurCreatedKeys(t *testing.T) {
	store := newFakeStore()
	registry := NewWithStore(store)
	if err := registry.Register(`C:\Tools\MarkownAnywhere.exe`); err != nil {
		t.Fatal(err)
	}
	if err := registry.Unregister(); err != nil {
		t.Fatal(err)
	}
	for _, item := range associations {
		if _, exists := store.keys[item.keyPath]; exists {
			t.Errorf("created key still exists: %s", item.keyPath)
		}
	}
	for key := range store.keys {
		if strings.Contains(key, "Applications\\Obsidian.exe") {
			t.Errorf("unexpected Obsidian application key touched: %s", key)
		}
	}
}

func TestRegisterRejectsEmptyExecutable(t *testing.T) {
	registry := NewWithStore(newFakeStore())
	if err := registry.Register(" "); !errors.Is(err, ErrEmptyExecutable) {
		t.Fatalf("expected empty executable error, got %v", err)
	}
}

func assertValue(t *testing.T, store *fakeStore, keyPath, valueName, want string) {
	t.Helper()
	got, ok := store.get(keyPath, valueName)
	if !ok || got != want {
		t.Fatalf("%s[%q] = %q, %v; want %q", keyPath, valueName, got, ok, want)
	}
}

func assertMissing(t *testing.T, store *fakeStore, keyPath, valueName string) {
	t.Helper()
	if _, ok := store.get(keyPath, valueName); ok {
		t.Fatalf("%s[%q] unexpectedly exists", keyPath, valueName)
	}
}

func TestKeyPathsUseWindowsRegistrySeparators(t *testing.T) {
	if filepath.Separator != '\\' {
		t.Skip("path separator assertion is Windows-specific")
	}
	if !strings.Contains(joinKey(classesRoot, ProgID), `Software\Classes\`) {
		t.Fatal("registry key path is malformed")
	}
}

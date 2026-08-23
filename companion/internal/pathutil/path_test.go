package pathutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeAbsoluteIsStable(t *testing.T) {
	first, err := NormalizeAbsolute(".")
	if err != nil {
		t.Fatal(err)
	}
	second, err := NormalizeAbsolute(first)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("normalized path changed: %q != %q", first, second)
	}
	if !filepath.IsAbs(first) {
		t.Fatalf("path is not absolute: %q", first)
	}
}

func TestNormalizeAbsoluteIsStableAfterTargetRemoval(t *testing.T) {
	source := filepath.Join(t.TempDir(), "note.md")
	if err := os.WriteFile(source, []byte("# note\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	before, err := NormalizeAbsolute(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	after, err := NormalizeAbsolute(source)
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatalf("normalized path changed after target removal: %q != %q", before, after)
	}
}

func TestRelativeToRejectsOutsidePath(t *testing.T) {
	_, err := RelativeTo(filepath.Join(t.TempDir(), "vault"), filepath.Join(t.TempDir(), "note.md"))
	if err == nil {
		t.Fatal("expected outside path to be rejected")
	}
}

func TestIsWithinTreatsDifferentWindowsVolumeAsOutside(t *testing.T) {
	root := t.TempDir()
	volume := filepath.VolumeName(root)
	if volume == "" {
		t.Skip("Windows volume paths are not available on this platform")
	}

	otherVolume := "C:\\"
	if strings.EqualFold(volume, "C:") {
		otherVolume = "D:\\"
	}
	inside, err := IsWithin(filepath.Join(otherVolume, "my-data", "content.md"), root)
	if err != nil {
		t.Fatal(err)
	}
	if inside {
		t.Fatal("expected a path on another volume to be outside the Vault")
	}
}

func TestLinkNameContainsHashAndFilename(t *testing.T) {
	name, err := LinkName(filepath.Join(t.TempDir(), "中文 note.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !IsManagedLinkName(name) {
		t.Fatalf("link name is not recognized: %q", name)
	}
	if !strings.HasSuffix(name, "-中文 note.md") {
		t.Fatalf("link name lost filename: %q", name)
	}
}

func TestIsManagedLinkName(t *testing.T) {
	valid := "0123456789ab-note.md"
	for _, name := range []string{valid, "0123456789ab-note.markdown"} {
		if !IsManagedLinkName(name) {
			t.Errorf("expected managed name: %q", name)
		}
	}
	for _, name := range []string{"note.md", "0123456789AB-note.md", "0123456789ab-"} {
		if IsManagedLinkName(name) {
			t.Errorf("expected unmanaged name: %q", name)
		}
	}
}

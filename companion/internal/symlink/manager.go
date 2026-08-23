package symlink

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/zhangcongke/markdown-anywhere/companion/internal/pathutil"
)

var (
	ErrSourceMissing = errors.New("source file does not exist")
	ErrLinkCollision = errors.New("managed link path is already used by another target")
)

type Link struct {
	Path     string
	Target   string
	External bool
	Existing bool
}

type LinkInfo struct {
	Path   string `json:"path"`
	Target string `json:"target"`
	Valid  bool   `json:"valid"`
}

type Manager struct {
	VaultRoot  string
	MirrorRoot string
}

func New(vaultRoot, mirrorFolder string) (Manager, error) {
	root, err := pathutil.Absolute(vaultRoot)
	if err != nil {
		return Manager{}, err
	}
	if mirrorFolder == "" {
		return Manager{}, errors.New("mirror folder must not be empty")
	}
	if filepath.IsAbs(mirrorFolder) {
		return Manager{}, errors.New("mirror folder must be relative to the Vault")
	}
	clean := filepath.Clean(mirrorFolder)
	if clean == "." || clean == ".." || len(clean) >= 2 && clean[:2] == ".."+string(filepath.Separator) {
		return Manager{}, errors.New("mirror folder must stay inside the Vault")
	}
	return Manager{
		VaultRoot:  root,
		MirrorRoot: filepath.Join(root, clean),
	}, nil
}

// Ensure returns the original path for Vault files, or creates/reuses a
// deterministic symlink for an external file.
func (m Manager) Ensure(source string) (Link, error) {
	sourceAbs, err := pathutil.Absolute(source)
	if err != nil {
		return Link{}, err
	}
	info, err := os.Stat(sourceAbs)
	if errors.Is(err, os.ErrNotExist) {
		return Link{}, fmt.Errorf("%w: %s", ErrSourceMissing, sourceAbs)
	}
	if err != nil {
		return Link{}, fmt.Errorf("stat source: %w", err)
	}
	if !info.Mode().IsRegular() {
		return Link{}, fmt.Errorf("source is not a regular file: %s", sourceAbs)
	}

	inside, err := pathutil.IsWithin(sourceAbs, m.VaultRoot)
	if err != nil {
		return Link{}, err
	}
	if inside {
		return Link{Path: sourceAbs, Target: sourceAbs, Existing: true}, nil
	}

	name, err := pathutil.LinkName(sourceAbs)
	if err != nil {
		return Link{}, err
	}
	if err := os.MkdirAll(m.MirrorRoot, 0o755); err != nil {
		return Link{}, fmt.Errorf("create mirror directory: %w", err)
	}
	linkPath := filepath.Join(m.MirrorRoot, name)

	linkInfo, err := os.Lstat(linkPath)
	if err == nil {
		if linkInfo.Mode()&os.ModeSymlink == 0 {
			return Link{}, fmt.Errorf("%w: %s", ErrLinkCollision, linkPath)
		}
		resolved, err := resolveTarget(linkPath)
		if err != nil {
			return Link{}, fmt.Errorf("read existing link: %w", err)
		}
		expected, err := pathutil.NormalizeAbsolute(sourceAbs)
		if err != nil {
			return Link{}, err
		}
		if resolved != expected {
			return Link{}, fmt.Errorf("%w: %s", ErrLinkCollision, linkPath)
		}
		return Link{Path: linkPath, Target: sourceAbs, External: true, Existing: true}, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return Link{}, fmt.Errorf("inspect link path: %w", err)
	}

	if err := os.Symlink(sourceAbs, linkPath); err != nil {
		return Link{}, fmt.Errorf("create symlink: %w", err)
	}
	return Link{Path: linkPath, Target: sourceAbs, External: true}, nil
}

func (m Manager) List() ([]LinkInfo, error) {
	entries, err := os.ReadDir(m.MirrorRoot)
	if errors.Is(err, os.ErrNotExist) {
		return []LinkInfo{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read mirror directory: %w", err)
	}

	links := make([]LinkInfo, 0, len(entries))
	for _, entry := range entries {
		if !pathutil.IsManagedLinkName(entry.Name()) {
			continue
		}
		linkPath := filepath.Join(m.MirrorRoot, entry.Name())
		info, err := os.Lstat(linkPath)
		if err != nil {
			return nil, fmt.Errorf("inspect link %s: %w", linkPath, err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		target, err := resolveTarget(linkPath)
		if err != nil {
			return nil, fmt.Errorf("read link %s: %w", linkPath, err)
		}
		owned, err := ownsLink(entry.Name(), target)
		if err != nil {
			return nil, err
		}
		if !owned {
			continue
		}
		_, statErr := os.Stat(target)
		if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return nil, fmt.Errorf("check link target %s: %w", target, statErr)
		}
		links = append(links, LinkInfo{
			Path:   linkPath,
			Target: target,
			Valid:  statErr == nil,
		})
	}

	sort.Slice(links, func(i, j int) bool { return links[i].Path < links[j].Path })
	return links, nil
}

func (m Manager) CleanupInvalid() ([]string, error) {
	links, err := m.List()
	if err != nil {
		return nil, err
	}
	removed := make([]string, 0)
	for _, link := range links {
		if link.Valid {
			continue
		}
		if err := os.Remove(link.Path); err != nil {
			return removed, fmt.Errorf("remove invalid link %s: %w", link.Path, err)
		}
		removed = append(removed, link.Path)
	}
	return removed, nil
}

func resolveTarget(linkPath string) (string, error) {
	raw, err := os.Readlink(linkPath)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(raw) {
		raw = filepath.Join(filepath.Dir(linkPath), raw)
	}
	return pathutil.NormalizeAbsolute(raw)
}

func ownsLink(name, target string) (bool, error) {
	prefix, err := pathutil.HashPrefix(target)
	if err != nil {
		return false, err
	}
	return len(name) > pathutil.HashPrefixLength && name[:pathutil.HashPrefixLength] == prefix, nil
}

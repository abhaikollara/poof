package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Registry represents the on-disk registry file.
type Registry struct {
	Version         int      `json:"version"`
	Entries         []Entry  `json:"entries"`
	AllowedPrefixes []string `json:"allowed_prefixes,omitempty"`
}

// Entry is a single tracked temporary directory.
type Entry struct {
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// registryDir returns the directory for the registry file.
func registryDir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "mehdir")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	return filepath.Join(home, ".config", "mehdir")
}

// RegistryPath returns the path to the registry JSON file.
func RegistryPath() string {
	return filepath.Join(registryDir(), "registry.json")
}

func lockPath() string {
	return RegistryPath() + ".lock"
}

// Lock acquires an exclusive file lock on the registry, blocking up to 5s.
// Returns the lock file, which must be passed to Unlock.
func Lock() (*os.File, error) {
	dir := registryDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("creating registry dir: %w", err)
	}

	f, err := os.OpenFile(lockPath(), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("opening lock file: %w", err)
	}

	// Try to acquire with a timeout via non-blocking attempts.
	deadline := time.Now().Add(5 * time.Second)
	for {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return f, nil
		}
		if time.Now().After(deadline) {
			f.Close()
			return nil, fmt.Errorf("could not acquire registry lock after 5s (another mehdir process running?)")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// Unlock releases the lock and closes the file.
func Unlock(f *os.File) {
	if f == nil {
		return
	}
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	f.Close()
}

// Load reads the registry from disk. Returns an empty registry if the file doesn't exist.
func Load() (*Registry, error) {
	path := RegistryPath()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Registry{Version: 1}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading registry: %w", err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("registry corrupt: %w", err)
	}
	return &reg, nil
}

// Save atomically writes the registry to disk via a temp file + rename.
func Save(reg *Registry) error {
	path := RegistryPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating registry dir: %w", err)
	}

	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling registry: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("writing temp registry: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming registry: %w", err)
	}
	return nil
}

// FindByPath finds an entry by its absolute path.
func (r *Registry) FindByPath(path string) (int, *Entry) {
	for i := range r.Entries {
		if r.Entries[i].Path == path {
			return i, &r.Entries[i]
		}
	}
	return -1, nil
}

// RemoveIndex removes the entry at index i.
func (r *Registry) RemoveIndex(i int) {
	r.Entries = append(r.Entries[:i], r.Entries[i+1:]...)
}

// AddAllowedPrefix registers a directory as an allowed parent for mehdir dirs.
// Deduplicates against existing prefixes.
func (r *Registry) AddAllowedPrefix(dir string) {
	cleaned := filepath.Clean(dir)
	for _, p := range r.AllowedPrefixes {
		if filepath.Clean(p) == cleaned {
			return
		}
	}
	r.AllowedPrefixes = append(r.AllowedPrefixes, cleaned)
}

// SafeToDelete checks whether a path is safe to RemoveAll.
// extraPrefixes are additional allowed parent directories (from the registry).
func SafeToDelete(path string, extraPrefixes []string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("path is not absolute: %s", path)
	}

	home, _ := os.UserHomeDir()
	forbidden := []string{"/", "/tmp", "/var/tmp"}
	if home != "" {
		forbidden = append(forbidden, home)
	}
	cleaned := filepath.Clean(path)
	for _, f := range forbidden {
		if cleaned == filepath.Clean(f) {
			return fmt.Errorf("refusing to delete protected path: %s", path)
		}
	}

	tmpDir := os.TempDir()
	allowedPrefixes := []string{
		filepath.Clean(tmpDir) + string(os.PathSeparator),
		"/tmp/",
		"/var/tmp/",
	}
	for _, ep := range extraPrefixes {
		allowedPrefixes = append(allowedPrefixes, filepath.Clean(ep)+string(os.PathSeparator))
	}
	// Deduplicate.
	seen := map[string]bool{}
	var prefixes []string
	for _, p := range allowedPrefixes {
		if !seen[p] {
			seen[p] = true
			prefixes = append(prefixes, p)
		}
	}

	inAllowed := false
	for _, prefix := range prefixes {
		if strings.HasPrefix(cleaned+string(os.PathSeparator), prefix) || strings.HasPrefix(cleaned, prefix) {
			inAllowed = true
			break
		}
	}
	if !inAllowed {
		return fmt.Errorf("path not under an allowed prefix: %s", path)
	}

	return nil
}

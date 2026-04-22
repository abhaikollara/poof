package sweep

import (
	"fmt"
	"os"
	"time"

	"abhai.dev/poof/internal/registry"
)

// Run removes expired entries from the registry, deleting their directories.
// Errors during deletion are logged to stderr but do not cause failure.
func Run(reg *registry.Registry, now time.Time) {
	var kept []registry.Entry
	for _, e := range reg.Entries {
		if now.After(e.ExpiresAt) {
			if err := safeRemove(e.Path, reg.AllowedPrefixes); err != nil {
				fmt.Fprintf(os.Stderr, "poof: sweep: skipping %s: %v\n", e.Path, err)
			}
			continue
		}
		kept = append(kept, e)
	}
	reg.Entries = kept
}

// GC removes entries whose directories no longer exist on disk.
func GC(reg *registry.Registry) int {
	var kept []registry.Entry
	removed := 0
	for _, e := range reg.Entries {
		if _, err := os.Stat(e.Path); os.IsNotExist(err) {
			removed++
			continue
		}
		kept = append(kept, e)
	}
	reg.Entries = kept
	return removed
}

func safeRemove(path string, extraPrefixes []string) error {
	if err := registry.SafeToDelete(path, extraPrefixes); err != nil {
		return err
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("removing %s: %w", path, err)
	}
	return nil
}

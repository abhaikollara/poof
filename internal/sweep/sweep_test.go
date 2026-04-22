package sweep

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"abhai.dev/mehdir/internal/registry"
)

func TestSweepRemovesExpired(t *testing.T) {
	expired := filepath.Join(os.TempDir(), "mehdir-expired-test")
	alive := filepath.Join(os.TempDir(), "mehdir-alive-test")
	os.MkdirAll(expired, 0700)
	os.MkdirAll(alive, 0700)
	defer os.RemoveAll(expired)
	defer os.RemoveAll(alive)

	now := time.Now().UTC()
	reg := &registry.Registry{
		Version: 1,
		Entries: []registry.Entry{
			{Path: expired, CreatedAt: now.Add(-2 * time.Hour), ExpiresAt: now.Add(-1 * time.Hour)},
			{Path: alive, CreatedAt: now, ExpiresAt: now.Add(1 * time.Hour)},
		},
	}

	Run(reg, now)

	if len(reg.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(reg.Entries))
	}
	if reg.Entries[0].Path != alive {
		t.Errorf("expected alive entry to remain, got %q", reg.Entries[0].Path)
	}

	if _, err := os.Stat(expired); !os.IsNotExist(err) {
		t.Error("expired directory was not removed")
	}
	if _, err := os.Stat(alive); err != nil {
		t.Error("alive directory should still exist")
	}
}

func TestGCRemovesOrphans(t *testing.T) {
	existing := filepath.Join(os.TempDir(), "mehdir-gc-exists")
	os.MkdirAll(existing, 0700)
	defer os.RemoveAll(existing)

	reg := &registry.Registry{
		Version: 1,
		Entries: []registry.Entry{
			{Path: existing},
			{Path: "/tmp/mehdir-gc-gone-does-not-exist"},
		},
	}

	removed := GC(reg)
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}
	if len(reg.Entries) != 1 || reg.Entries[0].Path != existing {
		t.Errorf("unexpected entries: %+v", reg.Entries)
	}
}

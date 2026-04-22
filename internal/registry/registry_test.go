package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSafeToDelete(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		extraPrefixes []string
		wantErr       bool
	}{
		{"valid temp path", "/tmp/mehdir-abc123", nil, false},
		{"valid var tmp", "/var/tmp/mehdir-abc123", nil, false},
		{"valid temp any name", "/tmp/foobar", nil, false},
		{"relative path", "mehdir-abc123", nil, true},
		{"root path", "/", nil, true},
		{"bare tmp", "/tmp", nil, true},
		{"bare var tmp", "/var/tmp", nil, true},
		{"outside temp", "/home/user/mehdir-abc123", nil, true},
		{"home dir", func() string { h, _ := os.UserHomeDir(); return h }(), nil, true},
		{"custom prefix allowed", "/projects/work/mydir", []string{"/projects/work"}, false},
		{"custom prefix not registered", "/projects/work/mydir", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SafeToDelete(tt.path, tt.extraPrefixes)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeToDelete(%q, %v) error = %v, wantErr %v", tt.path, tt.extraPrefixes, err, tt.wantErr)
			}
		})
	}
}

func TestAtomicWrite(t *testing.T) {
	dir := t.TempDir()

	t.Setenv("XDG_CONFIG_HOME", dir)

	reg := &Registry{
		Version: 1,
		Entries: []Entry{
			{
				Path:      "/tmp/mehdir-abc123",
				CreatedAt: time.Now().UTC(),
				ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
			},
		},
	}

	if err := Save(reg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	path := filepath.Join(dir, "mehdir", "registry.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading saved registry: %v", err)
	}

	var loaded Registry
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshaling: %v", err)
	}
	if len(loaded.Entries) != 1 || loaded.Entries[0].Path != "/tmp/mehdir-abc123" {
		t.Errorf("unexpected entries: %+v", loaded.Entries)
	}

	_, err = os.Stat(path + ".tmp")
	if !os.IsNotExist(err) {
		t.Error("temp file was not cleaned up")
	}
}

func TestFindByPath(t *testing.T) {
	reg := &Registry{
		Version: 1,
		Entries: []Entry{
			{Path: "/tmp/mehdir-foo"},
			{Path: "/tmp/mehdir-bar"},
		},
	}

	if i, e := reg.FindByPath("/tmp/mehdir-foo"); i != 0 || e.Path != "/tmp/mehdir-foo" {
		t.Errorf("expected to find foo at 0, got %d", i)
	}
	if i, e := reg.FindByPath("/tmp/mehdir-bar"); i != 1 || e.Path != "/tmp/mehdir-bar" {
		t.Errorf("expected to find bar at 1, got %d", i)
	}
	if i, _ := reg.FindByPath("/tmp/mehdir-nope"); i != -1 {
		t.Errorf("expected -1 for missing, got %d", i)
	}
}

func TestAddAllowedPrefix(t *testing.T) {
	reg := &Registry{Version: 1}
	reg.AddAllowedPrefix("/projects/work")
	reg.AddAllowedPrefix("/projects/work") // duplicate
	reg.AddAllowedPrefix("/other/dir")

	if len(reg.AllowedPrefixes) != 2 {
		t.Errorf("expected 2 prefixes, got %d: %v", len(reg.AllowedPrefixes), reg.AllowedPrefixes)
	}
}

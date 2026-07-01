package claudehome

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLinkSharedSkipsRuntimeStateAndSettings verifies that runtime-state
// entries and settings.json are never linked, so two profiles never share a
// live daemon/session, and settings.json is always regenerated (not linked)
// per profile.
func TestLinkSharedSkipsRuntimeStateAndSettings(t *testing.T) {
	realDir := t.TempDir()
	mirrorDir := t.TempDir()

	writeFile(t, filepath.Join(realDir, "settings.json"), `{}`)
	writeFile(t, filepath.Join(realDir, "CLAUDE.md"), "shared memory")
	writeFile(t, filepath.Join(realDir, "daemon.log"), "runtime log")
	if err := os.Mkdir(filepath.Join(realDir, "daemon"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(realDir, "plugins"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(realDir, "plugins", "a.json"), "plugin data")

	if err := linkShared(realDir, mirrorDir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Lstat(filepath.Join(mirrorDir, "settings.json")); err == nil {
		t.Error("settings.json must not be linked into the mirror")
	}
	if _, err := os.Lstat(filepath.Join(mirrorDir, "daemon")); err == nil {
		t.Error("runtime-state dir 'daemon' must not be linked into the mirror")
	}
	if _, err := os.Lstat(filepath.Join(mirrorDir, "daemon.log")); err == nil {
		t.Error("runtime-state file 'daemon.log' must not be linked into the mirror")
	}

	claudeMD, err := os.ReadFile(filepath.Join(mirrorDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md should be linked into the mirror: %v", err)
	}
	if string(claudeMD) != "shared memory" {
		t.Errorf("linked CLAUDE.md content mismatch: %q", claudeMD)
	}

	pluginData, err := os.ReadFile(filepath.Join(mirrorDir, "plugins", "a.json"))
	if err != nil {
		t.Fatalf("plugins dir should be linked into the mirror: %v", err)
	}
	if string(pluginData) != "plugin data" {
		t.Errorf("linked plugins content mismatch: %q", pluginData)
	}
}

// TestLinkSharedReflectsUpdatesToReal verifies that re-running linkShared
// (as Prepare does on every launch) picks up new content written to the real
// dir after the mirror already exists, since the mirror must track shared
// state rather than freeze it at first launch.
func TestLinkSharedReflectsUpdatesToReal(t *testing.T) {
	realDir := t.TempDir()
	mirrorDir := t.TempDir()
	writeFile(t, filepath.Join(realDir, "CLAUDE.md"), "v1")

	if err := linkShared(realDir, mirrorDir); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(realDir, "CLAUDE.md"), "v2")
	if err := linkShared(realDir, mirrorDir); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(mirrorDir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "v2" {
		t.Errorf("mirror did not reflect update to real dir: got %q", got)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

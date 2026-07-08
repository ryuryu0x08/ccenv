// Package claudehome builds a per-profile mirror of ~/.claude so that
// claude's background daemon and Workflow subagents (which resolve
// credentials/env from CLAUDE_CONFIG_DIR's settings.json rather than from
// the launching process's environment) see the same provider config as the
// foreground session.
//
// The mirror links every shared entry (plugins, skills, projects, memory,
// credentials, ...) back into the real ~/.claude so profiles keep one
// unified view of that state. Runtime-state entries (daemon socket,
// sessions, shell snapshots, ...) are excluded from the link set and left
// for claude to create fresh inside the mirror, so two profiles running
// concurrently never share a daemon/session that was started under a
// different profile's provider. settings.json is not linked either: it is
// regenerated from the real settings.json with only the profile's own
// provider keys applied on top.
package claudehome

import (
	"fmt"
	"os"
	"path/filepath"

	"ccenv/internal/claudecfg"
	"ccenv/internal/profile"
)

// runtimeStateEntries are ~/.claude entries tied to a running daemon/session
// that must not be shared between profiles launched concurrently. They are
// skipped during linking; claude recreates them inside the mirror as needed.
var runtimeStateEntries = map[string]bool{
	"daemon":          true,
	"daemon.log":      true,
	"sessions":        true,
	"session-data":    true,
	"session-env":     true,
	"shell-snapshots": true,
	"jobs":            true,
	"debug":           true,
	"tmp":             true,
}

// settingsFileName is excluded from linking; it is regenerated per profile.
const settingsFileName = "settings.json"

// nestedConfigDirName is a ".claude" entry that may exist inside the real
// ~/.claude. Linking it would place a ".claude" directory at the mirror root,
// turning the mirror dir itself into a project root: claude would then read
// mirror/.claude/settings.json as project-scope config (higher precedence than
// the mirror's own user-scope settings.json), silently overriding the profile.
// It is never linked.
const nestedConfigDirName = ".claude"

// Dir returns the mirror directory for the given profile name:
// ~/.ccenv/claudehome/<name>.
func Dir(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".ccenv", "claudehome", name), nil
}

// Prepare builds (or refreshes) the mirror directory for profile name so it
// can be used as CLAUDE_CONFIG_DIR, and returns its path. It links every
// shared entry from the real ~/.claude into the mirror (skipping
// runtimeStateEntries, settingsFileName, and nestedConfigDirName), then
// materializes
// settings.json with p's provider env applied on top of the real
// settings.json's other settings.
func Prepare(name string, p profile.Profile) (string, error) {
	realDir, err := claudecfg.Dir()
	if err != nil {
		return "", err
	}
	mirrorDir, err := Dir(name)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir mirror dir %s: %w", mirrorDir, err)
	}
	if err := linkShared(realDir, mirrorDir); err != nil {
		return "", err
	}

	realSettings := filepath.Join(realDir, settingsFileName)
	mirrorSettings := filepath.Join(mirrorDir, settingsFileName)
	env := p.EnvMap()
	if err := claudecfg.MaterializeProviderSettings(realSettings, mirrorSettings, env, profile.ManagedEnvKeys); err != nil {
		return "", fmt.Errorf("materialize settings for profile %q: %w", name, err)
	}
	return mirrorDir, nil
}

// linkShared re-links mirrorDir's shared entries to match realDir: every
// entry present in realDir (except runtimeStateEntries, settingsFileName, and
// nestedConfigDirName) gets a link in mirrorDir, replacing whatever was there
// before. Entries in
// mirrorDir that no longer exist in realDir are left alone (defensive; ccenv
// never wrote them there since only listed entries are ever linked).
func linkShared(realDir, mirrorDir string) error {
	entries, err := os.ReadDir(realDir)
	if err != nil {
		return fmt.Errorf("read %s: %w", realDir, err)
	}
	for _, e := range entries {
		name := e.Name()
		if runtimeStateEntries[name] || name == settingsFileName || name == nestedConfigDirName {
			continue
		}
		src := filepath.Join(realDir, name)
		dst := filepath.Join(mirrorDir, name)
		if err := relink(src, dst); err != nil {
			return fmt.Errorf("link %s: %w", name, err)
		}
	}
	return nil
}

// relink removes any existing entry at dst and re-creates it as a link to
// src, using the most capable link type the platform allows without
// elevation (see link_unix.go / link_windows.go).
func relink(src, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("remove stale %s: %w", dst, err)
	}
	return link(src, dst)
}

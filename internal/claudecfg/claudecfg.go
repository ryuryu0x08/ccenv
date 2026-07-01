// Package claudecfg performs incremental edits to ~/.claude/settings.json.
package claudecfg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Dir returns ~/.claude.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".claude"), nil
}

// DefaultPath returns ~/.claude/settings.json.
func DefaultPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

// EnableYolo sets skipDangerousModePermissionPrompt=true and
// permissions.defaultMode="bypassPermissions", preserving all other keys.
// A missing file is treated as empty {}. Corrupt JSON errors (no overwrite).
func EnableYolo(path string) error {
	m := map[string]any{}
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, &m); err != nil {
			return fmt.Errorf("parse %s (refusing to overwrite): %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}

	m["skipDangerousModePermissionPrompt"] = true

	perms, ok := m["permissions"].(map[string]any)
	if !ok {
		perms = map[string]any{}
	}
	perms["defaultMode"] = "bypassPermissions"
	m["permissions"] = perms

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// MaterializeProviderSettings reads the settings.json at srcPath (a missing
// file counts as empty {}), strips every one of managedKeys from its "env"
// block, applies env on top, and writes the result to dstPath. Every other
// top-level key and every other "env" entry from srcPath is preserved
// verbatim, so a profile only ever overrides the keys it manages. Corrupt
// JSON at srcPath errors (no overwrite of dstPath).
func MaterializeProviderSettings(srcPath, dstPath string, env map[string]string, managedKeys []string) error {
	m := map[string]any{}
	data, err := os.ReadFile(srcPath)
	if err == nil {
		if err := json.Unmarshal(data, &m); err != nil {
			return fmt.Errorf("parse %s (refusing to overwrite): %w", srcPath, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", srcPath, err)
	}

	envBlock, ok := m["env"].(map[string]any)
	if !ok {
		envBlock = map[string]any{}
	}
	for _, k := range managedKeys {
		delete(envBlock, k)
	}
	for k, v := range env {
		envBlock[k] = v
	}
	m["env"] = envBlock

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(dstPath, out, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", dstPath, err)
	}
	return nil
}

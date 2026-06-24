// Package claudecfg performs incremental edits to ~/.claude/settings.json.
package claudecfg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultPath returns ~/.claude/settings.json.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
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

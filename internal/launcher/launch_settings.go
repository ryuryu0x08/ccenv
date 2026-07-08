package launcher

import (
	"encoding/json"
	"fmt"

	"ccenv/internal/profile"
)

// providerSettingsJSON renders p's provider env as a compact settings JSON
// document ({"env": {...}}) suitable for claude's `--settings <json>` flag.
//
// `--settings` sits at command-line precedence: above local/project/user
// settings files, below only managed settings. Injecting the profile's
// provider env here (in addition to the CLAUDE_CONFIG_DIR mirror, which is
// only user scope) guarantees the profile's provider wins regardless of any
// .claude/settings.json that happens to live in the launch directory. That
// CWD-relative project-scope file is exactly what would otherwise override
// the mirror and make the effective model depend on where ccenv was run from.
func providerSettingsJSON(p profile.Profile) (string, error) {
	b, err := json.Marshal(map[string]any{"env": p.EnvMap()})
	if err != nil {
		return "", fmt.Errorf("marshal provider --settings json: %w", err)
	}
	return string(b), nil
}

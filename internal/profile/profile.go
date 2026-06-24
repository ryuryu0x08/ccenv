// Package profile defines the ccenv profile data model.
package profile

import "math"

// Profile describes one named launch configuration. Empty fields are omitted
// from TOML and skipped during environment injection.
type Profile struct {
	BaseURL           string `toml:"base_url,omitempty"`
	AuthToken         string `toml:"auth_token,omitempty"`
	ModelsURL         string `toml:"models_url,omitempty"`
	Model             string `toml:"model,omitempty"`
	AutoCompactWindow int    `toml:"auto_compact_window,omitempty"`
}

// DefaultCompactRatioPercent is the default auto-compact window as a percentage
// of the model's context length.
const DefaultCompactRatioPercent = 80.0

// CompactWindow computes the auto-compact token window from a context length and
// a percentage. A non-positive contextLen or percent yields 0 (disabled).
func CompactWindow(contextLen int, percent float64) int {
	if contextLen <= 0 || percent <= 0 {
		return 0
	}
	return int(math.Round(float64(contextLen) * percent / 100.0))
}

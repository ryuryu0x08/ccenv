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

	// Per-tier model overrides. Empty means that tier falls back to Model.
	// Setting any of these switches injection to per-tier mode (ANTHROPIC_MODEL
	// is then NOT injected, because it would override the per-tier variables).
	HaikuModel  string `toml:"haiku_model,omitempty"`
	SonnetModel string `toml:"sonnet_model,omitempty"`
	OpusModel   string `toml:"opus_model,omitempty"`
}

// HasTierModels reports whether any per-tier model override is set.
func (p Profile) HasTierModels() bool {
	return p.HaikuModel != "" || p.SonnetModel != "" || p.OpusModel != ""
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

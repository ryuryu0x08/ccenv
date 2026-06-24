// Package profile defines the ccenv profile data model.
package profile

// Profile describes one named launch configuration. Empty fields are omitted
// from TOML and skipped during environment injection.
type Profile struct {
	BaseURL           string `toml:"base_url,omitempty"`
	AuthToken         string `toml:"auth_token,omitempty"`
	ModelsURL         string `toml:"models_url,omitempty"`
	Model             string `toml:"model,omitempty"`
	AutoCompactWindow int    `toml:"auto_compact_window,omitempty"`
}

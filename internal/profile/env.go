package profile

import "strconv"

// Environment variable names a profile can inject. These are the single
// source of truth: both process-env injection (launcher.BuildEnv) and
// settings.json materialization (claudecfg.MaterializeProviderEnv) key off
// them, so a value ccenv doesn't manage can never leak between profiles.
const (
	EnvBaseURL           = "ANTHROPIC_BASE_URL"
	EnvAuthToken         = "ANTHROPIC_AUTH_TOKEN"
	EnvAPIKey            = "ANTHROPIC_API_KEY"
	EnvModel             = "ANTHROPIC_MODEL"
	EnvHaikuModel        = "ANTHROPIC_DEFAULT_HAIKU_MODEL"
	EnvSonnetModel       = "ANTHROPIC_DEFAULT_SONNET_MODEL"
	EnvOpusModel         = "ANTHROPIC_DEFAULT_OPUS_MODEL"
	EnvAutoCompactWindow = "CLAUDE_CODE_AUTO_COMPACT_WINDOW"
)

// ManagedEnvKeys lists every key EnvMap can ever produce.
var ManagedEnvKeys = []string{
	EnvBaseURL, EnvAuthToken, EnvAPIKey, EnvModel,
	EnvHaikuModel, EnvSonnetModel, EnvOpusModel, EnvAutoCompactWindow,
}

// EnvMap returns the KEY -> VALUE pairs this profile injects. Empty fields
// produce no entries (so a fully-empty "plan" profile returns an empty map).
func (p Profile) EnvMap() map[string]string {
	env := map[string]string{}
	if p.BaseURL != "" {
		env[EnvBaseURL] = p.BaseURL
	}
	if p.AuthToken != "" {
		// Custom endpoint (base_url set) uses a Bearer token; the official
		// endpoint (base_url empty) uses the console API key (x-api-key).
		if p.BaseURL != "" {
			env[EnvAuthToken] = p.AuthToken
		} else {
			env[EnvAPIKey] = p.AuthToken
		}
	}
	appendModelEnv(env, p)
	if p.AutoCompactWindow > 0 {
		env[EnvAutoCompactWindow] = strconv.Itoa(p.AutoCompactWindow)
	}
	return env
}

// appendModelEnv injects the model variables.
//
//   - No per-tier overrides: inject ANTHROPIC_MODEL plus all three
//     ANTHROPIC_DEFAULT_*_MODEL set to Model (backward-compatible single-model
//     behavior).
//   - Any per-tier override set: do NOT inject ANTHROPIC_MODEL (it would force a
//     single model and override the tiers). Inject the three
//     ANTHROPIC_DEFAULT_*_MODEL, each using its tier override when set and
//     falling back to Model otherwise.
//
// If Model is empty and no tier overrides are set, nothing is injected.
func appendModelEnv(env map[string]string, p Profile) {
	if !p.HasTierModels() {
		if p.Model != "" {
			env[EnvModel] = p.Model
			env[EnvHaikuModel] = p.Model
			env[EnvSonnetModel] = p.Model
			env[EnvOpusModel] = p.Model
		}
		return
	}
	tier := func(override string) string {
		if override != "" {
			return override
		}
		return p.Model
	}
	if v := tier(p.HaikuModel); v != "" {
		env[EnvHaikuModel] = v
	}
	if v := tier(p.SonnetModel); v != "" {
		env[EnvSonnetModel] = v
	}
	if v := tier(p.OpusModel); v != "" {
		env[EnvOpusModel] = v
	}
}

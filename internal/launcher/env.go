// Package launcher builds environment variables and launches claude.
package launcher

import (
	"strconv"

	"ccenv/internal/profile"
)

// BuildEnv returns the KEY=VALUE pairs a profile injects. Empty fields produce
// no entries (so a fully-empty "plan" profile returns an empty slice).
func BuildEnv(p profile.Profile) []string {
	var env []string
	if p.BaseURL != "" {
		env = append(env, "ANTHROPIC_BASE_URL="+p.BaseURL)
	}
	if p.AuthToken != "" {
		// Custom endpoint (base_url set) uses a Bearer token; the official
		// endpoint (base_url empty) uses the console API key (x-api-key).
		if p.BaseURL != "" {
			env = append(env, "ANTHROPIC_AUTH_TOKEN="+p.AuthToken)
		} else {
			env = append(env, "ANTHROPIC_API_KEY="+p.AuthToken)
		}
	}
	env = appendModelEnv(env, p)
	if p.AutoCompactWindow > 0 {
		env = append(env, "CLAUDE_CODE_AUTO_COMPACT_WINDOW="+strconv.Itoa(p.AutoCompactWindow))
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
func appendModelEnv(env []string, p profile.Profile) []string {
	if !p.HasTierModels() {
		if p.Model != "" {
			env = append(env,
				"ANTHROPIC_MODEL="+p.Model,
				"ANTHROPIC_DEFAULT_HAIKU_MODEL="+p.Model,
				"ANTHROPIC_DEFAULT_SONNET_MODEL="+p.Model,
				"ANTHROPIC_DEFAULT_OPUS_MODEL="+p.Model,
			)
		}
		return env
	}
	tier := func(override string) string {
		if override != "" {
			return override
		}
		return p.Model
	}
	if v := tier(p.HaikuModel); v != "" {
		env = append(env, "ANTHROPIC_DEFAULT_HAIKU_MODEL="+v)
	}
	if v := tier(p.SonnetModel); v != "" {
		env = append(env, "ANTHROPIC_DEFAULT_SONNET_MODEL="+v)
	}
	if v := tier(p.OpusModel); v != "" {
		env = append(env, "ANTHROPIC_DEFAULT_OPUS_MODEL="+v)
	}
	return env
}

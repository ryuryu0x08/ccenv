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
	if p.Model != "" {
		env = append(env,
			"ANTHROPIC_MODEL="+p.Model,
			"ANTHROPIC_DEFAULT_HAIKU_MODEL="+p.Model,
			"ANTHROPIC_DEFAULT_SONNET_MODEL="+p.Model,
			"ANTHROPIC_DEFAULT_OPUS_MODEL="+p.Model,
		)
	}
	if p.AutoCompactWindow > 0 {
		env = append(env, "CLAUDE_CODE_AUTO_COMPACT_WINDOW="+strconv.Itoa(p.AutoCompactWindow))
	}
	return env
}

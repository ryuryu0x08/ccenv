package launcher

import (
	"testing"

	"ccenv/internal/profile"
)

func has(env []string, kv string) bool {
	for _, e := range env {
		if e == kv {
			return true
		}
	}
	return false
}

func TestBuildEnvFull(t *testing.T) {
	p := profile.Profile{
		BaseURL: "http://x", AuthToken: "tok", Model: "m1", AutoCompactWindow: 999,
	}
	env := BuildEnv(p)
	for _, want := range []string{
		"ANTHROPIC_BASE_URL=http://x",
		"ANTHROPIC_AUTH_TOKEN=tok",
		"ANTHROPIC_MODEL=m1",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL=m1",
		"ANTHROPIC_DEFAULT_SONNET_MODEL=m1",
		"ANTHROPIC_DEFAULT_OPUS_MODEL=m1",
		"CLAUDE_CODE_AUTO_COMPACT_WINDOW=999",
	} {
		if !has(env, want) {
			t.Errorf("missing %q in %v", want, env)
		}
	}
}

func TestBuildEnvPlanEmpty(t *testing.T) {
	env := BuildEnv(profile.Profile{})
	if len(env) != 0 {
		t.Errorf("plan profile must inject nothing, got %v", env)
	}
}

func TestBuildEnvOfficialAPIKey(t *testing.T) {
	// No base_url => official endpoint => x-api-key via ANTHROPIC_API_KEY.
	env := BuildEnv(profile.Profile{AuthToken: "k"})
	if len(env) != 1 || env[0] != "ANTHROPIC_API_KEY=k" {
		t.Errorf("official API profile should inject ANTHROPIC_API_KEY only, got %v", env)
	}
}

func TestBuildEnvCustomEndpointBearer(t *testing.T) {
	// base_url set => custom endpoint => Bearer via ANTHROPIC_AUTH_TOKEN.
	env := BuildEnv(profile.Profile{BaseURL: "http://x", AuthToken: "k"})
	foundBearer := false
	for _, e := range env {
		if e == "ANTHROPIC_AUTH_TOKEN=k" {
			foundBearer = true
		}
		if e == "ANTHROPIC_API_KEY=k" {
			t.Errorf("custom endpoint must not set ANTHROPIC_API_KEY, got %v", env)
		}
	}
	if !foundBearer {
		t.Errorf("custom endpoint should set ANTHROPIC_AUTH_TOKEN, got %v", env)
	}
}

func hasPrefix(env []string, key string) bool {
	for _, e := range env {
		if len(e) >= len(key) && e[:len(key)] == key {
			return true
		}
	}
	return false
}

func TestBuildEnvTierModelsAllSet(t *testing.T) {
	// All three tiers set => no ANTHROPIC_MODEL, each tier uses its own value.
	p := profile.Profile{
		Model:       "main",
		HaikuModel:  "h1",
		SonnetModel: "s1",
		OpusModel:   "o1",
	}
	env := BuildEnv(p)
	if hasPrefix(env, "ANTHROPIC_MODEL=") {
		t.Errorf("tier mode must NOT inject ANTHROPIC_MODEL, got %v", env)
	}
	for _, want := range []string{
		"ANTHROPIC_DEFAULT_HAIKU_MODEL=h1",
		"ANTHROPIC_DEFAULT_SONNET_MODEL=s1",
		"ANTHROPIC_DEFAULT_OPUS_MODEL=o1",
	} {
		if !has(env, want) {
			t.Errorf("missing %q in %v", want, env)
		}
	}
}

func TestBuildEnvTierModelsPartialFallback(t *testing.T) {
	// Only sonnet overridden => haiku & opus fall back to Model; still no ANTHROPIC_MODEL.
	p := profile.Profile{
		Model:       "main",
		SonnetModel: "s1",
	}
	env := BuildEnv(p)
	if hasPrefix(env, "ANTHROPIC_MODEL=") {
		t.Errorf("tier mode must NOT inject ANTHROPIC_MODEL, got %v", env)
	}
	for _, want := range []string{
		"ANTHROPIC_DEFAULT_HAIKU_MODEL=main",
		"ANTHROPIC_DEFAULT_SONNET_MODEL=s1",
		"ANTHROPIC_DEFAULT_OPUS_MODEL=main",
	} {
		if !has(env, want) {
			t.Errorf("missing %q in %v", want, env)
		}
	}
}

func TestBuildEnvTierModelsNoMainFallback(t *testing.T) {
	// Tier set but Model empty => only the explicitly-set tier is injected;
	// tiers without an override and without a Model fallback emit nothing.
	p := profile.Profile{HaikuModel: "h1"}
	env := BuildEnv(p)
	if !has(env, "ANTHROPIC_DEFAULT_HAIKU_MODEL=h1") {
		t.Errorf("expected haiku tier, got %v", env)
	}
	if hasPrefix(env, "ANTHROPIC_DEFAULT_SONNET_MODEL=") || hasPrefix(env, "ANTHROPIC_DEFAULT_OPUS_MODEL=") {
		t.Errorf("tiers without override and no Model fallback must not be injected, got %v", env)
	}
}

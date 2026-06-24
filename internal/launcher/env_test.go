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

func TestBuildEnvPartial(t *testing.T) {
	env := BuildEnv(profile.Profile{AuthToken: "k"})
	if len(env) != 1 || env[0] != "ANTHROPIC_AUTH_TOKEN=k" {
		t.Errorf("api profile should inject only token, got %v", env)
	}
}

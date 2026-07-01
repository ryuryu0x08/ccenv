package profile

import "testing"

func has(env map[string]string, k, v string) bool {
	got, ok := env[k]
	return ok && got == v
}

func TestEnvMapFull(t *testing.T) {
	p := Profile{
		BaseURL: "http://x", AuthToken: "tok", Model: "m1", AutoCompactWindow: 999,
	}
	env := p.EnvMap()
	for k, v := range map[string]string{
		EnvBaseURL:           "http://x",
		EnvAuthToken:         "tok",
		EnvModel:             "m1",
		EnvHaikuModel:        "m1",
		EnvSonnetModel:       "m1",
		EnvOpusModel:         "m1",
		EnvAutoCompactWindow: "999",
	} {
		if !has(env, k, v) {
			t.Errorf("missing %s=%q in %v", k, v, env)
		}
	}
}

func TestEnvMapPlanEmpty(t *testing.T) {
	env := Profile{}.EnvMap()
	if len(env) != 0 {
		t.Errorf("plan profile must inject nothing, got %v", env)
	}
}

func TestEnvMapOfficialAPIKey(t *testing.T) {
	// No base_url => official endpoint => x-api-key via ANTHROPIC_API_KEY.
	env := Profile{AuthToken: "k"}.EnvMap()
	if len(env) != 1 || env[EnvAPIKey] != "k" {
		t.Errorf("official API profile should inject ANTHROPIC_API_KEY only, got %v", env)
	}
}

func TestEnvMapCustomEndpointBearer(t *testing.T) {
	// base_url set => custom endpoint => Bearer via ANTHROPIC_AUTH_TOKEN.
	env := Profile{BaseURL: "http://x", AuthToken: "k"}.EnvMap()
	if !has(env, EnvAuthToken, "k") {
		t.Errorf("custom endpoint should set ANTHROPIC_AUTH_TOKEN, got %v", env)
	}
	if _, ok := env[EnvAPIKey]; ok {
		t.Errorf("custom endpoint must not set ANTHROPIC_API_KEY, got %v", env)
	}
}

func TestEnvMapTierModelsAllSet(t *testing.T) {
	// All three tiers set => no ANTHROPIC_MODEL, each tier uses its own value.
	p := Profile{
		Model:       "main",
		HaikuModel:  "h1",
		SonnetModel: "s1",
		OpusModel:   "o1",
	}
	env := p.EnvMap()
	if _, ok := env[EnvModel]; ok {
		t.Errorf("tier mode must NOT inject ANTHROPIC_MODEL, got %v", env)
	}
	for k, v := range map[string]string{
		EnvHaikuModel:  "h1",
		EnvSonnetModel: "s1",
		EnvOpusModel:   "o1",
	} {
		if !has(env, k, v) {
			t.Errorf("missing %s=%q in %v", k, v, env)
		}
	}
}

func TestEnvMapTierModelsPartialFallback(t *testing.T) {
	// Only sonnet overridden => haiku & opus fall back to Model; still no ANTHROPIC_MODEL.
	p := Profile{
		Model:       "main",
		SonnetModel: "s1",
	}
	env := p.EnvMap()
	if _, ok := env[EnvModel]; ok {
		t.Errorf("tier mode must NOT inject ANTHROPIC_MODEL, got %v", env)
	}
	for k, v := range map[string]string{
		EnvHaikuModel:  "main",
		EnvSonnetModel: "s1",
		EnvOpusModel:   "main",
	} {
		if !has(env, k, v) {
			t.Errorf("missing %s=%q in %v", k, v, env)
		}
	}
}

func TestEnvMapTierModelsNoMainFallback(t *testing.T) {
	// Tier set but Model empty => only the explicitly-set tier is injected;
	// tiers without an override and without a Model fallback emit nothing.
	env := Profile{HaikuModel: "h1"}.EnvMap()
	if !has(env, EnvHaikuModel, "h1") {
		t.Errorf("expected haiku tier, got %v", env)
	}
	if _, ok := env[EnvSonnetModel]; ok {
		t.Errorf("tiers without override and no Model fallback must not be injected, got %v", env)
	}
	if _, ok := env[EnvOpusModel]; ok {
		t.Errorf("tiers without override and no Model fallback must not be injected, got %v", env)
	}
}

func TestCompactWindow(t *testing.T) {
	cases := []struct {
		ctx  int
		pct  float64
		want int
	}{
		{131072, 80, 104858},
		{262144, 80, 209715},
		{131072, 100, 131072},
		{0, 80, 0},
		{131072, 0, 0},
		{1000, -5, 0},
	}
	for _, c := range cases {
		if got := CompactWindow(c.ctx, c.pct); got != c.want {
			t.Errorf("CompactWindow(%d,%v)=%d want %d", c.ctx, c.pct, got, c.want)
		}
	}
}

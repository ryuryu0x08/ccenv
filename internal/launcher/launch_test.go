package launcher

import (
	"encoding/json"
	"strings"
	"testing"

	"ccenv/internal/profile"
)

// TestProviderSettingsJSONCarriesEnv verifies the --settings payload wraps the
// profile's full env map under an "env" key so it can override project-scope
// settings at command-line precedence.
func TestProviderSettingsJSONCarriesEnv(t *testing.T) {
	p := profile.Profile{
		BaseURL:   "https://www.cun.ai/v1",
		AuthToken: "sk-secret-token",
		Model:     "claude-sonnet-5",
	}
	got, err := providerSettingsJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Env map[string]string `json:"env"`
	}
	if err := json.Unmarshal([]byte(got), &doc); err != nil {
		t.Fatalf("payload is not valid json: %v (%s)", err, got)
	}
	if doc.Env[profile.EnvModel] != "claude-sonnet-5" {
		t.Errorf("ANTHROPIC_MODEL not in --settings env: %v", doc.Env)
	}
	if doc.Env[profile.EnvBaseURL] != "https://www.cun.ai/v1" {
		t.Errorf("ANTHROPIC_BASE_URL not in --settings env: %v", doc.Env)
	}
	if doc.Env[profile.EnvAuthToken] != "sk-secret-token" {
		t.Errorf("auth token must be present in the real payload (only masked for display): %v", doc.Env)
	}
}

func TestMaskSecret(t *testing.T) {
	cases := map[string]string{
		"sk-1234567890abcdef": "sk-123***",
		"short":               "***",
		"":                    "***",
	}
	for in, want := range cases {
		if got := maskSecret(in); got != want {
			t.Errorf("maskSecret(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestMaskSettingsRedactsToken verifies the displayed --settings string hides
// credential values that appear in the env map.
func TestMaskSettingsRedactsToken(t *testing.T) {
	env := map[string]string{profile.EnvAuthToken: "sk-topsecretvalue"}
	raw := `{"env":{"ANTHROPIC_AUTH_TOKEN":"sk-topsecretvalue"}}`
	got := maskSettings(raw, env)
	if strings.Contains(got, "sk-topsecretvalue") {
		t.Errorf("masked settings still leaks token: %s", got)
	}
	if !strings.Contains(got, "sk-top***") {
		t.Errorf("masked settings missing masked prefix: %s", got)
	}
}

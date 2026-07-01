package claudecfg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func read(t *testing.T, p string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestYoloCreatesWhenMissing(t *testing.T) {
	p := filepath.Join(t.TempDir(), "settings.json")
	if err := EnableYolo(p); err != nil {
		t.Fatal(err)
	}
	m := read(t, p)
	if m["skipDangerousModePermissionPrompt"] != true {
		t.Errorf("flag not set: %v", m)
	}
	perms, _ := m["permissions"].(map[string]any)
	if perms["defaultMode"] != "bypassPermissions" {
		t.Errorf("defaultMode not set: %v", m)
	}
}

func TestYoloPreservesExisting(t *testing.T) {
	p := filepath.Join(t.TempDir(), "settings.json")
	os.WriteFile(p, []byte(`{"theme":"dark","permissions":{"foo":"bar"}}`), 0o600)
	if err := EnableYolo(p); err != nil {
		t.Fatal(err)
	}
	m := read(t, p)
	if m["theme"] != "dark" {
		t.Errorf("lost theme key: %v", m)
	}
	perms, _ := m["permissions"].(map[string]any)
	if perms["foo"] != "bar" {
		t.Errorf("lost permissions.foo: %v", m)
	}
	if perms["defaultMode"] != "bypassPermissions" {
		t.Errorf("defaultMode not merged: %v", m)
	}
}

func TestYoloCorruptErrors(t *testing.T) {
	p := filepath.Join(t.TempDir(), "settings.json")
	os.WriteFile(p, []byte(`{ broken`), 0o600)
	if err := EnableYolo(p); err == nil {
		t.Error("expected error on corrupt json")
	}
}

func TestMaterializeProviderSettingsAppliesEnvAndPreservesOthers(t *testing.T) {
	src := filepath.Join(t.TempDir(), "settings.json")
	dst := filepath.Join(t.TempDir(), "mirror", "settings.json")
	os.WriteFile(src, []byte(`{
		"theme": "dark",
		"env": {
			"ANTHROPIC_BASE_URL": "https://stale.example",
			"ANTHROPIC_AUTH_TOKEN": "stale-token",
			"SOME_OTHER_VAR": "keep-me"
		}
	}`), 0o600)

	env := map[string]string{"ANTHROPIC_BASE_URL": "http://127.0.0.1:11434"}
	managed := []string{"ANTHROPIC_BASE_URL", "ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_API_KEY"}
	if err := MaterializeProviderSettings(src, dst, env, managed); err != nil {
		t.Fatal(err)
	}

	m := read(t, dst)
	if m["theme"] != "dark" {
		t.Errorf("lost non-env key: %v", m)
	}
	envBlock, _ := m["env"].(map[string]any)
	if envBlock["ANTHROPIC_BASE_URL"] != "http://127.0.0.1:11434" {
		t.Errorf("base url not overridden: %v", envBlock)
	}
	if _, ok := envBlock["ANTHROPIC_AUTH_TOKEN"]; ok {
		t.Errorf("stale managed key not cleared (profile doesn't set it): %v", envBlock)
	}
	if envBlock["SOME_OTHER_VAR"] != "keep-me" {
		t.Errorf("lost unmanaged env key: %v", envBlock)
	}
}

func TestMaterializeProviderSettingsMissingSrcTreatedEmpty(t *testing.T) {
	src := filepath.Join(t.TempDir(), "nope.json")
	dst := filepath.Join(t.TempDir(), "settings.json")
	env := map[string]string{"ANTHROPIC_MODEL": "m1"}
	if err := MaterializeProviderSettings(src, dst, env, []string{"ANTHROPIC_MODEL"}); err != nil {
		t.Fatal(err)
	}
	m := read(t, dst)
	envBlock, _ := m["env"].(map[string]any)
	if envBlock["ANTHROPIC_MODEL"] != "m1" {
		t.Errorf("expected model in fresh env block: %v", m)
	}
}

func TestMaterializeProviderSettingsCorruptSrcErrors(t *testing.T) {
	src := filepath.Join(t.TempDir(), "settings.json")
	dst := filepath.Join(t.TempDir(), "mirror", "settings.json")
	os.WriteFile(src, []byte(`{ broken`), 0o600)
	if err := MaterializeProviderSettings(src, dst, nil, nil); err == nil {
		t.Error("expected error on corrupt src json")
	}
	if _, err := os.Stat(dst); err == nil {
		t.Error("dst should not be written when src is corrupt")
	}
}

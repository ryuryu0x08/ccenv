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

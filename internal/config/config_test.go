package config

import (
	"path/filepath"
	"testing"

	"ccenv/internal/profile"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.toml")
	c := &Config{Profiles: map[string]profile.Profile{
		"local": {BaseURL: "http://127.0.0.1:11434", AuthToken: "local", Model: "m1", AutoCompactWindow: 1000},
		"plan":  {},
	}}
	if err := Save(p, c); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := Load(p)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Profiles["local"].BaseURL != "http://127.0.0.1:11434" {
		t.Errorf("base_url lost: %+v", got.Profiles["local"])
	}
	if _, ok := got.Profiles["plan"]; !ok {
		t.Errorf("empty profile lost")
	}
}

func TestLoadMissingReturnsEmpty(t *testing.T) {
	c, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatalf("missing file should be empty config, got err: %v", err)
	}
	if len(c.Profiles) != 0 {
		t.Errorf("expected empty, got %d", len(c.Profiles))
	}
}

func TestLoadCorruptErrors(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.toml")
	if err := writeFile(p, "this is = = not toml ["); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil {
		t.Errorf("expected error on corrupt toml")
	}
}

func TestCRUD(t *testing.T) {
	c := &Config{Profiles: map[string]profile.Profile{}}
	c.Set("a", profile.Profile{Model: "x"})
	if _, ok := c.Get("a"); !ok {
		t.Fatal("get after set failed")
	}
	if !c.Has("a") {
		t.Fatal("has failed")
	}
	if err := c.Remove("a"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if c.Has("a") {
		t.Fatal("still present after remove")
	}
	if err := c.Remove("ghost"); err == nil {
		t.Fatal("remove missing should error")
	}
}

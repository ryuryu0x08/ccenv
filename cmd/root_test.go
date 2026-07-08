package cmd

import (
	"reflect"
	"testing"
)

// TestStripGlobalFlags verifies that ccenv consumes only its own leading flags
// (--verbose/-v) before the profile name and leaves everything else — including
// the profile name, subcommand flags, and claude pass-through args — untouched.
func TestStripGlobalFlags(t *testing.T) {
	cases := []struct {
		name        string
		args        []string
		wantVerbose bool
		wantRest    []string
	}{
		{"no flags", []string{"cun"}, false, []string{"cun"}},
		{"verbose long", []string{"--verbose", "cun"}, true, []string{"cun"}},
		{"verbose short", []string{"-v", "cun"}, true, []string{"cun"}},
		{"verbose then passthrough", []string{"-v", "cun", "--dangerously-skip-permissions"}, true, []string{"cun", "--dangerously-skip-permissions"}},
		{"flag after profile is not consumed", []string{"cun", "-v"}, false, []string{"cun", "-v"}},
		{"stops at first non-global flag", []string{"-v", "--other", "cun"}, true, []string{"--other", "cun"}},
		{"empty", nil, false, []string{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotVerbose, gotRest := StripGlobalFlags(c.args)
			if gotVerbose != c.wantVerbose {
				t.Errorf("verbose = %v, want %v", gotVerbose, c.wantVerbose)
			}
			if len(gotRest) == 0 && len(c.wantRest) == 0 {
				return
			}
			if !reflect.DeepEqual(gotRest, c.wantRest) {
				t.Errorf("rest = %v, want %v", gotRest, c.wantRest)
			}
		})
	}
}

package cmd

import "testing"

// labelForModel must return "" whenever the current model has no matching option
// label. promptModel relies on this to decide whether to set survey.Select.Default
// at all — survey rejects a Default ("" included) that isn't one of Options.
func TestLabelForModel(t *testing.T) {
	opts := []string{"gpt-4 (ctx=128000)", "glm-5.2"}
	cases := []struct {
		name string
		cur  string
		want string
	}{
		{"empty current", "", ""},
		{"absent current", "missing-model", ""},
		{"match with ctx suffix", "gpt-4", "gpt-4 (ctx=128000)"},
		{"match plain", "glm-5.2", "glm-5.2"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := labelForModel(opts, c.cur); got != c.want {
				t.Fatalf("labelForModel(%q) = %q, want %q", c.cur, got, c.want)
			}
		})
	}
}

func TestModelIDFromLabel(t *testing.T) {
	cases := map[string]string{
		"glm-5.2":            "glm-5.2",
		"gpt-4 (ctx=128000)": "gpt-4",
		"no-suffix-model":    "no-suffix-model",
	}
	for in, want := range cases {
		if got := modelIDFromLabel(in); got != want {
			t.Errorf("modelIDFromLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

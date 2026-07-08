package launcher

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ccenv/internal/profile"
)

// secretEnvKeys are provider env keys whose values are credentials and must be
// masked in any diagnostic output.
var secretEnvKeys = map[string]bool{
	profile.EnvAuthToken: true,
	profile.EnvAPIKey:    true,
}

// maskSecret redacts a credential value, keeping only a short prefix so the
// user can still tell which token is in play without leaking it.
func maskSecret(v string) string {
	const keep = 6
	if len(v) <= keep {
		return "***"
	}
	return v[:keep] + "***"
}

// verboseLog writes ccenv's launch diagnostics to w. It prints the resolved
// profile env (secrets masked), the mirror dir / CLAUDE_CONFIG_DIR, the
// injected --settings payload, and claude's full argv. It also warns when a
// project-scope .claude/settings.json in the current working directory could
// shadow the profile's settings — the exact failure mode where the effective
// model silently changes with the launch directory.
func verboseLog(w io.Writer, name string, p profile.Profile, mirrorDir, settingsJSON string, argv []string) {
	fmt.Fprintf(w, "ccenv: launching profile %q\n", name)

	env := p.EnvMap()
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Fprintln(w, "ccenv: injected env:")
	for _, k := range keys {
		v := env[k]
		if secretEnvKeys[k] {
			v = maskSecret(v)
		}
		fmt.Fprintf(w, "ccenv:   %s=%s\n", k, v)
	}

	if mirrorDir != "" {
		fmt.Fprintf(w, "ccenv: CLAUDE_CONFIG_DIR=%s\n", mirrorDir)
	}
	if settingsJSON != "" {
		fmt.Fprintf(w, "ccenv: --settings=%s\n", maskSettings(settingsJSON, env))
	}
	fmt.Fprintf(w, "ccenv: exec: claude %s\n", strings.Join(argv, " "))

	warnShadowingProjectSettings(w)
}

// maskSettings redacts secret values inside the --settings JSON string for
// display, matching the same keys verboseLog masks in the env dump.
func maskSettings(settingsJSON string, env map[string]string) string {
	out := settingsJSON
	for k := range secretEnvKeys {
		if v, ok := env[k]; ok && v != "" {
			out = strings.ReplaceAll(out, v, maskSecret(v))
		}
	}
	return out
}

// warnShadowingProjectSettings checks whether the current working directory
// carries a project-scope .claude/settings.json. Such a file sits ABOVE the
// CLAUDE_CONFIG_DIR mirror (user scope) in Claude Code's precedence, so its
// env/model keys would override the profile — making the effective config
// depend on where ccenv was launched from. --settings injection defeats env
// overrides, but non-env keys (model, effortLevel, ...) still leak, so we
// surface the file for the user.
func warnShadowingProjectSettings(w io.Writer) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(w, "ccenv: (verbose) cannot resolve cwd to check project settings: %v\n", err)
		return
	}
	projectSettings := filepath.Join(cwd, ".claude", "settings.json")
	if _, err := os.Stat(projectSettings); err != nil {
		return
	}
	fmt.Fprintf(w, "ccenv: WARNING: %s exists and has higher precedence than the profile's\n", projectSettings)
	fmt.Fprintln(w, "ccenv:          user-scope settings. Its model/effortLevel/permissions keys can override")
	fmt.Fprintln(w, "ccenv:          the profile. Provider env is protected via --settings, but other keys are not.")
}

package launcher

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"ccenv/internal/claudehome"
	"ccenv/internal/profile"
)

// settingsFlag is claude's command-line flag for loading additional settings
// from a file path or inline JSON string.
const settingsFlag = "--settings"

// Launch finds claude in PATH, passes through args, and inherits stdio. It
// returns claude's exit code (0 on success) and an error. If claude ran but
// exited non-zero, the returned code reflects that and err is nil (the
// non-zero status is not a launch failure). If claude could not be started,
// code is 1 and err is set.
//
// A profile with any provider env (base_url/auth_token/model/...) is injected
// two complementary ways:
//
//   - CLAUDE_CONFIG_DIR points at a per-profile mirror of ~/.claude (see
//     claudehome) whose settings.json carries the profile's env. This is
//     user-scope config; claude's background daemon and Workflow subagents
//     resolve provider config from it (not from the launching process's env),
//     so it reaches everything claude spawns, foreground and background alike.
//   - `--settings <json>` carries the same env at command-line precedence,
//     which sits ABOVE local/project/user settings files (only managed
//     settings outrank it). Without this, a project-scope .claude/settings.json
//     in the launch directory would override the mirror's user-scope env,
//     making the effective provider/model depend on the current working
//     directory. --settings makes the profile win regardless of CWD.
//
// A profile with nothing to inject (bare official-account "plan" profile)
// skips both and runs against the real ~/.claude unmodified.
func Launch(name string, p profile.Profile, args []string, verbose bool) (int, error) {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return 1, fmt.Errorf("claude not found in PATH; install Claude Code or add it to PATH: %w", err)
	}

	env := os.Environ()
	argv := args
	var mirrorDir, settingsJSON string
	if len(p.EnvMap()) > 0 {
		mirrorDir, err = claudehome.Prepare(name, p)
		if err != nil {
			return 1, fmt.Errorf("prepare claude home for profile %q: %w", name, err)
		}
		env = append(env, "CLAUDE_CONFIG_DIR="+mirrorDir)

		settingsJSON, err = providerSettingsJSON(p)
		if err != nil {
			return 1, fmt.Errorf("build --settings for profile %q: %w", name, err)
		}
		// Prepend so a user-supplied --settings in args, if any, still parses;
		// claude merges multiple settings sources by precedence.
		argv = append([]string{settingsFlag, settingsJSON}, args...)
	}

	if verbose {
		verboseLog(os.Stderr, name, p, mirrorDir, settingsJSON, argv)
	}

	cmd := exec.Command(bin, argv...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode(), nil
		}
		return 1, fmt.Errorf("run claude: %w", err)
	}
	return 0, nil
}

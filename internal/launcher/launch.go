package launcher

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"ccenv/internal/claudehome"
	"ccenv/internal/profile"
)

// Launch finds claude in PATH, passes through args, and inherits stdio. It
// returns claude's exit code (0 on success) and an error. If claude ran but
// exited non-zero, the returned code reflects that and err is nil (the
// non-zero status is not a launch failure). If claude could not be started,
// code is 1 and err is set.
//
// A profile with any provider env (base_url/auth_token/model/...) is not
// injected as process environment variables directly: claude's background
// daemon and Workflow subagents resolve provider config from
// CLAUDE_CONFIG_DIR's settings.json rather than from the launching process's
// environment, so injecting ANTHROPIC_* into the process env would only
// reach the foreground session. Instead, Launch points CLAUDE_CONFIG_DIR at
// a per-profile mirror of ~/.claude (see claudehome) whose settings.json
// carries the profile's env; claude then propagates that env to everything
// it spawns, foreground and background alike. A profile with nothing to
// inject (bare official-account "plan" profile) skips this and runs against
// the real ~/.claude unmodified.
func Launch(name string, p profile.Profile, args []string) (int, error) {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return 1, fmt.Errorf("claude not found in PATH; install Claude Code or add it to PATH: %w", err)
	}
	env := os.Environ()
	if len(p.EnvMap()) > 0 {
		mirrorDir, err := claudehome.Prepare(name, p)
		if err != nil {
			return 1, fmt.Errorf("prepare claude home for profile %q: %w", name, err)
		}
		env = append(env, "CLAUDE_CONFIG_DIR="+mirrorDir)
	}
	cmd := exec.Command(bin, args...)
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

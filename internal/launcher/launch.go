package launcher

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"ccenv/internal/profile"
)

// Launch finds claude in PATH, injects the profile's env on top of the current
// environment, passes through args, and inherits stdio. It returns claude's
// exit code (0 on success) and an error. If claude ran but exited non-zero,
// the returned code reflects that and err is nil (the non-zero status is not a
// launch failure). If claude could not be started, code is 1 and err is set.
func Launch(p profile.Profile, args []string) (int, error) {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return 1, fmt.Errorf("claude not found in PATH; install Claude Code or add it to PATH: %w", err)
	}
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), BuildEnv(p)...)
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

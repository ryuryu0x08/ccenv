package launcher

import (
	"fmt"
	"os"
	"os/exec"

	"ccenv/internal/profile"
)

// Launch finds claude in PATH, injects the profile's env on top of the current
// environment, passes through args, and inherits stdio. Returns claude's exit
// error (if any).
func Launch(p profile.Profile, args []string) error {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude not found in PATH; install Claude Code or add it to PATH: %w", err)
	}
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), BuildEnv(p)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

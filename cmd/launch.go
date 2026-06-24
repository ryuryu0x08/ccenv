package cmd

import (
	"fmt"
	"strings"

	"ccenv/internal/launcher"
)

// RunLaunch starts claude using the named profile, passing through extraArgs.
// Returns claude's exit code and any launch error.
func RunLaunch(name string, extraArgs []string) (int, error) {
	c, _, err := loadConfig()
	if err != nil {
		return 1, err
	}
	p, ok := c.Get(name)
	if !ok {
		return 1, fmt.Errorf("profile %q does not exist (available: %s)", name, strings.Join(c.Names(), ", "))
	}
	return launcher.Launch(p, extraArgs)
}

package cmd

import (
	"fmt"

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
		return 1, profileNotFoundErr(name, c)
	}
	code, err := launcher.Launch(name, p, extraArgs)
	if err != nil {
		return code, fmt.Errorf("launch profile %q: %w", name, err)
	}
	return code, nil
}

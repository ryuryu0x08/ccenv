package cmd

import (
	"fmt"

	"ccenv/internal/launcher"
)

// RunLaunch starts claude using the named profile, passing through extraArgs.
// When verbose is true, ccenv prints launch diagnostics (resolved env, mirror
// dir, injected --settings, claude argv, and CWD project-settings warnings) to
// stderr before exec. Returns claude's exit code and any launch error.
func RunLaunch(name string, extraArgs []string, verbose bool) (int, error) {
	c, _, err := loadConfig()
	if err != nil {
		return 1, err
	}
	p, ok := c.Get(name)
	if !ok {
		return 1, profileNotFoundErr(name, c)
	}
	code, err := launcher.Launch(name, p, extraArgs, verbose)
	if err != nil {
		return code, fmt.Errorf("launch profile %q: %w", name, err)
	}
	return code, nil
}

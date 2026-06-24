package cmd

import (
	"fmt"
	"strings"

	"ccenv/internal/launcher"
)

// RunLaunch starts claude using the named profile, passing through extraArgs.
func RunLaunch(name string, extraArgs []string) error {
	c, _, err := loadConfig()
	if err != nil {
		return err
	}
	p, ok := c.Get(name)
	if !ok {
		return fmt.Errorf("profile %q does not exist (available: %s)", name, strings.Join(c.Names(), ", "))
	}
	return launcher.Launch(p, extraArgs)
}

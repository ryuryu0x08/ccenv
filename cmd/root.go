// Package cmd wires the ccenv CLI.
package cmd

import (
	"fmt"
	"strings"

	"ccenv/internal/config"
	"github.com/spf13/cobra"
)

// NewRoot builds the cobra root command with all subcommands attached.
func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "ccenv",
		Short:         "Launch Claude Code with named profiles",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Help(); err != nil {
				return fmt.Errorf("print help: %w", err)
			}
			fmt.Println()
			return runLs()
		},
	}
	root.AddCommand(newLsCmd(), newRmCmd(), newAddCmd(), newEditCmd(), newYoloCmd())
	return root
}

// verboseFlags are the leading ccenv-global flags that enable launch
// diagnostics. They are consumed by StripGlobalFlags before the profile name
// so they are never forwarded to claude.
var verboseFlags = map[string]bool{"--verbose": true, "-v": true}

// StripGlobalFlags consumes ccenv's own leading flags from args (currently just
// --verbose/-v) and returns whether verbose was set plus the remaining args.
// Only flags appearing before the first non-flag token (the profile name or
// subcommand) are consumed; the first unrecognized token stops scanning so
// subcommand flags and claude pass-through args are left untouched.
func StripGlobalFlags(args []string) (verbose bool, rest []string) {
	i := 0
	for i < len(args) && verboseFlags[args[i]] {
		verbose = true
		i++
	}
	return verbose, args[i:]
}

// IsReservedIn reports whether name matches a registered subcommand (or its
// aliases) of root, or the built-in help/completion commands. This is the
// single source of truth for routing: anything reserved goes through cobra,
// everything else is treated as a profile name.
func IsReservedIn(root *cobra.Command, name string) bool {
	for _, c := range root.Commands() {
		if c.Name() == name {
			return true
		}
		for _, a := range c.Aliases {
			if a == name {
				return true
			}
		}
	}
	return false
}

// loadConfig loads from the default path, surfacing a friendly error.
func loadConfig() (*config.Config, string, error) {
	path, err := config.DefaultPath()
	if err != nil {
		return nil, "", err
	}
	c, err := config.Load(path)
	if err != nil {
		return nil, "", err
	}
	return c, path, nil
}

// profileNotFoundErr formats a consistent not-found error listing available names.
func profileNotFoundErr(name string, c *config.Config) error {
	return fmt.Errorf("profile %q does not exist (available: %s)", name, strings.Join(c.Names(), ", "))
}

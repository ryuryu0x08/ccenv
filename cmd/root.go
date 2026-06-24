// Package cmd wires the ccenv CLI.
package cmd

import (
	"fmt"

	"ccenv/internal/config"
	"github.com/spf13/cobra"
)

// reserved are subcommand names that take priority over profile names.
var reserved = map[string]bool{
	"add": true, "edit": true, "ls": true, "rm": true, "yolo": true,
	"help": true, "completion": true,
}

// IsReserved reports whether name is a reserved subcommand word.
func IsReserved(name string) bool { return reserved[name] }

// NewRoot builds the cobra root command with all subcommands attached.
func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "ccenv",
		Short:         "Launch Claude Code with named profiles",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()
			fmt.Println()
			return runLs()
		},
	}
	root.AddCommand(newLsCmd(), newRmCmd(), newAddCmd(), newEditCmd(), newYoloCmd())
	return root
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

func saveConfig(path string, c *config.Config) error {
	return config.Save(path, c)
}

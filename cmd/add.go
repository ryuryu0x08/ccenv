package cmd

import (
	"fmt"

	"ccenv/internal/config"
	"ccenv/internal/profile"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name>",
		Short: "Create a profile interactively",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if IsReservedIn(NewRoot(), name) {
				return fmt.Errorf("%q is a reserved word; choose another profile name", name)
			}
			c, path, err := loadConfig()
			if err != nil {
				return err
			}
			if c.Has(name) {
				return fmt.Errorf("profile %q already exists; use: ccenv edit %s", name, name)
			}
			p, err := promptProfile(profile.Profile{})
			if err != nil {
				return fmt.Errorf("prompt profile %q: %w", name, err)
			}
			c.Set(name, p)
			if err := config.Save(path, c); err != nil {
				return fmt.Errorf("save profile %q: %w", name, err)
			}
			fmt.Printf("Created profile %q\n", name)
			return nil
		},
	}
}

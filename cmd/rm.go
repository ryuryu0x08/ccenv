package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Delete a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			c, path, err := loadConfig()
			if err != nil {
				return err
			}
			if err := c.Remove(name); err != nil {
				return fmt.Errorf("%w (available: %s)", err, strings.Join(c.Names(), ", "))
			}
			if err := saveConfig(path, c); err != nil {
				return err
			}
			fmt.Printf("Removed profile %q\n", name)
			return nil
		},
	}
}

package cmd

import (
	"fmt"

	"ccenv/internal/claudecfg"
	"github.com/spf13/cobra"
)

func newYoloCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "yolo",
		Short: "Enable bypass + dangerous mode in ~/.claude/settings.json",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := claudecfg.DefaultPath()
			if err != nil {
				return err
			}
			if err := claudecfg.EnableYolo(path); err != nil {
				return err
			}
			fmt.Printf("YOLO mode enabled in %s\n", path)
			return nil
		},
	}
}

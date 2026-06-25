package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List all profiles",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, args []string) error { return runLs() },
	}
}

func mask(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 4 {
		return "***"
	}
	return token[:4] + "***"
}

func runLs() error {
	c, _, err := loadConfig()
	if err != nil {
		return err
	}
	names := c.Names()
	if len(names) == 0 {
		fmt.Println("No profiles. Create one with: ccenv add <name>")
		return nil
	}
	fmt.Println("Profiles:")
	for _, n := range names {
		p := c.Profiles[n]
		base := p.BaseURL
		if base == "" {
			base = "(official default)"
		}
		fmt.Printf("  %-12s base=%s\n", n, base)
		fmt.Printf("               token=%s model=%s models_url=%v compact=%d\n",
			mask(p.AuthToken), p.Model, p.ModelsURL != "", p.AutoCompactWindow)
		if p.HasTierModels() {
			fmt.Printf("               tiers: haiku=%s sonnet=%s opus=%s\n",
				tierOrFallback(p.HaikuModel), tierOrFallback(p.SonnetModel), tierOrFallback(p.OpusModel))
		}
	}
	return nil
}

// tierOrFallback shows a tier model, or "(main)" when it falls back to Model.
func tierOrFallback(m string) string {
	if m == "" {
		return "(main)"
	}
	return m
}

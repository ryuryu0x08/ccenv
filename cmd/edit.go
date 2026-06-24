package cmd

import (
	"fmt"
	"strings"

	"ccenv/internal/profile"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			c, path, err := loadConfig()
			if err != nil {
				return err
			}
			cur, ok := c.Get(name)
			if !ok {
				return fmt.Errorf("profile %q does not exist (available: %s)", name, strings.Join(c.Names(), ", "))
			}

			fields := []string{"base_url", "auth_token", "models_url + model", "compact_window"}
			var pick []string
			if err := survey.AskOne(&survey.MultiSelect{
				Message: "选择要修改的字段 (空格选中，回车提交):",
				Options: fields,
			}, &pick); err != nil {
				return err
			}

			updated, err := editFields(cur, pick)
			if err != nil {
				return err
			}
			c.Set(name, updated)
			if err := saveConfig(path, c); err != nil {
				return err
			}
			fmt.Printf("Updated profile %q\n", name)
			return nil
		},
	}
}

// editFields prompts only for the selected fields. "models_url + model" reuses
// the full models flow via promptModelSection.
func editFields(cur profile.Profile, pick []string) (profile.Profile, error) {
	p := cur
	sel := map[string]bool{}
	for _, f := range pick {
		sel[f] = true
	}
	if sel["base_url"] {
		if err := survey.AskOne(&survey.Input{Message: "Base URL:", Default: p.BaseURL}, &p.BaseURL); err != nil {
			return p, err
		}
	}
	if sel["auth_token"] {
		if err := survey.AskOne(&survey.Input{Message: "Auth token:", Default: p.AuthToken}, &p.AuthToken); err != nil {
			return p, err
		}
	}
	if sel["models_url + model"] {
		np, err := promptModelSection(p)
		if err != nil {
			return p, err
		}
		p = np
	}
	if sel["compact_window"] {
		var s string
		if err := survey.AskOne(&survey.Input{
			Message: "Auto compact window (绝对 token 数，留空/0=不注入):",
			Default: fmt.Sprintf("%d", p.AutoCompactWindow),
		}, &s); err != nil {
			return p, err
		}
		n := 0
		fmt.Sscanf(s, "%d", &n)
		p.AutoCompactWindow = n
	}
	return p, nil
}

package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"ccenv/internal/config"
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
				return profileNotFoundErr(name, c)
			}

			fields := []string{"base_url", "auth_token", "models_url", "model", "compact_window"}
			var pick []string
			if err := survey.AskOne(&survey.MultiSelect{
				Message: "选择要修改的字段 (空格选中，回车提交):",
				Options: fields,
			}, &pick); err != nil {
				return fmt.Errorf("prompt fields to edit: %w", err)
			}

			updated, err := editFields(cur, pick)
			if err != nil {
				return err
			}
			c.Set(name, updated)
			if err := config.Save(path, c); err != nil {
				return fmt.Errorf("save profile %q: %w", name, err)
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
			return p, fmt.Errorf("prompt base url: %w", err)
		}
	}
	if sel["auth_token"] {
		if err := survey.AskOne(&survey.Input{Message: "Auth token:", Default: p.AuthToken}, &p.AuthToken); err != nil {
			return p, fmt.Errorf("prompt auth token: %w", err)
		}
	}
	if sel["models_url"] {
		np, err := promptModelsURL(p)
		if err != nil {
			return p, err
		}
		p = np
	}
	if sel["model"] {
		np, err := promptModel(p)
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
			return p, fmt.Errorf("prompt compact window: %w", err)
		}
		s = strings.TrimSpace(s)
		if s == "" {
			p.AutoCompactWindow = 0
		} else {
			n, err := strconv.Atoi(s)
			if err != nil {
				return p, fmt.Errorf("invalid compact window %q: %w", s, err)
			}
			p.AutoCompactWindow = n
		}
	}
	return p, nil
}

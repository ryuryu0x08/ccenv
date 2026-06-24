package cmd

import (
	"fmt"
	"strconv"

	"ccenv/internal/models"
	"ccenv/internal/profile"
	"github.com/AlecAivazis/survey/v2"
)

// promptProfile runs the full add interactive flow starting from `cur` (zero
// value for add). Returns the updated profile.
func promptProfile(cur profile.Profile) (profile.Profile, error) {
	p := cur
	if err := survey.AskOne(&survey.Input{
		Message: "Base URL (留空=官方默认 api.anthropic.com):",
		Default: p.BaseURL,
	}, &p.BaseURL); err != nil {
		return p, fmt.Errorf("prompt base url: %w", err)
	}
	if err := survey.AskOne(&survey.Input{
		Message: "Auth token / API key (留空=不注入，用于官方 plan):",
		Default: p.AuthToken,
	}, &p.AuthToken); err != nil {
		return p, fmt.Errorf("prompt auth token: %w", err)
	}
	return promptModelSection(p)
}

// promptManualModel asks for a model name by hand (used when there's no models
// API or the fetch failed) and disables the auto-compact window.
func promptManualModel(p profile.Profile) (profile.Profile, error) {
	p.AutoCompactWindow = 0
	if err := survey.AskOne(&survey.Input{Message: "模型名 (留空=不注入模型):", Default: p.Model}, &p.Model); err != nil {
		return p, fmt.Errorf("prompt model name: %w", err)
	}
	return p, nil
}

// promptModelSection asks whether there's a /v1/models API, then either fetches
// and selects a model (computing the compact window) or takes a manual model
// name. Shared by add (full flow) and edit (model-only).
func promptModelSection(p profile.Profile) (profile.Profile, error) {
	hasAPI := p.ModelsURL != ""
	if err := survey.AskOne(&survey.Confirm{
		Message: "该 endpoint 是否支持 OpenAI 兼容的 /v1/models 接口? (仅支持 OpenAI 格式，官方 Anthropic 不支持)",
		Default: hasAPI,
	}, &hasAPI); err != nil {
		return p, fmt.Errorf("prompt models api support: %w", err)
	}
	if !hasAPI {
		p.ModelsURL = ""
		return promptManualModel(p)
	}
	if err := survey.AskOne(&survey.Input{
		Message: "Models URL (完整 /v1/models 地址):",
		Default: p.ModelsURL,
	}, &p.ModelsURL); err != nil {
		return p, fmt.Errorf("prompt models url: %w", err)
	}
	list, err := models.Fetch(p.ModelsURL, p.AuthToken)
	if err != nil {
		fmt.Printf("拉取模型失败 (%v)，降级为手动填写模型名。\n", err)
		return promptManualModel(p)
	}
	opts := make([]string, len(list))
	ctxByLabel := map[string]int{}
	for i, m := range list {
		label := m.ID
		if m.ContextLength > 0 {
			label = fmt.Sprintf("%s (ctx=%d)", m.ID, m.ContextLength)
		}
		opts[i] = label
		ctxByLabel[label] = m.ContextLength
	}
	var chosen string
	if err := survey.AskOne(&survey.Select{Message: "选择默认模型:", Options: opts}, &chosen); err != nil {
		return p, fmt.Errorf("prompt model selection: %w", err)
	}
	p.Model = list[indexOf(opts, chosen)].ID
	if ctx := ctxByLabel[chosen]; ctx > 0 {
		pctStr := fmt.Sprintf("%g", profile.DefaultCompactRatioPercent)
		if err := survey.AskOne(&survey.Input{Message: "压缩窗口比例 (%，默认 80):", Default: pctStr}, &pctStr); err != nil {
			return p, fmt.Errorf("prompt compact ratio: %w", err)
		}
		pct, perr := strconv.ParseFloat(pctStr, 64)
		if perr != nil || pct <= 0 {
			pct = profile.DefaultCompactRatioPercent
		}
		p.AutoCompactWindow = profile.CompactWindow(ctx, pct)
	} else {
		p.AutoCompactWindow = 0
	}
	return p, nil
}

func indexOf(s []string, v string) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return 0
}

package cmd

import (
	"fmt"
	"math"
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
		return p, err
	}
	if err := survey.AskOne(&survey.Input{
		Message: "Auth token / API key (留空=不注入，用于官方 plan):",
		Default: p.AuthToken,
	}, &p.AuthToken); err != nil {
		return p, err
	}
	return promptModelSection(p)
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
		return p, err
	}
	if !hasAPI {
		p.ModelsURL = ""
		p.AutoCompactWindow = 0
		err := survey.AskOne(&survey.Input{Message: "模型名 (留空=不注入模型):", Default: p.Model}, &p.Model)
		return p, err
	}
	if err := survey.AskOne(&survey.Input{
		Message: "Models URL (完整 /v1/models 地址):",
		Default: p.ModelsURL,
	}, &p.ModelsURL); err != nil {
		return p, err
	}
	list, err := models.Fetch(p.ModelsURL, p.AuthToken)
	if err != nil {
		fmt.Printf("拉取模型失败 (%v)，降级为手动填写模型名。\n", err)
		p.AutoCompactWindow = 0
		err := survey.AskOne(&survey.Input{Message: "模型名 (留空=不注入模型):", Default: p.Model}, &p.Model)
		return p, err
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
		return p, err
	}
	p.Model = list[indexOf(opts, chosen)].ID
	if ctx := ctxByLabel[chosen]; ctx > 0 {
		pctStr := "80"
		if err := survey.AskOne(&survey.Input{Message: "压缩窗口比例 (%，默认 80):", Default: "80"}, &pctStr); err != nil {
			return p, err
		}
		pct, perr := strconv.ParseFloat(pctStr, 64)
		if perr != nil || pct <= 0 {
			pct = 80
		}
		p.AutoCompactWindow = int(math.Round(float64(ctx) * pct / 100.0))
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

package cmd

import (
	"fmt"
	"strconv"
	"strings"

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
	return promptTierModels(p, nil)
}

const tierFallbackLabel = "[留空回退到主 model]"

// promptTierModels optionally configures per-tier (Haiku/Sonnet/Opus) models.
// It first asks whether the user wants per-tier models at all; if not, all tier
// fields are cleared (single-model behavior). If yes, each tier is chosen from
// `opts` (an OpenAI model list, when available) prefixed with a fallback option,
// or entered by hand when `opts` is nil/empty. An empty/fallback choice leaves
// that tier falling back to the main Model.
func promptTierModels(p profile.Profile, opts []string) (profile.Profile, error) {
	want := p.HasTierModels()
	if err := survey.AskOne(&survey.Confirm{
		Message: "是否为不同层级 (Haiku/Sonnet/Opus) 配置不同模型? (默认否)",
		Default: want,
	}, &want); err != nil {
		return p, fmt.Errorf("prompt tier models toggle: %w", err)
	}
	if !want {
		p.HaikuModel, p.SonnetModel, p.OpusModel = "", "", ""
		return p, nil
	}
	var err error
	if p.HaikuModel, err = promptOneTier("Haiku", p.HaikuModel, p.Model, opts); err != nil {
		return p, err
	}
	if p.SonnetModel, err = promptOneTier("Sonnet", p.SonnetModel, p.Model, opts); err != nil {
		return p, err
	}
	if p.OpusModel, err = promptOneTier("Opus", p.OpusModel, p.Model, opts); err != nil {
		return p, err
	}
	return p, nil
}

// promptOneTier prompts for a single tier's model. With a non-empty opts list it
// uses a Select whose first option is the fallback sentinel; otherwise it uses a
// free-text Input. Returns "" when the user picks the fallback or leaves it
// blank (meaning: use the main model for this tier).
func promptOneTier(tier, cur, mainModel string, opts []string) (string, error) {
	msg := fmt.Sprintf("%s 层级模型 (留空回退到主 model: %s)", tier, mainModel)
	if len(opts) == 0 {
		var v string
		if err := survey.AskOne(&survey.Input{Message: msg, Default: cur}, &v); err != nil {
			return "", fmt.Errorf("prompt %s tier model: %w", tier, err)
		}
		return strings.TrimSpace(v), nil
	}
	choices := append([]string{tierFallbackLabel}, opts...)
	def := tierFallbackLabel
	if cur != "" {
		for _, o := range opts {
			if modelIDFromLabel(o) == cur {
				def = o
				break
			}
		}
	}
	var chosen string
	if err := survey.AskOne(&survey.Select{Message: msg, Options: choices, Default: def}, &chosen); err != nil {
		return "", fmt.Errorf("prompt %s tier model: %w", tier, err)
	}
	if chosen == tierFallbackLabel {
		return "", nil
	}
	return modelIDFromLabel(chosen), nil
}

// modelIDFromLabel recovers the raw model id from a list label that may carry a
// " (ctx=N)" suffix produced during model listing.
func modelIDFromLabel(label string) string {
	if i := strings.Index(label, " (ctx="); i >= 0 {
		return label[:i]
	}
	return label
}

// manualEntryLabel is the trailing Select option that drops to free-text model
// entry when a models list is available but the desired model isn't listed.
const manualEntryLabel = "[手动输入模型名]"

// promptModelSection is the full add-flow model step: configure the models URL
// then choose the model.
func promptModelSection(p profile.Profile) (profile.Profile, error) {
	p, err := promptModelsURL(p)
	if err != nil {
		return p, err
	}
	return promptModel(p)
}

// promptModelsURL toggles and edits the OpenAI-compatible /v1/models endpoint.
// Declining clears ModelsURL; it does not touch Model.
func promptModelsURL(p profile.Profile) (profile.Profile, error) {
	hasAPI := p.ModelsURL != ""
	if err := survey.AskOne(&survey.Confirm{
		Message: "该 endpoint 是否支持 OpenAI 兼容的 /v1/models 接口? (仅支持 OpenAI 格式，官方 Anthropic 不支持)",
		Default: hasAPI,
	}, &hasAPI); err != nil {
		return p, fmt.Errorf("prompt models api support: %w", err)
	}
	if !hasAPI {
		p.ModelsURL = ""
		return p, nil
	}
	if err := survey.AskOne(&survey.Input{
		Message: "Models URL (完整 /v1/models 地址):",
		Default: p.ModelsURL,
	}, &p.ModelsURL); err != nil {
		return p, fmt.Errorf("prompt models url: %w", err)
	}
	return p, nil
}

// promptModel selects the main model (and per-tier overrides). With a reachable
// ModelsURL it offers the fetched list plus a trailing manual-entry option;
// otherwise (no URL / fetch failed / manual chosen) it falls back to free text.
func promptModel(p profile.Profile) (profile.Profile, error) {
	if p.ModelsURL == "" {
		return promptManualModel(p)
	}
	list, err := models.Fetch(p.ModelsURL, p.AuthToken)
	if err != nil {
		fmt.Printf("拉取模型失败 (%v)，降级为手动填写模型名。\n", err)
		return promptManualModel(p)
	}
	opts, ctxByLabel := modelOptions(list)
	choices := append(append([]string{}, opts...), manualEntryLabel)
	var chosen string
	if err := survey.AskOne(&survey.Select{
		Message: "选择默认模型 (最后一项可手动输入):",
		Options: choices,
		Default: labelForModel(opts, p.Model),
	}, &chosen); err != nil {
		return p, fmt.Errorf("prompt model selection: %w", err)
	}
	if chosen == manualEntryLabel {
		return promptManualModel(p)
	}
	p.Model = modelIDFromLabel(chosen)
	np, err := promptCompactFromCtx(p, ctxByLabel[chosen])
	if err != nil {
		return p, err
	}
	return promptTierModels(np, opts)
}

// modelOptions builds Select labels (id with optional " (ctx=N)" suffix) and a
// label→context-length lookup from a fetched model list.
func modelOptions(list []models.Model) ([]string, map[string]int) {
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
	return opts, ctxByLabel
}

// labelForModel finds the option label matching model id `cur`, or "" if absent.
func labelForModel(opts []string, cur string) string {
	if cur == "" {
		return ""
	}
	for _, o := range opts {
		if modelIDFromLabel(o) == cur {
			return o
		}
	}
	return ""
}

// promptCompactFromCtx derives the auto-compact window from a model context
// length: zero ctx disables it, otherwise the user picks a percentage.
func promptCompactFromCtx(p profile.Profile, ctx int) (profile.Profile, error) {
	if ctx <= 0 {
		p.AutoCompactWindow = 0
		return p, nil
	}
	pctStr := fmt.Sprintf("%g", profile.DefaultCompactRatioPercent)
	if err := survey.AskOne(&survey.Input{Message: "压缩窗口比例 (%，默认 80):", Default: pctStr}, &pctStr); err != nil {
		return p, fmt.Errorf("prompt compact ratio: %w", err)
	}
	pct, perr := strconv.ParseFloat(pctStr, 64)
	if perr != nil || pct <= 0 {
		pct = profile.DefaultCompactRatioPercent
	}
	p.AutoCompactWindow = profile.CompactWindow(ctx, pct)
	return p, nil
}

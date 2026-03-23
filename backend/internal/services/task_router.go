package services

import "strings"

type AITextTask string

const (
	TextTaskGeneric        AITextTask = "generic"
	TextTaskLongformScript AITextTask = "longform_script"
	TextTaskNarration      AITextTask = "narration"
	TextTaskStoryboard     AITextTask = "storyboard"
	TextTaskShotPrompt     AITextTask = "shot_prompt"
	TextTaskReview         AITextTask = "review"
)

type TextProvider string

const (
	TextProviderCloudQwen   TextProvider = "cloud_qwen"
	TextProviderLocalVLLM   TextProvider = "local_vllm"
	TextProviderLocalOllama TextProvider = "local_ollama"
	TextProviderHybrid      TextProvider = "hybrid"
)

type VideoTask string

const (
	VideoTaskGeneric   VideoTask = "generic"
	VideoTaskScene     VideoTask = "scene"
	VideoTaskNarration VideoTask = "narration"
	VideoTaskPreview   VideoTask = "preview"
	VideoTaskPremium   VideoTask = "premium"
)

func parseCSVProviders(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	items := strings.Split(raw, ",")
	out := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		normalized := strings.ToLower(strings.TrimSpace(item))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func mergeProviderLists(primary []string, fallback ...string) []string {
	out := make([]string, 0, len(primary)+len(fallback))
	seen := map[string]struct{}{}
	for _, item := range append(primary, fallback...) {
		normalized := strings.ToLower(strings.TrimSpace(item))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

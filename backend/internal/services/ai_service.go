package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/richard9219/3kstory/internal/config"
)

type AIService struct {
	cfg *config.Config
}

func NewAIService(cfg *config.Config) *AIService {
	return &AIService{cfg: cfg}
}

func (s *AIService) GenerateScript(ctx context.Context, prompt string) (*ScriptResult, error) {
	return s.generateScriptForTask(ctx, TextTaskLongformScript, prompt)
}

func (s *AIService) GenerateStoryboard(ctx context.Context, prompt string) (*ScriptResult, error) {
	return s.generateScriptForTask(ctx, TextTaskStoryboard, prompt)
}

func (s *AIService) GenerateShotPromptPack(ctx context.Context, prompt string) (*ScriptResult, error) {
	return s.generateScriptForTask(ctx, TextTaskShotPrompt, prompt)
}

func (s *AIService) textProvidersForTask(task AITextTask) []TextProvider {
	defaultProvider := strings.ToLower(strings.TrimSpace(s.cfg.AI.AIProvider))
	if defaultProvider == "" {
		defaultProvider = string(TextProviderCloudQwen)
	}

	var configured []string
	switch task {
	case TextTaskLongformScript:
		configured = parseCSVProviders(s.cfg.AI.ScriptProviders)
		return s.normalizeTextProviders(mergeProviderLists(configured, defaultProvider, string(TextProviderCloudQwen), string(TextProviderLocalVLLM), string(TextProviderLocalOllama)))
	case TextTaskNarration:
		configured = parseCSVProviders(s.cfg.AI.NarrationProviders)
		return s.normalizeTextProviders(mergeProviderLists(configured, string(TextProviderCloudQwen), defaultProvider, string(TextProviderLocalVLLM), string(TextProviderLocalOllama)))
	case TextTaskStoryboard:
		configured = parseCSVProviders(s.cfg.AI.StoryboardProviders)
		return s.normalizeTextProviders(mergeProviderLists(configured, string(TextProviderLocalVLLM), string(TextProviderLocalOllama), defaultProvider, string(TextProviderCloudQwen)))
	case TextTaskShotPrompt:
		configured = parseCSVProviders(s.cfg.AI.ShotPromptProviders)
		return s.normalizeTextProviders(mergeProviderLists(configured, string(TextProviderLocalVLLM), string(TextProviderLocalOllama), defaultProvider, string(TextProviderCloudQwen)))
	case TextTaskReview:
		configured = parseCSVProviders(s.cfg.AI.ReviewProviders)
		return s.normalizeTextProviders(mergeProviderLists(configured, string(TextProviderCloudQwen), defaultProvider, string(TextProviderLocalVLLM), string(TextProviderLocalOllama)))
	default:
		return s.normalizeTextProviders(mergeProviderLists(nil, defaultProvider, string(TextProviderCloudQwen), string(TextProviderLocalVLLM), string(TextProviderLocalOllama)))
	}
}

func (s *AIService) normalizeTextProviders(items []string) []TextProvider {
	out := make([]TextProvider, 0, len(items))
	for _, item := range items {
		switch TextProvider(item) {
		case TextProviderCloudQwen, TextProviderLocalVLLM, TextProviderLocalOllama, TextProviderHybrid:
			out = append(out, TextProvider(item))
		}
	}
	if len(out) == 0 {
		return []TextProvider{TextProviderCloudQwen}
	}
	return out
}

func (s *AIService) generateScriptForTask(ctx context.Context, task AITextTask, prompt string) (*ScriptResult, error) {
	providers := s.textProvidersForTask(task)
	var lastErr error
	for _, provider := range providers {
		res, err := s.generateScriptWithProvider(ctx, provider, prompt)
		if err == nil {
			return res, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no configured text providers available for task %s", task)
	}
	return nil, lastErr
}

func (s *AIService) generateScriptWithProvider(ctx context.Context, provider TextProvider, prompt string) (*ScriptResult, error) {
	switch provider {
	case TextProviderLocalVLLM:
		return s.generateScriptWithVLLM(ctx, prompt)
	case TextProviderLocalOllama:
		return s.generateScriptWithOllama(ctx, prompt)
	case TextProviderHybrid:
		if res, err := s.generateScriptWithVLLM(ctx, prompt); err == nil {
			return res, nil
		}
		if res, err := s.generateScriptWithOllama(ctx, prompt); err == nil {
			return res, nil
		}
		return s.generateScriptWithCloudQwen(ctx, prompt)
	case TextProviderCloudQwen:
		fallthrough
	default:
		return s.generateScriptWithCloudQwen(ctx, prompt)
	}
}

func (s *AIService) scriptSystemPrompt() string {
	return `你是一个专业的短剧编剧和分镜导演。你必须只输出严格JSON，不要输出任何解释、Markdown、代码块标记。

JSON Schema（必须完全符合）：
{
  "title": "string",
  "genre": "string",
  "style": "string",
  "scenes": [
    {
      "scene_number": 1,
      "title": "string",
      "location": "string",
      "characters": [{"name": "string", "emotion": "string"}],
      "dialogue": "string",
      "shot_type": "string",
      "duration": 10
    }
  ]
}`
}

func (s *AIService) generateScriptWithCloudQwen(ctx context.Context, prompt string) (*ScriptResult, error) {
	requestBody := map[string]interface{}{
		"model": "qwen-plus",
		"messages": []map[string]string{
			{"role": "system", "content": s.scriptSystemPrompt()},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.8,
	}

	jsonData, _ := json.Marshal(requestBody)
	req, err := http.NewRequestWithContext(ctx, "POST", s.cfg.AI.QwenAPIBase+"/services/aigc/text-generation/generation", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.cfg.AI.QwenAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.AI.QwenAPIKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cloud_qwen API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Output struct {
			Text string `json:"text"`
		} `json:"output"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	return parseScriptResult(apiResp.Output.Text)
}

func (s *AIService) generateScriptWithVLLM(ctx context.Context, prompt string) (*ScriptResult, error) {
	baseURL := strings.TrimRight(s.cfg.AI.VLLMBaseURL, "/")
	endpoint := baseURL + "/v1/chat/completions"

	requestBody := map[string]interface{}{
		"model": s.cfg.AI.VLLMModelName,
		"messages": []map[string]string{
			{"role": "system", "content": s.scriptSystemPrompt()},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.8,
		"max_tokens":  s.cfg.AI.VLLMMaxTokens,
	}

	jsonData, _ := json.Marshal(requestBody)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(s.cfg.AI.VLLMTimeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("vLLM API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}
	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("vLLM returned no choices")
	}

	return parseScriptResult(apiResp.Choices[0].Message.Content)
}

func (s *AIService) generateScriptWithOllama(ctx context.Context, prompt string) (*ScriptResult, error) {
	baseURL := strings.TrimRight(s.cfg.AI.OLLAMABaseURL, "/")
	endpoint := baseURL + "/api/generate"

	// Ollama 的 /api/generate 不是 chat messages；这里把系统约束拼到 prompt 前面。
	fullPrompt := s.scriptSystemPrompt() + "\n\n用户需求：\n" + prompt

	requestBody := map[string]interface{}{
		"model":  s.cfg.AI.OLLAMAModelName,
		"prompt": fullPrompt,
		"stream": false,
		"options": map[string]interface{}{
			"num_predict": s.cfg.AI.OLLAMAMaxTokens,
			"temperature": 0.8,
		},
	}

	jsonData, _ := json.Marshal(requestBody)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(s.cfg.AI.OLLAMATimeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	return parseScriptResult(apiResp.Response)
}

func parseScriptResult(raw string) (*ScriptResult, error) {
	trimmed := strings.TrimSpace(raw)
	var result ScriptResult
	if err := json.Unmarshal([]byte(trimmed), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as ScriptResult JSON: %w", err)
	}
	return &result, nil
}

func (s *AIService) GenerateImage(ctx context.Context, prompt string) (string, error) {
	return fmt.Sprintf("https://placeholder.com/800x600?text=%s", prompt), nil
}

func (s *AIService) GenerateVideo(ctx context.Context, prompt string) (string, error) {
	return fmt.Sprintf("https://placeholder.com/video/%s.mp4", prompt), nil
}

type ScriptResult struct {
	Title  string        `json:"title"`
	Genre  string        `json:"genre"`
	Style  string        `json:"style"`
	Scenes []SceneDetail `json:"scenes"`
}

type SceneDetail struct {
	SceneNumber int          `json:"scene_number"`
	Title       string       `json:"title"`
	Location    string       `json:"location"`
	Characters  []CharDetail `json:"characters"`
	Dialogue    string       `json:"dialogue"`
	ShotType    string       `json:"shot_type"`
	Duration    int          `json:"duration"`
}

type CharDetail struct {
	Name    string `json:"name"`
	Emotion string `json:"emotion"`
}

type NarrationScriptRequest struct {
	MovieTitle     string
	Synopsis       string
	Style          string
	TargetDuration int
	CreativeBrief  string
}

type NarrationSegment struct {
	Title             string `json:"title"`
	NarrationText     string `json:"narration_text"`
	EstimatedDuration int    `json:"estimated_duration"`
}

type NarrationScriptResult struct {
	Title    string             `json:"title"`
	Style    string             `json:"style"`
	Segments []NarrationSegment `json:"segments"`
}

// GenerateNarrationScript 里程碑 1 版本：复用现有剧本能力组装解说分段。
func (s *AIService) GenerateNarrationScript(ctx context.Context, req NarrationScriptRequest) (*NarrationScriptResult, error) {
	if req.Style == "" {
		req.Style = "深度分析"
	}
	if req.TargetDuration <= 0 {
		req.TargetDuration = 90
	}

	prompt := fmt.Sprintf("请为《%s》生成%s风格的电影/电视剧解说稿，目标时长约%d秒。剧情简介：%s", req.MovieTitle, req.Style, req.TargetDuration, req.Synopsis)
	if extra := strings.TrimSpace(req.CreativeBrief); extra != "" {
		prompt += "。额外创作要求：" + extra
	}
	script, err := s.generateScriptForTask(ctx, TextTaskNarration, prompt)
	if err != nil {
		segments := []NarrationSegment{{
			Title:             "开场",
			NarrationText:     fmt.Sprintf("今天我们来解说《%s》。%s", req.MovieTitle, req.Synopsis),
			EstimatedDuration: req.TargetDuration,
		}}
		return &NarrationScriptResult{Title: req.MovieTitle + "解说", Style: req.Style, Segments: segments}, nil
	}

	segments := make([]NarrationSegment, 0, len(script.Scenes))
	each := req.TargetDuration
	if len(script.Scenes) > 0 {
		each = req.TargetDuration / len(script.Scenes)
		if each < 8 {
			each = 8
		}
	}
	for _, sc := range script.Scenes {
		text := strings.TrimSpace(sc.Dialogue)
		if text == "" {
			text = fmt.Sprintf("场景发生在%s，围绕%s展开。", sc.Location, sc.Title)
		}
		segments = append(segments, NarrationSegment{
			Title:             sc.Title,
			NarrationText:     text,
			EstimatedDuration: each,
		})
	}
	if len(segments) == 0 {
		segments = append(segments, NarrationSegment{
			Title:             "开场",
			NarrationText:     fmt.Sprintf("今天我们来解说《%s》。", req.MovieTitle),
			EstimatedDuration: req.TargetDuration,
		})
	}

	title := script.Title
	if strings.TrimSpace(title) == "" {
		title = req.MovieTitle + "解说"
	}

	return &NarrationScriptResult{
		Title:    title,
		Style:    req.Style,
		Segments: segments,
	}, nil
}

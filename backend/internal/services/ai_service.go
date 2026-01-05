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
	switch strings.ToLower(strings.TrimSpace(s.cfg.AI.AIProvider)) {
	case "local_vllm":
		return s.generateScriptWithVLLM(ctx, prompt)
	case "local_ollama":
		return s.generateScriptWithOllama(ctx, prompt)
	case "hybrid":
		// Minimal failover: vLLM -> Ollama -> cloud_qwen
		if res, err := s.generateScriptWithVLLM(ctx, prompt); err == nil {
			return res, nil
		}
		if res, err := s.generateScriptWithOllama(ctx, prompt); err == nil {
			return res, nil
		}
		return s.generateScriptWithCloudQwen(ctx, prompt)
	case "cloud_qwen", "":
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

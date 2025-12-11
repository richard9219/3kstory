package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LocalQwenConfig 本地 Qwen 模型配置
type LocalQwenConfig struct {
	BaseURL    string // vLLM 或 Ollama 的 API 地址
	ModelName  string // 模型名称
	MaxTokens  int    // 最大生成 token 数
	APIType    string // "vllm" 或 "ollama"
	Timeout    time.Duration
}

// LocalQwenService 本地 Qwen 模型服务
type LocalQwenService struct {
	config     LocalQwenConfig
	httpClient *http.Client
}

// NewLocalQwenService 创建本地 Qwen 服务实例
func NewLocalQwenService(cfg LocalQwenConfig) *LocalQwenService {
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}
	
	return &LocalQwenService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// GenerateScript 生成短剧脚本
func (s *LocalQwenService) GenerateScript(ctx context.Context, prompt string) (string, error) {
	if s.config.APIType == "ollama" {
		return s.generateWithOllama(ctx, prompt)
	}
	return s.generateWithVLLM(ctx, prompt)
}

// generateWithVLLM 使用 vLLM API 生成
func (s *LocalQwenService) generateWithVLLM(ctx context.Context, prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": s.config.ModelName,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是一个专业的短剧编剧，擅长创作引人入胜的故事情节。",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  s.config.MaxTokens,
		"temperature": 0.7,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		s.config.BaseURL+"/v1/chat/completions",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response generated")
	}
	
	return result.Choices[0].Message.Content, nil
}

// generateWithOllama 使用 Ollama API 生成
func (s *LocalQwenService) generateWithOllama(ctx context.Context, prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":  s.config.ModelName,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"num_predict": s.config.MaxTokens,
			"temperature": 0.7,
		},
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		s.config.BaseURL+"/api/generate",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Response string `json:"response"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Response, nil
}

// HealthCheck 检查服务健康状态
func (s *LocalQwenService) HealthCheck(ctx context.Context) error {
	var endpoint string
	if s.config.APIType == "ollama" {
		endpoint = s.config.BaseURL + "/api/tags"
	} else {
		endpoint = s.config.BaseURL + "/v1/models"
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}
	
	return nil
}

// ReviewContent 使用 Qwen2-VL 审核内容
func (s *LocalQwenService) ReviewContent(ctx context.Context, content string) (bool, string, error) {
	prompt := fmt.Sprintf(`请审核以下内容是否符合规范，是否包含敏感信息：

内容：
%s

请以JSON格式返回审核结果：
{"approved": true/false, "reason": "审核原因"}`, content)

	response, err := s.GenerateScript(ctx, prompt)
	if err != nil {
		return false, "", err
	}
	
	var result struct {
		Approved bool   `json:"approved"`
		Reason   string `json:"reason"`
	}
	
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// 如果无法解析 JSON，则默认通过
		return true, "无法解析审核结果，默认通过", nil
	}
	
	return result.Approved, result.Reason, nil
}

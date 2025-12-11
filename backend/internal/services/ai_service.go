package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	systemPrompt := `You are a professional screenwriter for short dramas.`

	requestBody := map[string]interface{}{
		"model": "qwen-plus",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
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
	req.Header.Set("Authorization", "Bearer "+s.cfg.AI.QwenAPIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var apiResp struct {
		Output struct {
			Text string `json:"text"`
		} `json:"output"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var result ScriptResult
	if err := json.Unmarshal([]byte(apiResp.Output.Text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
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

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

// VideoProvider defines which video generation service to use
type VideoProvider string

const (
	ProviderRunway VideoProvider = "runway"
	ProviderPika   VideoProvider = "pika"
)

// VideoService handles video generation via third-party APIs
type VideoService struct {
	cfg *config.Config
}

func NewVideoService(cfg *config.Config) *VideoService {
	return &VideoService{cfg: cfg}
}

// VideoGenerationRequest represents a video generation request
type VideoGenerationRequest struct {
	ProjectID   uint
	SceneID     uint
	Prompt      string
	Provider    VideoProvider
	ImageURL    string // for image-to-video
	Duration    int    // seconds (1-60)
	AspectRatio string // "16:9" or "9:16"
}

// VideoGenerationResult represents the result of video generation
type VideoGenerationResult struct {
	VideoID     string
	VideoURL    string
	Provider    VideoProvider
	Status      string // "pending", "processing", "completed", "failed"
	Duration    int
	Resolution  string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// GenerateVideo handles video generation with specified provider
func (s *VideoService) GenerateVideo(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	switch req.Provider {
	case ProviderRunway:
		return s.generateWithRunway(ctx, req)
	case ProviderPika:
		return s.generateWithPika(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported video provider: %s", req.Provider)
	}
}

// generateWithRunway generates video using Runway API
func (s *VideoService) generateWithRunway(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	// Runway API endpoint for Gen-3 model
	endpoint := "https://api.runwayml.com/v1/generations"

	requestBody := map[string]interface{}{
		"model": "gen3",
		"prompt": map[string]interface{}{
			"text": req.Prompt,
		},
		"duration":     req.Duration,
		"aspect_ratio": req.AspectRatio,
	}

	// If image URL provided, use image-to-video generation
	if req.ImageURL != "" {
		requestBody["prompt"] = map[string]interface{}{
			"image": req.ImageURL,
			"text":  req.Prompt,
		}
	}

	jsonData, _ := json.Marshal(requestBody)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.AI.RunwayAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("runway API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("runway API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		ID        string   `json:"id"`
		Status    string   `json:"status"`
		Output    []string `json:"output"`
		CreatedAt string   `json:"created_at"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse runway response: %w", err)
	}

	result := &VideoGenerationResult{
		VideoID:    apiResp.ID,
		Provider:   ProviderRunway,
		Status:     apiResp.Status,
		Duration:   req.Duration,
		Resolution: "1080p",
		CreatedAt:  time.Now(),
	}

	// If generation completed immediately
	if len(apiResp.Output) > 0 {
		result.VideoURL = apiResp.Output[0]
		result.Status = "completed"
		now := time.Now()
		result.CompletedAt = &now
	}

	return result, nil
}

// generateWithPika generates video using Pika API
func (s *VideoService) generateWithPika(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	// Pika API endpoint
	endpoint := "https://api.pika.art/v1/generations"

	// Prepare request body
	requestBody := map[string]interface{}{
		"prompt":       req.Prompt,
		"duration":     req.Duration,
		"aspect_ratio": req.AspectRatio,
	}

	// If image URL provided, use image expansion mode
	if req.ImageURL != "" {
		requestBody["mode"] = "image-expand"
		requestBody["image_url"] = req.ImageURL
	}

	jsonData, _ := json.Marshal(requestBody)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.AI.PikaAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("pika API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("pika API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		GenerationID string `json:"generation_id"`
		Status       string `json:"status"`
		VideoURL     string `json:"video_url"`
		CreatedAt    string `json:"created_at"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse pika response: %w", err)
	}

	result := &VideoGenerationResult{
		VideoID:    apiResp.GenerationID,
		VideoURL:   apiResp.VideoURL,
		Provider:   ProviderPika,
		Status:     apiResp.Status,
		Duration:   req.Duration,
		Resolution: "1080p",
		CreatedAt:  time.Now(),
	}

	return result, nil
}

// PollVideoStatus checks the status of a video generation job
func (s *VideoService) PollVideoStatus(ctx context.Context, videoID string, provider VideoProvider) (*VideoGenerationResult, error) {
	var endpoint string
	var authHeader string

	switch provider {
	case ProviderRunway:
		endpoint = fmt.Sprintf("https://api.runwayml.com/v1/generations/%s", videoID)
		authHeader = "Bearer " + s.cfg.AI.RunwayAPIKey
	case ProviderPika:
		endpoint = fmt.Sprintf("https://api.pika.art/v1/generations/%s", videoID)
		authHeader = "Bearer " + s.cfg.AI.PikaAPIKey
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("status poll failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status poll error (status %d): %s", resp.StatusCode, string(body))
	}

	var statusResp struct {
		ID       string   `json:"id"`
		Status   string   `json:"status"`
		Output   []string `json:"output"`
		VideoURL string   `json:"video_url"`
	}

	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to parse status response: %w", err)
	}

	result := &VideoGenerationResult{
		VideoID:   statusResp.ID,
		Provider:  provider,
		Status:    statusResp.Status,
		CreatedAt: time.Now(),
	}

	// Get video URL from response
	if statusResp.VideoURL != "" {
		result.VideoURL = statusResp.VideoURL
	} else if len(statusResp.Output) > 0 {
		result.VideoURL = statusResp.Output[0]
	}

	return result, nil
}

// FailoverGenerate attempts to generate video with primary provider, falls back to secondary
func (s *VideoService) FailoverGenerate(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	// Try primary provider
	result, err := s.GenerateVideo(ctx, req)
	if err == nil {
		return result, nil
	}

	// Log the error and attempt fallback
	fmt.Printf("Primary provider %s failed: %v, attempting fallback\n", req.Provider, err)

	// Switch to fallback provider
	if req.Provider == ProviderRunway {
		req.Provider = ProviderPika
	} else {
		req.Provider = ProviderRunway
	}

	result, fallbackErr := s.GenerateVideo(ctx, req)
	if fallbackErr != nil {
		return nil, fmt.Errorf("both providers failed. Primary: %w, Fallback: %w", err, fallbackErr)
	}

	return result, nil
}

// GenerateVideoTask represents an async video generation task
type GenerateVideoTask struct {
	ID          uint
	ProjectID   uint
	SceneID     uint
	Status      string // pending, processing, completed, failed
	VideoURL    string
	ErrorMsg    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

// SaveVideoTask persists a video generation task
func (s *VideoService) SaveVideoTask(ctx context.Context, task *GenerateVideoTask) error {
	// This would be implemented with database persistence
	// For now, placeholder
	return nil
}

// GetVideoTask retrieves a video generation task by ID
func (s *VideoService) GetVideoTask(ctx context.Context, taskID uint) (*GenerateVideoTask, error) {
	// This would be implemented with database retrieval
	// For now, placeholder
	return nil, nil
}

// ListVideoTasks retrieves all video tasks for a project
func (s *VideoService) ListVideoTasks(ctx context.Context, projectID uint) ([]*GenerateVideoTask, error) {
	// This would be implemented with database query
	// For now, placeholder
	return nil, nil
}

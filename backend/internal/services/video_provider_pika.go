package services

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/richard9219/3kstory/internal/config"
)

type pikaVideoProvider struct {
	runtime providerRuntime
	key     string
}

func newPikaVideoProvider(cfg *config.Config) videoProviderClient {
	return &pikaVideoProvider{runtime: providerRuntime{cfg: cfg}, key: strings.TrimSpace(cfg.AI.PikaAPIKey)}
}

func (p *pikaVideoProvider) Name() VideoProvider { return ProviderPika }

func (p *pikaVideoProvider) ValidateConfig() error {
	if p.key == "" {
		return p.runtime.configError(ProviderPika, "PIKA_API_KEY is not configured")
	}
	return nil
}

func (p *pikaVideoProvider) HealthCheck(ctx context.Context) ProviderHealth {
	now := time.Now().Format(time.RFC3339)
	if err := p.ValidateConfig(); err != nil {
		return ProviderHealth{
			Provider:   ProviderPika,
			Configured: false,
			Healthy:    false,
			Message:    err.Error(),
			ErrorKind:  string(VideoErrorConfig),
			CheckedAt:  now,
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.pika.art/v1/generations/health", nil)
	if err != nil {
		return ProviderHealth{Provider: ProviderPika, Configured: true, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorProvider), CheckedAt: now}
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.key)
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(httpReq)
	if err != nil {
		return ProviderHealth{Provider: ProviderPika, Configured: true, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorProvider), CheckedAt: now}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return ProviderHealth{Provider: ProviderPika, Configured: true, Healthy: true, Message: "Pika endpoint reachable", CheckedAt: now}
	}

	return ProviderHealth{Provider: ProviderPika, Configured: true, Healthy: false, Message: "Pika endpoint unavailable", ErrorKind: string(VideoErrorProvider), CheckedAt: now}
}

func (p *pikaVideoProvider) Generate(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	if err := p.ValidateConfig(); err != nil {
		return nil, err
	}
	requestBody := map[string]interface{}{
		"prompt":       req.Prompt,
		"duration":     req.Duration,
		"aspect_ratio": req.AspectRatio,
	}
	if req.ImageURL != "" {
		requestBody["mode"] = "image-expand"
		requestBody["image_url"] = req.ImageURL
	}
	jsonData, _ := json.Marshal(requestBody)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.pika.art/v1/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.key)
	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, &VideoProviderError{Provider: ProviderPika, Kind: VideoErrorProvider, Retryable: true, Message: "request failed", Cause: err}
	}
	defer resp.Body.Close()
	body, _ := ioReadAll(resp)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, p.runtime.classifyHTTPError(ProviderPika, resp.StatusCode, body, "generation")
	}
	var apiResp struct {
		GenerationID string `json:"generation_id"`
		ID           string `json:"id"`
		Status       string `json:"status"`
		VideoURL     string `json:"video_url"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &VideoProviderError{Provider: ProviderPika, Kind: VideoErrorProvider, Message: "invalid response payload", Cause: err}
	}
	result := &VideoGenerationResult{VideoID: firstNonEmpty(apiResp.GenerationID, apiResp.ID), VideoURL: apiResp.VideoURL, Provider: ProviderPika, Status: normalizeVideoStatus(apiResp.Status), Duration: req.Duration, Resolution: "1080p", CreatedAt: time.Now()}
	if result.VideoURL != "" && result.Status == "pending" {
		result.Status = "completed"
	}
	return result, nil
}

func (p *pikaVideoProvider) PollStatus(ctx context.Context, videoID string) (*VideoGenerationResult, error) {
	if err := p.ValidateConfig(); err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.pika.art/v1/generations/"+videoID, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.key)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, p.runtime.pollingError(ProviderPika, "status request failed", err)
	}
	defer resp.Body.Close()
	body, _ := ioReadAll(resp)
	if resp.StatusCode != http.StatusOK {
		return nil, p.runtime.classifyHTTPError(ProviderPika, resp.StatusCode, body, "status polling")
	}
	var statusResp struct {
		GenerationID string   `json:"generation_id"`
		ID           string   `json:"id"`
		Status       string   `json:"status"`
		Output       []string `json:"output"`
		VideoURL     string   `json:"video_url"`
	}
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, p.runtime.pollingError(ProviderPika, "invalid status payload", err)
	}
	result := &VideoGenerationResult{VideoID: firstNonEmpty(statusResp.GenerationID, statusResp.ID, videoID), Provider: ProviderPika, Status: normalizeVideoStatus(statusResp.Status), CreatedAt: time.Now()}
	if statusResp.VideoURL != "" {
		result.VideoURL = statusResp.VideoURL
	} else if len(statusResp.Output) > 0 {
		result.VideoURL = statusResp.Output[0]
	}
	return result, nil
}

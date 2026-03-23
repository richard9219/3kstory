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

type runwayVideoProvider struct {
	runtime providerRuntime
	key     string
}

func newRunwayVideoProvider(cfg *config.Config) videoProviderClient {
	return &runwayVideoProvider{runtime: providerRuntime{cfg: cfg}, key: strings.TrimSpace(cfg.AI.RunwayAPIKey)}
}

func (p *runwayVideoProvider) Name() VideoProvider { return ProviderRunway }

func (p *runwayVideoProvider) ValidateConfig() error {
	if p.key == "" {
		return p.runtime.configError(ProviderRunway, "RUNWAY_API_KEY is not configured")
	}
	return nil
}

func (p *runwayVideoProvider) HealthCheck(ctx context.Context) ProviderHealth {
	now := time.Now().Format(time.RFC3339)
	if err := p.ValidateConfig(); err != nil {
		return ProviderHealth{
			Provider:   ProviderRunway,
			Configured: false,
			Healthy:    false,
			Message:    err.Error(),
			ErrorKind:  string(VideoErrorConfig),
			CheckedAt:  now,
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.runwayml.com/v1/generations/health", nil)
	if err != nil {
		return ProviderHealth{Provider: ProviderRunway, Configured: true, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorProvider), CheckedAt: now}
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.key)
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(httpReq)
	if err != nil {
		return ProviderHealth{Provider: ProviderRunway, Configured: true, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorProvider), CheckedAt: now}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return ProviderHealth{Provider: ProviderRunway, Configured: true, Healthy: true, Message: "Runway endpoint reachable", CheckedAt: now}
	}

	return ProviderHealth{Provider: ProviderRunway, Configured: true, Healthy: false, Message: "Runway endpoint unavailable", ErrorKind: string(VideoErrorProvider), CheckedAt: now}
}

func (p *runwayVideoProvider) Generate(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	if err := p.ValidateConfig(); err != nil {
		return nil, err
	}
	requestBody := map[string]interface{}{
		"model": "gen3",
		"prompt": map[string]interface{}{
			"text": req.Prompt,
		},
		"duration":     req.Duration,
		"aspect_ratio": req.AspectRatio,
	}
	if req.ImageURL != "" {
		requestBody["prompt"] = map[string]interface{}{
			"image": req.ImageURL,
			"text":  req.Prompt,
		}
	}
	jsonData, _ := json.Marshal(requestBody)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.runwayml.com/v1/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.key)

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, &VideoProviderError{Provider: ProviderRunway, Kind: VideoErrorProvider, Retryable: true, Message: "request failed", Cause: err}
	}
	defer resp.Body.Close()
	body, _ := ioReadAll(resp)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, p.runtime.classifyHTTPError(ProviderRunway, resp.StatusCode, body, "generation")
	}
	var apiResp struct {
		ID     string   `json:"id"`
		Status string   `json:"status"`
		Output []string `json:"output"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &VideoProviderError{Provider: ProviderRunway, Kind: VideoErrorProvider, Message: "invalid response payload", Cause: err}
	}
	result := &VideoGenerationResult{VideoID: apiResp.ID, Provider: ProviderRunway, Status: normalizeVideoStatus(apiResp.Status), Duration: req.Duration, Resolution: "1080p", CreatedAt: time.Now()}
	if len(apiResp.Output) > 0 {
		result.VideoURL = apiResp.Output[0]
		result.Status = "completed"
		now := time.Now()
		result.CompletedAt = &now
	}
	return result, nil
}

func (p *runwayVideoProvider) PollStatus(ctx context.Context, videoID string) (*VideoGenerationResult, error) {
	if err := p.ValidateConfig(); err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.runwayml.com/v1/generations/"+videoID, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.key)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, p.runtime.pollingError(ProviderRunway, "status request failed", err)
	}
	defer resp.Body.Close()
	body, _ := ioReadAll(resp)
	if resp.StatusCode != http.StatusOK {
		return nil, p.runtime.classifyHTTPError(ProviderRunway, resp.StatusCode, body, "status polling")
	}
	var statusResp struct {
		ID       string   `json:"id"`
		Status   string   `json:"status"`
		Output   []string `json:"output"`
		VideoURL string   `json:"video_url"`
	}
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, p.runtime.pollingError(ProviderRunway, "invalid status payload", err)
	}
	result := &VideoGenerationResult{VideoID: firstNonEmpty(statusResp.ID, videoID), Provider: ProviderRunway, Status: normalizeVideoStatus(statusResp.Status), CreatedAt: time.Now()}
	if statusResp.VideoURL != "" {
		result.VideoURL = statusResp.VideoURL
	} else if len(statusResp.Output) > 0 {
		result.VideoURL = statusResp.Output[0]
	}
	return result, nil
}

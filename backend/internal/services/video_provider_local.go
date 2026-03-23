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

type localVideoProvider struct {
	runtime providerRuntime
	url     string
}

func newLocalVideoProvider(cfg *config.Config) videoProviderClient {
	return &localVideoProvider{runtime: providerRuntime{cfg: cfg}, url: strings.TrimRight(strings.TrimSpace(cfg.AI.VideoServiceURL), "/")}
}

func (p *localVideoProvider) Name() VideoProvider { return ProviderLocal }

func (p *localVideoProvider) ValidateConfig() error {
	if p.url == "" {
		return p.runtime.configError(ProviderLocal, "AI_VIDEO_SERVICE_URL is not configured")
	}
	return nil
}

func (p *localVideoProvider) HealthCheck(ctx context.Context) ProviderHealth {
	now := time.Now().Format(time.RFC3339)
	if err := p.ValidateConfig(); err != nil {
		return ProviderHealth{
			Provider:   ProviderLocal,
			Configured: false,
			Healthy:    false,
			Message:    err.Error(),
			ErrorKind:  string(VideoErrorConfig),
			CheckedAt:  now,
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url+"/health", nil)
	if err != nil {
		return ProviderHealth{Provider: ProviderLocal, Configured: true, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorProvider), CheckedAt: now}
	}
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(httpReq)
	if err != nil {
		return ProviderHealth{Provider: ProviderLocal, Configured: true, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorProvider), CheckedAt: now}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return ProviderHealth{Provider: ProviderLocal, Configured: true, Healthy: true, Message: "Local video service reachable", CheckedAt: now}
	}

	return ProviderHealth{Provider: ProviderLocal, Configured: true, Healthy: false, Message: "Local video service unavailable", ErrorKind: string(VideoErrorProvider), CheckedAt: now}
}

func (p *localVideoProvider) Generate(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	if err := p.ValidateConfig(); err != nil {
		return nil, err
	}
	requestBody := map[string]interface{}{
		"prompt":            req.Prompt,
		"image_url":         req.ImageURL,
		"duration":          req.Duration,
		"aspect_ratio":      req.AspectRatio,
		"scene_id":          req.SceneID,
		"project_id":        req.ProjectID,
		"source_video_path": req.SourceVideoPath,
		"source_video_url":  req.SourceVideoURL,
	}
	if strings.TrimSpace(req.Mode) != "" {
		requestBody["mode"] = req.Mode
	}
	if len(req.NarrationSegments) > 0 {
		requestBody["segments"] = req.NarrationSegments
	}
	jsonData, _ := json.Marshal(requestBody)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, &VideoProviderError{Provider: ProviderLocal, Kind: VideoErrorProvider, Retryable: true, Message: "request failed", Cause: err}
	}
	defer resp.Body.Close()
	body, _ := ioReadAll(resp)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, p.runtime.classifyHTTPError(ProviderLocal, resp.StatusCode, body, "generation")
	}
	var apiResp struct {
		VideoID  string   `json:"video_id"`
		Status   string   `json:"status"`
		VideoURL string   `json:"video_url"`
		ID       string   `json:"id"`
		Output   []string `json:"output"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &VideoProviderError{Provider: ProviderLocal, Kind: VideoErrorProvider, Message: "invalid response payload", Cause: err}
	}
	result := &VideoGenerationResult{VideoID: firstNonEmpty(apiResp.VideoID, apiResp.ID), VideoURL: apiResp.VideoURL, Provider: ProviderLocal, Status: normalizeVideoStatus(apiResp.Status), Duration: req.Duration, CreatedAt: time.Now()}
	if result.VideoURL == "" && len(apiResp.Output) > 0 {
		result.VideoURL = apiResp.Output[0]
	}
	return result, nil
}

func (p *localVideoProvider) PollStatus(ctx context.Context, videoID string) (*VideoGenerationResult, error) {
	if err := p.ValidateConfig(); err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url+"/"+videoID, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, p.runtime.pollingError(ProviderLocal, "status request failed", err)
	}
	defer resp.Body.Close()
	body, _ := ioReadAll(resp)
	if resp.StatusCode != http.StatusOK {
		return nil, p.runtime.classifyHTTPError(ProviderLocal, resp.StatusCode, body, "status polling")
	}
	var statusResp struct {
		ID       string   `json:"id"`
		Status   string   `json:"status"`
		Output   []string `json:"output"`
		VideoURL string   `json:"video_url"`
	}
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, p.runtime.pollingError(ProviderLocal, "invalid status payload", err)
	}
	result := &VideoGenerationResult{VideoID: firstNonEmpty(statusResp.ID, videoID), Provider: ProviderLocal, Status: normalizeVideoStatus(statusResp.Status), CreatedAt: time.Now()}
	if statusResp.VideoURL != "" {
		result.VideoURL = statusResp.VideoURL
	} else if len(statusResp.Output) > 0 {
		result.VideoURL = statusResp.Output[0]
	}
	return result, nil
}

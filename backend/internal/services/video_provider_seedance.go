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

type seedanceVideoProvider struct {
	runtime    providerRuntime
	key        string
	baseURL    string
	createPath string
	statusPath string
	model      string
	defaultRes string
}

func newSeedanceVideoProvider(cfg *config.Config) videoProviderClient {
	return &seedanceVideoProvider{
		runtime:    providerRuntime{cfg: cfg},
		key:        strings.TrimSpace(cfg.AI.SeedanceAPIKey),
		baseURL:    strings.TrimRight(strings.TrimSpace(cfg.AI.SeedanceAPIBase), "/"),
		createPath: strings.TrimSpace(cfg.AI.SeedanceCreatePath),
		statusPath: strings.TrimSpace(cfg.AI.SeedanceStatusPath),
		model:      strings.TrimSpace(cfg.AI.SeedanceVideoModel),
		defaultRes: strings.TrimSpace(cfg.AI.SeedanceVideoResolution),
	}
}

func (p *seedanceVideoProvider) Name() VideoProvider { return ProviderSeedance }

func (p *seedanceVideoProvider) ValidateConfig() error {
	if p.key == "" {
		return p.runtime.configError(ProviderSeedance, "SEEDANCE_API_KEY is not configured")
	}
	if p.baseURL == "" {
		return p.runtime.configError(ProviderSeedance, "SEEDANCE_API_BASE is not configured")
	}
	if p.createPath == "" || p.statusPath == "" {
		return p.runtime.configError(ProviderSeedance, "SEEDANCE_CREATE_PATH / SEEDANCE_STATUS_PATH are not configured")
	}
	return nil
}

func (p *seedanceVideoProvider) HealthCheck(ctx context.Context) ProviderHealth {
	now := time.Now().Format(time.RFC3339)
	if err := p.ValidateConfig(); err != nil {
		return ProviderHealth{Provider: ProviderSeedance, Configured: false, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorConfig), CheckedAt: now}
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodOptions, joinBaseAndPath(p.baseURL, p.createPath), nil)
	if err != nil {
		return ProviderHealth{Provider: ProviderSeedance, Configured: true, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorProvider), CheckedAt: now}
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.key)
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(httpReq)
	if err != nil {
		return ProviderHealth{Provider: ProviderSeedance, Configured: true, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorProvider), CheckedAt: now}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return ProviderHealth{Provider: ProviderSeedance, Configured: true, Healthy: true, Message: "Seedance endpoint reachable", CheckedAt: now}
	}
	return ProviderHealth{Provider: ProviderSeedance, Configured: true, Healthy: false, Message: "Seedance endpoint unavailable", ErrorKind: string(VideoErrorProvider), CheckedAt: now}
}

func (p *seedanceVideoProvider) Generate(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	if err := p.ValidateConfig(); err != nil {
		return nil, err
	}

	requestBody := map[string]interface{}{
		"model":        firstNonEmpty(req.Model, p.model),
		"prompt":       req.Prompt,
		"duration":     req.Duration,
		"aspect_ratio": req.AspectRatio,
		"resolution":   firstNonEmpty(req.Resolution, p.defaultRes),
	}
	if requestBody["model"] == "" {
		requestBody["model"] = "seedance-1-0-pro-250528"
	}
	if req.ImageURL != "" {
		requestBody["image_url"] = req.ImageURL
	}
	if req.LastFrameImageURL != "" {
		requestBody["last_frame_image_url"] = req.LastFrameImageURL
	}
	if req.CallbackURL != "" {
		requestBody["callback_url"] = req.CallbackURL
	}
	for key, value := range req.ExtraData {
		requestBody[key] = value
	}

	jsonData, _ := json.Marshal(requestBody)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, joinBaseAndPath(p.baseURL, p.createPath), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.key)

	resp, err := (&http.Client{Timeout: 45 * time.Second}).Do(httpReq)
	if err != nil {
		return nil, &VideoProviderError{Provider: ProviderSeedance, Kind: VideoErrorProvider, Retryable: true, Message: "request failed", Cause: err}
	}
	defer resp.Body.Close()
	body, _ := ioReadAll(resp)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, p.runtime.classifyHTTPError(ProviderSeedance, resp.StatusCode, body, "generation")
	}

	var apiResp struct {
		ID        string `json:"id"`
		TaskID    string `json:"task_id"`
		Status    string `json:"status"`
		VideoURL  string `json:"video_url"`
		OutputURL string `json:"output_url"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &VideoProviderError{Provider: ProviderSeedance, Kind: VideoErrorProvider, Message: "invalid response payload", Cause: err}
	}

	result := &VideoGenerationResult{
		VideoID:    firstNonEmpty(apiResp.TaskID, apiResp.ID),
		VideoURL:   firstNonEmpty(apiResp.VideoURL, apiResp.OutputURL),
		Provider:   ProviderSeedance,
		Status:     normalizeVideoStatus(apiResp.Status),
		Duration:   req.Duration,
		Resolution: firstNonEmpty(req.Resolution, p.defaultRes),
		CreatedAt:  time.Now(),
	}
	if result.Status == "" {
		if result.VideoURL != "" {
			result.Status = "completed"
		} else {
			result.Status = "pending"
		}
	}
	return result, nil
}

func (p *seedanceVideoProvider) PollStatus(ctx context.Context, videoID string) (*VideoGenerationResult, error) {
	if err := p.ValidateConfig(); err != nil {
		return nil, err
	}
	statusURL := joinBaseAndPath(p.baseURL, interpolatePath(p.statusPath, map[string]string{"task_id": videoID}))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.key)
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(httpReq)
	if err != nil {
		return nil, p.runtime.pollingError(ProviderSeedance, "status request failed", err)
	}
	defer resp.Body.Close()
	body, _ := ioReadAll(resp)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, p.runtime.classifyHTTPError(ProviderSeedance, resp.StatusCode, body, "status polling")
	}

	var statusResp struct {
		ID        string `json:"id"`
		TaskID    string `json:"task_id"`
		Status    string `json:"status"`
		VideoURL  string `json:"video_url"`
		OutputURL string `json:"output_url"`
		Result    struct {
			VideoURL  string `json:"video_url"`
			OutputURL string `json:"output_url"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, p.runtime.pollingError(ProviderSeedance, "invalid status payload", err)
	}
	return &VideoGenerationResult{
		VideoID:   firstNonEmpty(statusResp.TaskID, statusResp.ID, videoID),
		VideoURL:  firstNonEmpty(statusResp.VideoURL, statusResp.OutputURL, statusResp.Result.VideoURL, statusResp.Result.OutputURL),
		Provider:  ProviderSeedance,
		Status:    normalizeVideoStatus(statusResp.Status),
		CreatedAt: time.Now(),
	}, nil
}

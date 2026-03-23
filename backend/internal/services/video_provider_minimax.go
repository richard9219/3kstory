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

type miniMaxVideoProvider struct {
	runtime    providerRuntime
	key        string
	baseURL    string
	model      string
	resolution string
}

func newMiniMaxVideoProvider(cfg *config.Config) videoProviderClient {
	return &miniMaxVideoProvider{
		runtime:    providerRuntime{cfg: cfg},
		key:        strings.TrimSpace(cfg.AI.MiniMaxAPIKey),
		baseURL:    strings.TrimRight(strings.TrimSpace(cfg.AI.MiniMaxAPIBase), "/"),
		model:      strings.TrimSpace(cfg.AI.MiniMaxVideoModel),
		resolution: strings.TrimSpace(cfg.AI.MiniMaxVideoResolution),
	}
}

func (p *miniMaxVideoProvider) Name() VideoProvider { return ProviderMiniMax }

func (p *miniMaxVideoProvider) ValidateConfig() error {
	if p.key == "" {
		return p.runtime.configError(ProviderMiniMax, "MINIMAX_API_KEY is not configured")
	}
	if p.baseURL == "" {
		return p.runtime.configError(ProviderMiniMax, "MINIMAX_API_BASE is not configured")
	}
	return nil
}

func (p *miniMaxVideoProvider) HealthCheck(ctx context.Context) ProviderHealth {
	now := time.Now().Format(time.RFC3339)
	if err := p.ValidateConfig(); err != nil {
		return ProviderHealth{Provider: ProviderMiniMax, Configured: false, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorConfig), CheckedAt: now}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, joinBaseAndPath(p.baseURL, "/v1/files/retrieve"), nil)
	if err != nil {
		return ProviderHealth{Provider: ProviderMiniMax, Configured: true, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorProvider), CheckedAt: now}
	}
	req.Header.Set("Authorization", "Bearer "+p.key)
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return ProviderHealth{Provider: ProviderMiniMax, Configured: true, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorProvider), CheckedAt: now}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return ProviderHealth{Provider: ProviderMiniMax, Configured: true, Healthy: true, Message: "MiniMax endpoint reachable", CheckedAt: now}
	}
	return ProviderHealth{Provider: ProviderMiniMax, Configured: true, Healthy: false, Message: "MiniMax endpoint unavailable", ErrorKind: string(VideoErrorProvider), CheckedAt: now}
}

func (p *miniMaxVideoProvider) Generate(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	if err := p.ValidateConfig(); err != nil {
		return nil, err
	}

	model := firstNonEmpty(req.Model, p.model)
	if model == "" {
		model = "MiniMax-Hailuo-2.3-Fast"
	}
	resolution := firstNonEmpty(req.Resolution, p.resolution)
	if resolution == "" {
		resolution = "768P"
	}

	requestBody := map[string]interface{}{
		"model":        model,
		"prompt":       req.Prompt,
		"resolution":   resolution,
		"aspect_ratio": req.AspectRatio,
	}
	if req.Duration > 0 {
		requestBody["duration"] = req.Duration
	}
	if req.Seed > 0 {
		requestBody["seed"] = req.Seed
	}
	if req.ImageURL != "" {
		requestBody["first_frame_image"] = req.ImageURL
	}
	if req.LastFrameImageURL != "" {
		requestBody["last_frame_image"] = req.LastFrameImageURL
	}
	if req.CallbackURL != "" {
		requestBody["callback_url"] = req.CallbackURL
	}
	for key, value := range req.ExtraData {
		requestBody[key] = value
	}

	jsonData, _ := json.Marshal(requestBody)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, joinBaseAndPath(p.baseURL, "/v1/video_generation"), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.key)

	resp, err := (&http.Client{Timeout: 45 * time.Second}).Do(httpReq)
	if err != nil {
		return nil, &VideoProviderError{Provider: ProviderMiniMax, Kind: VideoErrorProvider, Retryable: true, Message: "request failed", Cause: err}
	}
	defer resp.Body.Close()
	body, _ := ioReadAll(resp)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, p.runtime.classifyHTTPError(ProviderMiniMax, resp.StatusCode, body, "generation")
	}

	var apiResp struct {
		TaskID   string `json:"task_id"`
		ID       string `json:"id"`
		Status   string `json:"status"`
		BaseResp struct {
			StatusCode int    `json:"status_code"`
			StatusMsg  string `json:"status_msg"`
		} `json:"base_resp"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &VideoProviderError{Provider: ProviderMiniMax, Kind: VideoErrorProvider, Message: "invalid response payload", Cause: err}
	}

	result := &VideoGenerationResult{
		VideoID:    firstNonEmpty(apiResp.TaskID, apiResp.ID),
		Provider:   ProviderMiniMax,
		Status:     normalizeVideoStatus(apiResp.Status),
		Duration:   req.Duration,
		Resolution: resolution,
		CreatedAt:  time.Now(),
	}
	if result.Status == "" {
		result.Status = "pending"
	}
	return result, nil
}

func (p *miniMaxVideoProvider) PollStatus(ctx context.Context, videoID string) (*VideoGenerationResult, error) {
	if err := p.ValidateConfig(); err != nil {
		return nil, err
	}
	statusURL := setQueryParam(joinBaseAndPath(p.baseURL, "/v1/query/video_generation"), "task_id", videoID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.key)

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(httpReq)
	if err != nil {
		return nil, p.runtime.pollingError(ProviderMiniMax, "status request failed", err)
	}
	defer resp.Body.Close()
	body, _ := ioReadAll(resp)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, p.runtime.classifyHTTPError(ProviderMiniMax, resp.StatusCode, body, "status polling")
	}

	var statusResp struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
		FileID string `json:"file_id"`
		File   struct {
			FileID string `json:"file_id"`
		} `json:"file"`
	}
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, p.runtime.pollingError(ProviderMiniMax, "invalid status payload", err)
	}

	result := &VideoGenerationResult{
		VideoID:   firstNonEmpty(statusResp.TaskID, videoID),
		Provider:  ProviderMiniMax,
		Status:    normalizeVideoStatus(statusResp.Status),
		CreatedAt: time.Now(),
	}

	fileID := firstNonEmpty(statusResp.FileID, statusResp.File.FileID)
	if fileID != "" {
		videoURL, err := p.lookupFileURL(ctx, fileID)
		if err == nil && videoURL != "" {
			result.VideoURL = videoURL
			if result.Status == "pending" {
				result.Status = "completed"
			}
		}
	}

	return result, nil
}

func (p *miniMaxVideoProvider) lookupFileURL(ctx context.Context, fileID string) (string, error) {
	reqURL := setQueryParam(joinBaseAndPath(p.baseURL, "/v1/files/retrieve"), "file_id", fileID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.key)
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioReadAll(resp)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", p.runtime.classifyHTTPError(ProviderMiniMax, resp.StatusCode, body, "file retrieve")
	}
	var fileResp struct {
		File struct {
			FileID      string `json:"file_id"`
			DownloadURL string `json:"download_url"`
			URL         string `json:"url"`
		} `json:"file"`
		DownloadURL string `json:"download_url"`
		URL         string `json:"url"`
	}
	if err := json.Unmarshal(body, &fileResp); err != nil {
		return "", err
	}
	return firstNonEmpty(fileResp.File.DownloadURL, fileResp.File.URL, fileResp.DownloadURL, fileResp.URL), nil
}

package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/richard9219/3kstory/internal/config"
)

type comfyVideoProvider struct {
	runtime      providerRuntime
	baseURL      string
	apiKey       string
	workflowDir  string
	outputNodeID string
}

func newComfyVideoProvider(cfg *config.Config) videoProviderClient {
	return &comfyVideoProvider{
		runtime:      providerRuntime{cfg: cfg},
		baseURL:      strings.TrimRight(strings.TrimSpace(cfg.AI.ComfyBaseURL), "/"),
		apiKey:       strings.TrimSpace(cfg.AI.ComfyAPIKey),
		workflowDir:  strings.TrimSpace(cfg.AI.ComfyWorkflowDir),
		outputNodeID: strings.TrimSpace(cfg.AI.ComfyOutputNodeID),
	}
}

func (p *comfyVideoProvider) Name() VideoProvider { return ProviderComfy }

func (p *comfyVideoProvider) ValidateConfig() error {
	if p.baseURL == "" {
		return p.runtime.configError(ProviderComfy, "COMFY_BASE_URL is not configured")
	}
	return nil
}

func (p *comfyVideoProvider) HealthCheck(ctx context.Context) ProviderHealth {
	now := time.Now().Format(time.RFC3339)
	if err := p.ValidateConfig(); err != nil {
		return ProviderHealth{Provider: ProviderComfy, Configured: false, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorConfig), CheckedAt: now}
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, joinBaseAndPath(p.baseURL, "/history"), nil)
	if err != nil {
		return ProviderHealth{Provider: ProviderComfy, Configured: true, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorProvider), CheckedAt: now}
	}
	p.applyAuth(httpReq)
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(httpReq)
	if err != nil {
		return ProviderHealth{Provider: ProviderComfy, Configured: true, Healthy: false, Message: err.Error(), ErrorKind: string(VideoErrorProvider), CheckedAt: now}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return ProviderHealth{Provider: ProviderComfy, Configured: true, Healthy: true, Message: "Comfy endpoint reachable", CheckedAt: now}
	}
	return ProviderHealth{Provider: ProviderComfy, Configured: true, Healthy: false, Message: "Comfy endpoint unavailable", ErrorKind: string(VideoErrorProvider), CheckedAt: now}
}

func (p *comfyVideoProvider) Generate(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	if err := p.ValidateConfig(); err != nil {
		return nil, err
	}
	workflow, err := p.resolveWorkflow(req)
	if err != nil {
		return nil, &VideoProviderError{Provider: ProviderComfy, Kind: VideoErrorInvalidRequest, Message: err.Error()}
	}
	if err := p.injectWorkflowInputs(workflow, req); err != nil {
		return nil, &VideoProviderError{Provider: ProviderComfy, Kind: VideoErrorInvalidRequest, Message: err.Error()}
	}

	requestBody := map[string]interface{}{
		"prompt":    workflow,
		"client_id": fmt.Sprintf("3kstory-%d", time.Now().UnixNano()),
	}
	jsonData, _ := json.Marshal(requestBody)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, joinBaseAndPath(p.baseURL, "/prompt"), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	p.applyAuth(httpReq)
	resp, err := (&http.Client{Timeout: 45 * time.Second}).Do(httpReq)
	if err != nil {
		return nil, &VideoProviderError{Provider: ProviderComfy, Kind: VideoErrorProvider, Retryable: true, Message: "request failed", Cause: err}
	}
	defer resp.Body.Close()
	body, _ := ioReadAll(resp)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, p.runtime.classifyHTTPError(ProviderComfy, resp.StatusCode, body, "generation")
	}

	var apiResp struct {
		PromptID string `json:"prompt_id"`
		ID       string `json:"id"`
		Number   int    `json:"number"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &VideoProviderError{Provider: ProviderComfy, Kind: VideoErrorProvider, Message: "invalid response payload", Cause: err}
	}

	return &VideoGenerationResult{
		VideoID:    firstNonEmpty(apiResp.PromptID, apiResp.ID),
		Provider:   ProviderComfy,
		Status:     "pending",
		Duration:   req.Duration,
		Resolution: req.Resolution,
		CreatedAt:  time.Now(),
	}, nil
}

func (p *comfyVideoProvider) PollStatus(ctx context.Context, videoID string) (*VideoGenerationResult, error) {
	if err := p.ValidateConfig(); err != nil {
		return nil, err
	}
	reqURL := joinBaseAndPath(p.baseURL, "/history/"+videoID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	p.applyAuth(httpReq)
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(httpReq)
	if err != nil {
		return nil, p.runtime.pollingError(ProviderComfy, "status request failed", err)
	}
	defer resp.Body.Close()
	body, _ := ioReadAll(resp)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, p.runtime.classifyHTTPError(ProviderComfy, resp.StatusCode, body, "status polling")
	}

	var history map[string]struct {
		Status struct {
			StatusStr string `json:"status_str"`
		} `json:"status"`
		Outputs map[string]map[string]interface{} `json:"outputs"`
	}
	if err := json.Unmarshal(body, &history); err != nil {
		return nil, p.runtime.pollingError(ProviderComfy, "invalid status payload", err)
	}
	entry, ok := history[videoID]
	if !ok {
		return &VideoGenerationResult{
			VideoID:   videoID,
			Provider:  ProviderComfy,
			Status:    "processing",
			CreatedAt: time.Now(),
		}, nil
	}

	result := &VideoGenerationResult{
		VideoID:   videoID,
		Provider:  ProviderComfy,
		Status:    normalizeVideoStatus(entry.Status.StatusStr),
		CreatedAt: time.Now(),
	}
	if result.Status == "" {
		result.Status = "processing"
	}
	if videoURL := p.findWorkflowOutputURL(entry.Outputs); videoURL != "" {
		result.VideoURL = videoURL
		if result.Status != "failed" {
			result.Status = "completed"
		}
	}
	return result, nil
}

func (p *comfyVideoProvider) applyAuth(req *http.Request) {
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
		req.Header.Set("X-API-Key", p.apiKey)
	}
}

func (p *comfyVideoProvider) resolveWorkflow(req *VideoGenerationRequest) (map[string]interface{}, error) {
	if len(req.Workflow) > 0 {
		return cloneJSONMap(req.Workflow), nil
	}
	workflowPath := strings.TrimSpace(req.WorkflowPath)
	if workflowPath == "" {
		return nil, fmt.Errorf("comfy provider requires workflow or workflow_path")
	}
	if !filepath.IsAbs(workflowPath) && p.workflowDir != "" {
		workflowPath = filepath.Join(p.workflowDir, workflowPath)
	}
	return loadWorkflowFromPath(workflowPath)
}

func (p *comfyVideoProvider) injectWorkflowInputs(workflow map[string]interface{}, req *VideoGenerationRequest) error {
	overrides, _ := req.ExtraData["workflow_inputs"].(map[string]interface{})
	if len(overrides) > 0 {
		for nodeID, raw := range overrides {
			node, ok := workflow[nodeID].(map[string]interface{})
			if !ok {
				continue
			}
			inputs, _ := node["inputs"].(map[string]interface{})
			if inputs == nil {
				inputs = map[string]interface{}{}
				node["inputs"] = inputs
			}
			if mapped, ok := raw.(map[string]interface{}); ok {
				for key, value := range mapped {
					inputs[key] = value
				}
			}
		}
	}

	promptNodeID, _ := req.ExtraData["prompt_node_id"].(string)
	promptInputName, _ := req.ExtraData["prompt_input_name"].(string)
	if promptInputName == "" {
		promptInputName = "text"
	}
	if promptNodeID != "" && strings.TrimSpace(req.Prompt) != "" {
		if node, ok := workflow[promptNodeID].(map[string]interface{}); ok {
			inputs, _ := node["inputs"].(map[string]interface{})
			if inputs == nil {
				inputs = map[string]interface{}{}
				node["inputs"] = inputs
			}
			inputs[promptInputName] = req.Prompt
		}
	}

	imageNodeID, _ := req.ExtraData["image_node_id"].(string)
	imageInputName, _ := req.ExtraData["image_input_name"].(string)
	if imageInputName == "" {
		imageInputName = "image"
	}
	if imageNodeID != "" && strings.TrimSpace(req.ImageURL) != "" {
		if node, ok := workflow[imageNodeID].(map[string]interface{}); ok {
			inputs, _ := node["inputs"].(map[string]interface{})
			if inputs == nil {
				inputs = map[string]interface{}{}
				node["inputs"] = inputs
			}
			inputs[imageInputName] = req.ImageURL
		}
	}
	return nil
}

func (p *comfyVideoProvider) findWorkflowOutputURL(outputs map[string]map[string]interface{}) string {
	if len(outputs) == 0 {
		return ""
	}

	nodeIDs := make([]string, 0, len(outputs))
	if p.outputNodeID != "" {
		nodeIDs = append(nodeIDs, p.outputNodeID)
	}
	for nodeID := range outputs {
		if nodeID != p.outputNodeID {
			nodeIDs = append(nodeIDs, nodeID)
		}
	}

	for _, nodeID := range nodeIDs {
		node := outputs[nodeID]
		for _, key := range []string{"gifs", "videos", "images"} {
			items, ok := node[key].([]interface{})
			if !ok {
				continue
			}
			for _, item := range items {
				entry, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				if direct, _ := entry["url"].(string); direct != "" {
					return direct
				}
				filename, _ := entry["filename"].(string)
				folderType, _ := entry["type"].(string)
				subfolder, _ := entry["subfolder"].(string)
				if filename != "" {
					viewURL := joinBaseAndPath(p.baseURL, "/view")
					viewURL = setQueryParam(viewURL, "filename", filename)
					viewURL = setQueryParam(viewURL, "type", folderType)
					viewURL = setQueryParam(viewURL, "subfolder", subfolder)
					return viewURL
				}
			}
		}
	}
	return ""
}

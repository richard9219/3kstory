package services

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/richard9219/3kstory/internal/config"
)

type ModelProviderStatus struct {
	Name       string `json:"name"`
	Category   string `json:"category"`
	Configured bool   `json:"configured"`
	Healthy    bool   `json:"healthy"`
	Message    string `json:"message"`
	CheckedAt  string `json:"checked_at"`
}

type ModelProviderAlert struct {
	Name             string `json:"name"`
	Category         string `json:"category"`
	FailureStreak    int    `json:"failure_streak"`
	FailureThreshold int    `json:"failure_threshold"`
	Alerting         bool   `json:"alerting"`
	LastFailureAt    string `json:"last_failure_at,omitempty"`
	LastSuccessAt    string `json:"last_success_at,omitempty"`
	LastMessage      string `json:"last_message,omitempty"`
	UpdatedAt        string `json:"updated_at"`
}

type ModelProbeTaskStatus struct {
	Enabled          bool   `json:"enabled"`
	IntervalSeconds  int    `json:"interval_seconds"`
	FailureThreshold int    `json:"failure_threshold"`
	LastProbeAt      string `json:"last_probe_at,omitempty"`
	NextProbeAt      string `json:"next_probe_at,omitempty"`
	Running          bool   `json:"running"`
}

type ModelCenterOverview struct {
	PreferredVideoProvider string                `json:"preferred_video_provider"`
	VideoProviders         []ProviderHealth      `json:"video_providers"`
	TextProviders          []ModelProviderStatus `json:"text_providers"`
	ImageProviders         []ModelProviderStatus `json:"image_providers"`
	TTSProviders           []ModelProviderStatus `json:"tts_providers"`
	TaskRoutes             []TaskRouteStatus     `json:"task_routes"`
	Alerts                 []ModelProviderAlert  `json:"alerts"`
	ProbeTask              ModelProbeTaskStatus  `json:"probe_task"`
}

type TaskRouteStatus struct {
	Task      string   `json:"task"`
	Category  string   `json:"category"`
	Providers []string `json:"providers"`
}

type ModelCenterService struct {
	cfg          *config.Config
	videoService *VideoService

	mu                sync.RWMutex
	cachedOverview    *ModelCenterOverview
	alertState        map[string]ModelProviderAlert
	probeInterval     time.Duration
	failureThreshold  int
	lastProbeAt       time.Time
	probeRunning      bool
	backgroundProbeOn bool
}

func NewModelCenterService(cfg *config.Config, videoService *VideoService) *ModelCenterService {
	return &ModelCenterService{
		cfg:              cfg,
		videoService:     videoService,
		alertState:       make(map[string]ModelProviderAlert),
		probeInterval:    time.Duration(cfg.AI.ModelProbeInterval) * time.Second,
		failureThreshold: cfg.AI.ModelFailThreshold,
	}
}

func (s *ModelCenterService) StartBackgroundProbe(ctx context.Context) {
	s.mu.Lock()
	if s.backgroundProbeOn {
		s.mu.Unlock()
		return
	}
	s.backgroundProbeOn = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(s.probeInterval)
		defer ticker.Stop()

		s.runProbeCycle(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.runProbeCycle(ctx)
			}
		}
	}()
}

func (s *ModelCenterService) TriggerProbe(ctx context.Context) *ModelCenterOverview {
	s.runProbeCycle(ctx)
	return s.GetOverview(ctx)
}

func (s *ModelCenterService) GetOverview(ctx context.Context) *ModelCenterOverview {
	s.mu.RLock()
	hasCache := s.cachedOverview != nil
	s.mu.RUnlock()

	if !hasCache {
		s.runProbeCycle(ctx)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cachedOverview == nil {
		return &ModelCenterOverview{}
	}
	clone := *s.cachedOverview
	clone.VideoProviders = append([]ProviderHealth(nil), s.cachedOverview.VideoProviders...)
	clone.TextProviders = append([]ModelProviderStatus(nil), s.cachedOverview.TextProviders...)
	clone.ImageProviders = append([]ModelProviderStatus(nil), s.cachedOverview.ImageProviders...)
	clone.TTSProviders = append([]ModelProviderStatus(nil), s.cachedOverview.TTSProviders...)
	clone.TaskRoutes = append([]TaskRouteStatus(nil), s.cachedOverview.TaskRoutes...)
	clone.Alerts = append([]ModelProviderAlert(nil), s.cachedOverview.Alerts...)
	return &clone
}

func (s *ModelCenterService) runProbeCycle(ctx context.Context) {
	s.mu.Lock()
	if s.probeRunning {
		s.mu.Unlock()
		return
	}
	s.probeRunning = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.probeRunning = false
		s.mu.Unlock()
	}()

	now := time.Now()
	overview := s.collectOverview(ctx, now)
	alerts := s.updateAlerts(overview, now)

	overview.Alerts = alerts
	overview.ProbeTask = ModelProbeTaskStatus{
		Enabled:          true,
		IntervalSeconds:  int(s.probeInterval / time.Second),
		FailureThreshold: s.failureThreshold,
		LastProbeAt:      now.Format(time.RFC3339),
		NextProbeAt:      now.Add(s.probeInterval).Format(time.RFC3339),
		Running:          false,
	}

	s.mu.Lock()
	s.lastProbeAt = now
	s.cachedOverview = overview
	s.mu.Unlock()
}

func (s *ModelCenterService) collectOverview(ctx context.Context, now time.Time) *ModelCenterOverview {
	checkedAt := now.Format(time.RFC3339)
	videoStatuses := s.videoService.ProviderHealthStatuses(ctx)

	qwenConfigured := strings.TrimSpace(s.cfg.AI.QwenAPIKey) != ""
	vllmConfigured := strings.TrimSpace(s.cfg.AI.VLLMBaseURL) != ""
	ollamaConfigured := strings.TrimSpace(s.cfg.AI.OLLAMABaseURL) != ""
	imageConfigured := strings.TrimSpace(s.cfg.AI.ImageServiceURL) != ""
	minimaxConfigured := strings.TrimSpace(s.cfg.AI.MiniMaxAPIKey) != "" && strings.TrimSpace(s.cfg.AI.MiniMaxAPIBase) != ""
	seedanceConfigured := strings.TrimSpace(s.cfg.AI.SeedanceAPIKey) != "" && strings.TrimSpace(s.cfg.AI.SeedanceAPIBase) != ""
	comfyConfigured := strings.TrimSpace(s.cfg.AI.ComfyBaseURL) != ""

	vllmHealthy, vllmMessage := s.checkEndpoint(ctx, s.cfg.AI.VLLMBaseURL, "vLLM base URL missing")
	ollamaHealthy, ollamaMessage := s.checkEndpoint(ctx, s.cfg.AI.OLLAMABaseURL, "Ollama base URL missing")
	imageHealthy, imageMessage := s.checkEndpoint(ctx, s.cfg.AI.ImageServiceURL, "Image service URL missing")
	comfyHealthy, comfyMessage := s.checkEndpoint(ctx, s.cfg.AI.ComfyBaseURL, "Comfy base URL missing")

	if !qwenConfigured {
		vllmMessage = configMessage(vllmConfigured, vllmMessage, "vLLM base URL missing")
		ollamaMessage = configMessage(ollamaConfigured, ollamaMessage, "Ollama base URL missing")
		imageMessage = configMessage(imageConfigured, imageMessage, "Image service URL missing")
	}

	return &ModelCenterOverview{
		PreferredVideoProvider: string(s.videoService.PreferredProvider()),
		VideoProviders:         videoStatuses,
		TextProviders: []ModelProviderStatus{
			{
				Name:       "cloud_qwen",
				Category:   "text",
				Configured: qwenConfigured,
				Healthy:    qwenConfigured,
				Message:    configMessage(qwenConfigured, "Qwen API key configured", "Qwen API key missing"),
				CheckedAt:  checkedAt,
			},
			{
				Name:       "local_vllm",
				Category:   "text",
				Configured: vllmConfigured,
				Healthy:    vllmConfigured && vllmHealthy,
				Message:    configMessage(vllmConfigured, vllmMessage, "vLLM base URL missing"),
				CheckedAt:  checkedAt,
			},
			{
				Name:       "local_ollama",
				Category:   "text",
				Configured: ollamaConfigured,
				Healthy:    ollamaConfigured && ollamaHealthy,
				Message:    configMessage(ollamaConfigured, ollamaMessage, "Ollama base URL missing"),
				CheckedAt:  checkedAt,
			},
			{
				Name:       "minimax_video",
				Category:   "video",
				Configured: minimaxConfigured,
				Healthy:    minimaxConfigured,
				Message:    configMessage(minimaxConfigured, "MiniMax API configured", "MiniMax API missing"),
				CheckedAt:  checkedAt,
			},
			{
				Name:       "seedance_video",
				Category:   "video",
				Configured: seedanceConfigured,
				Healthy:    seedanceConfigured,
				Message:    configMessage(seedanceConfigured, "Seedance API configured", "Seedance API missing"),
				CheckedAt:  checkedAt,
			},
			{
				Name:       "comfy_workflow",
				Category:   "video",
				Configured: comfyConfigured,
				Healthy:    comfyConfigured && comfyHealthy,
				Message:    configMessage(comfyConfigured, comfyMessage, "Comfy base URL missing"),
				CheckedAt:  checkedAt,
			},
		},
		ImageProviders: []ModelProviderStatus{
			{
				Name:       "image_service",
				Category:   "image",
				Configured: imageConfigured,
				Healthy:    imageConfigured && imageHealthy,
				Message:    configMessage(imageConfigured, imageMessage, "Image service URL missing"),
				CheckedAt:  checkedAt,
			},
		},
		TTSProviders: []ModelProviderStatus{
			{
				Name:       "edge_tts",
				Category:   "tts",
				Configured: true,
				Healthy:    true,
				Message:    "Built-in TTS pipeline enabled",
				CheckedAt:  checkedAt,
			},
		},
		TaskRoutes: []TaskRouteStatus{
			{
				Task:      string(TextTaskLongformScript),
				Category:  "text",
				Providers: stringifyTextProviders(NewAIService(s.cfg).textProvidersForTask(TextTaskLongformScript)),
			},
			{
				Task:      string(TextTaskNarration),
				Category:  "text",
				Providers: stringifyTextProviders(NewAIService(s.cfg).textProvidersForTask(TextTaskNarration)),
			},
			{
				Task:      string(TextTaskStoryboard),
				Category:  "text",
				Providers: stringifyTextProviders(NewAIService(s.cfg).textProvidersForTask(TextTaskStoryboard)),
			},
			{
				Task:      string(TextTaskShotPrompt),
				Category:  "text",
				Providers: stringifyTextProviders(NewAIService(s.cfg).textProvidersForTask(TextTaskShotPrompt)),
			},
			{
				Task:      string(VideoTaskScene),
				Category:  "video",
				Providers: stringifyVideoProviders(s.videoService.providerOrderForTask(VideoTaskScene)),
			},
			{
				Task:      string(VideoTaskNarration),
				Category:  "video",
				Providers: stringifyVideoProviders(s.videoService.providerOrderForTask(VideoTaskNarration)),
			},
			{
				Task:      string(VideoTaskPreview),
				Category:  "video",
				Providers: stringifyVideoProviders(s.videoService.providerOrderForTask(VideoTaskPreview)),
			},
			{
				Task:      string(VideoTaskPremium),
				Category:  "video",
				Providers: stringifyVideoProviders(s.videoService.providerOrderForTask(VideoTaskPremium)),
			},
		},
	}
}

func stringifyTextProviders(items []TextProvider) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, string(item))
	}
	return out
}

func stringifyVideoProviders(items []VideoProvider) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, string(item))
	}
	return out
}

func (s *ModelCenterService) updateAlerts(overview *ModelCenterOverview, now time.Time) []ModelProviderAlert {
	s.mu.Lock()
	defer s.mu.Unlock()

	seen := make(map[string]struct{})
	alerts := make([]ModelProviderAlert, 0)

	for _, provider := range overview.VideoProviders {
		key := "video:" + string(provider.Provider)
		seen[key] = struct{}{}
		alerts = append(alerts, s.nextAlertState(key, string(provider.Provider), "video", provider.Healthy, provider.Message, now))
	}

	for _, provider := range overview.TextProviders {
		key := provider.Category + ":" + provider.Name
		seen[key] = struct{}{}
		alerts = append(alerts, s.nextAlertState(key, provider.Name, provider.Category, provider.Healthy, provider.Message, now))
	}

	for _, provider := range overview.ImageProviders {
		key := provider.Category + ":" + provider.Name
		seen[key] = struct{}{}
		alerts = append(alerts, s.nextAlertState(key, provider.Name, provider.Category, provider.Healthy, provider.Message, now))
	}

	for _, provider := range overview.TTSProviders {
		key := provider.Category + ":" + provider.Name
		seen[key] = struct{}{}
		alerts = append(alerts, s.nextAlertState(key, provider.Name, provider.Category, provider.Healthy, provider.Message, now))
	}

	for key := range s.alertState {
		if _, ok := seen[key]; !ok {
			delete(s.alertState, key)
		}
	}

	return alerts
}

func (s *ModelCenterService) nextAlertState(key, name, category string, healthy bool, message string, now time.Time) ModelProviderAlert {
	prev := s.alertState[key]
	next := prev
	next.Name = name
	next.Category = category
	next.FailureThreshold = s.failureThreshold
	next.LastMessage = message
	next.UpdatedAt = now.Format(time.RFC3339)

	if healthy {
		next.FailureStreak = 0
		next.Alerting = false
		next.LastSuccessAt = now.Format(time.RFC3339)
	} else {
		next.FailureStreak = prev.FailureStreak + 1
		next.Alerting = next.FailureStreak >= s.failureThreshold
		next.LastFailureAt = now.Format(time.RFC3339)
	}

	s.alertState[key] = next
	return next
}

func (s *ModelCenterService) checkEndpoint(ctx context.Context, rawURL, missingMsg string) (bool, string) {
	url := strings.TrimSpace(rawURL)
	if url == "" {
		return false, missingMsg
	}

	probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, url, nil)
	if err != nil {
		return false, "invalid endpoint URL"
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, "endpoint unreachable"
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		return false, "endpoint unhealthy"
	}

	return true, "endpoint reachable"
}

func configMessage(configured bool, okMessage, missingMessage string) string {
	if configured {
		return okMessage
	}
	return missingMessage
}

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/richard9219/3kstory/internal/config"
	"github.com/richard9219/3kstory/internal/models"
	"gorm.io/gorm"
)

// VideoProvider defines which video generation service to use
type VideoProvider string

const (
	ProviderRunway   VideoProvider = "runway"
	ProviderPika     VideoProvider = "pika"
	ProviderLocal    VideoProvider = "local"
	ProviderMiniMax  VideoProvider = "minimax"
	ProviderSeedance VideoProvider = "seedance"
	ProviderComfy    VideoProvider = "comfy"
)

// VideoService handles video generation via third-party APIs
type VideoService struct {
	cfg       *config.Config
	db        *gorm.DB
	providers map[VideoProvider]videoProviderClient
}

func NewVideoService(cfg *config.Config, db *gorm.DB) *VideoService {
	service := &VideoService{cfg: cfg, db: db}
	service.providers = map[VideoProvider]videoProviderClient{
		ProviderRunway:   newRunwayVideoProvider(cfg),
		ProviderPika:     newPikaVideoProvider(cfg),
		ProviderLocal:    newLocalVideoProvider(cfg),
		ProviderMiniMax:  newMiniMaxVideoProvider(cfg),
		ProviderSeedance: newSeedanceVideoProvider(cfg),
		ProviderComfy:    newComfyVideoProvider(cfg),
	}
	return service
}

func (s *VideoService) PreferredProvider() VideoProvider {
	return s.PreferredProviderForTask(VideoTaskGeneric)
}

func (s *VideoService) PreferredProviderForTask(task VideoTask) VideoProvider {
	configured := s.ConfiguredProvidersForTask(task)
	if len(configured) > 0 {
		return configured[0]
	}
	if task == VideoTaskNarration {
		return ProviderLocal
	}
	return ProviderRunway
}

func (s *VideoService) ConfiguredProvidersForTask(task VideoTask) []VideoProvider {
	order := s.providerOrderForTask(task)
	configured := make([]VideoProvider, 0, len(order))
	for _, provider := range order {
		client, ok := s.providers[provider]
		if !ok {
			continue
		}
		if err := client.ValidateConfig(); err == nil {
			configured = append(configured, provider)
		}
	}
	return configured
}

func (s *VideoService) providerOrderForTask(task VideoTask) []VideoProvider {
	var configured []string
	switch task {
	case VideoTaskScene:
		configured = parseCSVProviders(s.cfg.AI.SceneVideoProviders)
		return normalizeVideoProviders(mergeProviderLists(configured, string(ProviderComfy), string(ProviderLocal), string(ProviderMiniMax), string(ProviderPika), string(ProviderSeedance), string(ProviderRunway)))
	case VideoTaskNarration:
		configured = parseCSVProviders(s.cfg.AI.NarrationVideoProviders)
		return normalizeVideoProviders(mergeProviderLists(configured, string(ProviderLocal), string(ProviderComfy), string(ProviderMiniMax), string(ProviderPika), string(ProviderSeedance), string(ProviderRunway)))
	case VideoTaskPreview:
		configured = parseCSVProviders(s.cfg.AI.PreviewVideoProviders)
		return normalizeVideoProviders(mergeProviderLists(configured, string(ProviderComfy), string(ProviderLocal), string(ProviderMiniMax), string(ProviderPika), string(ProviderSeedance), string(ProviderRunway)))
	case VideoTaskPremium:
		configured = parseCSVProviders(s.cfg.AI.PremiumVideoProviders)
		return normalizeVideoProviders(mergeProviderLists(configured, string(ProviderSeedance), string(ProviderMiniMax), string(ProviderRunway), string(ProviderPika), string(ProviderComfy), string(ProviderLocal)))
	default:
		return normalizeVideoProviders(mergeProviderLists(nil, string(ProviderRunway), string(ProviderPika), string(ProviderMiniMax), string(ProviderSeedance), string(ProviderComfy), string(ProviderLocal)))
	}
}

func normalizeVideoProviders(items []string) []VideoProvider {
	out := make([]VideoProvider, 0, len(items))
	for _, item := range items {
		switch VideoProvider(item) {
		case ProviderRunway, ProviderPika, ProviderLocal, ProviderMiniMax, ProviderSeedance, ProviderComfy:
			out = append(out, VideoProvider(item))
		}
	}
	if len(out) == 0 {
		return []VideoProvider{ProviderRunway, ProviderPika, ProviderMiniMax, ProviderSeedance, ProviderComfy, ProviderLocal}
	}
	return out
}

func (s *VideoService) ConfiguredProviders() []VideoProvider {
	return s.ConfiguredProvidersForTask(VideoTaskGeneric)
}

func (s *VideoService) GenerateVideoForTask(ctx context.Context, task VideoTask, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	provider := req.Provider
	if provider == "" {
		provider = s.PreferredProviderForTask(task)
	}
	req.Provider = provider
	client, err := s.providerClient(provider)
	if err != nil {
		return nil, err
	}
	result, err := client.Generate(ctx, req)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *VideoService) FailoverGenerateForTask(ctx context.Context, task VideoTask, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	if req.Provider == "" {
		req.Provider = s.PreferredProviderForTask(task)
	}

	result, err := s.GenerateVideoForTask(ctx, task, req)
	if err == nil {
		return result, nil
	}

	fmt.Printf("Primary provider %s failed for task %s: %v, attempting fallback\n", req.Provider, task, err)

	fallbacks := s.ConfiguredProvidersForTask(task)
	if len(fallbacks) == 0 {
		return nil, err
	}

	var fallbackErr error
	for _, provider := range fallbacks {
		if provider == req.Provider {
			continue
		}
		fallbackReq := *req
		fallbackReq.Provider = provider
		result, fallbackErr = s.GenerateVideoForTask(ctx, task, &fallbackReq)
		if fallbackErr == nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("all configured providers failed for task %s. Primary: %w, Last fallback: %w", task, err, fallbackErr)
}

func (s *VideoService) PreferredProviderLegacy() VideoProvider {
	configured := s.ConfiguredProviders()
	if len(configured) > 0 {
		return configured[0]
	}
	return ProviderRunway
}

func (s *VideoService) ProviderHealthStatuses(ctx context.Context) []ProviderHealth {
	order := []VideoProvider{ProviderRunway, ProviderPika, ProviderMiniMax, ProviderSeedance, ProviderComfy, ProviderLocal}
	statuses := make([]ProviderHealth, 0, len(order))
	for _, provider := range order {
		client, ok := s.providers[provider]
		if !ok {
			continue
		}
		statuses = append(statuses, client.HealthCheck(ctx))
	}
	return statuses
}

func (s *VideoService) providerClient(provider VideoProvider) (videoProviderClient, error) {
	client, ok := s.providers[provider]
	if !ok {
		return nil, &VideoProviderError{Provider: provider, Kind: VideoErrorUnsupported, Message: "provider is not registered"}
	}
	return client, nil
}

// VideoGenerationRequest represents a video generation request
type VideoGenerationRequest struct {
	ProjectID          uint
	SceneID            uint
	Prompt             string
	Provider           VideoProvider
	Model              string
	ImageURL           string // for image-to-video
	LastFrameImageURL  string
	ReferenceImageURLs []string
	Duration           int    // seconds (1-60)
	AspectRatio        string // "16:9" or "9:16"
	Resolution         string
	Mode               string
	Seed               int
	WorkflowPath       string
	Workflow           models.JSONMap
	CallbackURL        string
	ExtraData          models.JSONMap
	SourceVideoPath    string
	SourceVideoURL     string
	NarrationSegments  []LocalNarrationSegment
}

type LocalNarrationSegment struct {
	Title             string `json:"title"`
	NarrationText     string `json:"narration_text"`
	EstimatedDuration int    `json:"estimated_duration"`
	AudioURL          string `json:"audio_url,omitempty"`
	AudioPath         string `json:"audio_path,omitempty"`
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
	return s.GenerateVideoForTask(ctx, VideoTaskGeneric, req)
}

// PollVideoStatus checks the status of a video generation job
func (s *VideoService) PollVideoStatus(ctx context.Context, videoID string, provider VideoProvider) (*VideoGenerationResult, error) {
	client, err := s.providerClient(provider)
	if err != nil {
		return nil, err
	}
	return client.PollStatus(ctx, videoID)
}

// FailoverGenerate attempts to generate video with primary provider, falls back to secondary
func (s *VideoService) FailoverGenerate(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	return s.FailoverGenerateForTask(ctx, VideoTaskGeneric, req)
}

// GenerateVideoTask represents an async video generation task
type GenerateVideoTask struct {
	ID          uint
	ProjectID   uint
	UserID      uint
	SceneID     uint
	TaskType    string
	Provider    string
	VideoID     string
	Status      string // pending, processing, completed, failed
	Title       string
	VideoURL    string
	InputData   models.JSONMap
	OutputData  models.JSONMap
	ErrorMsg    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

// SaveVideoTask persists a video generation task
func (s *VideoService) SaveVideoTask(ctx context.Context, task *GenerateVideoTask) error {
	if s.db == nil {
		return fmt.Errorf("database is not initialized")
	}

	var sceneID *uint
	if task.SceneID > 0 {
		s := task.SceneID
		sceneID = &s
	}

	m := &models.VideoTask{
		ID:          task.ID,
		UserID:      task.UserID,
		ProjectID:   task.ProjectID,
		SceneID:     sceneID,
		TaskType:    task.TaskType,
		Title:       task.Title,
		Provider:    task.Provider,
		VideoID:     task.VideoID,
		VideoURL:    task.VideoURL,
		Status:      task.Status,
		InputData:   task.InputData,
		OutputData:  task.OutputData,
		ErrorMsg:    task.ErrorMsg,
		CompletedAt: task.CompletedAt,
	}

	if m.TaskType == "" {
		m.TaskType = "generate_video"
	}

	if task.ID > 0 {
		return s.db.WithContext(ctx).Model(&models.VideoTask{}).Where("id = ?", task.ID).Updates(m).Error
	}
	if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	task.ID = m.ID
	task.CreatedAt = m.CreatedAt
	task.UpdatedAt = m.UpdatedAt
	return nil
}

// GetVideoTask retrieves a video generation task by ID
func (s *VideoService) GetVideoTask(ctx context.Context, taskID uint) (*GenerateVideoTask, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var m models.VideoTask
	if err := s.db.WithContext(ctx).First(&m, taskID).Error; err != nil {
		return nil, err
	}
	return mapVideoTaskModel(&m), nil
}

// ListVideoTasks retrieves all video tasks for a project
func (s *VideoService) ListVideoTasks(ctx context.Context, projectID uint) ([]*GenerateVideoTask, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var list []models.VideoTask
	err := s.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]*GenerateVideoTask, 0, len(list))
	for i := range list {
		out = append(out, mapVideoTaskModel(&list[i]))
	}
	return out, nil
}

func mapVideoTaskModel(m *models.VideoTask) *GenerateVideoTask {
	res := &GenerateVideoTask{
		ID:          m.ID,
		ProjectID:   m.ProjectID,
		UserID:      m.UserID,
		TaskType:    m.TaskType,
		Provider:    m.Provider,
		VideoID:     m.VideoID,
		Title:       m.Title,
		Status:      m.Status,
		VideoURL:    m.VideoURL,
		InputData:   m.InputData,
		OutputData:  m.OutputData,
		ErrorMsg:    m.ErrorMsg,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		CompletedAt: m.CompletedAt,
	}
	if m.SceneID != nil {
		res.SceneID = *m.SceneID
	}
	return res
}

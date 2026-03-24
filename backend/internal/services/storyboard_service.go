package services

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/richard9219/3kstory/internal/config"
	"github.com/richard9219/3kstory/internal/models"
	"gorm.io/gorm"
)

type StoryboardService struct {
	db           *gorm.DB
	videoService *VideoService
	cfg          *config.Config
	uploaders    *platformUploaderFactory
}

type StoryboardVersionNode struct {
	Root     models.StoryboardShot   `json:"root"`
	Versions []models.StoryboardShot `json:"versions"`
}

type StoryboardTimeline struct {
	ProjectID       uint                    `json:"project_id"`
	TotalDurationMs int                     `json:"total_duration_ms"`
	ReadyShotCount  int                     `json:"ready_shot_count"`
	TotalShotCount  int                     `json:"total_shot_count"`
	LatestExport    *models.VideoTask       `json:"latest_export,omitempty"`
	Shots           []models.StoryboardShot `json:"shots"`
}

type UpdateShotInput struct {
	TrackIndex           *int
	Chapter              *string
	Title                *string
	Description          *string
	CameraLanguage       *string
	EmotionTone          *string
	Duration             *int
	AspectRatio          *string
	Prompt               *string
	NegativePrompt       *string
	ReferenceImageURL    *string
	TimelineStartMs      *int
	TimelineDurationMs   *int
	TransitionType       *string
	TransitionDurationMs *int
	Locked               *bool
	Status               *string
}

type GenerateShotClipInput struct {
	Provider          VideoProvider
	Model             string
	Duration          int
	AspectRatio       string
	Prompt            string
	ReferenceImageURL string
	WorkflowPath      string
	NegativePrompt    string
}

type DirectorTemplateInput struct {
	Name                 string  `json:"name"`
	Slug                 string  `json:"slug"`
	SampleFrameURL       string  `json:"sample_frame_url"`
	SampleVideoURL       string  `json:"sample_video_url"`
	PromptPrefix         string  `json:"prompt_prefix"`
	CameraLanguage       string  `json:"camera_language"`
	EmotionTone          string  `json:"emotion_tone"`
	TransitionType       string  `json:"transition_type"`
	TransitionDurationMs int     `json:"transition_duration_ms"`
	GenreKeywords        string  `json:"genre_keywords"`
	WeightNarrative      float64 `json:"weight_narrative"`
	WeightVisual         float64 `json:"weight_visual"`
	WeightEmotion        float64 `json:"weight_emotion"`
	WeightRhythm         float64 `json:"weight_rhythm"`
	WeightContinuity     float64 `json:"weight_continuity"`
}

type AutoDirectorStrategyInput struct {
	Genre       string `json:"genre"`
	TemplateID  *uint  `json:"template_id"`
	Apply       bool   `json:"apply"`
	TunePercent int    `json:"tune_percent"`
}

type AutoDirectorStrategyResult struct {
	Genre          string                  `json:"genre"`
	Applied        bool                    `json:"applied"`
	TunePercent    int                     `json:"tune_percent"`
	Selected       models.DirectorTemplate `json:"selected"`
	PredictedScore float64                 `json:"predicted_score"`
	ShotUpdates    int                     `json:"shot_updates"`
}

type DirectorABCompareInput struct {
	TemplateAID   uint   `json:"template_a_id"`
	TemplateBID   uint   `json:"template_b_id"`
	Genre         string `json:"genre"`
	ApplyBest     bool   `json:"apply_best"`
	TunePercent   int    `json:"tune_percent"`
	RenderBestCut bool   `json:"render_best_cut"`
}

type DirectorABCompareResult struct {
	Genre             string                  `json:"genre"`
	TemplateA         models.DirectorTemplate `json:"template_a"`
	TemplateB         models.DirectorTemplate `json:"template_b"`
	ScoreA            float64                 `json:"score_a"`
	ScoreB            float64                 `json:"score_b"`
	WinnerTemplateID  uint                    `json:"winner_template_id"`
	WinnerTemplate    string                  `json:"winner_template"`
	Applied           bool                    `json:"applied"`
	RenderedExportID  string                  `json:"rendered_export_id"`
	PredictedGain     float64                 `json:"predicted_gain"`
	ComparedShotCount int                     `json:"compared_shot_count"`
}

func NewStoryboardService(db *gorm.DB, videoService *VideoService, cfg *config.Config) *StoryboardService {
	return &StoryboardService{db: db, videoService: videoService, cfg: cfg, uploaders: newPlatformUploaderFactory(cfg)}
}

func (s *StoryboardService) validateProjectOwnership(ctx context.Context, projectID uint, userID uint) error {
	var project models.Project
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return fmt.Errorf("project not found or no permission")
	}
	return nil
}

func (s *StoryboardService) ListProjectShots(ctx context.Context, projectID uint, userID uint) ([]models.StoryboardShot, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}
	var shots []models.StoryboardShot
	if err := s.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Order("sort_order ASC, shot_number ASC").
		Find(&shots).Error; err != nil {
		return nil, err
	}
	return shots, nil
}

func (s *StoryboardService) CreateShot(ctx context.Context, shot *models.StoryboardShot) error {
	if err := s.validateProjectOwnership(ctx, shot.ProjectID, shot.UserID); err != nil {
		return err
	}
	if shot.ShotNumber <= 0 {
		var count int64
		if err := s.db.WithContext(ctx).Model(&models.StoryboardShot{}).Where("project_id = ?", shot.ProjectID).Count(&count).Error; err == nil {
			shot.ShotNumber = int(count) + 1
			shot.SortOrder = int(count) + 1
		}
	}
	if shot.SortOrder <= 0 {
		shot.SortOrder = shot.ShotNumber
	}
	if shot.AspectRatio == "" {
		shot.AspectRatio = "16:9"
	}
	if shot.Status == "" {
		shot.Status = "draft"
	}
	if shot.ClipStatus == "" {
		shot.ClipStatus = "draft"
	}
	if shot.Duration <= 0 {
		shot.Duration = 5
	}
	if shot.TrackIndex <= 0 {
		shot.TrackIndex = 1
	}
	if shot.TimelineDurationMs <= 0 {
		shot.TimelineDurationMs = shot.Duration * 1000
	}
	if shot.TransitionType == "" {
		shot.TransitionType = "cut"
	}
	if shot.Version <= 0 {
		shot.Version = 1
	}
	return s.db.WithContext(ctx).Create(shot).Error
}

func (s *StoryboardService) BulkCreateShots(ctx context.Context, shots []models.StoryboardShot) (int, error) {
	if len(shots) == 0 {
		return 0, fmt.Errorf("shots required")
	}
	projectID := shots[0].ProjectID
	userID := shots[0].UserID
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return 0, err
	}
	for idx := range shots {
		if shots[idx].ProjectID != projectID || shots[idx].UserID != userID {
			return 0, fmt.Errorf("all shots must belong to the same project and user")
		}
		if err := s.CreateShot(ctx, &shots[idx]); err != nil {
			return idx, err
		}
	}
	return len(shots), nil
}

func (s *StoryboardService) ReorderShots(ctx context.Context, projectID uint, userID uint, orderedIDs []uint) error {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return err
	}
	if len(orderedIDs) == 0 {
		return fmt.Errorf("ordered shot ids required")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing []models.StoryboardShot
		if err := tx.Where("project_id = ? AND user_id = ?", projectID, userID).Order("sort_order ASC, shot_number ASC").Find(&existing).Error; err != nil {
			return err
		}
		if len(existing) != len(orderedIDs) {
			return fmt.Errorf("ordered ids do not match project shot count")
		}

		index := make(map[uint]int, len(orderedIDs))
		for i, id := range orderedIDs {
			index[id] = i + 1
		}

		for _, shot := range existing {
			order, ok := index[shot.ID]
			if !ok {
				return fmt.Errorf("ordered ids missing shot %d", shot.ID)
			}
			if err := tx.Model(&models.StoryboardShot{}).Where("id = ?", shot.ID).Updates(map[string]interface{}{
				"sort_order":  order,
				"shot_number": order,
			}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *StoryboardService) CreateShotVersion(ctx context.Context, projectID uint, userID uint, shotID uint, note string) (*models.StoryboardShot, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}

	var base models.StoryboardShot
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ? AND user_id = ?", shotID, projectID, userID).First(&base).Error; err != nil {
		return nil, fmt.Errorf("shot not found or no permission")
	}

	rootID := base.ID
	if base.RootShotID != nil {
		rootID = *base.RootShotID
	}

	var maxVersion int
	if err := s.db.WithContext(ctx).
		Model(&models.StoryboardShot{}).
		Where("project_id = ? AND user_id = ? AND (id = ? OR root_shot_id = ?)", projectID, userID, rootID, rootID).
		Select("COALESCE(MAX(version), 1)").
		Scan(&maxVersion).Error; err != nil {
		return nil, err
	}

	next := base
	next.ID = 0
	next.Version = maxVersion + 1
	next.ParentShotID = &base.ID
	next.RootShotID = &rootID
	next.VersionNote = note
	next.Status = "draft"

	if err := s.db.WithContext(ctx).Create(&next).Error; err != nil {
		return nil, err
	}
	return &next, nil
}

func (s *StoryboardService) ListVersionTree(ctx context.Context, projectID uint, userID uint) ([]StoryboardVersionNode, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}

	var shots []models.StoryboardShot
	if err := s.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Order("sort_order ASC, version ASC, created_at ASC").
		Find(&shots).Error; err != nil {
		return nil, err
	}

	nodeMap := make(map[uint]*StoryboardVersionNode)
	for _, shot := range shots {
		rootID := shot.ID
		if shot.RootShotID != nil {
			rootID = *shot.RootShotID
		}
		node, ok := nodeMap[rootID]
		if !ok {
			node = &StoryboardVersionNode{}
			nodeMap[rootID] = node
		}
		if shot.RootShotID == nil {
			node.Root = shot
			continue
		}
		node.Versions = append(node.Versions, shot)
	}

	roots := make([]uint, 0, len(nodeMap))
	for rootID := range nodeMap {
		roots = append(roots, rootID)
	}
	sort.Slice(roots, func(i, j int) bool {
		left := nodeMap[roots[i]].Root.SortOrder
		right := nodeMap[roots[j]].Root.SortOrder
		if left == right {
			return roots[i] < roots[j]
		}
		return left < right
	})

	out := make([]StoryboardVersionNode, 0, len(roots))
	for _, rootID := range roots {
		node := nodeMap[rootID]
		sort.Slice(node.Versions, func(i, j int) bool {
			return node.Versions[i].Version < node.Versions[j].Version
		})
		out = append(out, *node)
	}
	return out, nil
}

func (s *StoryboardService) BootstrapFromScenes(ctx context.Context, projectID uint, userID uint) (int, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return 0, err
	}

	var existing int64
	if err := s.db.WithContext(ctx).Model(&models.StoryboardShot{}).Where("project_id = ?", projectID).Count(&existing).Error; err != nil {
		return 0, err
	}
	if existing > 0 {
		return 0, nil
	}

	var scenes []models.Scene
	if err := s.db.WithContext(ctx).Where("project_id = ?", projectID).Order("scene_number ASC").Find(&scenes).Error; err != nil {
		return 0, err
	}

	if len(scenes) == 0 {
		return 0, nil
	}

	shots := make([]models.StoryboardShot, 0, len(scenes))
	for idx, scene := range scenes {
		sceneID := scene.ID
		shots = append(shots, models.StoryboardShot{
			UserID:             userID,
			ProjectID:          projectID,
			SceneID:            &sceneID,
			TrackIndex:         1,
			Chapter:            fmt.Sprintf("第%d章", idx+1),
			ShotNumber:         idx + 1,
			SortOrder:          idx + 1,
			Title:              scene.Title,
			Description:        scene.Description,
			CameraLanguage:     scene.ShotType,
			Duration:           scene.Duration,
			TimelineStartMs:    idx * scene.Duration * 1000,
			TimelineDurationMs: scene.Duration * 1000,
			TransitionType:     "cut",
			AspectRatio:        "16:9",
			Prompt:             scene.PromptForVideo,
			Status:             "draft",
			ClipStatus:         "draft",
			Version:            1,
		})
	}

	if err := s.db.WithContext(ctx).Create(&shots).Error; err != nil {
		return 0, err
	}

	return len(shots), nil
}

func (s *StoryboardService) GetTimeline(ctx context.Context, projectID uint, userID uint) (*StoryboardTimeline, error) {
	shots, err := s.ListProjectShots(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	currentStart := 0
	readyCount := 0
	for i := range shots {
		if shots[i].TimelineDurationMs <= 0 {
			shots[i].TimelineDurationMs = maxStoryboardInt(shots[i].Duration*1000, 1000)
		}
		if shots[i].TimelineStartMs <= 0 && i > 0 {
			shots[i].TimelineStartMs = currentStart
		}
		currentStart = shots[i].TimelineStartMs + shots[i].TimelineDurationMs - shots[i].TransitionDurationMs
		if shots[i].ClipStatus == "completed" && strings.TrimSpace(shots[i].ClipVideoURL) != "" {
			readyCount++
		}
	}

	var latestExport models.VideoTask
	var exportPtr *models.VideoTask
	if err := s.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ? AND task_type = ?", projectID, userID, "director_cut").
		Order("created_at DESC").
		First(&latestExport).Error; err == nil {
		exportPtr = &latestExport
	}

	return &StoryboardTimeline{
		ProjectID:       projectID,
		TotalDurationMs: maxStoryboardInt(currentStart, 0),
		ReadyShotCount:  readyCount,
		TotalShotCount:  len(shots),
		LatestExport:    exportPtr,
		Shots:           shots,
	}, nil
}

func (s *StoryboardService) UpdateShot(ctx context.Context, projectID uint, userID uint, shotID uint, in UpdateShotInput) (*models.StoryboardShot, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}
	var shot models.StoryboardShot
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ? AND user_id = ?", shotID, projectID, userID).First(&shot).Error; err != nil {
		return nil, fmt.Errorf("shot not found or no permission")
	}
	updates := map[string]interface{}{}
	if in.TrackIndex != nil {
		updates["track_index"] = *in.TrackIndex
	}
	if in.Chapter != nil {
		updates["chapter"] = *in.Chapter
	}
	if in.Title != nil {
		updates["title"] = *in.Title
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}
	if in.CameraLanguage != nil {
		updates["camera_language"] = *in.CameraLanguage
	}
	if in.EmotionTone != nil {
		updates["emotion_tone"] = *in.EmotionTone
	}
	if in.Duration != nil {
		updates["duration"] = *in.Duration
	}
	if in.AspectRatio != nil {
		updates["aspect_ratio"] = *in.AspectRatio
	}
	if in.Prompt != nil {
		updates["prompt"] = *in.Prompt
	}
	if in.NegativePrompt != nil {
		updates["negative_prompt"] = *in.NegativePrompt
	}
	if in.ReferenceImageURL != nil {
		updates["reference_image_url"] = *in.ReferenceImageURL
	}
	if in.TimelineStartMs != nil {
		updates["timeline_start_ms"] = *in.TimelineStartMs
	}
	if in.TimelineDurationMs != nil {
		updates["timeline_duration_ms"] = *in.TimelineDurationMs
	}
	if in.TransitionType != nil {
		updates["transition_type"] = *in.TransitionType
	}
	if in.TransitionDurationMs != nil {
		updates["transition_duration_ms"] = *in.TransitionDurationMs
	}
	if in.Locked != nil {
		updates["locked"] = *in.Locked
	}
	if in.Status != nil {
		updates["status"] = *in.Status
	}
	if len(updates) == 0 {
		return &shot, nil
	}
	if err := s.db.WithContext(ctx).Model(&shot).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).First(&shot, shot.ID).Error; err != nil {
		return nil, err
	}
	return &shot, nil
}

func (s *StoryboardService) GenerateShotClip(ctx context.Context, projectID uint, userID uint, shotID uint, in GenerateShotClipInput) (*models.StoryboardShot, error) {
	if s.videoService == nil {
		return nil, fmt.Errorf("video service not initialized")
	}
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}
	var shot models.StoryboardShot
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ? AND user_id = ?", shotID, projectID, userID).First(&shot).Error; err != nil {
		return nil, fmt.Errorf("shot not found or no permission")
	}
	if shot.Locked {
		return nil, fmt.Errorf("shot is locked")
	}

	durationSeconds := in.Duration
	if durationSeconds <= 0 {
		durationSeconds = maxStoryboardInt(shot.TimelineDurationMs/1000, shot.Duration)
	}
	aspectRatio := strings.TrimSpace(in.AspectRatio)
	if aspectRatio == "" {
		aspectRatio = shot.AspectRatio
	}
	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" {
		prompt = shot.Prompt
	}
	imageURL := strings.TrimSpace(in.ReferenceImageURL)
	if imageURL == "" {
		imageURL = shot.ReferenceImageURL
	}
	provider := in.Provider
	if provider == "" {
		provider = s.videoService.PreferredProviderForTask(VideoTaskScene)
	}

	if err := s.db.WithContext(ctx).Model(&shot).Updates(map[string]interface{}{
		"clip_status": "processing",
		"status":      "processing",
	}).Error; err != nil {
		return nil, err
	}

	result, err := s.videoService.FailoverGenerateForTask(ctx, VideoTaskScene, &VideoGenerationRequest{
		ProjectID:    projectID,
		SceneID:      derefUint(shot.SceneID),
		Prompt:       prompt,
		Provider:     provider,
		Model:        in.Model,
		ImageURL:     imageURL,
		Resolution:   "720p",
		WorkflowPath: in.WorkflowPath,
		Duration:     durationSeconds,
		AspectRatio:  aspectRatio,
		ExtraData: models.JSONMap{
			"storyboard_shot_id": shot.ID,
			"negative_prompt":    in.NegativePrompt,
		},
	})
	if err != nil {
		_ = s.db.WithContext(ctx).Model(&shot).Updates(map[string]interface{}{
			"clip_status": "failed",
			"status":      "failed",
			"clip_notes":  err.Error(),
		}).Error
		return nil, err
	}

	clipScore := 0.0
	if result.Status == "completed" {
		clipScore = 0.8
	}
	updates := map[string]interface{}{
		"prompt":               prompt,
		"reference_image_url":  imageURL,
		"aspect_ratio":         aspectRatio,
		"timeline_duration_ms": durationSeconds * 1000,
		"duration":             durationSeconds,
		"clip_provider":        string(result.Provider),
		"clip_video_id":        result.VideoID,
		"clip_video_url":       result.VideoURL,
		"clip_status":          result.Status,
		"clip_score":           clipScore,
		"clip_notes":           "generated from storyboard timeline",
		"status":               ternaryStoryboardStatus(result.Status == "completed", "completed", result.Status),
	}
	if err := s.db.WithContext(ctx).Model(&shot).Updates(updates).Error; err != nil {
		return nil, err
	}

	task := &GenerateVideoTask{
		UserID:    userID,
		ProjectID: projectID,
		SceneID:   derefUint(shot.SceneID),
		TaskType:  "storyboard_shot_video",
		Title:     shot.Title,
		Provider:  string(result.Provider),
		VideoID:   result.VideoID,
		Status:    result.Status,
		VideoURL:  result.VideoURL,
		Score:     clipScore,
		InputData: models.JSONMap{
			"storyboard_shot_id": shot.ID,
			"prompt":             prompt,
			"duration":           durationSeconds,
		},
		OutputData: models.JSONMap{
			"clip_video_url": result.VideoURL,
		},
	}
	_ = s.videoService.SaveVideoTask(ctx, task)

	if err := s.db.WithContext(ctx).First(&shot, shot.ID).Error; err != nil {
		return nil, err
	}
	return &shot, nil
}

func (s *StoryboardService) RegenerateShotClip(ctx context.Context, projectID uint, userID uint, shotID uint, in GenerateShotClipInput, note string) (*models.StoryboardShot, error) {
	version, err := s.CreateShotVersion(ctx, projectID, userID, shotID, note)
	if err != nil {
		return nil, err
	}
	version.Prompt = firstNonEmptyString(strings.TrimSpace(in.Prompt), version.Prompt)
	version.ReferenceImageURL = firstNonEmptyString(strings.TrimSpace(in.ReferenceImageURL), version.ReferenceImageURL)
	version.ClipStatus = "draft"
	if err := s.db.WithContext(ctx).Save(version).Error; err != nil {
		return nil, err
	}
	return s.GenerateShotClip(ctx, projectID, userID, version.ID, in)
}

func (s *StoryboardService) RenderDirectorCut(ctx context.Context, projectID uint, userID uint) (*models.VideoTask, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}
	shots, err := s.ListProjectShots(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	selected := make([]models.StoryboardShot, 0, len(shots))
	totalDurationMs := 0
	weightedScoreSum := 0.0
	weightedDurationSum := 0.0
	for _, shot := range shots {
		if shot.ClipStatus == "completed" && strings.TrimSpace(shot.ClipVideoURL) != "" {
			selected = append(selected, shot)
			durationMs := maxStoryboardInt(shot.TimelineDurationMs, shot.Duration*1000)
			totalDurationMs += durationMs
			durationSec := float64(durationMs) / 1000
			weightedScoreSum += shot.ClipScore * durationSec
			weightedDurationSum += durationSec
		}
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no completed storyboard clips available")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not found")
	}

	exportID := fmt.Sprintf("director_%d", time.Now().UnixNano())
	outputDir := filepath.Join(".local", "exports")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, err
	}
	workDir, err := os.MkdirTemp("", "3kstory-director-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(workDir)

	localFiles := make([]string, 0, len(selected))
	durationSeconds := make([]float64, 0, len(selected))
	for i, shot := range selected {
		clipPath, dlErr := s.materializeClip(ctx, workDir, i, shot.ClipVideoURL)
		if dlErr != nil {
			return nil, dlErr
		}
		normalizedPath, normalizeErr := s.normalizeClipForTimeline(ctx, workDir, i, clipPath, maxStoryboardInt(shot.TimelineDurationMs, shot.Duration*1000))
		if normalizeErr != nil {
			return nil, normalizeErr
		}
		localFiles = append(localFiles, normalizedPath)
		durationSeconds = append(durationSeconds, math.Max(float64(maxStoryboardInt(shot.TimelineDurationMs, shot.Duration*1000))/1000, 0.2))
	}

	transitionDurations := make([]float64, 0, len(selected))
	transitionKinds := make([]string, 0, len(selected))
	for i := range selected {
		if i == 0 {
			transitionDurations = append(transitionDurations, 0)
			transitionKinds = append(transitionKinds, "cut")
			continue
		}
		prev := selected[i-1]
		d := float64(maxStoryboardInt(prev.TransitionDurationMs, 0)) / 1000
		if prev.TransitionType == "cut" || d <= 0 {
			d = 0.02
		}
		transitionDurations = append(transitionDurations, math.Min(d, math.Max(0.02, durationSeconds[i]-0.05)))
		transitionKinds = append(transitionKinds, prev.TransitionType)
	}

	outputPath := filepath.Join(outputDir, exportID+".mp4")
	if len(localFiles) == 1 {
		cmd := exec.CommandContext(ctx, "ffmpeg",
			"-y",
			"-i", localFiles[0],
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-pix_fmt", "yuv420p",
			"-c:a", "aac",
			"-b:a", "192k",
			outputPath,
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("render director cut failed: %v: %s", err, strings.TrimSpace(string(out)))
		}
	} else {
		args := []string{"-y"}
		for _, file := range localFiles {
			args = append(args, "-i", file)
		}

		videoLabel := "[v0]"
		audioLabel := "[a0]"
		offset := durationSeconds[0]
		filters := []string{
			"[0:v]settb=AVTB,format=yuv420p[v0]",
			"[0:a]aformat=sample_rates=44100:channel_layouts=stereo[a0]",
		}
		for i := 1; i < len(localFiles); i++ {
			filters = append(filters,
				fmt.Sprintf("[%d:v]settb=AVTB,format=yuv420p[v%d]", i, i),
				fmt.Sprintf("[%d:a]aformat=sample_rates=44100:channel_layouts=stereo[a%d]", i, i),
			)

			transitionName := mapTransitionType(transitionKinds[i])
			td := math.Max(0.02, transitionDurations[i])
			offset = math.Max(offset-td, 0)
			nextVideoLabel := fmt.Sprintf("[vxf%d]", i)
			nextAudioLabel := fmt.Sprintf("[axf%d]", i)
			filters = append(filters,
				fmt.Sprintf("%s[v%d]xfade=transition=%s:duration=%.3f:offset=%.3f%s", videoLabel, i, transitionName, td, offset, nextVideoLabel),
				fmt.Sprintf("%s[a%d]acrossfade=d=%.3f:c1=tri:c2=tri%s", audioLabel, i, td, nextAudioLabel),
			)

			videoLabel = nextVideoLabel
			audioLabel = nextAudioLabel
			offset += durationSeconds[i]
		}

		filterComplex := strings.Join(filters, ";")
		args = append(args,
			"-filter_complex", filterComplex,
			"-map", videoLabel,
			"-map", audioLabel,
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-pix_fmt", "yuv420p",
			"-c:a", "aac",
			"-b:a", "192k",
			outputPath,
		)

		cmd := exec.CommandContext(ctx, "ffmpeg", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("render director cut failed: %v: %s", err, strings.TrimSpace(string(out)))
		}
	}

	qualityScore := 0.0
	if weightedDurationSum > 0 {
		qualityScore = roundStoryboardScore(weightedScoreSum / weightedDurationSum)
	}
	qualityGatePassed, gateReason, threshold := s.evaluateDirectorGate(selected, qualityScore)

	videoURL := fmt.Sprintf("/api/v1/projects/%d/storyboard-timeline/exports/%s", projectID, exportID)
	task := &models.VideoTask{
		UserID:    userID,
		ProjectID: projectID,
		TaskType:  "director_cut",
		Title:     "Director Cut",
		Provider:  "local",
		VideoID:   exportID,
		VideoURL:  videoURL,
		Status:    "completed",
		Score:     qualityScore,
		OutputData: models.JSONMap{
			"output_path":          outputPath,
			"shot_count":           len(selected),
			"timeline_duration_ms": totalDurationMs,
			"quality_score":        qualityScore,
			"quality_threshold":    threshold,
			"quality_gate_passed":  qualityGatePassed,
			"quality_gate_reason":  gateReason,
			"render_mode":          "transition_graph",
		},
		CompletedAt: ptrTime(time.Now()),
	}
	if err := s.db.WithContext(ctx).Create(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

func (s *StoryboardService) ResolveDirectorCutPath(ctx context.Context, projectID uint, userID uint, exportID string) (string, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return "", err
	}
	var task models.VideoTask
	if err := s.db.WithContext(ctx).Where("project_id = ? AND user_id = ? AND task_type = ? AND video_id = ?", projectID, userID, "director_cut", exportID).First(&task).Error; err != nil {
		return "", fmt.Errorf("director cut not found")
	}
	path, _ := task.OutputData["output_path"].(string)
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("director cut path missing")
	}
	return path, nil
}

func (s *StoryboardService) CheckDirectorCutGate(ctx context.Context, projectID uint, userID uint, exportID string) (bool, string, float64, float64, error) {
	task, err := s.getDirectorCutTask(ctx, projectID, userID, exportID)
	if err != nil {
		return false, "", 0, 0, err
	}
	score := task.Score
	if raw, ok := task.OutputData["quality_score"]; ok {
		score = asStoryboardFloat(raw, score)
	}
	threshold := s.directorQualityThreshold()
	if raw, ok := task.OutputData["quality_threshold"]; ok {
		threshold = asStoryboardFloat(raw, threshold)
	}
	if score >= threshold {
		return true, "", score, threshold, nil
	}
	return false, fmt.Sprintf("score %.3f below threshold %.2f", score, threshold), score, threshold, nil
}

func (s *StoryboardService) PublishDirectorCut(ctx context.Context, projectID uint, userID uint, exportID string, platform string) (*models.VideoTask, error) {
	return s.publishDirectorCutInternal(ctx, projectID, userID, exportID, platform, nil)
}

func (s *StoryboardService) RetryDirectorPublish(ctx context.Context, projectID uint, userID uint, recordID uint) (*models.VideoTask, *models.DirectorPublishRecord, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, nil, err
	}
	var record models.DirectorPublishRecord
	if err := s.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND user_id = ?", recordID, projectID, userID).
		First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, fmt.Errorf("publish record not found")
		}
		return nil, nil, err
	}
	if strings.TrimSpace(record.ExportID) == "" {
		return nil, nil, fmt.Errorf("publish record missing export id")
	}
	retryFromID := record.ID
	task, err := s.publishDirectorCutInternal(ctx, projectID, userID, record.ExportID, record.Platform, &retryFromID)
	if err != nil {
		return nil, nil, err
	}
	var newest models.DirectorPublishRecord
	if err := s.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ? AND export_id = ? AND platform = ?", projectID, userID, record.ExportID, record.Platform).
		Order("id DESC").
		First(&newest).Error; err != nil {
		return task, nil, err
	}
	return task, &newest, nil
}

func (s *StoryboardService) ListDirectorPublishHistory(ctx context.Context, projectID uint, userID uint, exportID string) ([]models.DirectorPublishRecord, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}
	query := s.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Order("created_at DESC")
	if strings.TrimSpace(exportID) != "" {
		query = query.Where("export_id = ?", strings.TrimSpace(exportID))
	}
	var records []models.DirectorPublishRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (s *StoryboardService) ListDirectorTemplates(ctx context.Context, projectID uint, userID uint) ([]models.DirectorTemplate, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}
	if err := s.ensureBuiltinDirectorTemplates(ctx, projectID, userID); err != nil {
		return nil, err
	}
	var templates []models.DirectorTemplate
	if err := s.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Order("is_builtin DESC, id ASC").
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (s *StoryboardService) CreateDirectorTemplate(ctx context.Context, projectID uint, userID uint, in DirectorTemplateInput) (*models.DirectorTemplate, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}
	template := models.DirectorTemplate{
		UserID:               userID,
		ProjectID:            projectID,
		Name:                 firstNonEmptyString(strings.TrimSpace(in.Name), "自定义导演模板"),
		Slug:                 firstNonEmptyString(strings.TrimSpace(in.Slug), strings.ToLower(strings.ReplaceAll(strings.TrimSpace(in.Name), " ", "_"))),
		SampleFrameURL:       strings.TrimSpace(in.SampleFrameURL),
		SampleVideoURL:       strings.TrimSpace(in.SampleVideoURL),
		PromptPrefix:         strings.TrimSpace(in.PromptPrefix),
		CameraLanguage:       strings.TrimSpace(in.CameraLanguage),
		EmotionTone:          strings.TrimSpace(in.EmotionTone),
		TransitionType:       normalizeTransitionType(in.TransitionType),
		TransitionDurationMs: maxStoryboardInt(in.TransitionDurationMs, 0),
		GenreKeywords:        strings.TrimSpace(in.GenreKeywords),
		WeightNarrative:      clamp01(in.WeightNarrative),
		WeightVisual:         clamp01(in.WeightVisual),
		WeightEmotion:        clamp01(in.WeightEmotion),
		WeightRhythm:         clamp01(in.WeightRhythm),
		WeightContinuity:     clamp01(in.WeightContinuity),
		IsBuiltin:            false,
	}
	normalizeTemplateWeights(&template)
	if err := s.db.WithContext(ctx).Create(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (s *StoryboardService) UpdateDirectorTemplate(ctx context.Context, projectID uint, userID uint, templateID uint, in DirectorTemplateInput) (*models.DirectorTemplate, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}
	var template models.DirectorTemplate
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ? AND user_id = ?", templateID, projectID, userID).First(&template).Error; err != nil {
		return nil, fmt.Errorf("template not found")
	}
	template.Name = firstNonEmptyString(strings.TrimSpace(in.Name), template.Name)
	if strings.TrimSpace(in.Slug) != "" {
		template.Slug = strings.TrimSpace(in.Slug)
	}
	if strings.TrimSpace(in.SampleFrameURL) != "" {
		template.SampleFrameURL = strings.TrimSpace(in.SampleFrameURL)
	}
	if strings.TrimSpace(in.SampleVideoURL) != "" {
		template.SampleVideoURL = strings.TrimSpace(in.SampleVideoURL)
	}
	if strings.TrimSpace(in.PromptPrefix) != "" {
		template.PromptPrefix = strings.TrimSpace(in.PromptPrefix)
	}
	if strings.TrimSpace(in.CameraLanguage) != "" {
		template.CameraLanguage = strings.TrimSpace(in.CameraLanguage)
	}
	if strings.TrimSpace(in.EmotionTone) != "" {
		template.EmotionTone = strings.TrimSpace(in.EmotionTone)
	}
	if strings.TrimSpace(in.TransitionType) != "" {
		template.TransitionType = normalizeTransitionType(in.TransitionType)
	}
	if in.TransitionDurationMs >= 0 {
		template.TransitionDurationMs = in.TransitionDurationMs
	}
	if strings.TrimSpace(in.GenreKeywords) != "" {
		template.GenreKeywords = strings.TrimSpace(in.GenreKeywords)
	}
	if in.WeightNarrative > 0 || in.WeightVisual > 0 || in.WeightEmotion > 0 || in.WeightRhythm > 0 || in.WeightContinuity > 0 {
		template.WeightNarrative = clamp01(in.WeightNarrative)
		template.WeightVisual = clamp01(in.WeightVisual)
		template.WeightEmotion = clamp01(in.WeightEmotion)
		template.WeightRhythm = clamp01(in.WeightRhythm)
		template.WeightContinuity = clamp01(in.WeightContinuity)
		normalizeTemplateWeights(&template)
	}
	if err := s.db.WithContext(ctx).Save(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (s *StoryboardService) DeleteDirectorTemplate(ctx context.Context, projectID uint, userID uint, templateID uint) error {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return err
	}
	var template models.DirectorTemplate
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ? AND user_id = ?", templateID, projectID, userID).First(&template).Error; err != nil {
		return fmt.Errorf("template not found")
	}
	if template.IsBuiltin {
		return fmt.Errorf("builtin template cannot be deleted")
	}
	return s.db.WithContext(ctx).Delete(&models.DirectorTemplate{}, template.ID).Error
}

func (s *StoryboardService) AutoDirectorStrategy(ctx context.Context, projectID uint, userID uint, in AutoDirectorStrategyInput) (*AutoDirectorStrategyResult, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}
	shots, err := s.ListProjectShots(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if len(shots) == 0 {
		return nil, fmt.Errorf("no shots found in project")
	}
	templates, err := s.ListDirectorTemplates(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if len(templates) == 0 {
		return nil, fmt.Errorf("no director templates available")
	}
	genre := strings.TrimSpace(strings.ToLower(in.Genre))
	tune := clampInt(in.TunePercent, 10, 100)
	selected := templates[0]
	if in.TemplateID != nil {
		for _, item := range templates {
			if item.ID == *in.TemplateID {
				selected = item
				break
			}
		}
	} else {
		projectTitle, projectPrompt := s.readProjectTitleAndPrompt(ctx, projectID, userID)
		selected = s.pickTemplateByGenre(templates, genre, projectTitle, projectPrompt)
	}
	predicted := s.predictTemplateScore(selected, shots, genre)
	updated := 0
	if in.Apply {
		updated, err = s.applyTemplateToShots(ctx, projectID, userID, selected, shots, genre, tune)
		if err != nil {
			return nil, err
		}
	}
	return &AutoDirectorStrategyResult{
		Genre:          genre,
		Applied:        in.Apply,
		TunePercent:    tune,
		Selected:       selected,
		PredictedScore: predicted,
		ShotUpdates:    updated,
	}, nil
}

func (s *StoryboardService) CompareDirectorAB(ctx context.Context, projectID uint, userID uint, in DirectorABCompareInput) (*DirectorABCompareResult, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}
	if in.TemplateAID == 0 || in.TemplateBID == 0 {
		return nil, fmt.Errorf("template_a_id and template_b_id are required")
	}
	templates, err := s.ListDirectorTemplates(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	var tA *models.DirectorTemplate
	var tB *models.DirectorTemplate
	for i := range templates {
		if templates[i].ID == in.TemplateAID {
			tA = &templates[i]
		}
		if templates[i].ID == in.TemplateBID {
			tB = &templates[i]
		}
	}
	if tA == nil || tB == nil {
		return nil, fmt.Errorf("template not found")
	}
	shots, err := s.ListProjectShots(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if len(shots) == 0 {
		return nil, fmt.Errorf("no shots found in project")
	}
	genre := strings.TrimSpace(strings.ToLower(in.Genre))
	scoreA := s.predictTemplateScore(*tA, shots, genre)
	scoreB := s.predictTemplateScore(*tB, shots, genre)
	winner := tA
	if scoreB > scoreA {
		winner = tB
	}
	applied := false
	renderedExportID := ""
	if in.ApplyBest {
		tune := clampInt(in.TunePercent, 10, 100)
		if _, err := s.applyTemplateToShots(ctx, projectID, userID, *winner, shots, genre, tune); err != nil {
			return nil, err
		}
		applied = true
		if in.RenderBestCut {
			if task, renderErr := s.RenderDirectorCut(ctx, projectID, userID); renderErr == nil {
				renderedExportID = task.VideoID
			}
		}
	}
	return &DirectorABCompareResult{
		Genre:             genre,
		TemplateA:         *tA,
		TemplateB:         *tB,
		ScoreA:            scoreA,
		ScoreB:            scoreB,
		WinnerTemplateID:  winner.ID,
		WinnerTemplate:    winner.Name,
		Applied:           applied,
		RenderedExportID:  renderedExportID,
		PredictedGain:     roundStoryboardScore(math.Abs(scoreA - scoreB)),
		ComparedShotCount: len(shots),
	}, nil
}

func (s *StoryboardService) ensureBuiltinDirectorTemplates(ctx context.Context, projectID uint, userID uint) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.DirectorTemplate{}).Where("project_id = ? AND user_id = ?", projectID, userID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	builtins := []models.DirectorTemplate{
		{UserID: userID, ProjectID: projectID, Name: "黑泽明", Slug: "kurosawa", PromptPrefix: "high-contrast monochrome mood, dynamic weather, samurai-like blocking", CameraLanguage: "长焦压缩 + 低机位横移 + 风雨动势", EmotionTone: "肃杀、宿命、张力", TransitionType: "wipe", TransitionDurationMs: 260, GenreKeywords: "历史,战争,宿命", WeightNarrative: 0.18, WeightVisual: 0.28, WeightEmotion: 0.2, WeightRhythm: 0.22, WeightContinuity: 0.12, IsBuiltin: true},
		{UserID: userID, ProjectID: projectID, Name: "张艺谋", Slug: "zhangyimou", PromptPrefix: "bold color choreography, ceremonial composition, poetic epic tableau", CameraLanguage: "色块构图 + 仪式化调度 + 大景别对称", EmotionTone: "浓烈、仪式、情绪外放", TransitionType: "fade", TransitionDurationMs: 320, GenreKeywords: "史诗,情感,古装", WeightNarrative: 0.16, WeightVisual: 0.34, WeightEmotion: 0.24, WeightRhythm: 0.16, WeightContinuity: 0.1, IsBuiltin: true},
		{UserID: userID, ProjectID: projectID, Name: "陈凯歌", Slug: "chenkaige", PromptPrefix: "operatic visual language, historical reflection, lyrical dramatic pacing", CameraLanguage: "戏剧性推轨 + 人物群像层次", EmotionTone: "抒情、史诗、反思", TransitionType: "match", TransitionDurationMs: 280, GenreKeywords: "历史,抒情,人物", WeightNarrative: 0.26, WeightVisual: 0.22, WeightEmotion: 0.24, WeightRhythm: 0.14, WeightContinuity: 0.14, IsBuiltin: true},
		{UserID: userID, ProjectID: projectID, Name: "冯小刚", Slug: "fengxiaogang", PromptPrefix: "urban realism, brisk rhythm, satirical social texture, handheld immediacy", CameraLanguage: "生活流手持 + 快节奏切换", EmotionTone: "写实、讽刺、群体烟火气", TransitionType: "cut", TransitionDurationMs: 120, GenreKeywords: "都市,现实,喜剧", WeightNarrative: 0.2, WeightVisual: 0.16, WeightEmotion: 0.2, WeightRhythm: 0.3, WeightContinuity: 0.14, IsBuiltin: true},
	}
	for i := range builtins {
		normalizeTemplateWeights(&builtins[i])
	}
	return s.db.WithContext(ctx).Create(&builtins).Error
}

func (s *StoryboardService) readProjectTitleAndPrompt(ctx context.Context, projectID uint, userID uint) (string, string) {
	var project models.Project
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return "", ""
	}
	return strings.TrimSpace(project.Title), strings.TrimSpace(project.Prompt)
}

func (s *StoryboardService) pickTemplateByGenre(templates []models.DirectorTemplate, genre string, projectTitle string, projectPrompt string) models.DirectorTemplate {
	best := templates[0]
	bestScore := -1.0
	corpus := strings.ToLower(strings.TrimSpace(genre + " " + projectTitle + " " + projectPrompt))
	for _, item := range templates {
		score := 0.4
		for _, token := range strings.Split(strings.ToLower(item.GenreKeywords), ",") {
			token = strings.TrimSpace(token)
			if token != "" && strings.Contains(corpus, token) {
				score += 0.18
			}
		}
		score += item.WeightVisual*0.1 + item.WeightEmotion*0.1 + item.WeightRhythm*0.1
		if score > bestScore {
			best = item
			bestScore = score
		}
	}
	return best
}

func (s *StoryboardService) predictTemplateScore(template models.DirectorTemplate, shots []models.StoryboardShot, genre string) float64 {
	if len(shots) == 0 {
		return 0
	}
	clipAvg := 0.55
	withScore := 0
	for _, shot := range shots {
		if shot.ClipScore > 0 {
			clipAvg += shot.ClipScore
			withScore++
		}
	}
	if withScore > 0 {
		clipAvg = clipAvg / float64(withScore+1)
	}
	genreFit := 0.5
	for _, token := range strings.Split(strings.ToLower(template.GenreKeywords), ",") {
		token = strings.TrimSpace(token)
		if token != "" && strings.Contains(strings.ToLower(genre), token) {
			genreFit += 0.1
		}
	}
	styleStrength := template.WeightNarrative*0.2 + template.WeightVisual*0.28 + template.WeightEmotion*0.22 + template.WeightRhythm*0.18 + template.WeightContinuity*0.12
	return roundStoryboardScore(clamp01(clipAvg*0.55 + genreFit*0.2 + styleStrength*0.25))
}

func (s *StoryboardService) applyTemplateToShots(ctx context.Context, projectID uint, userID uint, template models.DirectorTemplate, shots []models.StoryboardShot, genre string, tunePercent int) (int, error) {
	updated := 0
	tune := float64(clampInt(tunePercent, 10, 100)) / 100
	for _, shot := range shots {
		if shot.Locked {
			continue
		}
		prefix := strings.TrimSpace(template.PromptPrefix)
		nextPrompt := strings.TrimSpace(shot.Prompt)
		if prefix != "" {
			nextPrompt = strings.TrimSpace(prefix + ". " + nextPrompt)
		}
		if genre != "" {
			nextPrompt = strings.TrimSpace(nextPrompt + fmt.Sprintf(". genre focus: %s", genre))
		}
		nextTransitionDuration := int(float64(template.TransitionDurationMs) * (0.6 + 0.4*tune))
		updates := map[string]interface{}{
			"camera_language":        firstNonEmptyString(template.CameraLanguage, shot.CameraLanguage),
			"emotion_tone":           firstNonEmptyString(template.EmotionTone, shot.EmotionTone),
			"transition_type":        normalizeTransitionType(template.TransitionType),
			"transition_duration_ms": maxStoryboardInt(nextTransitionDuration, 0),
			"prompt":                 nextPrompt,
		}
		if err := s.db.WithContext(ctx).Model(&models.StoryboardShot{}).Where("id = ? AND project_id = ? AND user_id = ?", shot.ID, projectID, userID).Updates(updates).Error; err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}

func (s *StoryboardService) publishDirectorCutInternal(ctx context.Context, projectID uint, userID uint, exportID string, platform string, retriedFromID *uint) (*models.VideoTask, error) {
	task, err := s.getDirectorCutTask(ctx, projectID, userID, exportID)
	if err != nil {
		return nil, err
	}
	platform = strings.TrimSpace(strings.ToLower(platform))
	if platform == "" {
		return nil, fmt.Errorf("platform is required")
	}
	if platform != models.PlatformDouyin && platform != models.PlatformBilibili && platform != models.PlatformWeibo && platform != models.PlatformXiaohongshu {
		return nil, fmt.Errorf("unsupported platform")
	}

	passed, reason, score, threshold, err := s.CheckDirectorCutGate(ctx, projectID, userID, exportID)
	if err != nil {
		return nil, err
	}
	if !passed {
		return nil, fmt.Errorf(reason)
	}

	var account models.PlatformAccount
	if err := s.db.WithContext(ctx).Where("user_id = ? AND platform = ?", userID, platform).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("platform account not bound: %s", platform)
		}
		return nil, err
	}

	uploader, err := s.uploaders.Uploader(platform)
	if err != nil {
		return nil, err
	}

	videoPath, resolveErr := s.ResolveDirectorCutPath(ctx, projectID, userID, exportID)
	if resolveErr != nil {
		return nil, resolveErr
	}

	attemptNo := 1
	if retriedFromID != nil {
		var maxAttempt int
		_ = s.db.WithContext(ctx).
			Model(&models.DirectorPublishRecord{}).
			Where("project_id = ? AND user_id = ? AND export_id = ? AND platform = ?", projectID, userID, exportID, platform).
			Select("COALESCE(MAX(attempt_no), 0)").
			Scan(&maxAttempt).Error
		attemptNo = maxStoryboardInt(maxAttempt+1, 2)
	}

	record := &models.DirectorPublishRecord{
		UserID:        userID,
		ProjectID:     projectID,
		VideoTaskID:   task.ID,
		ExportID:      exportID,
		Platform:      platform,
		Status:        "pending",
		AttemptNo:     attemptNo,
		RetriedFromID: retriedFromID,
		RequestPayload: models.JSONMap{
			"title":      task.Title,
			"desc":       "Director Cut 自动发布",
			"video_path": videoPath,
		},
	}
	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}

	receipt, uploadErr := uploader.Upload(ctx, &account, PublishUploadRequest{
		Title:     firstNonEmptyString(strings.TrimSpace(task.Title), "Director Cut"),
		Desc:      "Director Cut 自动发布",
		VideoPath: videoPath,
	})

	responsePayload := models.JSONMap{}
	if receipt != nil {
		responsePayload = models.JSONMap{
			"platform":        receipt.Platform,
			"status":          receipt.Status,
			"receipt_id":      receipt.ReceiptID,
			"remote_video_id": receipt.RemoteVideoID,
			"remote_url":      receipt.RemoteURL,
			"http_status":     receipt.HTTPStatus,
			"request_id":      receipt.RequestID,
			"raw_body":        receipt.RawBody,
			"received_at":     receipt.ReceivedAt.UTC().Format(time.RFC3339),
		}
	}

	now := time.Now()
	recordUpdate := map[string]interface{}{
		"response_payload": responsePayload,
		"updated_at":       now,
	}
	if uploadErr != nil {
		recordUpdate["status"] = "failed"
		recordUpdate["error_msg"] = uploadErr.Error()
		recordUpdate["completed_at"] = now
	} else {
		recordUpdate["status"] = "success"
		recordUpdate["error_msg"] = ""
		recordUpdate["completed_at"] = now
		if receipt != nil {
			recordUpdate["receipt_id"] = receipt.ReceiptID
			recordUpdate["remote_video_id"] = receipt.RemoteVideoID
			recordUpdate["remote_url"] = receipt.RemoteURL
		}
	}
	if err := s.db.WithContext(ctx).Model(&models.DirectorPublishRecord{}).Where("id = ?", record.ID).Updates(recordUpdate).Error; err != nil {
		return nil, err
	}

	if task.OutputData == nil {
		task.OutputData = models.JSONMap{}
	}
	publishedMap := map[string]interface{}{}
	if existing, ok := task.OutputData["published_platforms"].(map[string]interface{}); ok {
		for k, v := range existing {
			publishedMap[k] = v
		}
	}
	publishedAt := now.UTC().Format(time.RFC3339)
	platformStatus := "failed"
	platformReason := ""
	if uploadErr == nil {
		platformStatus = "submitted"
	} else {
		platformReason = uploadErr.Error()
	}
	entry := map[string]interface{}{
		"published_at": publishedAt,
		"status":       platformStatus,
		"reason":       platformReason,
		"score":        score,
		"threshold":    threshold,
		"record_id":    record.ID,
	}
	if receipt != nil {
		entry["receipt_id"] = receipt.ReceiptID
		entry["request_id"] = receipt.RequestID
		entry["http_status"] = receipt.HTTPStatus
		entry["remote_video_id"] = receipt.RemoteVideoID
		entry["remote_url"] = receipt.RemoteURL
	}
	publishedMap[platform] = entry
	task.OutputData["published_platforms"] = publishedMap
	task.OutputData["last_publish_platform"] = platform
	task.OutputData["last_publish_at"] = publishedAt
	task.OutputData["quality_gate_passed"] = true
	task.OutputData["quality_gate_reason"] = ""

	if err := s.db.WithContext(ctx).Model(&models.VideoTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"output_data": task.OutputData,
		"updated_at":  now,
	}).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).First(task, task.ID).Error; err != nil {
		return nil, err
	}
	if uploadErr != nil {
		return nil, uploadErr
	}
	return task, nil
}

func (s *StoryboardService) getDirectorCutTask(ctx context.Context, projectID uint, userID uint, exportID string) (*models.VideoTask, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}
	var task models.VideoTask
	if err := s.db.WithContext(ctx).Where("project_id = ? AND user_id = ? AND task_type = ? AND video_id = ?", projectID, userID, "director_cut", exportID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("director cut not found")
		}
		return nil, err
	}
	return &task, nil
}

func (s *StoryboardService) evaluateDirectorGate(selected []models.StoryboardShot, precomputedScore float64) (bool, string, float64) {
	threshold := s.directorQualityThreshold()
	score := precomputedScore
	if score <= 0 {
		weightedScore := 0.0
		weightedDuration := 0.0
		for _, shot := range selected {
			d := float64(maxStoryboardInt(shot.TimelineDurationMs, shot.Duration*1000)) / 1000
			weightedScore += shot.ClipScore * d
			weightedDuration += d
		}
		if weightedDuration > 0 {
			score = roundStoryboardScore(weightedScore / weightedDuration)
		}
	}
	if score >= threshold {
		return true, "", threshold
	}
	return false, fmt.Sprintf("score %.3f below threshold %.2f", score, threshold), threshold
}

func (s *StoryboardService) directorQualityThreshold() float64 {
	if s.cfg == nil {
		return 0.72
	}
	if s.cfg.AI.PublishQualityThreshold <= 0 {
		return 0.72
	}
	return s.cfg.AI.PublishQualityThreshold
}

func (s *StoryboardService) normalizeClipForTimeline(ctx context.Context, dir string, index int, sourcePath string, durationMs int) (string, error) {
	outputPath := filepath.Join(dir, fmt.Sprintf("normalized-%02d.mp4", index+1))
	durationSeconds := math.Max(float64(maxStoryboardInt(durationMs, 1000))/1000, 1)
	if hasAudioTrack(ctx, sourcePath) {
		cmd := exec.CommandContext(ctx, "ffmpeg",
			"-y",
			"-i", sourcePath,
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-pix_fmt", "yuv420p",
			"-c:a", "aac",
			"-ar", "44100",
			"-ac", "2",
			outputPath,
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("normalize clip failed: %v: %s", err, strings.TrimSpace(string(out)))
		}
		return outputPath, nil
	}
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-i", sourcePath,
		"-f", "lavfi",
		"-t", fmt.Sprintf("%.3f", durationSeconds),
		"-i", "anullsrc=channel_layout=stereo:sample_rate=44100",
		"-shortest",
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-ar", "44100",
		"-ac", "2",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("normalize clip failed: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return outputPath, nil
}

func hasAudioTrack(ctx context.Context, filePath string) bool {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "a",
		"-show_entries", "stream=index",
		"-of", "csv=p=0",
		filePath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

func mapTransitionType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "fade":
		return "fade"
	case "wipe":
		return "wipeleft"
	case "match":
		return "smoothleft"
	default:
		return "fade"
	}
}

func (s *StoryboardService) materializeClip(ctx context.Context, dir string, index int, clipURL string) (string, error) {
	if strings.TrimSpace(clipURL) == "" {
		return "", fmt.Errorf("clip url is empty")
	}
	filePath := filepath.Join(dir, fmt.Sprintf("clip-%02d.mp4", index+1))
	if strings.HasPrefix(clipURL, "http://") || strings.HasPrefix(clipURL, "https://") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, clipURL, nil)
		if err != nil {
			return "", err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", fmt.Errorf("download clip failed with status %d", resp.StatusCode)
		}
		f, err := os.Create(filePath)
		if err != nil {
			return "", err
		}
		defer f.Close()
		if _, err := io.Copy(f, resp.Body); err != nil {
			return "", err
		}
		return filePath, nil
	}
	content, err := os.ReadFile(clipURL)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		return "", err
	}
	return filePath, nil
}

func derefUint(v *uint) uint {
	if v == nil {
		return 0
	}
	return *v
}

func maxStoryboardInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func firstNonEmptyString(a string, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func ternaryStoryboardStatus(condition bool, whenTrue string, whenFalse string) string {
	if condition {
		return whenTrue
	}
	return whenFalse
}

func asStoryboardFloat(value interface{}, fallback float64) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func roundStoryboardScore(v float64) float64 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 1
	}
	return math.Round(v*1000) / 1000
}

func normalizeTransitionType(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "fade", "wipe", "match", "cut":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "cut"
	}
}

func normalizeTemplateWeights(t *models.DirectorTemplate) {
	total := t.WeightNarrative + t.WeightVisual + t.WeightEmotion + t.WeightRhythm + t.WeightContinuity
	if total <= 0 {
		t.WeightNarrative = 0.2
		t.WeightVisual = 0.2
		t.WeightEmotion = 0.2
		t.WeightRhythm = 0.2
		t.WeightContinuity = 0.2
		return
	}
	t.WeightNarrative = roundStoryboardScore(clamp01(t.WeightNarrative / total))
	t.WeightVisual = roundStoryboardScore(clamp01(t.WeightVisual / total))
	t.WeightEmotion = roundStoryboardScore(clamp01(t.WeightEmotion / total))
	t.WeightRhythm = roundStoryboardScore(clamp01(t.WeightRhythm / total))
	t.WeightContinuity = roundStoryboardScore(clamp01(t.WeightContinuity / total))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func clampInt(v int, min int, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

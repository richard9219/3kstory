package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/richard9219/3kstory/internal/models"
	"gorm.io/gorm"
)

type NarrationService struct {
	db           *gorm.DB
	aiService    *AIService
	videoService *VideoService
	ttsService   *TTSService
}

func NewNarrationService(db *gorm.DB, aiService *AIService, videoService *VideoService, ttsService *TTSService) *NarrationService {
	return &NarrationService{
		db:           db,
		aiService:    aiService,
		videoService: videoService,
		ttsService:   ttsService,
	}
}

type GenerateNarrationInput struct {
	ProjectID      uint
	UserID         uint
	MovieTitle     string
	Synopsis       string
	Style          string
	TargetDuration int
	Voice          string
	Speed          float64
	AspectRatio    string
	Provider       VideoProvider
}

func (s *NarrationService) GenerateNarrationVideo(ctx context.Context, in GenerateNarrationInput) (*models.VideoTask, error) {
	narration, err := s.aiService.GenerateNarrationScript(ctx, NarrationScriptRequest{
		MovieTitle:     in.MovieTitle,
		Synopsis:       in.Synopsis,
		Style:          in.Style,
		TargetDuration: in.TargetDuration,
	})
	if err != nil {
		return nil, err
	}

	segmentsWithAudio := make([]map[string]interface{}, 0, len(narration.Segments))
	videoSegments := make([]LocalNarrationSegment, 0, len(narration.Segments))
	texts := make([]string, 0, len(narration.Segments))
	for _, seg := range narration.Segments {
		audioURL, synthErr := s.ttsService.Synthesize(seg.NarrationText, in.Voice, in.Speed)
		if synthErr != nil {
			return nil, synthErr
		}
		segmentsWithAudio = append(segmentsWithAudio, map[string]interface{}{
			"title":              seg.Title,
			"narration_text":     seg.NarrationText,
			"estimated_duration": seg.EstimatedDuration,
			"audio_url":          audioURL,
		})
		videoSegments = append(videoSegments, LocalNarrationSegment{
			Title:             seg.Title,
			NarrationText:     seg.NarrationText,
			EstimatedDuration: seg.EstimatedDuration,
			AudioURL:          audioURL,
		})
		texts = append(texts, seg.NarrationText)
	}

	videoPrompt := fmt.Sprintf("《%s》%s风格解说。%s", in.MovieTitle, in.Style, strings.Join(texts, " "))
	videoResult, err := s.videoService.GenerateVideo(ctx, &VideoGenerationRequest{
		ProjectID:         in.ProjectID,
		Prompt:            videoPrompt,
		Provider:          in.Provider,
		Duration:          in.TargetDuration,
		AspectRatio:       in.AspectRatio,
		Mode:              "narration",
		NarrationSegments: videoSegments,
	})
	if err != nil {
		return nil, err
	}

	var completedAt *time.Time
	if videoResult.Status == "completed" {
		now := time.Now()
		completedAt = &now
	}

	task := &models.VideoTask{
		UserID:    in.UserID,
		ProjectID: in.ProjectID,
		TaskType:  "narration_video",
		Title:     narration.Title,
		Provider:  string(videoResult.Provider),
		VideoID:   videoResult.VideoID,
		VideoURL:  videoResult.VideoURL,
		Status:    videoResult.Status,
		InputData: models.JSONMap{
			"movie_title":      in.MovieTitle,
			"synopsis":         in.Synopsis,
			"style":            in.Style,
			"target_duration":  in.TargetDuration,
			"voice":            in.Voice,
			"speed":            in.Speed,
			"aspect_ratio":     in.AspectRatio,
			"generated_prompt": videoPrompt,
		},
		OutputData: models.JSONMap{
			"segments": segmentsWithAudio,
		},
		CompletedAt: completedAt,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, err
	}

	return task, nil
}

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/richard9219/3kstory/internal/config"
	"github.com/richard9219/3kstory/internal/models"
	"gorm.io/gorm"
)

type NarrationService struct {
	db           *gorm.DB
	cfg          *config.Config
	aiService    *AIService
	videoService *VideoService
	ttsService   *TTSService
}

func NewNarrationService(db *gorm.DB, cfg *config.Config, aiService *AIService, videoService *VideoService, ttsService *TTSService) *NarrationService {
	return &NarrationService{
		db:           db,
		cfg:          cfg,
		aiService:    aiService,
		videoService: videoService,
		ttsService:   ttsService,
	}
}

type GenerateNarrationInput struct {
	ProjectID       uint
	UserID          uint
	MovieTitle      string
	Synopsis        string
	Style           string
	TargetDuration  int
	Voice           string
	Speed           float64
	AspectRatio     string
	Provider        VideoProvider
	SourceVideoPath string
	SourceVideoURL  string
	CreativeBrief   string
}

func (s *NarrationService) GenerateNarrationVideo(ctx context.Context, in GenerateNarrationInput) (*models.VideoTask, error) {
	if in.Provider == "" {
		in.Provider = s.videoService.PreferredProviderForTask(VideoTaskNarration)
	}

	narration, err := s.aiService.GenerateNarrationScript(ctx, NarrationScriptRequest{
		MovieTitle:     in.MovieTitle,
		Synopsis:       in.Synopsis,
		Style:          in.Style,
		TargetDuration: in.TargetDuration,
		CreativeBrief:  in.CreativeBrief,
	})
	if err != nil {
		return nil, err
	}

	segmentsWithAudio := make([]map[string]interface{}, 0, len(narration.Segments))
	videoSegments := make([]LocalNarrationSegment, 0, len(narration.Segments))
	texts := make([]string, 0, len(narration.Segments))
	for _, seg := range narration.Segments {
		audioPath, synthErr := s.ttsService.Synthesize(seg.NarrationText, in.Voice, in.Speed)
		if synthErr != nil {
			return nil, synthErr
		}
		segmentsWithAudio = append(segmentsWithAudio, map[string]interface{}{
			"title":              seg.Title,
			"narration_text":     seg.NarrationText,
			"estimated_duration": seg.EstimatedDuration,
			"audio_path":         audioPath,
		})
		videoSegments = append(videoSegments, LocalNarrationSegment{
			Title:             seg.Title,
			NarrationText:     seg.NarrationText,
			EstimatedDuration: seg.EstimatedDuration,
			AudioPath:         audioPath,
		})
		texts = append(texts, seg.NarrationText)
	}

	videoPrompt := fmt.Sprintf("《%s》%s风格解说。%s", in.MovieTitle, in.Style, strings.Join(texts, " "))
	videoResult, err := s.videoService.FailoverGenerateForTask(ctx, VideoTaskNarration, &VideoGenerationRequest{
		ProjectID:         in.ProjectID,
		Prompt:            videoPrompt,
		Provider:          in.Provider,
		Duration:          in.TargetDuration,
		AspectRatio:       in.AspectRatio,
		Mode:              "narration",
		SourceVideoPath:   in.SourceVideoPath,
		SourceVideoURL:    in.SourceVideoURL,
		NarrationSegments: videoSegments,
	})
	if err != nil {
		return nil, err
	}

	assetBundle, err := s.writeNarrationAssets(videoResult.VideoID, narration, in, videoPrompt)
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
			"movie_title":       in.MovieTitle,
			"synopsis":          in.Synopsis,
			"style":             in.Style,
			"target_duration":   in.TargetDuration,
			"voice":             in.Voice,
			"speed":             in.Speed,
			"aspect_ratio":      in.AspectRatio,
			"source_video_path": in.SourceVideoPath,
			"source_video_url":  in.SourceVideoURL,
			"creative_brief":    in.CreativeBrief,
			"generated_prompt":  videoPrompt,
		},
		OutputData: models.JSONMap{
			"segments":             segmentsWithAudio,
			"script_text_path":     assetBundle.TextPath,
			"script_markdown_path": assetBundle.MarkdownPath,
			"script_json_path":     assetBundle.JSONPath,
			"script_text_url":      assetBundle.TextURL,
			"script_markdown_url":  assetBundle.MarkdownURL,
			"script_json_url":      assetBundle.JSONURL,
		},
		CompletedAt: completedAt,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, err
	}

	return task, nil
}

type narrationAssetBundle struct {
	TextPath     string
	MarkdownPath string
	JSONPath     string
	TextURL      string
	MarkdownURL  string
	JSONURL      string
}

func (s *NarrationService) writeNarrationAssets(videoID string, narration *NarrationScriptResult, in GenerateNarrationInput, generatedPrompt string) (*narrationAssetBundle, error) {
	outputDir := s.cfg.AI.NarrationOutputDir
	if strings.TrimSpace(outputDir) == "" {
		outputDir = filepath.Join(".local", "videos", "narration")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create narration output dir failed: %w", err)
	}

	baseName := sanitizeFileComponent(in.MovieTitle)
	if baseName == "" {
		baseName = "narration"
	}
	if strings.TrimSpace(videoID) != "" {
		baseName += "-" + videoID
	} else {
		baseName += "-" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	textPath := filepath.Join(outputDir, baseName+".txt")
	mdPath := filepath.Join(outputDir, baseName+".md")
	jsonPath := filepath.Join(outputDir, baseName+".json")

	textBody := s.buildNarrationText(narration, in)
	mdBody := s.buildNarrationMarkdown(narration, in, generatedPrompt)
	jsonBody := s.buildNarrationJSON(narration, in, generatedPrompt)

	if err := os.WriteFile(textPath, []byte(textBody), 0o644); err != nil {
		return nil, fmt.Errorf("write narration text failed: %w", err)
	}
	if err := os.WriteFile(mdPath, []byte(mdBody), 0o644); err != nil {
		return nil, fmt.Errorf("write narration markdown failed: %w", err)
	}
	if err := os.WriteFile(jsonPath, []byte(jsonBody), 0o644); err != nil {
		return nil, fmt.Errorf("write narration json failed: %w", err)
	}

	publicBase := strings.TrimRight(s.cfg.AI.NarrationPublicBase, "/")
	return &narrationAssetBundle{
		TextPath:     textPath,
		MarkdownPath: mdPath,
		JSONPath:     jsonPath,
		TextURL:      publicBase + "/" + filepath.Base(textPath),
		MarkdownURL:  publicBase + "/" + filepath.Base(mdPath),
		JSONURL:      publicBase + "/" + filepath.Base(jsonPath),
	}, nil
}

func (s *NarrationService) buildNarrationText(narration *NarrationScriptResult, in GenerateNarrationInput) string {
	lines := make([]string, 0, len(narration.Segments)+4)
	lines = append(lines, narration.Title)
	lines = append(lines, "")
	for idx, seg := range narration.Segments {
		lines = append(lines, fmt.Sprintf("%d. %s", idx+1, strings.TrimSpace(seg.Title)))
		lines = append(lines, strings.TrimSpace(seg.NarrationText))
		lines = append(lines, "")
	}
	if brief := strings.TrimSpace(in.CreativeBrief); brief != "" {
		lines = append(lines, "创作要求：")
		lines = append(lines, brief)
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func (s *NarrationService) buildNarrationMarkdown(narration *NarrationScriptResult, in GenerateNarrationInput, generatedPrompt string) string {
	lines := []string{
		"# " + narration.Title,
		"",
		"## 元信息",
		"",
		"- 影片/主题：" + in.MovieTitle,
		"- 风格：" + in.Style,
		"- 目标时长：" + strconv.Itoa(in.TargetDuration) + " 秒",
		"- 画幅：" + in.AspectRatio,
	}
	if strings.TrimSpace(in.CreativeBrief) != "" {
		lines = append(lines, "- 创作要求："+strings.TrimSpace(in.CreativeBrief))
	}
	lines = append(lines, "", "## 解说稿", "")
	for idx, seg := range narration.Segments {
		lines = append(lines, fmt.Sprintf("### %d. %s", idx+1, strings.TrimSpace(seg.Title)))
		lines = append(lines, "")
		lines = append(lines, strings.TrimSpace(seg.NarrationText))
		lines = append(lines, "")
	}
	lines = append(lines, "## 生成提示词", "", generatedPrompt, "")
	return strings.Join(lines, "\n")
}

func (s *NarrationService) buildNarrationJSON(narration *NarrationScriptResult, in GenerateNarrationInput, generatedPrompt string) string {
	payload := map[string]interface{}{
		"title":            narration.Title,
		"movie_title":      in.MovieTitle,
		"style":            in.Style,
		"target_duration":  in.TargetDuration,
		"aspect_ratio":     in.AspectRatio,
		"creative_brief":   in.CreativeBrief,
		"generated_prompt": generatedPrompt,
		"segments":         narration.Segments,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

func sanitizeFileComponent(in string) string {
	value := strings.ToLower(strings.TrimSpace(in))
	if value == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		" ", "-",
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)
	value = replacer.Replace(value)
	return strings.Trim(value, "-")
}

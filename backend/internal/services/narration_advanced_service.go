package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/richard9219/3kstory/internal/models"
)

type GenerateNarrationAdvancedInput struct {
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
	ProviderMode    string
	Providers       []VideoProvider
	CandidateCount  int
	QualityMode     string
	ScoreProfile    string
	AutoPick        bool
	SourceVideoPath string
	SourceVideoURL  string
	CreativeBrief   string
}

type VideoJobDetail struct {
	Job              *models.VideoJob     `json:"job"`
	PipelineStatus   map[string]string    `json:"pipeline_status"`
	Candidates       []*GenerateVideoTask `json:"candidates"`
	SelectedVideoID  string               `json:"selected_video_id"`
	SelectedTaskID   *uint                `json:"selected_task_id"`
	TopScore         float64              `json:"top_score"`
	ScoreProfile     string               `json:"score_profile"`
	QualityMode      string               `json:"quality_mode"`
	ProviderMode     string               `json:"provider_mode"`
	CandidateSummary []map[string]float64 `json:"candidate_summary"`
}

type scoringWeights struct {
	Sync      float64
	Read      float64
	Visual    float64
	Audio     float64
	Stability float64
	Cost      float64
}

func (s *NarrationService) GenerateNarrationAdvanced(ctx context.Context, in GenerateNarrationAdvancedInput) (*models.VideoJob, error) {
	if in.Style == "" {
		in.Style = "深度分析"
	}
	if in.TargetDuration <= 0 {
		in.TargetDuration = 90
	}
	if in.AspectRatio == "" {
		in.AspectRatio = "16:9"
	}
	if in.Speed <= 0 {
		in.Speed = 1.0
	}
	if in.CandidateCount <= 0 {
		in.CandidateCount = 3
	}
	if in.CandidateCount > 8 {
		in.CandidateCount = 8
	}
	if strings.TrimSpace(in.QualityMode) == "" {
		in.QualityMode = "fast"
	}
	if strings.TrimSpace(in.ScoreProfile) == "" {
		in.ScoreProfile = "default"
	}
	if in.ProviderMode == "" {
		in.ProviderMode = "single"
	}

	jobID, err := randomJobID("vj")
	if err != nil {
		return nil, err
	}
	publishThreshold := s.cfg.AI.PublishQualityThreshold

	job := &models.VideoJob{
		JobID:            jobID,
		UserID:           in.UserID,
		ProjectID:        in.ProjectID,
		PipelineType:     "narration_advanced",
		Status:           "queued",
		QueueStatus:      "queued",
		CandidateCount:   in.CandidateCount,
		QualityMode:      in.QualityMode,
		ScoreProfile:     in.ScoreProfile,
		ProviderMode:     normalizeProviderMode(in.ProviderMode),
		AutoPick:         in.AutoPick,
		PublishThreshold: publishThreshold,
		RequestData: models.JSONMap{
			"movie_title":       in.MovieTitle,
			"synopsis":          in.Synopsis,
			"style":             in.Style,
			"target_duration":   in.TargetDuration,
			"voice":             in.Voice,
			"speed":             in.Speed,
			"aspect_ratio":      in.AspectRatio,
			"provider_mode":     normalizeProviderMode(in.ProviderMode),
			"providers":         toProviderStrings(in.Providers),
			"candidate_count":   in.CandidateCount,
			"quality_mode":      in.QualityMode,
			"score_profile":     in.ScoreProfile,
			"auto_pick":         in.AutoPick,
			"source_video_path": in.SourceVideoPath,
			"source_video_url":  in.SourceVideoURL,
			"creative_brief":    in.CreativeBrief,
		},
		ResultData: models.JSONMap{},
	}

	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, err
	}

	select {
	case s.advancedQueue <- advancedNarrationJob{JobID: job.JobID, Input: in}:
		return job, nil
	default:
		_ = s.db.WithContext(ctx).Model(job).Updates(map[string]interface{}{
			"status":               "failed",
			"queue_status":         "failed",
			"error_msg":            "job queue is full, please retry later",
			"publish_gate_passed":  false,
			"publish_block_reason": "queue_overload",
		}).Error
		return nil, fmt.Errorf("advanced narration queue is full")
	}
}

func (s *NarrationService) processAdvancedNarrationJob(ctx context.Context, jobID string, in GenerateNarrationAdvancedInput) {
	var job models.VideoJob
	if err := s.db.WithContext(ctx).Where("job_id = ?", jobID).First(&job).Error; err != nil {
		return
	}

	now := time.Now()
	_ = s.db.WithContext(ctx).Model(&job).Updates(map[string]interface{}{
		"status":       "processing",
		"queue_status": "processing",
		"started_at":   &now,
		"result_data": models.JSONMap{
			"stage": "script",
		},
	}).Error

	narration, err := s.aiService.GenerateNarrationScript(ctx, NarrationScriptRequest{
		MovieTitle:     in.MovieTitle,
		Synopsis:       in.Synopsis,
		Style:          in.Style,
		TargetDuration: in.TargetDuration,
		CreativeBrief:  in.CreativeBrief,
	})
	if err != nil {
		_ = s.db.WithContext(ctx).Model(&job).Updates(map[string]interface{}{
			"status":               "failed",
			"queue_status":         "failed",
			"error_msg":            err.Error(),
			"publish_gate_passed":  false,
			"publish_block_reason": "script_failed",
		}).Error
		return
	}

	_ = s.db.WithContext(ctx).Model(&job).Update("result_data", models.JSONMap{"stage": "tts"}).Error

	segmentsWithAudio := make([]map[string]interface{}, 0, len(narration.Segments))
	videoSegments := make([]LocalNarrationSegment, 0, len(narration.Segments))
	texts := make([]string, 0, len(narration.Segments))
	for _, seg := range narration.Segments {
		audioPath, synthErr := s.ttsService.Synthesize(seg.NarrationText, in.Voice, in.Speed)
		if synthErr != nil {
			_ = s.db.WithContext(ctx).Model(&job).Updates(map[string]interface{}{
				"status":               "failed",
				"queue_status":         "failed",
				"error_msg":            synthErr.Error(),
				"publish_gate_passed":  false,
				"publish_block_reason": "tts_failed",
			}).Error
			return
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
	_ = s.db.WithContext(ctx).Model(&job).Update("result_data", models.JSONMap{"stage": "render"}).Error
	providers := s.resolveCandidateProviders(in)
	candidates := make([]*GenerateVideoTask, 0, in.CandidateCount)

	for i := 0; i < in.CandidateCount; i++ {
		provider := providers[i%len(providers)]
		generationReq := &VideoGenerationRequest{
			ProjectID:         in.ProjectID,
			Prompt:            videoPrompt,
			Provider:          provider,
			Duration:          in.TargetDuration,
			AspectRatio:       in.AspectRatio,
			Mode:              "narration",
			SourceVideoPath:   in.SourceVideoPath,
			SourceVideoURL:    in.SourceVideoURL,
			NarrationSegments: videoSegments,
			Seed:              1000 + i,
			ExtraData: models.JSONMap{
				"quality_mode": in.QualityMode,
				"job_id":       job.JobID,
			},
		}

		result, genErr := s.videoService.FailoverGenerateForTask(ctx, VideoTaskNarration, generationReq)
		task := &GenerateVideoTask{
			UserID:      in.UserID,
			ProjectID:   in.ProjectID,
			TaskType:    "narration_video",
			JobID:       job.JobID,
			CandidateNo: i + 1,
			Title:       narration.Title,
			Provider:    string(provider),
			InputData: models.JSONMap{
				"movie_title":      in.MovieTitle,
				"synopsis":         in.Synopsis,
				"style":            in.Style,
				"target_duration":  in.TargetDuration,
				"voice":            in.Voice,
				"speed":            in.Speed,
				"aspect_ratio":     in.AspectRatio,
				"creative_brief":   in.CreativeBrief,
				"generated_prompt": videoPrompt,
				"quality_mode":     in.QualityMode,
				"score_profile":    in.ScoreProfile,
			},
			OutputData: models.JSONMap{
				"segments":         segmentsWithAudio,
				"candidate_no":     i + 1,
				"provider_mode":    normalizeProviderMode(in.ProviderMode),
				"selected":         false,
				"job_id":           job.JobID,
				"score_profile":    in.ScoreProfile,
				"quality_mode":     in.QualityMode,
				"provider_request": provider,
			},
		}

		if genErr != nil {
			task.Status = "failed"
			task.ErrorMsg = genErr.Error()
			task.Score = 0
			task.ScoreDetail = models.JSONMap{"error": genErr.Error()}
		} else {
			task.VideoID = result.VideoID
			task.VideoURL = result.VideoURL
			task.Status = result.Status
			if result.Status == "completed" {
				now := time.Now()
				task.CompletedAt = &now
			}
			score, detail := s.scoreNarrationCandidate(*task, in, videoSegments)
			task.Score = score
			task.ScoreDetail = detail
		}

		if err := s.videoService.SaveVideoTask(ctx, task); err != nil {
			_ = s.db.WithContext(ctx).Model(&job).Updates(map[string]interface{}{
				"status":               "failed",
				"queue_status":         "failed",
				"error_msg":            err.Error(),
				"publish_gate_passed":  false,
				"publish_block_reason": "persist_failed",
			}).Error
			return
		}
		candidates = append(candidates, task)
	}

	_ = s.db.WithContext(ctx).Model(&job).Update("result_data", models.JSONMap{"stage": "score"}).Error
	s.sortAndRankCandidates(candidates)
	for _, task := range candidates {
		if err := s.videoService.SaveVideoTask(ctx, task); err != nil {
			_ = s.db.WithContext(ctx).Model(&job).Updates(map[string]interface{}{
				"status":               "failed",
				"queue_status":         "failed",
				"error_msg":            err.Error(),
				"publish_gate_passed":  false,
				"publish_block_reason": "rank_persist_failed",
			}).Error
			return
		}
	}

	_ = s.db.WithContext(ctx).Model(&job).Update("result_data", models.JSONMap{"stage": "selection"}).Error
	selected := s.pickCandidate(candidates, in.AutoPick)
	publishGatePassed := false
	publishBlockReason := "no_candidate_selected"
	if selected != nil {
		selected.IsSelected = true
		if err := s.videoService.SaveVideoTask(ctx, selected); err != nil {
			_ = s.db.WithContext(ctx).Model(&job).Updates(map[string]interface{}{
				"status":               "failed",
				"queue_status":         "failed",
				"error_msg":            err.Error(),
				"publish_gate_passed":  false,
				"publish_block_reason": "select_persist_failed",
			}).Error
			return
		}
		job.SelectedTaskID = &selected.ID
		job.SelectedVideoID = selected.VideoID
		publishGatePassed = selected.Score >= job.PublishThreshold
		if publishGatePassed {
			publishBlockReason = ""
		} else {
			publishBlockReason = fmt.Sprintf("score %.3f below threshold %.2f", selected.Score, job.PublishThreshold)
		}
	} else {
		publishBlockReason = "auto_pick_disabled_or_no_completed_candidate"
	}

	now = time.Now()
	job.CompletedAt = &now
	job.Status = "completed"
	job.QueueStatus = "done"
	job.PublishGatePassed = publishGatePassed
	job.PublishBlockReason = publishBlockReason
	job.ResultData = models.JSONMap{
		"candidate_count":      len(candidates),
		"top_score":            topCandidateScore(candidates),
		"selected_video":       job.SelectedVideoID,
		"auto_pick":            in.AutoPick,
		"publish_gate_passed":  publishGatePassed,
		"publish_block_reason": publishBlockReason,
		"publish_threshold":    job.PublishThreshold,
		"stage":                "done",
	}

	if err := s.db.WithContext(ctx).Save(job).Error; err != nil {
		return
	}
}

func (s *NarrationService) GetVideoJobDetail(ctx context.Context, userID uint, projectID uint, jobID string) (*VideoJobDetail, error) {
	var job models.VideoJob
	if err := s.db.WithContext(ctx).Where("job_id = ? AND user_id = ? AND project_id = ?", jobID, userID, projectID).First(&job).Error; err != nil {
		return nil, err
	}

	candidates, err := s.videoService.ListVideoTasksByJob(ctx, userID, projectID, jobID)
	if err != nil {
		return nil, err
	}

	stage, _ := job.ResultData["stage"].(string)
	if stage == "" {
		stage = "queued"
	}
	pipelineStatus := map[string]string{
		"script":    "pending",
		"tts":       "pending",
		"render":    "pending",
		"score":     "pending",
		"selection": "pending",
		"gate":      ternaryString(job.PublishGatePassed, "passed", "blocked"),
	}
	if job.QueueStatus == "queued" {
		pipelineStatus["script"] = "queued"
	}
	if stage == "script" || stage == "tts" || stage == "render" || stage == "score" || stage == "selection" || stage == "done" {
		pipelineStatus["script"] = "completed"
	}
	if stage == "tts" || stage == "render" || stage == "score" || stage == "selection" || stage == "done" {
		pipelineStatus["tts"] = "completed"
	}
	if stage == "render" || stage == "score" || stage == "selection" || stage == "done" {
		pipelineStatus["render"] = ternaryString(job.Status == "failed", "failed", "completed")
	}
	if stage == "score" || stage == "selection" || stage == "done" {
		pipelineStatus["score"] = ternaryString(job.Status == "failed", "failed", "completed")
	}
	if stage == "selection" || stage == "done" {
		pipelineStatus["selection"] = ternaryString(job.Status == "failed", "failed", "completed")
	}

	detail := &VideoJobDetail{
		Job:             &job,
		Candidates:      candidates,
		SelectedVideoID: job.SelectedVideoID,
		SelectedTaskID:  job.SelectedTaskID,
		ScoreProfile:    job.ScoreProfile,
		QualityMode:     job.QualityMode,
		ProviderMode:    job.ProviderMode,
		PipelineStatus:  pipelineStatus,
	}

	if len(candidates) > 0 {
		detail.TopScore = candidates[0].Score
		summary := make([]map[string]float64, 0, len(candidates))
		for _, c := range candidates {
			summary = append(summary, map[string]float64{
				"candidate_no": float64(c.CandidateNo),
				"score":        c.Score,
				"rank":         float64(c.Rank),
			})
		}
		detail.CandidateSummary = summary
	}

	return detail, nil
}

func (s *NarrationService) SelectVideoCandidate(ctx context.Context, userID uint, projectID uint, jobID string, videoID string, reason string) (*GenerateVideoTask, error) {
	task, err := s.videoService.SelectVideoTaskByVideoID(ctx, userID, projectID, jobID, videoID)
	if err != nil {
		return nil, err
	}

	var job models.VideoJob
	if err := s.db.WithContext(ctx).Where("job_id = ? AND user_id = ? AND project_id = ?", jobID, userID, projectID).First(&job).Error; err != nil {
		return nil, err
	}

	resultData := job.ResultData
	if resultData == nil {
		resultData = models.JSONMap{}
	}
	if strings.TrimSpace(reason) != "" {
		resultData["selection_reason"] = reason
	}
	publishGatePassed := task.Score >= job.PublishThreshold
	publishBlockReason := ""
	if !publishGatePassed {
		publishBlockReason = fmt.Sprintf("score %.3f below threshold %.2f", task.Score, job.PublishThreshold)
	}
	resultData["publish_gate_passed"] = publishGatePassed
	resultData["publish_block_reason"] = publishBlockReason
	resultData["selected_video"] = videoID

	updates := map[string]interface{}{
		"selected_video_id":    videoID,
		"selected_task_id":     task.ID,
		"result_data":          resultData,
		"publish_gate_passed":  publishGatePassed,
		"publish_block_reason": publishBlockReason,
	}
	if err := s.db.WithContext(ctx).Model(&models.VideoJob{}).
		Where("job_id = ? AND user_id = ? AND project_id = ?", jobID, userID, projectID).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	return task, nil
}

func (s *NarrationService) CheckAutoPublishGate(ctx context.Context, userID uint, projectID uint, jobID string) (bool, string, error) {
	var job models.VideoJob
	if err := s.db.WithContext(ctx).Where("job_id = ? AND user_id = ? AND project_id = ?", jobID, userID, projectID).First(&job).Error; err != nil {
		return false, "", err
	}
	if job.PublishGatePassed {
		return true, "", nil
	}
	if strings.TrimSpace(job.PublishBlockReason) != "" {
		return false, job.PublishBlockReason, nil
	}
	return false, "quality gate blocked", nil
}

func (s *NarrationService) resolveCandidateProviders(in GenerateNarrationAdvancedInput) []VideoProvider {
	mode := normalizeProviderMode(in.ProviderMode)
	if mode == "single" {
		if in.Provider != "" {
			return []VideoProvider{in.Provider}
		}
		return []VideoProvider{s.videoService.PreferredProviderForTask(VideoTaskNarration)}
	}

	providers := make([]VideoProvider, 0)
	for _, p := range in.Providers {
		if p != "" {
			providers = append(providers, p)
		}
	}
	if len(providers) == 0 {
		providers = s.videoService.ConfiguredProvidersForTask(VideoTaskNarration)
	}
	if len(providers) == 0 {
		providers = []VideoProvider{s.videoService.PreferredProviderForTask(VideoTaskNarration)}
	}
	return providers
}

func (s *NarrationService) scoreNarrationCandidate(task GenerateVideoTask, in GenerateNarrationAdvancedInput, segments []LocalNarrationSegment) (float64, models.JSONMap) {
	weights := scoringProfileWeights(in.ScoreProfile)
	totalAudioSec := 0.0
	totalChars := 0
	missingAudio := 0
	for _, seg := range segments {
		segSec := float64(seg.EstimatedDuration)
		if segSec <= 0 {
			segSec = 6
		}
		totalAudioSec += segSec
		totalChars += len([]rune(strings.TrimSpace(seg.NarrationText)))
		if strings.TrimSpace(seg.AudioPath) == "" && strings.TrimSpace(seg.AudioURL) == "" {
			missingAudio++
		}
	}
	if totalAudioSec <= 0 {
		totalAudioSec = float64(in.TargetDuration)
	}

	videoDuration, vw, vh := probeVideoInfo(task.VideoURL)

	syncScore := boundedScore(1 - math.Min(math.Abs(totalAudioSec-float64(in.TargetDuration))/math.Max(float64(in.TargetDuration), 1), 1))
	if videoDuration > 0 {
		videoDrift := math.Abs(videoDuration-totalAudioSec) / math.Max(totalAudioSec, 1)
		syncScore = boundedScore((syncScore + (1 - math.Min(videoDrift, 1))) / 2)
	}

	cps := float64(totalChars) / math.Max(totalAudioSec, 1)
	readability := readabilityScore(cps)

	visual := 0.65
	if vw > 0 && vh > 0 {
		if strings.TrimSpace(in.AspectRatio) == "9:16" {
			if vh >= vw {
				visual = 0.95
			} else {
				visual = 0.7
			}
		} else {
			if vw >= vh {
				visual = 0.95
			} else {
				visual = 0.7
			}
		}
		if minInt(vw, vh) < 720 {
			visual = math.Max(0.55, visual-0.15)
		}
	}

	audioQuality := boundedScore(1 - float64(missingAudio)/math.Max(float64(len(segments)), 1))
	stability := 0.0
	if task.Status == "completed" {
		stability = 1.0
	}

	costScore := boundedScore(1 - estimatedCost(task.Provider, in.TargetDuration)/4.0)
	total :=
		weights.Sync*syncScore +
			weights.Read*readability +
			weights.Visual*visual +
			weights.Audio*audioQuality +
			weights.Stability*stability +
			weights.Cost*costScore

	detail := models.JSONMap{
		"sync":            roundScore(syncScore),
		"readability":     roundScore(readability),
		"visual":          roundScore(visual),
		"audio":           roundScore(audioQuality),
		"stability":       roundScore(stability),
		"cost_efficiency": roundScore(costScore),
		"cps":             roundScore(cps),
		"target_duration": in.TargetDuration,
		"video_duration":  roundScore(videoDuration),
		"resolution":      fmt.Sprintf("%dx%d", vw, vh),
		"profile":         in.ScoreProfile,
		"weights": models.JSONMap{
			"sync":      weights.Sync,
			"read":      weights.Read,
			"visual":    weights.Visual,
			"audio":     weights.Audio,
			"stability": weights.Stability,
			"cost":      weights.Cost,
		},
	}
	return roundScore(total), detail
}

func scoringProfileWeights(profile string) scoringWeights {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case "short_drama":
		return scoringWeights{Sync: 0.25, Read: 0.15, Visual: 0.30, Audio: 0.10, Stability: 0.10, Cost: 0.10}
	case "movie_narration":
		return scoringWeights{Sync: 0.35, Read: 0.25, Visual: 0.15, Audio: 0.15, Stability: 0.05, Cost: 0.05}
	default:
		return scoringWeights{Sync: 0.30, Read: 0.20, Visual: 0.20, Audio: 0.15, Stability: 0.10, Cost: 0.05}
	}
}

func (s *NarrationService) sortAndRankCandidates(tasks []*GenerateVideoTask) {
	sort.SliceStable(tasks, func(i, j int) bool {
		if tasks[i].Score == tasks[j].Score {
			return tasks[i].CandidateNo < tasks[j].CandidateNo
		}
		return tasks[i].Score > tasks[j].Score
	})
	for i := range tasks {
		tasks[i].Rank = i + 1
		tasks[i].IsSelected = false
	}
}

func (s *NarrationService) pickCandidate(tasks []*GenerateVideoTask, autoPick bool) *GenerateVideoTask {
	if len(tasks) == 0 || !autoPick {
		return nil
	}
	for _, t := range tasks {
		if t.Status == "completed" {
			return t
		}
	}
	return tasks[0]
}

func topCandidateScore(tasks []*GenerateVideoTask) float64 {
	if len(tasks) == 0 {
		return 0
	}
	return tasks[0].Score
}

func normalizeProviderMode(mode string) string {
	m := strings.ToLower(strings.TrimSpace(mode))
	if m == "multi" {
		return "multi"
	}
	return "single"
}

func toProviderStrings(in []VideoProvider) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		if v == "" {
			continue
		}
		out = append(out, string(v))
	}
	return out
}

func randomJobID(prefix string) (string, error) {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%d_%s", prefix, time.Now().Unix(), hex.EncodeToString(buf)), nil
}

func probeVideoInfo(input string) (duration float64, width int, height int) {
	if strings.TrimSpace(input) == "" {
		return 0, 0, 0
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return 0, 0, 0
	}
	out, err := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		input,
	).Output()
	if err != nil {
		return 0, 0, 0
	}
	lines := strings.Fields(strings.TrimSpace(string(out)))
	if len(lines) < 3 {
		return 0, 0, 0
	}
	width, _ = strconv.Atoi(lines[0])
	height, _ = strconv.Atoi(lines[1])
	duration, _ = strconv.ParseFloat(lines[2], 64)
	return duration, width, height
}

func readabilityScore(cps float64) float64 {
	switch {
	case cps <= 4.5:
		return 1
	case cps <= 6.0:
		return 0.82
	case cps <= 7.5:
		return 0.62
	case cps <= 9.0:
		return 0.45
	default:
		return 0.3
	}
}

func estimatedCost(provider string, durationSec int) float64 {
	basePerMin := map[string]float64{
		"runway":   2.8,
		"pika":     1.9,
		"local":    0.2,
		"minimax":  1.2,
		"seedance": 2.2,
		"comfy":    0.5,
	}
	base := basePerMin[strings.ToLower(strings.TrimSpace(provider))]
	if base <= 0 {
		base = 1.0
	}
	minutes := math.Max(float64(durationSec)/60.0, 0.2)
	return base * minutes
}

func roundScore(v float64) float64 {
	return math.Round(v*1000) / 1000
}

func boundedScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ternaryString(condition bool, a string, b string) string {
	if condition {
		return a
	}
	return b
}

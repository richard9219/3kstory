package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/richard9219/3kstory/internal/services"
)

type NarrationHandler struct {
	projectService   *services.ProjectService
	narrationService *services.NarrationService
}

func NewNarrationHandler(projectService *services.ProjectService, narrationService *services.NarrationService) *NarrationHandler {
	return &NarrationHandler{
		projectService:   projectService,
		narrationService: narrationService,
	}
}

type GenerateNarrationRequest struct {
	MovieTitle      string  `json:"movie_title" binding:"required"`
	Synopsis        string  `json:"synopsis"`
	Style           string  `json:"style"`
	TargetDuration  int     `json:"target_duration" binding:"min=10,max=600"`
	Voice           string  `json:"voice"`
	Speed           float64 `json:"speed" binding:"min=0.5,max=2"`
	Provider        string  `json:"provider" binding:"omitempty,oneof=runway pika local minimax seedance comfy"`
	AspectRatio     string  `json:"aspect_ratio" binding:"omitempty,oneof=16:9 9:16"`
	SourceVideoPath string  `json:"source_video_path"`
	SourceVideoURL  string  `json:"source_video_url"`
	CreativeBrief   string  `json:"creative_brief"`
}

type GenerateNarrationAdvancedRequest struct {
	MovieTitle      string   `json:"movie_title" binding:"required"`
	Synopsis        string   `json:"synopsis"`
	Style           string   `json:"style"`
	TargetDuration  int      `json:"target_duration" binding:"min=10,max=600"`
	Voice           string   `json:"voice"`
	Speed           float64  `json:"speed" binding:"min=0.5,max=2"`
	Provider        string   `json:"provider" binding:"omitempty,oneof=runway pika local minimax seedance comfy"`
	ProviderMode    string   `json:"provider_mode" binding:"omitempty,oneof=single multi"`
	Providers       []string `json:"providers"`
	CandidateCount  int      `json:"candidate_count" binding:"min=1,max=8"`
	QualityMode     string   `json:"quality_mode" binding:"omitempty,oneof=fast quality"`
	ScoreProfile    string   `json:"score_profile" binding:"omitempty,oneof=default short_drama movie_narration"`
	AutoPick        *bool    `json:"auto_pick"`
	AspectRatio     string   `json:"aspect_ratio" binding:"omitempty,oneof=16:9 9:16"`
	SourceVideoPath string   `json:"source_video_path"`
	SourceVideoURL  string   `json:"source_video_url"`
	CreativeBrief   string   `json:"creative_brief"`
}

type SelectCandidateRequest struct {
	VideoID string `json:"video_id" binding:"required"`
	Reason  string `json:"reason"`
}

type AutoPublishRequest struct {
	Platform string `json:"platform"`
}

// GenerateNarrationVideo 生成电影/电视剧解说视频
// POST /api/v1/projects/:id/generate-narration
func (h *NarrationHandler) GenerateNarrationVideo(c *gin.Context) {
	projectID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	projectID := uint(projectID64)

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	project, err := h.projectService.GetProjectWithScenes(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if project.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this project"})
		return
	}

	var req GenerateNarrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Style == "" {
		req.Style = "深度分析"
	}
	if req.TargetDuration == 0 {
		req.TargetDuration = 90
	}
	if req.AspectRatio == "" {
		req.AspectRatio = "16:9"
	}
	if req.Speed == 0 {
		req.Speed = 1.0
	}

	task, err := h.narrationService.GenerateNarrationVideo(c.Request.Context(), services.GenerateNarrationInput{
		ProjectID:       projectID,
		UserID:          userID.(uint),
		MovieTitle:      req.MovieTitle,
		Synopsis:        req.Synopsis,
		Style:           req.Style,
		TargetDuration:  req.TargetDuration,
		Voice:           req.Voice,
		Speed:           req.Speed,
		Provider:        services.VideoProvider(req.Provider),
		AspectRatio:     req.AspectRatio,
		SourceVideoPath: req.SourceVideoPath,
		SourceVideoURL:  req.SourceVideoURL,
		CreativeBrief:   req.CreativeBrief,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "narration video generation failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"task_id":   task.ID,
		"video_id":  task.VideoID,
		"status":    task.Status,
		"provider":  task.Provider,
		"video_url": task.VideoURL,
	})
}

// GenerateNarrationAdvanced 生成多候选解说视频并自动评分选片
// POST /api/v1/projects/:id/generate-narration-advanced
func (h *NarrationHandler) GenerateNarrationAdvanced(c *gin.Context) {
	projectID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	projectID := uint(projectID64)

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	project, err := h.projectService.GetProjectWithScenes(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if project.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this project"})
		return
	}

	var req GenerateNarrationAdvancedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	providers := make([]services.VideoProvider, 0, len(req.Providers))
	for _, item := range req.Providers {
		providers = append(providers, services.VideoProvider(item))
	}

	autoPick := true
	if req.AutoPick != nil {
		autoPick = *req.AutoPick
	}

	job, err := h.narrationService.GenerateNarrationAdvanced(c.Request.Context(), services.GenerateNarrationAdvancedInput{
		ProjectID:       projectID,
		UserID:          userID.(uint),
		MovieTitle:      req.MovieTitle,
		Synopsis:        req.Synopsis,
		Style:           req.Style,
		TargetDuration:  req.TargetDuration,
		Voice:           req.Voice,
		Speed:           req.Speed,
		Provider:        services.VideoProvider(req.Provider),
		ProviderMode:    req.ProviderMode,
		Providers:       providers,
		CandidateCount:  req.CandidateCount,
		QualityMode:     req.QualityMode,
		ScoreProfile:    req.ScoreProfile,
		AutoPick:        autoPick,
		AspectRatio:     req.AspectRatio,
		SourceVideoPath: req.SourceVideoPath,
		SourceVideoURL:  req.SourceVideoURL,
		CreativeBrief:   req.CreativeBrief,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "advanced narration generation failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":   job.JobID,
		"status":   job.Status,
		"message":  "job queued, poll /video-jobs/:jobID for progress",
		"queue":    job.QueueStatus,
		"estimate": "processing asynchronously",
	})
}

// GetVideoJob 获取高级视频任务聚合状态
// GET /api/v1/projects/:id/video-jobs/:jobID
func (h *NarrationHandler) GetVideoJob(c *gin.Context) {
	projectID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	projectID := uint(projectID64)

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	project, err := h.projectService.GetProjectWithScenes(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if project.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this project"})
		return
	}

	jobID := c.Param("jobID")
	detail, err := h.narrationService.GetVideoJobDetail(c.Request.Context(), userID.(uint), projectID, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": detail})
}

// SelectVideoCandidate 人工改选候选视频
// POST /api/v1/projects/:id/video-jobs/:jobID/select
func (h *NarrationHandler) SelectVideoCandidate(c *gin.Context) {
	projectID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	projectID := uint(projectID64)

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	project, err := h.projectService.GetProjectWithScenes(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if project.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this project"})
		return
	}

	var req SelectCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := c.Param("jobID")
	task, err := h.narrationService.SelectVideoCandidate(c.Request.Context(), userID.(uint), projectID, jobID, req.VideoID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to select candidate", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id":   jobID,
		"video_id": task.VideoID,
		"task_id":  task.ID,
		"score":    task.Score,
		"rank":     task.Rank,
		"status":   "selected",
	})
}

// AutoPublishWithGate 执行发布前质量阈值门禁校验
// POST /api/v1/projects/:id/video-jobs/:jobID/auto-publish
func (h *NarrationHandler) AutoPublishWithGate(c *gin.Context) {
	projectID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	projectID := uint(projectID64)

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	project, err := h.projectService.GetProjectWithScenes(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if project.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this project"})
		return
	}

	var req AutoPublishRequest
	_ = c.ShouldBindJSON(&req)

	jobID := c.Param("jobID")
	passed, reason, err := h.narrationService.CheckAutoPublishGate(c.Request.Context(), userID.(uint), projectID, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found", "details": err.Error()})
		return
	}

	if !passed {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"job_id":   jobID,
			"platform": req.Platform,
			"status":   "blocked",
			"reason":   reason,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id":   jobID,
		"platform": req.Platform,
		"status":   "passed",
		"message":  "quality gate passed; ready for publish workflow",
	})
}

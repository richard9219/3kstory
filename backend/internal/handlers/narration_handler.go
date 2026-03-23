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

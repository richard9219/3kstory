package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/richard9219/3kstory/internal/services"
)

type VideoHandler struct {
	videoService   *services.VideoService
	projectService *services.ProjectService
}

func NewVideoHandler(videoService *services.VideoService, projectService *services.ProjectService) *VideoHandler {
	return &VideoHandler{
		videoService:   videoService,
		projectService: projectService,
	}
}

// GenerateVideoRequest represents the request to generate a video
type GenerateVideoRequest struct {
	SceneID     uint   `json:"scene_id" binding:"required"`
	Prompt      string `json:"prompt" binding:"required"`
	Provider    string `json:"provider" binding:"required,oneof=runway pika"`
	ImageURL    string `json:"image_url"`
	Duration    int    `json:"duration" binding:"min=1,max=60"`
	AspectRatio string `json:"aspect_ratio" binding:"oneof=16:9 9:16"`
}

// GenerateVideoResponse represents the response from video generation
type GenerateVideoResponse struct {
	TaskID   uint   `json:"task_id"`
	VideoID  string `json:"video_id"`
	Status   string `json:"status"`
	Provider string `json:"provider"`
	Message  string `json:"message,omitempty"`
	VideoURL string `json:"video_url,omitempty"`
}

// GenerateVideo generates a video for a scene
// POST /api/v1/projects/:id/generate-video
func (h *VideoHandler) GenerateVideo(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	// Verify project ownership
	project, err := h.projectService.GetProjectWithScenes(uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	if project.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this project"})
		return
	}

	var req GenerateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Duration == 0 {
		req.Duration = 30
	}
	if req.AspectRatio == "" {
		req.AspectRatio = "16:9"
	}

	// Create video generation request
	videoReq := &services.VideoGenerationRequest{
		ProjectID:   uint(projectID),
		SceneID:     req.SceneID,
		Prompt:      req.Prompt,
		Provider:    services.VideoProvider(req.Provider),
		ImageURL:    req.ImageURL,
		Duration:    req.Duration,
		AspectRatio: req.AspectRatio,
	}

	// Generate video with failover support
	result, err := h.videoService.FailoverGenerate(c, videoReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Video generation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, GenerateVideoResponse{
		VideoID:  result.VideoID,
		Status:   result.Status,
		Provider: string(result.Provider),
		Message:  "Video generation started. Poll for status updates.",
		VideoURL: result.VideoURL,
	})
}

// GetVideoStatusRequest represents the request to get video status
type GetVideoStatusRequest struct {
	VideoID  string `json:"video_id" binding:"required"`
	Provider string `json:"provider" binding:"required,oneof=runway pika"`
}

// GetVideoStatus retrieves the status of a video generation job
// GET /api/v1/projects/:id/video-status
func (h *VideoHandler) GetVideoStatus(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	// Verify project ownership
	project, err := h.projectService.GetProjectWithScenes(uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	if project.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this project"})
		return
	}

	var req GetVideoStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.videoService.PollVideoStatus(c, req.VideoID, services.VideoProvider(req.Provider))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get video status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GenerateVideoResponse{
		VideoID:  result.VideoID,
		Status:   result.Status,
		Provider: string(result.Provider),
		VideoURL: result.VideoURL,
	})
}

// ListVideosRequest represents the request to list videos
type ListVideosRequest struct {
	Status string `form:"status"`
	Limit  int    `form:"limit" binding:"max=100"`
	Offset int    `form:"offset"`
}

// ListVideos retrieves all videos for a project
// GET /api/v1/projects/:id/videos
func (h *VideoHandler) ListVideos(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	// Verify project ownership
	project, err := h.projectService.GetProjectWithScenes(uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	if project.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this project"})
		return
	}

	var req ListVideosRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 20
	}

	tasks, err := h.videoService.ListVideoTasks(c, uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve videos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(tasks),
		"data":  tasks,
	})
}

// CancelVideoGeneration cancels an ongoing video generation job
// DELETE /api/v1/projects/:id/video/:videoID
func (h *VideoHandler) CancelVideoGeneration(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	// Verify project ownership
	project, err := h.projectService.GetProjectWithScenes(uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	if project.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this project"})
		return
	}

	videoID := c.Param("videoID")

	// TODO: Implement video cancellation logic with provider APIs

	c.JSON(http.StatusOK, gin.H{
		"message":  "Video generation cancelled successfully",
		"video_id": videoID,
	})
}

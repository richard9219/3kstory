package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/richard9219/3kstory/internal/models"
	"github.com/richard9219/3kstory/internal/services"
)

type StoryboardHandler struct {
	service *services.StoryboardService
}

func NewStoryboardHandler(service *services.StoryboardService) *StoryboardHandler {
	return &StoryboardHandler{service: service}
}

func (h *StoryboardHandler) parseProjectAndUser(c *gin.Context) (uint, uint, bool) {
	projectID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return 0, 0, false
	}
	userID := c.GetUint("user_id")
	return uint(projectID64), userID, true
}

func (h *StoryboardHandler) ListProjectShots(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	shots, err := h.service.ListProjectShots(c.Request.Context(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": shots, "total": len(shots)})
}

type createShotRequest struct {
	SceneID           *uint  `json:"scene_id"`
	Chapter           string `json:"chapter"`
	ShotNumber        int    `json:"shot_number"`
	SortOrder         int    `json:"sort_order"`
	Title             string `json:"title" binding:"required"`
	Description       string `json:"description"`
	CameraLanguage    string `json:"camera_language"`
	Duration          int    `json:"duration"`
	AspectRatio       string `json:"aspect_ratio"`
	Prompt            string `json:"prompt"`
	NegativePrompt    string `json:"negative_prompt"`
	ReferenceImageURL string `json:"reference_image_url"`
	Status            string `json:"status"`
}

type reorderShotsRequest struct {
	OrderedIDs []uint `json:"ordered_ids" binding:"required"`
}

type importShotsRequest struct {
	Shots []createShotRequest `json:"shots" binding:"required"`
}

type createShotVersionRequest struct {
	SourceShotID uint   `json:"source_shot_id" binding:"required"`
	VersionNote  string `json:"version_note"`
}

func (h *StoryboardHandler) CreateShot(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	var req createShotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	shot := &models.StoryboardShot{
		UserID:            userID,
		ProjectID:         projectID,
		SceneID:           req.SceneID,
		Chapter:           req.Chapter,
		ShotNumber:        req.ShotNumber,
		SortOrder:         req.SortOrder,
		Title:             req.Title,
		Description:       req.Description,
		CameraLanguage:    req.CameraLanguage,
		Duration:          req.Duration,
		AspectRatio:       req.AspectRatio,
		Prompt:            req.Prompt,
		NegativePrompt:    req.NegativePrompt,
		ReferenceImageURL: req.ReferenceImageURL,
		Status:            req.Status,
		Version:           1,
	}
	if err := h.service.CreateShot(c.Request.Context(), shot); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, shot)
}

func (h *StoryboardHandler) ReorderShots(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}

	var req reorderShotsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ReorderShots(c.Request.Context(), projectID, userID, req.OrderedIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reordered": len(req.OrderedIDs)})
}

func (h *StoryboardHandler) ImportShots(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}

	var req importShotsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shots := make([]models.StoryboardShot, 0, len(req.Shots))
	for _, item := range req.Shots {
		shots = append(shots, models.StoryboardShot{
			UserID:            userID,
			ProjectID:         projectID,
			SceneID:           item.SceneID,
			Chapter:           item.Chapter,
			ShotNumber:        item.ShotNumber,
			SortOrder:         item.SortOrder,
			Title:             item.Title,
			Description:       item.Description,
			CameraLanguage:    item.CameraLanguage,
			Duration:          item.Duration,
			AspectRatio:       item.AspectRatio,
			Prompt:            item.Prompt,
			NegativePrompt:    item.NegativePrompt,
			ReferenceImageURL: item.ReferenceImageURL,
			Status:            item.Status,
			Version:           1,
		})
	}

	count, err := h.service.BulkCreateShots(c.Request.Context(), shots)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "imported": count})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"imported": count})
}

func (h *StoryboardHandler) CreateShotVersion(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}

	var req createShotVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shot, err := h.service.CreateShotVersion(c.Request.Context(), projectID, userID, req.SourceShotID, req.VersionNote)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, shot)
}

func (h *StoryboardHandler) GetVersionTree(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}

	nodes, err := h.service.ListVersionTree(c.Request.Context(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": nodes, "total": len(nodes)})
}

func (h *StoryboardHandler) BootstrapFromScenes(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	count, err := h.service.BootstrapFromScenes(c.Request.Context(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bootstrapped": count})
}

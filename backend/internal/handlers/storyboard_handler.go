package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/richard9219/3kstory/internal/models"
	"github.com/richard9219/3kstory/internal/services"
)

type shotActionRequest struct {
	Provider          string `json:"provider"`
	Model             string `json:"model"`
	Duration          int    `json:"duration"`
	AspectRatio       string `json:"aspect_ratio"`
	Prompt            string `json:"prompt"`
	NegativePrompt    string `json:"negative_prompt"`
	ReferenceImageURL string `json:"reference_image_url"`
	WorkflowPath      string `json:"workflow_path"`
	VersionNote       string `json:"version_note"`
}

type updateShotRequest struct {
	TrackIndex           *int    `json:"track_index"`
	Chapter              *string `json:"chapter"`
	Title                *string `json:"title"`
	Description          *string `json:"description"`
	CameraLanguage       *string `json:"camera_language"`
	EmotionTone          *string `json:"emotion_tone"`
	Duration             *int    `json:"duration"`
	AspectRatio          *string `json:"aspect_ratio"`
	Prompt               *string `json:"prompt"`
	NegativePrompt       *string `json:"negative_prompt"`
	ReferenceImageURL    *string `json:"reference_image_url"`
	TimelineStartMs      *int    `json:"timeline_start_ms"`
	TimelineDurationMs   *int    `json:"timeline_duration_ms"`
	TransitionType       *string `json:"transition_type"`
	TransitionDurationMs *int    `json:"transition_duration_ms"`
	Locked               *bool   `json:"locked"`
	Status               *string `json:"status"`
}

type directorPublishRequest struct {
	Platform string `json:"platform"`
}

type directorPublishRetryRequest struct {
	Reason string `json:"reason"`
}

type directorTemplateRequest struct {
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

type autoDirectorStrategyRequest struct {
	Genre       string `json:"genre"`
	TemplateID  *uint  `json:"template_id"`
	Apply       bool   `json:"apply"`
	TunePercent int    `json:"tune_percent"`
}

type directorABCompareRequest struct {
	TemplateAID   uint   `json:"template_a_id" binding:"required"`
	TemplateBID   uint   `json:"template_b_id" binding:"required"`
	Genre         string `json:"genre"`
	ApplyBest     bool   `json:"apply_best"`
	TunePercent   int    `json:"tune_percent"`
	RenderBestCut bool   `json:"render_best_cut"`
}

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

func (h *StoryboardHandler) GetTimeline(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	timeline, err := h.service.GetTimeline(c.Request.Context(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, timeline)
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

func (h *StoryboardHandler) UpdateShot(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	shotID64, err := strconv.ParseUint(c.Param("shotID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shot id"})
		return
	}
	var req updateShotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	shot, err := h.service.UpdateShot(c.Request.Context(), projectID, userID, uint(shotID64), services.UpdateShotInput{
		TrackIndex:           req.TrackIndex,
		Chapter:              req.Chapter,
		Title:                req.Title,
		Description:          req.Description,
		CameraLanguage:       req.CameraLanguage,
		EmotionTone:          req.EmotionTone,
		Duration:             req.Duration,
		AspectRatio:          req.AspectRatio,
		Prompt:               req.Prompt,
		NegativePrompt:       req.NegativePrompt,
		ReferenceImageURL:    req.ReferenceImageURL,
		TimelineStartMs:      req.TimelineStartMs,
		TimelineDurationMs:   req.TimelineDurationMs,
		TransitionType:       req.TransitionType,
		TransitionDurationMs: req.TransitionDurationMs,
		Locked:               req.Locked,
		Status:               req.Status,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shot)
}

func (h *StoryboardHandler) GenerateShotClip(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	shotID64, err := strconv.ParseUint(c.Param("shotID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shot id"})
		return
	}
	var req shotActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	shot, err := h.service.GenerateShotClip(c.Request.Context(), projectID, userID, uint(shotID64), services.GenerateShotClipInput{
		Provider:          services.VideoProvider(req.Provider),
		Model:             req.Model,
		Duration:          req.Duration,
		AspectRatio:       req.AspectRatio,
		Prompt:            req.Prompt,
		NegativePrompt:    req.NegativePrompt,
		ReferenceImageURL: req.ReferenceImageURL,
		WorkflowPath:      req.WorkflowPath,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shot)
}

func (h *StoryboardHandler) RegenerateShotClip(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	shotID64, err := strconv.ParseUint(c.Param("shotID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shot id"})
		return
	}
	var req shotActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	shot, err := h.service.RegenerateShotClip(c.Request.Context(), projectID, userID, uint(shotID64), services.GenerateShotClipInput{
		Provider:          services.VideoProvider(req.Provider),
		Model:             req.Model,
		Duration:          req.Duration,
		AspectRatio:       req.AspectRatio,
		Prompt:            req.Prompt,
		NegativePrompt:    req.NegativePrompt,
		ReferenceImageURL: req.ReferenceImageURL,
		WorkflowPath:      req.WorkflowPath,
	}, req.VersionNote)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shot)
}

func (h *StoryboardHandler) RenderDirectorCut(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	task, err := h.service.RenderDirectorCut(c.Request.Context(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *StoryboardHandler) StreamDirectorCut(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	exportID := c.Param("exportID")
	path, err := h.service.ResolveDirectorCutPath(c.Request.Context(), projectID, userID, exportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.File(path)
}

func (h *StoryboardHandler) CheckDirectorCutGate(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	exportID := c.Param("exportID")
	passed, reason, score, threshold, err := h.service.CheckDirectorCutGate(c.Request.Context(), projectID, userID, exportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if !passed {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"status":    "blocked",
			"reason":    reason,
			"score":     score,
			"threshold": threshold,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":    "passed",
		"reason":    "",
		"score":     score,
		"threshold": threshold,
	})
}

func (h *StoryboardHandler) PublishDirectorCut(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	exportID := c.Param("exportID")
	var req directorPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := h.service.PublishDirectorCut(c.Request.Context(), projectID, userID, exportID, req.Platform)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "below threshold") {
			c.JSON(http.StatusPreconditionFailed, gin.H{"status": "blocked", "reason": msg})
			return
		}
		if strings.Contains(msg, "not bound") {
			c.JSON(http.StatusPreconditionRequired, gin.H{"status": "blocked", "reason": msg})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   "published",
		"platform": req.Platform,
		"video_id": task.VideoID,
		"task_id":  task.ID,
	})
}

func (h *StoryboardHandler) ListDirectorPublishHistory(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	exportID := strings.TrimSpace(c.Query("export_id"))
	records, err := h.service.ListDirectorPublishHistory(c.Request.Context(), projectID, userID, exportID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": len(records), "data": records})
}

func (h *StoryboardHandler) RetryDirectorPublish(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	recordID64, err := strconv.ParseUint(c.Param("recordID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid record id"})
		return
	}
	var req directorPublishRetryRequest
	_ = c.ShouldBindJSON(&req)
	task, record, retryErr := h.service.RetryDirectorPublish(c.Request.Context(), projectID, userID, uint(recordID64))
	if retryErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": retryErr.Error(), "reason": req.Reason})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":         "retried",
		"reason":         req.Reason,
		"video_id":       task.VideoID,
		"task_id":        task.ID,
		"publish_record": record,
	})
}

func (h *StoryboardHandler) ListDirectorTemplates(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	templates, err := h.service.ListDirectorTemplates(c.Request.Context(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": len(templates), "data": templates})
}

func (h *StoryboardHandler) CreateDirectorTemplate(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	var req directorTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, err := h.service.CreateDirectorTemplate(c.Request.Context(), projectID, userID, services.DirectorTemplateInput{
		Name:                 req.Name,
		Slug:                 req.Slug,
		SampleFrameURL:       req.SampleFrameURL,
		SampleVideoURL:       req.SampleVideoURL,
		PromptPrefix:         req.PromptPrefix,
		CameraLanguage:       req.CameraLanguage,
		EmotionTone:          req.EmotionTone,
		TransitionType:       req.TransitionType,
		TransitionDurationMs: req.TransitionDurationMs,
		GenreKeywords:        req.GenreKeywords,
		WeightNarrative:      req.WeightNarrative,
		WeightVisual:         req.WeightVisual,
		WeightEmotion:        req.WeightEmotion,
		WeightRhythm:         req.WeightRhythm,
		WeightContinuity:     req.WeightContinuity,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, template)
}

func (h *StoryboardHandler) UpdateDirectorTemplate(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	templateID64, err := strconv.ParseUint(c.Param("templateID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}
	var req directorTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, updateErr := h.service.UpdateDirectorTemplate(c.Request.Context(), projectID, userID, uint(templateID64), services.DirectorTemplateInput{
		Name:                 req.Name,
		Slug:                 req.Slug,
		SampleFrameURL:       req.SampleFrameURL,
		SampleVideoURL:       req.SampleVideoURL,
		PromptPrefix:         req.PromptPrefix,
		CameraLanguage:       req.CameraLanguage,
		EmotionTone:          req.EmotionTone,
		TransitionType:       req.TransitionType,
		TransitionDurationMs: req.TransitionDurationMs,
		GenreKeywords:        req.GenreKeywords,
		WeightNarrative:      req.WeightNarrative,
		WeightVisual:         req.WeightVisual,
		WeightEmotion:        req.WeightEmotion,
		WeightRhythm:         req.WeightRhythm,
		WeightContinuity:     req.WeightContinuity,
	})
	if updateErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": updateErr.Error()})
		return
	}
	c.JSON(http.StatusOK, template)
}

func (h *StoryboardHandler) DeleteDirectorTemplate(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	templateID64, err := strconv.ParseUint(c.Param("templateID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}
	if err := h.service.DeleteDirectorTemplate(c.Request.Context(), projectID, userID, uint(templateID64)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *StoryboardHandler) AutoDirectorStrategy(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	var req autoDirectorStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.service.AutoDirectorStrategy(c.Request.Context(), projectID, userID, services.AutoDirectorStrategyInput{
		Genre:       req.Genre,
		TemplateID:  req.TemplateID,
		Apply:       req.Apply,
		TunePercent: req.TunePercent,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *StoryboardHandler) CompareDirectorAB(c *gin.Context) {
	projectID, userID, ok := h.parseProjectAndUser(c)
	if !ok {
		return
	}
	var req directorABCompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.service.CompareDirectorAB(c.Request.Context(), projectID, userID, services.DirectorABCompareInput{
		TemplateAID:   req.TemplateAID,
		TemplateBID:   req.TemplateBID,
		Genre:         req.Genre,
		ApplyBest:     req.ApplyBest,
		TunePercent:   req.TunePercent,
		RenderBestCut: req.RenderBestCut,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

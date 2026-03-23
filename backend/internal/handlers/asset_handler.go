package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/richard9219/3kstory/internal/models"
	"github.com/richard9219/3kstory/internal/services"
)

type AssetHandler struct {
	service *services.AssetService
}

func NewAssetHandler(service *services.AssetService) *AssetHandler {
	return &AssetHandler{service: service}
}

func parseOptionalProjectID(c *gin.Context) (*uint, error) {
	pid := c.Query("project_id")
	if pid == "" {
		return nil, nil
	}
	id64, err := strconv.ParseUint(pid, 10, 32)
	if err != nil {
		return nil, err
	}
	id := uint(id64)
	return &id, nil
}

type createRoleAssetRequest struct {
	ProjectID    *uint  `json:"project_id"`
	Name         string `json:"name" binding:"required"`
	RoleType     string `json:"role_type"`
	Description  string `json:"description"`
	AvatarURL    string `json:"avatar_url"`
	VoicePreset  string `json:"voice_preset"`
	StylePrompt  string `json:"style_prompt"`
	NegativeHint string `json:"negative_hint"`
	Tags         string `json:"tags"`
}

type updateRoleAssetRequest struct {
	ProjectID    *uint  `json:"project_id"`
	Name         string `json:"name"`
	RoleType     string `json:"role_type"`
	Description  string `json:"description"`
	AvatarURL    string `json:"avatar_url"`
	VoicePreset  string `json:"voice_preset"`
	StylePrompt  string `json:"style_prompt"`
	NegativeHint string `json:"negative_hint"`
	Tags         string `json:"tags"`
}

func (h *AssetHandler) CreateRoleAsset(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req createRoleAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	asset := &models.RoleAsset{
		UserID:       userID,
		ProjectID:    req.ProjectID,
		Name:         req.Name,
		RoleType:     req.RoleType,
		Description:  req.Description,
		AvatarURL:    req.AvatarURL,
		VoicePreset:  req.VoicePreset,
		StylePrompt:  req.StylePrompt,
		NegativeHint: req.NegativeHint,
		Tags:         req.Tags,
	}
	if err := h.service.CreateRoleAsset(c.Request.Context(), asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, asset)
}

func (h *AssetHandler) ListRoleAssets(c *gin.Context) {
	userID := c.GetUint("user_id")
	projectID, err := parseOptionalProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}
	keyword := strings.TrimSpace(c.Query("q"))
	tags := strings.Split(c.Query("tags"), ",")
	assets, err := h.service.ListRoleAssets(c.Request.Context(), userID, projectID, keyword, tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list role assets"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": assets, "total": len(assets)})
}

func (h *AssetHandler) UpdateRoleAsset(c *gin.Context) {
	userID := c.GetUint("user_id")
	assetID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role asset id"})
		return
	}

	var req updateRoleAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.UpdateRoleAsset(c.Request.Context(), userID, uint(assetID64), &models.RoleAsset{
		ProjectID:    req.ProjectID,
		Name:         req.Name,
		RoleType:     req.RoleType,
		Description:  req.Description,
		AvatarURL:    req.AvatarURL,
		VoicePreset:  req.VoicePreset,
		StylePrompt:  req.StylePrompt,
		NegativeHint: req.NegativeHint,
		Tags:         req.Tags,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *AssetHandler) DeleteRoleAsset(c *gin.Context) {
	userID := c.GetUint("user_id")
	assetID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role asset id"})
		return
	}

	if err := h.service.DeleteRoleAsset(c.Request.Context(), userID, uint(assetID64)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

type createPromptTemplateRequest struct {
	ProjectID    *uint  `json:"project_id"`
	Name         string `json:"name" binding:"required"`
	TemplateType string `json:"template_type" binding:"required"`
	ProviderType string `json:"provider_type"`
	Content      string `json:"content" binding:"required"`
	Variables    string `json:"variables"`
	Tags         string `json:"tags"`
}

type updatePromptTemplateRequest struct {
	ProjectID    *uint  `json:"project_id"`
	Name         string `json:"name"`
	TemplateType string `json:"template_type"`
	ProviderType string `json:"provider_type"`
	Content      string `json:"content"`
	Variables    string `json:"variables"`
	Tags         string `json:"tags"`
}

func (h *AssetHandler) CreatePromptTemplate(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req createPromptTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tpl := &models.PromptTemplate{
		UserID:       userID,
		ProjectID:    req.ProjectID,
		Name:         req.Name,
		TemplateType: req.TemplateType,
		ProviderType: req.ProviderType,
		Content:      req.Content,
		Variables:    req.Variables,
		Tags:         req.Tags,
	}
	if err := h.service.CreatePromptTemplate(c.Request.Context(), tpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tpl)
}

func (h *AssetHandler) ListPromptTemplates(c *gin.Context) {
	userID := c.GetUint("user_id")
	projectID, err := parseOptionalProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}
	typeFilter := c.Query("template_type")
	keyword := strings.TrimSpace(c.Query("q"))
	tags := strings.Split(c.Query("tags"), ",")
	list, err := h.service.ListPromptTemplates(c.Request.Context(), userID, projectID, typeFilter, keyword, tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list prompt templates"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": len(list)})
}

func (h *AssetHandler) UpdatePromptTemplate(c *gin.Context) {
	userID := c.GetUint("user_id")
	templateID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}

	var req updatePromptTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.UpdatePromptTemplate(c.Request.Context(), userID, uint(templateID64), &models.PromptTemplate{
		ProjectID:    req.ProjectID,
		Name:         req.Name,
		TemplateType: req.TemplateType,
		ProviderType: req.ProviderType,
		Content:      req.Content,
		Variables:    req.Variables,
		Tags:         req.Tags,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *AssetHandler) DeletePromptTemplate(c *gin.Context) {
	userID := c.GetUint("user_id")
	templateID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}

	if err := h.service.DeletePromptTemplate(c.Request.Context(), userID, uint(templateID64)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

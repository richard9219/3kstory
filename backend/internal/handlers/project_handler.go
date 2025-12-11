package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/richard9219/3kstory/internal/models"
	"github.com/richard9219/3kstory/internal/services"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	service *services.ProjectService
	db      *gorm.DB
}

func NewProjectHandler(service *services.ProjectService, db *gorm.DB) *ProjectHandler {
	return &ProjectHandler{
		service: service,
		db:      db,
	}
}

type CreateProjectRequest struct {
	Title  string `json:"title"`
	Prompt string `json:"prompt" binding:"required"`
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.service.CreateProject(userID, req.Prompt, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	userID := c.GetUint("user_id")

	var projects []models.Project
	if err := h.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	projectID := c.Param("id")
	userID := c.GetUint("user_id")

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	projectID := c.Param("id")
	userID := c.GetUint("user_id")

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		project.Title = req.Title
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.Status != "" {
		project.Status = req.Status
	}

	if err := h.db.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	projectID := c.Param("id")
	userID := c.GetUint("user_id")

	result := h.db.Where("id = ? AND user_id = ?", projectID, userID).Delete(&models.Project{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

func (h *ProjectHandler) GetScenes(c *gin.Context) {
	projectID := c.Param("id")
	userID := c.GetUint("user_id")

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var scenes []models.Scene
	if err := h.db.Where("project_id = ?", projectID).Order("scene_number ASC").Find(&scenes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scenes"})
		return
	}

	c.JSON(http.StatusOK, scenes)
}

func (h *ProjectHandler) GenerateScenes(c *gin.Context) {
	projectID := c.Param("id")
	userID := c.GetUint("user_id")

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	go func() {
		h.service.GenerateScenes(c.Request.Context(), project.ID)
	}()

	c.JSON(http.StatusAccepted, gin.H{"message": "Scene generation started"})
}

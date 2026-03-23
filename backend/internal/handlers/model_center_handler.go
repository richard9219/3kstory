package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/richard9219/3kstory/internal/services"
)

type ModelCenterHandler struct {
	service *services.ModelCenterService
}

func NewModelCenterHandler(service *services.ModelCenterService) *ModelCenterHandler {
	return &ModelCenterHandler{service: service}
}

// GetOverview 获取模型中心配置和健康状态。
// GET /api/v1/model-center/overview
func (h *ModelCenterHandler) GetOverview(c *gin.Context) {
	overview := h.service.GetOverview(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"data": overview})
}

// TriggerProbe 触发一次主动探活。
// POST /api/v1/model-center/probe
func (h *ModelCenterHandler) TriggerProbe(c *gin.Context) {
	overview := h.service.TriggerProbe(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"data": overview})
}

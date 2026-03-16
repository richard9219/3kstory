package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/richard9219/3kstory/internal/services"
)

type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
	platformService  *services.PlatformService
}

func NewAnalyticsHandler(analyticsService *services.AnalyticsService, platformService *services.PlatformService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		platformService:  platformService,
	}
}

// GetSummary 获取运营看板总览（里程碑 1 先返回内部统计 + 绑定平台列表）
// GET /api/v1/analytics/summary
func (h *AnalyticsHandler) GetSummary(c *gin.Context) {
	userID := c.GetUint("user_id")
	summary, err := h.analyticsService.GetSummary(userID, h.platformService.ConfiguredPlatforms())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query summary"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// ListVideos 获取近期生成/发布记录
// GET /api/v1/analytics/videos?limit=20
func (h *AnalyticsHandler) ListVideos(c *gin.Context) {
	userID := c.GetUint("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	list, err := h.analyticsService.ListRecentVideos(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query videos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(list),
		"data":  list,
	})
}

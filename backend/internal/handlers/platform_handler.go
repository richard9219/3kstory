package handlers

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/richard9219/3kstory/internal/models"
	"github.com/richard9219/3kstory/internal/services"
)

type PlatformHandler struct {
	platformService *services.PlatformService
}

func NewPlatformHandler(platformService *services.PlatformService) *PlatformHandler {
	return &PlatformHandler{platformService: platformService}
}

// ListPlatforms 列出当前用户已绑定的平台账号
// GET /api/v1/platforms
func (h *PlatformHandler) ListPlatforms(c *gin.Context) {
	userID, _ := c.Get("user_id")
	list, err := h.platformService.ListByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// ConfiguredPlatforms 返回后端已配置 OAuth 的平台列表（未配置的平台前端可显示「暂未开放」等）
// GET /api/v1/platforms/configured
func (h *PlatformHandler) ConfiguredPlatforms(c *gin.Context) {
	list := h.platformService.ConfiguredPlatforms()
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// GetAuthorizeURL 获取指定平台的授权链接，前端跳转
// GET /api/v1/platforms/:platform/authorize
func (h *PlatformHandler) GetAuthorizeURL(c *gin.Context) {
	platform := c.Param("platform")
	if !isValidPlatform(platform) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid platform"})
		return
	}
	if platform == models.PlatformXiaohongshu {
		c.JSON(http.StatusBadRequest, gin.H{"error": "xiaohongshu is not available in milestone 1 yet"})
		return
	}
	userID, _ := c.Get("user_id")
	url, err := h.platformService.GetAuthorizeURL(platform, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"authorize_url": url})
}

// OAuthCallback 各平台 OAuth 回调（无需登录）；用 code+state 换 token 后重定向到前端
// GET /api/v1/platforms/:platform/callback?code=xxx&state=xxx
func (h *PlatformHandler) OAuthCallback(c *gin.Context) {
	platform := c.Param("platform")
	if !isValidPlatform(platform) {
		c.Redirect(http.StatusFound, h.platformService.PlatformBoundURL(false, "invalid_platform"))
		return
	}
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.Redirect(http.StatusFound, h.platformService.PlatformBoundURL(false, "missing_code_or_state"))
		return
	}
	redirectSuccess, redirectFail, err := h.platformService.ExchangeCode(platform, code, state)
	if err != nil {
		c.Redirect(http.StatusFound, redirectFail+"&err="+url.QueryEscape(err.Error()))
		return
	}
	c.Redirect(http.StatusFound, redirectSuccess)
}

// Disconnect 解绑指定平台
// DELETE /api/v1/platforms/:platform
func (h *PlatformHandler) Disconnect(c *gin.Context) {
	platform := c.Param("platform")
	if !isValidPlatform(platform) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid platform"})
		return
	}
	userID, _ := c.Get("user_id")
	if err := h.platformService.Disconnect(userID.(uint), platform); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "disconnected"})
}

func isValidPlatform(p string) bool {
	switch p {
	case models.PlatformDouyin, models.PlatformXiaohongshu, models.PlatformBilibili, models.PlatformWeibo:
		return true
	default:
		return false
	}
}

package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/richard9219/3kstory/internal/config"
	"github.com/richard9219/3kstory/internal/models"
	"gorm.io/gorm"
)

const oauthStatePrefix = "oauth_state:"
const oauthStateTTL = 10 * time.Minute

// PlatformService 第三方平台 OAuth 与账号绑定
type PlatformService struct {
	db  *gorm.DB
	cfg *config.Config
	rdb RedisStore
}

type RedisStore interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
}

func NewPlatformService(db *gorm.DB, cfg *config.Config, rdb RedisStore) *PlatformService {
	return &PlatformService{db: db, cfg: cfg, rdb: rdb}
}

// GetOAuthConfig 返回指定平台的 OAuth 配置（用于生成授权链接）
func (s *PlatformService) GetOAuthConfig(platform string) (clientID, redirectURI, scope string, ok bool) {
	base := strings.TrimSuffix(s.cfg.BaseURL, "/")
	switch platform {
	case models.PlatformDouyin:
		item := s.cfg.Platform.Douyin
		if item.RedirectURI != "" {
			return item.ClientID, item.RedirectURI, item.Scope, true
		}
		return item.ClientID, base + "/api/v1/platforms/douyin/callback", item.Scope, item.ClientID != ""
	case models.PlatformXiaohongshu:
		item := s.cfg.Platform.Xiaohongshu
		if item.RedirectURI != "" {
			return item.ClientID, item.RedirectURI, item.Scope, true
		}
		return item.ClientID, base + "/api/v1/platforms/xiaohongshu/callback", item.Scope, item.ClientID != ""
	case models.PlatformBilibili:
		item := s.cfg.Platform.Bilibili
		if item.RedirectURI != "" {
			return item.ClientID, item.RedirectURI, item.Scope, true
		}
		return item.ClientID, base + "/api/v1/platforms/bilibili/callback", item.Scope, item.ClientID != ""
	case models.PlatformWeibo:
		item := s.cfg.Platform.Weibo
		if item.RedirectURI != "" {
			return item.ClientID, item.RedirectURI, item.Scope, true
		}
		return item.ClientID, base + "/api/v1/platforms/weibo/callback", item.Scope, item.ClientID != ""
	default:
		return "", "", "", false
	}
}

// GetAuthorizeURL 生成授权跳转 URL，并将 state 与 userID 存入 Redis
func (s *PlatformService) GetAuthorizeURL(platform string, userID uint) (authorizeURL string, err error) {
	clientID, redirectURI, scope, ok := s.GetOAuthConfig(platform)
	if !ok {
		return "", fmt.Errorf("platform %s not configured", platform)
	}
	state := genOAuthState()
	key := oauthStatePrefix + state
	if err = s.rdb.Set(context.Background(), key, userID, oauthStateTTL); err != nil {
		return "", err
	}
	var raw string
	switch platform {
	case models.PlatformDouyin:
		// 抖音: client_key, response_type=code, scope, redirect_uri, state
		raw = "https://open.douyin.com/platform/oauth/connect/?client_key=%s&response_type=code&scope=%s&redirect_uri=%s&state=%s"
		authorizeURL = fmt.Sprintf(raw, url.QueryEscape(clientID), url.QueryEscape(scope), url.QueryEscape(redirectURI), url.QueryEscape(state))
	case models.PlatformXiaohongshu:
		// 小红书: appId, redirectUri, state
		raw = "https://ark.xiaohongshu.com/ark/authorization?appId=%s&redirectUri=%s&state=%s"
		authorizeURL = fmt.Sprintf(raw, url.QueryEscape(clientID), url.QueryEscape(redirectURI), url.QueryEscape(state))
	case models.PlatformBilibili:
		raw = "https://passport.bilibili.com/oauth2/authorize?client_id=%s&response_type=code&redirect_uri=%s&state=%s"
		if scope != "" {
			raw = "https://passport.bilibili.com/oauth2/authorize?client_id=%s&response_type=code&redirect_uri=%s&state=%s&scope=%s"
			authorizeURL = fmt.Sprintf(raw, url.QueryEscape(clientID), url.QueryEscape(redirectURI), url.QueryEscape(state), url.QueryEscape(scope))
		} else {
			authorizeURL = fmt.Sprintf(raw, url.QueryEscape(clientID), url.QueryEscape(redirectURI), url.QueryEscape(state))
		}
	case models.PlatformWeibo:
		raw = "https://api.weibo.com/oauth2/authorize?client_id=%s&response_type=code&redirect_uri=%s&state=%s"
		if scope != "" {
			raw = "https://api.weibo.com/oauth2/authorize?client_id=%s&response_type=code&redirect_uri=%s&state=%s&scope=%s"
			authorizeURL = fmt.Sprintf(raw, url.QueryEscape(clientID), url.QueryEscape(redirectURI), url.QueryEscape(state), url.QueryEscape(scope))
		} else {
			authorizeURL = fmt.Sprintf(raw, url.QueryEscape(clientID), url.QueryEscape(redirectURI), url.QueryEscape(state))
		}
	default:
		return "", fmt.Errorf("unsupported platform: %s", platform)
	}
	return authorizeURL, nil
}

// ExchangeCode 用 code 换 token 并写入 platform_accounts；返回前端跳转 URL（成功/失败）
func (s *PlatformService) ExchangeCode(platform string, code, state string) (redirectSuccess, redirectFail string, err error) {
	frontBase := strings.TrimSuffix(s.cfg.FrontendURL, "/")
	if frontBase == "" {
		frontBase = strings.TrimSuffix(s.cfg.BaseURL, "/")
	}
	redirectSuccess = frontBase + "/platform-bound?ok=1"
	redirectFail = frontBase + "/platform-bound?ok=0"
	key := oauthStatePrefix + state
	userIDStr, err := s.rdb.Get(context.Background(), key)
	if err != nil || userIDStr == "" {
		return redirectSuccess, redirectFail, fmt.Errorf("invalid or expired state")
	}
	var userID uint
	if _, e := fmt.Sscanf(userIDStr, "%d", &userID); e != nil {
		return redirectSuccess, redirectFail, fmt.Errorf("invalid state payload")
	}
	_ = s.rdb.Del(context.Background(), key)

	var acc *models.PlatformAccount
	switch platform {
	case models.PlatformBilibili:
		acc, err = s.exchangeBilibili(code, userID)
	case models.PlatformWeibo:
		acc, err = s.exchangeWeibo(code, userID)
	case models.PlatformDouyin:
		acc, err = s.exchangeDouyin(code, userID)
	case models.PlatformXiaohongshu:
		acc, err = s.exchangeXiaohongshu(code, userID)
	default:
		err = fmt.Errorf("unsupported platform: %s", platform)
	}
	if err != nil {
		return redirectSuccess, redirectFail, err
	}
	if err = s.upsertAccount(acc); err != nil {
		return redirectSuccess, redirectFail, err
	}
	return redirectSuccess, redirectFail, nil
}

func (s *PlatformService) upsertAccount(acc *models.PlatformAccount) error {
	var existing models.PlatformAccount
	err := s.db.Where("user_id = ? AND platform = ?", acc.UserID, acc.Platform).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return s.db.Create(acc).Error
	}
	if err != nil {
		return err
	}
	acc.ID = existing.ID
	return s.db.Model(&existing).Updates(map[string]interface{}{
		"open_id":       acc.OpenID,
		"union_id":      acc.UnionID,
		"access_token":  acc.AccessToken,
		"refresh_token": acc.RefreshToken,
		"expires_at":    acc.ExpiresAt,
		"nickname":      acc.Nickname,
		"avatar_url":    acc.AvatarURL,
		"extra":         acc.Extra,
		"updated_at":    time.Now(),
	}).Error
}

func (s *PlatformService) exchangeBilibili(code string, userID uint) (*models.PlatformAccount, error) {
	item := s.cfg.Platform.Bilibili
	if item.ClientID == "" || item.ClientSecret == "" {
		return nil, fmt.Errorf("bilibili oauth not configured")
	}
	redirectURI := item.RedirectURI
	if redirectURI == "" {
		redirectURI = strings.TrimSuffix(s.cfg.BaseURL, "/") + "/api/v1/platforms/bilibili/callback"
	}
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", item.ClientID)
	form.Set("client_secret", item.ClientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", redirectURI)
	req, _ := http.NewRequest(http.MethodPost, "https://api.bilibili.com/x/account-oauth2/v1/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int64  `json:"expires_in"`
			Mid          int64  `json:"mid"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Code != 0 {
		return nil, fmt.Errorf("bilibili token: %s", out.Message)
	}
	exp := time.Now().Add(time.Duration(out.Data.ExpiresIn) * time.Second)
	acc := &models.PlatformAccount{
		UserID:       userID,
		Platform:     models.PlatformBilibili,
		OpenID:       fmt.Sprintf("%d", out.Data.Mid),
		AccessToken:  out.Data.AccessToken,
		RefreshToken: out.Data.RefreshToken,
		ExpiresAt:    &exp,
	}
	acc.Nickname = "B站用户"
	if name, avatar, err := s.fetchBilibiliUserInfo(out.Data.AccessToken, out.Data.Mid); err == nil {
		if name != "" {
			acc.Nickname = name
		}
		if avatar != "" {
			acc.AvatarURL = avatar
		}
	}
	return acc, nil
}

func (s *PlatformService) exchangeWeibo(code string, userID uint) (*models.PlatformAccount, error) {
	item := s.cfg.Platform.Weibo
	if item.ClientID == "" || item.ClientSecret == "" {
		return nil, fmt.Errorf("weibo oauth not configured")
	}
	redirectURI := item.RedirectURI
	if redirectURI == "" {
		redirectURI = strings.TrimSuffix(s.cfg.BaseURL, "/") + "/api/v1/platforms/weibo/callback"
	}
	form := url.Values{}
	form.Set("client_id", item.ClientID)
	form.Set("client_secret", item.ClientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	req, _ := http.NewRequest(http.MethodPost, "https://api.weibo.com/oauth2/access_token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		UID         string `json:"uid"`
		Error       string `json:"error"`
		ErrorCode   int    `json:"error_code"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Error != "" {
		return nil, fmt.Errorf("weibo token: %s (code %d)", out.Error, out.ErrorCode)
	}
	exp := time.Now().Add(time.Duration(out.ExpiresIn) * time.Second)
	acc := &models.PlatformAccount{
		UserID:      userID,
		Platform:    models.PlatformWeibo,
		OpenID:      out.UID,
		AccessToken: out.AccessToken,
		ExpiresAt:   &exp,
	}
	acc.Nickname = "微博用户"
	if name, avatar, err := s.fetchWeiboUserInfo(out.AccessToken, out.UID); err == nil {
		if name != "" {
			acc.Nickname = name
		}
		if avatar != "" {
			acc.AvatarURL = avatar
		}
	}
	return acc, nil
}

func (s *PlatformService) exchangeDouyin(code string, userID uint) (*models.PlatformAccount, error) {
	item := s.cfg.Platform.Douyin
	if item.ClientID == "" || item.ClientSecret == "" {
		return nil, fmt.Errorf("douyin oauth not configured")
	}
	redirectURI := item.RedirectURI
	if redirectURI == "" {
		redirectURI = strings.TrimSuffix(s.cfg.BaseURL, "/") + "/api/v1/platforms/douyin/callback"
	}
	form := url.Values{}
	form.Set("client_key", item.ClientID)
	form.Set("client_secret", item.ClientSecret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", redirectURI)
	req, _ := http.NewRequest(http.MethodPost, "https://open.douyin.com/oauth/access_token/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Data struct {
			AccessToken   string `json:"access_token"`
			RefreshToken  string `json:"refresh_token"`
			ExpiresIn     int64  `json:"expires_in"`
			OpenID        string `json:"open_id"`
			UnionID       string `json:"union_id"`
			ErrorCode     int64  `json:"error_code"`
			Description   string `json:"description"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Data.ErrorCode != 0 {
		return nil, fmt.Errorf("douyin token: %s", out.Data.Description)
	}
	exp := time.Now().Add(time.Duration(out.Data.ExpiresIn) * time.Second)
	acc := &models.PlatformAccount{
		UserID:       userID,
		Platform:     models.PlatformDouyin,
		OpenID:       out.Data.OpenID,
		UnionID:      out.Data.UnionID,
		AccessToken:  out.Data.AccessToken,
		RefreshToken: out.Data.RefreshToken,
		ExpiresAt:    &exp,
	}
	acc.Nickname = "抖音用户"
	if name, avatar, err := s.fetchDouyinUserInfo(out.Data.AccessToken, out.Data.OpenID); err == nil {
		if name != "" {
			acc.Nickname = name
		}
		if avatar != "" {
			acc.AvatarURL = avatar
		}
	}
	return acc, nil
}

func (s *PlatformService) exchangeXiaohongshu(code string, userID uint) (*models.PlatformAccount, error) {
	// 小红书使用 Ark 开放平台（https://ark.xiaohongshu.com），非标准 OAuth2：
	// 授权后需调用 ark/open_api 通用接口，请求需 sign、timestamp 等签名参数，且不同能力有不同 path。
	// 接入前需在「小红书开放平台-应用管理」创建应用并获取 app_id、app_secret，配置授权回调 redirectUri。
	// 实现时参考：https://xiaohongshu.apifox.cn 或 小红书开放平台文档。
	item := s.cfg.Platform.Xiaohongshu
	if item.ClientID == "" || item.ClientSecret == "" {
		return nil, fmt.Errorf("xiaohongshu oauth not configured")
	}
	_ = code
	_ = item
	return nil, fmt.Errorf("小红书绑定暂未开放，请先使用抖音、B站或微博")
}

// ConfiguredPlatforms 返回已在环境变量中配置了 OAuth 的平台列表（前端可据此展示「可绑定」状态）
func (s *PlatformService) ConfiguredPlatforms() []string {
	var out []string
	if s.cfg.Platform.Douyin.ClientID != "" && s.cfg.Platform.Douyin.ClientSecret != "" {
		out = append(out, models.PlatformDouyin)
	}
	if s.cfg.Platform.Xiaohongshu.ClientID != "" && s.cfg.Platform.Xiaohongshu.ClientSecret != "" {
		out = append(out, models.PlatformXiaohongshu)
	}
	if s.cfg.Platform.Bilibili.ClientID != "" && s.cfg.Platform.Bilibili.ClientSecret != "" {
		out = append(out, models.PlatformBilibili)
	}
	if s.cfg.Platform.Weibo.ClientID != "" && s.cfg.Platform.Weibo.ClientSecret != "" {
		out = append(out, models.PlatformWeibo)
	}
	return out
}

// ListByUserID 返回用户已绑定的平台账号列表；返回前对即将过期的 token 尝试刷新（B站/抖音）
func (s *PlatformService) ListByUserID(userID uint) ([]*models.PlatformAccountResponse, error) {
	var list []models.PlatformAccount
	if err := s.db.Where("user_id = ?", userID).Find(&list).Error; err != nil {
		return nil, err
	}
	for i := range list {
		_ = s.RefreshTokenIfNeeded(&list[i])
	}
	out := make([]*models.PlatformAccountResponse, 0, len(list))
	for i := range list {
		out = append(out, list[i].ToResponse())
	}
	return out, nil
}

// Disconnect 解绑指定平台
func (s *PlatformService) Disconnect(userID uint, platform string) error {
	return s.db.Where("user_id = ? AND platform = ?", userID, platform).Delete(&models.PlatformAccount{}).Error
}

// RefreshTokenIfNeeded 在 access_token 即将过期时刷新（B站、抖音支持）；过期或无 refresh_token 则不刷新
func (s *PlatformService) RefreshTokenIfNeeded(acc *models.PlatformAccount) error {
	if acc.ExpiresAt == nil || acc.RefreshToken == "" {
		return nil
	}
	// 提前 5 分钟视为即将过期
	if time.Until(*acc.ExpiresAt) > 5*time.Minute {
		return nil
	}
	switch acc.Platform {
	case models.PlatformBilibili:
		return s.refreshBilibiliToken(acc)
	case models.PlatformDouyin:
		return s.refreshDouyinToken(acc)
	default:
		return nil
	}
}

func (s *PlatformService) refreshBilibiliToken(acc *models.PlatformAccount) error {
	item := s.cfg.Platform.Bilibili
	if item.ClientID == "" || item.ClientSecret == "" {
		return nil
	}
	form := url.Values{}
	form.Set("client_id", item.ClientID)
	form.Set("client_secret", item.ClientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", acc.RefreshToken)
	req, _ := http.NewRequest(http.MethodPost, "https://api.bilibili.com/x/account-oauth2/v1/refresh_token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var out struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int64  `json:"expires_in"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	if out.Code != 0 {
		return fmt.Errorf("bilibili refresh: %s", out.Message)
	}
	exp := time.Now().Add(time.Duration(out.Data.ExpiresIn) * time.Second)
	return s.db.Model(acc).Updates(map[string]interface{}{
		"access_token":  out.Data.AccessToken,
		"refresh_token": out.Data.RefreshToken,
		"expires_at":    exp,
		"updated_at":    time.Now(),
	}).Error
}

func (s *PlatformService) refreshDouyinToken(acc *models.PlatformAccount) error {
	item := s.cfg.Platform.Douyin
	if item.ClientID == "" || item.ClientSecret == "" {
		return nil
	}
	form := url.Values{}
	form.Set("client_key", item.ClientID)
	form.Set("client_secret", item.ClientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", acc.RefreshToken)
	req, _ := http.NewRequest(http.MethodPost, "https://open.douyin.com/oauth/refresh_token/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var out struct {
		Data struct {
			AccessToken   string `json:"access_token"`
			RefreshToken  string `json:"refresh_token"`
			ExpiresIn     int64  `json:"expires_in"`
			ErrorCode     int64  `json:"error_code"`
			Description   string `json:"description"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	if out.Data.ErrorCode != 0 {
		return fmt.Errorf("douyin refresh: %s", out.Data.Description)
	}
	exp := time.Now().Add(time.Duration(out.Data.ExpiresIn) * time.Second)
	return s.db.Model(acc).Updates(map[string]interface{}{
		"access_token":  out.Data.AccessToken,
		"refresh_token": out.Data.RefreshToken,
		"expires_at":    exp,
		"updated_at":    time.Now(),
	}).Error
}

// fetchWeiboUserInfo 微博 GET /2/users/show.json
func (s *PlatformService) fetchWeiboUserInfo(accessToken, uid string) (nickname, avatar string, err error) {
	u := "https://api.weibo.com/2/users/show.json?access_token=" + url.QueryEscape(accessToken) + "&uid=" + url.QueryEscape(uid)
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	var out struct {
		Name            string `json:"name"`
		ProfileImageURL string `json:"profile_image_url"`
		Error           string `json:"error"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", err
	}
	if out.Error != "" {
		return "", "", fmt.Errorf("%s", out.Error)
	}
	return out.Name, out.ProfileImageURL, nil
}

// fetchDouyinUserInfo 抖音 POST /oauth/userinfo/ 请求头 access-token，Body 为空
func (s *PlatformService) fetchDouyinUserInfo(accessToken, openID string) (nickname, avatar string, err error) {
	req, _ := http.NewRequest(http.MethodPost, "https://open.douyin.com/oauth/userinfo/", nil)
	req.Header.Set("access-token", accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	var out struct {
		Data struct {
			Nickname  string      `json:"nickname"`
			Avatar    string      `json:"avatar"`
			ErrorCode interface{} `json:"error_code"` // 成功为 "0"，失败为数字
		} `json:"data"`
		Extra struct {
			LogID string `json:"logid"`
		} `json:"extra"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", err
	}
	switch v := out.Data.ErrorCode.(type) {
	case string:
		if v != "0" && v != "" {
			return "", "", fmt.Errorf("douyin userinfo error_code: %s", v)
		}
	case float64:
		if v != 0 {
			return "", "", fmt.Errorf("douyin userinfo error_code: %.0f", v)
		}
	}
	return out.Data.Nickname, out.Data.Avatar, nil
}

// fetchBilibiliUserInfo B 站开放平台用户信息（部分接口需签名，此处尝试简单方式）
func (s *PlatformService) fetchBilibiliUserInfo(accessToken string, mid int64) (nickname, avatar string, err error) {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.bilibili.com/x/account-oauth2/v1/info?mid=%d", mid), nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	var out struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Name    string `json:"name"`
			Face    string `json:"face"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", err
	}
	if out.Code != 0 {
		return "", "", fmt.Errorf("bilibili info: %s", out.Message)
	}
	return out.Data.Name, out.Data.Face, nil
}

func genOAuthState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:22]
}

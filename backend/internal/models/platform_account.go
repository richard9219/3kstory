package models

import (
	"time"
)

// Platform 支持的视频平台
const (
	PlatformDouyin       = "douyin"
	PlatformXiaohongshu  = "xiaohongshu"
	PlatformBilibili     = "bilibili"
	PlatformWeibo        = "weibo"
)

// PlatformAccount 用户绑定的第三方平台账号
type PlatformAccount struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"uniqueIndex:idx_user_platform;not null" json:"user_id"`
	Platform     string    `gorm:"uniqueIndex:idx_user_platform;size:32;not null" json:"platform"` // douyin, xiaohongshu, bilibili, weibo
	OpenID       string    `gorm:"size:128" json:"-"`                                              // 平台 open_id
	UnionID      string    `gorm:"size:128" json:"-"`
	AccessToken  string    `gorm:"size:1024;not null" json:"-"`
	RefreshToken string    `gorm:"size:1024" json:"-"`
	ExpiresAt    *time.Time `json:"-"` // access_token 过期时间
	Nickname     string    `gorm:"size:100" json:"nickname"`
	AvatarURL    string    `gorm:"size:500" json:"avatar_url"`
	Extra        string    `gorm:"type:jsonb" json:"-"` // 平台返回的其它信息 JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 表名
func (PlatformAccount) TableName() string {
	return "platform_accounts"
}

// PlatformAccountResponse 对外展示，不含 token
type PlatformAccountResponse struct {
	ID        uint      `json:"id"`
	Platform  string    `json:"platform"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

func (p *PlatformAccount) ToResponse() *PlatformAccountResponse {
	return &PlatformAccountResponse{
		ID:        p.ID,
		Platform:  p.Platform,
		Nickname:  p.Nickname,
		AvatarURL: p.AvatarURL,
		CreatedAt: p.CreatedAt,
	}
}

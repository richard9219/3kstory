package models

import "time"

// RoleAsset 保存项目可复用角色资产信息。
type RoleAsset struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	ProjectID    *uint     `gorm:"index" json:"project_id,omitempty"`
	Name         string    `gorm:"size:120;not null" json:"name"`
	RoleType     string    `gorm:"size:50" json:"role_type"`
	Description  string    `gorm:"type:text" json:"description"`
	AvatarURL    string    `gorm:"size:500" json:"avatar_url"`
	VoicePreset  string    `gorm:"size:100" json:"voice_preset"`
	StylePrompt  string    `gorm:"type:text" json:"style_prompt"`
	NegativeHint string    `gorm:"type:text" json:"negative_hint"`
	Tags         string    `gorm:"size:500" json:"tags"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (RoleAsset) TableName() string {
	return "role_assets"
}

// PromptTemplate 保存提示词模板，用于解说、短剧、电影和广告等任务复用。
type PromptTemplate struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	ProjectID    *uint     `gorm:"index" json:"project_id,omitempty"`
	Name         string    `gorm:"size:120;not null" json:"name"`
	TemplateType string    `gorm:"size:50;index" json:"template_type"`
	ProviderType string    `gorm:"size:50" json:"provider_type"`
	Content      string    `gorm:"type:text;not null" json:"content"`
	Variables    string    `gorm:"type:text" json:"variables"`
	Tags         string    `gorm:"size:500" json:"tags"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (PromptTemplate) TableName() string {
	return "prompt_templates"
}

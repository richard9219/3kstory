package models

import "time"

// DirectorTemplate 导演风格模板，支持可配置权重用于自动导演策略与 A/B 对比。
type DirectorTemplate struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	UserID               uint      `gorm:"not null;index" json:"user_id"`
	ProjectID            uint      `gorm:"not null;index" json:"project_id"`
	Name                 string    `gorm:"size:80;not null" json:"name"`
	Slug                 string    `gorm:"size:80;not null;index" json:"slug"`
	SampleFrameURL       string    `gorm:"size:1000" json:"sample_frame_url"`
	SampleVideoURL       string    `gorm:"size:1000" json:"sample_video_url"`
	PromptPrefix         string    `gorm:"size:500" json:"prompt_prefix"`
	CameraLanguage       string    `gorm:"size:120" json:"camera_language"`
	EmotionTone          string    `gorm:"size:120" json:"emotion_tone"`
	TransitionType       string    `gorm:"size:40;default:cut" json:"transition_type"`
	TransitionDurationMs int       `gorm:"default:200" json:"transition_duration_ms"`
	GenreKeywords        string    `gorm:"size:255" json:"genre_keywords"`
	WeightNarrative      float64   `gorm:"default:0.2" json:"weight_narrative"`
	WeightVisual         float64   `gorm:"default:0.2" json:"weight_visual"`
	WeightEmotion        float64   `gorm:"default:0.2" json:"weight_emotion"`
	WeightRhythm         float64   `gorm:"default:0.2" json:"weight_rhythm"`
	WeightContinuity     float64   `gorm:"default:0.2" json:"weight_continuity"`
	IsBuiltin            bool      `gorm:"default:false" json:"is_builtin"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (DirectorTemplate) TableName() string {
	return "director_templates"
}

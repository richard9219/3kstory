package models

import "time"

// StoryboardShot 保存项目内镜头数据，用于短剧和电影阶段的分镜生产。
type StoryboardShot struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	UserID               uint      `gorm:"not null;index" json:"user_id"`
	ProjectID            uint      `gorm:"not null;index" json:"project_id"`
	SceneID              *uint     `gorm:"index" json:"scene_id,omitempty"`
	TrackIndex           int       `gorm:"default:1;index" json:"track_index"`
	Chapter              string    `gorm:"size:120" json:"chapter"`
	ShotNumber           int       `gorm:"not null;index" json:"shot_number"`
	SortOrder            int       `gorm:"not null;default:0;index" json:"sort_order"`
	Title                string    `gorm:"size:200" json:"title"`
	Description          string    `gorm:"type:text" json:"description"`
	CameraLanguage       string    `gorm:"size:80" json:"camera_language"`
	EmotionTone          string    `gorm:"size:80" json:"emotion_tone"`
	TimelineStartMs      int       `gorm:"default:0" json:"timeline_start_ms"`
	TimelineDurationMs   int       `gorm:"default:5000" json:"timeline_duration_ms"`
	TransitionType       string    `gorm:"size:40;default:cut" json:"transition_type"`
	TransitionDurationMs int       `gorm:"default:0" json:"transition_duration_ms"`
	Duration             int       `gorm:"default:5" json:"duration"`
	AspectRatio          string    `gorm:"size:10;default:16:9" json:"aspect_ratio"`
	Prompt               string    `gorm:"type:text" json:"prompt"`
	NegativePrompt       string    `gorm:"type:text" json:"negative_prompt"`
	ReferenceImageURL    string    `gorm:"size:500" json:"reference_image_url"`
	ClipProvider         string    `gorm:"size:30" json:"clip_provider"`
	ClipVideoID          string    `gorm:"size:100;index" json:"clip_video_id"`
	ClipVideoURL         string    `gorm:"size:1000" json:"clip_video_url"`
	ClipStatus           string    `gorm:"size:20;default:draft;index" json:"clip_status"`
	ClipScore            float64   `gorm:"default:0" json:"clip_score"`
	ClipNotes            string    `gorm:"size:255" json:"clip_notes"`
	Locked               bool      `gorm:"default:false" json:"locked"`
	Status               string    `gorm:"size:20;default:draft;index" json:"status"`
	Version              int       `gorm:"default:1" json:"version"`
	ParentShotID         *uint     `gorm:"index" json:"parent_shot_id,omitempty"`
	RootShotID           *uint     `gorm:"index" json:"root_shot_id,omitempty"`
	VersionNote          string    `gorm:"size:255" json:"version_note"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (StoryboardShot) TableName() string {
	return "storyboard_shots"
}

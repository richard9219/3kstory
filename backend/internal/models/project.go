package models

import (
"database/sql/driver"
"encoding/json"
"time"
)

type Project struct {
ID             uint      `gorm:"primaryKey" json:"id"`
UserID         uint      `gorm:"not null;index" json:"user_id"`
Title          string    `gorm:"size:200;not null" json:"title"`
Description    string    `gorm:"type:text" json:"description"`
Prompt         string    `gorm:"type:text;not null" json:"prompt"`
Genre          string    `gorm:"size:50;index" json:"genre"`
Style          string    `gorm:"size:50" json:"style"`
TargetDuration int       `gorm:"default:30" json:"target_duration"`
CoverURL       string    `gorm:"size:500" json:"cover_url"`
Status         string    `gorm:"size:20;default:draft;index" json:"status"`
ViewCount      int       `gorm:"default:0" json:"view_count"`
LikeCount      int       `gorm:"default:0" json:"like_count"`
CreatedAt      time.Time `json:"created_at"`
UpdatedAt      time.Time `json:"updated_at"`

User   User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
Scenes []Scene `gorm:"foreignKey:ProjectID" json:"scenes,omitempty"`
}

type Scene struct {
ID              uint           `gorm:"primaryKey" json:"id"`
ProjectID       uint           `gorm:"not null;index" json:"project_id"`
SceneNumber     int            `gorm:"not null" json:"scene_number"`
Title           string         `gorm:"size:200" json:"title"`
Description     string         `gorm:"type:text" json:"description"`
Location        string         `gorm:"size:100" json:"location"`
Characters      CharacterArray `gorm:"type:jsonb" json:"characters"`
Dialogue        string         `gorm:"type:text" json:"dialogue"`
ShotType        string         `gorm:"size:50" json:"shot_type"`
Duration        int            `gorm:"default:5" json:"duration"`
MediaType       string         `gorm:"size:20;default:image" json:"media_type"`
MediaURL        string         `gorm:"size:500" json:"media_url"`
PromptForImage  string         `gorm:"type:text" json:"prompt_for_image"`
PromptForVideo  string         `gorm:"type:text" json:"prompt_for_video"`
Status          string         `gorm:"size:20;default:pending;index" json:"status"`
CreatedAt       time.Time      `json:"created_at"`
UpdatedAt       time.Time      `json:"updated_at"`
}

type Character struct {
Name    string `json:"name"`
Emotion string `json:"emotion"`
}

type CharacterArray []Character

func (c CharacterArray) Value() (driver.Value, error) {
return json.Marshal(c)
}

func (c *CharacterArray) Scan(value interface{}) error {
bytes, ok := value.([]byte)
if !ok {
return nil
}
return json.Unmarshal(bytes, c)
}

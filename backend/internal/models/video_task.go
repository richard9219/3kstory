package models

import "time"

// VideoTask 保存视频生成与解说视频任务记录，用于看板统计与近期记录展示。
type VideoTask struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"not null;index" json:"user_id"`
	ProjectID   uint       `gorm:"not null;index" json:"project_id"`
	SceneID     *uint      `gorm:"index" json:"scene_id"`
	TaskType    string     `gorm:"size:50;not null;index" json:"task_type"` // generate_video / narration_video
	Title       string     `gorm:"size:200" json:"title"`
	Provider    string     `gorm:"size:30;index" json:"provider"`
	VideoID     string     `gorm:"size:100;index" json:"video_id"`
	VideoURL    string     `gorm:"size:1000" json:"video_url"`
	Status      string     `gorm:"size:20;default:pending;index" json:"status"`
	InputData   JSONMap    `gorm:"type:jsonb" json:"input_data"`
	OutputData  JSONMap    `gorm:"type:jsonb" json:"output_data"`
	ErrorMsg    string     `gorm:"type:text" json:"error_msg"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (VideoTask) TableName() string {
	return "video_tasks"
}

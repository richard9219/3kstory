package models

import "time"

// DirectorPublishRecord 记录导演版导出到平台发布的真实上传回执与重试链路。
type DirectorPublishRecord struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UserID          uint       `gorm:"not null;index" json:"user_id"`
	ProjectID       uint       `gorm:"not null;index" json:"project_id"`
	VideoTaskID     uint       `gorm:"not null;index" json:"video_task_id"`
	ExportID        string     `gorm:"size:128;not null;index" json:"export_id"`
	Platform        string     `gorm:"size:32;not null;index" json:"platform"`
	Status          string     `gorm:"size:20;not null;index" json:"status"` // pending/success/failed
	AttemptNo       int        `gorm:"default:1" json:"attempt_no"`
	RetriedFromID   *uint      `gorm:"index" json:"retried_from_id"`
	ReceiptID       string     `gorm:"size:128;index" json:"receipt_id"`
	RemoteVideoID   string     `gorm:"size:256" json:"remote_video_id"`
	RemoteURL       string     `gorm:"size:1000" json:"remote_url"`
	RequestPayload  JSONMap    `gorm:"type:jsonb" json:"request_payload"`
	ResponsePayload JSONMap    `gorm:"type:jsonb" json:"response_payload"`
	ErrorMsg        string     `gorm:"type:text" json:"error_msg"`
	CompletedAt     *time.Time `json:"completed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (DirectorPublishRecord) TableName() string {
	return "director_publish_records"
}

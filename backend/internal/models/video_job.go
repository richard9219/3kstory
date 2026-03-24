package models

import "time"

// VideoJob represents an advanced generation workflow run that may produce multiple candidates.
type VideoJob struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	JobID              string     `gorm:"size:64;uniqueIndex" json:"job_id"`
	UserID             uint       `gorm:"not null;index" json:"user_id"`
	ProjectID          uint       `gorm:"not null;index" json:"project_id"`
	PipelineType       string     `gorm:"size:50;not null;index" json:"pipeline_type"`
	Status             string     `gorm:"size:20;default:pending;index" json:"status"`
	QueueStatus        string     `gorm:"size:20;default:queued;index" json:"queue_status"`
	CandidateCount     int        `gorm:"default:1" json:"candidate_count"`
	QualityMode        string     `gorm:"size:20;default:fast" json:"quality_mode"`
	ScoreProfile       string     `gorm:"size:30;default:default" json:"score_profile"`
	ProviderMode       string     `gorm:"size:20;default:single" json:"provider_mode"`
	AutoPick           bool       `gorm:"default:true" json:"auto_pick"`
	PublishThreshold   float64    `gorm:"default:0.72" json:"publish_threshold"`
	PublishGatePassed  bool       `gorm:"default:false" json:"publish_gate_passed"`
	PublishBlockReason string     `gorm:"size:255" json:"publish_block_reason"`
	SelectedTaskID     *uint      `gorm:"index" json:"selected_task_id"`
	SelectedVideoID    string     `gorm:"size:100;index" json:"selected_video_id"`
	RequestData        JSONMap    `gorm:"type:jsonb" json:"request_data"`
	ResultData         JSONMap    `gorm:"type:jsonb" json:"result_data"`
	ErrorMsg           string     `gorm:"type:text" json:"error_msg"`
	StartedAt          *time.Time `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (VideoJob) TableName() string {
	return "video_jobs"
}

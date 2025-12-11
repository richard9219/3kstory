package models

import (
"database/sql/driver"
"encoding/json"
"time"
)

type AITask struct {
ID           uint       `gorm:"primaryKey" json:"id"`
ProjectID    *uint      `gorm:"index" json:"project_id"`
SceneID      *uint      `gorm:"index" json:"scene_id"`
TaskType     string     `gorm:"size:50;not null;index" json:"task_type"`
ModelName    string     `gorm:"size:100" json:"model_name"`
InputData    JSONMap    `gorm:"type:jsonb" json:"input_data"`
OutputData   JSONMap    `gorm:"type:jsonb" json:"output_data"`
Status       string     `gorm:"size:20;default:pending;index" json:"status"`
ErrorMessage string     `gorm:"type:text" json:"error_message"`
RetryCount   int        `gorm:"default:0" json:"retry_count"`
StartedAt    *time.Time `json:"started_at"`
CompletedAt  *time.Time `json:"completed_at"`
DurationMs   int        `json:"duration_ms"`
CreatedAt    time.Time  `json:"created_at"`
}

type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
bytes, ok := value.([]byte)
if !ok {
return nil
}
return json.Unmarshal(bytes, j)
}

package models

import (
"time"
)

type User struct {
ID           uint      `gorm:"primaryKey" json:"id"`
Username     string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
Email        string    `gorm:"uniqueIndex;size:100;not null" json:"email"`
PasswordHash string    `gorm:"size:255;not null" json:"-"`
Nickname     string    `gorm:"size:50" json:"nickname"`
AvatarURL    string    `gorm:"size:500" json:"avatar_url"`
Points       int       `gorm:"default:100" json:"points"`
Role         string    `gorm:"size:20;default:user" json:"role"`
Status       string    `gorm:"size:20;default:active" json:"status"`
CreatedAt    time.Time `json:"created_at"`
UpdatedAt    time.Time `json:"updated_at"`
}

type UserResponse struct {
ID        uint      `json:"id"`
Username  string    `json:"username"`
Email     string    `json:"email"`
Nickname  string    `json:"nickname"`
AvatarURL string    `json:"avatar_url"`
Points    int       `json:"points"`
Role      string    `json:"role"`
CreatedAt time.Time `json:"created_at"`
}

func (u *User) ToResponse() *UserResponse {
return &UserResponse{
ID:        u.ID,
Username:  u.Username,
Email:     u.Email,
Nickname:  u.Nickname,
AvatarURL: u.AvatarURL,
Points:    u.Points,
Role:      u.Role,
CreatedAt: u.CreatedAt,
}
}

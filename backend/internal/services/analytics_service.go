package services

import (
	"github.com/richard9219/3kstory/internal/models"
	"gorm.io/gorm"
)

type AnalyticsService struct {
	db *gorm.DB
}

func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

type AnalyticsSummary struct {
	TotalProjects      int64                             `json:"total_projects"`
	TotalGenerated     int64                             `json:"total_generated_videos"`
	TotalCompleted     int64                             `json:"total_completed_videos"`
	TotalInteractions  int64                             `json:"total_interactions"`
	BoundPlatforms     []*models.PlatformAccountResponse `json:"bound_platforms"`
	ConfiguredPlatform []string                          `json:"configured_platforms"`
}

func (s *AnalyticsService) GetSummary(userID uint, configured []string) (*AnalyticsSummary, error) {
	var totalProjects int64
	if err := s.db.Model(&models.Project{}).Where("user_id = ?", userID).Count(&totalProjects).Error; err != nil {
		return nil, err
	}

	var totalGenerated int64
	if err := s.db.Model(&models.VideoTask{}).Where("user_id = ?", userID).Count(&totalGenerated).Error; err != nil {
		return nil, err
	}

	var totalCompleted int64
	if err := s.db.Model(&models.VideoTask{}).Where("user_id = ? AND status = ?", userID, "completed").Count(&totalCompleted).Error; err != nil {
		return nil, err
	}

	var likes int64
	if err := s.db.Model(&models.Project{}).Where("user_id = ?", userID).Select("COALESCE(SUM(like_count),0)").Scan(&likes).Error; err != nil {
		return nil, err
	}

	var views int64
	if err := s.db.Model(&models.Project{}).Where("user_id = ?", userID).Select("COALESCE(SUM(view_count),0)").Scan(&views).Error; err != nil {
		return nil, err
	}

	var accounts []models.PlatformAccount
	if err := s.db.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		return nil, err
	}

	bound := make([]*models.PlatformAccountResponse, 0, len(accounts))
	for i := range accounts {
		bound = append(bound, accounts[i].ToResponse())
	}

	return &AnalyticsSummary{
		TotalProjects:      totalProjects,
		TotalGenerated:     totalGenerated,
		TotalCompleted:     totalCompleted,
		TotalInteractions:  likes + views,
		BoundPlatforms:     bound,
		ConfiguredPlatform: configured,
	}, nil
}

func (s *AnalyticsService) ListRecentVideos(userID uint, limit int) ([]models.VideoTask, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var list []models.VideoTask
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}

	return list, nil
}

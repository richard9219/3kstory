package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/richard9219/3kstory/internal/models"
	"gorm.io/gorm"
)

type StoryboardService struct {
	db *gorm.DB
}

type StoryboardVersionNode struct {
	Root     models.StoryboardShot   `json:"root"`
	Versions []models.StoryboardShot `json:"versions"`
}

func NewStoryboardService(db *gorm.DB) *StoryboardService {
	return &StoryboardService{db: db}
}

func (s *StoryboardService) validateProjectOwnership(ctx context.Context, projectID uint, userID uint) error {
	var project models.Project
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return fmt.Errorf("project not found or no permission")
	}
	return nil
}

func (s *StoryboardService) ListProjectShots(ctx context.Context, projectID uint, userID uint) ([]models.StoryboardShot, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}
	var shots []models.StoryboardShot
	if err := s.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Order("sort_order ASC, shot_number ASC").
		Find(&shots).Error; err != nil {
		return nil, err
	}
	return shots, nil
}

func (s *StoryboardService) CreateShot(ctx context.Context, shot *models.StoryboardShot) error {
	if err := s.validateProjectOwnership(ctx, shot.ProjectID, shot.UserID); err != nil {
		return err
	}
	if shot.ShotNumber <= 0 {
		var count int64
		if err := s.db.WithContext(ctx).Model(&models.StoryboardShot{}).Where("project_id = ?", shot.ProjectID).Count(&count).Error; err == nil {
			shot.ShotNumber = int(count) + 1
			shot.SortOrder = int(count) + 1
		}
	}
	if shot.SortOrder <= 0 {
		shot.SortOrder = shot.ShotNumber
	}
	if shot.AspectRatio == "" {
		shot.AspectRatio = "16:9"
	}
	if shot.Status == "" {
		shot.Status = "draft"
	}
	if shot.Duration <= 0 {
		shot.Duration = 5
	}
	if shot.Version <= 0 {
		shot.Version = 1
	}
	return s.db.WithContext(ctx).Create(shot).Error
}

func (s *StoryboardService) BulkCreateShots(ctx context.Context, shots []models.StoryboardShot) (int, error) {
	if len(shots) == 0 {
		return 0, fmt.Errorf("shots required")
	}
	projectID := shots[0].ProjectID
	userID := shots[0].UserID
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return 0, err
	}
	for idx := range shots {
		if shots[idx].ProjectID != projectID || shots[idx].UserID != userID {
			return 0, fmt.Errorf("all shots must belong to the same project and user")
		}
		if err := s.CreateShot(ctx, &shots[idx]); err != nil {
			return idx, err
		}
	}
	return len(shots), nil
}

func (s *StoryboardService) ReorderShots(ctx context.Context, projectID uint, userID uint, orderedIDs []uint) error {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return err
	}
	if len(orderedIDs) == 0 {
		return fmt.Errorf("ordered shot ids required")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing []models.StoryboardShot
		if err := tx.Where("project_id = ? AND user_id = ?", projectID, userID).Order("sort_order ASC, shot_number ASC").Find(&existing).Error; err != nil {
			return err
		}
		if len(existing) != len(orderedIDs) {
			return fmt.Errorf("ordered ids do not match project shot count")
		}

		index := make(map[uint]int, len(orderedIDs))
		for i, id := range orderedIDs {
			index[id] = i + 1
		}

		for _, shot := range existing {
			order, ok := index[shot.ID]
			if !ok {
				return fmt.Errorf("ordered ids missing shot %d", shot.ID)
			}
			if err := tx.Model(&models.StoryboardShot{}).Where("id = ?", shot.ID).Updates(map[string]interface{}{
				"sort_order":  order,
				"shot_number": order,
			}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *StoryboardService) CreateShotVersion(ctx context.Context, projectID uint, userID uint, shotID uint, note string) (*models.StoryboardShot, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}

	var base models.StoryboardShot
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ? AND user_id = ?", shotID, projectID, userID).First(&base).Error; err != nil {
		return nil, fmt.Errorf("shot not found or no permission")
	}

	rootID := base.ID
	if base.RootShotID != nil {
		rootID = *base.RootShotID
	}

	var maxVersion int
	if err := s.db.WithContext(ctx).
		Model(&models.StoryboardShot{}).
		Where("project_id = ? AND user_id = ? AND (id = ? OR root_shot_id = ?)", projectID, userID, rootID, rootID).
		Select("COALESCE(MAX(version), 1)").
		Scan(&maxVersion).Error; err != nil {
		return nil, err
	}

	next := base
	next.ID = 0
	next.Version = maxVersion + 1
	next.ParentShotID = &base.ID
	next.RootShotID = &rootID
	next.VersionNote = note
	next.Status = "draft"

	if err := s.db.WithContext(ctx).Create(&next).Error; err != nil {
		return nil, err
	}
	return &next, nil
}

func (s *StoryboardService) ListVersionTree(ctx context.Context, projectID uint, userID uint) ([]StoryboardVersionNode, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return nil, err
	}

	var shots []models.StoryboardShot
	if err := s.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Order("sort_order ASC, version ASC, created_at ASC").
		Find(&shots).Error; err != nil {
		return nil, err
	}

	nodeMap := make(map[uint]*StoryboardVersionNode)
	for _, shot := range shots {
		rootID := shot.ID
		if shot.RootShotID != nil {
			rootID = *shot.RootShotID
		}
		node, ok := nodeMap[rootID]
		if !ok {
			node = &StoryboardVersionNode{}
			nodeMap[rootID] = node
		}
		if shot.RootShotID == nil {
			node.Root = shot
			continue
		}
		node.Versions = append(node.Versions, shot)
	}

	roots := make([]uint, 0, len(nodeMap))
	for rootID := range nodeMap {
		roots = append(roots, rootID)
	}
	sort.Slice(roots, func(i, j int) bool {
		left := nodeMap[roots[i]].Root.SortOrder
		right := nodeMap[roots[j]].Root.SortOrder
		if left == right {
			return roots[i] < roots[j]
		}
		return left < right
	})

	out := make([]StoryboardVersionNode, 0, len(roots))
	for _, rootID := range roots {
		node := nodeMap[rootID]
		sort.Slice(node.Versions, func(i, j int) bool {
			return node.Versions[i].Version < node.Versions[j].Version
		})
		out = append(out, *node)
	}
	return out, nil
}

func (s *StoryboardService) BootstrapFromScenes(ctx context.Context, projectID uint, userID uint) (int, error) {
	if err := s.validateProjectOwnership(ctx, projectID, userID); err != nil {
		return 0, err
	}

	var existing int64
	if err := s.db.WithContext(ctx).Model(&models.StoryboardShot{}).Where("project_id = ?", projectID).Count(&existing).Error; err != nil {
		return 0, err
	}
	if existing > 0 {
		return 0, nil
	}

	var scenes []models.Scene
	if err := s.db.WithContext(ctx).Where("project_id = ?", projectID).Order("scene_number ASC").Find(&scenes).Error; err != nil {
		return 0, err
	}

	if len(scenes) == 0 {
		return 0, nil
	}

	shots := make([]models.StoryboardShot, 0, len(scenes))
	for idx, scene := range scenes {
		sceneID := scene.ID
		shots = append(shots, models.StoryboardShot{
			UserID:         userID,
			ProjectID:      projectID,
			SceneID:        &sceneID,
			Chapter:        fmt.Sprintf("第%d章", idx+1),
			ShotNumber:     idx + 1,
			SortOrder:      idx + 1,
			Title:          scene.Title,
			Description:    scene.Description,
			CameraLanguage: scene.ShotType,
			Duration:       scene.Duration,
			AspectRatio:    "16:9",
			Prompt:         scene.PromptForVideo,
			Status:         "draft",
			Version:        1,
		})
	}

	if err := s.db.WithContext(ctx).Create(&shots).Error; err != nil {
		return 0, err
	}

	return len(shots), nil
}

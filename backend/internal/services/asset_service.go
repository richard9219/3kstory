package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/richard9219/3kstory/internal/models"
	"gorm.io/gorm"
)

type AssetService struct {
	db *gorm.DB
}

func NewAssetService(db *gorm.DB) *AssetService {
	return &AssetService{db: db}
}

func (s *AssetService) validateProjectOwnership(ctx context.Context, userID uint, projectID *uint) error {
	if projectID == nil {
		return nil
	}
	var project models.Project
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", *projectID, userID).First(&project).Error; err != nil {
		return fmt.Errorf("project not found or no permission")
	}
	return nil
}

func (s *AssetService) CreateRoleAsset(ctx context.Context, asset *models.RoleAsset) error {
	if err := s.validateProjectOwnership(ctx, asset.UserID, asset.ProjectID); err != nil {
		return err
	}
	asset.Tags = normalizeTags(asset.Tags)
	return s.db.WithContext(ctx).Create(asset).Error
}

func (s *AssetService) ListRoleAssets(ctx context.Context, userID uint, projectID *uint, keyword string, tags []string) ([]models.RoleAsset, error) {
	query := s.db.WithContext(ctx).Where("user_id = ?", userID)
	if projectID != nil {
		query = query.Where("project_id = ? OR project_id IS NULL", *projectID)
	}
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ? OR style_prompt ILIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	for _, tag := range sanitizeTags(tags) {
		query = query.Where("LOWER(tags) LIKE ?", "%"+strings.ToLower(tag)+"%")
	}
	var assets []models.RoleAsset
	if err := query.Order("updated_at DESC").Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (s *AssetService) UpdateRoleAsset(ctx context.Context, userID uint, assetID uint, req *models.RoleAsset) (*models.RoleAsset, error) {
	var asset models.RoleAsset
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", assetID, userID).First(&asset).Error; err != nil {
		return nil, fmt.Errorf("asset not found or no permission")
	}

	if req.ProjectID != nil {
		if err := s.validateProjectOwnership(ctx, userID, req.ProjectID); err != nil {
			return nil, err
		}
	}

	updates := map[string]interface{}{}
	if req.ProjectID != nil {
		updates["project_id"] = *req.ProjectID
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	updates["role_type"] = req.RoleType
	updates["description"] = req.Description
	updates["avatar_url"] = req.AvatarURL
	updates["voice_preset"] = req.VoicePreset
	updates["style_prompt"] = req.StylePrompt
	updates["negative_hint"] = req.NegativeHint
	updates["tags"] = normalizeTags(req.Tags)

	if err := s.db.WithContext(ctx).Model(&asset).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).First(&asset, asset.ID).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (s *AssetService) DeleteRoleAsset(ctx context.Context, userID uint, assetID uint) error {
	res := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", assetID, userID).Delete(&models.RoleAsset{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("asset not found or no permission")
	}
	return nil
}

func (s *AssetService) CreatePromptTemplate(ctx context.Context, tpl *models.PromptTemplate) error {
	if err := s.validateProjectOwnership(ctx, tpl.UserID, tpl.ProjectID); err != nil {
		return err
	}
	tpl.Tags = normalizeTags(tpl.Tags)
	return s.db.WithContext(ctx).Create(tpl).Error
}

func (s *AssetService) ListPromptTemplates(ctx context.Context, userID uint, projectID *uint, templateType string, keyword string, tags []string) ([]models.PromptTemplate, error) {
	query := s.db.WithContext(ctx).Where("user_id = ?", userID)
	if projectID != nil {
		query = query.Where("project_id = ? OR project_id IS NULL", *projectID)
	}
	if templateType != "" {
		query = query.Where("template_type = ?", templateType)
	}
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		query = query.Where("name ILIKE ? OR content ILIKE ? OR provider_type ILIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	for _, tag := range sanitizeTags(tags) {
		query = query.Where("LOWER(tags) LIKE ?", "%"+strings.ToLower(tag)+"%")
	}
	var templates []models.PromptTemplate
	if err := query.Order("updated_at DESC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (s *AssetService) UpdatePromptTemplate(ctx context.Context, userID uint, templateID uint, req *models.PromptTemplate) (*models.PromptTemplate, error) {
	var tpl models.PromptTemplate
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", templateID, userID).First(&tpl).Error; err != nil {
		return nil, fmt.Errorf("template not found or no permission")
	}

	if req.ProjectID != nil {
		if err := s.validateProjectOwnership(ctx, userID, req.ProjectID); err != nil {
			return nil, err
		}
	}

	updates := map[string]interface{}{}
	if req.ProjectID != nil {
		updates["project_id"] = *req.ProjectID
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.TemplateType != "" {
		updates["template_type"] = req.TemplateType
	}
	updates["provider_type"] = req.ProviderType
	if req.Content != "" {
		updates["content"] = req.Content
	}
	updates["variables"] = req.Variables
	updates["tags"] = normalizeTags(req.Tags)

	if err := s.db.WithContext(ctx).Model(&tpl).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).First(&tpl, tpl.ID).Error; err != nil {
		return nil, err
	}
	return &tpl, nil
}

func (s *AssetService) DeletePromptTemplate(ctx context.Context, userID uint, templateID uint) error {
	res := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", templateID, userID).Delete(&models.PromptTemplate{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("template not found or no permission")
	}
	return nil
}

func sanitizeTags(tags []string) []string {
	cleaned := make([]string, 0, len(tags))
	seen := make(map[string]struct{})
	for _, tag := range tags {
		t := strings.ToLower(strings.TrimSpace(tag))
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		cleaned = append(cleaned, t)
	}
	return cleaned
}

func normalizeTags(raw string) string {
	parts := strings.Split(raw, ",")
	return strings.Join(sanitizeTags(parts), ",")
}

package services

import (
	"context"
	"fmt"

	"github.com/richard9219/3kstory/internal/models"
	"gorm.io/gorm"
)

type ProjectService struct {
	db        *gorm.DB
	aiService *AIService
}

func NewProjectService(db *gorm.DB, aiService *AIService) *ProjectService {
	return &ProjectService{
		db:        db,
		aiService: aiService,
	}
}

func (s *ProjectService) CreateProject(userID uint, prompt, title string) (*models.Project, error) {
	project := &models.Project{
		UserID: userID,
		Title:  title,
		Prompt: prompt,
		Status: "draft",
	}

	if err := s.db.Create(project).Error; err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) GenerateScenes(ctx context.Context, projectID uint) error {
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return err
	}

	project.Status = "processing"
	s.db.Save(&project)

	script, err := s.aiService.GenerateScript(ctx, project.Prompt)
	if err != nil {
		project.Status = "failed"
		s.db.Save(&project)
		return err
	}

	project.Genre = script.Genre
	project.Style = script.Style
	if project.Title == "" {
		project.Title = script.Title
	}

	for _, sceneDetail := range script.Scenes {
		scene := models.Scene{
			ProjectID:   projectID,
			SceneNumber: sceneDetail.SceneNumber,
			Title:       sceneDetail.Title,
			Location:    sceneDetail.Location,
			Dialogue:    sceneDetail.Dialogue,
			ShotType:    sceneDetail.ShotType,
			Duration:    sceneDetail.Duration,
			Status:      "pending",
		}

		var chars models.CharacterArray
		for _, c := range sceneDetail.Characters {
			chars = append(chars, models.Character{
				Name:    c.Name,
				Emotion: c.Emotion,
			})
		}
		scene.Characters = chars

		scene.PromptForImage = fmt.Sprintf("%s, %s, %s", sceneDetail.Location, sceneDetail.Title, script.Style)

		if err := s.db.Create(&scene).Error; err != nil {
			continue
		}

		imageURL, err := s.aiService.GenerateImage(ctx, scene.PromptForImage)
		if err == nil {
			scene.MediaURL = imageURL
			scene.MediaType = "image"
			scene.Status = "completed"
			s.db.Save(&scene)
		}
	}

	project.Status = "completed"
	s.db.Save(&project)

	return nil
}

func (s *ProjectService) GetProjectWithScenes(projectID uint) (*models.Project, error) {
	var project models.Project
	err := s.db.Preload("Scenes").Preload("User").First(&project, projectID).Error
	return &project, err
}

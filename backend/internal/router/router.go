package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/richard9219/3kstory/internal/config"
	"github.com/richard9219/3kstory/internal/handlers"
	"github.com/richard9219/3kstory/internal/middleware"
	"github.com/richard9219/3kstory/internal/services"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, cfg *config.Config) {
	aiService := services.NewAIService(cfg)
	projectService := services.NewProjectService(db, aiService)
	videoService := services.NewVideoService(cfg)

	authHandler := handlers.NewAuthHandler(db, cfg)
	projectHandler := handlers.NewProjectHandler(projectService, db)
	videoHandler := handlers.NewVideoHandler(videoService, projectService)

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		authorized := v1.Group("")
		authorized.Use(middleware.AuthRequired(cfg))
		{
			users := authorized.Group("/users")
			{
				users.GET("/me", authHandler.GetProfile)
				users.PUT("/me", authHandler.UpdateProfile)
			}

			projects := authorized.Group("/projects")
			{
				projects.POST("", projectHandler.CreateProject)
				projects.GET("", projectHandler.ListProjects)
				projects.GET("/:id", projectHandler.GetProject)
				projects.PUT("/:id", projectHandler.UpdateProject)
				projects.DELETE("/:id", projectHandler.DeleteProject)
				projects.GET("/:id/scenes", projectHandler.GetScenes)
				projects.POST("/:id/generate", projectHandler.GenerateScenes)

				// Video generation endpoints (Milestone 1.1)
				projects.POST("/:id/generate-video", videoHandler.GenerateVideo)
				projects.POST("/:id/video-status", videoHandler.GetVideoStatus)
				projects.GET("/:id/videos", videoHandler.ListVideos)
				projects.DELETE("/:id/video/:videoID", videoHandler.CancelVideoGeneration)
			}
		}
	}
}

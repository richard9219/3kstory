package router

import (
	"context"

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
	videoService := services.NewVideoService(cfg, db)
	modelCenterService := services.NewModelCenterService(cfg, videoService)
	modelCenterService.StartBackgroundProbe(context.Background())
	assetService := services.NewAssetService(db)
	storyboardService := services.NewStoryboardService(db)
	ttsService := services.NewTTSService()
	narrationService := services.NewNarrationService(db, cfg, aiService, videoService, ttsService)
	platformService := services.NewPlatformService(db, cfg, &services.RedisAdapter{Client: rdb})
	analyticsService := services.NewAnalyticsService(db)

	authHandler := handlers.NewAuthHandler(db, cfg)
	projectHandler := handlers.NewProjectHandler(projectService, db)
	videoHandler := handlers.NewVideoHandler(videoService, projectService)
	platformHandler := handlers.NewPlatformHandler(platformService)
	narrationHandler := handlers.NewNarrationHandler(projectService, narrationService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, platformService)
	modelCenterHandler := handlers.NewModelCenterHandler(modelCenterService)
	assetHandler := handlers.NewAssetHandler(assetService)
	storyboardHandler := handlers.NewStoryboardHandler(storyboardService)

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// 平台相关：已配置列表与 OAuth 回调无需登录
		platformsPublic := v1.Group("/platforms")
		{
			platformsPublic.GET("/configured", platformHandler.ConfiguredPlatforms)
			platformsPublic.GET("/:platform/callback", platformHandler.OAuthCallback)
		}

		authorized := v1.Group("")
		authorized.Use(middleware.AuthRequired(cfg))
		{
			modelCenter := authorized.Group("/model-center")
			{
				modelCenter.GET("/overview", modelCenterHandler.GetOverview)
				modelCenter.POST("/probe", modelCenterHandler.TriggerProbe)
			}

			assets := authorized.Group("/assets")
			{
				assets.GET("/roles", assetHandler.ListRoleAssets)
				assets.POST("/roles", assetHandler.CreateRoleAsset)
				assets.PUT("/roles/:id", assetHandler.UpdateRoleAsset)
				assets.DELETE("/roles/:id", assetHandler.DeleteRoleAsset)
				assets.GET("/prompt-templates", assetHandler.ListPromptTemplates)
				assets.POST("/prompt-templates", assetHandler.CreatePromptTemplate)
				assets.PUT("/prompt-templates/:id", assetHandler.UpdatePromptTemplate)
				assets.DELETE("/prompt-templates/:id", assetHandler.DeletePromptTemplate)
			}

			analytics := authorized.Group("/analytics")
			{
				analytics.GET("/summary", analyticsHandler.GetSummary)
				analytics.GET("/videos", analyticsHandler.ListVideos)
			}

			users := authorized.Group("/users")
			{
				users.GET("/me", authHandler.GetProfile)
				users.PUT("/me", authHandler.UpdateProfile)
			}

			platforms := authorized.Group("/platforms")
			{
				platforms.GET("", platformHandler.ListPlatforms)
				platforms.GET("/:platform/authorize", platformHandler.GetAuthorizeURL)
				platforms.DELETE("/:platform", platformHandler.Disconnect)
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
				projects.POST("/:id/generate-narration", narrationHandler.GenerateNarrationVideo)
				projects.POST("/:id/video-status", videoHandler.GetVideoStatus)
				projects.GET("/:id/videos", videoHandler.ListVideos)
				projects.DELETE("/:id/video/:videoID", videoHandler.CancelVideoGeneration)
				projects.GET("/:id/storyboard-shots", storyboardHandler.ListProjectShots)
				projects.POST("/:id/storyboard-shots", storyboardHandler.CreateShot)
				projects.POST("/:id/storyboard-shots/import", storyboardHandler.ImportShots)
				projects.POST("/:id/storyboard-shots/bootstrap", storyboardHandler.BootstrapFromScenes)
				projects.POST("/:id/storyboard-shots/reorder", storyboardHandler.ReorderShots)
				projects.POST("/:id/storyboard-shots/version", storyboardHandler.CreateShotVersion)
				projects.GET("/:id/storyboard-shots/version-tree", storyboardHandler.GetVersionTree)
			}
		}
	}
}

package main

import (
"log"
"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/richard9219/3kstory/internal/config"
	"github.com/richard9219/3kstory/internal/database"
	"github.com/richard9219/3kstory/internal/middleware"
	"github.com/richard9219/3kstory/internal/router"
)

func main() {
if err := godotenv.Load(); err != nil {
log.Println("No .env file found, using system environment variables")
}

cfg := config.Load()

db, err := database.InitDB(cfg)
if err != nil {
log.Fatalf("Failed to connect database: %v", err)
}

rdb := database.InitRedis(cfg)

if cfg.Env == "production" {
gin.SetMode(gin.ReleaseMode)
}

r := gin.Default()

r.Use(middleware.CORS())
r.Use(middleware.Logger())
r.Use(middleware.Recovery())

r.GET("/health", func(c *gin.Context) {
c.JSON(200, gin.H{"status": "ok"})
})

router.SetupRoutes(r, db, rdb, cfg)

port := os.Getenv("PORT")
if port == "" {
port = "8080"
}

log.Printf("Server starting on port %s...", port)
if err := r.Run(":" + port); err != nil {
log.Fatalf("Failed to start server: %v", err)
}
}

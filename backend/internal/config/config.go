package config

import (
	"os"
	"strconv"
)

type Config struct {
	Env      string
	Port     string
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OSS      OSSConfig
	AI       AIConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret      string
	ExpireHours int
}

type OSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
	BaseURL         string
}

type AIConfig struct {
	QwenAPIKey       string
	QwenAPIBase      string
	TextServiceURL   string
	ImageServiceURL  string
	VideoServiceURL  string
	ReviewServiceURL string
	RunwayAPIKey     string
	PikaAPIKey       string
}

func Load() *Config {
	expireHours, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "168"))

	return &Config{
		Env:  getEnv("ENV", "development"),
		Port: getEnv("PORT", "8080"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "3kvedio"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "change_me"),
			ExpireHours: expireHours,
		},
		OSS: OSSConfig{
			Endpoint:        getEnv("OSS_ENDPOINT", ""),
			AccessKeyID:     getEnv("OSS_ACCESS_KEY_ID", ""),
			AccessKeySecret: getEnv("OSS_ACCESS_KEY_SECRET", ""),
			BucketName:      getEnv("OSS_BUCKET_NAME", ""),
			BaseURL:         getEnv("OSS_BASE_URL", ""),
		},
		AI: AIConfig{
			QwenAPIKey:       getEnv("QWEN_API_KEY", ""),
			QwenAPIBase:      getEnv("QWEN_API_BASE", ""),
			TextServiceURL:   getEnv("AI_TEXT_SERVICE_URL", ""),
			ImageServiceURL:  getEnv("AI_IMAGE_SERVICE_URL", ""),
			VideoServiceURL:  getEnv("AI_VIDEO_SERVICE_URL", ""),
			ReviewServiceURL: getEnv("AI_REVIEW_SERVICE_URL", ""),
			RunwayAPIKey:     getEnv("RUNWAY_API_KEY", ""),
			PikaAPIKey:       getEnv("PIKA_API_KEY", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

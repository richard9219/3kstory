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
	AIProvider       string
	QwenAPIKey       string
	QwenAPIBase      string
	VLLMBaseURL      string
	VLLMModelName    string
	VLLMMaxTokens    int
	VLLMTimeout      int
	OLLAMABaseURL    string
	OLLAMAModelName  string
	OLLAMAMaxTokens  int
	OLLAMATimeout    int
	TextServiceURL   string
	ImageServiceURL  string
	VideoServiceURL  string
	ReviewServiceURL string
	RunwayAPIKey     string
	PikaAPIKey       string
}

func Load() *Config {
	expireHours, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "168"))
	vllmMaxTokens, _ := strconv.Atoi(getEnv("VLLM_MAX_TOKENS", "2048"))
	vllmTimeout, _ := strconv.Atoi(getEnv("VLLM_TIMEOUT", "60"))
	ollamaMaxTokens, _ := strconv.Atoi(getEnv("OLLAMA_MAX_TOKENS", "2048"))
	ollamaTimeout, _ := strconv.Atoi(getEnv("OLLAMA_TIMEOUT", "60"))

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
			AIProvider:       getEnv("AI_PROVIDER", "cloud_qwen"),
			QwenAPIKey:       getEnv("QWEN_API_KEY", ""),
			QwenAPIBase:      getEnv("QWEN_API_BASE", ""),
			VLLMBaseURL:      getEnv("VLLM_BASE_URL", "http://localhost:8000"),
			VLLMModelName:    getEnv("VLLM_MODEL_NAME", "qwen2.5-7b"),
			VLLMMaxTokens:    vllmMaxTokens,
			VLLMTimeout:      vllmTimeout,
			OLLAMABaseURL:    getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
			OLLAMAModelName:  getEnv("OLLAMA_MODEL_NAME", "qwen2.5:7b"),
			OLLAMAMaxTokens:  ollamaMaxTokens,
			OLLAMATimeout:    ollamaTimeout,
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

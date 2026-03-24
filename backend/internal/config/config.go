package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env         string
	Port        string
	BaseURL     string // 后端对外 BaseURL，用于 OAuth redirect_uri
	FrontendURL string // 前端地址，OAuth 成功后跳转
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	OSS         OSSConfig
	AI          AIConfig
	Platform    PlatformConfig
}

// PlatformConfig 各视频平台 OAuth 配置
type PlatformConfig struct {
	Douyin      PlatformOAuthItem
	Xiaohongshu PlatformOAuthItem
	Bilibili    PlatformOAuthItem
	Weibo       PlatformOAuthItem
}

type PlatformOAuthItem struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scope        string
	PublishAPI   string
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
	AIProvider              string
	ScriptProviders         string
	NarrationProviders      string
	StoryboardProviders     string
	ShotPromptProviders     string
	ReviewProviders         string
	SceneVideoProviders     string
	NarrationVideoProviders string
	PreviewVideoProviders   string
	PremiumVideoProviders   string
	MiniMaxAPIKey           string
	MiniMaxAPIBase          string
	MiniMaxVideoModel       string
	MiniMaxVideoResolution  string
	SeedanceAPIKey          string
	SeedanceAPIBase         string
	SeedanceCreatePath      string
	SeedanceStatusPath      string
	SeedanceVideoModel      string
	SeedanceVideoResolution string
	ComfyAPIKey             string
	ComfyBaseURL            string
	ComfyWorkflowDir        string
	ComfyOutputNodeID       string
	QwenAPIKey              string
	QwenAPIBase             string
	VLLMBaseURL             string
	VLLMModelName           string
	VLLMMaxTokens           int
	VLLMTimeout             int
	OLLAMABaseURL           string
	OLLAMAModelName         string
	OLLAMAMaxTokens         int
	OLLAMATimeout           int
	TextServiceURL          string
	ImageServiceURL         string
	VideoServiceURL         string
	ReviewServiceURL        string
	RunwayAPIKey            string
	PikaAPIKey              string
	ModelProbeInterval      int
	ModelFailThreshold      int
	VideoJobQueueWorkers    int
	PublishQualityThreshold float64
	NarrationOutputDir      string
	NarrationPublicBase     string
}

func Load() *Config {
	expireHours, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "168"))
	vllmMaxTokens, _ := strconv.Atoi(getEnv("VLLM_MAX_TOKENS", "2048"))
	vllmTimeout, _ := strconv.Atoi(getEnv("VLLM_TIMEOUT", "60"))
	ollamaMaxTokens, _ := strconv.Atoi(getEnv("OLLAMA_MAX_TOKENS", "2048"))
	ollamaTimeout, _ := strconv.Atoi(getEnv("OLLAMA_TIMEOUT", "60"))
	modelProbeInterval, _ := strconv.Atoi(getEnv("MODEL_PROBE_INTERVAL", "60"))
	modelFailThreshold, _ := strconv.Atoi(getEnv("MODEL_FAIL_THRESHOLD", "3"))
	videoJobQueueWorkers, _ := strconv.Atoi(getEnv("VIDEO_JOB_QUEUE_WORKERS", "2"))
	publishQualityThreshold, _ := strconv.ParseFloat(getEnv("PUBLISH_QUALITY_THRESHOLD", "0.72"), 64)

	return &Config{
		Env:         getEnv("ENV", "development"),
		Port:        getEnv("PORT", "8080"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
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
			AIProvider:              getEnv("AI_PROVIDER", "cloud_qwen"),
			ScriptProviders:         getEnv("AI_SCRIPT_PROVIDERS", ""),
			NarrationProviders:      getEnv("AI_NARRATION_PROVIDERS", ""),
			StoryboardProviders:     getEnv("AI_STORYBOARD_PROVIDERS", ""),
			ShotPromptProviders:     getEnv("AI_SHOT_PROMPT_PROVIDERS", ""),
			ReviewProviders:         getEnv("AI_REVIEW_PROVIDERS", ""),
			SceneVideoProviders:     getEnv("AI_SCENE_VIDEO_PROVIDERS", ""),
			NarrationVideoProviders: getEnv("AI_NARRATION_VIDEO_PROVIDERS", ""),
			PreviewVideoProviders:   getEnv("AI_PREVIEW_VIDEO_PROVIDERS", ""),
			PremiumVideoProviders:   getEnv("AI_PREMIUM_VIDEO_PROVIDERS", ""),
			MiniMaxAPIKey:           getEnv("MINIMAX_API_KEY", ""),
			MiniMaxAPIBase:          strings.TrimRight(getEnv("MINIMAX_API_BASE", "https://api.minimax.io"), "/"),
			MiniMaxVideoModel:       getEnv("MINIMAX_VIDEO_MODEL", "MiniMax-Hailuo-2.3-Fast"),
			MiniMaxVideoResolution:  getEnv("MINIMAX_VIDEO_RESOLUTION", "768P"),
			SeedanceAPIKey:          getEnv("SEEDANCE_API_KEY", ""),
			SeedanceAPIBase:         strings.TrimRight(getEnv("SEEDANCE_API_BASE", ""), "/"),
			SeedanceCreatePath:      getEnv("SEEDANCE_CREATE_PATH", "/contents/generations/tasks"),
			SeedanceStatusPath:      getEnv("SEEDANCE_STATUS_PATH", "/contents/generations/tasks/{task_id}"),
			SeedanceVideoModel:      getEnv("SEEDANCE_VIDEO_MODEL", "seedance-1-0-pro-250528"),
			SeedanceVideoResolution: getEnv("SEEDANCE_VIDEO_RESOLUTION", "720p"),
			ComfyAPIKey:             getEnv("COMFY_API_KEY", ""),
			ComfyBaseURL:            strings.TrimRight(getEnv("COMFY_BASE_URL", "http://localhost:8188"), "/"),
			ComfyWorkflowDir:        getEnv("COMFY_WORKFLOW_DIR", "workflows/comfy"),
			ComfyOutputNodeID:       getEnv("COMFY_OUTPUT_NODE_ID", ""),
			QwenAPIKey:              getEnv("QWEN_API_KEY", ""),
			QwenAPIBase:             getEnv("QWEN_API_BASE", ""),
			VLLMBaseURL:             getEnv("VLLM_BASE_URL", "http://localhost:8000"),
			VLLMModelName:           getEnv("VLLM_MODEL_NAME", "qwen2.5-7b"),
			VLLMMaxTokens:           vllmMaxTokens,
			VLLMTimeout:             vllmTimeout,
			OLLAMABaseURL:           getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
			OLLAMAModelName:         getEnv("OLLAMA_MODEL_NAME", "qwen2.5:7b"),
			OLLAMAMaxTokens:         ollamaMaxTokens,
			OLLAMATimeout:           ollamaTimeout,
			TextServiceURL:          getEnv("AI_TEXT_SERVICE_URL", ""),
			ImageServiceURL:         getEnv("AI_IMAGE_SERVICE_URL", ""),
			VideoServiceURL:         getEnv("AI_VIDEO_SERVICE_URL", ""),
			ReviewServiceURL:        getEnv("AI_REVIEW_SERVICE_URL", ""),
			RunwayAPIKey:            getEnv("RUNWAY_API_KEY", ""),
			PikaAPIKey:              getEnv("PIKA_API_KEY", ""),
			ModelProbeInterval:      maxInt(modelProbeInterval, 15),
			ModelFailThreshold:      maxInt(modelFailThreshold, 1),
			VideoJobQueueWorkers:    maxInt(videoJobQueueWorkers, 1),
			PublishQualityThreshold: clampFloat(publishQualityThreshold, 0.1, 1),
			NarrationOutputDir:      getEnv("NARRATION_OUTPUT_DIR", ".local/videos/narration"),
			NarrationPublicBase:     strings.TrimRight(getEnv("NARRATION_PUBLIC_BASE", "http://localhost:8003/files/narration"), "/"),
		},
		Platform: PlatformConfig{
			Douyin: PlatformOAuthItem{
				ClientID:     getEnv("DOUYIN_CLIENT_KEY", ""),
				ClientSecret: getEnv("DOUYIN_CLIENT_SECRET", ""),
				RedirectURI:  getEnv("DOUYIN_REDIRECT_URI", ""),
				Scope:        getEnv("DOUYIN_SCOPE", "user_info,video.list,video.publish"),
				PublishAPI:   getEnv("DOUYIN_PUBLISH_API", "https://open.douyin.com/video/upload/"),
			},
			Xiaohongshu: PlatformOAuthItem{
				ClientID:     getEnv("XIAOHONGSHU_CLIENT_ID", ""),
				ClientSecret: getEnv("XIAOHONGSHU_CLIENT_SECRET", ""),
				RedirectURI:  getEnv("XIAOHONGSHU_REDIRECT_URI", ""),
				Scope:        getEnv("XIAOHONGSHU_SCOPE", "user_info,note_publish"),
				PublishAPI:   getEnv("XIAOHONGSHU_PUBLISH_API", "https://edith.xiaohongshu.com/api/sns/v1/note/publish"),
			},
			Bilibili: PlatformOAuthItem{
				ClientID:     getEnv("BILIBILI_CLIENT_ID", ""),
				ClientSecret: getEnv("BILIBILI_CLIENT_SECRET", ""),
				RedirectURI:  getEnv("BILIBILI_REDIRECT_URI", ""),
				Scope:        getEnv("BILIBILI_SCOPE", ""),
				PublishAPI:   getEnv("BILIBILI_PUBLISH_API", "https://member.bilibili.com/x/vu/web/add/v3"),
			},
			Weibo: PlatformOAuthItem{
				ClientID:     getEnv("WEIBO_CLIENT_ID", ""),
				ClientSecret: getEnv("WEIBO_CLIENT_SECRET", ""),
				RedirectURI:  getEnv("WEIBO_REDIRECT_URI", ""),
				Scope:        getEnv("WEIBO_SCOPE", ""),
				PublishAPI:   getEnv("WEIBO_PUBLISH_API", "https://api.weibo.com/2/statuses/share.json"),
			},
		},
	}
}

func maxInt(v, min int) int {
	if v < min {
		return min
	}
	return v
}

func clampFloat(v float64, min float64, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

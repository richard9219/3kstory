package services

import (
	"fmt"
	"net/url"
	"strings"
)

// TTSService 里程碑 1 的轻量实现：先产出可追踪的音频占位 URL，后续可替换真实 TTS SDK。
type TTSService struct{}

func NewTTSService() *TTSService {
	return &TTSService{}
}

func (s *TTSService) Synthesize(text, voice string, speed float64) (string, error) {
	preview := strings.TrimSpace(text)
	if len(preview) > 32 {
		preview = preview[:32]
	}
	if voice == "" {
		voice = "female_cn"
	}
	if speed <= 0 {
		speed = 1.0
	}
	return fmt.Sprintf("tts://synth?voice=%s&speed=%.2f&text=%s", url.QueryEscape(voice), speed, url.QueryEscape(preview)), nil
}

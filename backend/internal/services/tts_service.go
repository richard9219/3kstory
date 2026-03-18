package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// TTSService 在本地优先使用系统语音能力生成真实音频文件。
type TTSService struct {
	outputDir string
}

func NewTTSService() *TTSService {
	dir := strings.TrimSpace(os.Getenv("TTS_OUTPUT_DIR"))
	if dir == "" {
		dir = filepath.Join(".local", "tts")
	}
	return &TTSService{outputDir: dir}
}

func (s *TTSService) Synthesize(text, voice string, speed float64) (string, error) {
	content := strings.TrimSpace(text)
	if content == "" {
		return "", fmt.Errorf("tts text is empty")
	}
	if speed <= 0 {
		speed = 1.0
	}
	if err := os.MkdirAll(s.outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create tts output dir failed: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		return s.synthesizeWithSay(content, voice, speed)
	default:
		if path, err := s.synthesizeWithEspeak(content, voice, speed); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no supported local TTS engine found; macOS requires `say`, Linux can use `espeak`")
}

func (s *TTSService) synthesizeWithSay(text, voice string, speed float64) (string, error) {
	if _, err := exec.LookPath("say"); err != nil {
		return "", err
	}

	if voice = normalizeMacVoice(voice); voice == "" {
		voice = "Ting-Ting"
	}

	rate := int(175 * speed)
	if rate < 120 {
		rate = 120
	}
	if rate > 320 {
		rate = 320
	}

	outPath := filepath.Join(s.outputDir, fmt.Sprintf("tts-%d.aiff", time.Now().UnixNano()))
	cmd := exec.Command("say", "-v", voice, "-r", strconv.Itoa(rate), "-o", outPath, text)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("say failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return outPath, nil
}

func (s *TTSService) synthesizeWithEspeak(text, voice string, speed float64) (string, error) {
	if _, err := exec.LookPath("espeak"); err != nil {
		return "", err
	}

	if voice == "" {
		voice = "zh"
	}
	rate := int(175 * speed)
	if rate < 120 {
		rate = 120
	}
	if rate > 320 {
		rate = 320
	}

	outPath := filepath.Join(s.outputDir, fmt.Sprintf("tts-%d.wav", time.Now().UnixNano()))
	cmd := exec.Command("espeak", "-v", voice, "-s", strconv.Itoa(rate), "-w", outPath, text)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("espeak failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return outPath, nil
}

func normalizeMacVoice(voice string) string {
	switch strings.ToLower(strings.TrimSpace(voice)) {
	case "", "female_cn", "ting-ting", "tingting", "zh":
		return "Ting-Ting"
	case "male_cn", "sin-ji", "sinji":
		return "Sin-ji"
	case "mei-jia", "meijia":
		return "Mei-Jia"
	default:
		return voice
	}
}

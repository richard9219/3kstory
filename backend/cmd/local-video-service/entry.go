package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type generateRequest struct {
	Prompt          string            `json:"prompt"`
	ImageURL        string            `json:"image_url"`
	Duration        int               `json:"duration"`
	AspectRatio     string            `json:"aspect_ratio"`
	SceneID         uint              `json:"scene_id"`
	ProjectID       uint              `json:"project_id"`
	Mode            string            `json:"mode"`
	SourceVideoPath string            `json:"source_video_path"`
	SourceVideoURL  string            `json:"source_video_url"`
	Segments        []generateSegment `json:"segments"`
}

type generateSegment struct {
	Title             string `json:"title"`
	NarrationText     string `json:"narration_text"`
	EstimatedDuration int    `json:"estimated_duration"`
	AudioURL          string `json:"audio_url"`
	AudioPath         string `json:"audio_path"`
}

type generateResponse struct {
	VideoID  string `json:"video_id"`
	Status   string `json:"status"`
	VideoURL string `json:"video_url"`
	Message  string `json:"message,omitempty"`
}

type videoJob struct {
	ID       string
	Status   string
	VideoURL string
	Error    string
}

type server struct {
	publicURL string
	outputDir string

	mu   sync.RWMutex
	jobs map[string]*videoJob
}

type renderParams struct {
	Prompt          string
	ImageURL        string
	SourceVideoPath string
	SourceVideoURL  string
	W               int
	H               int
	Seconds         int
	OutPath         string
}

func main() {
	port := getenv("LOCAL_VIDEO_PORT", "8003")
	addr := ":" + port
	public := getenv("LOCAL_VIDEO_PUBLIC_BASE", "http://localhost:"+port)
	outputDir := getenv("LOCAL_VIDEO_OUTPUT_DIR", filepath.Join(".local", "videos"))

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Fatalf("ffmpeg not found. Install it first (macOS): brew install ffmpeg")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		log.Fatalf("ffprobe not found. Install it first (macOS): brew install ffmpeg")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("failed to create output dir: %v", err)
	}

	s := &server{publicURL: strings.TrimRight(public, "/"), outputDir: outputDir, jobs: map[string]*videoJob{}}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/v1/generate", s.handleGenerate)
	mux.HandleFunc("/v1/generate/", s.handleGetStatus)
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(outputDir))))

	log.Printf("local-video-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func (s *server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "prompt is required"})
		return
	}

	dur := req.Duration
	if dur <= 0 {
		dur = 10
	}
	if dur > 600 {
		dur = 600
	}
	wpx, hpx := aspectToSize(req.AspectRatio)

	id, err := randomID(12)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create id"})
		return
	}
	outFile := filepath.Join(s.outputDir, id+".mp4")
	publicURL := fmt.Sprintf("%s/files/%s.mp4", s.publicURL, id)

	job := &videoJob{ID: id, Status: "processing", VideoURL: publicURL}
	s.mu.Lock()
	s.jobs[id] = job
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	renderReq := renderParams{
		Prompt:          req.Prompt,
		ImageURL:        strings.TrimSpace(req.ImageURL),
		SourceVideoPath: strings.TrimSpace(req.SourceVideoPath),
		SourceVideoURL:  strings.TrimSpace(req.SourceVideoURL),
		W:               wpx,
		H:               hpx,
		Seconds:         dur,
		OutPath:         outFile,
	}

	var renderErr error
	if strings.TrimSpace(req.Mode) == "narration" && len(req.Segments) > 0 {
		renderErr = renderNarrationVideo(ctx, renderReq, req.Segments)
	} else {
		renderErr = renderVideo(ctx, renderReq)
	}
	if renderErr != nil {
		s.mu.Lock()
		job.Status = "failed"
		job.Error = renderErr.Error()
		s.mu.Unlock()
		writeJSON(w, http.StatusInternalServerError, generateResponse{VideoID: id, Status: "failed", Message: renderErr.Error()})
		return
	}

	s.mu.Lock()
	job.Status = "completed"
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, generateResponse{VideoID: id, Status: "completed", VideoURL: publicURL})
}

func (s *server) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	id := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/v1/generate/"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "video_id is required"})
		return
	}
	s.mu.RLock()
	job := s.jobs[id]
	s.mu.RUnlock()
	if job == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	resp := generateResponse{VideoID: job.ID, Status: job.Status, VideoURL: job.VideoURL}
	if job.Error != "" {
		resp.Message = job.Error
	}
	writeJSON(w, http.StatusOK, resp)
}

func renderVideo(ctx context.Context, p renderParams) error {
	if p.Seconds <= 0 {
		return errors.New("invalid duration")
	}
	textPath := p.OutPath + ".txt"
	if err := os.WriteFile(textPath, []byte(p.Prompt), 0o644); err != nil {
		return fmt.Errorf("failed to write text: %w", err)
	}
	defer os.Remove(textPath)

	args := []string{"-y"}
	if p.ImageURL != "" {
		imgPath, err := downloadToTemp(ctx, p.ImageURL)
		if err != nil {
			return err
		}
		defer os.Remove(imgPath)
		args = append(args, "-loop", "1", "-i", imgPath, "-t", strconv.Itoa(p.Seconds))
		vf := fmt.Sprintf("scale=%d:%d,format=yuv420p,drawtext=%s", p.W, p.H, drawTextFilter(textPath))
		args = append(args, "-vf", vf, "-r", "30", p.OutPath)
	} else {
		args = append(args, "-f", "lavfi", "-i", fmt.Sprintf("color=c=black:s=%dx%d:d=%d", p.W, p.H, p.Seconds))
		vf := fmt.Sprintf("drawtext=%s,format=yuv420p", drawTextFilter(textPath))
		args = append(args, "-vf", vf, "-r", "30", p.OutPath)
	}

	if err := runFFmpeg(ctx, args...); err != nil {
		return fmt.Errorf("ffmpeg failed: %w", err)
	}
	if _, err := os.Stat(p.OutPath); err != nil {
		return fmt.Errorf("output not created: %w", err)
	}
	return nil
}

func renderNarrationVideo(ctx context.Context, p renderParams, segments []generateSegment) error {
	if p.Seconds <= 0 {
		return errors.New("invalid duration")
	}

	source := strings.TrimSpace(p.SourceVideoPath)
	if source == "" {
		source = strings.TrimSpace(p.SourceVideoURL)
	}

	segmentFiles := make([]string, 0, len(segments))
	tempPaths := make([]string, 0, len(segments)*2)
	defer func() {
		for _, f := range tempPaths {
			_ = os.Remove(f)
		}
	}()

	clipStarts := make([]float64, len(segments))
	if source != "" {
		if starts, err := buildClipStarts(ctx, source, segments); err == nil {
			clipStarts = starts
		}
	}

	for i, seg := range segments {
		segDuration := segmentDuration(seg)
		segmentOut := filepath.Join(filepath.Dir(p.OutPath), fmt.Sprintf("%s-seg-%02d.mp4", strings.TrimSuffix(filepath.Base(p.OutPath), ".mp4"), i))
		audioInput := strings.TrimSpace(seg.AudioPath)
		if audioInput == "" {
			audioInput = strings.TrimSpace(seg.AudioURL)
		}

		args := []string{"-y"}
		if source != "" {
			args = append(args, "-ss", formatSeconds(clipStarts[i]), "-t", strconv.Itoa(segDuration), "-i", source)
		} else if p.ImageURL != "" {
			imgPath, err := downloadToTemp(ctx, p.ImageURL)
			if err != nil {
				return err
			}
			tempPaths = append(tempPaths, imgPath)
			args = append(args, "-loop", "1", "-i", imgPath)
		} else {
			args = append(args, "-f", "lavfi", "-i", fmt.Sprintf("color=c=black:s=%dx%d:d=%d", p.W, p.H, segDuration))
		}

		if audioInput != "" {
			args = append(args, "-i", audioInput)
		} else {
			args = append(args, "-f", "lavfi", "-t", strconv.Itoa(segDuration), "-i", "anullsrc=r=44100:cl=stereo")
		}

		textPath, err := writeSegmentText(seg)
		if err != nil {
			return err
		}
		tempPaths = append(tempPaths, textPath)

		visualFilter := buildNarrationVisualFilter(p.W, p.H, source == "", textPath)
		args = append(
			args,
			"-t", strconv.Itoa(segDuration),
			"-vf", visualFilter,
			"-af", "apad",
			"-map", "0:v:0",
			"-map", "1:a:0",
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-pix_fmt", "yuv420p",
			"-c:a", "aac",
			"-b:a", "192k",
			"-shortest",
			segmentOut,
		)

		if err := runFFmpeg(ctx, args...); err != nil {
			return fmt.Errorf("render narration segment %d failed: %w", i+1, err)
		}
		segmentFiles = append(segmentFiles, segmentOut)
		tempPaths = append(tempPaths, segmentOut)
	}

	if len(segmentFiles) == 0 {
		return renderVideo(ctx, p)
	}

	listFile := p.OutPath + ".concat.txt"
	var builder strings.Builder
	for _, f := range segmentFiles {
		builder.WriteString("file '")
		builder.WriteString(strings.ReplaceAll(f, "'", "'\\''"))
		builder.WriteString("'")
		builder.WriteString("\n")
	}
	if err := os.WriteFile(listFile, []byte(builder.String()), 0o644); err != nil {
		return fmt.Errorf("failed to write concat list: %w", err)
	}
	defer os.Remove(listFile)

	if err := runFFmpeg(ctx,
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", listFile,
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "192k",
		p.OutPath,
	); err != nil {
		return fmt.Errorf("concat narration video failed: %w", err)
	}

	if _, err := os.Stat(p.OutPath); err != nil {
		return fmt.Errorf("output not created: %w", err)
	}
	return nil
}

func buildClipStarts(ctx context.Context, source string, segments []generateSegment) ([]float64, error) {
	duration, err := probeDuration(ctx, source)
	if err != nil {
		return nil, err
	}
	totalSegmentSeconds := 0
	for _, seg := range segments {
		totalSegmentSeconds += segmentDuration(seg)
	}
	if totalSegmentSeconds <= 0 || duration <= 0 {
		return nil, fmt.Errorf("invalid clip duration")
	}

	available := duration - float64(totalSegmentSeconds)
	if available < 0 {
		available = 0
	}
	step := 0.0
	if len(segments) > 1 {
		step = available / float64(len(segments)-1)
	}

	starts := make([]float64, len(segments))
	cursor := 0.0
	for i, seg := range segments {
		maxStart := duration - float64(segmentDuration(seg))
		if maxStart < 0 {
			maxStart = 0
		}
		if cursor > maxStart {
			cursor = maxStart
		}
		starts[i] = cursor
		cursor += step
	}
	return starts, nil
}

func buildNarrationVisualFilter(w, h int, isStatic bool, textPath string) string {
	base := fmt.Sprintf("scale=%d:%d", w, h)
	if isStatic {
		base = fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d", w, h, w, h)
	}
	return fmt.Sprintf("%s,drawtext=%s,format=yuv420p", base, drawBottomTextFilter(textPath))
}

func writeSegmentText(seg generateSegment) (string, error) {
	line := strings.TrimSpace(seg.Title)
	if line != "" {
		line += "："
	}
	line += strings.TrimSpace(seg.NarrationText)
	if line == "" {
		line = "电影解说片段"
	}
	tmp, err := os.CreateTemp("", "3kstory-seg-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create segment text file: %w", err)
	}
	if _, err := tmp.WriteString(line); err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("failed to write segment text: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("failed to close segment text file: %w", err)
	}
	return tmp.Name(), nil
}

func drawTextFontOption() string {
	if font := findFontFile(); font != "" {
		return "fontfile=" + font + ":"
	}
	return ""
}

func drawTextFilter(textPath string) string {
	return fmt.Sprintf(
		"%stextfile=%s:fontcolor=white:fontsize=36:box=1:boxcolor=black@0.55:boxborderw=18:x=40:y=40:line_spacing=10",
		drawTextFontOption(),
		escapeFFmpegPath(textPath),
	)
}

func drawBottomTextFilter(textPath string) string {
	return fmt.Sprintf(
		"%stextfile=%s:fontcolor=white:fontsize=34:box=1:boxcolor=black@0.62:boxborderw=18:x=40:y=h-180:line_spacing=10",
		drawTextFontOption(),
		escapeFFmpegPath(textPath),
	)
}

func escapeFFmpegPath(path string) string {
	replacer := strings.NewReplacer("\\", "\\\\", ":", "\\:", "'", "\\'")
	return replacer.Replace(path)
}

func downloadToTemp(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to download image (status %d): %s", resp.StatusCode, string(b))
	}
	f, err := os.CreateTemp("", "3kstory-img-*.bin")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = os.Remove(f.Name())
		}
	}()
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func findFontFile() string {
	for _, p := range []string{
		"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
		"/System/Library/Fonts/Supplemental/PingFang.ttc",
		"/System/Library/Fonts/Supplemental/Arial.ttf",
		"/Library/Fonts/Arial Unicode.ttf",
		"/Library/Fonts/Arial.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func aspectToSize(ar string) (int, int) {
	if strings.TrimSpace(ar) == "9:16" {
		return 720, 1280
	}
	return 1280, 720
}

func segmentDuration(seg generateSegment) int {
	d := seg.EstimatedDuration
	if d <= 0 {
		d = 6
	}
	if audioPath := strings.TrimSpace(seg.AudioPath); audioPath != "" {
		if sec, err := probeDuration(context.Background(), audioPath); err == nil {
			audioSeconds := int(sec + 0.999)
			if audioSeconds > d {
				d = audioSeconds
			}
		}
	}
	return d
}

func probeDuration(ctx context.Context, input string) (float64, error) {
	out, err := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		input,
	).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	sec, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, fmt.Errorf("parse ffprobe duration failed: %w", err)
	}
	return sec, nil
}

func runFFmpeg(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func formatSeconds(sec float64) string {
	return strconv.FormatFloat(sec, 'f', 3, 64)
}

func randomID(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getenv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

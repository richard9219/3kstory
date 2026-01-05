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
	Prompt      string `json:"prompt"`
	ImageURL    string `json:"image_url"`
	Duration    int    `json:"duration"`
	AspectRatio string `json:"aspect_ratio"`
	SceneID     uint   `json:"scene_id"`
	ProjectID   uint   `json:"project_id"`
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

func main() {
	port := getenv("LOCAL_VIDEO_PORT", "8003")
	addr := ":" + port
	public := getenv("LOCAL_VIDEO_PUBLIC_BASE", "http://localhost:"+port)
	outputDir := getenv("LOCAL_VIDEO_OUTPUT_DIR", filepath.Join(".local", "videos"))

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Fatalf("ffmpeg not found. Install it first (macOS): brew install ffmpeg")
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
	if dur > 60 {
		dur = 60
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

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()
	if err := renderVideo(ctx, renderParams{Prompt: req.Prompt, ImageURL: strings.TrimSpace(req.ImageURL), W: wpx, H: hpx, Seconds: dur, OutPath: outFile}); err != nil {
		s.mu.Lock()
		job.Status = "failed"
		job.Error = err.Error()
		s.mu.Unlock()
		writeJSON(w, http.StatusInternalServerError, generateResponse{VideoID: id, Status: "failed", Message: err.Error()})
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

type renderParams struct {
	Prompt   string
	ImageURL string
	W        int
	H        int
	Seconds  int
	OutPath  string
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

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if _, err := os.Stat(p.OutPath); err != nil {
		return fmt.Errorf("output not created: %w", err)
	}
	return nil
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

func drawTextFilter(textPath string) string {
	fontOpt := ""
	if font := findFontFile(); font != "" {
		fontOpt = "fontfile=" + font + ":"
	}
	return fmt.Sprintf(
		"%stextfile=%s:fontcolor=white:fontsize=36:box=1:boxcolor=black@0.55:boxborderw=18:x=40:y=40:line_spacing=10",
		fontOpt,
		textPath,
	)
}

func findFontFile() string {
	for _, p := range []string{
		"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
		"/System/Library/Fonts/Supplemental/Arial.ttf",
		"/System/Library/Fonts/Supplemental/Helvetica.ttf",
		"/Library/Fonts/Arial Unicode.ttf",
		"/Library/Fonts/Arial.ttf",
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

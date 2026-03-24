package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/richard9219/3kstory/internal/config"
	"github.com/richard9219/3kstory/internal/models"
)

type PublishUploadRequest struct {
	Title     string
	Desc      string
	VideoPath string
	CoverPath string
}

type PublishUploadReceipt struct {
	Platform      string
	Status        string
	ReceiptID     string
	RemoteVideoID string
	RemoteURL     string
	HTTPStatus    int
	RequestID     string
	RawBody       string
	ReceivedAt    time.Time
}

type PlatformUploader interface {
	Upload(ctx context.Context, account *models.PlatformAccount, req PublishUploadRequest) (*PublishUploadReceipt, error)
}

type platformUploaderFactory struct {
	cfg    *config.Config
	client *http.Client
}

func newPlatformUploaderFactory(cfg *config.Config) *platformUploaderFactory {
	return &platformUploaderFactory{
		cfg: cfg,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (f *platformUploaderFactory) Uploader(platform string) (PlatformUploader, error) {
	switch platform {
	case models.PlatformDouyin:
		return &douyinUploader{baseUploader{cfg: f.cfg, client: f.client}}, nil
	case models.PlatformXiaohongshu:
		return &xiaohongshuUploader{baseUploader{cfg: f.cfg, client: f.client}}, nil
	case models.PlatformBilibili:
		return &bilibiliUploader{baseUploader{cfg: f.cfg, client: f.client}}, nil
	case models.PlatformWeibo:
		return &weiboUploader{baseUploader{cfg: f.cfg, client: f.client}}, nil
	default:
		return nil, fmt.Errorf("unsupported platform uploader: %s", platform)
	}
}

type baseUploader struct {
	cfg    *config.Config
	client *http.Client
}

func (u *baseUploader) buildMultipartBody(req PublishUploadRequest, platform string) (*bytes.Buffer, string, error) {
	videoPath := strings.TrimSpace(req.VideoPath)
	if videoPath == "" {
		return nil, "", fmt.Errorf("video path is required")
	}
	file, err := os.Open(videoPath)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	_ = writer.WriteField("title", strings.TrimSpace(req.Title))
	_ = writer.WriteField("desc", strings.TrimSpace(req.Desc))
	_ = writer.WriteField("platform", platform)

	videoPart, err := writer.CreateFormFile("video", filepath.Base(videoPath))
	if err != nil {
		return nil, "", err
	}
	if _, err := io.Copy(videoPart, file); err != nil {
		return nil, "", err
	}

	if strings.TrimSpace(req.CoverPath) != "" {
		if coverFile, openErr := os.Open(req.CoverPath); openErr == nil {
			defer coverFile.Close()
			if coverPart, createErr := writer.CreateFormFile("cover", filepath.Base(req.CoverPath)); createErr == nil {
				_, _ = io.Copy(coverPart, coverFile)
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return buf, writer.FormDataContentType(), nil
}

func (u *baseUploader) parseReceipt(platform string, resp *http.Response) (*PublishUploadReceipt, error) {
	body, _ := io.ReadAll(resp.Body)
	rawBody := strings.TrimSpace(string(body))
	if len(rawBody) > 8000 {
		rawBody = rawBody[:8000]
	}
	requestID := strings.TrimSpace(resp.Header.Get("x-tt-logid"))
	if requestID == "" {
		requestID = strings.TrimSpace(resp.Header.Get("x-request-id"))
	}
	if requestID == "" {
		requestID = strings.TrimSpace(resp.Header.Get("x-bili-trace-id"))
	}

	receipt := &PublishUploadReceipt{
		Platform:   platform,
		Status:     "failed",
		HTTPStatus: resp.StatusCode,
		RequestID:  requestID,
		RawBody:    rawBody,
		ReceivedAt: time.Now(),
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err == nil {
		receipt.RemoteVideoID = firstNonEmptyString(
			asString(payload["video_id"]),
			firstNonEmptyString(
				asString(payload["vid"]),
				firstNonEmptyString(asString(payload["aweme_id"]), asString(payload["bvid"])),
			),
		)
		receipt.RemoteURL = firstNonEmptyString(
			asString(payload["share_url"]),
			firstNonEmptyString(asString(payload["url"]), asString(payload["video_url"])),
		)
		receipt.ReceiptID = firstNonEmptyString(
			asString(payload["request_id"]),
			firstNonEmptyString(asString(payload["trace_id"]), requestID),
		)
		if code, ok := payload["code"]; ok {
			if asFloat(code) == 0 && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				receipt.Status = "success"
			}
		}
	}

	if receipt.ReceiptID == "" {
		receipt.ReceiptID = requestID
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && receipt.Status != "success" {
		receipt.Status = "success"
	}
	if strings.Contains(strings.ToLower(rawBody), "error") || strings.Contains(strings.ToLower(rawBody), "invalid") {
		receipt.Status = "failed"
	}
	if receipt.Status != "success" {
		return receipt, fmt.Errorf("platform upload failed: http=%d request_id=%s", resp.StatusCode, requestID)
	}
	return receipt, nil
}

type douyinUploader struct{ baseUploader }

type xiaohongshuUploader struct{ baseUploader }

type bilibiliUploader struct{ baseUploader }

type weiboUploader struct{ baseUploader }

func (u *douyinUploader) Upload(ctx context.Context, account *models.PlatformAccount, req PublishUploadRequest) (*PublishUploadReceipt, error) {
	endpoint := strings.TrimSpace(u.cfg.Platform.Douyin.PublishAPI)
	if endpoint == "" {
		return nil, fmt.Errorf("douyin publish api not configured")
	}
	body, contentType, err := u.buildMultipartBody(req, models.PlatformDouyin)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("access-token", account.AccessToken)
	httpReq.Header.Set("Authorization", "Bearer "+account.AccessToken)
	resp, err := u.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return u.parseReceipt(models.PlatformDouyin, resp)
}

func (u *xiaohongshuUploader) Upload(ctx context.Context, account *models.PlatformAccount, req PublishUploadRequest) (*PublishUploadReceipt, error) {
	endpoint := strings.TrimSpace(u.cfg.Platform.Xiaohongshu.PublishAPI)
	if endpoint == "" {
		return nil, fmt.Errorf("xiaohongshu publish api not configured")
	}
	body, contentType, err := u.buildMultipartBody(req, models.PlatformXiaohongshu)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("Authorization", "Bearer "+account.AccessToken)
	resp, err := u.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return u.parseReceipt(models.PlatformXiaohongshu, resp)
}

func (u *bilibiliUploader) Upload(ctx context.Context, account *models.PlatformAccount, req PublishUploadRequest) (*PublishUploadReceipt, error) {
	endpoint := strings.TrimSpace(u.cfg.Platform.Bilibili.PublishAPI)
	if endpoint == "" {
		return nil, fmt.Errorf("bilibili publish api not configured")
	}
	body, contentType, err := u.buildMultipartBody(req, models.PlatformBilibili)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("Authorization", "Bearer "+account.AccessToken)
	resp, err := u.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return u.parseReceipt(models.PlatformBilibili, resp)
}

func (u *weiboUploader) Upload(ctx context.Context, account *models.PlatformAccount, req PublishUploadRequest) (*PublishUploadReceipt, error) {
	endpoint := strings.TrimSpace(u.cfg.Platform.Weibo.PublishAPI)
	if endpoint == "" {
		return nil, fmt.Errorf("weibo publish api not configured")
	}
	body, contentType, err := u.buildMultipartBody(req, models.PlatformWeibo)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", contentType)
	resp, err := u.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return u.parseReceipt(models.PlatformWeibo, resp)
}

func asString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case json.Number:
		return x.String()
	case float64:
		if x == float64(int64(x)) {
			return fmt.Sprintf("%d", int64(x))
		}
		return fmt.Sprintf("%.6f", x)
	case int:
		return fmt.Sprintf("%d", x)
	case int64:
		return fmt.Sprintf("%d", x)
	default:
		return ""
	}
}

func asFloat(v interface{}) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case json.Number:
		if parsed, err := x.Float64(); err == nil {
			return parsed
		}
	case string:
		if parsed, err := json.Number(strings.TrimSpace(x)).Float64(); err == nil {
			return parsed
		}
	}
	return -1
}

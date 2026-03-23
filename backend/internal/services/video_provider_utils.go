package services

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func normalizeVideoStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "queued", "submitted", "created", "pending":
		return "pending"
	case "processing", "running", "in_progress":
		return "processing"
	case "completed", "succeeded", "success", "done":
		return "completed"
	case "failed", "error", "cancelled", "canceled":
		return "failed"
	default:
		return strings.ToLower(strings.TrimSpace(status))
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func ioReadAll(resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

func joinBaseAndPath(base, path string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	path = "/" + strings.TrimLeft(strings.TrimSpace(path), "/")
	if base == "" {
		return path
	}
	return base + path
}

func interpolatePath(template string, replacements map[string]string) string {
	out := template
	for key, value := range replacements {
		out = strings.ReplaceAll(out, "{"+key+"}", value)
	}
	return out
}

func loadWorkflowFromPath(path string) (map[string]interface{}, error) {
	body, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	var workflow map[string]interface{}
	if err := json.Unmarshal(body, &workflow); err != nil {
		return nil, err
	}
	return workflow, nil
}

func cloneJSONMap(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(in))
	for key, value := range in {
		if child, ok := value.(map[string]interface{}); ok {
			out[key] = cloneJSONMap(child)
			continue
		}
		out[key] = value
	}
	return out
}

func setQueryParam(rawURL, key, value string) string {
	if strings.TrimSpace(value) == "" {
		return rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	query := parsed.Query()
	query.Set(key, value)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

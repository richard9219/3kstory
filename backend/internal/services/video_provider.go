package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/richard9219/3kstory/internal/config"
)

type VideoErrorKind string

const (
	VideoErrorConfig         VideoErrorKind = "config"
	VideoErrorAuth           VideoErrorKind = "auth"
	VideoErrorRateLimit      VideoErrorKind = "rate_limit"
	VideoErrorInvalidRequest VideoErrorKind = "invalid_request"
	VideoErrorProvider       VideoErrorKind = "provider"
	VideoErrorPolling        VideoErrorKind = "polling"
	VideoErrorUnsupported    VideoErrorKind = "unsupported"
)

type VideoProviderError struct {
	Provider   VideoProvider
	Kind       VideoErrorKind
	StatusCode int
	Retryable  bool
	Message    string
	Cause      error
}

func (e *VideoProviderError) Error() string {
	provider := string(e.Provider)
	if provider == "" {
		provider = "video"
	}
	if e.Message != "" {
		return fmt.Sprintf("%s %s error: %s", provider, e.Kind, e.Message)
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s %s error: %v", provider, e.Kind, e.Cause)
	}
	return fmt.Sprintf("%s %s error", provider, e.Kind)
}

func (e *VideoProviderError) Unwrap() error {
	return e.Cause
}

type videoProviderClient interface {
	Name() VideoProvider
	ValidateConfig() error
	HealthCheck(ctx context.Context) ProviderHealth
	Generate(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error)
	PollStatus(ctx context.Context, videoID string) (*VideoGenerationResult, error)
}

type ProviderHealth struct {
	Provider   VideoProvider `json:"provider"`
	Configured bool          `json:"configured"`
	Healthy    bool          `json:"healthy"`
	Message    string        `json:"message"`
	ErrorKind  string        `json:"error_kind,omitempty"`
	CheckedAt  string        `json:"checked_at"`
}

type providerRuntime struct {
	cfg *config.Config
}

func (r providerRuntime) classifyHTTPError(provider VideoProvider, status int, body []byte, operation string) error {
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(status)
	}
	err := &VideoProviderError{
		Provider:   provider,
		StatusCode: status,
		Message:    fmt.Sprintf("%s failed: %s", operation, message),
	}
	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		err.Kind = VideoErrorAuth
	case status == http.StatusTooManyRequests:
		err.Kind = VideoErrorRateLimit
		err.Retryable = true
	case status == http.StatusBadRequest || status == http.StatusUnprocessableEntity:
		err.Kind = VideoErrorInvalidRequest
	case status >= 500:
		err.Kind = VideoErrorProvider
		err.Retryable = true
	default:
		err.Kind = VideoErrorProvider
	}
	return err
}

func (r providerRuntime) configError(provider VideoProvider, message string) error {
	return &VideoProviderError{
		Provider: provider,
		Kind:     VideoErrorConfig,
		Message:  message,
	}
}

func (r providerRuntime) pollingError(provider VideoProvider, message string, cause error) error {
	return &VideoProviderError{
		Provider:  provider,
		Kind:      VideoErrorPolling,
		Message:   message,
		Cause:     cause,
		Retryable: true,
	}
}

package yandex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"github.com/artem13815/hr/multiagent/internal/domain"
	"github.com/artem13815/hr/multiagent/internal/usecase"
)

const endpoint = "https://ai.api.cloud.yandex.net/v1/responses"

// Config carries everything the adapter needs to dial Yandex. Built once
// in bootstrap from config.YandexCloud + RateLimit.
type Config struct {
	FolderID        string
	APIKey          string
	Model           string // e.g. "qwen3.6-35b-a3b/latest"
	RequestTimeout  time.Duration
	RateLimitRPS    float64
	RateLimitBurst  int
	MaxOutputTokens int
}

type Client struct {
	cfg     Config
	http    *http.Client
	limiter *rate.Limiter
}

func New(cfg Config) *Client {
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 60 * time.Second
	}
	return &Client{
		cfg:     cfg,
		http:    &http.Client{Timeout: cfg.RequestTimeout},
		limiter: rate.NewLimiter(rate.Limit(cfg.RateLimitRPS), cfg.RateLimitBurst),
	}
}

func (c *Client) modelURI() string {
	return fmt.Sprintf("gpt://%s/%s", c.cfg.FolderID, c.cfg.Model)
}

type requestBody struct {
	Model           string  `json:"model"`
	Temperature     float32 `json:"temperature"`
	Instructions    string  `json:"instructions"`
	Input           string  `json:"input"`
	MaxOutputTokens int     `json:"max_output_tokens"`
}

type responseBody struct {
	Output []struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func (r responseBody) text() string {
	for _, o := range r.Output {
		if len(o.Content) > 0 {
			return o.Content[0].Text
		}
	}
	return ""
}

// Complete sends one inference call. Errors map to typed sentinels:
//   - rate-limit waits respect the caller's ctx (Wait returns ctx err on
//     deadline / cancel — we propagate it).
//   - 429 / 5xx / network errors -> ErrLLMUnavailable so the caller treats
//     them as soft failures.
//   - other non-2xx -> wrapped error including status code.
//   - empty body / unparseable response -> ErrLLMInvalidResponse.
//
// A single call is a single HTTP request. No retry inside the adapter —
// retries belong to the caller (usecase decides whether to fall back to
// heuristic or surface the failure).
func (c *Client) Complete(ctx context.Context, in domain.CompletionRequest) (string, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return "", err
	}

	maxOut := in.MaxOutputTokens
	if maxOut == 0 {
		maxOut = c.cfg.MaxOutputTokens
	}

	payload, err := json.Marshal(requestBody{
		Model:           c.modelURI(),
		Temperature:     in.Temperature,
		Instructions:    in.Instructions,
		Input:           in.Input,
		MaxOutputTokens: maxOut,
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+c.cfg.APIKey)
	req.Header.Set("OpenAI-Project", c.cfg.FolderID)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", usecase.ErrLLMUnavailable, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			slog.Warn("yandex llm transient failure",
				"status", resp.StatusCode,
				"body", truncate(string(body), 256),
			)
			return "", fmt.Errorf("%w: status=%d", usecase.ErrLLMUnavailable, resp.StatusCode)
		}
		return "", fmt.Errorf("yandex llm: status=%d body=%s", resp.StatusCode, truncate(string(body), 256))
	}

	var parsed responseBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("%w: unmarshal: %v", usecase.ErrLLMInvalidResponse, err)
	}
	text := parsed.text()
	if text == "" {
		return "", fmt.Errorf("%w: empty completion", usecase.ErrLLMInvalidResponse)
	}
	return text, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// compile-time guard: keep the Client honest about implementing the port.
var _ usecase.LLM = (*Client)(nil)

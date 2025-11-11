package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Client is a minimal OpenRouter (OpenAI-compatible) chat completions client.
type Client struct {
	APIKey   string
	BaseURL  string
	Model    string
	AppTitle string
	Referer  string
	httpDo   *http.Client
}

func New(apiKey, baseURL, model, appTitle, referer string) *Client {
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	return &Client{
		APIKey:   apiKey,
		BaseURL:  baseURL,
		Model:    model,
		AppTitle: appTitle,
		Referer:  referer,
		httpDo: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionsRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	// Keep defaults conservative; callers can change by editing fields if needed.
	Temperature float32 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

type chatChoice struct {
	Index   int `json:"index"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
}

type chatCompletionsResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
}

// Ask sends resume text to the LLM with a given instruction and returns the model reply.
func (c *Client) Ask(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if c.APIKey == "" {
		return "", errors.New("openrouter api key is empty")
	}
	model := c.Model
	if model == "" {
		model = "qwen/qwen2.5-32b-instruct"
	}
	reqBody := chatCompletionsRequest{
		Model: model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.2,
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("%s/chat/completions", c.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	if c.Referer != "" {
		httpReq.Header.Set("HTTP-Referer", c.Referer)
	}
	if c.AppTitle != "" {
		httpReq.Header.Set("X-Title", c.AppTitle)
	}

	resp, err := c.httpDo.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errMap map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errMap)
		return "", fmt.Errorf("openrouter http %d: %v", resp.StatusCode, errMap)
	}
	var out chatCompletionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", errors.New("no choices returned by model")
	}
	return out.Choices[0].Message.Content, nil
}

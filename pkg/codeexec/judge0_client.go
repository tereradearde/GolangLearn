package codeexec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client клиент для Judge0 API
type Client struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

// NewClient создает новый клиент Judge0
func NewClient(apiURL, apiKey string) *Client {
	return &Client{
		apiURL: apiURL,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// LanguageID маппинг языков на ID Judge0
var LanguageID = map[string]int{
	"python":     71, // Python 3
	"javascript": 63, // Node.js
	"java":       62, // Java
	"go":         60, // Go
	"cpp":        54, // C++17
}

// SubmissionRequest запрос на выполнение кода
type SubmissionRequest struct {
	SourceCode   string `json:"source_code"`
	LanguageID   int    `json:"language_id"`
	Stdin        string `json:"stdin,omitempty"`
	CPUTimeLimit int    `json:"cpu_time_limit,omitempty"` // seconds
	MemoryLimit  int    `json:"memory_limit,omitempty"`   // KB
}

// SubmissionResponse ответ от Judge0
type SubmissionResponse struct {
	Token string `json:"token"`
}

// SubmissionResult результат выполнения
type SubmissionResult struct {
	Status struct {
		ID          int    `json:"id"`
		Description string `json:"description"`
	} `json:"status"`
	Stdout        string `json:"stdout"`
	Stderr        string `json:"stderr"`
	CompileOutput string `json:"compile_output"`
	Time          string `json:"time"`   // seconds
	Memory        int    `json:"memory"` // KB
	ExitCode      int    `json:"exit_code"`
	ExitSignal    *int   `json:"exit_signal"`
}

// Submit создает submission и возвращает token
func (c *Client) Submit(ctx context.Context, req SubmissionRequest) (string, error) {
	url := c.apiURL + "/submissions"
	if c.apiKey != "" {
		url += "?base64_encoded=false&wait=false"
	} else {
		url += "?base64_encoded=false&wait=false"
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-RapidAPI-Key", c.apiKey)
		httpReq.Header.Set("X-RapidAPI-Host", "judge0-ce.p.rapidapi.com")
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("judge0 api error: %d - %s", resp.StatusCode, string(body))
	}

	var submissionResp SubmissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&submissionResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return submissionResp.Token, nil
}

// GetResult получает результат выполнения по token
func (c *Client) GetResult(ctx context.Context, token string) (*SubmissionResult, error) {
	url := c.apiURL + "/submissions/" + token + "?base64_encoded=false"

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if c.apiKey != "" {
		httpReq.Header.Set("X-RapidAPI-Key", c.apiKey)
		httpReq.Header.Set("X-RapidAPI-Host", "judge0-ce.p.rapidapi.com")
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("judge0 api error: %d - %s", resp.StatusCode, string(body))
	}

	var result SubmissionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// SubmitAndWait создает submission и ждет результата (для синхронного выполнения)
func (c *Client) SubmitAndWait(ctx context.Context, req SubmissionRequest, timeout time.Duration) (*SubmissionResult, error) {
	token, err := c.Submit(ctx, req)
	if err != nil {
		return nil, err
	}

	// Ждем результат с polling
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			result, err := c.GetResult(ctx, token)
			if err != nil {
				return nil, err
			}

			// Status 3 = Accepted (completed)
			if result.Status.ID == 3 {
				return result, nil
			}

			// Status 1-2 = In Queue / Processing
			if result.Status.ID <= 2 {
				continue
			}

			// Другие статусы (ошибки компиляции, runtime errors и т.д.)
			return result, nil
		}
	}

	return nil, fmt.Errorf("timeout waiting for result")
}

// IsAvailable проверяет доступность API
func (c *Client) IsAvailable() bool {
	return c.apiURL != ""
}

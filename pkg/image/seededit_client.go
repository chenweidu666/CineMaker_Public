package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type SeededitClient struct {
	BaseURL    string
	APIKey     string
	Model      string
	Endpoint   string
	HTTPClient *http.Client
}

type SeededitRequest struct {
	Model         string  `json:"model"`
	Prompt        string  `json:"prompt"`
	Image         string  `json:"image"`
	Seed          int64   `json:"seed,omitempty"`
	GuidanceScale float64 `json:"guidance_scale,omitempty"`
	Size          string  `json:"size,omitempty"`
	Watermark     bool    `json:"watermark"`
}

func NewSeededitClient(baseURL, apiKey, model, endpoint string) *SeededitClient {
	if endpoint == "" {
		endpoint = "/api/v3/images/generations"
	}
	if model == "" {
		model = "doubao-seededit-3-0-i2i-250628"
	}
	return &SeededitClient{
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Model:    model,
		Endpoint: endpoint,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

func (c *SeededitClient) GenerateImage(prompt string, opts ...ImageOption) (*ImageResult, error) {
	options := &ImageOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if len(options.ReferenceImages) == 0 {
		return nil, fmt.Errorf("seededit requires a source image (pass via ReferenceImages)")
	}

	sourceImage := options.ReferenceImages[0]

	model := c.Model
	if options.Model != "" {
		model = options.Model
	}

	guidanceScale := options.CfgScale
	if guidanceScale <= 0 {
		guidanceScale = 5.5
	}

	reqBody := SeededitRequest{
		Model:         model,
		Prompt:        prompt,
		Image:         sourceImage,
		GuidanceScale: guidanceScale,
		Size:          "adaptive",
		Watermark:     false,
	}

	if options.Seed > 0 {
		reqBody.Seed = options.Seed
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := c.BaseURL + c.Endpoint
	slog.Info("[Seededit] Sending request", "url", url, "model", model, "prompt_len", len(prompt))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	slog.Info("[Seededit] Response received", "status", resp.StatusCode, "body_len", len(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		if strings.Contains(string(body), "SensitiveContentDetected") ||
			strings.Contains(string(body), "sensitive") {
			return nil, fmt.Errorf("内容安全审核未通过：编辑指令触发了安全策略，请修改后重试")
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result VolcEngineImageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("seededit error: %v", result.Error)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no edited image returned")
	}

	return &ImageResult{
		Status:    "completed",
		ImageURL:  result.Data[0].URL,
		Completed: true,
	}, nil
}

func (c *SeededitClient) GetTaskStatus(taskID string) (*ImageResult, error) {
	return nil, fmt.Errorf("not supported for Seededit (synchronous generation)")
}

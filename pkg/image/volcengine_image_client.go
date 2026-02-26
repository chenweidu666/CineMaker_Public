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

type VolcEngineImageClient struct {
	BaseURL       string
	APIKey        string
	Model         string
	Endpoint      string
	QueryEndpoint string
	HTTPClient    *http.Client
}

// VolcEngineImageRequest 按官方文档 https://www.volcengine.com/docs/82379/1824121
type VolcEngineImageRequest struct {
	Model                            string                          `json:"model"`
	Prompt                           string                          `json:"prompt"`
	Image                            []string                        `json:"image,omitempty"`
	SequentialImageGeneration        string                          `json:"sequential_image_generation,omitempty"`
	SequentialImageGenerationOptions *SeqImageGenOptions              `json:"sequential_image_generation_options,omitempty"`
	ResponseFormat                   string                          `json:"response_format,omitempty"`
	Stream                           bool                            `json:"stream"`
	Size                             string                          `json:"size,omitempty"`
	Watermark                        bool                            `json:"watermark"`
	OptimizePromptOptions            *OptimizePromptOptions           `json:"optimize_prompt_options,omitempty"`
}

type SeqImageGenOptions struct {
	MaxImages int `json:"max_images,omitempty"`
}

// OptimizePromptOptions Seedream 4.0 支持 fast/standard，4.5 仅 standard
type OptimizePromptOptions struct {
	Mode string `json:"mode"` // "standard" or "fast"
}

type VolcEngineImageResponse struct {
	Model   string `json:"model"`
	Created int64  `json:"created"`
	Data    []struct {
		URL  string `json:"url"`
		Size string `json:"size"`
	} `json:"data"`
	Usage struct {
		GeneratedImages int `json:"generated_images"`
		OutputTokens    int `json:"output_tokens"`
		TotalTokens     int `json:"total_tokens"`
	} `json:"usage"`
	Error interface{} `json:"error,omitempty"`
}

func NewVolcEngineImageClient(baseURL, apiKey, model, endpoint, queryEndpoint string) *VolcEngineImageClient {
	if endpoint == "" {
		endpoint = "/api/v3/images/generations"
	}
	if queryEndpoint == "" {
		queryEndpoint = endpoint
	}
	return &VolcEngineImageClient{
		BaseURL:       baseURL,
		APIKey:        apiKey,
		Model:         model,
		Endpoint:      endpoint,
		QueryEndpoint: queryEndpoint,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

func (c *VolcEngineImageClient) GenerateImage(prompt string, opts ...ImageOption) (*ImageResult, error) {
	options := &ImageOptions{
		Size:    "2K",
		Quality: "standard",
	}

	for _, opt := range opts {
		opt(options)
	}

	model := c.Model
	if options.Model != "" {
		model = options.Model
	}

	jsonData, err := c.buildSeedreamRequest(model, prompt, options)
	if err != nil {
		return nil, err
	}

	url := c.BaseURL + c.Endpoint
	slog.Info("[VolcEngine Image] Sending request", "url", url, "model", model, "prompt_len", len(prompt))

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

	slog.Info("[VolcEngine Image] Response received", "status", resp.StatusCode, "body_len", len(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		if strings.Contains(string(body), "SensitiveContentDetected") ||
			strings.Contains(string(body), "sensitive") {
			return nil, fmt.Errorf("内容安全审核未通过：提示词触发了火山引擎安全策略，请修改后重试")
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result VolcEngineImageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("volcengine error: %v", result.Error)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no image generated")
	}

	return &ImageResult{
		Status:    "completed",
		ImageURL:  result.Data[0].URL,
		Completed: true,
	}, nil
}

// buildSeedreamRequest 构建 Seedream（文生图/图生图）请求体
func (c *VolcEngineImageClient) buildSeedreamRequest(model, prompt string, options *ImageOptions) ([]byte, error) {
	promptText := prompt
	if options.NegativePrompt != "" {
		promptText += fmt.Sprintf("。负面描述：%s", options.NegativePrompt)
	}

	size := resolveVolcEngineSize(options.Size, options.Width, options.Height)

	reqBody := VolcEngineImageRequest{
		Model:          model,
		Prompt:         promptText,
		ResponseFormat: "url",
		Stream:         false,
		Size:           size,
		Watermark:      false,
	}

	reqBody.OptimizePromptOptions = &OptimizePromptOptions{Mode: "standard"}

	if len(options.ReferenceImages) > 0 {
		reqBody.Image = options.ReferenceImages
		reqBody.SequentialImageGeneration = "disabled"
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal seedream request: %w", err)
	}
	return data, nil
}


func (c *VolcEngineImageClient) GetTaskStatus(taskID string) (*ImageResult, error) {
	return nil, fmt.Errorf("not supported for VolcEngine Seedream (synchronous generation)")
}

// TestConnection 测试 Seedream 图片 API 连通性，使用最小请求验证 API Key
func (c *VolcEngineImageClient) TestConnection() error {
	_, err := c.GenerateImage("a red circle", WithSize("1K"))
	return err
}

// resolveVolcEngineSize 处理 size 参数。
// 官方文档支持两种互斥方式：
//   - 分辨率档位: "1K", "2K", "4K"
//   - 像素值: "WxH" 如 "2048x2048"（范围 1280x720 ~ 4096x4096）
func resolveVolcEngineSize(sizeStr string, width, height int) string {
	// 已经是合法的分辨率档位
	if sizeStr == "1K" || sizeStr == "2K" || sizeStr == "4K" {
		return sizeStr
	}

	// 如果有明确的 width/height，用像素值方式
	if width > 0 && height > 0 {
		return fmt.Sprintf("%dx%d", width, height)
	}

	// 如果是 "WxH" 格式的字符串，直接透传
	if sizeStr != "" {
		var w, h int
		if _, err := fmt.Sscanf(sizeStr, "%dx%d", &w, &h); err == nil && w >= 720 && h >= 720 {
			return sizeStr
		}
	}

	return "2K"
}

package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAIClient struct {
	BaseURL    string
	APIKey     string
	Model      string
	Endpoint   string
	HTTPClient *http.Client
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MultimodalContentPart represents a single part in a multimodal message (text or image).
type MultimodalContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// MultimodalChatMessage supports both plain text and multimodal content arrays.
type MultimodalChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []MultimodalContentPart
}

type ChatCompletionRequest struct {
	Model               string        `json:"model"`
	Messages            []ChatMessage `json:"messages"`
	Temperature         float64       `json:"temperature,omitempty"`
	MaxTokens           *int          `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int          `json:"max_completion_tokens,omitempty"`
	TopP                float64       `json:"top_p,omitempty"`
	Stream              bool          `json:"stream,omitempty"`
}

// MultimodalChatCompletionRequest is like ChatCompletionRequest but accepts multimodal messages.
type MultimodalChatCompletionRequest struct {
	Model               string                  `json:"model"`
	Messages            []MultimodalChatMessage `json:"messages"`
	Temperature         float64                 `json:"temperature,omitempty"`
	MaxTokens           *int                    `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int                    `json:"max_completion_tokens,omitempty"`
	TopP                float64                 `json:"top_p,omitempty"`
	Stream              bool                    `json:"stream,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ImageGenerationRequest struct {
	Model  string `json:"model,omitempty"`
	Prompt string `json:"prompt"`
	N      int    `json:"n,omitempty"`
	Size   string `json:"size,omitempty"`
}

type ImageGenerationResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL     string `json:"url"`
		B64JSON string `json:"b64_json"`
	} `json:"data"`
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func NewOpenAIClient(baseURL, apiKey, model, endpoint string) *OpenAIClient {
	if endpoint == "" {
		endpoint = "/v1/chat/completions"
	}

	return &OpenAIClient{
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Model:    model,
		Endpoint: endpoint,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

func (c *OpenAIClient) ChatCompletion(messages []ChatMessage, options ...func(*ChatCompletionRequest)) (*ChatCompletionResponse, error) {
	req := &ChatCompletionRequest{
		Model:    c.Model,
		Messages: messages,
	}

	for _, option := range options {
		option(req)
	}

	return c.sendChatRequest(req)
}

func (c *OpenAIClient) sendChatRequest(req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	resp, err := c.doChatRequest(req)
	if err == nil {
		return resp, nil
	}

	if shouldRetryWithMaxCompletionTokens(err, req) {
		tokens := *req.MaxTokens
		retryReq := *req
		retryReq.MaxTokens = nil
		retryReq.MaxCompletionTokens = &tokens
		fmt.Printf("OpenAI: retrying with max_completion_tokens=%d\n", tokens)
		return c.doChatRequest(&retryReq)
	}

	return nil, err
}

func (c *OpenAIClient) doChatRequest(req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("OpenAI: Failed to marshal request: %v\n", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.BaseURL + c.Endpoint

	// 打印请求信息
	fmt.Printf("OpenAI: Sending request to: %s\n", url)
	fmt.Printf("OpenAI: BaseURL=%s, Endpoint=%s, Model=%s\n", c.BaseURL, c.Endpoint, c.Model)
	requestPreview := string(jsonData)
	if len(jsonData) > 300 {
		requestPreview = string(jsonData[:300]) + "..."
	}
	fmt.Printf("OpenAI: Request body: %s\n", requestPreview)

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("OpenAI: Failed to create request: %v\n", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	fmt.Printf("OpenAI: Executing HTTP request...\n")
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		fmt.Printf("OpenAI: HTTP request failed: %v\n", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("OpenAI: Received response with status: %d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("OpenAI: Failed to read response body: %v\n", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("OpenAI: API error (status %d): %s\n", resp.StatusCode, string(body))
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		// 内容安全审核拦截 — 返回友好提示
		if errResp.Error.Code == "data_inspection_failed" ||
			strings.Contains(errResp.Error.Message, "inappropriate content") {
			return nil, fmt.Errorf("⚠️ 内容安全审核未通过：当前剧本内容触发了阿里云 DashScope 的安全策略，请求被拒绝。\n\n💡 建议：请切换到「火山引擎」模型（如 doubao-1-5-pro），火山引擎对内容审核更宽松，可在 AI 配置中选择火山引擎的文本服务。")
		}
		return nil, fmt.Errorf("API error: %s", errResp.Error.Message)
	}

	// 打印响应体用于调试
	bodyPreview := string(body)
	if len(body) > 500 {
		bodyPreview = string(body[:500]) + "..."
	}
	fmt.Printf("OpenAI: Response body: %s\n", bodyPreview)

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		errorPreview := string(body)
		if len(body) > 200 {
			errorPreview = string(body[:200])
		}
		fmt.Printf("OpenAI: Failed to parse response: %v\n", err)
		return nil, fmt.Errorf("failed to unmarshal response: %w, body preview: %s", err, errorPreview)
	}

	fmt.Printf("OpenAI: Successfully parsed response, choices count: %d\n", len(chatResp.Choices))

	if len(chatResp.Choices) == 0 {
		fmt.Printf("OpenAI: No choices in response\n")
		return nil, fmt.Errorf("no choices in response")
	}

	// 检查 finish_reason，处理内容过滤的情况
	if len(chatResp.Choices) > 0 {
		finishReason := chatResp.Choices[0].FinishReason
		content := chatResp.Choices[0].Message.Content
		usage := chatResp.Usage

		fmt.Printf("OpenAI: finish_reason=%s, content_length=%d\n", finishReason, len(content))

		if finishReason == "content_filter" {
			return nil, fmt.Errorf("AI内容被安全过滤器拦截，可能因为：\n1. 请求内容触发了安全策略\n2. 生成的内容包含敏感信息\n3. 建议：调整输入内容或联系API提供商调整过滤策略")
		}

		if usage.TotalTokens == 0 && finishReason != "stop" {
			return nil, fmt.Errorf("AI返回内容为空 (finish_reason: %s)，可能的原因：\n1. 内容被过滤\n2. Token限制\n3. API异常", finishReason)
		}
	}

	return &chatResp, nil
}

func WithTemperature(temp float64) func(*ChatCompletionRequest) {
	return func(req *ChatCompletionRequest) {
		req.Temperature = temp
	}
}

func WithMaxTokens(tokens int) func(*ChatCompletionRequest) {
	return func(req *ChatCompletionRequest) {
		req.MaxTokens = &tokens
	}
}

func WithTopP(topP float64) func(*ChatCompletionRequest) {
	return func(req *ChatCompletionRequest) {
		req.TopP = topP
	}
}

func (c *OpenAIClient) GenerateText(prompt string, systemPrompt string, options ...func(*ChatCompletionRequest)) (string, *TokenUsage, error) {
	messages := []ChatMessage{}

	if systemPrompt != "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: prompt,
	})

	resp, err := c.ChatCompletion(messages, options...)
	if err != nil {
		return "", nil, err
	}

	if len(resp.Choices) == 0 {
		return "", nil, fmt.Errorf("no response from API")
	}

	usage := &TokenUsage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}

	return resp.Choices[0].Message.Content, usage, nil
}

func (c *OpenAIClient) GenerateImage(prompt string, size string, n int) ([]string, error) {
	// 图片生成端点通常是 /v1/images/generations
	// 如果 c.Endpoint 是 chat 端点，我们需要将其替换
	// 这是一个简单的处理逻辑，实际可能需要更复杂的配置
	imageEndpoint := "/v1/images/generations"

	// 如果 BaseURL 是类似 api.openai.com，那么直接拼接
	url := c.BaseURL + imageEndpoint

	reqBody := ImageGenerationRequest{
		Prompt: prompt,
		N:      n,
		Size:   size,
		Model:  c.Model, // 如果是DALL-E 3，模型名很重要
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			// 内容安全审核拦截 — 返回友好提示
			if errResp.Error.Code == "data_inspection_failed" ||
				strings.Contains(errResp.Error.Message, "inappropriate content") {
				return nil, fmt.Errorf("⚠️ 内容安全审核未通过：图片生成提示词触发了阿里云的安全策略，请求被拒绝。\n\n💡 建议：请切换到「火山引擎」图片模型（如 Seedream），可在 AI 配置中选择火山引擎的图片服务。")
			}
			return nil, fmt.Errorf("API error: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var imgResp ImageGenerationResponse
	if err := json.Unmarshal(body, &imgResp); err != nil {
		return nil, err
	}

	var urls []string
	for _, data := range imgResp.Data {
		if data.URL != "" {
			urls = append(urls, data.URL)
		} else if data.B64JSON != "" {
			// 如果返回的是base64，添加前缀
			urls = append(urls, "data:image/png;base64,"+data.B64JSON)
		}
	}

	return urls, nil
}

func (c *OpenAIClient) TestConnection() error {
	fmt.Printf("OpenAI: TestConnection called with BaseURL=%s, Endpoint=%s, Model=%s\n", c.BaseURL, c.Endpoint, c.Model)

	_, _, err := c.GenerateText("Hello", "", WithMaxTokens(50))
	if err != nil {
		fmt.Printf("OpenAI: TestConnection failed: %v\n", err)
	} else {
		fmt.Printf("OpenAI: TestConnection succeeded\n")
	}
	return err
}

func shouldRetryWithMaxCompletionTokens(err error, req *ChatCompletionRequest) bool {
	if err == nil || req == nil || req.MaxTokens == nil || req.MaxCompletionTokens != nil {
		return false
	}

	msg := err.Error()
	if strings.Contains(msg, "Unsupported parameter: 'max_tokens'") {
		return true
	}
	if strings.Contains(msg, "max_tokens is not supported") {
		return true
	}
	if strings.Contains(msg, "max_completion_tokens") {
		return true
	}
	return false
}

// Volcengine Responses API types (for doubao-seed-1.6-vision etc.)
type responsesContentPart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

type responsesMessage struct {
	Role    string                 `json:"role"`
	Content []responsesContentPart `json:"content"`
}

type responsesRequest struct {
	Model string             `json:"model"`
	Input []responsesMessage `json:"input"`
}

type responsesOutputItem struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Status  string `json:"status"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

type responsesResponse struct {
	ID     string                `json:"id"`
	Output []responsesOutputItem `json:"output"`
	Usage  struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// isResponsesAPI returns true if the endpoint is the Volcengine Responses API format.
func (c *OpenAIClient) isResponsesAPI() bool {
	return strings.Contains(c.Endpoint, "responses")
}

// GenerateTextWithImages sends a multimodal request with text and base64-encoded images.
// imageDataURLs should be data URIs like "data:image/jpeg;base64,..." or plain URLs.
// Automatically detects and uses either OpenAI chat/completions or Volcengine Responses API format.
func (c *OpenAIClient) GenerateTextWithImages(prompt string, imageDataURLs []string, systemPrompt string, maxTokens int) (string, *TokenUsage, error) {
	if c.isResponsesAPI() {
		return c.generateWithResponsesAPI(prompt, imageDataURLs, systemPrompt)
	}
	return c.generateWithChatCompletionsVision(prompt, imageDataURLs, systemPrompt, maxTokens)
}

// generateWithResponsesAPI uses the Volcengine /responses endpoint format.
func (c *OpenAIClient) generateWithResponsesAPI(prompt string, imageDataURLs []string, systemPrompt string) (string, *TokenUsage, error) {
	var input []responsesMessage

	if systemPrompt != "" {
		input = append(input, responsesMessage{
			Role:    "system",
			Content: []responsesContentPart{{Type: "input_text", Text: systemPrompt}},
		})
	}

	var parts []responsesContentPart
	for _, imgURL := range imageDataURLs {
		parts = append(parts, responsesContentPart{Type: "input_image", ImageURL: imgURL})
	}
	parts = append(parts, responsesContentPart{Type: "input_text", Text: prompt})

	input = append(input, responsesMessage{Role: "user", Content: parts})

	reqBody := responsesRequest{Model: c.Model, Input: input}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("marshal responses request: %w", err)
	}

	url := c.BaseURL + c.Endpoint
	fmt.Printf("Vision (Responses API): %s model=%s images=%d\n", url, c.Model, len(imageDataURLs))

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return "", nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		preview := string(body)
		if len(preview) > 500 {
			preview = preview[:500]
		}
		return "", nil, fmt.Errorf("Vision API error (status %d): %s", resp.StatusCode, preview)
	}

	var rr responsesResponse
	if err := json.Unmarshal(body, &rr); err != nil {
		preview := string(body)
		if len(preview) > 300 {
			preview = preview[:300]
		}
		return "", nil, fmt.Errorf("unmarshal responses: %w, raw: %s", err, preview)
	}

	if rr.Error != nil && rr.Error.Message != "" {
		return "", nil, fmt.Errorf("API error: [%s] %s", rr.Error.Code, rr.Error.Message)
	}

	fmt.Printf("Vision (Responses API): output items=%d\n", len(rr.Output))

	var text string
	for _, item := range rr.Output {
		for _, part := range item.Content {
			if part.Type == "output_text" || part.Type == "text" {
				text += part.Text
			}
		}
	}
	if text == "" {
		bodyPreview := string(body)
		if len(bodyPreview) > 300 {
			bodyPreview = bodyPreview[:300]
		}
		return "", nil, fmt.Errorf("no text content in responses output, raw: %s", bodyPreview)
	}

	usage := &TokenUsage{
		PromptTokens:     rr.Usage.InputTokens,
		CompletionTokens: rr.Usage.OutputTokens,
		TotalTokens:      rr.Usage.TotalTokens,
	}

	return text, usage, nil
}

// generateWithChatCompletionsVision uses the OpenAI-compatible chat/completions format.
func (c *OpenAIClient) generateWithChatCompletionsVision(prompt string, imageDataURLs []string, systemPrompt string, maxTokens int) (string, *TokenUsage, error) {
	var messages []MultimodalChatMessage

	if systemPrompt != "" {
		messages = append(messages, MultimodalChatMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	contentParts := []MultimodalContentPart{
		{Type: "text", Text: prompt},
	}
	for _, imgURL := range imageDataURLs {
		contentParts = append(contentParts, MultimodalContentPart{
			Type:     "image_url",
			ImageURL: &ImageURL{URL: imgURL, Detail: "auto"},
		})
	}

	messages = append(messages, MultimodalChatMessage{
		Role:    "user",
		Content: contentParts,
	})

	reqBody := MultimodalChatCompletionRequest{
		Model:    c.Model,
		Messages: messages,
	}
	if maxTokens > 0 {
		reqBody.MaxTokens = &maxTokens
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("marshal multimodal request: %w", err)
	}

	url := c.BaseURL + c.Endpoint
	fmt.Printf("Vision (Chat API): %s model=%s images=%d\n", url, c.Model, len(imageDataURLs))

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return "", nil, fmt.Errorf("send multimodal request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		preview := string(body)
		if len(preview) > 500 {
			preview = preview[:500]
		}
		return "", nil, fmt.Errorf("Vision API error (status %d): %s", resp.StatusCode, preview)
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", nil, fmt.Errorf("unmarshal vision response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", nil, fmt.Errorf("no choices in vision response")
	}

	usage := &TokenUsage{
		PromptTokens:     chatResp.Usage.PromptTokens,
		CompletionTokens: chatResp.Usage.CompletionTokens,
		TotalTokens:      chatResp.Usage.TotalTokens,
	}

	return chatResp.Choices[0].Message.Content, usage, nil
}

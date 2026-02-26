package ai

// TokenUsage holds token consumption from an AI API response.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIClient 定义文本生成客户端接口
type AIClient interface {
	GenerateText(prompt string, systemPrompt string, options ...func(*ChatCompletionRequest)) (string, *TokenUsage, error)
	GenerateImage(prompt string, size string, n int) ([]string, error)
	TestConnection() error
}

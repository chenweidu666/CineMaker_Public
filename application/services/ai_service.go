package services

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cinemaker/backend/domain"
	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/ai"
	"github.com/cinemaker/backend/pkg/logger"
	"gorm.io/gorm"
)

type AIService struct {
	db            *gorm.DB
	log           *logger.Logger
	msgLogService *AIMessageLogService
}

func NewAIService(db *gorm.DB, log *logger.Logger) *AIService {
	return &AIService{
		db:            db,
		log:           log,
		msgLogService: NewAIMessageLogService(db, log),
	}
}

func (s *AIService) GetMessageLogService() *AIMessageLogService {
	return s.msgLogService
}

type CreateAIConfigRequest struct {
	ServiceType   string            `json:"service_type" binding:"required,oneof=text image video"`
	Name          string            `json:"name" binding:"required,min=1,max=100"`
	Provider      string            `json:"provider" binding:"required"`
	BaseURL       string            `json:"base_url" binding:"required,url"`
	APIKey        string            `json:"api_key" binding:"required"`
	Model         models.ModelField `json:"model" binding:"required"`
	Endpoint      string            `json:"endpoint"`
	QueryEndpoint string            `json:"query_endpoint"`
	Priority      int               `json:"priority"`
	IsDefault     bool              `json:"is_default"`
	Settings      string            `json:"settings"`
}

type UpdateAIConfigRequest struct {
	Name          string             `json:"name" binding:"omitempty,min=1,max=100"`
	Provider      string             `json:"provider"`
	BaseURL       string             `json:"base_url" binding:"omitempty,url"`
	APIKey        string             `json:"api_key"`
	Model         *models.ModelField `json:"model"`
	Endpoint      string             `json:"endpoint"`
	QueryEndpoint string             `json:"query_endpoint"`
	Priority      *int               `json:"priority"`
	IsDefault     bool               `json:"is_default"`
	IsActive      bool               `json:"is_active"`
	Settings      string             `json:"settings"`
}

type TestConnectionRequest struct {
	BaseURL  string            `json:"base_url" binding:"required,url"`
	APIKey   string            `json:"api_key" binding:"required"`
	Model    models.ModelField `json:"model" binding:"required"`
	Provider string            `json:"provider"`
	Endpoint string            `json:"endpoint"`
}

func (s *AIService) CreateConfig(req *CreateAIConfigRequest) (*models.AIServiceConfig, error) {
	// 根据 provider 和 service_type 自动设置 endpoint
	endpoint := req.Endpoint
	queryEndpoint := req.QueryEndpoint

	if endpoint == "" {
		switch req.Provider {
		case "gemini", "google":
			if req.ServiceType == "text" {
				endpoint = "/v1beta/models/{model}:generateContent"
			} else if req.ServiceType == "image" {
				endpoint = "/v1beta/models/{model}:generateContent"
			}
		case "openai":
			if req.ServiceType == "text" {
				endpoint = "/chat/completions"
			} else if req.ServiceType == "image" {
				endpoint = "/images/generations"
			} else if req.ServiceType == "video" {
				endpoint = "/videos"
				if queryEndpoint == "" {
					queryEndpoint = "/videos/{taskId}"
				}
			}
		case "doubao", "volcengine", "volces":
			if req.ServiceType == "video" {
				endpoint = "/contents/generations/tasks"
				if queryEndpoint == "" {
					queryEndpoint = "/generations/tasks/{taskId}"
				}
			}
		default:
			// 默认使用 OpenAI 格式
			if req.ServiceType == "text" {
				endpoint = "/chat/completions"
			} else if req.ServiceType == "image" {
				endpoint = "/images/generations"
			}
		}
	}

	config := &models.AIServiceConfig{
		ServiceType:   req.ServiceType,
		Name:          req.Name,
		Provider:      req.Provider,
		BaseURL:       req.BaseURL,
		APIKey:        req.APIKey,
		Model:         req.Model,
		Endpoint:      endpoint,
		QueryEndpoint: queryEndpoint,
		Priority:      req.Priority,
		IsDefault:     req.IsDefault,
		IsActive:      true,
		Settings:      req.Settings,
	}

	if err := s.db.Create(config).Error; err != nil {
		s.log.Errorw("Failed to create AI config", "error", err)
		return nil, err
	}

	s.log.Infow("AI config created", "config_id", config.ID, "provider", req.Provider, "endpoint", endpoint)
	return config, nil
}

func (s *AIService) GetConfig(configID uint) (*models.AIServiceConfig, error) {
	var config models.AIServiceConfig
	err := s.db.Where("id = ? ", configID).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrConfigNotFound
		}
		return nil, err
	}
	return &config, nil
}

func (s *AIService) ListConfigs(serviceType string) ([]models.AIServiceConfig, error) {
	var configs []models.AIServiceConfig
	query := s.db

	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}

	err := query.Order("priority DESC, created_at DESC").Find(&configs).Error
	if err != nil {
		s.log.Errorw("Failed to list AI configs", "error", err)
		return nil, err
	}

	return configs, nil
}

func (s *AIService) UpdateConfig(configID uint, req *UpdateAIConfigRequest) (*models.AIServiceConfig, error) {
	var config models.AIServiceConfig
	if err := s.db.Where("id = ? ", configID).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrConfigNotFound
		}
		return nil, err
	}

	tx := s.db.Begin()

	// 不再需要is_default独占逻辑

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Provider != "" {
		updates["provider"] = req.Provider
	}
	if req.BaseURL != "" {
		updates["base_url"] = req.BaseURL
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	if req.Model != nil && len(*req.Model) > 0 {
		updates["model"] = *req.Model
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}

	// 如果提供了 provider，根据 provider 和 service_type 自动设置 endpoint
	if req.Provider != "" && req.Endpoint == "" {
		provider := req.Provider
		serviceType := config.ServiceType

		switch provider {
		case "gemini", "google":
			if serviceType == "text" || serviceType == "image" {
				updates["endpoint"] = "/v1beta/models/{model}:generateContent"
			}
		case "openai":
			if serviceType == "text" {
				updates["endpoint"] = "/chat/completions"
			} else if serviceType == "image" {
				updates["endpoint"] = "/images/generations"
			} else if serviceType == "video" {
				updates["endpoint"] = "/videos"
				updates["query_endpoint"] = "/videos/{taskId}"
			}
		}
	} else if req.Endpoint != "" {
		updates["endpoint"] = req.Endpoint
	}

	// 允许清空query_endpoint，所以不检查是否为空
	updates["query_endpoint"] = req.QueryEndpoint
	if req.Settings != "" {
		updates["settings"] = req.Settings
	}
	updates["is_default"] = req.IsDefault
	updates["is_active"] = req.IsActive

	if err := tx.Model(&config).Updates(updates).Error; err != nil {
		tx.Rollback()
		s.log.Errorw("Failed to update AI config", "error", err)
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	s.log.Infow("AI config updated", "config_id", configID)
	return &config, nil
}

func (s *AIService) DeleteConfig(configID uint) error {
	result := s.db.Where("id = ? ", configID).Delete(&models.AIServiceConfig{})

	if result.Error != nil {
		s.log.Errorw("Failed to delete AI config", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrConfigNotFound
	}

	s.log.Infow("AI config deleted", "config_id", configID)
	return nil
}

func (s *AIService) TestConnection(req *TestConnectionRequest) error {
	s.log.Infow("TestConnection called", "baseURL", req.BaseURL, "provider", req.Provider, "endpoint", req.Endpoint, "modelCount", len(req.Model))

	// 使用第一个模型进行测试
	model := ""
	if len(req.Model) > 0 {
		model = req.Model[0]
	}
	s.log.Infow("Using model for test", "model", model, "provider", req.Provider)

	// 根据 provider 参数选择客户端
	var client ai.AIClient
	var endpoint string

	switch req.Provider {
	case "gemini", "google":
		// Gemini
		s.log.Infow("Using Gemini client", "baseURL", req.BaseURL)
		endpoint = "/v1beta/models/{model}:generateContent"
		client = ai.NewGeminiClient(req.BaseURL, req.APIKey, model, endpoint)
	case "openai":
		// OpenAI 格式
		s.log.Infow("Using OpenAI-compatible client", "baseURL", req.BaseURL, "provider", req.Provider)
		endpoint = req.Endpoint
		if endpoint == "" {
			endpoint = "/chat/completions"
		}
		client = ai.NewOpenAIClient(req.BaseURL, req.APIKey, model, endpoint)
	default:
		// 默认使用 OpenAI 格式
		s.log.Infow("Using default OpenAI-compatible client", "baseURL", req.BaseURL)
		endpoint = req.Endpoint
		if endpoint == "" {
			endpoint = "/chat/completions"
		}
		client = ai.NewOpenAIClient(req.BaseURL, req.APIKey, model, endpoint)
	}

	s.log.Infow("Calling TestConnection on client", "endpoint", endpoint)
	err := client.TestConnection()
	if err != nil {
		s.log.Errorw("TestConnection failed", "error", err)
	} else {
		s.log.Infow("TestConnection succeeded")
	}
	return err
}

func (s *AIService) GetDefaultConfig(serviceType string) (*models.AIServiceConfig, error) {
	// Priority: environment variables > database
	if cfg := s.configFromEnv(serviceType); cfg != nil {
		return cfg, nil
	}

	var config models.AIServiceConfig
	err := s.db.Where("service_type = ? AND is_active = ?", serviceType, true).
		Order("priority DESC, created_at DESC").
		First(&config).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no active %s config found: set %s_API_KEY env var or add config in settings",
				serviceType, strings.ToUpper(serviceType))
		}
		return nil, err
	}

	return &config, nil
}

// configFromEnv builds an AIServiceConfig from environment variables.
// Env var pattern: {TYPE}_API_KEY, {TYPE}_BASE_URL, {TYPE}_MODEL, {TYPE}_PROVIDER
// e.g. TEXT_API_KEY, VISION_BASE_URL, IMAGE_MODEL
func (s *AIService) configFromEnv(serviceType string) *models.AIServiceConfig {
	prefix := strings.ToUpper(serviceType)
	apiKey := os.Getenv(prefix + "_API_KEY")
	if apiKey == "" {
		return nil
	}

	baseURL := os.Getenv(prefix + "_BASE_URL")
	model := os.Getenv(prefix + "_MODEL")
	provider := os.Getenv(prefix + "_PROVIDER")
	endpoint := os.Getenv(prefix + "_ENDPOINT")

	if provider == "" {
		provider = "openai"
	}
	if baseURL == "" {
		switch provider {
		case "doubao", "volcengine":
			baseURL = "https://ark.cn-beijing.volces.com/api"
		default:
			baseURL = "https://api.openai.com"
		}
	}

	var modelField models.ModelField
	if model != "" {
		modelField = models.ModelField{model}
	}

	return &models.AIServiceConfig{
		ServiceType: serviceType,
		Provider:    provider,
		Name:        prefix + " (env)",
		BaseURL:     baseURL,
		APIKey:      apiKey,
		Model:       modelField,
		Endpoint:    endpoint,
		IsActive:    true,
		Priority:    9999,
	}
}

// GetConfigForModel 根据服务类型和模型名称获取优先级最高的激活配置
func (s *AIService) GetConfigForModel(serviceType string, modelName string) (*models.AIServiceConfig, error) {
	var configs []models.AIServiceConfig
	err := s.db.Where("service_type = ? AND is_active = ?", serviceType, true).
		Order("priority DESC, created_at DESC").
		Find(&configs).Error

	if err != nil {
		return nil, err
	}

	// 查找包含指定模型的配置
	for _, config := range configs {
		for _, model := range config.Model {
			if model == modelName {
				return &config, nil
			}
		}
	}

	return nil, errors.New("no active config found for model: " + modelName)
}

func (s *AIService) GetAIClient(serviceType string) (ai.AIClient, error) {
	config, err := s.GetDefaultConfig(serviceType)
	if err != nil {
		return nil, err
	}

	// 使用第一个模型
	model := ""
	if len(config.Model) > 0 {
		model = config.Model[0]
	}

	// 使用数据库配置中的 endpoint，如果为空则根据 provider 设置默认值
	endpoint := config.Endpoint
	if endpoint == "" {
		switch config.Provider {
		case "gemini", "google":
			endpoint = "/v1beta/models/{model}:generateContent"
		default:
			endpoint = "/chat/completions"
		}
	}

	// 根据 provider 创建对应的客户端
	switch config.Provider {
	case "gemini", "google":
		return ai.NewGeminiClient(config.BaseURL, config.APIKey, model, endpoint), nil
	default:
		// openai 等其他厂商都使用 OpenAI 格式
		return ai.NewOpenAIClient(config.BaseURL, config.APIKey, model, endpoint), nil
	}
}

// GetAIClientForModel 根据服务类型和模型名称获取对应的AI客户端
func (s *AIService) GetAIClientForModel(serviceType string, modelName string) (ai.AIClient, error) {
	config, err := s.GetConfigForModel(serviceType, modelName)
	if err != nil {
		return nil, err
	}

	// 使用数据库配置中的 endpoint，如果为空则根据 provider 设置默认值
	endpoint := config.Endpoint
	if endpoint == "" {
		switch config.Provider {
		case "gemini", "google":
			endpoint = "/v1beta/models/{model}:generateContent"
		default:
			endpoint = "/chat/completions"
		}
	}

	// 根据 provider 创建对应的客户端
	switch config.Provider {
	case "gemini", "google":
		return ai.NewGeminiClient(config.BaseURL, config.APIKey, modelName, endpoint), nil
	default:
		// openai 等其他厂商都使用 OpenAI 格式
		return ai.NewOpenAIClient(config.BaseURL, config.APIKey, modelName, endpoint), nil
	}
}

func (s *AIService) GenerateText(prompt string, systemPrompt string, options ...func(*ai.ChatCompletionRequest)) (string, error) {
	return s.GenerateTextWithLog(prompt, systemPrompt, "general", nil, options...)
}

// GenerateTextWithLog calls GenerateText and records an AI message log.
func (s *AIService) GenerateTextWithLog(prompt string, systemPrompt string, purpose string, dramaID *uint, options ...func(*ai.ChatCompletionRequest)) (string, error) {
	config, _ := s.GetDefaultConfig("text")
	provider, model := "", ""
	if config != nil {
		provider = config.Provider
		if len(config.Model) > 0 {
			model = config.Model[0]
		}
	}

	logID := s.msgLogService.LogRequest(LogEntry{
		DramaID:      dramaID,
		ServiceType:  "text",
		Purpose:      purpose,
		Provider:     provider,
		Model:        model,
		SystemPrompt: systemPrompt,
		UserPrompt:   prompt,
	})

	start := time.Now()
	client, err := s.GetAIClient("text")
	if err != nil {
		s.msgLogService.UpdateFailed(logID, err.Error(), time.Since(start).Milliseconds())
		return "", fmt.Errorf("failed to get AI client: %w", err)
	}

	result, usage, err := client.GenerateText(prompt, systemPrompt, options...)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		s.msgLogService.UpdateFailed(logID, err.Error(), elapsed)
		return "", err
	}

	promptTokens, outputTokens := 0, 0
	if usage != nil {
		promptTokens = usage.PromptTokens
		outputTokens = usage.CompletionTokens
	}
	s.msgLogService.UpdateSuccess(logID, result, elapsed, promptTokens, outputTokens)
	return result, nil
}

// GenerateTextForModel calls GenerateText with a specific model and records a log.
func (s *AIService) GenerateTextForModel(prompt string, systemPrompt string, model string, purpose string, dramaID *uint, options ...func(*ai.ChatCompletionRequest)) (string, error) {
	if model == "" {
		return s.GenerateTextWithLog(prompt, systemPrompt, purpose, dramaID, options...)
	}

	client, err := s.GetAIClientForModel("text", model)
	if err != nil {
		s.log.Warnw("Failed to get client for model, falling back to default", "model", model, "error", err)
		return s.GenerateTextWithLog(prompt, systemPrompt, purpose, dramaID, options...)
	}

	config, _ := s.GetConfigForModel("text", model)
	provider := ""
	if config != nil {
		provider = config.Provider
	}

	logID := s.msgLogService.LogRequest(LogEntry{
		DramaID:      dramaID,
		ServiceType:  "text",
		Purpose:      purpose,
		Provider:     provider,
		Model:        model,
		SystemPrompt: systemPrompt,
		UserPrompt:   prompt,
	})

	start := time.Now()
	result, usage, err := client.GenerateText(prompt, systemPrompt, options...)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		s.msgLogService.UpdateFailed(logID, err.Error(), elapsed)
		return "", err
	}
	promptTokens, outputTokens := 0, 0
	if usage != nil {
		promptTokens = usage.PromptTokens
		outputTokens = usage.CompletionTokens
	}
	s.msgLogService.UpdateSuccess(logID, result, elapsed, promptTokens, outputTokens)
	return result, nil
}

func (s *AIService) GenerateImage(prompt string, size string, n int) ([]string, error) {
	client, err := s.GetAIClient("image")
	if err != nil {
		return nil, fmt.Errorf("failed to get AI client for image: %w", err)
	}

	return client.GenerateImage(prompt, size, n)
}

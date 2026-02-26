package services

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/cinemaker/backend/pkg/ai"
	"github.com/cinemaker/backend/pkg/logger"
)

type PromptTranslator struct {
	aiService *AIService
	log       *logger.Logger
}

func NewPromptTranslator(aiService *AIService, log *logger.Logger) *PromptTranslator {
	return &PromptTranslator{
		aiService: aiService,
		log:       log,
	}
}

const translateSystemPrompt = `You are a professional prompt translator for AI image and video generation.

Your task: Translate the user's Chinese prompt into English, optimized for AI image/video generation models (like DALL-E, Midjourney, Stable Diffusion, VolcEngine Seedream, etc.).

Rules:
1. Translate ALL Chinese text to English accurately.
2. Preserve any existing English text, technical terms, style keywords, and formatting as-is.
3. Keep special tokens like "imageRatio:16:9", negative prompts, and parameter-like syntax unchanged.
4. Maintain the artistic intent and visual description quality.
5. Output ONLY the translated prompt text, nothing else. No explanations, no markdown, no quotes.
6. If the input is already fully in English, return it unchanged.
7. Preserve line breaks and paragraph structure.`

// TranslatePrompt translates a Chinese prompt to English using LLM.
// Returns the original prompt if translation fails or prompt is already English.
func (t *PromptTranslator) TranslatePrompt(prompt string) (string, error) {
	if prompt == "" {
		return prompt, nil
	}

	if !containsChinese(prompt) {
		t.log.Debugw("Prompt is already in English, skipping translation", "prompt_length", len(prompt))
		return prompt, nil
	}

	t.log.Infow("Translating prompt to English",
		"original_length", len(prompt),
		"preview", truncateString(prompt, 100))

	translated, err := t.aiService.GenerateText(
		prompt,
		translateSystemPrompt,
		ai.WithMaxTokens(4096),
		ai.WithTemperature(0.3),
	)
	if err != nil {
		return "", fmt.Errorf("prompt translation failed: %w", err)
	}

	translated = strings.TrimSpace(translated)
	if translated == "" {
		return "", fmt.Errorf("translation returned empty result")
	}

	t.log.Infow("Prompt translated successfully",
		"original_length", len(prompt),
		"translated_length", len(translated),
		"preview", truncateString(translated, 100))

	return translated, nil
}

// TranslatePromptWithFallback translates the prompt, falling back to the original on error.
func (t *PromptTranslator) TranslatePromptWithFallback(prompt string) string {
	translated, err := t.TranslatePrompt(prompt)
	if err != nil {
		t.log.Warnw("Prompt translation failed, using original Chinese prompt",
			"error", err,
			"prompt_preview", truncateString(prompt, 80))
		return prompt
	}
	return translated
}

func containsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

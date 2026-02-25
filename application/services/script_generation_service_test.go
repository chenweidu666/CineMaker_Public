package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAIRejectionResponse(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "normal script",
			text:     "这是一个正常的剧本内容，包含了很多对话和场景描述...",
			expected: false,
		},
		{
			name:     "long text",
			text:     "这是一个很长的文本，超过了500个字符。" + string(make([]byte, 600)),
			expected: false,
		},
		{
			name:     "rejection with inappropriate",
			text:     "这个请求不适当，我无法为你完成",
			expected: true,
		},
		{
			name:     "rejection with moral",
			text:     "这违反了公序良俗，我不能创作",
			expected: true,
		},
		{
			name:     "rejection with cannot",
			text:     "我不能按照你的要求完成这个任务",
			expected: true,
		},
		{
			name:     "rejection with AI",
			text:     "作为一个AI，我无法协助这个请求",
			expected: true,
		},
		{
			name:     "rejection with sensitive",
			text:     "这涉及敏感内容，无法满足你的要求",
			expected: true,
		},
		{
			name:     "short normal text",
			text:     "好的",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAIRejectionResponse(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMinInt(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a is smaller",
			a:        5,
			b:        10,
			expected: 5,
		},
		{
			name:     "b is smaller",
			a:        10,
			b:        5,
			expected: 5,
		},
		{
			name:     "equal values",
			a:        5,
			b:        5,
			expected: 5,
		},
		{
			name:     "negative values",
			a:        -10,
			b:        -5,
			expected: -10,
		},
		{
			name:     "zero values",
			a:        0,
			b:        0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := minInt(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

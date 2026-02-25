package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractInitialPose(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		expected string
	}{
		{
			name:     "simple action",
			action:   "站立",
			expected: "站立",
		},
		{
			name:     "action with process word",
			action:   "站立然后坐下",
			expected: "站立",
		},
		{
			name:     "action with multiple process words",
			action:   "站立接着向前走",
			expected: "站立",
		},
		{
			name:     "action with punctuation",
			action:   "站立，然后坐下",
			expected: "站立",
		},
		{
			name:     "complex action",
			action:   "双手交叉放在胸前，然后慢慢放下",
			expected: "双手交叉放在胸前",
		},
		{
			name:     "action with direction",
			action:   "站立，然后向左转",
			expected: "站立",
		},
		{
			name:     "action with speed",
			action:   "站立然后向后倒",
			expected: "站立",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractInitialPose(tt.action)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractSimpleLocation(t *testing.T) {
	tests := []struct {
		name     string
		location string
		expected string
	}{
		{
			name:     "simple location",
			location: "客厅",
			expected: "客厅",
		},
		{
			name:     "location with dot separator",
			location: "客厅·沙发",
			expected: "客厅",
		},
		{
			name:     "location with Chinese comma",
			location: "客厅，沙发",
			expected: "客厅",
		},
		{
			name:     "location with English comma",
			location: "客厅,沙发",
			expected: "客厅",
		},
		{
			name:     "long location",
			location: "这是一个非常长的场景描述超过了15个字符",
			expected: "这是一个非",
		},
		{
			name:     "location with spaces",
			location: " 客厅 ",
			expected: "客厅",
		},
		{
			name:     "empty location",
			location: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSimpleLocation(tt.location)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractSimplePose(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		expected string
	}{
		{
			name:     "short action",
			action:   "站立",
			expected: "站立",
		},
		{
			name:     "long action",
			action:   "这是一个非常长的动作描述超过了十个字符",
			expected: "这是一个非常长的动作",
		},
		{
			name:     "action with spaces",
			action:   " 站立 ",
			expected: "站立",
		},
		{
			name:     "empty action",
			action:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSimplePose(tt.action)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFirstFramePose(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		expected string
	}{
		{
			name:     "simple action",
			action:   "站立",
			expected: "站立",
		},
		{
			name:     "action with process words",
			action:   "站立然后坐下",
			expected: "站立",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFirstFramePose(tt.action)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractCompositionType(t *testing.T) {
	tests := []struct {
		name     string
		shotType string
		expected string
	}{
		{
			name:     "close-up shot",
			shotType: "特写",
			expected: "特写",
		},
		{
			name:     "medium shot",
			shotType: "中景",
			expected: "中景",
		},
		{
			name:     "long shot",
			shotType: "远景",
			expected: "远景",
		},
		{
			name:     "empty shot type",
			shotType: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCompositionType(tt.shotType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeString(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "nil pointer",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty string",
			input:    func() *string { s := ""; return &s }(),
			expected: "",
		},
		{
			name:     "normal string",
			input:    func() *string { s := "hello"; return &s }(),
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMaxTokensForModel(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected int
	}{
		{
			name:     "GPT-4",
			model:    "gpt-4",
			expected: 16384,
		},
		{
			name:     "GPT-3.5",
			model:    "gpt-3.5-turbo",
			expected: 16384,
		},
		{
			name:     "Doubao model",
			model:    "doubao-pro",
			expected: 12288,
		},
		{
			name:     "Lite model",
			model:    "gpt-4-lite",
			expected: 8192,
		},
		{
			name:     "unknown model",
			model:    "unknown-model",
			expected: 16384,
		},
		{
			name:     "empty model",
			model:    "",
			expected: 16384,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMaxTokensForModel(tt.model)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMin(t *testing.T) {
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
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "nil pointer",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty string",
			input:    func() *string { s := ""; return &s }(),
			expected: "",
		},
		{
			name:     "normal string",
			input:    func() *string { s := "hello"; return &s }(),
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

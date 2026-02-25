package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStringValue(t *testing.T) {
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
		{
			name:     "string with spaces",
			input:    func() *string { s := "hello world"; return &s }(),
			expected: "hello world",
		},
		{
			name:     "string with special characters",
			input:    func() *string { s := "你好世界"; return &s }(),
			expected: "你好世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

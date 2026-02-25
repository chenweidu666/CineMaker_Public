package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateImageURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "empty URL",
			url:      "",
			expected: "",
		},
		{
			name:     "short normal URL",
			url:      "http://example.com/image.jpg",
			expected: "http://example.com/image.jpg",
		},
		{
			name:     "long normal URL",
			url:      "http://example.com/" + "a" + string(make([]byte, 150)),
			expected: "http://example.com/a" + string(make([]byte, 100)) + "...",
		},
		{
			name:     "short data URI",
			url:      "data:image/png;base64,short",
			expected: "data:image/png;base64,short",
		},
		{
			name:     "long data URI",
			url:      "data:image/png;base64," + "a" + string(make([]byte, 100)),
			expected: "data:image/png;base64,a" + string(make([]byte, 50)) + "...[base64 data]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateImageURL(tt.url)
			if len(tt.url) > 100 && !strings.HasPrefix(tt.url, "data:") {
				assert.Equal(t, 103, len(result))
				assert.True(t, strings.HasSuffix(result, "..."))
			} else if len(tt.url) > 50 && strings.HasPrefix(tt.url, "data:") {
				assert.Equal(t, 66, len(result))
				assert.True(t, strings.HasSuffix(result, "...[base64 data]"))
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

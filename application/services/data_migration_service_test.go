package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "URL with query parameters",
			url:      "https://example.com/image.jpg?width=800&height=600",
			expected: ".jpg",
		},
		{
			name:     "URL with fragment",
			url:      "https://example.com/image.png#section",
			expected: ".png",
		},
		{
			name:     "URL with both query and fragment",
			url:      "https://example.com/image.gif?param=value#section",
			expected: ".gif",
		},
		{
			name:     "URL without extension",
			url:      "https://example.com/image",
			expected: ".jpg",
		},
		{
			name:     "URL with uppercase extension",
			url:      "https://example.com/image.JPG",
			expected: ".jpg",
		},
		{
			name:     "URL with mixed case extension",
			url:      "https://example.com/image.PnG",
			expected: ".png",
		},
		{
			name:     "Local path with extension",
			url:      "/static/images/photo.jpg",
			expected: ".jpg",
		},
		{
			name:     "WebP extension",
			url:      "https://example.com/image.webp",
			expected: ".webp",
		},
		{
			name:     "Extension too long",
			url:      "https://example.com/image.verylongextensionthatshouldbetruncated",
			expected: ".jpg",
		},
		{
			name:     "Simple URL",
			url:      "https://example.com/image.jpeg",
			expected: ".jpeg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DataMigrationService{}
			result := s.extractFileExtension(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

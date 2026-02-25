package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectImageMimeType(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "PNG image",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x00},
			expected: "image/png",
		},
		{
			name:     "JPEG image",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x00, 0x00, 0x00, 0x00},
			expected: "image/jpeg",
		},
		{
			name:     "GIF image",
			data:     []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: "image/gif",
		},
		{
			name:     "WebP image",
			data:     []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50},
			expected: "image/webp",
		},
		{
			name:     "Unknown format - defaults to JPEG",
			data:     []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: "image/jpeg",
		},
		{
			name:     "Too short data - defaults to JPEG",
			data:     []byte{0x00},
			expected: "image/jpeg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectImageMimeType(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestImageToBase64_LocalFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.png")

	// 创建测试文件（PNG header，至少12字节）
	err := os.WriteFile(testFile, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x00}, 0644)
	assert.NoError(t, err)

	// 测试
	result, err := ImageToBase64(testFile)
	assert.NoError(t, err)
	assert.Contains(t, result, "data:image/png;base64,")
}

func TestImageToBase64_LocalFileNotFound(t *testing.T) {
	_, err := ImageToBase64("/non/existent/file.png")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read local image file")
}

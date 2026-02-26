package image

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageOptions(t *testing.T) {
	tests := []struct {
		name   string
		option ImageOption
		verify func(*testing.T, *ImageOptions)
	}{
		{
			name:   "WithNegativePrompt",
			option: WithNegativePrompt("no text"),
			verify: func(t *testing.T, opts *ImageOptions) {
				assert.Equal(t, "no text", opts.NegativePrompt)
			},
		},
		{
			name:   "WithSize",
			option: WithSize("1024x1024"),
			verify: func(t *testing.T, opts *ImageOptions) {
				assert.Equal(t, "1024x1024", opts.Size)
			},
		},
		{
			name:   "WithQuality",
			option: WithQuality("high"),
			verify: func(t *testing.T, opts *ImageOptions) {
				assert.Equal(t, "high", opts.Quality)
			},
		},
		{
			name:   "WithStyle",
			option: WithStyle("realistic"),
			verify: func(t *testing.T, opts *ImageOptions) {
				assert.Equal(t, "realistic", opts.Style)
			},
		},
		{
			name:   "WithSteps",
			option: WithSteps(50),
			verify: func(t *testing.T, opts *ImageOptions) {
				assert.Equal(t, 50, opts.Steps)
			},
		},
		{
			name:   "WithCfgScale",
			option: WithCfgScale(7.5),
			verify: func(t *testing.T, opts *ImageOptions) {
				assert.Equal(t, 7.5, opts.CfgScale)
			},
		},
		{
			name:   "WithSeed",
			option: WithSeed(12345),
			verify: func(t *testing.T, opts *ImageOptions) {
				assert.Equal(t, int64(12345), opts.Seed)
			},
		},
		{
			name:   "WithModel",
			option: WithModel("dall-e-3"),
			verify: func(t *testing.T, opts *ImageOptions) {
				assert.Equal(t, "dall-e-3", opts.Model)
			},
		},
		{
			name:   "WithDimensions",
			option: WithDimensions(1920, 1080),
			verify: func(t *testing.T, opts *ImageOptions) {
				assert.Equal(t, 1920, opts.Width)
				assert.Equal(t, 1080, opts.Height)
			},
		},
		{
			name:   "WithReferenceImages",
			option: WithReferenceImages([]string{"http://example.com/ref1.jpg", "http://example.com/ref2.jpg"}),
			verify: func(t *testing.T, opts *ImageOptions) {
				assert.Len(t, opts.ReferenceImages, 2)
				assert.Equal(t, "http://example.com/ref1.jpg", opts.ReferenceImages[0])
				assert.Equal(t, "http://example.com/ref2.jpg", opts.ReferenceImages[1])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ImageOptions{}
			tt.option(opts)
			tt.verify(t, opts)
		})
	}
}

func TestImageOptionsChaining(t *testing.T) {
	opts := &ImageOptions{}

	WithSize("1024x1024")(opts)
	WithQuality("high")(opts)
	WithStyle("realistic")(opts)
	WithSteps(30)(opts)

	assert.Equal(t, "1024x1024", opts.Size)
	assert.Equal(t, "high", opts.Quality)
	assert.Equal(t, "realistic", opts.Style)
	assert.Equal(t, 30, opts.Steps)
}

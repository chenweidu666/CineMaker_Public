package video

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVideoOptions(t *testing.T) {
	tests := []struct {
		name   string
		option VideoOption
		verify func(*testing.T, *VideoOptions)
	}{
		{
			name:   "WithModel",
			option: WithModel("runway-gen3"),
			verify: func(t *testing.T, opts *VideoOptions) {
				assert.Equal(t, "runway-gen3", opts.Model)
			},
		},
		{
			name:   "WithDuration",
			option: WithDuration(5),
			verify: func(t *testing.T, opts *VideoOptions) {
				assert.Equal(t, 5, opts.Duration)
			},
		},
		{
			name:   "WithFPS",
			option: WithFPS(24),
			verify: func(t *testing.T, opts *VideoOptions) {
				assert.Equal(t, 24, opts.FPS)
			},
		},
		{
			name:   "WithResolution",
			option: WithResolution("720p"),
			verify: func(t *testing.T, opts *VideoOptions) {
				assert.Equal(t, "720p", opts.Resolution)
			},
		},
		{
			name:   "WithAspectRatio",
			option: WithAspectRatio("16:9"),
			verify: func(t *testing.T, opts *VideoOptions) {
				assert.Equal(t, "16:9", opts.AspectRatio)
			},
		},
		{
			name:   "WithStyle",
			option: WithStyle("realistic"),
			verify: func(t *testing.T, opts *VideoOptions) {
				assert.Equal(t, "realistic", opts.Style)
			},
		},
		{
			name:   "WithMotionLevel",
			option: WithMotionLevel(50),
			verify: func(t *testing.T, opts *VideoOptions) {
				assert.Equal(t, 50, opts.MotionLevel)
			},
		},
		{
			name:   "WithCameraMotion",
			option: WithCameraMotion("pan"),
			verify: func(t *testing.T, opts *VideoOptions) {
				assert.Equal(t, "pan", opts.CameraMotion)
			},
		},
		{
			name:   "WithSeed",
			option: WithSeed(12345),
			verify: func(t *testing.T, opts *VideoOptions) {
				assert.Equal(t, int64(12345), opts.Seed)
			},
		},
		{
			name:   "WithFirstFrame",
			option: WithFirstFrame("http://example.com/first.jpg"),
			verify: func(t *testing.T, opts *VideoOptions) {
				assert.Equal(t, "http://example.com/first.jpg", opts.FirstFrameURL)
			},
		},
		{
			name:   "WithLastFrame",
			option: WithLastFrame("http://example.com/last.jpg"),
			verify: func(t *testing.T, opts *VideoOptions) {
				assert.Equal(t, "http://example.com/last.jpg", opts.LastFrameURL)
			},
		},
		{
			name:   "WithReferenceImages",
			option: WithReferenceImages([]string{"http://example.com/ref1.jpg", "http://example.com/ref2.jpg"}),
			verify: func(t *testing.T, opts *VideoOptions) {
				assert.Len(t, opts.ReferenceImageURLs, 2)
				assert.Equal(t, "http://example.com/ref1.jpg", opts.ReferenceImageURLs[0])
				assert.Equal(t, "http://example.com/ref2.jpg", opts.ReferenceImageURLs[1])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &VideoOptions{}
			tt.option(opts)
			tt.verify(t, opts)
		})
	}
}

func TestVideoOptionsChaining(t *testing.T) {
	opts := &VideoOptions{}

	WithModel("runway-gen3")(opts)
	WithDuration(5)(opts)
	WithResolution("720p")(opts)
	WithAspectRatio("16:9")(opts)
	WithSeed(12345)(opts)

	assert.Equal(t, "runway-gen3", opts.Model)
	assert.Equal(t, 5, opts.Duration)
	assert.Equal(t, "720p", opts.Resolution)
	assert.Equal(t, "16:9", opts.AspectRatio)
	assert.Equal(t, int64(12345), opts.Seed)
}

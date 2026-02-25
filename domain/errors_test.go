package domain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var sentinelErrors = []struct {
	name    string
	err     error
	message string
}{
	{"ErrDramaNotFound", ErrDramaNotFound, "drama not found"},
	{"ErrEpisodeNotFound", ErrEpisodeNotFound, "episode not found"},
	{"ErrSceneNotFound", ErrSceneNotFound, "scene not found"},
	{"ErrPropNotFound", ErrPropNotFound, "prop not found"},
	{"ErrCharacterNotFound", ErrCharacterNotFound, "character not found"},
	{"ErrLibraryItemNotFound", ErrLibraryItemNotFound, "library item not found"},
	{"ErrConfigNotFound", ErrConfigNotFound, "config not found"},
	{"ErrUnauthorized", ErrUnauthorized, "unauthorized"},
	{"ErrCharacterNoImage", ErrCharacterNoImage, "character has no image"},
}

func TestDomainErrors_NotNil(t *testing.T) {
	for _, tt := range sentinelErrors {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.err)
		})
	}
}

func TestDomainErrors_Message(t *testing.T) {
	for _, tt := range sentinelErrors {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.message, tt.err.Error())
		})
	}
}

func TestDomainErrors_Is(t *testing.T) {
	for _, tt := range sentinelErrors {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, errors.Is(tt.err, tt.err))
		})
	}
}

func TestDomainErrors_IsDifferentiates(t *testing.T) {
	allSentinels := []error{
		ErrDramaNotFound, ErrEpisodeNotFound, ErrSceneNotFound,
		ErrPropNotFound, ErrCharacterNotFound, ErrLibraryItemNotFound,
		ErrConfigNotFound, ErrUnauthorized, ErrCharacterNoImage,
	}
	for _, tt := range sentinelErrors {
		t.Run(tt.name, func(t *testing.T) {
			for _, other := range allSentinels {
				if other == tt.err {
					continue
				}
				assert.False(t, errors.Is(tt.err, other), "expected %q not to be %q", tt.err, other)
			}
		})
	}
}

func TestDomainErrors_WrapAndUnwrap(t *testing.T) {
	for _, tt := range sentinelErrors {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := fmt.Errorf("operation failed: %w", tt.err)
			assert.True(t, errors.Is(wrapped, tt.err))
			assert.Equal(t, tt.message, tt.err.Error())
		})
	}
}

package domain

import "errors"

var (
	ErrDramaNotFound       = errors.New("drama not found")
	ErrEpisodeNotFound     = errors.New("episode not found")
	ErrSceneNotFound       = errors.New("scene not found")
	ErrPropNotFound        = errors.New("prop not found")
	ErrCharacterNotFound   = errors.New("character not found")
	ErrLibraryItemNotFound = errors.New("library item not found")
	ErrConfigNotFound      = errors.New("config not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrCharacterNoImage    = errors.New("character has no image")
)

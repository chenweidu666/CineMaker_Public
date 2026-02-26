package storage

import "io"

// Storage defines the interface for file storage backends (local, COS, S3, etc.)
type Storage interface {
	Upload(file io.Reader, filename, category string) (string, error)
	Delete(url string) error
	GetURL(path string) string
	DownloadFromURL(url, category string) (string, error)
	DownloadFromURLWithPath(url, category string) (*DownloadResult, error)
	GetAbsolutePath(relativePath string) string
}

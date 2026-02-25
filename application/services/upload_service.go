package services

import (
	"io"
	"time"

	"github.com/cinemaker/backend/infrastructure/storage"
	"github.com/cinemaker/backend/pkg/logger"
)

type UploadService struct {
	store storage.Storage
	log   *logger.Logger
}

func NewUploadService(store storage.Storage, log *logger.Logger) *UploadService {
	return &UploadService{
		store: store,
		log:   log,
	}
}

type UploadResult struct {
	URL       string
	LocalPath string
}

func (s *UploadService) UploadFile(file io.Reader, fileName, contentType string, category string) (*UploadResult, error) {
	url, err := s.store.Upload(file, fileName, category)
	if err != nil {
		s.log.Errorw("Failed to upload file", "error", err, "filename", fileName)
		return nil, err
	}

	s.log.Infow("File uploaded successfully", "url", url)
	return &UploadResult{
		URL:       url,
		LocalPath: url,
	}, nil
}

func (s *UploadService) UploadCharacterImage(file io.Reader, fileName, contentType string) (*UploadResult, error) {
	return s.UploadFile(file, fileName, contentType, "characters")
}

func (s *UploadService) DeleteFile(fileURL string) error {
	return s.store.Delete(fileURL)
}

func (s *UploadService) GetPresignedURL(objectName string, expiry time.Duration) (string, error) {
	return s.store.GetURL(objectName), nil
}

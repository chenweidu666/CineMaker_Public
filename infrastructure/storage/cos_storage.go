package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// COSStorage implements Storage using Tencent Cloud COS.
type COSStorage struct {
	client   *cos.Client
	bucket   string
	region   string
	cdnURL   string // optional CDN domain, e.g. "https://cdn.example.com"
	cosURL   string // e.g. "https://bucket-xxx.cos.ap-shanghai.myqcloud.com"
	localTmp string // local temp dir for tools that need file paths (ffmpeg, etc.)
}

// Compile-time check
var _ Storage = (*COSStorage)(nil)

type COSConfig struct {
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	SecretID  string `mapstructure:"secret_id"`
	SecretKey string `mapstructure:"secret_key"`
	CDNURL    string `mapstructure:"cdn_url"`
}

func NewCOSStorage(cfg COSConfig, localTmpDir string) (*COSStorage, error) {
	bucketURL, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region))
	serviceURL, _ := url.Parse(fmt.Sprintf("https://cos.%s.myqcloud.com", cfg.Region))

	client := cos.NewClient(&cos.BaseURL{
		BucketURL:  bucketURL,
		ServiceURL: serviceURL,
	}, &http.Client{
		Timeout: 10 * time.Minute,
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})

	if err := os.MkdirAll(localTmpDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create COS temp directory: %w", err)
	}

	cosURL := bucketURL.String()
	cdnURL := strings.TrimRight(cfg.CDNURL, "/")
	if cdnURL == "" {
		cdnURL = cosURL
	}

	return &COSStorage{
		client:   client,
		bucket:   cfg.Bucket,
		region:   cfg.Region,
		cdnURL:   cdnURL,
		cosURL:   cosURL,
		localTmp: localTmpDir,
	}, nil
}

func (s *COSStorage) Upload(file io.Reader, filename, category string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	uniqueID := uuid.New().String()[:8]
	ext := filepath.Ext(filename)
	objectKey := fmt.Sprintf("%s/%s_%s%s", category, timestamp, uniqueID, ext)

	opt := &cos.ObjectPutOptions{
		ACLHeaderOptions: &cos.ACLHeaderOptions{
			XCosACL: "public-read",
		},
	}
	_, err := s.client.Object.Put(context.Background(), objectKey, file, opt)
	if err != nil {
		return "", fmt.Errorf("COS upload failed: %w", err)
	}

	return fmt.Sprintf("%s/%s", s.cdnURL, objectKey), nil
}

func (s *COSStorage) Delete(fileURL string) error {
	if fileURL == "" {
		return nil
	}

	objectKey := s.urlToObjectKey(fileURL)
	if objectKey == "" {
		return nil
	}

	_, err := s.client.Object.Delete(context.Background(), objectKey)
	if err != nil {
		return fmt.Errorf("COS delete failed: %w", err)
	}
	return nil
}

func (s *COSStorage) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", s.cdnURL, strings.TrimLeft(path, "/"))
}

func (s *COSStorage) DownloadFromURL(remoteURL, category string) (string, error) {
	result, err := s.DownloadFromURLWithPath(remoteURL, category)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

func (s *COSStorage) DownloadFromURLWithPath(remoteURL, category string) (*DownloadResult, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: HTTP %d", resp.StatusCode)
	}

	ext := getFileExtension(remoteURL, resp.Header.Get("Content-Type"))
	timestamp := time.Now().Format("20060102_150405")
	uniqueID := uuid.New().String()[:8]
	filename := fmt.Sprintf("%s_%s%s", timestamp, uniqueID, ext)
	objectKey := fmt.Sprintf("%s/%s", category, filename)

	// Save to local temp first (for tools that need local paths)
	tmpDir := filepath.Join(s.localTmp, category)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	tmpFile := filepath.Join(tmpDir, filename)
	dst, err := os.Create(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	if _, err := io.Copy(dst, resp.Body); err != nil {
		dst.Close()
		return nil, fmt.Errorf("failed to save temp file: %w", err)
	}
	dst.Close()

	// Upload to COS
	f, err := os.Open(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open temp file: %w", err)
	}
	defer f.Close()

	putOpt := &cos.ObjectPutOptions{
		ACLHeaderOptions: &cos.ACLHeaderOptions{
			XCosACL: "public-read",
		},
	}
	_, err = s.client.Object.Put(context.Background(), objectKey, f, putOpt)
	if err != nil {
		return nil, fmt.Errorf("COS upload failed: %w", err)
	}

	cosURL := fmt.Sprintf("%s/%s", s.cdnURL, objectKey)
	return &DownloadResult{
		URL:          cosURL,
		RelativePath: objectKey,
		AbsolutePath: tmpFile,
	}, nil
}

// GetAbsolutePath downloads the object to local temp and returns the local path.
// This is needed for tools like ffmpeg that require local file paths.
func (s *COSStorage) GetAbsolutePath(relativePath string) string {
	localPath := filepath.Join(s.localTmp, relativePath)

	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}

	dir := filepath.Dir(localPath)
	os.MkdirAll(dir, 0755)

	resp, err := s.client.Object.Get(context.Background(), relativePath, nil)
	if err != nil {
		return localPath
	}
	defer resp.Body.Close()

	dst, err := os.Create(localPath)
	if err != nil {
		return localPath
	}
	defer dst.Close()
	io.Copy(dst, resp.Body)

	return localPath
}

func (s *COSStorage) urlToObjectKey(fileURL string) string {
	for _, prefix := range []string{s.cdnURL + "/", s.cosURL + "/"} {
		if strings.HasPrefix(fileURL, prefix) {
			return strings.TrimPrefix(fileURL, prefix)
		}
	}
	if !strings.HasPrefix(fileURL, "http") {
		return strings.TrimLeft(fileURL, "/")
	}
	return ""
}

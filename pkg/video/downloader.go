package video

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DownloadResult holds info about a successfully downloaded video.
type DownloadResult struct {
	FilePath string  `json:"file_path"`
	Title    string  `json:"title"`
	Duration float64 `json:"duration"`
}

// CookieFilePaths are the locations to search for cookie files (Netscape format).
var CookieFilePaths = []string{
	"data/cookies.txt",
	"data/cookies_*.txt",
	"cookies.txt",
	"/app/data/cookies.txt",
	"/app/data/cookies_*.txt",
	"/app/cookies.txt",
}

// findCookieFile returns the first existing cookie file path, or empty string.
// Supports glob patterns like "data/cookies_*.txt".
func findCookieFile() string {
	for _, p := range CookieFilePaths {
		if strings.Contains(p, "*") {
			matches, _ := filepath.Glob(p)
			if len(matches) > 0 {
				// Return the most recently modified match
				best := matches[0]
				for _, m := range matches[1:] {
					if info, err := os.Stat(m); err == nil {
						if bestInfo, err2 := os.Stat(best); err2 == nil && info.ModTime().After(bestInfo.ModTime()) {
							best = m
						}
					}
				}
				return best
			}
		} else {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}

// buildYtDlpArgs constructs common yt-dlp arguments with cookie and UA support.
func buildYtDlpArgs(videoURL string, extraArgs ...string) []string {
	args := []string{
		"--no-warnings",
		"--user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	}

	if cookieFile := findCookieFile(); cookieFile != "" {
		args = append(args, "--cookies", cookieFile)
		fmt.Printf("yt-dlp: using cookies from %s\n", cookieFile)
	}

	args = append(args, extraArgs...)
	args = append(args, videoURL)
	return args
}

// needsCookies checks whether a URL is known to require cookies.
func needsCookies(url string) bool {
	cookieSites := []string{"xiaohongshu.com", "xhslink.com", "douyin.com"}
	lower := strings.ToLower(url)
	for _, site := range cookieSites {
		if strings.Contains(lower, site) {
			return true
		}
	}
	return false
}

// friendlyError translates yt-dlp errors into user-friendly Chinese messages.
func friendlyError(videoURL string, rawErr error, stderr string) error {
	msg := rawErr.Error() + " " + stderr

	if strings.Contains(msg, "No video formats found") || strings.Contains(msg, "Unsupported URL") {
		if needsCookies(videoURL) {
			siteName := "该平台"
			if strings.Contains(videoURL, "xiaohongshu") || strings.Contains(videoURL, "xhslink") {
				siteName = "小红书"
			} else if strings.Contains(videoURL, "douyin") {
				siteName = "抖音"
			}
			return fmt.Errorf("%s 需要登录 Cookie 才能下载视频。\n\n请按以下步骤操作：\n1. 用浏览器登录 %s\n2. 安装浏览器扩展「Get cookies.txt LOCALLY」\n3. 导出 cookies.txt 文件\n4. 将文件放到项目根目录 data/cookies.txt\n5. 重新提交分析", siteName, siteName)
		}
		return fmt.Errorf("无法从该链接获取视频，可能是图文内容或链接已失效")
	}

	if strings.Contains(msg, "HTTP Error 404") || strings.Contains(msg, "not found") {
		return fmt.Errorf("视频不存在或已被删除 (404)")
	}

	if strings.Contains(msg, "HTTP Error 403") {
		return fmt.Errorf("访问被拒绝，可能需要登录 Cookie（放到 data/cookies.txt）")
	}

	if strings.Contains(msg, "executable file not found") {
		return fmt.Errorf("yt-dlp 未安装，请重建 Docker 镜像")
	}

	return fmt.Errorf("下载失败: %s", rawErr.Error())
}

// DownloadVideo downloads a video from URL using yt-dlp.
func DownloadVideo(videoURL string, outputDir string) (*DownloadResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	// Step 1: Get video info
	infoArgs := buildYtDlpArgs(videoURL, "--no-download", "--print-json")
	infoCmd := exec.CommandContext(ctx, "yt-dlp", infoArgs...)

	var stderrBuf strings.Builder
	infoCmd.Stderr = &stderrBuf

	infoOut, err := infoCmd.Output()
	if err != nil {
		return nil, friendlyError(videoURL, err, stderrBuf.String())
	}

	var info struct {
		Title    string  `json:"title"`
		Duration float64 `json:"duration"`
	}
	if err := json.Unmarshal(infoOut, &info); err != nil {
		info.Title = "video"
	}

	safeTitle := sanitizeFilename(info.Title)
	if safeTitle == "" {
		safeTitle = "video"
	}

	outputTemplate := filepath.Join(outputDir, safeTitle+".%(ext)s")

	// Step 2: Download
	dlArgs := buildYtDlpArgs(videoURL,
		"-o", outputTemplate,
		"--merge-output-format", "mp4",
		"--no-playlist",
	)
	dlCmd := exec.CommandContext(ctx, "yt-dlp", dlArgs...)
	dlCmd.Stdout = os.Stdout
	dlCmd.Stderr = os.Stderr

	if err := dlCmd.Run(); err != nil {
		return nil, friendlyError(videoURL, err, "")
	}

	// Step 3: Find downloaded file
	filePath := filepath.Join(outputDir, safeTitle+".mp4")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		matches, _ := filepath.Glob(filepath.Join(outputDir, safeTitle+".*"))
		for _, m := range matches {
			ext := strings.ToLower(filepath.Ext(m))
			if ext == ".mp4" || ext == ".mkv" || ext == ".webm" || ext == ".mov" {
				filePath = m
				break
			}
		}
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Last resort: find any video file in output dir
		entries, _ := os.ReadDir(outputDir)
		for _, e := range entries {
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if ext == ".mp4" || ext == ".mkv" || ext == ".webm" || ext == ".mov" {
				filePath = filepath.Join(outputDir, e.Name())
				break
			}
		}
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("下载完成但找不到视频文件")
	}

	dur := info.Duration
	if dur == 0 {
		dur, _ = GetVideoDuration(filePath)
	}

	return &DownloadResult{
		FilePath: filePath,
		Title:    info.Title,
		Duration: dur,
	}, nil
}

// IsValidVideoURL performs a basic check on whether a URL looks like a supported video site.
func IsValidVideoURL(url string) bool {
	supportedDomains := []string{
		"xiaohongshu.com", "xhslink.com",
		"bilibili.com", "b23.tv",
		"douyin.com",
		"youtube.com", "youtu.be",
		"kuaishou.com",
		"weibo.com",
		"tiktok.com",
	}
	lower := strings.ToLower(url)
	for _, domain := range supportedDomains {
		if strings.Contains(lower, domain) {
			return true
		}
	}
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_",
		"|", "_", "\n", "_", "\r", "_",
	)
	name = replacer.Replace(name)
	if len(name) > 100 {
		name = name[:100]
	}
	return strings.TrimSpace(name)
}

// 火山方舟 API Key 三模型连通性测试
//
// 与 Web 端「测试连接」逻辑一致：对文本、图片、视频三个接口发轻量请求，
// 使用不存在的 model，鉴权通过则返回 404。三个均返回 404 才视为通过。
//
// 使用方式:
//
//	export ARK_API_KEY=你的API_Key
//	go run ./cmd/test-ark
//
//	或: ARK_API_KEY=xxx go run ./cmd/test-ark
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	baseURL      = "https://ark.cn-beijing.volces.com/api/v3"
	invalidModel = "invalid_model_conn_test_001"
)

func testEndpoint(apiKey, endpoint string, body any, name string) error {
	url := baseURL + endpoint
	data, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%s: 鉴权失败(status %d)，请检查 API Key", name, resp.StatusCode)
	}
	if resp.StatusCode != http.StatusNotFound {
		bodyBytes, _ := io.ReadAll(resp.Body)
		preview := string(bodyBytes)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return fmt.Errorf("%s: 未返回 404(status %d): %s", name, resp.StatusCode, preview)
	}
	return nil
}

func maskKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return key[:4] + "****"
}

func main() {
	apiKey := strings.TrimSpace(os.Getenv("ARK_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("TEXT_API_KEY"))
	}
	if apiKey == "" {
		fmt.Println("错误: 请设置环境变量 ARK_API_KEY 或 TEXT_API_KEY")
		fmt.Println("示例: export ARK_API_KEY=your-api-key")
		os.Exit(1)
	}

	fmt.Println("火山方舟三模型连通性测试")
	fmt.Println("  base_url:", baseURL)
	fmt.Println("  api_key:", maskKey(apiKey))
	fmt.Println()

	tests := []struct {
		name     string
		endpoint string
		body     any
	}{
		{"文本(doubao)", "/chat/completions", map[string]any{
			"model": invalidModel, "messages": []map[string]string{{"role": "user", "content": "hi"}}, "max_tokens": 1,
		}},
		{"图片(seedream)", "/images/generations", map[string]any{
			"model": invalidModel, "prompt": "x", "size": "1K", "response_format": "url",
		}},
		{"视频(seedance)", "/contents/generations/tasks", map[string]any{
			"model": invalidModel, "content": []map[string]string{{"type": "text", "text": "test"}}, "generate_audio": false,
		}},
	}

	for _, t := range tests {
		if err := testEndpoint(apiKey, t.endpoint, t.body, t.name); err != nil {
			fmt.Println("  ✗", err)
			os.Exit(1)
		}
		fmt.Println("  ✓", t.name+": 404 (通过)")
	}

	fmt.Println()
	fmt.Println("✓ 文本、图片、视频三个模型测试通过，可保存！")
}

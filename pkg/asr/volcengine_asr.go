package asr

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TranscriptSegment represents one segment of transcribed speech.
type TranscriptSegment struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Text      string  `json:"text"`
	Speaker   string  `json:"speaker,omitempty"`
}

// ProgressFunc is called during ASR polling with status updates.
type ProgressFunc func(status string, attempt int)

// ASRClient handles speech-to-text transcription via Volcengine BigASR.
type ASRClient struct {
	AppID      string
	AccessKey  string
	ResourceID string
	HTTPClient *http.Client
	OnProgress ProgressFunc
}

const (
	bigASRSubmitURL = "https://openspeech.bytedance.com/api/v3/auc/bigmodel/submit"
	bigASRQueryURL  = "https://openspeech.bytedance.com/api/v3/auc/bigmodel/query"
)

func NewASRClient(appID, accessKey string) *ASRClient {
	return &ASRClient{
		AppID:      appID,
		AccessKey:  accessKey,
		ResourceID: "volc.bigasr.auc",
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}
}

type bigASRSubmitRequest struct {
	User    map[string]string      `json:"user"`
	Audio   map[string]interface{} `json:"audio"`
	Request map[string]interface{} `json:"request"`
}

type bigASRQueryResponse struct {
	Result *struct {
		Text       string `json:"text"`
		Utterances []struct {
			Text      string `json:"text"`
			StartTime int    `json:"start_time"`
			EndTime   int    `json:"end_time"`
		} `json:"utterances"`
	} `json:"result"`
	AudioInfo *struct {
		Duration int `json:"duration"`
	} `json:"audio_info"`
}

// TranscribeFile reads a local audio file, base64-encodes it, and submits
// to BigASR via the submit/query async flow.
func (c *ASRClient) TranscribeFile(audioPath string) ([]TranscriptSegment, error) {
	data, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, fmt.Errorf("read audio file: %w", err)
	}
	b64 := base64.StdEncoding.EncodeToString(data)

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(audioPath)), ".")
	if ext == "" {
		ext = "wav"
	}

	requestID := uuid.New().String()
	if err := c.submitTaskBase64(requestID, b64, ext); err != nil {
		return nil, fmt.Errorf("submit ASR task: %w", err)
	}
	fmt.Printf("BigASR: submitted task %s (base64, %d bytes, format=%s) appID=%s resID=%s accessKey=%s...\n",
		requestID, len(data), ext, c.AppID, c.ResourceID, c.AccessKey[:8])

	return c.pollResult(requestID)
}

// TranscribeFileByURL submits an audio URL to BigASR and polls for the result.
func (c *ASRClient) TranscribeFileByURL(audioURL, audioFormat string) ([]TranscriptSegment, error) {
	requestID := uuid.New().String()
	if err := c.submitTaskURL(requestID, audioURL, audioFormat); err != nil {
		return nil, fmt.Errorf("submit ASR task: %w", err)
	}
	fmt.Printf("BigASR: submitted task %s, audio=%s\n", requestID, audioURL)

	return c.pollResult(requestID)
}

func (c *ASRClient) pollResult(requestID string) ([]TranscriptSegment, error) {
	for attempt := 0; attempt < 60; attempt++ {
		wait := 3 * time.Second
		if attempt > 5 {
			wait = 5 * time.Second
		}
		if attempt > 15 {
			wait = 10 * time.Second
		}
		time.Sleep(wait)

		if c.OnProgress != nil {
			c.OnProgress("polling", attempt+1)
		}

		segments, done, err := c.queryTask(requestID)
		if err != nil {
			return nil, fmt.Errorf("query ASR task: %w", err)
		}
		if done {
			fmt.Printf("BigASR: task %s completed, %d segments\n", requestID, len(segments))
			return segments, nil
		}
		fmt.Printf("BigASR: task %s still processing (attempt %d)\n", requestID, attempt+1)
	}
	return nil, fmt.Errorf("ASR task timed out after polling")
}

func (c *ASRClient) submitTaskBase64(requestID, b64Data, audioFormat string) error {
	return c.doSubmit(requestID, map[string]interface{}{
		"format": audioFormat,
		"data":   b64Data,
	})
}

func (c *ASRClient) submitTaskURL(requestID, audioURL, audioFormat string) error {
	return c.doSubmit(requestID, map[string]interface{}{
		"format": audioFormat,
		"url":    audioURL,
	})
}

func (c *ASRClient) doSubmit(requestID string, audioField map[string]interface{}) error {
	body := bigASRSubmitRequest{
		User:  map[string]string{"uid": c.AppID},
		Audio: audioField,
		Request: map[string]interface{}{
			"model_name":          "bigmodel",
			"enable_itn":         true,
			"enable_punc":        true,
			"show_utterances":    true,
			"enable_speaker_info": true,
		},
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", bigASRSubmitURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-App-Key", c.AppID)
	req.Header.Set("X-Api-Access-Key", c.AccessKey)
	req.Header.Set("X-Api-Resource-Id", c.ResourceID)
	req.Header.Set("X-Api-Request-Id", requestID)
	req.Header.Set("X-Api-Sequence", "-1")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	statusCode := resp.Header.Get("X-Api-Status-Code")
	statusMsg := resp.Header.Get("X-Api-Message")

	if statusCode != "20000000" {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("submit failed: status=%s msg=%s body=%s", statusCode, statusMsg, string(respBody))
	}

	return nil
}

func (c *ASRClient) queryTask(requestID string) ([]TranscriptSegment, bool, error) {
	jsonData := []byte("{}")

	req, err := http.NewRequest("POST", bigASRQueryURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, false, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-App-Key", c.AppID)
	req.Header.Set("X-Api-Access-Key", c.AccessKey)
	req.Header.Set("X-Api-Resource-Id", c.ResourceID)
	req.Header.Set("X-Api-Request-Id", requestID)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	statusCode := resp.Header.Get("X-Api-Status-Code")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}

	switch statusCode {
	case "20000001", "20000002":
		return nil, false, nil
	case "20000000":
		var qr bigASRQueryResponse
		if err := json.Unmarshal(body, &qr); err != nil {
			return nil, true, fmt.Errorf("parse result: %w, raw: %s", err, string(body[:min(len(body), 300)]))
		}

		if qr.Result == nil {
			return []TranscriptSegment{}, true, nil
		}

		if len(qr.Result.Utterances) > 0 {
			segments := make([]TranscriptSegment, 0, len(qr.Result.Utterances))
			for _, u := range qr.Result.Utterances {
				segments = append(segments, TranscriptSegment{
					StartTime: float64(u.StartTime) / 1000.0,
					EndTime:   float64(u.EndTime) / 1000.0,
					Text:      u.Text,
				})
			}
			return segments, true, nil
		}

		if qr.Result.Text != "" {
			return []TranscriptSegment{{Text: qr.Result.Text}}, true, nil
		}

		return []TranscriptSegment{}, true, nil
	case "20000003":
		return []TranscriptSegment{}, true, nil
	default:
		return nil, true, fmt.Errorf("ASR query error: status=%s body=%s", statusCode, string(body[:min(len(body), 300)]))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

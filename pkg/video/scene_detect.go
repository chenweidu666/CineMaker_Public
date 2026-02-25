package video

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Shot represents a detected scene/shot boundary in the video.
type Shot struct {
	Index     int     `json:"index"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	FramePath string  `json:"frame_path"`
}

// SceneDetectResult holds the full output of video processing.
type SceneDetectResult struct {
	Shots     []Shot  `json:"shots"`
	AudioPath string  `json:"audio_path"`
	Duration  float64 `json:"duration"`
}

// GetVideoDuration returns the video duration in seconds.
func GetVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe duration failed: %w", err)
	}
	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}

// DetectScenes uses ffprobe scene detection to find shot boundaries.
// threshold is the scene change sensitivity (0.0-1.0), lower = more sensitive.
func DetectScenes(videoPath string, threshold float64, outputDir string) (*SceneDetectResult, error) {
	if threshold <= 0 {
		threshold = 0.3
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	duration, err := GetVideoDuration(videoPath)
	if err != nil {
		return nil, err
	}

	// Use ffprobe to detect scene changes
	filter := fmt.Sprintf("select=gt(scene\\,%g),showinfo", threshold)
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-f", "lavfi",
		"-i", fmt.Sprintf("movie=%s,%s", escapeFFmpegPath(videoPath), filter),
		"-show_entries", "frame=pts_time",
		"-of", "json",
	)
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Scene detection failed (%v), falling back to interval sampling\n", err)
		return sampleAtIntervals(videoPath, duration, outputDir)
	}

	var probeResult struct {
		Frames []struct {
			PtsTime string `json:"pts_time"`
		} `json:"frames"`
	}
	if err := json.Unmarshal(out, &probeResult); err != nil || len(probeResult.Frames) == 0 {
		fmt.Printf("No scenes detected or parse error, falling back to interval sampling\n")
		return sampleAtIntervals(videoPath, duration, outputDir)
	}

	var timestamps []float64
	timestamps = append(timestamps, 0)
	for _, f := range probeResult.Frames {
		ts, err := strconv.ParseFloat(f.PtsTime, 64)
		if err == nil && ts > 0.5 {
			timestamps = append(timestamps, ts)
		}
	}
	sort.Float64s(timestamps)

	maxShots := 20
	if len(timestamps) > maxShots {
		step := len(timestamps) / maxShots
		var sampled []float64
		for i := 0; i < len(timestamps); i += step {
			sampled = append(sampled, timestamps[i])
		}
		timestamps = sampled
	}

	shots := make([]Shot, 0, len(timestamps))
	for i, ts := range timestamps {
		endTime := duration
		if i+1 < len(timestamps) {
			endTime = timestamps[i+1]
		}

		framePath := filepath.Join(outputDir, fmt.Sprintf("frame_%03d.jpg", i))
		if err := extractFrame(videoPath, ts, framePath); err != nil {
			fmt.Printf("Warning: failed to extract frame at %f: %v\n", ts, err)
			continue
		}

		shots = append(shots, Shot{
			Index:     i,
			StartTime: ts,
			EndTime:   endTime,
			FramePath: framePath,
		})
	}

	audioPath := filepath.Join(outputDir, "audio.wav")
	if err := ExtractAudio(videoPath, audioPath); err != nil {
		fmt.Printf("Warning: failed to extract audio: %v\n", err)
		audioPath = ""
	}

	return &SceneDetectResult{
		Shots:     shots,
		AudioPath: audioPath,
		Duration:  duration,
	}, nil
}

func sampleAtIntervals(videoPath string, duration float64, outputDir string) (*SceneDetectResult, error) {
	interval := 5.0
	if duration > 120 {
		interval = duration / 20.0
	} else if duration > 60 {
		interval = duration / 15.0
	}

	var shots []Shot
	for ts := 0.0; ts < duration; ts += interval {
		idx := len(shots)
		endTime := ts + interval
		if endTime > duration {
			endTime = duration
		}

		framePath := filepath.Join(outputDir, fmt.Sprintf("frame_%03d.jpg", idx))
		if err := extractFrame(videoPath, ts, framePath); err != nil {
			continue
		}

		shots = append(shots, Shot{
			Index:     idx,
			StartTime: ts,
			EndTime:   endTime,
			FramePath: framePath,
		})
	}

	audioPath := filepath.Join(outputDir, "audio.wav")
	if err := ExtractAudio(videoPath, audioPath); err != nil {
		audioPath = ""
	}

	return &SceneDetectResult{
		Shots:     shots,
		AudioPath: audioPath,
		Duration:  duration,
	}, nil
}

func extractFrame(videoPath string, timestamp float64, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-ss", fmt.Sprintf("%f", timestamp),
		"-i", videoPath,
		"-frames:v", "1",
		"-q:v", "2",
		"-y",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("extract frame: %w, output: %s", err, string(out))
	}
	return nil
}

// ExtractAudio extracts audio as 16kHz mono WAV for ASR processing.
func ExtractAudio(videoPath string, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vn",
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		"-y",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("extract audio: %w, output: %s", err, string(out))
	}
	return nil
}

// ExtractAudioSegment extracts a time-range of audio as 16kHz mono WAV.
func ExtractAudioSegment(videoPath string, startTime, endTime float64, outputPath string) error {
	duration := endTime - startTime
	if duration < 0.3 {
		return fmt.Errorf("segment too short: %.2fs", duration)
	}
	cmd := exec.Command("ffmpeg",
		"-ss", fmt.Sprintf("%f", startTime),
		"-t", fmt.Sprintf("%f", duration),
		"-i", videoPath,
		"-vn",
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		"-y",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("extract audio segment: %w, output: %s", err, string(out))
	}
	return nil
}

func escapeFFmpegPath(path string) string {
	path = strings.ReplaceAll(path, "'", "'\\''")
	return "'" + path + "'"
}

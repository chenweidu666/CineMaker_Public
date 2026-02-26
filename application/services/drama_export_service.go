package services

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/logger"
	"gorm.io/gorm"
)

// --- Export JSON schema ---

type ExportBundle struct {
	Version    string              `json:"version"`
	ExportedAt string              `json:"exported_at"`
	Drama      ExportDrama         `json:"drama"`
	Characters []ExportCharacter   `json:"characters"`
	Scenes     []ExportScene       `json:"scenes"`
	Props      []ExportProp        `json:"props"`
	Episodes   []ExportEpisode     `json:"episodes"`
}

type ExportDrama struct {
	Title         string   `json:"title"`
	Description   string   `json:"description,omitempty"`
	Genre         string   `json:"genre,omitempty"`
	Style         string   `json:"style"`
	Tags          []string `json:"tags,omitempty"`
	TotalEpisodes int      `json:"total_episodes"`
	GenerateAudio bool     `json:"generate_audio"`
}

type ExportCharacter struct {
	RefID           string   `json:"_ref_id"`
	Name            string   `json:"name"`
	Role            string   `json:"role,omitempty"`
	Gender          string   `json:"gender,omitempty"`
	AgeDescription  string   `json:"age_description,omitempty"`
	Description     string   `json:"description,omitempty"`
	Appearance      string   `json:"appearance,omitempty"`
	Personality     string   `json:"personality,omitempty"`
	VoiceStyle      string   `json:"voice_style,omitempty"`
	Prompt          string   `json:"prompt,omitempty"`
	Image           string   `json:"image,omitempty"`
	ReferenceImages []string `json:"reference_images,omitempty"`
}

type ExportScene struct {
	RefID           string   `json:"_ref_id"`
	Name            string   `json:"name,omitempty"`
	Location        string   `json:"location"`
	Time            string   `json:"time,omitempty"`
	Prompt          string   `json:"prompt,omitempty"`
	Image           string   `json:"image,omitempty"`
	ReferenceImages []string `json:"reference_images,omitempty"`
}

type ExportProp struct {
	RefID           string   `json:"_ref_id"`
	Name            string   `json:"name"`
	Type            string   `json:"type,omitempty"`
	Description     string   `json:"description,omitempty"`
	Prompt          string   `json:"prompt,omitempty"`
	Image           string   `json:"image,omitempty"`
	ReferenceImages []string `json:"reference_images,omitempty"`
}

type ExportEpisode struct {
	RefID         string             `json:"_ref_id"`
	EpisodeNumber int                `json:"episode_number"`
	Title         string             `json:"title"`
	ScriptContent string             `json:"script_content,omitempty"`
	Description   string             `json:"description,omitempty"`
	Duration      int                `json:"duration"`
	CharacterRefs []string           `json:"character_refs,omitempty"`
	Storyboards   []ExportStoryboard `json:"storyboards"`
}

type ExportStoryboard struct {
	StoryboardNumber int      `json:"storyboard_number"`
	Title            string   `json:"title,omitempty"`
	Location         string   `json:"location,omitempty"`
	Time             string   `json:"time,omitempty"`
	ShotType         string   `json:"shot_type,omitempty"`
	Angle            string   `json:"angle,omitempty"`
	Movement         string   `json:"movement,omitempty"`
	Action           string   `json:"action,omitempty"`
	Result           string   `json:"result,omitempty"`
	Atmosphere       string   `json:"atmosphere,omitempty"`
	ImagePrompt      string   `json:"image_prompt,omitempty"`
	VideoPrompt      string   `json:"video_prompt,omitempty"`
	BgmPrompt        string   `json:"bgm_prompt,omitempty"`
	SoundEffect      string   `json:"sound_effect,omitempty"`
	Dialogue         string   `json:"dialogue,omitempty"`
	Description      string   `json:"description,omitempty"`
	FirstFrameDesc   string   `json:"first_frame_desc,omitempty"`
	MiddleActionDesc string   `json:"middle_action_desc,omitempty"`
	LastFrameDesc    string   `json:"last_frame_desc,omitempty"`
	Duration         int      `json:"duration"`
	SceneRef         string   `json:"scene_ref,omitempty"`
	CharacterRefs    []string `json:"character_refs,omitempty"`
	PropRefs         []string `json:"prop_refs,omitempty"`
}

// --- Service ---

type DramaExportService struct {
	db          *gorm.DB
	log         *logger.Logger
	storagePath string
}

func NewDramaExportService(db *gorm.DB, log *logger.Logger, storagePath string) *DramaExportService {
	return &DramaExportService{db: db, log: log, storagePath: storagePath}
}

// ExportDramaToZip creates a ZIP file containing drama.json + images/.
func (s *DramaExportService) ExportDramaToZip(dramaID uint, outputPath string) error {
	drama, err := s.loadFullDrama(dramaID)
	if err != nil {
		return fmt.Errorf("load drama: %w", err)
	}

	zipFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create zip: %w", err)
	}
	defer zipFile.Close()

	zw := zip.NewWriter(zipFile)
	defer zw.Close()

	imgCollector := &imageCollector{
		zw:          zw,
		storagePath: s.storagePath,
		added:       make(map[string]string),
	}

	bundle := s.buildBundle(drama, imgCollector)

	jsonData, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	w, err := zw.Create("drama.json")
	if err != nil {
		return fmt.Errorf("create drama.json in zip: %w", err)
	}
	if _, err := w.Write(jsonData); err != nil {
		return fmt.Errorf("write drama.json: %w", err)
	}

	return nil
}

// ExportAnalysisToZip exports a video analysis result as a drama ZIP without first importing.
func (s *DramaExportService) ExportAnalysisToZip(taskID string, outputPath string) error {
	var task models.VideoAnalysisTask
	if err := s.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return fmt.Errorf("task not found: %w", err)
	}
	if task.Status != "done" {
		return fmt.Errorf("task not complete (status: %s)", task.Status)
	}

	var result models.AnalysisResult
	if err := json.Unmarshal(task.Result, &result); err != nil {
		return fmt.Errorf("parse result: %w", err)
	}

	zipFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create zip: %w", err)
	}
	defer zipFile.Close()

	zw := zip.NewWriter(zipFile)
	defer zw.Close()

	imgCollector := &imageCollector{
		zw:          zw,
		storagePath: s.storagePath,
		added:       make(map[string]string),
	}

	// Build shot frame URL map from stageData
	shotFrameURLs := make(map[int]string)
	if len(task.StageData) > 0 {
		var sd map[string]interface{}
		if err := json.Unmarshal(task.StageData, &sd); err == nil {
			if detect, ok := sd["detect"].(map[string]interface{}); ok {
				if shots, ok := detect["shots"].([]interface{}); ok {
					for _, sh := range shots {
						if shMap, ok := sh.(map[string]interface{}); ok {
							idxFloat, ok := shMap["index"].(float64)
							if !ok {
								continue
							}
							idx := int(idxFloat)
							if url, ok := shMap["frame_url"].(string); ok && url != "" {
								shotFrameURLs[idx] = url
							}
						}
					}
				}
			}
		}
	}

	// Build character ref images
	charRefImages := make(map[string][]string)
	for _, shot := range result.Shots {
		frameURL := shotFrameURLs[shot.Index]
		if frameURL == "" {
			continue
		}
		for _, name := range shot.Characters {
			urls := charRefImages[name]
			found := false
			for _, u := range urls {
				if u == frameURL {
					found = true
					break
				}
			}
			if !found {
				charRefImages[name] = append(urls, frameURL)
			}
		}
	}

	// Build characters
	characters := make([]ExportCharacter, len(result.Characters))
	charRefIDMap := make(map[string]string)
	for i, ch := range result.Characters {
		refID := fmt.Sprintf("char_%d", i)
		charRefIDMap[ch.Name] = refID
		ec := ExportCharacter{
			RefID:       refID,
			Name:        ch.Name,
			Role:        ch.Role,
			Description: ch.Description,
		}
		if urls, ok := charRefImages[ch.Name]; ok {
			for j, url := range urls {
				zipPath := imgCollector.addFromURL(url, fmt.Sprintf("images/characters/%s_ref_%d", refID, j))
				if zipPath != "" {
					ec.ReferenceImages = append(ec.ReferenceImages, zipPath)
				}
			}
		}
		characters[i] = ec
	}

	// Build scenes (deduplicated)
	type sceneEntry struct {
		refID    string
		location string
		prompt   string
		frameURL string
	}
	locationMap := make(map[string]*sceneEntry)
	var scenes []ExportScene
	for _, shot := range result.Shots {
		loc := shot.Location
		if loc == "" {
			loc = "未知场景"
		}
		if _, exists := locationMap[loc]; !exists {
			refID := fmt.Sprintf("scene_%d", len(locationMap))
			entry := &sceneEntry{refID: refID, location: loc, prompt: shot.Description, frameURL: shotFrameURLs[shot.Index]}
			locationMap[loc] = entry
		}
	}
	for _, entry := range locationMap {
		es := ExportScene{
			RefID:    entry.refID,
			Name:     entry.location,
			Location: entry.location,
			Prompt:   entry.prompt,
		}
		if entry.frameURL != "" {
			zipPath := imgCollector.addFromURL(entry.frameURL, fmt.Sprintf("images/scenes/%s_ref_0", entry.refID))
			if zipPath != "" {
				es.ReferenceImages = []string{zipPath}
			}
		}
		scenes = append(scenes, es)
	}

	// Build storyboards
	storyboards := make([]ExportStoryboard, len(result.Shots))
	for i, shot := range result.Shots {
		loc := shot.Location
		if loc == "" {
			loc = "未知场景"
		}
		sceneRef := ""
		if entry, ok := locationMap[loc]; ok {
			sceneRef = entry.refID
		}
		var charRefs []string
		for _, name := range shot.Characters {
			if ref, ok := charRefIDMap[name]; ok {
				charRefs = append(charRefs, ref)
			}
		}
		storyboards[i] = ExportStoryboard{
			StoryboardNumber: i + 1,
			Description:      shot.Description,
			Dialogue:         shot.Dialogue,
			Duration:         int(shot.EndTime - shot.StartTime),
			Atmosphere:       shot.Mood,
			Location:         shot.Location,
			SceneRef:         sceneRef,
			CharacterRefs:    charRefs,
		}
	}

	title := result.Title
	if title == "" {
		title = task.Title
	}

	bundle := &ExportBundle{
		Version:    "1.0",
		ExportedAt: time.Now().Format(time.RFC3339),
		Drama: ExportDrama{
			Title:         title,
			Description:   result.Summary,
			Style:         "realistic",
			Tags:          result.Tags,
			TotalEpisodes: 1,
			GenerateAudio: true,
		},
		Characters: characters,
		Scenes:     scenes,
		Props:      nil,
		Episodes: []ExportEpisode{{
			RefID:         "ep_0",
			EpisodeNumber: 1,
			Title:         "第1集",
			Storyboards:   storyboards,
		}},
	}

	jsonData, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	w, err := zw.Create("drama.json")
	if err != nil {
		return fmt.Errorf("create drama.json in zip: %w", err)
	}
	if _, err := w.Write(jsonData); err != nil {
		return fmt.Errorf("write drama.json: %w", err)
	}

	return nil
}

func (s *DramaExportService) loadFullDrama(dramaID uint) (*models.Drama, error) {
	var drama models.Drama
	err := s.db.
		Preload("Characters").
		Preload("Characters.Children").
		Preload("Scenes").
		Preload("Props").
		Preload("Episodes").
		Preload("Episodes.Storyboards").
		Preload("Episodes.Storyboards.Characters").
		Preload("Episodes.Storyboards.Props").
		Preload("Episodes.Characters").
		First(&drama, dramaID).Error
	if err != nil {
		return nil, err
	}
	return &drama, nil
}

func (s *DramaExportService) buildBundle(drama *models.Drama, img *imageCollector) *ExportBundle {
	charIDToRef := make(map[uint]string)
	sceneIDToRef := make(map[uint]string)
	propIDToRef := make(map[uint]string)

	// Characters
	characters := make([]ExportCharacter, len(drama.Characters))
	for i, ch := range drama.Characters {
		refID := fmt.Sprintf("char_%d", i)
		charIDToRef[ch.ID] = refID
		ec := ExportCharacter{
			RefID:       refID,
			Name:        ch.Name,
			Role:        ptrStr(ch.Role),
			Gender:      ptrStr(ch.Gender),
			AgeDescription: ptrStr(ch.AgeDescription),
			Description: ptrStr(ch.Description),
			Appearance:  ptrStr(ch.Appearance),
			Personality: ptrStr(ch.Personality),
			VoiceStyle:  ptrStr(ch.VoiceStyle),
			Prompt:      ptrStr(ch.Prompt),
		}
		ec.Image = img.addEntityImage(ch.ImageURL, ch.LocalPath, fmt.Sprintf("images/characters/%s_image", refID))
		ec.ReferenceImages = img.addRefImages(ch.ReferenceImages, fmt.Sprintf("images/characters/%s_ref", refID))
		characters[i] = ec
	}

	// Scenes
	scenes := make([]ExportScene, len(drama.Scenes))
	for i, sc := range drama.Scenes {
		refID := fmt.Sprintf("scene_%d", i)
		sceneIDToRef[sc.ID] = refID
		es := ExportScene{
			RefID:    refID,
			Name:     ptrStr(sc.Name),
			Location: sc.Location,
			Time:     sc.Time,
			Prompt:   sc.Prompt,
		}
		es.Image = img.addEntityImage(sc.ImageURL, sc.LocalPath, fmt.Sprintf("images/scenes/%s_image", refID))
		es.ReferenceImages = img.addRefImages(sc.ReferenceImages, fmt.Sprintf("images/scenes/%s_ref", refID))
		scenes[i] = es
	}

	// Props
	props := make([]ExportProp, len(drama.Props))
	for i, p := range drama.Props {
		refID := fmt.Sprintf("prop_%d", i)
		propIDToRef[p.ID] = refID
		ep := ExportProp{
			RefID:       refID,
			Name:        p.Name,
			Type:        ptrStr(p.Type),
			Description: ptrStr(p.Description),
			Prompt:      ptrStr(p.Prompt),
		}
		ep.Image = img.addEntityImage(p.ImageURL, p.LocalPath, fmt.Sprintf("images/props/%s_image", refID))
		ep.ReferenceImages = img.addRefImages(p.ReferenceImages, fmt.Sprintf("images/props/%s_ref", refID))
		props[i] = ep
	}

	// Episodes + Storyboards
	episodes := make([]ExportEpisode, len(drama.Episodes))
	for i, ep := range drama.Episodes {
		epRefID := fmt.Sprintf("ep_%d", i)

		var charRefs []string
		for _, ch := range ep.Characters {
			if ref, ok := charIDToRef[ch.ID]; ok {
				charRefs = append(charRefs, ref)
			}
		}

		sbs := make([]ExportStoryboard, len(ep.Storyboards))
		for j, sb := range ep.Storyboards {
			esb := ExportStoryboard{
				StoryboardNumber: sb.StoryboardNumber,
				Title:            ptrStr(sb.Title),
				Location:         ptrStr(sb.Location),
				Time:             ptrStr(sb.Time),
				ShotType:         ptrStr(sb.ShotType),
				Angle:            ptrStr(sb.Angle),
				Movement:         ptrStr(sb.Movement),
				Action:           ptrStr(sb.Action),
				Result:           ptrStr(sb.Result),
				Atmosphere:       ptrStr(sb.Atmosphere),
				ImagePrompt:      ptrStr(sb.ImagePrompt),
				VideoPrompt:      ptrStr(sb.VideoPrompt),
				BgmPrompt:        ptrStr(sb.BgmPrompt),
				SoundEffect:      ptrStr(sb.SoundEffect),
				Dialogue:         ptrStr(sb.Dialogue),
				Description:      ptrStr(sb.Description),
				FirstFrameDesc:   ptrStr(sb.FirstFrameDesc),
				MiddleActionDesc: ptrStr(sb.MiddleActionDesc),
				LastFrameDesc:    ptrStr(sb.LastFrameDesc),
				Duration:         sb.Duration,
			}
			if sb.SceneID != nil {
				if ref, ok := sceneIDToRef[*sb.SceneID]; ok {
					esb.SceneRef = ref
				}
			}
			for _, ch := range sb.Characters {
				if ref, ok := charIDToRef[ch.ID]; ok {
					esb.CharacterRefs = append(esb.CharacterRefs, ref)
				}
			}
			for _, p := range sb.Props {
				if ref, ok := propIDToRef[p.ID]; ok {
					esb.PropRefs = append(esb.PropRefs, ref)
				}
			}
			sbs[j] = esb
		}

		episodes[i] = ExportEpisode{
			RefID:         epRefID,
			EpisodeNumber: ep.EpisodeNum,
			Title:         ep.Title,
			ScriptContent: ptrStr(ep.ScriptContent),
			Description:   ptrStr(ep.Description),
			Duration:      ep.Duration,
			CharacterRefs: charRefs,
			Storyboards:   sbs,
		}
	}

	var tags []string
	if drama.Tags != nil {
		json.Unmarshal(drama.Tags, &tags)
	}

	return &ExportBundle{
		Version:    "1.0",
		ExportedAt: time.Now().Format(time.RFC3339),
		Drama: ExportDrama{
			Title:         drama.Title,
			Description:   ptrStr(drama.Description),
			Genre:         ptrStr(drama.Genre),
			Style:         drama.Style,
			Tags:          tags,
			TotalEpisodes: drama.TotalEpisodes,
			GenerateAudio: drama.GenerateAudio == nil || *drama.GenerateAudio,
		},
		Characters: characters,
		Scenes:     scenes,
		Props:      props,
		Episodes:   episodes,
	}
}

// --- Image collector ---

type imageCollector struct {
	zw          *zip.Writer
	storagePath string
	added       map[string]string // local path -> zip path
}

func (ic *imageCollector) addEntityImage(imageURL, localPath *string, zipPrefix string) string {
	path := ""
	if localPath != nil && *localPath != "" {
		path = *localPath
	} else if imageURL != nil && *imageURL != "" {
		path = ic.urlToLocalPath(*imageURL)
	}
	if path == "" {
		return ""
	}
	return ic.addFile(path, zipPrefix)
}

func (ic *imageCollector) addRefImages(refJSON []byte, zipPrefix string) []string {
	if len(refJSON) == 0 {
		return nil
	}
	var urls []string
	if err := json.Unmarshal(refJSON, &urls); err != nil {
		return nil
	}
	var result []string
	for i, url := range urls {
		zipPath := ic.addFromURL(url, fmt.Sprintf("%s_%d", zipPrefix, i))
		if zipPath != "" {
			result = append(result, zipPath)
		}
	}
	return result
}

func (ic *imageCollector) addFromURL(url, zipPrefix string) string {
	localPath := ic.urlToLocalPath(url)
	if localPath == "" {
		return ""
	}
	return ic.addFile(localPath, zipPrefix)
}

func (ic *imageCollector) addFile(localPath, zipPrefix string) string {
	if existing, ok := ic.added[localPath]; ok {
		return existing
	}

	absPath := localPath
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(ic.storagePath, localPath)
	}

	f, err := os.Open(absPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	ext := filepath.Ext(localPath)
	if ext == "" {
		ext = ".jpg"
	}
	zipPath := zipPrefix + ext

	w, err := ic.zw.Create(zipPath)
	if err != nil {
		return ""
	}
	if _, err := io.Copy(w, f); err != nil {
		return ""
	}

	ic.added[localPath] = zipPath
	return zipPath
}

func (ic *imageCollector) urlToLocalPath(url string) string {
	if url == "" {
		return ""
	}
	const prefix = "/static/"
	if strings.HasPrefix(url, prefix) {
		return filepath.Join(ic.storagePath, url[len(prefix):])
	}
	if strings.HasPrefix(url, "http") {
		return ""
	}
	return url
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

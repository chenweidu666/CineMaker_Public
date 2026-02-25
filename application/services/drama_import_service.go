package services

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/logger"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DramaImportService struct {
	db          *gorm.DB
	log         *logger.Logger
	storagePath string
}

func NewDramaImportService(db *gorm.DB, log *logger.Logger, storagePath string) *DramaImportService {
	return &DramaImportService{db: db, log: log, storagePath: storagePath}
}

// ImportDramaFromZip reads a zip file and creates a full Drama with all associations.
// Returns the created drama ID.
func (s *DramaImportService) ImportDramaFromZip(zipPath string) (uint, error) {
	tmpDir, err := os.MkdirTemp("", "drama_import_*")
	if err != nil {
		return 0, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := unzipTo(zipPath, tmpDir); err != nil {
		return 0, fmt.Errorf("unzip: %w", err)
	}

	jsonData, err := os.ReadFile(filepath.Join(tmpDir, "drama.json"))
	if err != nil {
		return 0, fmt.Errorf("read drama.json: %w", err)
	}

	var bundle ExportBundle
	if err := json.Unmarshal(jsonData, &bundle); err != nil {
		return 0, fmt.Errorf("parse drama.json: %w", err)
	}
	if bundle.Version == "" {
		return 0, fmt.Errorf("missing version in drama.json")
	}

	var dramaID uint
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var txErr error
		dramaID, txErr = s.doImport(tx, &bundle, tmpDir)
		return txErr
	})
	if err != nil {
		return 0, err
	}
	return dramaID, nil
}

func (s *DramaImportService) doImport(tx *gorm.DB, bundle *ExportBundle, tmpDir string) (uint, error) {
	genAudio := bundle.Drama.GenerateAudio
	tagsJSON, _ := json.Marshal(bundle.Drama.Tags)

	drama := models.Drama{
		Title:         bundle.Drama.Title,
		Description:   strPtr(bundle.Drama.Description),
		Genre:         strPtr(bundle.Drama.Genre),
		Style:         bundle.Drama.Style,
		Tags:          datatypes.JSON(tagsJSON),
		TotalEpisodes: bundle.Drama.TotalEpisodes,
		GenerateAudio: &genAudio,
		Status:        "draft",
	}
	if err := tx.Create(&drama).Error; err != nil {
		return 0, fmt.Errorf("create drama: %w", err)
	}

	storageDir := filepath.Join(s.storagePath, fmt.Sprintf("imported_%d", drama.ID))
	os.MkdirAll(storageDir, 0755)

	charRefToID := make(map[string]uint)
	sceneRefToID := make(map[string]uint)
	propRefToID := make(map[string]uint)

	// Characters
	for _, ec := range bundle.Characters {
		ch := models.Character{
			DramaID:     drama.ID,
			Name:        ec.Name,
			Role:        strPtr(ec.Role),
			Gender:      strPtr(ec.Gender),
			AgeDescription: strPtr(ec.AgeDescription),
			Description: strPtr(ec.Description),
			Appearance:  strPtr(ec.Appearance),
			Personality: strPtr(ec.Personality),
			VoiceStyle:  strPtr(ec.VoiceStyle),
			Prompt:      strPtr(ec.Prompt),
		}
		if ec.Image != "" {
			localPath := s.saveImage(tmpDir, ec.Image, storageDir)
			if localPath != "" {
				url := s.localPathToURL(localPath)
				ch.ImageURL = &url
				ch.LocalPath = &localPath
			}
		}
		refImgs := s.saveRefImages(tmpDir, ec.ReferenceImages, storageDir)
		if len(refImgs) > 0 {
			data, _ := json.Marshal(refImgs)
			ch.ReferenceImages = datatypes.JSON(data)
		}
		if err := tx.Create(&ch).Error; err != nil {
			return 0, fmt.Errorf("create character %s: %w", ec.Name, err)
		}
		charRefToID[ec.RefID] = ch.ID
	}

	// Scenes
	for _, es := range bundle.Scenes {
		sc := models.Scene{
			DramaID:  drama.ID,
			Name:     strPtr(es.Name),
			Location: es.Location,
			Time:     es.Time,
			Prompt:   es.Prompt,
			Status:   "pending",
		}
		if es.Image != "" {
			localPath := s.saveImage(tmpDir, es.Image, storageDir)
			if localPath != "" {
				url := s.localPathToURL(localPath)
				sc.ImageURL = &url
				sc.LocalPath = &localPath
			}
		}
		refImgs := s.saveRefImages(tmpDir, es.ReferenceImages, storageDir)
		if len(refImgs) > 0 {
			data, _ := json.Marshal(refImgs)
			sc.ReferenceImages = datatypes.JSON(data)
		}
		if err := tx.Create(&sc).Error; err != nil {
			return 0, fmt.Errorf("create scene %s: %w", es.Location, err)
		}
		sceneRefToID[es.RefID] = sc.ID
	}

	// Props
	for _, ep := range bundle.Props {
		p := models.Prop{
			DramaID:     drama.ID,
			Name:        ep.Name,
			Type:        strPtr(ep.Type),
			Description: strPtr(ep.Description),
			Prompt:      strPtr(ep.Prompt),
		}
		if ep.Image != "" {
			localPath := s.saveImage(tmpDir, ep.Image, storageDir)
			if localPath != "" {
				url := s.localPathToURL(localPath)
				p.ImageURL = &url
				p.LocalPath = &localPath
			}
		}
		refImgs := s.saveRefImages(tmpDir, ep.ReferenceImages, storageDir)
		if len(refImgs) > 0 {
			data, _ := json.Marshal(refImgs)
			p.ReferenceImages = datatypes.JSON(data)
		}
		if err := tx.Create(&p).Error; err != nil {
			return 0, fmt.Errorf("create prop %s: %w", ep.Name, err)
		}
		propRefToID[ep.RefID] = p.ID
	}

	// Episodes + Storyboards + Associations
	for _, ee := range bundle.Episodes {
		episode := models.Episode{
			DramaID:    drama.ID,
			EpisodeNum: ee.EpisodeNumber,
			Title:      ee.Title,
			ScriptContent: strPtr(ee.ScriptContent),
			Description:   strPtr(ee.Description),
			Duration:      ee.Duration,
			Status:        "draft",
		}
		if err := tx.Create(&episode).Error; err != nil {
			return 0, fmt.Errorf("create episode %s: %w", ee.Title, err)
		}

		// Episode-Character M2M
		var epChars []models.Character
		for _, ref := range ee.CharacterRefs {
			if cID, ok := charRefToID[ref]; ok {
				epChars = append(epChars, models.Character{ID: cID})
			}
		}
		if len(epChars) > 0 {
			tx.Model(&episode).Association("Characters").Append(&epChars)
		}

		for _, esb := range ee.Storyboards {
			sb := models.Storyboard{
				EpisodeID:        episode.ID,
				StoryboardNumber: esb.StoryboardNumber,
				Title:            strPtr(esb.Title),
				Location:         strPtr(esb.Location),
				Time:             strPtr(esb.Time),
				ShotType:         strPtr(esb.ShotType),
				Angle:            strPtr(esb.Angle),
				Movement:         strPtr(esb.Movement),
				Action:           strPtr(esb.Action),
				Result:           strPtr(esb.Result),
				Atmosphere:       strPtr(esb.Atmosphere),
				ImagePrompt:      strPtr(esb.ImagePrompt),
				VideoPrompt:      strPtr(esb.VideoPrompt),
				BgmPrompt:        strPtr(esb.BgmPrompt),
				SoundEffect:      strPtr(esb.SoundEffect),
				Dialogue:         strPtr(esb.Dialogue),
				Description:      strPtr(esb.Description),
				FirstFrameDesc:   strPtr(esb.FirstFrameDesc),
				MiddleActionDesc: strPtr(esb.MiddleActionDesc),
				LastFrameDesc:    strPtr(esb.LastFrameDesc),
				Duration:         esb.Duration,
				Status:           "pending",
			}
			if esb.SceneRef != "" {
				if sID, ok := sceneRefToID[esb.SceneRef]; ok {
					sb.SceneID = &sID
				}
			}
			if err := tx.Create(&sb).Error; err != nil {
				return 0, fmt.Errorf("create storyboard #%d: %w", esb.StoryboardNumber, err)
			}

			// Storyboard-Character M2M
			var sbChars []models.Character
			for _, ref := range esb.CharacterRefs {
				if cID, ok := charRefToID[ref]; ok {
					sbChars = append(sbChars, models.Character{ID: cID})
				}
			}
			if len(sbChars) > 0 {
				tx.Model(&sb).Association("Characters").Append(&sbChars)
			}

			// Storyboard-Prop M2M
			var sbProps []models.Prop
			for _, ref := range esb.PropRefs {
				if pID, ok := propRefToID[ref]; ok {
					sbProps = append(sbProps, models.Prop{ID: pID})
				}
			}
			if len(sbProps) > 0 {
				tx.Model(&sb).Association("Props").Append(&sbProps)
			}
		}
	}

	return drama.ID, nil
}

func (s *DramaImportService) saveImage(tmpDir, zipRelPath, storageDir string) string {
	srcPath := filepath.Join(tmpDir, zipRelPath)
	if _, err := os.Stat(srcPath); err != nil {
		return ""
	}
	dstPath := filepath.Join(storageDir, filepath.Base(zipRelPath))
	os.MkdirAll(filepath.Dir(dstPath), 0755)
	if err := copyFile(srcPath, dstPath); err != nil {
		return ""
	}
	return dstPath
}

func (s *DramaImportService) saveRefImages(tmpDir string, zipPaths []string, storageDir string) []string {
	var urls []string
	for _, zp := range zipPaths {
		localPath := s.saveImage(tmpDir, zp, storageDir)
		if localPath != "" {
			urls = append(urls, s.localPathToURL(localPath))
		}
	}
	return urls
}

func (s *DramaImportService) localPathToURL(localPath string) string {
	rel, err := filepath.Rel(s.storagePath, localPath)
	if err != nil {
		return "/static/" + filepath.Base(localPath)
	}
	return "/static/" + filepath.ToSlash(rel)
}

// --- Helpers ---

func unzipTo(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue // zip slip protection
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(target), 0755)

		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

package services

import (
	"fmt"

	"github.com/cinemaker/backend/domain/models"
)

// UpdateStoryboard 更新分镜的所有字段，并重新生成提示词
func (s *StoryboardService) UpdateStoryboard(storyboardID string, updates map[string]interface{}) error {
	// 查找分镜
	var storyboard models.Storyboard
	if err := s.db.First(&storyboard, storyboardID).Error; err != nil {
		return fmt.Errorf("storyboard not found: %w", err)
	}

	// 构建用于重新生成提示词的Storyboard结构
	sb := Storyboard{
		ShotNumber: storyboard.StoryboardNumber,
	}

	// 从updates中提取字段并更新
	updateData := make(map[string]interface{})

	if val, ok := updates["title"].(string); ok && val != "" {
		updateData["title"] = val
		sb.Title = val
	}
	if val, ok := updates["shot_type"].(string); ok && val != "" {
		updateData["shot_type"] = val
		sb.ShotType = val
	}
	if val, ok := updates["angle"].(string); ok && val != "" {
		updateData["angle"] = val
		sb.Angle = val
	}
	if val, ok := updates["movement"].(string); ok && val != "" {
		updateData["movement"] = val
		sb.Movement = val
	}
	if val, ok := updates["location"].(string); ok && val != "" {
		updateData["location"] = val
		sb.Location = val
	}
	if val, ok := updates["time"].(string); ok && val != "" {
		updateData["time"] = val
		sb.Time = val
	}
	if val, ok := updates["action"].(string); ok && val != "" {
		updateData["action"] = val
		sb.Action = val
	}
	if val, ok := updates["dialogue"].(string); ok && val != "" {
		updateData["dialogue"] = val
		sb.Dialogue = val
	}
	if val, ok := updates["result"].(string); ok && val != "" {
		updateData["result"] = val
		sb.Result = val
	}
	if val, ok := updates["atmosphere"].(string); ok && val != "" {
		updateData["atmosphere"] = val
		sb.Atmosphere = val
	}
	if val, ok := updates["description"].(string); ok && val != "" {
		updateData["description"] = val
	}
	frameDescChanged := false
	if val, ok := updates["first_frame_desc"].(string); ok {
		updateData["first_frame_desc"] = val
		sb.FirstFrameDesc = val
		frameDescChanged = true
	}
	if val, ok := updates["middle_action_desc"].(string); ok {
		updateData["middle_action_desc"] = val
		sb.MiddleActionDesc = val
		frameDescChanged = true
	}
	if val, ok := updates["last_frame_desc"].(string); ok {
		updateData["last_frame_desc"] = val
		sb.LastFrameDesc = val
		frameDescChanged = true
	}
	// 任一帧描述变化时，自动从三段拼合 action（供下游视频/图片提示词使用）
	if frameDescChanged {
		// 从DB补全未更新的字段
		if sb.FirstFrameDesc == "" && storyboard.FirstFrameDesc != nil {
			sb.FirstFrameDesc = *storyboard.FirstFrameDesc
		}
		if sb.MiddleActionDesc == "" && storyboard.MiddleActionDesc != nil {
			sb.MiddleActionDesc = *storyboard.MiddleActionDesc
		}
		if sb.LastFrameDesc == "" && storyboard.LastFrameDesc != nil {
			sb.LastFrameDesc = *storyboard.LastFrameDesc
		}
		composed := composeActionFromFrameDescs(sb.FirstFrameDesc, sb.MiddleActionDesc, sb.LastFrameDesc)
		if composed != "" {
			updateData["action"] = composed
			sb.Action = composed
		}
	}
	if val, ok := updates["bgm_prompt"].(string); ok && val != "" {
		updateData["bgm_prompt"] = val
		sb.BgmPrompt = val
	}
	if val, ok := updates["sound_effect"].(string); ok && val != "" {
		updateData["sound_effect"] = val
		sb.SoundEffect = val
	}
	if val, ok := updates["duration"].(float64); ok {
		updateData["duration"] = int(val)
		sb.Duration = int(val)
	}
	if val, ok := updates["scene_id"].(float64); ok {
		sceneID := uint(val)
		updateData["scene_id"] = sceneID
	}

	// 检查是否直接传入了video_prompt
	hasVideoPrompt := false
	if val, ok := updates["video_prompt"].(string); ok && val != "" {
		updateData["video_prompt"] = val
		hasVideoPrompt = true
	}

	// 如果没有直接传入video_prompt，则使用当前数据库值填充缺失字段并重新生成
	if !hasVideoPrompt {
		if sb.Title == "" && storyboard.Title != nil {
			sb.Title = *storyboard.Title
		}
		if sb.ShotType == "" && storyboard.ShotType != nil {
			sb.ShotType = *storyboard.ShotType
		}
		if sb.Angle == "" && storyboard.Angle != nil {
			sb.Angle = *storyboard.Angle
		}
		if sb.Movement == "" && storyboard.Movement != nil {
			sb.Movement = *storyboard.Movement
		}
		if sb.Location == "" && storyboard.Location != nil {
			sb.Location = *storyboard.Location
		}
		if sb.Time == "" && storyboard.Time != nil {
			sb.Time = *storyboard.Time
		}
		if sb.Action == "" && storyboard.Action != nil {
			sb.Action = *storyboard.Action
		}
		if sb.Dialogue == "" && storyboard.Dialogue != nil {
			sb.Dialogue = *storyboard.Dialogue
		}
		if sb.Result == "" && storyboard.Result != nil {
			sb.Result = *storyboard.Result
		}
		if sb.Atmosphere == "" && storyboard.Atmosphere != nil {
			sb.Atmosphere = *storyboard.Atmosphere
		}
		if sb.BgmPrompt == "" && storyboard.BgmPrompt != nil {
			sb.BgmPrompt = *storyboard.BgmPrompt
		}
		if sb.SoundEffect == "" && storyboard.SoundEffect != nil {
			sb.SoundEffect = *storyboard.SoundEffect
		}
		if sb.Duration == 0 {
			sb.Duration = storyboard.Duration
		}

		// 只重新生成video_prompt
		// image_prompt不自动更新，因为可能对应多张已生成的帧图片
		videoPrompt := s.generateVideoPrompt(sb)

		updateData["video_prompt"] = videoPrompt
	}

	// 更新数据库
	if err := s.db.Model(&storyboard).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update storyboard: %w", err)
	}

	// 处理角色关联更新（many2many 关系）
	if charIDs, ok := updates["character_ids"]; ok {
		var characterIDs []uint
		switch ids := charIDs.(type) {
		case []interface{}:
			for _, id := range ids {
				switch v := id.(type) {
				case float64:
					characterIDs = append(characterIDs, uint(v))
				case string:
					// 尝试解析字符串类型的ID
					var parsed uint64
					if _, err := fmt.Sscanf(v, "%d", &parsed); err == nil {
						characterIDs = append(characterIDs, uint(parsed))
					}
				}
			}
		}

		// 查询要关联的角色
		var characters []models.Character
		if len(characterIDs) > 0 {
			if err := s.db.Where("id IN ?", characterIDs).Find(&characters).Error; err != nil {
				s.log.Warnw("Failed to load characters for association", "error", err, "character_ids", characterIDs)
			}
		}

		// 替换 many2many 关联（先清空再添加）
		if err := s.db.Model(&storyboard).Association("Characters").Replace(characters); err != nil {
			s.log.Errorw("Failed to update character associations", "error", err, "storyboard_id", storyboardID)
			return fmt.Errorf("failed to update character associations: %w", err)
		}

		s.log.Infow("Updated storyboard character associations",
			"storyboard_id", storyboardID,
			"character_count", len(characters))
	}

	s.log.Infow("Storyboard updated successfully",
		"storyboard_id", storyboardID,
		"fields_updated", len(updateData))

	return nil
}

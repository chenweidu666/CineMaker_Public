import type {
  CreateDramaRequest,
  Drama,
  DramaListQuery,
  DramaStats,
  UpdateDramaRequest
} from '../types/drama'
import request from '../utils/request'

export const dramaAPI = {
  list(params?: DramaListQuery) {
    return request.get<{
      items: Drama[]
      pagination: {
        page: number
        page_size: number
        total: number
        total_pages: number
      }
    }>('/dramas', { params })
  },

  create(data: CreateDramaRequest) {
    return request.post<Drama>('/dramas', data)
  },

  get(id: string) {
    return request.get<Drama>(`/dramas/${id}`)
  },

  update(id: string, data: UpdateDramaRequest) {
    return request.put<Drama>(`/dramas/${id}`, data)
  },

  delete(id: string) {
    return request.delete(`/dramas/${id}`)
  },

  getStats() {
    return request.get<DramaStats>('/dramas/stats')
  },

  saveOutline(id: string, data: { title: string; summary: string; genre?: string; tags?: string[] }) {
    return request.put(`/dramas/${id}/outline`, data)
  },

  getCharacters(dramaId: string) {
    return request.get(`/dramas/${dramaId}/characters`)
  },

  saveCharacters(id: string, data: any[], episodeId?: string) {
    return request.put(`/dramas/${id}/characters`, {
      characters: data,
      episode_id: episodeId ? parseInt(episodeId) : undefined
    })
  },

  updateCharacter(id: number, data: any) {
    return request.put(`/characters/${id}`, data)
  },

  saveEpisodes(id: string, data: any[]) {
    return request.put(`/dramas/${id}/episodes`, { episodes: data })
  },

  updateEpisodeTitle(id: string, episodeNumber: number, title: string, status?: string) {
    return request.put(`/dramas/${id}/episodes/title`, { 
      episode_number: episodeNumber, 
      title: title,
      ...(status ? { status } : {})
    })
  },

  saveProgress(id: string, data: { current_step: string; step_data?: any }) {
    return request.put(`/dramas/${id}/progress`, data)
  },

  generateStoryboard(episodeId: string) {
    return request.post(`/episodes/${episodeId}/storyboards`)
  },

  getBackgrounds(episodeId: string) {
    return request.get(`/images/episode/${episodeId}/backgrounds`)
  },

  extractBackgrounds(episodeId: string, model?: string) {
    return request.post<{ task_id: string; status: string; message: string }>(`/images/episode/${episodeId}/backgrounds/extract`, { model })
  },

  batchGenerateBackgrounds(episodeId: string) {
    return request.post(`/images/episode/${episodeId}/batch`)
  },

  generateSingleBackground(backgroundId: number, dramaId: string, prompt: string) {
    return request.post('/images', {
      background_id: backgroundId,
      drama_id: dramaId,
      prompt: prompt
    })
  },

  getStoryboards(episodeId: string) {
    return request.get(`/episodes/${episodeId}/storyboards`)
  },

  updateStoryboard(storyboardId: string, data: any) {
    return request.put(`/storyboards/${storyboardId}`, data)
  },

  updateScene(sceneId: string, data: {
    background_id?: string;
    characters?: string[];
    location?: string;
    time?: string;
    prompt?: string;
    action?: string;
    dialogue?: string;
    description?: string;
    duration?: number;
    image_url?: string;
    local_path?: string;
  }) {
    return request.put(`/scenes/${sceneId}`, data)
  },

  createScene(data: {
    drama_id: number;
    episode_id?: number;
    location: string;
    time?: string;
    prompt?: string;
    description?: string;
    image_url?: string;
    local_path?: string;
  }) {
    return request.post('/scenes', data)
  },

  generateSceneImage(data: { scene_id: number; prompt?: string; model?: string; reference_images?: string[] }) {
    return request.post<{ image_generation: { id: number } }>('/scenes/generate-image', data)
  },

  updateScenePrompt(sceneId: string, name: string, prompt: string, reference_images?: any[], image_orientation?: string) {
    return request.put(`/scenes/${sceneId}/prompt`, { name, prompt, reference_images, image_orientation })
  },

  polishPrompt(data: { prompt: string; type: string; orientation: string; style: string; reference_images?: string[] }) {
    return request.post<{ polished_prompt: string }>('/scenes/polish-prompt', data)
  },

  deleteScene(sceneId: string) {
    return request.delete(`/scenes/${sceneId}`)
  },

  createCharacter(data: {
    drama_id: number;
    episode_id?: number;
    parent_id?: number;
    name: string;
    outfit_name?: string;
    role?: string;
    appearance?: string;
    personality?: string;
    description?: string;
    prompt?: string;
    image_url?: string;
    local_path?: string;
    reference_images?: any[];
    image_orientation?: string;
  }) {
    return request.post('/characters', data)
  },

  getScenes(dramaId: string) {
    return request.get(`/scenes?drama_id=${dramaId}`)
  },

  // 完成集数制作（触发视频合成）
  finalizeEpisode(episodeId: string, timelineData?: any) {
    return request.post(`/episodes/${episodeId}/finalize`, timelineData || {})
  },

  createStoryboard(data: {
    episode_id: number;
    storyboard_number: number;
    title?: string;
    description?: string;
    action?: string;
    dialogue?: string;
    scene_id?: number;
    duration: number;
  }) {
    return request.post('/storyboards', data)
  },

  deleteStoryboard(storyboardId: number) {
    return request.delete(`/storyboards/${storyboardId}`)
  },

  generateVideoPrompt(storyboardId: number, model?: string, duration?: number, enableSubtitle?: boolean, generateAudio?: boolean, aspectRatio?: string, includeDialogue?: boolean) {
    const requestData = { model, duration, enable_subtitle: enableSubtitle, generate_audio: generateAudio, aspect_ratio: aspectRatio, include_dialogue: includeDialogue };
    return request.post<{ video_prompt: string }>(`/storyboards/${storyboardId}/video-prompt`, requestData)
  },

  saveEpisodeScript(episodeId: string, scriptContent: string) {
    return request.put(`/episodes/${episodeId}/script`, { script_content: scriptContent })
  },

  saveEpisodeDescription(episodeId: string, description: string) {
    return request.put(`/episodes/${episodeId}/description`, { description })
  },

  addCharacterToEpisode(episodeId: number, characterId: number) {
    return request.post(`/episodes/${episodeId}/characters`, { character_id: characterId })
  },

  addSceneToEpisode(episodeId: number, sceneId: number) {
    return request.post(`/episodes/${episodeId}/scenes`, { scene_id: sceneId })
  },

  addPropToEpisode(episodeId: number, propId: number) {
    return request.post(`/episodes/${episodeId}/props`, { prop_id: propId })
  },

  removeCharacterFromEpisode(episodeId: number, characterId: number) {
    return request.delete(`/episodes/${episodeId}/characters/${characterId}`)
  },

  removeSceneFromEpisode(episodeId: number, sceneId: number) {
    return request.delete(`/episodes/${episodeId}/scenes/${sceneId}`)
  },

  removePropFromEpisode(episodeId: number, propId: number) {
    return request.delete(`/episodes/${episodeId}/props/${propId}`)
  }
}

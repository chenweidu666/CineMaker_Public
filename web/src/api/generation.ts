import type {
    GenerateCharactersRequest
} from '../types/generation'
import request from '../utils/request'

export const generationAPI = {
  generateCharacters(data: GenerateCharactersRequest) {
    return request.post<{ task_id: string; status: string; message: string }>('/generation/characters', data)
  },

  generateStoryboard(episodeId: string, model?: string, shotCount?: number) {
    return request.post<{ task_id: string; status: string; message: string }>(`/episodes/${episodeId}/storyboards`, { model, shot_count: shotCount || 10 })
  },

  rewriteScript(episodeId: number, model?: string) {
    return request.post<{ task_id: string; status: string; message: string }>('/generation/rewrite-script', { episode_id: episodeId, model })
  },

  getTaskStatus(taskId: string) {
    return request.get<{
      id: string
      type: string
      status: string
      progress: number
      message?: string
      error?: string
      result?: string
      created_at: string
      updated_at: string
      completed_at?: string
    }>(`/tasks/${taskId}`)
  },

  // V3: 程序解析提取角色和场景（同步，不走AI）
  parseExtract(episodeId: number) {
    return request.post<{
      characters: Array<{ name: string; identity: string; personality: string; appearance: string }>
      scenes: Array<{ location: string; time: string; description: string }>
      message: string
    }>('/generation/parse-extract', { episode_id: episodeId })
  }
}

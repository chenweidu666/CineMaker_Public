/**
 * 统一资源 API 适配层
 * 角色 / 场景 / 道具三个模块统一为相同的接口签名
 */
import request from '../utils/request'

export interface ResourceAPI {
  list(dramaId: string | number): Promise<any[]>
  create(dramaId: string | number, data: Record<string, any>): Promise<any>
  update(id: number | string, data: Record<string, any>): Promise<void>
  delete(id: number | string): Promise<void>
  generateImage(id: number | string, prompt?: string, referenceImages?: string[]): Promise<any>
  saveImageConfig(id: number | string, config: { prompt?: string; reference_images?: string[]; image_orientation?: string }): Promise<void>
  extractFromScript(episodeId: number): Promise<{ task_id: string }>
  polishPrompt(data: { prompt: string; type: string; orientation: string; style: string; reference_images?: string[] }): Promise<{ polished_prompt: string }>
}

/** 角色 API */
export const characterAPI: ResourceAPI = {
  async list(dramaId) {
    const drama = await request.get<any>(`/dramas/${dramaId}`)
    return drama.characters || []
  },
  async create(dramaId, data) {
    return request.post('/characters', { ...data, drama_id: Number(dramaId) })
  },
  async update(id, data) {
    return request.put(`/characters/${id}`, data)
  },
  async delete(id) {
    return request.delete(`/characters/${id}`)
  },
  async generateImage(id, prompt, referenceImages) {
    return request.post(`/characters/${id}/generate-image`, {
      prompt,
      reference_images: referenceImages?.length ? referenceImages : undefined
    })
  },
  async saveImageConfig(id, config) {
    return request.put(`/characters/${id}`, config)
  },
  async extractFromScript(episodeId) {
    return request.post(`/episodes/${episodeId}/characters/extract`)
  },
  async polishPrompt(data) {
    return request.post('/scenes/polish-prompt', data)
  }
}

/** 场景 API */
export const sceneAPI: ResourceAPI = {
  async list(dramaId) {
    return request.get(`/scenes?drama_id=${dramaId}`)
  },
  async create(dramaId, data) {
    return request.post('/scenes', { ...data, drama_id: Number(dramaId) })
  },
  async update(id, data) {
    return request.put(`/scenes/${id}`, data)
  },
  async delete(id) {
    return request.delete(`/scenes/${id}`)
  },
  async generateImage(id, prompt, referenceImages) {
    return request.post('/scenes/generate-image', {
      scene_id: Number(id),
      prompt,
      reference_images: referenceImages?.length ? referenceImages : undefined
    })
  },
  async saveImageConfig(id, config) {
    return request.put(`/scenes/${id}/prompt`, {
      name: '',
      prompt: config.prompt,
      reference_images: config.reference_images,
      image_orientation: config.image_orientation
    })
  },
  async extractFromScript(episodeId) {
    return request.post(`/images/episode/${episodeId}/backgrounds/extract`)
  },
  async polishPrompt(data) {
    return request.post('/scenes/polish-prompt', data)
  }
}

/** 道具 API */
export const propAPI: ResourceAPI = {
  async list(dramaId) {
    return request.get(`/dramas/${dramaId}/props`)
  },
  async create(dramaId, data) {
    return request.post('/props', { ...data, drama_id: Number(dramaId) })
  },
  async update(id, data) {
    return request.put(`/props/${id}`, data)
  },
  async delete(id) {
    return request.delete(`/props/${id}`)
  },
  async generateImage(id, prompt, referenceImages) {
    return request.post(`/props/${id}/generate`, {
      prompt,
      reference_images: referenceImages?.length ? referenceImages : undefined
    })
  },
  async saveImageConfig(id, config) {
    return request.put(`/props/${id}`, config)
  },
  async extractFromScript(episodeId) {
    return request.post(`/episodes/${episodeId}/props/extract`)
  },
  async polishPrompt(data) {
    return request.post('/scenes/polish-prompt', data)
  }
}

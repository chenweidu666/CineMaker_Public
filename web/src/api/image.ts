import type {
  GenerateImageRequest,
  EditImageRequest,
  ImageGeneration,
  ImageGenerationListParams
} from '../types/image'
import request from '../utils/request'

export const imageAPI = {
  generateImage(data: GenerateImageRequest) {
    return request.post<ImageGeneration>('/images', data)
  },

  previewImagePrompt(data: GenerateImageRequest) {
    return request.post<any>('/images/preview', data)
  },

  batchGenerateForEpisode(episodeId: number, options?: { skipExisting?: boolean; frameType?: string }) {
    const params = new URLSearchParams()
    if (options?.skipExisting) params.set('skip_existing', 'true')
    if (options?.frameType) params.set('frame_type', options.frameType)
    const query = params.toString() ? `?${params.toString()}` : ''
    return request.post<ImageGeneration[]>(`/images/episode/${episodeId}/batch${query}`)
  },

  getImage(id: number) {
    return request.get<ImageGeneration>(`/images/${id}`)
  },

  listImages(params: ImageGenerationListParams) {
    return request.get<{
      items: ImageGeneration[]
      pagination: {
        page: number
        page_size: number
        total: number
        total_pages: number
      }
    }>('/images', { params })
  },

  deleteImage(id: number) {
    return request.delete(`/images/${id}`)
  },

  editImage(data: EditImageRequest) {
    return request.post<ImageGeneration>('/images/edit', data)
  },

  listEditsBySource(sourceImageId: number) {
    return request.get<ImageGeneration[]>(`/images/${sourceImageId}/edits`)
  },

  replaceWithEdit(sourceImageId: number, editImageId: number) {
    return request.post(`/images/${sourceImageId}/replace`, { edit_image_id: editImageId })
  },

  // 上传图片并创建图片生成记录
  uploadImage(data: {
    storyboard_id: number
    drama_id: number
    frame_type: string
    image_url: string
    prompt?: string
  }) {
    return request.post<ImageGeneration>('/images/upload', data)
  }
}

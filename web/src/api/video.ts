import type {
  GenerateVideoRequest,
  VideoGeneration,
  VideoGenerationListParams
} from '../types/video'
import request from '../utils/request'

export const videoAPI = {
  generateVideo(data: GenerateVideoRequest) {
    return request.post<VideoGeneration>('/videos', data)
  },

  generateFromImage(imageGenId: number) {
    return request.post<VideoGeneration>(`/videos/image/${imageGenId}`)
  },

  batchGenerateForEpisode(episodeId: number, skipExisting = false) {
    const query = skipExisting ? '?skip_existing=true' : ''
    return request.post<VideoGeneration[]>(`/videos/episode/${episodeId}/batch${query}`)
  },

  // V3 链式视频生成已禁用
  // chainGenerateForEpisode(episodeId: number, options?: { ... }) { ... },

  getVideo(id: number) {
    return request.get<VideoGeneration>(`/videos/${id}`)
  },

  listVideos(params: VideoGenerationListParams) {
    return request.get<{
      items: VideoGeneration[]
      pagination: {
        page: number
        page_size: number
        total: number
        total_pages: number
      }
    }>('/videos', { params })
  },

  deleteVideo(id: number) {
    return request.delete(`/videos/${id}`)
  },

  /**
   * V3: 从指定分镜的已完成视频截取尾帧
   * videoId 可选，指定具体视频；不传则取最新的
   */
  extractLastFrame(storyboardId: number, videoId?: number) {
    const query = videoId ? `?video_id=${videoId}` : ''
    return request.post<{
      success: boolean
      has_video: boolean
      frame_path?: string
      video_id?: number
      message: string
      videos?: { id: number; created_at: string; duration?: number; model: string }[]
    }>(`/videos/storyboard/${storyboardId}/extract-last-frame${query}`)
  }
}

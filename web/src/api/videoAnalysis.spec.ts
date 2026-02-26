import { describe, it, expect, vi, beforeEach } from 'vitest'
import { videoAnalysisAPI, dramaExportImportAPI } from './videoAnalysis'
import request from '../utils/request'

vi.mock('../utils/request', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    defaults: { baseURL: '/api/v1' }
  }
}))

describe('videoAnalysisAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('应该调用 GET /video-analysis', () => {
      videoAnalysisAPI.list()
      expect(request.get).toHaveBeenCalledWith('/video-analysis')
    })
  })

  describe('upload', () => {
    it('应该调用 POST /video-analysis/upload 并传入 FormData', () => {
      const file = new File(['content'], 'test.mp4', { type: 'video/mp4' })
      videoAnalysisAPI.upload(file)
      expect(request.post).toHaveBeenCalledTimes(1)
      const [url, body, options] = (request.post as ReturnType<typeof vi.fn>).mock.calls[0]
      expect(url).toBe('/video-analysis/upload')
      expect(body).toBeInstanceOf(FormData)
      expect((body as FormData).get('video')).toBe(file)
      expect(options).toEqual({
        headers: { 'Content-Type': 'multipart/form-data' },
        timeout: 600000
      })
    })
  })

  describe('analyzeFromURL', () => {
    it('应该调用 POST /video-analysis/from-url 并传入 url', () => {
      videoAnalysisAPI.analyzeFromURL('https://example.com/video.mp4')
      expect(request.post).toHaveBeenCalledWith('/video-analysis/from-url', {
        url: 'https://example.com/video.mp4'
      })
    })
  })

  describe('getStatus', () => {
    it('应该调用 GET /video-analysis/:taskId/status', () => {
      videoAnalysisAPI.getStatus('task-123')
      expect(request.get).toHaveBeenCalledWith('/video-analysis/task-123/status')
    })
  })

  describe('retry', () => {
    it('应该调用 POST /video-analysis/:taskId/retry', () => {
      videoAnalysisAPI.retry('task-123')
      expect(request.post).toHaveBeenCalledWith('/video-analysis/task-123/retry')
    })
  })

  describe('resynthesize', () => {
    it('应该调用 POST /video-analysis/:taskId/resynthesize 默认 include_audio 为 true', () => {
      videoAnalysisAPI.resynthesize('task-123')
      expect(request.post).toHaveBeenCalledWith('/video-analysis/task-123/resynthesize', {
        include_audio: true
      })
    })

    it('应该传入 include_audio: false 当指定时', () => {
      videoAnalysisAPI.resynthesize('task-123', false)
      expect(request.post).toHaveBeenCalledWith('/video-analysis/task-123/resynthesize', {
        include_audio: false
      })
    })
  })

  describe('deleteTask', () => {
    it('应该调用 DELETE /video-analysis/:taskId', () => {
      videoAnalysisAPI.deleteTask('task-123')
      expect(request.delete).toHaveBeenCalledWith('/video-analysis/task-123')
    })
  })

  describe('importToDrama', () => {
    it('应该调用 POST /video-analysis/:taskId/import 不带 title', () => {
      videoAnalysisAPI.importToDrama('task-123')
      expect(request.post).toHaveBeenCalledWith('/video-analysis/task-123/import', {
        title: undefined
      })
    })

    it('应该调用 POST /video-analysis/:taskId/import 带 title', () => {
      videoAnalysisAPI.importToDrama('task-123', 'My Drama')
      expect(request.post).toHaveBeenCalledWith('/video-analysis/task-123/import', {
        title: 'My Drama'
      })
    })
  })
})

describe('dramaExportImportAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('importDrama', () => {
    it('应该调用 POST /dramas/import 并传入 FormData', () => {
      const file = new File(['content'], 'drama.json', { type: 'application/json' })
      dramaExportImportAPI.importDrama(file)
      expect(request.post).toHaveBeenCalledTimes(1)
      const [url, body, options] = (request.post as ReturnType<typeof vi.fn>).mock.calls[0]
      expect(url).toBe('/dramas/import')
      expect(body).toBeInstanceOf(FormData)
      expect((body as FormData).get('file')).toBe(file)
      expect(options).toEqual({
        headers: { 'Content-Type': 'multipart/form-data' },
        timeout: 600000
      })
    })
  })
})

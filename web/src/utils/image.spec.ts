import { describe, it, expect, vi } from 'vitest'
import { fixImageUrl, getImageUrl, hasImage, getVideoUrl, hasVideo } from './image'

describe('image.ts', () => {
  describe('fixImageUrl', () => {
    it('返回空字符串当输入为空', () => {
      expect(fixImageUrl('')).toBe('')
      expect(fixImageUrl('' as any)).toBe('')
    })

    it('返回原URL当已经是完整URL', () => {
      expect(fixImageUrl('http://example.com/image.jpg')).toBe('http://example.com/image.jpg')
      expect(fixImageUrl('https://example.com/image.jpg')).toBe('https://example.com/image.jpg')
    })

    it('返回原URL当是data URL', () => {
      expect(fixImageUrl('data:image/png;base64,abc123')).toBe('data:image/png;base64,abc123')
    })

    it('添加API基础URL当是相对路径', () => {
      const originalEnv = import.meta.env.VITE_API_BASE_URL
      import.meta.env.VITE_API_BASE_URL = 'http://api.example.com'
      
      expect(fixImageUrl('/images/test.jpg')).toBe('http://api.example.com/images/test.jpg')
      expect(fixImageUrl('images/test.jpg')).toBe('http://api.example.comimages/test.jpg')
      
      import.meta.env.VITE_API_BASE_URL = originalEnv
    })
  })

  describe('getImageUrl', () => {
    it('返回空字符串当item为空', () => {
      expect(getImageUrl(null)).toBe('')
      expect(getImageUrl(undefined)).toBe('')
      expect(getImageUrl('' as any)).toBe('')
    })

    it('使用local_path优先', () => {
      const item = {
        local_path: 'images/test.jpg',
        image_url: 'http://example.com/image.jpg'
      }
      expect(getImageUrl(item)).toBe('/static/images/test.jpg')
    })

    it('回退到image_url当没有local_path', () => {
      const item = {
        image_url: 'http://example.com/image.jpg'
      }
      expect(getImageUrl(item)).toBe('http://example.com/image.jpg')
    })

    it('返回空字符串当没有图片路径', () => {
      const item = {
        name: 'test'
      }
      expect(getImageUrl(item)).toBe('')
    })
  })

  describe('hasImage', () => {
    it('返回true当有local_path', () => {
      expect(hasImage({ local_path: 'images/test.jpg' })).toBe(true)
    })

    it('返回true当有image_url', () => {
      expect(hasImage({ image_url: 'http://example.com/image.jpg' })).toBe(true)
    })

    it('返回false当没有图片路径', () => {
      expect(hasImage({ name: 'test' })).toBe(false)
      expect(hasImage(null)).toBe(false)
      expect(hasImage(undefined)).toBe(false)
    })
  })

  describe('getVideoUrl', () => {
    it('返回空字符串当item为空', () => {
      expect(getVideoUrl(null)).toBe('')
      expect(getVideoUrl(undefined)).toBe('')
      expect(getVideoUrl('' as any)).toBe('')
    })

    it('使用local_path优先', () => {
      const item = {
        local_path: 'videos/test.mp4',
        video_url: 'http://example.com/video.mp4'
      }
      expect(getVideoUrl(item)).toBe('/static/videos/test.mp4')
    })

    it('返回原URL当local_path已经是完整URL', () => {
      const item = {
        local_path: 'http://example.com/video.mp4',
        video_url: 'http://example.com/other.mp4'
      }
      expect(getVideoUrl(item)).toBe('http://example.com/video.mp4')
    })

    it('回退到video_url当没有local_path', () => {
      const item = {
        video_url: 'http://example.com/video.mp4'
      }
      expect(getVideoUrl(item)).toBe('http://example.com/video.mp4')
    })

    it('回退到url当用于assets', () => {
      const item = {
        url: 'http://example.com/video.mp4'
      }
      expect(getVideoUrl(item)).toBe('http://example.com/video.mp4')
    })

    it('返回空字符串当没有视频路径', () => {
      const item = {
        name: 'test'
      }
      expect(getVideoUrl(item)).toBe('')
    })
  })

  describe('hasVideo', () => {
    it('返回true当有local_path', () => {
      expect(hasVideo({ local_path: 'videos/test.mp4' })).toBe(true)
    })

    it('返回true当有video_url', () => {
      expect(hasVideo({ video_url: 'http://example.com/video.mp4' })).toBe(true)
    })

    it('返回true当有url', () => {
      expect(hasVideo({ url: 'http://example.com/video.mp4' })).toBe(true)
    })

    it('返回false当没有视频路径', () => {
      expect(hasVideo({ name: 'test' })).toBe(false)
      expect(hasVideo(null)).toBe(false)
      expect(hasVideo(undefined)).toBe(false)
    })
  })
})

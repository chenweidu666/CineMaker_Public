import { describe, it, expect, vi, beforeEach } from 'vitest'
import { propAPI } from './prop'
import request from '../utils/request'

vi.mock('../utils/request', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }
}))

describe('propAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('应该调用 GET /dramas/:dramaId/props', () => {
      propAPI.list('123')
      expect(request.get).toHaveBeenCalledWith('/dramas/123/props')
    })

    it('应该支持数字 dramaId', () => {
      propAPI.list(456)
      expect(request.get).toHaveBeenCalledWith('/dramas/456/props')
    })
  })

  describe('create', () => {
    it('应该调用 POST /props', () => {
      const data = { drama_id: 1, name: 'Sword', description: 'A sword' }
      propAPI.create(data)
      expect(request.post).toHaveBeenCalledWith('/props', data)
    })
  })

  describe('update', () => {
    it('应该调用 PUT /props/:id', () => {
      const data = { name: 'Updated Prop', description: 'Updated' }
      propAPI.update(1, data)
      expect(request.put).toHaveBeenCalledWith('/props/1', data)
    })
  })

  describe('delete', () => {
    it('应该调用 DELETE /props/:id', () => {
      propAPI.delete(1)
      expect(request.delete).toHaveBeenCalledWith('/props/1')
    })
  })

  describe('extractFromScript', () => {
    it('应该调用 POST /episodes/:episodeId/props/extract', () => {
      propAPI.extractFromScript(10)
      expect(request.post).toHaveBeenCalledWith('/episodes/10/props/extract')
    })
  })

  describe('generateImage', () => {
    it('应该调用 POST /props/:id/generate 带 prompt 和 reference_images', () => {
      propAPI.generateImage(1, 'a red sword', ['url1', 'url2'])
      expect(request.post).toHaveBeenCalledWith('/props/1/generate', {
        prompt: 'a red sword',
        reference_images: ['url1', 'url2']
      })
    })

    it('应该不带 reference_images 当为空数组', () => {
      propAPI.generateImage(1, 'prompt', [])
      expect(request.post).toHaveBeenCalledWith('/props/1/generate', {
        prompt: 'prompt',
        reference_images: undefined
      })
    })

    it('应该只传 prompt 当不传 referenceImages', () => {
      propAPI.generateImage(1, 'prompt')
      expect(request.post).toHaveBeenCalledWith('/props/1/generate', {
        prompt: 'prompt',
        reference_images: undefined
      })
    })

    it('应该不带可选参数', () => {
      propAPI.generateImage(1)
      expect(request.post).toHaveBeenCalledWith('/props/1/generate', {
        prompt: undefined,
        reference_images: undefined
      })
    })
  })

  describe('associateWithStoryboard', () => {
    it('应该调用 POST /storyboards/:storyboardId/props 并传入 prop_ids', () => {
      propAPI.associateWithStoryboard(5, [1, 2, 3])
      expect(request.post).toHaveBeenCalledWith('/storyboards/5/props', {
        prop_ids: [1, 2, 3]
      })
    })
  })
})

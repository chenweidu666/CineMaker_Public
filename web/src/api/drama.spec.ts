import { describe, it, expect, vi, beforeEach } from 'vitest'
import { dramaAPI } from './drama'
import request from '../utils/request'

vi.mock('../utils/request', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }
}))

describe('dramaAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('应该调用 GET /dramas', () => {
      dramaAPI.list({ page: 1, page_size: 10 })
      expect(request.get).toHaveBeenCalledWith('/dramas', { 
        params: { page: 1, page_size: 10 } 
      })
    })

    it('应该不带参数调用', () => {
      dramaAPI.list()
      expect(request.get).toHaveBeenCalledWith('/dramas', { params: undefined })
    })
  })

  describe('create', () => {
    it('应该调用 POST /dramas', () => {
      const data = { title: 'Test Drama', script: 'Test script' }
      dramaAPI.create(data)
      expect(request.post).toHaveBeenCalledWith('/dramas', data)
    })
  })

  describe('get', () => {
    it('应该调用 GET /dramas/:id', () => {
      dramaAPI.get('123')
      expect(request.get).toHaveBeenCalledWith('/dramas/123')
    })
  })

  describe('update', () => {
    it('应该调用 PUT /dramas/:id', () => {
      const data = { title: 'Updated Title' }
      dramaAPI.update('123', data)
      expect(request.put).toHaveBeenCalledWith('/dramas/123', data)
    })
  })

  describe('delete', () => {
    it('应该调用 DELETE /dramas/:id', () => {
      dramaAPI.delete('123')
      expect(request.delete).toHaveBeenCalledWith('/dramas/123')
    })
  })

  describe('getStats', () => {
    it('应该调用 GET /dramas/stats', () => {
      dramaAPI.getStats()
      expect(request.get).toHaveBeenCalledWith('/dramas/stats')
    })
  })

  describe('saveOutline', () => {
    it('应该调用 PUT /dramas/:id/outline', () => {
      const data = { title: 'Title', summary: 'Summary' }
      dramaAPI.saveOutline('123', data)
      expect(request.put).toHaveBeenCalledWith('/dramas/123/outline', data)
    })
  })

  describe('getCharacters', () => {
    it('应该调用 GET /dramas/:id/characters', () => {
      dramaAPI.getCharacters('123')
      expect(request.get).toHaveBeenCalledWith('/dramas/123/characters')
    })
  })

  describe('saveCharacters', () => {
    it('应该调用 PUT /dramas/:id/characters', () => {
      const characters = [{ name: 'Character 1' }]
      dramaAPI.saveCharacters('123', characters)
      expect(request.put).toHaveBeenCalledWith('/dramas/123/characters', {
        characters,
        episode_id: undefined
      })
    })

    it('应该包含 episode_id 当提供时', () => {
      const characters = [{ name: 'Character 1' }]
      dramaAPI.saveCharacters('123', characters, '1')
      expect(request.put).toHaveBeenCalledWith('/dramas/123/characters', {
        characters,
        episode_id: 1
      })
    })
  })

  describe('updateCharacter', () => {
    it('应该调用 PUT /characters/:id', () => {
      const data = { name: 'Updated Character' }
      dramaAPI.updateCharacter(1, data)
      expect(request.put).toHaveBeenCalledWith('/characters/1', data)
    })
  })

  describe('saveEpisodes', () => {
    it('应该调用 PUT /dramas/:id/episodes', () => {
      const episodes = [{ title: 'Episode 1' }]
      dramaAPI.saveEpisodes('123', episodes)
      expect(request.put).toHaveBeenCalledWith('/dramas/123/episodes', { episodes })
    })
  })

  describe('updateEpisodeTitle', () => {
    it('应该调用 PUT /dramas/:id/episodes/title', () => {
      dramaAPI.updateEpisodeTitle('123', 1, 'New Title')
      expect(request.put).toHaveBeenCalledWith('/dramas/123/episodes/title', {
        episode_number: 1,
        title: 'New Title'
      })
    })
  })

  describe('saveProgress', () => {
    it('应该调用 PUT /dramas/:id/progress', () => {
      const data = { current_step: 'script_edit' }
      dramaAPI.saveProgress('123', data)
      expect(request.put).toHaveBeenCalledWith('/dramas/123/progress', data)
    })

    it('应该包含 step_data 当提供时', () => {
      const data = { current_step: 'script_edit', step_data: { key: 'value' } }
      dramaAPI.saveProgress('123', data)
      expect(request.put).toHaveBeenCalledWith('/dramas/123/progress', data)
    })
  })

  describe('generateStoryboard', () => {
    it('应该调用 POST /episodes/:id/storyboards', () => {
      dramaAPI.generateStoryboard('123')
      expect(request.post).toHaveBeenCalledWith('/episodes/123/storyboards')
    })
  })

  describe('getBackgrounds', () => {
    it('应该调用 GET /images/episode/:id/backgrounds', () => {
      dramaAPI.getBackgrounds('123')
      expect(request.get).toHaveBeenCalledWith('/images/episode/123/backgrounds')
    })
  })

  describe('extractBackgrounds', () => {
    it('应该调用 POST /images/episode/:id/backgrounds/extract', () => {
      dramaAPI.extractBackgrounds('123', 'model-name')
      expect(request.post).toHaveBeenCalledWith('/images/episode/123/backgrounds/extract', {
        model: 'model-name'
      })
    })

    it('应该不带 model 参数', () => {
      dramaAPI.extractBackgrounds('123')
      expect(request.post).toHaveBeenCalledWith('/images/episode/123/backgrounds/extract', {
        model: undefined
      })
    })
  })

  describe('batchGenerateBackgrounds', () => {
    it('应该调用 POST /images/episode/:id/batch', () => {
      dramaAPI.batchGenerateBackgrounds('123')
      expect(request.post).toHaveBeenCalledWith('/images/episode/123/batch')
    })
  })

  describe('generateSingleBackground', () => {
    it('应该调用 POST /images', () => {
      const data = {
        background_id: 1,
        drama_id: '123',
        prompt: 'Test prompt'
      }
      dramaAPI.generateSingleBackground(1, '123', 'Test prompt')
      expect(request.post).toHaveBeenCalledWith('/images', data)
    })
  })

  describe('getStoryboards', () => {
    it('应该调用 GET /episodes/:id/storyboards', () => {
      dramaAPI.getStoryboards('123')
      expect(request.get).toHaveBeenCalledWith('/episodes/123/storyboards')
    })
  })

  describe('updateStoryboard', () => {
    it('应该调用 PUT /storyboards/:id', () => {
      const data = { title: 'Updated' }
      dramaAPI.updateStoryboard('123', data)
      expect(request.put).toHaveBeenCalledWith('/storyboards/123', data)
    })
  })

  describe('updateScene', () => {
    it('应该调用 PUT /scenes/:id', () => {
      const data = { location: 'New Location' }
      dramaAPI.updateScene('123', data)
      expect(request.put).toHaveBeenCalledWith('/scenes/123', data)
    })
  })

  describe('createScene', () => {
    it('应该调用 POST /scenes', () => {
      const data = {
        drama_id: 1,
        location: 'Test Location'
      }
      dramaAPI.createScene(data)
      expect(request.post).toHaveBeenCalledWith('/scenes', data)
    })
  })

  describe('generateSceneImage', () => {
    it('应该调用 POST /scenes/generate-image', () => {
      const data = { scene_id: 1, prompt: 'Test' }
      dramaAPI.generateSceneImage(data)
      expect(request.post).toHaveBeenCalledWith('/scenes/generate-image', data)
    })
  })

  describe('updateScenePrompt', () => {
    it('应该调用 PUT /scenes/:id/prompt', () => {
      dramaAPI.updateScenePrompt('123', 'Scene Name', 'New prompt')
      expect(request.put).toHaveBeenCalledWith('/scenes/123/prompt', {
        name: 'Scene Name',
        prompt: 'New prompt',
        reference_images: undefined,
        image_orientation: undefined
      })
    })

    it('应该包含可选参数', () => {
      dramaAPI.updateScenePrompt('123', 'Scene Name', 'prompt', ['ref1.jpg'], 'horizontal')
      expect(request.put).toHaveBeenCalledWith('/scenes/123/prompt', {
        name: 'Scene Name',
        prompt: 'prompt',
        reference_images: ['ref1.jpg'],
        image_orientation: 'horizontal'
      })
    })
  })

  describe('deleteScene', () => {
    it('应该调用 DELETE /scenes/:id', () => {
      dramaAPI.deleteScene('123')
      expect(request.delete).toHaveBeenCalledWith('/scenes/123')
    })
  })

  describe('finalizeEpisode', () => {
    it('应该调用 POST /episodes/:id/finalize', () => {
      dramaAPI.finalizeEpisode('123')
      expect(request.post).toHaveBeenCalledWith('/episodes/123/finalize', {})
    })

    it('应该包含 timelineData 当提供时', () => {
      const timelineData = { scenes: [] }
      dramaAPI.finalizeEpisode('123', timelineData)
      expect(request.post).toHaveBeenCalledWith('/episodes/123/finalize', timelineData)
    })
  })

  describe('createStoryboard', () => {
    it('应该调用 POST /storyboards', () => {
      const data = {
        episode_id: 1,
        storyboard_number: 1,
        duration: 5
      }
      dramaAPI.createStoryboard(data)
      expect(request.post).toHaveBeenCalledWith('/storyboards', data)
    })
  })

  describe('deleteStoryboard', () => {
    it('应该调用 DELETE /storyboards/:id', () => {
      dramaAPI.deleteStoryboard(123)
      expect(request.delete).toHaveBeenCalledWith('/storyboards/123')
    })
  })

  describe('generateVideoPrompt', () => {
    it('应该调用 POST /storyboards/:id/video-prompt', () => {
      dramaAPI.generateVideoPrompt(123, 'model', 5)
      expect(request.post).toHaveBeenCalledWith('/storyboards/123/video-prompt', {
        model: 'model',
        duration: 5
      })
    })

    it('应该不带可选参数', () => {
      dramaAPI.generateVideoPrompt(123)
      expect(request.post).toHaveBeenCalledWith('/storyboards/123/video-prompt', {
        model: undefined,
        duration: undefined
      })
    })
  })
})

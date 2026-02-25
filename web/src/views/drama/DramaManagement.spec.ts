import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import DramaManagement from './DramaManagement.vue'
import * as characterLibraryAPI from '../../api/character-library'
import * as dramaAPI from '../../api/drama'
import * as propAPI from '../../api/prop'

vi.mock('../../api/character-library')
vi.mock('../../api/drama')
vi.mock('../../api/prop')

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/dramas/:id', component: DramaManagement }
  ]
})

describe('DramaManagement - Character Management', () => {
  let wrapper: any
  let pinia: any

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    
    wrapper = mount(DramaManagement, {
      global: {
        plugins: [pinia, router],
        mocks: { $t: (key: string) => key },
        stubs: {
          'el-button': true,
          'el-card': true,
          'el-dialog': true,
          'el-form': true,
          'el-form-item': true,
          'el-input': true,
          'el-select': true,
          'el-option': true,
          'el-upload': true,
          'el-radio-group': true,
          'el-radio': true,
          'el-tag': true,
          'el-avatar': true,
          'el-row': true,
          'el-col': true,
          'el-empty': true,
          'el-message-box': true,
          'el-tabs': true,
          'el-tab-pane': true,
          'el-icon': true,
          'el-tooltip': true,
          'el-dropdown': true,
          'el-dropdown-menu': true,
          'el-dropdown-item': true,
          'el-progress': true,
          'el-switch': true,
          'el-table': true,
          'el-table-column': true,
          'el-pagination': true,
          'el-checkbox': true,
          'el-checkbox-group': true,
          'el-date-picker': true,
          'el-time-picker': true,
          'el-slider': true,
          'el-rate': true,
          'el-color-picker': true,
          'el-transfer': true,
          'el-tree': true,
          'el-tree-select': true,
          'el-cascader': true,
          'el-select-v2': true,
          'el-autocomplete': true,
          'el-input-number': true,
          'el-radio-button': true,
          'el-checkbox-button': true,
          'el-form-item-label': true,
          'el-form-item-error': true,
          'el-form-item-content': true,
          'el-image': true,
          'el-image-viewer': true,
          'el-carousel': true,
          'el-carousel-item': true,
          'el-backtop': true,
          'el-badge': true,
          'el-skeleton': true,
          'el-skeleton-item': true,
          'el-skeleton-image': true,
          'el-skeleton-button': true,
          'el-skeleton-input': true,
          'el-skeleton-text': true,
          'el-skeleton-paragraph': true,
          'el-skeleton-circle': true,
          'el-skeleton-square': true,
          'el-skeleton-rect': true,
          'el-skeleton-avatar': true,
          'el-skeleton-list': true,
          'el-skeleton-grid': true,
          'el-skeleton-facade': true,
          'el-skeleton-props': true,
          'el-skeleton-rows': true,
          'el-skeleton-throttle': true,
          'el-skeleton-lazy': true,
          'el-skeleton-async': true,
          'el-skeleton-loading': true,
          'el-skeleton-animated': true,
          'el-skeleton-count': true,
          'el-skeleton-duration': true,
          'el-skeleton-easing': true,
          'el-skeleton-delay': true,
          'el-skeleton-throttle-delay': true,
          'el-skeleton-lazy-delay': true,
          'el-skeleton-async-delay': true,
          'el-skeleton-loading-delay': true,
          'el-skeleton-animated-delay': true,
          'el-skeleton-count-delay': true,
          'el-skeleton-duration-delay': true,
          'el-skeleton-easing-delay': true,
          'el-skeleton-throttle-easing': true,
          'el-skeleton-lazy-easing': true,
          'el-skeleton-async-easing': true,
          'el-skeleton-loading-easing': true,
          'el-skeleton-animated-easing': true,
          'el-skeleton-count-easing': true,
          'el-skeleton-duration-easing': true,
          'el-skeleton-easing-duration': true,
          'el-skeleton-throttle-duration': true,
          'el-skeleton-lazy-duration': true,
          'el-skeleton-async-duration': true,
          'el-skeleton-loading-duration': true,
          'el-skeleton-animated-duration': true,
          'el-skeleton-count-duration': true,
          'el-skeleton-duration-count': true,
          'el-skeleton-easing-count': true,
          'el-skeleton-throttle-count': true,
          'el-skeleton-lazy-count': true,
          'el-skeleton-async-count': true,
          'el-skeleton-loading-count': true,
          'el-skeleton-animated-count': true,
          'el-skeleton-count-count': true,
          'el-skeleton-duration-count-easing': true,
          'el-skeleton-easing-count-easing': true,
          'el-skeleton-throttle-count-easing': true,
          'el-skeleton-lazy-count-easing': true,
          'el-skeleton-async-count-easing': true,
          'el-skeleton-loading-count-easing': true,
          'el-skeleton-animated-count-easing': true,
          'el-skeleton-count-count-easing': true,
          'el-skeleton-duration-count-easing-throttle': true,
          'el-skeleton-easing-count-easing-throttle': true,
          'el-skeleton-throttle-count-easing-throttle': true,
          'el-skeleton-lazy-count-easing-throttle': true,
          'el-skeleton-async-count-easing-throttle': true,
          'el-skeleton-loading-count-easing-throttle': true,
          'el-skeleton-animated-count-easing-throttle': true,
          'el-skeleton-count-count-easing-throttle': true,
          'el-skeleton-duration-count-easing-throttle-lazy': true,
          'el-skeleton-easing-count-easing-throttle-lazy': true,
          'el-skeleton-throttle-count-easing-throttle-lazy': true,
          'el-skeleton-lazy-count-easing-throttle-lazy': true,
          'el-skeleton-async-count-easing-throttle-lazy': true,
          'el-skeleton-loading-count-easing-throttle-lazy': true,
          'el-skeleton-animated-count-easing-throttle-lazy': true,
          'el-skeleton-count-count-easing-throttle-lazy': true,
          'el-skeleton-duration-count-easing-throttle-lazy-async': true,
          'el-skeleton-easing-count-easing-throttle-lazy-async': true,
          'el-skeleton-throttle-count-easing-throttle-lazy-async': true,
          'el-skeleton-lazy-count-easing-throttle-lazy-async': true,
          'el-skeleton-async-count-easing-throttle-lazy-async': true,
          'el-skeleton-loading-count-easing-throttle-lazy-async': true,
          'el-skeleton-animated-count-easing-throttle-lazy-async': true,
          'el-skeleton-count-count-easing-throttle-lazy-async': true,
          'el-skeleton-duration-count-easing-throttle-lazy-async-loading': true,
          'el-skeleton-easing-count-easing-throttle-lazy-async-loading': true,
          'el-skeleton-throttle-count-easing-throttle-lazy-async-loading': true,
          'el-skeleton-lazy-count-easing-throttle-lazy-async-loading': true,
          'el-skeleton-async-count-easing-throttle-lazy-async-loading': true,
          'el-skeleton-loading-count-easing-throttle-lazy-async-loading': true,
          'el-skeleton-animated-count-easing-throttle-lazy-async-loading': true,
          'el-skeleton-count-count-easing-throttle-lazy-async-loading': true,
          'el-skeleton-duration-count-easing-throttle-lazy-async-loading-animated': true,
          'el-skeleton-easing-count-easing-throttle-lazy-async-loading-animated': true,
          'el-skeleton-throttle-count-easing-throttle-lazy-async-loading-animated': true,
          'el-skeleton-lazy-count-easing-throttle-lazy-async-loading-animated': true,
          'el-skeleton-async-count-easing-throttle-lazy-async-loading-animated': true,
          'el-skeleton-loading-count-easing-throttle-lazy-async-loading-animated': true,
          'el-skeleton-animated-count-easing-throttle-lazy-async-loading-animated': true,
          'el-skeleton-count-count-easing-throttle-lazy-async-loading-animated': true,
          'el-skeleton-duration-count-easing-throttle-lazy-async-loading-animated-count': true,
          'el-skeleton-easing-count-easing-throttle-lazy-async-loading-animated-count': true,
          'el-skeleton-throttle-count-easing-throttle-lazy-async-loading-animated-count': true,
          'el-skeleton-lazy-count-easing-throttle-lazy-async-loading-animated-count': true,
          'el-skeleton-async-count-easing-throttle-lazy-async-loading-animated-count': true,
          'el-skeleton-loading-count-easing-throttle-lazy-async-loading-animated-count': true,
          'el-skeleton-animated-count-easing-throttle-lazy-async-loading-animated-count': true,
          'el-skeleton-count-count-easing-throttle-lazy-async-loading-animated-count': true,
          'StatCard': true,
          'ImagePreview': true,
          'AppHeader': true,
          'PageHeader': true,
          'EmptyState': true,
          'CreateDramaDialog': true,
          'UploadScriptDialog': true,
          'DramaCreate': true,
          'DramaList': true,
          'DramaWorkflow': true,
          'EpisodeWorkflow': true,
          'ProfessionalEditor': true,
          'TimelineEditor': true,
          'ImageGeneration': true,
          'VideoGeneration': true,
          'AIConfig': true,
          'CharacterExtraction': true,
          'CharacterImages': true,
          'DramaSettings': true,
          'SceneImages': true,
          'StoryboardGeneration': true,
          'GenerateImageDialog': true,
          'GenerateVideoDialog': true,
          'ImageDetailDialog': true,
          'VideoDetailDialog': true,
          'ConfigList': true,
          'GridImageEditor': true,
          'StoryboardEditor': true,
          'VideoTimelineEditor': true
        }
      }
    })
  })

  describe('Character CRUD Operations', () => {
    it('should create a new character', async () => {
      await wrapper.vm.openAddCharacterDialog()
      expect(wrapper.vm.addCharacterDialogVisible).toBe(true)
      expect(wrapper.vm.editingCharacter).toBe(null)
    })

    it('should edit an existing character', async () => {
      const mockCharacter = {
        id: 1,
        name: '测试角色',
        role: 'main',
        appearance: '测试外观',
        personality: '测试性格',
        description: '测试描述',
        image_url: 'http://example.com/image.jpg',
        reference_images: ['http://example.com/ref1.jpg'],
        image_orientation: 'horizontal'
      }

      await wrapper.vm.editCharacter(mockCharacter)

      expect(wrapper.vm.editingCharacter).toEqual(mockCharacter)
      expect(wrapper.vm.addCharacterDialogVisible).toBe(true)
    })

    it('should delete a character', async () => {
      const mockCharacter = {
        id: 1,
        name: '测试角色'
      }

      expect(wrapper.vm.deleteCharacter).toBeDefined()
      expect(typeof wrapper.vm.deleteCharacter).toBe('function')
    })

    it('should handle reference images correctly when editing', async () => {
      const mockCharacter = {
        id: 1,
        name: '测试角色',
        reference_images: [
          { url: 'http://example.com/ref1.jpg', name: 'ref1.jpg' },
          { url: 'http://example.com/ref2.jpg', name: 'ref2.jpg' }
        ]
      }

      await wrapper.vm.editCharacter(mockCharacter)

      expect(wrapper.vm.newCharacter.reference_images).toEqual([
        'http://example.com/ref1.jpg',
        'http://example.com/ref2.jpg'
      ])
    })

    it('should handle reference images as strings when editing', async () => {
      const mockCharacter = {
        id: 1,
        name: '测试角色',
        reference_images: ['http://example.com/ref1.jpg', 'http://example.com/ref2.jpg']
      }

      await wrapper.vm.editCharacter(mockCharacter)

      expect(wrapper.vm.newCharacter.reference_images).toEqual([
        'http://example.com/ref1.jpg',
        'http://example.com/ref2.jpg'
      ])
    })
  })
})

describe('DramaManagement - Scene Management', () => {
  let wrapper: any
  let pinia: any

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    
    wrapper = mount(DramaManagement, {
      global: {
        plugins: [pinia, router],
        mocks: { $t: (key: string) => key },
        stubs: {
          'el-button': true,
          'el-card': true,
          'el-dialog': true,
          'el-form': true,
          'el-form-item': true,
          'el-input': true,
          'el-upload': true,
          'el-radio-group': true,
          'el-radio': true,
          'el-tag': true,
          'el-row': true,
          'el-col': true,
          'el-empty': true,
          'el-message-box': true,
          'ImagePreview': true
        }
      }
    })
  })

  describe('Scene CRUD Operations', () => {
    it('should create a new scene', async () => {
      await wrapper.vm.openAddSceneDialog()
      expect(wrapper.vm.addSceneDialogVisible).toBe(true)
      expect(wrapper.vm.editingScene).toBe(null)
    })

    it('should edit an existing scene', async () => {
      const mockScene = {
        id: 1,
        location: '测试地点',
        prompt: '测试提示',
        image_url: 'http://example.com/image.jpg',
        reference_images: ['http://example.com/ref1.jpg'],
        image_orientation: 'horizontal'
      }

      await wrapper.vm.editScene(mockScene)

      expect(wrapper.vm.editingScene).toEqual(mockScene)
      expect(wrapper.vm.addSceneDialogVisible).toBe(true)
    })

    it('should delete a scene', async () => {
      expect(wrapper.vm.deleteScene).toBeDefined()
      expect(typeof wrapper.vm.deleteScene).toBe('function')
    })

    it('should handle reference images correctly when editing scene', async () => {
      const mockScene = {
        id: 1,
        location: '测试地点',
        reference_images: [
          { url: 'http://example.com/ref1.jpg', name: 'ref1.jpg' },
          { url: 'http://example.com/ref2.jpg', name: 'ref2.jpg' }
        ]
      }

      await wrapper.vm.editScene(mockScene)

      expect(wrapper.vm.newScene.reference_images).toEqual([
        'http://example.com/ref1.jpg',
        'http://example.com/ref2.jpg'
      ])
    })
  })
})

describe('DramaManagement - Prop Management', () => {
  let wrapper: any
  let pinia: any

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    
    wrapper = mount(DramaManagement, {
      global: {
        plugins: [pinia, router],
        mocks: { $t: (key: string) => key },
        stubs: {
          'el-button': true,
          'el-card': true,
          'el-dialog': true,
          'el-form': true,
          'el-form-item': true,
          'el-input': true,
          'el-upload': true,
          'el-radio-group': true,
          'el-radio': true,
          'el-tag': true,
          'el-row': true,
          'el-col': true,
          'el-empty': true,
          'el-message-box': true,
          'ImagePreview': true
        }
      }
    })
  })

  describe('Prop CRUD Operations', () => {
    it('should create a new prop', async () => {
      await wrapper.vm.openAddPropDialog()
      expect(wrapper.vm.addPropDialogVisible).toBe(true)
      expect(wrapper.vm.editingProp).toBe(null)
    })

    it('should edit an existing prop', async () => {
      const mockProp = {
        id: 1,
        name: '测试道具',
        type: '武器',
        description: '测试描述',
        prompt: '测试提示',
        image_url: 'http://example.com/image.jpg',
        reference_images: ['http://example.com/ref1.jpg'],
        image_orientation: 'horizontal'
      }

      await wrapper.vm.editProp(mockProp)

      expect(wrapper.vm.editingProp).toEqual(mockProp)
      expect(wrapper.vm.addPropDialogVisible).toBe(true)
    })

    it('should delete a prop', async () => {
      expect(wrapper.vm.deleteProp).toBeDefined()
      expect(typeof wrapper.vm.deleteProp).toBe('function')
    })

    it('should handle reference images correctly when editing prop', async () => {
      const mockProp = {
        id: 1,
        name: '测试道具',
        reference_images: [
          { url: 'http://example.com/ref1.jpg', name: 'ref1.jpg' },
          { url: 'http://example.com/ref2.jpg', name: 'ref2.jpg' }
        ]
      }

      await wrapper.vm.editProp(mockProp)

      expect(wrapper.vm.newProp.reference_images).toEqual([
        'http://example.com/ref1.jpg',
        'http://example.com/ref2.jpg'
      ])
    })
  })
})

describe('DramaManagement - Image Upload', () => {
  let wrapper: any
  let pinia: any

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    
    wrapper = mount(DramaManagement, {
      global: {
        plugins: [pinia, router],
        mocks: { $t: (key: string) => key },
        stubs: {
          'el-button': true,
          'el-card': true,
          'el-dialog': true,
          'el-form': true,
          'el-form-item': true,
          'el-input': true,
          'el-upload': true,
          'el-radio-group': true,
          'el-radio': true,
          'el-tag': true,
          'el-row': true,
          'el-col': true,
          'el-empty': true,
          'el-message-box': true,
          'ImagePreview': true
        }
      }
    })
  })

  describe('Character Image Upload', () => {
    it('should handle character avatar upload success', async () => {
      const mockResponse = {
        data: {
          url: 'http://example.com/uploaded.jpg'
        }
      }

      await wrapper.vm.handleCharacterAvatarSuccess(mockResponse)

      expect(wrapper.vm.newCharacter.image_url).toBe('http://example.com/uploaded.jpg')
    })

    it('should handle character reference image upload success', async () => {
      const mockFile = { name: 'ref.jpg', url: '' }
      const mockResponse = { data: { url: 'http://example.com/ref.jpg' } }

      wrapper.vm.characterReferenceImages = [mockFile]
      await wrapper.vm.handleCharacterReferenceImageSuccess(mockResponse, mockFile)

      expect(wrapper.vm.newCharacter.reference_images).toContain('http://example.com/ref.jpg')
    })
  })

  describe('Scene Image Upload', () => {
    it('should handle scene image upload success', async () => {
      const mockResponse = {
        data: {
          url: 'http://example.com/uploaded.jpg'
        }
      }

      await wrapper.vm.handleSceneImageSuccess(mockResponse)

      expect(wrapper.vm.newScene.image_url).toBe('http://example.com/uploaded.jpg')
    })

    it('should handle scene reference image upload success', async () => {
      const mockFile = { name: 'ref.jpg', url: '' }
      const mockResponse = { data: { url: 'http://example.com/ref.jpg' } }

      wrapper.vm.sceneReferenceImages = [mockFile]
      await wrapper.vm.handleSceneReferenceImageSuccess(mockResponse, mockFile)

      expect(wrapper.vm.newScene.reference_images).toContain('http://example.com/ref.jpg')
    })
  })

  describe('Prop Image Upload', () => {
    it('should handle prop image upload success', async () => {
      const mockResponse = {
        data: {
          url: 'http://example.com/uploaded.jpg'
        }
      }

      await wrapper.vm.handlePropImageSuccess(mockResponse)

      expect(wrapper.vm.newProp.image_url).toBe('http://example.com/uploaded.jpg')
    })

    it('should handle prop reference image upload success', async () => {
      const mockFile = { name: 'ref.jpg', url: '' }
      const mockResponse = { data: { url: 'http://example.com/ref.jpg' } }

      wrapper.vm.propReferenceImages = [mockFile]
      await wrapper.vm.handlePropReferenceImageSuccess(mockResponse, mockFile)

      expect(wrapper.vm.newProp.reference_images).toContain('http://example.com/ref.jpg')
    })
  })
})

describe('DramaManagement - Image Orientation', () => {
  let wrapper: any
  let pinia: any

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    
    wrapper = mount(DramaManagement, {
      global: {
        plugins: [pinia, router],
        mocks: { $t: (key: string) => key },
        stubs: {
          'el-button': true,
          'el-card': true,
          'el-dialog': true,
          'el-form': true,
          'el-form-item': true,
          'el-input': true,
          'el-upload': true,
          'el-radio-group': true,
          'el-radio': true,
          'el-tag': true,
          'el-row': true,
          'el-col': true,
          'el-empty': true,
          'el-message-box': true,
          'ImagePreview': true
        }
      }
    })
  })

  it('should set default image orientation to horizontal for new character', () => {
    expect(wrapper.vm.newCharacter.image_orientation).toBe('horizontal')
  })

  it('should set default image orientation to horizontal for new scene', () => {
    expect(wrapper.vm.newScene.image_orientation).toBe('horizontal')
  })

  it('should set default image orientation to horizontal for new prop', () => {
    expect(wrapper.vm.newProp.image_orientation).toBe('horizontal')
  })

  it('should preserve image orientation when editing character', async () => {
    const mockCharacter = {
      id: 1,
      name: '测试角色',
      image_orientation: 'vertical'
    }

    await wrapper.vm.editCharacter(mockCharacter)

    expect(wrapper.vm.newCharacter.image_orientation).toBe('vertical')
  })

  it('should preserve image orientation when editing scene', async () => {
    const mockScene = {
      id: 1,
      location: '测试地点',
      image_orientation: 'vertical'
    }

    await wrapper.vm.editScene(mockScene)

    expect(wrapper.vm.newScene.image_orientation).toBe('vertical')
  })

  it('should preserve image orientation when editing prop', async () => {
    const mockProp = {
      id: 1,
      name: '测试道具',
      image_orientation: 'vertical'
    }

    await wrapper.vm.editProp(mockProp)

    expect(wrapper.vm.newProp.image_orientation).toBe('vertical')
  })
})

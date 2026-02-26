import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ActionButton from './ActionButton.vue'
import { Delete } from '@element-plus/icons-vue'

describe('ActionButton.vue', () => {
  describe('基础渲染', () => {
    it('应该渲染图标按钮', () => {
      const wrapper = mount(ActionButton, {
        props: { icon: Delete }
      })
      expect(wrapper.find('button').exists()).toBe(true)
    })

    it('应该渲染图标', () => {
      const wrapper = mount(ActionButton, {
        props: { icon: Delete }
      })
      expect(wrapper.html()).toContain('svg')
    })
  })

  describe('variant', () => {
    it('应该有 default variant 类', () => {
      const wrapper = mount(ActionButton, {
        props: { icon: Delete }
      })
      expect(wrapper.find('button').classes()).toContain('default')
    })

    it('应该有 primary variant 类', () => {
      const wrapper = mount(ActionButton, {
        props: { icon: Delete, variant: 'primary' }
      })
      expect(wrapper.find('button').classes()).toContain('primary')
    })

    it('应该有 danger variant 类', () => {
      const wrapper = mount(ActionButton, {
        props: { icon: Delete, variant: 'danger' }
      })
      expect(wrapper.find('button').classes()).toContain('danger')
    })
  })

  describe('disabled', () => {
    it('不应该有 disabled 类当未禁用时', () => {
      const wrapper = mount(ActionButton, {
        props: { icon: Delete, disabled: false }
      })
      const button = wrapper.find('button')
      expect(button.classes()).not.toContain('disabled')
      expect(button.attributes('disabled')).toBeUndefined()
    })

    it('应该有 disabled 类当禁用时', () => {
      const wrapper = mount(ActionButton, {
        props: { icon: Delete, disabled: true }
      })
      const button = wrapper.find('button')
      expect(button.classes()).toContain('disabled')
      expect(button.attributes('disabled')).toBeDefined()
    })
  })

  describe('事件', () => {
    it('应该触发 click 事件', async () => {
      const wrapper = mount(ActionButton, {
        props: { icon: Delete }
      })
      
      await wrapper.find('button').trigger('click')
      expect(wrapper.emitted('click')).toBeTruthy()
      expect(wrapper.emitted('click')).toHaveLength(1)
    })

    it('不应该触发 click 事件当禁用时', async () => {
      const wrapper = mount(ActionButton, {
        props: { icon: Delete, disabled: true }
      })
      
      await wrapper.find('button').trigger('click')
      expect(wrapper.emitted('click')).toBeFalsy()
    })
  })

  describe('样式类', () => {
    it('应该有 action-button 类', () => {
      const wrapper = mount(ActionButton, {
        props: { icon: Delete }
      })
      expect(wrapper.find('button').classes()).toContain('action-button')
    })
  })
})

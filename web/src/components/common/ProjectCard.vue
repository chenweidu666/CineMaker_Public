<template>
  <!-- Project card component - Compact design with hover actions -->
  <!-- 项目卡片组件 - 紧凑设计，悬停显示操作 -->
  <article 
    class="project-card"
    @click="$emit('click')"
    tabindex="0"
    @keydown.enter="$emit('click')"
  >
    <!-- Gradient header with icon / 渐变头部区域 -->
    <div class="card-header">
      <el-icon class="header-icon"><Film /></el-icon>
      <!-- Hover actions / 悬停操作区 -->
      <div class="hover-actions" @click.stop>
        <slot name="actions"></slot>
      </div>
    </div>

    <!-- Card content / 卡片内容 -->
    <div class="card-body">
      <h3 class="card-title">{{ title }}</h3>
      <p v-if="description" class="card-description">{{ description }}</p>
      
      <!-- Footer section / 底部区域 -->
      <div class="card-footer">
        <div class="footer-row">
          <span class="meta-time">{{ formattedDate }}</span>
          <span class="episode-label">共 {{ episodeCount }} 集</span>
        </div>
        <span v-if="styleLabel" class="style-tag" :class="'style-' + dramaStyle">{{ styleLabel }}</span>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Film } from '@element-plus/icons-vue'

/**
 * ProjectCard - Reusable project/drama card component
 * 项目卡片组件 - 可复用的项目展示卡片
 */
const props = withDefaults(defineProps<{
  title: string
  description?: string
  updatedAt: string
  episodeCount?: number
  dramaStyle?: string
}>(), {
  description: '',
  episodeCount: 0,
  dramaStyle: ''
})

const styleNameMap: Record<string, string> = {
  realistic: '写实',
  comic: '漫画',
}

const styleLabel = computed(() => {
  if (!props.dramaStyle) return ''
  return styleNameMap[props.dramaStyle] || props.dramaStyle
})

defineEmits<{
  click: []
}>()

// Format date / 格式化日期
const formattedDate = computed(() => {
  const date = new Date(props.updatedAt)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
})
</script>

<style scoped>
/* Card Container / 卡片容器 */
.project-card {
  position: relative;
  display: flex;
  flex-direction: column;
  background: var(--glass-bg-heavy);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-xl);
  overflow: hidden;
  cursor: pointer;
  transition: all var(--transition-normal);
  width: 200px;
}

.project-card:hover {
  border-color: var(--accent);
  transform: translateY(-3px);
  box-shadow: var(--shadow-card-hover), 0 0 20px rgba(14, 165, 233, 0.1);
}

.project-card:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

/* Card Header / 卡片头部 */
.card-header {
  position: relative;
  height: 120px;
  background: linear-gradient(135deg, var(--accent) 0%, #06b6d4 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-icon {
  font-size: 28px;
  color: rgba(255, 255, 255, 0.8);
}

/* Hover Actions / 悬停操作区 */
.hover-actions {
  position: absolute;
  top: 8px;
  right: 8px;
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity var(--transition-fast);
  z-index: 10;
}

.project-card:hover .hover-actions {
  opacity: 1;
}

/* Body Section / 内容区域 */
.card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 12px;
  gap: 10px;
}

.card-title {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-description {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.4;
}

/* Footer Section / 底部区域 */
.card-footer {
  margin-top: auto;
  padding-top: 8px;
  border-top: 1px solid var(--glass-border);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.footer-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.meta-time,
.episode-label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.style-tag {
  display: inline-block;
  width: fit-content;
  padding: 2px 10px;
  font-size: 0.7rem;
  font-weight: 500;
  border-radius: var(--radius-full);
  color: #fff;
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
}

.style-realistic { background: rgba(59, 130, 246, 0.85); }
.style-comic { background: rgba(245, 158, 11, 0.85); }

:deep(.action-button) {
  width: 28px !important;
  height: 28px !important;
  padding: 0 !important;
  background: var(--bg-secondary) !important;
  border: none !important;
}
</style>

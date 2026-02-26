<template>
  <div class="detail-section">
    <div class="detail-label">{{ label }}</div>
    <div
      class="detail-value collapsible-text"
      :class="{ collapsed: !expanded, expandable: needsCollapse }"
    >
      <span ref="textRef">{{ text }}</span>
    </div>
    <button v-if="needsCollapse" class="toggle-btn" @click="expanded = !expanded">
      {{ expanded ? '收起' : '展开全部' }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'

const props = defineProps<{ label: string; text: string; defaultCollapsed?: boolean }>()

const expanded = ref(!props.defaultCollapsed)
const needsCollapse = ref(false)
const textRef = ref<HTMLElement>()

const COLLAPSED_HEIGHT = 80

onMounted(async () => {
  await nextTick()
  if (textRef.value && textRef.value.scrollHeight > COLLAPSED_HEIGHT + 10) {
    needsCollapse.value = true
  }
})
</script>

<style>
.collapsible-text {
  position: relative;
  overflow: hidden;
  transition: max-height 0.25s ease;
}

.collapsible-text.collapsed {
  max-height: 80px;
}

.collapsible-text.collapsed.expandable::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 28px;
  background: linear-gradient(transparent, var(--el-bg-color, #fff));
  pointer-events: none;
}

.toggle-btn {
  display: inline-block;
  margin-top: 4px;
  padding: 0;
  border: none;
  background: none;
  color: var(--el-color-primary, #409eff);
  font-size: 12px;
  cursor: pointer;
}

.toggle-btn:hover {
  opacity: 0.8;
}
</style>

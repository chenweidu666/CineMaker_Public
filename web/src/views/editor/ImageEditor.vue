<template>
  <div class="image-editor-page page-container">
    <!-- 顶栏 -->
    <header class="editor-topbar glass-card">
      <div class="topbar-left">
        <button class="back-btn" @click="router.back()">
          <el-icon :size="16"><ArrowLeft /></el-icon>
          <span>返回</span>
        </button>
        <div class="topbar-divider" />
        <h1 class="topbar-title">图片微调</h1>
        <el-tag v-if="sourceImage" type="info" size="small" effect="plain">
          #{{ sourceImage.id }}
          <template v-if="sourceImage.frame_type"> · {{ frameTypeLabel(sourceImage.frame_type) }}</template>
        </el-tag>
      </div>
    </header>

    <!-- 主体 -->
    <main class="editor-main" v-loading="loading">
      <!-- 左侧：原图 -->
      <section class="main-preview">
        <div class="preview-card glass-card">
          <div class="preview-card-header">
            <span class="card-title">原始图片</span>
            <el-tag v-if="sourceImage" size="small" type="success" effect="plain">
              {{ sourceImage.size || '—' }}
            </el-tag>
          </div>
          <div class="preview-canvas">
            <el-image
              v-if="originalImageUrl"
              :src="originalImageUrl"
              fit="contain"
              :preview-src-list="[originalImageUrl]"
              preview-teleported
              class="canvas-img"
            />
            <div v-else class="canvas-empty">
              <el-icon :size="48" color="var(--text-muted)"><Picture /></el-icon>
              <p>加载中...</p>
            </div>
          </div>
        </div>
      </section>

      <!-- 右侧：控制面板 -->
      <aside class="main-sidebar">
        <div class="sidebar-card glass-card">
          <div class="card-title">编辑指令</div>

          <!-- 快捷指令 -->
          <div class="preset-chips">
            <button
              v-for="preset in presetInstructions"
              :key="preset.label"
              class="chip"
              :class="{ active: editPrompt === preset.prompt }"
              @click="editPrompt = preset.prompt"
            >{{ preset.label }}</button>
          </div>

          <!-- 输入框 -->
          <el-input
            v-model="editPrompt"
            type="textarea"
            :rows="4"
            placeholder="描述需要修改的内容，例如：让两个人的眼神四目相对、修正手部姿态..."
            maxlength="1000"
            show-word-limit
            resize="none"
          />

          <!-- 高级选项 -->
          <el-collapse class="adv-collapse">
            <el-collapse-item title="高级选项" name="advanced">
              <div class="adv-row">
                <label>引导强度</label>
                <el-slider v-model="guidanceScale" :min="1" :max="15" :step="0.5" style="flex:1" />
                <span class="adv-val">{{ guidanceScale }}</span>
              </div>
              <div class="adv-row">
                <label>随机种子</label>
                <el-input-number v-model="seed" :min="0" :max="999999999" controls-position="right" style="width:140px" placeholder="留空随机" />
                <el-button text size="small" @click="seed = undefined">重置</el-button>
              </div>
            </el-collapse-item>
          </el-collapse>

          <!-- 提交 -->
          <el-button
            type="primary"
            size="large"
            :loading="generating"
            :disabled="!editPrompt.trim() || !sourceImage"
            @click="submitEdit"
            class="submit-btn"
          >
            <el-icon v-if="!generating"><MagicStick /></el-icon>
            {{ generating ? '正在编辑...' : '开始编辑' }}
          </el-button>
        </div>
      </aside>
    </main>

    <!-- 编辑结果 -->
    <section class="results-strip" v-if="editResults.length > 0">
      <div class="strip-header">
        <span class="card-title">编辑结果</span>
        <el-tag size="small" type="info" effect="plain">{{ editResults.length }} 张</el-tag>
      </div>
      <div class="strip-scroll custom-scrollbar">
        <div
          v-for="item in editResults"
          :key="item.id"
          class="result-card"
          :class="{
            clickable: item.status === 'completed',
            'is-processing': item.status === 'pending' || item.status === 'processing'
          }"
          @click="openPreview(item)"
        >
          <div class="result-thumb">
            <!-- 完成 -->
            <el-image
              v-if="item.status === 'completed' && getImageUrl(item)"
              :src="getImageUrl(item)!"
              fit="cover"
              lazy
              loading="lazy"
              class="thumb-img"
            />
            <!-- 加载中 -->
            <div v-else-if="item.status === 'pending' || item.status === 'processing'" class="thumb-state">
              <div class="pulse-ring" />
              <el-icon class="is-loading" :size="22"><Loading /></el-icon>
              <span>生成中...</span>
            </div>
            <!-- 失败 -->
            <div v-else-if="item.status === 'failed'" class="thumb-state failed">
              <el-icon :size="22"><CircleCloseFilled /></el-icon>
              <span>失败</span>
            </div>
          </div>
          <div class="result-meta">
            <p class="result-prompt" :title="item.prompt">{{ item.prompt }}</p>
            <el-tag
              :type="statusTagType(item.status)"
              size="small"
              effect="plain"
            >{{ statusLabel(item.status) }}</el-tag>
          </div>
          <!-- hover 覆盖层 -->
          <div class="result-overlay" v-if="item.status === 'completed'">
            <el-icon :size="20"><ZoomIn /></el-icon>
            <span>查看 & 替换</span>
          </div>
        </div>
      </div>
    </section>

    <!-- 预览弹窗 -->
    <el-dialog
      v-model="previewVisible"
      width="88%"
      top="3vh"
      :show-close="true"
      :close-on-click-modal="true"
      destroy-on-close
      class="preview-dialog"
    >
      <template #header>
        <div class="dialog-title-row">
          <span class="dialog-title">对比预览</span>
          <el-tag type="info" size="small" effect="plain" v-if="previewImage">
            编辑 #{{ previewImage.id }}
          </el-tag>
        </div>
      </template>

      <div class="dialog-body" v-if="previewImage">
        <div class="compare-grid">
          <div class="compare-card">
            <div class="compare-badge">原图</div>
            <el-image :src="originalImageUrl" fit="contain" class="compare-img" />
          </div>
          <div class="compare-card">
            <div class="compare-badge accent">编辑结果</div>
            <el-image :src="getImageUrl(previewImage)!" fit="contain" class="compare-img" />
          </div>
        </div>
        <div class="dialog-prompt-bar">
          <el-icon :size="14" color="var(--accent)"><MagicStick /></el-icon>
          <span>{{ previewImage.prompt }}</span>
        </div>
      </div>

      <template #footer>
        <div class="dialog-actions">
          <el-button @click="previewVisible = false">关闭</el-button>
          <el-button type="danger" plain @click="deleteEditResult" :loading="deleting">
            删除此结果
          </el-button>
          <el-button type="primary" @click="confirmReplace" :loading="replacing">
            确认替换原图
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, MagicStick, Loading, CircleCloseFilled, Picture, ZoomIn } from '@element-plus/icons-vue'
import { imageAPI } from '@/api/image'
import type { ImageGeneration } from '@/types/image'
import { getImageUrl as resolveImageUrl } from '@/utils/image'

const route = useRoute()
const router = useRouter()

const sourceImage = ref<ImageGeneration | null>(null)
const editResults = ref<ImageGeneration[]>([])
const editPrompt = ref('')
const guidanceScale = ref(5.5)
const seed = ref<number | undefined>(undefined)
const loading = ref(false)
const generating = ref(false)
const replacing = ref(false)
const deleting = ref(false)

const previewVisible = ref(false)
const previewImage = ref<ImageGeneration | null>(null)

const pollingIds = ref<Set<number>>(new Set())
const pollingTimer = ref<ReturnType<typeof setInterval> | null>(null)

const presetInstructions = [
  { label: '眼神对视', prompt: '让两个人物的眼神四目相对，目光注视对方' },
  { label: '修正手部', prompt: '修正人物手部姿态，使其自然协调' },
  { label: '调整表情', prompt: '让人物面部表情更加自然放松' },
  { label: '增强光影', prompt: '增强画面的光影效果，使光线更加柔和自然' },
  { label: '背景虚化', prompt: '将背景进行适度虚化，突出前景人物' },
  { label: '调整构图', prompt: '优化画面构图，使主体更居中突出' },
]

const originalImageUrl = computed(() => {
  if (!sourceImage.value) return ''
  return getImageUrl(sourceImage.value) || ''
})

function getImageUrl(img: ImageGeneration): string | null {
  return resolveImageUrl(img) || null
}

function frameTypeLabel(ft: string): string {
  const m: Record<string, string> = { first: '首帧', last: '尾帧', key: '关键帧', action: '动作序列', panel: '分镜板' }
  return m[ft] || ft
}

function statusLabel(s: string): string {
  const m: Record<string, string> = { pending: '等待中', processing: '生成中', completed: '完成', failed: '失败' }
  return m[s] || s
}

function statusTagType(s: string): 'success' | 'warning' | 'danger' | 'info' {
  const m: Record<string, 'success' | 'warning' | 'danger' | 'info'> = {
    completed: 'success', processing: 'warning', pending: 'info', failed: 'danger'
  }
  return m[s] || 'info'
}

/* ---- 数据 ---- */

async function loadSourceImage() {
  const id = Number(route.params.id)
  if (!id) { ElMessage.error('缺少图片 ID'); return }
  loading.value = true
  try {
    sourceImage.value = await imageAPI.getImage(id)
    const edits = await imageAPI.listEditsBySource(id)
    editResults.value = edits || []
    for (const e of editResults.value) {
      if (e.status === 'pending' || e.status === 'processing') pollingIds.value.add(e.id)
    }
    if (pollingIds.value.size > 0) startPolling()
  } catch (e: any) {
    ElMessage.error('加载失败: ' + (e.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

async function submitEdit() {
  if (!editPrompt.value.trim() || !sourceImage.value) return
  generating.value = true
  try {
    const result = await imageAPI.editImage({
      source_image_id: sourceImage.value.id,
      prompt: editPrompt.value.trim(),
      guidance_scale: guidanceScale.value,
      seed: seed.value,
    })
    editResults.value.unshift(result)
    pollingIds.value.add(result.id)
    startPolling()
  } catch (e: any) {
    ElMessage.error('提交失败: ' + (e.message || '未知错误'))
    generating.value = false
  }
}

function startPolling() {
  if (pollingTimer.value) return
  pollingTimer.value = setInterval(async () => {
    if (pollingIds.value.size === 0) { stopPolling(); return }
    for (const pid of pollingIds.value) {
      try {
        const updated = await imageAPI.getImage(pid)
        const idx = editResults.value.findIndex(r => r.id === pid)
        if (idx >= 0) editResults.value[idx] = updated
        if (updated.status === 'completed' || updated.status === 'failed') {
          pollingIds.value.delete(pid)
          if (updated.status === 'completed') ElMessage.success('编辑完成')
          else ElMessage.error('编辑失败: ' + (updated.error_msg || ''))
        }
      } catch { /* ignore */ }
    }
    if (pollingIds.value.size === 0) { stopPolling(); generating.value = false }
  }, 3000)
}

function stopPolling() {
  if (pollingTimer.value) { clearInterval(pollingTimer.value); pollingTimer.value = null }
}

/* ---- 预览 / 操作 ---- */

function openPreview(item: ImageGeneration) {
  if (item.status !== 'completed') return
  previewImage.value = item
  previewVisible.value = true
}

async function confirmReplace() {
  if (!previewImage.value || !sourceImage.value) return
  replacing.value = true
  try {
    await imageAPI.replaceWithEdit(sourceImage.value.id, previewImage.value.id)
    ElMessage.success('替换成功')
    previewVisible.value = false
    router.back()
  } catch (e: any) {
    ElMessage.error('替换失败: ' + (e.message || ''))
  } finally {
    replacing.value = false
  }
}

async function deleteEditResult() {
  if (!previewImage.value) return
  deleting.value = true
  try {
    await imageAPI.deleteImage(previewImage.value.id)
    editResults.value = editResults.value.filter(r => r.id !== previewImage.value!.id)
    previewVisible.value = false
    ElMessage.success('已删除')
  } catch (e: any) {
    ElMessage.error('删除失败: ' + (e.message || ''))
  } finally {
    deleting.value = false
  }
}

/* ---- 生命周期 ---- */

onMounted(() => loadSourceImage())
onUnmounted(() => stopPolling())

watch(() => route.params.id, () => {
  stopPolling(); pollingIds.value.clear(); editResults.value = []; editPrompt.value = ''
  loadSourceImage()
})
</script>

<style scoped>
/* ========== 页面整体 ========== */
.image-editor-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  gap: var(--space-4);
  padding-bottom: var(--space-8);
}

/* ========== 顶栏 ========== */
.editor-topbar {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-3) var(--space-5);
  margin: var(--space-3) var(--space-4) 0;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.topbar-divider {
  width: 1px;
  height: 20px;
  background: var(--border-primary);
}

.topbar-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.02em;
  margin: 0;
}

/* ========== 主体布局 ========== */
.editor-main {
  display: grid;
  grid-template-columns: 1fr 380px;
  gap: var(--space-4);
  padding: 0 var(--space-4);
  align-items: start;
}

/* ========== 左侧预览 ========== */
.preview-card {
  padding: var(--space-4);
}

.preview-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-3);
}

.card-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.preview-canvas {
  position: relative;
  width: 100%;
  min-height: 360px;
  max-height: 65vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--border-primary);
}

.canvas-img {
  max-width: 100%;
  max-height: 65vh;
}

.canvas-img :deep(img) {
  max-height: 65vh;
  object-fit: contain;
}

.canvas-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  color: var(--text-muted);
  font-size: 0.875rem;
}

/* ========== 右侧面板 ========== */
.sidebar-card {
  padding: var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.preset-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.chip {
  padding: 5px 12px;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-secondary);
  background: var(--bg-primary);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-full);
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.chip:hover {
  color: var(--accent);
  border-color: var(--accent);
  background: var(--accent-light);
}

.chip.active {
  color: var(--text-inverse);
  background: var(--accent);
  border-color: var(--accent);
}

.adv-collapse :deep(.el-collapse-item__header) {
  font-size: 0.8125rem;
  color: var(--text-muted);
  height: 32px;
  border-bottom: none;
  background: transparent;
}

.adv-collapse :deep(.el-collapse-item__wrap) {
  border-bottom: none;
  background: transparent;
}

.adv-collapse :deep(.el-collapse-item__content) {
  padding-bottom: 0;
}

.adv-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
}

.adv-row label {
  font-size: 0.8125rem;
  color: var(--text-secondary);
  white-space: nowrap;
  min-width: 60px;
}

.adv-val {
  font-size: 0.8125rem;
  color: var(--accent);
  font-weight: 600;
  min-width: 28px;
  text-align: right;
}

.submit-btn {
  width: 100%;
  height: 44px;
  font-size: 0.9375rem;
  font-weight: 600;
}

/* ========== 结果条 ========== */
.results-strip {
  padding: 0 var(--space-4);
}

.strip-header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-bottom: var(--space-3);
}

.strip-scroll {
  display: flex;
  gap: var(--space-3);
  overflow-x: auto;
  padding-bottom: var(--space-2);
}

.result-card {
  position: relative;
  flex-shrink: 0;
  width: 200px;
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-xl);
  overflow: hidden;
  transition: all var(--transition-normal);
  box-shadow: var(--shadow-card);
}

.result-card.clickable {
  cursor: pointer;
}

.result-card.clickable:hover {
  border-color: var(--accent);
  box-shadow: var(--shadow-card-hover);
  transform: translateY(-3px);
}

.result-card.is-processing {
  border-color: var(--warning);
  border-style: dashed;
}

.result-thumb {
  width: 100%;
  height: 140px;
  background: var(--bg-primary);
  overflow: hidden;
}

.thumb-img {
  width: 100%;
  height: 100%;
}

.thumb-img :deep(img) {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.thumb-state {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: var(--text-muted);
  font-size: 0.75rem;
  position: relative;
}

.thumb-state.failed {
  color: var(--error);
}

.pulse-ring {
  position: absolute;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  border: 2px solid var(--warning);
  animation: pulse-ring-anim 1.8s ease-out infinite;
}

@keyframes pulse-ring-anim {
  0% { transform: scale(0.6); opacity: 1; }
  100% { transform: scale(1.4); opacity: 0; }
}

.result-meta {
  padding: 10px 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
}

.result-prompt {
  flex: 1;
  font-size: 0.75rem;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin: 0;
}

/* hover 覆盖 */
.result-overlay {
  position: absolute;
  inset: 0;
  background: var(--bg-overlay);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: white;
  font-size: 0.8125rem;
  font-weight: 500;
  opacity: 0;
  transition: opacity var(--transition-fast);
  pointer-events: none;
  backdrop-filter: blur(2px);
}

.result-card.clickable:hover .result-overlay {
  opacity: 1;
}

/* ========== 预览弹窗 ========== */
.dialog-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.dialog-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--text-primary);
}

.dialog-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.compare-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-4);
}

.compare-card {
  position: relative;
  background: var(--bg-primary);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 300px;
}

.compare-badge {
  position: absolute;
  top: var(--space-3);
  left: var(--space-3);
  z-index: 2;
  padding: 4px 12px;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-secondary);
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-full);
  backdrop-filter: blur(8px);
}

.compare-badge.accent {
  color: var(--accent);
  border-color: var(--accent);
  background: var(--accent-light);
}

.compare-img {
  max-height: 58vh;
}

.compare-img :deep(img) {
  max-height: 58vh;
  object-fit: contain;
}

.dialog-prompt-bar {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  background: var(--bg-primary);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
}
</style>

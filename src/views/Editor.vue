<template>
  <div class="h-[calc(100vh-56px)] flex overflow-hidden bg-gray-100">
    <!-- Left: Storyboard List -->
    <div class="w-64 flex-shrink-0 bg-white border-r border-gray-200 flex flex-col">
      <div class="p-3 border-b border-gray-100">
        <div class="flex items-center gap-2">
          <el-button text size="small" @click="$router.push(`/drama/${drama.id}`)">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <span class="text-sm font-medium text-gray-700 truncate">EP01 · {{ episode.title.split('｜')[0] }}</span>
        </div>
      </div>
      <div class="flex-1 overflow-y-auto">
        <div
          v-for="sb in storyboards"
          :key="sb.id"
          class="flex items-center gap-3 px-3 py-2 cursor-pointer border-b border-gray-50 transition-colors"
          :class="currentId === sb.id ? 'bg-blue-50 border-l-2 border-l-blue-500' : 'hover:bg-gray-50'"
          @click="selectStoryboard(sb.id)"
        >
          <div class="w-16 h-10 rounded overflow-hidden bg-gray-100 flex-shrink-0">
            <img :src="sb.composedImage" class="w-full h-full object-cover" />
          </div>
          <div class="min-w-0">
            <div class="text-sm font-medium text-gray-800">S{{ String(sb.number).padStart(2, '0') }} {{ sb.title }}</div>
            <div class="text-xs text-gray-400">{{ sb.duration }}s</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Center: Preview -->
    <div class="flex-1 flex flex-col min-w-0">
      <div class="p-3 bg-white border-b border-gray-200 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="text-sm font-medium text-gray-700">S{{ String(current.number).padStart(2, '0') }} {{ current.title }}</span>
          <el-tag size="small" effect="plain">{{ current.duration }}s</el-tag>
        </div>
        <div class="flex items-center gap-2">
          <el-tag v-if="currentScene" size="small" effect="plain" type="info">
            {{ currentScene.name }}
          </el-tag>
        </div>
      </div>
      <div class="flex-1 flex items-center justify-center p-6 bg-gray-900/5">
        <div class="relative max-w-2xl w-full aspect-video rounded-lg overflow-hidden shadow-lg bg-black">
          <!-- Video mode -->
          <video
            v-if="showVideo"
            ref="videoEl"
            :src="current.videoUrl"
            class="w-full h-full object-contain"
            controls
            autoplay
            @ended="showVideo = false"
          />
          <!-- Image + Play button mode -->
          <template v-else>
            <img :src="current.composedImage" class="w-full h-full object-contain" />
            <div
              v-if="current.videoUrl"
              class="absolute inset-0 flex items-center justify-center cursor-pointer group"
              @click="playVideo"
            >
              <div class="w-16 h-16 rounded-full bg-black/40 group-hover:bg-black/60 flex items-center justify-center transition-colors backdrop-blur-sm">
                <svg class="w-8 h-8 text-white ml-1" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M8 5v14l11-7z" />
                </svg>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- Right: Properties Panel -->
    <div class="w-96 flex-shrink-0 bg-white border-l border-gray-200 flex flex-col overflow-hidden">
      <el-tabs v-model="rightTab" class="flex-1 flex flex-col right-tabs">
        <el-tab-pane label="镜头属性" name="props" class="flex-1 overflow-y-auto">
          <div class="p-4 space-y-4">
            <div>
              <label class="text-xs font-medium text-gray-500 mb-1 block">场景</label>
              <div v-if="currentScene" class="flex items-center gap-2 bg-gray-50 rounded p-2">
                <img :src="currentScene.imageUrl" class="w-12 h-8 rounded object-cover" />
                <div>
                  <div class="text-sm text-gray-800">{{ currentScene.name }}</div>
                  <div class="text-xs text-gray-400">{{ currentScene.atmosphere }}</div>
                </div>
              </div>
            </div>

            <div>
              <label class="text-xs font-medium text-gray-500 mb-1 block">出场角色</label>
              <div class="flex flex-wrap gap-2">
                <div
                  v-for="char in currentChars"
                  :key="char.id"
                  class="flex items-center gap-2 bg-gray-50 rounded px-2 py-1"
                >
                  <img :src="char.imageUrl" class="w-6 h-6 rounded-full object-cover" />
                  <span class="text-sm text-gray-700">{{ char.name }}·{{ char.outfitName }}</span>
                </div>
              </div>
            </div>

            <div>
              <label class="text-xs font-medium text-gray-500 mb-1 block">首帧描述</label>
              <div class="bg-gray-50 rounded p-2 text-sm text-gray-600 leading-relaxed">
                {{ current.firstFrameDesc }}
              </div>
            </div>

            <div>
              <label class="text-xs font-medium text-gray-500 mb-1 block">中间动作</label>
              <div class="bg-gray-50 rounded p-2 text-sm text-gray-600 leading-relaxed">
                {{ current.middleActionDesc }}
              </div>
            </div>

            <div>
              <label class="text-xs font-medium text-gray-500 mb-1 block">尾帧描述</label>
              <div class="bg-gray-50 rounded p-2 text-sm text-gray-600 leading-relaxed">
                {{ current.lastFrameDesc }}
              </div>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="镜头画面" name="images" class="flex-1 overflow-y-auto">
          <div class="p-4 space-y-4">
            <!-- Frame type tabs -->
            <div class="frame-type-tabs">
              <div
                class="frame-type-tab"
                :class="{ active: selectedFrameType === 'first' }"
                @click="selectedFrameType = 'first'"
              >
                <span class="frame-dot first"></span>
                首帧
                <el-badge
                  v-if="current.firstFrameImages?.length"
                  :value="current.firstFrameImages.length"
                  type="primary"
                  class="ml-1"
                />
              </div>
              <div
                class="frame-type-tab"
                :class="{ active: selectedFrameType === 'last' }"
                @click="selectedFrameType = 'last'"
              >
                <span class="frame-dot last"></span>
                尾帧
                <el-badge
                  v-if="current.lastFrameImages?.length"
                  :value="current.lastFrameImages.length"
                  type="warning"
                  class="ml-1"
                />
              </div>
            </div>

            <!-- Generated frame images -->
            <div>
              <label class="text-xs font-medium text-gray-500 mb-2 block">
                {{ selectedFrameType === 'first' ? '首帧' : '尾帧' }}图片
              </label>
              <div class="grid grid-cols-2 gap-2">
                <div
                  v-for="(img, idx) in currentFrameImages"
                  :key="idx"
                  class="rounded-lg overflow-hidden border-2 cursor-pointer transition-all"
                  :class="selectedFrameIdx === idx ? 'border-blue-500 shadow-md' : 'border-gray-100 hover:border-gray-300'"
                  @click="selectedFrameIdx = idx"
                >
                  <img :src="img" class="w-full aspect-[16/9] object-cover" />
                </div>
              </div>
            </div>

            <!-- Reference images -->
            <div>
              <label class="text-xs font-medium text-gray-500 mb-2 block">参考图</label>
              <div class="grid grid-cols-3 gap-2">
                <div v-for="char in currentChars" :key="char.id" class="rounded overflow-hidden border border-gray-100">
                  <img :src="char.imageUrl" class="w-full aspect-square object-cover" />
                  <div class="text-xs text-center py-1 text-gray-500 truncate px-1">{{ char.name }}·{{ char.outfitName }}</div>
                </div>
                <div v-if="currentScene" class="rounded overflow-hidden border border-gray-100">
                  <img :src="currentScene.imageUrl" class="w-full aspect-square object-cover" />
                  <div class="text-xs text-center py-1 text-gray-500 truncate px-1">{{ currentScene.name }}</div>
                </div>
              </div>
            </div>

            <el-button type="primary" class="w-full" :loading="imageGenerating" @click="mockGenerateImage">
              <el-icon class="mr-1"><Picture /></el-icon>
              {{ imageGenerating ? '生成中...' : `生成${selectedFrameType === 'first' ? '首' : '尾'}帧图片` }}
            </el-button>
          </div>
        </el-tab-pane>

        <el-tab-pane label="视频生成" name="video" class="flex-1 overflow-y-auto">
          <div class="p-4 space-y-4">
            <!-- First/Last frame comparison -->
            <div>
              <label class="text-xs font-medium text-gray-500 mb-2 block">首尾帧对比</label>
              <div class="flex items-center gap-2">
                <div class="flex-1 rounded-lg overflow-hidden border border-gray-200 bg-gray-50">
                  <div class="flex items-center gap-1 px-2 py-1 bg-gray-100 border-b border-gray-200">
                    <span class="frame-dot first"></span>
                    <span class="text-xs text-gray-600">首帧</span>
                  </div>
                  <img
                    v-if="current.firstFrameImages?.length"
                    :src="current.firstFrameImages[0]"
                    class="w-full aspect-video object-cover"
                  />
                  <div v-else class="w-full aspect-video flex items-center justify-center text-xs text-gray-400">
                    请先生成首帧
                  </div>
                </div>
                <div class="text-gray-300 text-lg flex-shrink-0">→</div>
                <div class="flex-1 rounded-lg overflow-hidden border border-gray-200 bg-gray-50">
                  <div class="flex items-center gap-1 px-2 py-1 bg-gray-100 border-b border-gray-200">
                    <span class="frame-dot last"></span>
                    <span class="text-xs text-gray-600">尾帧</span>
                    <el-tag size="small" type="info" effect="plain" class="ml-auto" style="font-size:10px">可选</el-tag>
                  </div>
                  <img
                    v-if="current.lastFrameImages?.length"
                    :src="current.lastFrameImages[current.lastFrameImages.length - 1]"
                    class="w-full aspect-video object-cover"
                  />
                  <div v-else class="w-full aspect-video flex items-center justify-center text-xs text-gray-400">
                    请先生成尾帧
                  </div>
                </div>
              </div>
            </div>

            <!-- Video parameters -->
            <div>
              <label class="text-xs font-medium text-gray-500 mb-1 block">生成参数</label>
              <div class="space-y-2">
                <div class="flex items-center justify-between text-sm">
                  <span class="text-gray-500">模型</span>
                  <el-tag size="small" effect="plain">Seedance 1.5 Pro</el-tag>
                </div>
                <div class="flex items-center justify-between text-sm">
                  <span class="text-gray-500">分辨率</span>
                  <span class="text-gray-700">1280×720</span>
                </div>
                <div class="flex items-center justify-between text-sm">
                  <span class="text-gray-500">时长</span>
                  <span class="text-gray-700">{{ current.duration }}s</span>
                </div>
                <div class="flex items-center justify-between text-sm">
                  <span class="text-gray-500">生成模式</span>
                  <el-tag size="small" effect="plain" type="success">首尾帧</el-tag>
                </div>
              </div>
            </div>

            <!-- Video status -->
            <div v-if="current.videoUrl" class="bg-green-50 rounded-lg p-3 flex items-center gap-2">
              <el-icon size="18" class="text-green-500"><SuccessFilled /></el-icon>
              <div>
                <p class="text-sm text-green-700">视频已生成</p>
                <p class="text-xs text-green-500">时长 {{ current.duration }}s</p>
              </div>
            </div>

            <el-button type="success" class="w-full" :loading="videoGenerating" @click="mockGenerateVideo">
              <el-icon class="mr-1"><VideoCamera /></el-icon>
              {{ videoGenerating ? '生成中...' : '生成视频' }}
            </el-button>

            <div v-if="videoProgress > 0 && videoProgress < 100" class="space-y-1">
              <el-progress :percentage="videoProgress" :stroke-width="6" />
              <p class="text-xs text-gray-400 text-center">AI 模型正在生成视频...</p>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import {
  drama, episode, storyboards,
  getCharactersForStoryboard, getSceneForStoryboard,
} from '../data/mockData.js'

const currentId = ref(storyboards[0].id)
const rightTab = ref('props')
const showVideo = ref(false)
const videoEl = ref(null)
const imageGenerating = ref(false)
const videoGenerating = ref(false)
const videoProgress = ref(0)
const selectedFrameType = ref('first')
const selectedFrameIdx = ref(0)

const current = computed(() => storyboards.find(s => s.id === currentId.value))
const currentScene = computed(() => getSceneForStoryboard(current.value))
const currentChars = computed(() => getCharactersForStoryboard(current.value))
const currentFrameImages = computed(() =>
  selectedFrameType.value === 'first'
    ? (current.value.firstFrameImages || [])
    : (current.value.lastFrameImages || [])
)

function selectStoryboard(id) {
  currentId.value = id
  showVideo.value = false
  videoProgress.value = 0
  selectedFrameIdx.value = 0
}

function playVideo() {
  showVideo.value = true
}

function mockGenerateImage() {
  imageGenerating.value = true
  setTimeout(() => {
    imageGenerating.value = false
    ElMessage.success('图片生成完成（演示）')
  }, 3000)
}

function mockGenerateVideo() {
  videoGenerating.value = true
  videoProgress.value = 0
  const timer = setInterval(() => {
    videoProgress.value += Math.random() * 15 + 5
    if (videoProgress.value >= 100) {
      videoProgress.value = 100
      clearInterval(timer)
      setTimeout(() => {
        videoGenerating.value = false
        videoProgress.value = 0
        showVideo.value = true
        ElMessage.success('视频生成完成（演示）')
      }, 500)
    }
  }, 500)
}
</script>

<style scoped>
.right-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow: hidden;
}
.right-tabs :deep(.el-tab-pane) {
  height: 100%;
  overflow-y: auto;
}
.right-tabs :deep(.el-tabs__header) {
  margin-bottom: 0;
  padding: 0 12px;
}

.frame-type-tabs {
  display: flex;
  gap: 0;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  overflow: hidden;
}
.frame-type-tab {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 0;
  font-size: 13px;
  color: #606266;
  cursor: pointer;
  background: #f5f7fa;
  transition: all 0.2s;
  border-right: 1px solid #e4e7ed;
}
.frame-type-tab:last-child {
  border-right: none;
}
.frame-type-tab.active {
  background: #fff;
  color: #303133;
  font-weight: 500;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}
.frame-type-tab:hover:not(.active) {
  background: #ebeef5;
}
.frame-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}
.frame-dot.first {
  background: #409eff;
}
.frame-dot.last {
  background: #e6a23c;
}
</style>

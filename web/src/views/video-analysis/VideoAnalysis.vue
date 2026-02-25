<template>
  <div class="page-container">
    <div class="content-wrapper">
      <AppHeader :fixed="false" :show-logo="false">
        <template #left>
          <el-button text @click="$router.push('/')" class="back-btn">
            <el-icon><ArrowLeft /></el-icon>
            <span>返回</span>
          </el-button>
          <div class="page-title">
            <h1>视频分析</h1>
            <span class="subtitle">从视频自动生成剧本</span>
          </div>
        </template>
      </AppHeader>

      <div class="analysis-content">
        <!-- Input Section -->
        <div v-if="!currentTask" class="input-section">
          <el-card class="input-card">
            <template #header>
              <div class="card-header">
                <el-icon><VideoCamera /></el-icon>
                <span>选择视频来源</span>
              </div>
            </template>

            <el-tabs v-model="inputMode" class="input-tabs">
              <!-- URL Input -->
              <el-tab-pane label="粘贴链接" name="url">
                <div class="url-input-section">
                  <p class="hint">支持 B站、YouTube 等平台视频链接直接下载分析</p>
                  <el-alert
                    type="warning"
                    :closable="true"
                    show-icon
                    class="cookie-tip"
                  >
                    <template #title>
                      小红书/抖音/快手等平台需登录才能访问，建议自行下载视频后通过「上传视频」分析
                    </template>
                  </el-alert>
                  <el-input
                    v-model="videoURL"
                    placeholder="粘贴视频链接，例如 https://www.bilibili.com/video/..."
                    size="large"
                    clearable
                    @keyup.enter="startAnalyzeFromURL"
                  >
                    <template #prefix>
                      <el-icon><Link /></el-icon>
                    </template>
                  </el-input>
                  <el-button
                    type="primary"
                    size="large"
                    :loading="submitting"
                    :disabled="!videoURL.trim()"
                    @click="startAnalyzeFromURL"
                    class="submit-btn"
                  >
                    <el-icon><VideoPlay /></el-icon>
                    下载并分析
                  </el-button>
                </div>
              </el-tab-pane>

              <!-- File Upload -->
              <el-tab-pane label="上传视频" name="upload">
                <div class="upload-section">
                  <el-alert type="success" :closable="true" show-icon class="download-tip">
                    <template #title>
                      <span>小红书/抖音视频可先通过 <a href="https://www.hellotik.app/zh/rednote" target="_blank" rel="noopener">HelloTik.app</a> 免费无水印下载，再上传分析</span>
                    </template>
                  </el-alert>
                  <el-upload
                    ref="uploadRef"
                    :auto-upload="false"
                    :limit="1"
                    :on-change="handleFileChange"
                    accept="video/*"
                    drag
                    class="video-upload"
                  >
                    <el-icon class="upload-icon"><Upload /></el-icon>
                    <div class="upload-text">拖拽视频到此处，或<em>点击上传</em></div>
                    <template #tip>
                      <div class="upload-tip">支持 MP4、MOV、AVI 等格式，最大 500MB</div>
                    </template>
                  </el-upload>
                  <el-button
                    v-if="uploadFile"
                    type="primary"
                    size="large"
                    :loading="submitting"
                    @click="startUploadAnalyze"
                    class="submit-btn"
                  >
                    <el-icon><VideoPlay /></el-icon>
                    开始分析
                  </el-button>
                </div>
              </el-tab-pane>
            </el-tabs>
          </el-card>

          <!-- History -->
          <el-card v-if="tasks.length > 0" class="history-card">
            <template #header>
              <div class="card-header">
                <el-icon><Clock /></el-icon>
                <span>分析历史</span>
              </div>
            </template>
            <div class="task-list">
              <div
                v-for="task in tasks"
                :key="task.task_id"
                class="task-item"
                @click="viewTask(task)"
              >
                <div class="task-info">
                  <span class="task-title">{{ task.title || '未命名视频' }}</span>
                  <span class="task-time">{{ formatTime(task.created_at) }}</span>
                </div>
                <div class="task-status">
                  <el-tag :type="statusTagType(task.status)" size="small">
                    {{ statusText(task.status) }}
                  </el-tag>
                  <el-tag v-if="task.imported_drama_id" type="success" size="small">已导入</el-tag>
                  <el-button
                    v-if="task.status === 'failed' || task.status === 'done'"
                    size="small"
                    type="warning"
                    :icon="RefreshRight"
                    circle
                    title="重新分析"
                    @click.stop="quickRetry(task)"
                  />
                  <el-button
                    v-if="task.status === 'failed'"
                    size="small"
                    type="danger"
                    :icon="Delete"
                    circle
                    title="删除任务"
                    @click.stop="handleDelete(task)"
                  />
                </div>
              </div>
            </div>
          </el-card>
        </div>

        <!-- Processing / Result Section -->
        <div v-else class="result-section">
          <el-button :icon="ArrowLeft" link @click="currentTask = null" class="back-link">
            返回列表
          </el-button>

          <!-- Vertical Pipeline Timeline -->
          <div class="pipeline-title-bar">
            <div class="pipeline-title-left">
              <h3>{{ currentTask.title || '视频分析' }}</h3>
              <div v-if="analysisResult?.tags?.length" class="pipeline-tags">
                <el-tag v-for="tag in analysisResult.tags" :key="tag" size="small" type="info" effect="plain" round>{{ tag }}</el-tag>
              </div>
            </div>
            <div class="pipeline-title-right">
              <el-tag v-if="currentTask.status === 'processing' || currentTask.status === 'downloading'" type="warning" effect="dark" size="small">处理中</el-tag>
              <el-tag v-else-if="currentTask.status === 'done'" type="success" effect="dark" size="small">已完成</el-tag>
              <el-tag v-else-if="currentTask.status === 'failed'" type="danger" effect="dark" size="small">失败</el-tag>
              <el-button
                v-if="currentTask.status === 'failed' || currentTask.status === 'done'"
                type="warning"
                size="small"
                :loading="retrying"
                @click="handleRetry"
                :icon="RefreshRight"
              >重新分析</el-button>
              <el-button
                v-if="currentTask.status === 'failed'"
                type="danger"
                size="small"
                @click="handleDelete(currentTask)"
                :icon="Delete"
              >删除</el-button>
            </div>
          </div>

          <!-- Error -->
          <el-alert
            v-if="currentTask.status === 'failed'"
            type="error"
            :title="'分析失败'"
            :description="currentTask.error_msg"
            show-icon
            :closable="false"
            style="margin-bottom:12px"
          />

          <div class="vtimeline">
            <div
              v-for="(step, idx) in pipelineSteps"
              :key="step.key"
              class="vt-step"
              :class="[`vt-${getPipelineStatus(step)}`]"
            >
              <!-- Left: icon + vertical line -->
              <div class="vt-rail">
                <div class="vt-icon">
                  <el-icon v-if="getPipelineStatus(step) === 'done'"><CircleCheckFilled /></el-icon>
                  <el-icon v-else-if="getPipelineStatus(step) === 'active'" class="vt-pulse"><component :is="step.icon" /></el-icon>
                  <el-icon v-else-if="getPipelineStatus(step) === 'failed'"><component :is="step.icon" /></el-icon>
                  <el-icon v-else><component :is="step.icon" /></el-icon>
                </div>
                <div v-if="idx < pipelineSteps.length - 1" class="vt-line" :class="{ filled: getPipelineStatus(step) === 'done' }"></div>
              </div>

              <!-- Right: content -->
              <div class="vt-content">
                <div class="vt-header" @click.stop="toggleStepDetail(step.key)">
                  <span class="vt-label">{{ step.label }}</span>
                  <el-tag v-if="getPipelineStatus(step) === 'active'" type="warning" size="small" effect="plain">进行中</el-tag>
                  <el-tag v-if="getPipelineStatus(step) === 'done'" type="success" size="small" effect="plain">完成</el-tag>
                  <el-tag v-if="getPipelineStatus(step) === 'failed'" type="danger" size="small" effect="plain">失败</el-tag>

                  <template v-if="step.key === 'synthesize' && analysisResult">
                    <div class="vt-header-actions" @click.stop>
                      <el-popover placement="bottom" :width="280" trigger="click">
                        <template #reference>
                          <el-button size="small" :loading="resynthesizing" class="header-action-btn">
                            <el-icon><RefreshRight /></el-icon>
                            重新合成
                          </el-button>
                        </template>
                        <div class="resynth-popover">
                          <div class="resynth-option">
                            <span>包含语音对白</span>
                            <el-switch v-model="resynthIncludeAudio" :disabled="!hasAudioData" />
                          </div>
                          <p v-if="!hasAudioData" class="resynth-hint">该视频无音频数据</p>
                          <el-button type="primary" size="small" style="width:100%;margin-top:10px" :loading="resynthesizing" @click="handleResynthesize">
                            开始合成
                          </el-button>
                        </div>
                      </el-popover>
                      <el-button
                        v-if="currentTask.imported_drama_id"
                        type="success"
                        size="small"
                        class="header-action-btn"
                        @click="router.push(`/dramas/${currentTask.imported_drama_id}`)"
                      >
                        <el-icon><VideoPlay /></el-icon>
                        查看剧本
                      </el-button>
                      <!-- [暂时隐藏] 导入剧本功能待版本稳定后开发，见 docs/TODO-hidden-features.md -->
                      <!-- <el-button
                        v-else
                        type="primary"
                        size="small"
                        class="header-action-btn"
                        :loading="importing"
                        @click="handleImport"
                      >
                        <el-icon><Download /></el-icon>
                        导入剧本
                      </el-button> -->
                    </div>
                  </template>

                  <el-icon
                    v-if="hasStepData(step.key) || getPipelineStatus(step) === 'done'"
                    class="vt-fold-icon"
                    :class="{ rotated: !isStepCollapsed(step.key) }"
                  ><ArrowDown /></el-icon>
                </div>

                <!-- Inline progress for active step -->
                <div v-if="getPipelineStatus(step) === 'active'" class="vt-progress">
                  <el-progress :percentage="currentTask.progress" :stroke-width="4" :show-text="false" striped striped-flow />
                  <span class="vt-progress-text">{{ stepProgressText(step.key) }}</span>
                </div>

                <!-- Step result panel (auto-show, can fold) -->
                <div v-show="hasStepData(step.key) && !isStepCollapsed(step.key)" class="vt-panel">

                    <!-- Download -->
                    <template v-if="step.key === 'download' && sd?.download">
                      <div class="vt-kv"><span>标题</span><span>{{ sd.download.title || currentTask.title }}</span></div>
                      <div class="vt-kv"><span>时长</span><span>{{ formatDuration(sd.download.duration || currentTask.duration) }}</span></div>
                      <div v-if="currentTask.video_path && sd.download.status !== 'failed'" class="download-video-player">
                        <video
                          controls
                          preload="metadata"
                          :src="videoStaticURL"
                          class="video-player"
                        />
                      </div>
                    </template>

                    <!-- Detect -->
                    <template v-if="step.key === 'detect' && sd?.detect">
                      <div class="vt-kv"><span>镜头数</span><span>{{ sd.detect.shot_count }}</span></div>
                      <div class="vt-kv"><span>时长</span><span>{{ formatDuration(sd.detect.duration) }}</span></div>
                      <div class="vt-kv"><span>音频</span><span>{{ sd.detect.has_audio ? '有' : '无' }}</span></div>
                      <div v-if="sd.detect.shots?.length" class="detect-shots-grid">
                        <div v-for="sh in sd.detect.shots" :key="sh.index" class="detect-shot-card">
                          <el-image
                            v-if="sh.frame_url"
                            :src="sh.frame_url"
                            fit="cover"
                            lazy
                            loading="lazy"
                            class="detect-shot-img"
                            :preview-src-list="sd.detect.shots.map((s: any) => s.frame_url).filter(Boolean)"
                            :initial-index="sh.index"
                            preview-teleported
                          />
                          <div class="detect-shot-info">
                            <span class="shot-badge">#{{ sh.index + 1 }}</span>
                            <span class="shot-range">{{ formatDuration(sh.start_time) }} → {{ formatDuration(sh.end_time) }}</span>
                          </div>
                        </div>
                      </div>
                    </template>

                    <!-- Transcribe -->
                    <template v-if="step.key === 'transcribe' && sd?.transcribe">
                      <p v-if="sd.transcribe.reason" class="vt-hint">{{ sd.transcribe.reason }}</p>
                      <p v-if="sd.transcribe.error" class="text-danger">{{ sd.transcribe.error }}</p>
                      <div v-if="sd.transcribe.status === 'processing'" class="vt-inline-progress" style="margin-bottom: 8px;">
                        <el-icon class="is-loading"><Loading /></el-icon>
                        <span>{{ sd.transcribe.message || '语音识别中...' }}</span>
                      </div>
                      <div v-if="sd.transcribe.shots?.length" class="transcript-shots-list">
                        <div
                          v-for="sh in sd.transcribe.shots"
                          :key="sh.shot_index"
                          class="transcript-shot-item"
                          :class="{ 'transcript-pending': sh.status === 'pending', 'transcript-silent': sh.status === 'silent' }"
                        >
                          <div class="transcript-shot-header">
                            <span class="transcript-index">#{{ sh.shot_index + 1 }}</span>
                            <span class="transcript-time">{{ formatDuration(sh.start_time) }} → {{ formatDuration(sh.end_time) }}</span>
                            <el-tag v-if="sh.status === 'done' && sh.text" type="success" size="small" effect="plain">有语音</el-tag>
                            <el-tag v-else-if="sh.status === 'silent'" type="info" size="small" effect="plain">无语音</el-tag>
                            <el-tag v-else-if="sh.status === 'failed'" type="danger" size="small" effect="plain">失败</el-tag>
                            <span v-else-if="sh.status === 'pending'" class="transcript-waiting">
                              <el-icon class="is-loading"><Loading /></el-icon> 等待识别
                            </span>
                          </div>
                          <p v-if="sh.text" class="transcript-shot-text">{{ sh.text }}</p>
                        </div>
                      </div>
                      <!-- Fallback: legacy segments format -->
                      <div v-else-if="sd.transcribe.segments?.length" class="transcript-list">
                        <div v-for="(seg, i) in sd.transcribe.segments" :key="i" class="transcript-row">
                          <span class="transcript-index">#{{ i + 1 }}</span>
                          <span class="transcript-time">{{ formatDuration(seg.start) }} → {{ formatDuration(seg.end) }}</span>
                          <span class="transcript-text">{{ seg.text }}</span>
                        </div>
                      </div>
                      <p v-if="sd.transcribe.status === 'done' && !sd.transcribe.shots?.length && !sd.transcribe.segments?.length" class="vt-hint">无识别结果（音频可能无语音内容）</p>
                    </template>

                    <!-- Analyze -->
                    <template v-if="step.key === 'analyze' && sd?.analyze">
                      <div v-if="sd.analyze.status === 'processing'" class="vt-inline-progress" style="margin-bottom: 8px;">
                        <el-icon class="is-loading"><Loading /></el-icon>
                        <span>画面分析 {{ sd.analyze.done || 0 }}/{{ sd.analyze.total || '?' }} 帧</span>
                      </div>
                      <div v-if="sd.analyze.frames?.length" class="frame-analysis-list">
                        <div v-for="f in sd.analyze.frames" :key="f.index" class="frame-item" :class="{ 'frame-pending': !f.description }">
                          <div class="frame-thumb" v-if="f.frame_url">
                            <span class="frame-badge">#{{ f.index + 1 }}</span>
                            <el-image
                              :src="f.frame_url"
                              fit="cover"
                              class="frame-img"
                              lazy
                              loading="lazy"
                              :preview-src-list="[f.frame_url]"
                              preview-teleported
                            />
                          </div>
                          <div class="frame-badge" v-else>#{{ f.index + 1 }}</div>
                          <div v-if="f.description" class="frame-desc" :class="{ 'frame-failed': f.description?.startsWith('[分析失败]') }" v-html="renderMd(f.description)"></div>
                          <p v-else class="frame-desc frame-waiting">
                            <el-icon class="is-loading"><Loading /></el-icon>
                            等待分析...
                          </p>
                        </div>
                      </div>
                    </template>

                    <!-- Synthesize -->
                    <template v-if="step.key === 'synthesize' && analysisResult">
                      <div class="synth-result">
                        <div class="synth-title-row">
                          <h4>{{ analysisResult.title }}</h4>
                          <div class="synth-actions">
                            <el-popover placement="bottom" :width="280" trigger="click">
                              <template #reference>
                                <el-button size="small" :loading="resynthesizing">
                                  <el-icon><RefreshRight /></el-icon>
                                  重新合成
                                </el-button>
                              </template>
                              <div class="resynth-popover">
                                <div class="resynth-option">
                                  <span>包含语音对白</span>
                                  <el-switch v-model="resynthIncludeAudio" :disabled="!hasAudioData" />
                                </div>
                                <p v-if="!hasAudioData" class="resynth-hint">该视频无音频数据</p>
                                <el-button type="primary" size="small" style="width:100%;margin-top:10px" :loading="resynthesizing" @click="handleResynthesize">
                                  开始合成
                                </el-button>
                              </div>
                            </el-popover>
                            <el-button
                              v-if="currentTask.imported_drama_id"
                              type="success"
                              size="small"
                              @click="router.push(`/dramas/${currentTask.imported_drama_id}`)"
                            >
                              <el-icon><VideoPlay /></el-icon>
                              查看剧本
                            </el-button>
                            <!-- [暂时隐藏] 导入剧本功能待版本稳定后开发，见 docs/TODO-hidden-features.md -->
                            <!-- <el-button
                              :type="currentTask.imported_drama_id ? 'default' : 'primary'"
                              size="small"
                              :loading="importing"
                              @click="handleImport"
                            >
                              <el-icon><DocumentAdd /></el-icon>
                              {{ currentTask.imported_drama_id ? '重新导入' : '导入为剧本' }}
                            </el-button> -->
                            <el-button
                              size="small"
                              @click="handleExportAnalysis"
                            >
                              <el-icon><Download /></el-icon>
                              导出 ZIP
                            </el-button>
                          </div>
                        </div>
                        <p class="synth-summary">{{ analysisResult.summary }}</p>
                        <div v-if="analysisResult.tags?.length" class="synth-tags">
                          <el-tag v-for="tag in analysisResult.tags" :key="tag" size="small" round effect="plain">{{ tag }}</el-tag>
                        </div>

                        <div class="synth-section">
                          <h5>角色 ({{ analysisResult.characters.length }})</h5>
                          <div class="character-card-list">
                            <div v-for="(ch, ci) in analysisResult.characters" :key="ci" class="character-card-item">
                              <div class="char-card-header">
                                <span class="char-card-name">{{ ch.name }}</span>
                                <el-tag :type="ch.role === '主角' ? 'danger' : 'info'" size="small">{{ ch.role }}</el-tag>
                              </div>
                              <p class="char-card-desc">{{ ch.description }}</p>
                            </div>
                          </div>
                        </div>

                        <div class="synth-section">
                          <h5>分镜 ({{ analysisResult.shots.length }})</h5>
                          <div class="shots-list">
                            <div v-for="shot in analysisResult.shots" :key="shot.index" class="shot-item">
                              <div class="shot-header">
                                <span class="shot-index">#{{ shot.index + 1 }}</span>
                                <span v-if="shot.title" class="shot-title-text">{{ shot.title }}</span>
                                <span class="shot-time">{{ formatDuration(shot.start_time) }} - {{ formatDuration(shot.end_time) }}</span>
                                <el-tag v-if="shot.shot_type" size="small" type="success">{{ shot.shot_type }}</el-tag>
                                <el-tag v-if="shot.location" size="small" type="info">{{ shot.location }}</el-tag>
                                <el-tag v-if="shot.mood" size="small" type="warning">{{ shot.mood }}</el-tag>
                              </div>
                              <div class="shot-body">
                                <div v-if="shotFrameURL(shot.index)" class="shot-frame-thumb">
                                  <el-image
                                    :src="shotFrameURL(shot.index)"
                                    fit="cover"
                                    lazy
                                    loading="lazy"
                                    class="shot-frame-img"
                                    :preview-src-list="allFrameURLs"
                                    :initial-index="shot.index"
                                    preview-teleported
                                  />
                                </div>
                                <div class="shot-text-content">
                                  <p v-if="shot.first_frame_desc" class="shot-frame-desc">
                                    <strong>首帧：</strong>{{ shot.first_frame_desc }}
                                  </p>
                                  <p v-if="shot.middle_action_desc" class="shot-frame-desc shot-middle">
                                    <strong>过程：</strong>{{ shot.middle_action_desc }}
                                  </p>
                                  <p v-if="shot.last_frame_desc" class="shot-frame-desc">
                                    <strong>尾帧：</strong>{{ shot.last_frame_desc }}
                                  </p>
                                  <p v-if="!shot.first_frame_desc && shot.description" class="shot-desc">{{ shot.description }}</p>
                                  <p v-if="shot.dialogue" class="shot-dialogue">
                                    <el-icon><ChatDotRound /></el-icon>
                                    {{ shot.dialogue }}
                                  </p>
                                  <div v-if="shot.characters?.length" class="shot-characters">
                                    <el-tag v-for="name in shot.characters" :key="name" size="small">{{ name }}</el-tag>
                                  </div>
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </template>

                  </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft, ArrowDown, VideoCamera, Link, VideoPlay, Upload, Clock,
  DocumentAdd, ChatDotRound, RefreshRight, Download, Scissor,
  Microphone, Picture, EditPen, CircleCheckFilled, Delete, Loading
} from '@element-plus/icons-vue'
import { marked } from 'marked'
import { videoAnalysisAPI } from '@/api/videoAnalysis'
import type { VideoAnalysisTask, AnalysisResult, StageData } from '@/api/videoAnalysis'
import { AppHeader } from '@/components/common'

marked.setOptions({ breaks: true, gfm: true })
const renderMd = (text: string): string => {
  if (!text) return ''
  return marked.parse(text) as string
}

const router = useRouter()

const inputMode = ref('url')
const videoURL = ref('')
const uploadFile = ref<File | null>(null)
const submitting = ref(false)
const importing = ref(false)
const retrying = ref(false)
const resynthesizing = ref(false)
const resynthIncludeAudio = ref(true)

const tasks = ref<VideoAnalysisTask[]>([])
const currentTask = ref<VideoAnalysisTask | null>(null)

let pollTimer: ReturnType<typeof setInterval> | null = null

const pipelineSteps = [
  { key: 'download',    label: '下载视频',   icon: Download,    stages: ['downloading', 'downloaded'] },
  { key: 'detect',      label: '场景切分',   icon: Scissor,     stages: ['detecting'] },
  { key: 'transcribe',  label: '语音提取',   icon: Microphone,  stages: ['transcribing', 'analyzing'] },
  { key: 'analyze',     label: '画面分析',   icon: Picture,     stages: ['analyzing'] },
  { key: 'synthesize',  label: '剧本合成',   icon: EditPen,     stages: ['synthesizing'] },
]

const stageOrder = ['downloading', 'downloaded', 'detecting', 'transcribing', 'analyzing', 'synthesizing', 'complete']

const getPipelineStatus = (step: typeof pipelineSteps[0]) => {
  const task = currentTask.value
  if (!task) return 'pending'

  // Check stageData for per-step status (handles parallel steps)
  const stepSD = sd.value ? (sd.value as Record<string, any>)[step.key] : null
  if (stepSD) {
    if (stepSD.status === 'done' || stepSD.status === 'skipped') return 'done'
    if (stepSD.status === 'failed') return 'failed'
    if (stepSD.status === 'processing') return 'active'
  }

  if (task.status === 'failed') {
    const currentIdx = stageOrder.indexOf(task.stage || '')
    const stepMaxIdx = Math.max(...step.stages.map(s => stageOrder.indexOf(s)))
    const stepMinIdx = Math.min(...step.stages.map(s => stageOrder.indexOf(s)))
    if (currentIdx > stepMaxIdx) return 'done'
    if (currentIdx >= stepMinIdx && currentIdx <= stepMaxIdx) return 'failed'
    return 'pending'
  }
  if (task.stage === 'complete') return 'done'
  const currentIdx = stageOrder.indexOf(task.stage || '')
  const stepMaxIdx = Math.max(...step.stages.map(s => stageOrder.indexOf(s)))
  const stepMinIdx = Math.min(...step.stages.map(s => stageOrder.indexOf(s)))
  if (currentIdx > stepMaxIdx) return 'done'
  if (currentIdx >= stepMinIdx && currentIdx <= stepMaxIdx) return 'active'
  return 'pending'
}

const collapsedSteps = reactive<Record<string, boolean>>({})

const isStepCollapsed = (key: string): boolean => {
  return !!collapsedSteps[key]
}

const toggleStepDetail = (key: string) => {
  if (!hasStepData(key)) return
  collapsedSteps[key] = !collapsedSteps[key]
}

const stepProgressText = (key: string): string => {
  const data = sd.value
  if (!data) return stageText(currentTask.value?.stage)

  if (key === 'analyze' && data.analyze) {
    const a = data.analyze as any
    if (a.done !== undefined && a.total !== undefined) {
      return `画面分析 ${a.done}/${a.total} 帧`
    }
  }
  if (key === 'transcribe' && data.transcribe) {
    const t = data.transcribe as any
    if (t.done !== undefined && t.total !== undefined) {
      return `语音识别 ${t.done}/${t.total} 镜头`
    }
    if (t.message) return t.message
  }
  return stageText(currentTask.value?.stage)
}

const sd = computed<StageData | null>(() => {
  const raw = currentTask.value?.stage_data
  if (!raw) return null
  if (typeof raw === 'string') {
    try { return JSON.parse(raw) } catch { return null }
  }
  return raw as StageData
})

const hasStepData = (key: string) => {
  if (!sd.value) return false
  return !!(sd.value as Record<string, unknown>)[key]
}

const analysisResult = computed<AnalysisResult | null>(() => {
  if (!currentTask.value?.result) return null
  if (typeof currentTask.value.result === 'string') {
    try { return JSON.parse(currentTask.value.result) } catch { return null }
  }
  return currentTask.value.result as unknown as AnalysisResult
})

const shotFrameURL = (shotIndex: number): string => {
  const detectShots = sd.value?.detect?.shots
  if (!detectShots?.length) return ''
  const match = detectShots.find((s: any) => s.index === shotIndex)
  return match?.frame_url || ''
}

const allFrameURLs = computed<string[]>(() => {
  const detectShots = sd.value?.detect?.shots
  if (!detectShots?.length) return []
  return detectShots.map((s: any) => s.frame_url).filter(Boolean)
})

const loadTasks = async () => {
  try {
    const res = await videoAnalysisAPI.list()
    tasks.value = res.items || []
  } catch {
    // silently fail
  }
}

const handleFileChange = (file: any) => {
  uploadFile.value = file.raw
}

const startAnalyzeFromURL = async () => {
  if (!videoURL.value.trim()) return
  submitting.value = true
  try {
    const res = await videoAnalysisAPI.analyzeFromURL(videoURL.value.trim())
    ElMessage.success(res.message || '开始分析')
    startPolling(res.task_id)
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error?.message || err?.message || '提交失败')
  } finally {
    submitting.value = false
  }
}

const startUploadAnalyze = async () => {
  if (!uploadFile.value) return
  submitting.value = true
  try {
    const res = await videoAnalysisAPI.upload(uploadFile.value)
    ElMessage.success(res.message || '开始分析')
    startPolling(res.task_id)
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error?.message || err?.message || '上传失败')
  } finally {
    submitting.value = false
  }
}

const startPolling = (taskId: string) => {
  stopPolling()
  currentTask.value = {
    task_id: taskId, status: 'pending', progress: 0, stage: 'starting'
  } as VideoAnalysisTask

  pollTimer = setInterval(async () => {
    try {
      const task = await videoAnalysisAPI.getStatus(taskId)
      currentTask.value = task
      if (task.status === 'done' || task.status === 'failed') {
        stopPolling()
        loadTasks()
      }
    } catch {
      // continue polling
    }
  }, 2000)
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

const viewTask = (task: VideoAnalysisTask) => {
  currentTask.value = task
  if (task.status === 'pending' || task.status === 'downloading' || task.status === 'processing') {
    startPolling(task.task_id)
  }
}

const quickRetry = async (task: VideoAnalysisTask) => {
  try {
    const res = await videoAnalysisAPI.retry(task.task_id)
    ElMessage.success(res.message || '已重新开始分析')
    currentTask.value = { ...task, status: 'processing', progress: 10, stage: 'retrying' }
    startPolling(res.task_id)
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error?.message || err?.message || '重试失败')
  }
}

const handleRetry = async () => {
  if (!currentTask.value) return
  retrying.value = true
  try {
    const res = await videoAnalysisAPI.retry(currentTask.value.task_id)
    ElMessage.success(res.message || '已重新开始分析')
    startPolling(res.task_id)
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error?.message || err?.message || '重试失败')
  } finally {
    retrying.value = false
  }
}

const handleImport = async () => {
  if (!currentTask.value) return
  importing.value = true
  try {
    const res = await videoAnalysisAPI.importToDrama(currentTask.value.task_id)
    ElMessage.success(`导入成功！剧本「${res.title}」已创建`)
    currentTask.value.imported_drama_id = res.drama_id
    loadTasks()
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error?.message || err?.message || '导入失败')
  } finally {
    importing.value = false
  }
}

const handleExportAnalysis = () => {
  if (!currentTask.value) return
  videoAnalysisAPI.exportAnalysis(currentTask.value.task_id)
}

const hasAudioData = computed(() => {
  if (!sd.value?.transcribe) return false
  const t = sd.value.transcribe
  if (t.status === 'skipped') return false
  return !!(t.shots?.length || t.segments?.length)
})

const handleResynthesize = async () => {
  if (!currentTask.value) return
  resynthesizing.value = true
  try {
    const res = await videoAnalysisAPI.resynthesize(currentTask.value.task_id, resynthIncludeAudio.value)
    ElMessage.success(res.message || '正在重新合成剧本')
    startPolling(res.task_id)
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error?.message || err?.message || '重新合成失败')
  } finally {
    resynthesizing.value = false
  }
}

const videoStaticURL = computed(() => {
  const vp = currentTask.value?.video_path
  if (!vp) return ''
  const prefix = 'data/storage/'
  const idx = vp.indexOf(prefix)
  if (idx >= 0) return '/static/' + vp.substring(idx + prefix.length)
  return '/static/' + vp
})

const handleDelete = async (task: VideoAnalysisTask) => {
  try {
    await ElMessageBox.confirm('确定删除该任务？相关文件也会被清理。', '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  try {
    await videoAnalysisAPI.deleteTask(task.task_id)
    ElMessage.success('已删除')
    if (currentTask.value?.task_id === task.task_id) {
      currentTask.value = null
      stopPolling()
    }
    loadTasks()
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error?.message || err?.message || '删除失败')
  }
}

const statusTagType = (status: string) => {
  const map: Record<string, string> = {
    done: 'success', failed: 'danger', processing: 'warning',
    downloading: 'warning', pending: 'info'
  }
  return map[status] || 'info'
}

const statusText = (status: string) => {
  const map: Record<string, string> = {
    done: '完成', failed: '失败', processing: '分析中',
    downloading: '下载中', pending: '等待中'
  }
  return map[status] || status
}

const stageText = (stage: string) => {
  const map: Record<string, string> = {
    created: '准备中...', downloading: '正在下载视频...',
    downloaded: '下载完成，准备分析...', detecting: '正在检测镜头...',
    analyzing: '正在分析画面...', synthesizing: '正在生成剧本...',
    retrying: '正在重新分析...', complete: '分析完成'
  }
  return map[stage] || stage || '处理中...'
}

const formatTime = (dateStr: string) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

const formatDuration = (seconds: number) => {
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}

onMounted(() => {
  loadTasks()
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.content-wrapper {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 var(--space-6);
}

.page-title {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.page-title h1 {
  font-size: 20px;
  font-weight: 600;
  margin: 0;
}

.subtitle {
  color: var(--text-secondary);
  font-size: 13px;
}

.back-btn {
  font-size: 14px;
}

.analysis-content {
  padding: var(--space-6) 0;
}

/* Input Section */
.input-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.input-card {
  border-radius: var(--radius-xl);
}

.card-header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-weight: 600;
  font-size: 15px;
}

.input-tabs :deep(.el-tabs__header) {
  margin-bottom: var(--space-5);
}

.url-input-section,
.upload-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.hint {
  color: var(--text-secondary);
  font-size: 13px;
  margin: 0;
}

.cookie-tip {
  margin-bottom: var(--space-1);
}

.submit-btn {
  align-self: flex-end;
  min-width: 160px;
}

.video-upload {
  width: 100%;
}

.video-upload :deep(.el-upload-dragger) {
  padding: 40px 20px;
  border-radius: var(--radius-xl);
  border: 2px dashed var(--glass-border);
  background: var(--glass-bg);
  transition: all var(--transition-fast);
}

.video-upload :deep(.el-upload-dragger:hover) {
  border-color: var(--accent);
}

.upload-icon {
  font-size: 40px;
  color: var(--text-muted);
  margin-bottom: var(--space-3);
}

.upload-text {
  color: var(--text-primary);
}

.upload-tip {
  color: var(--text-secondary);
  font-size: 12px;
  margin-top: var(--space-2);
}

/* History */
.history-card {
  border-radius: var(--radius-xl);
}

.task-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.task-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.task-item:hover {
  background: var(--bg-card-hover);
}

.task-info {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.task-title {
  font-weight: 500;
  font-size: 14px;
}

.task-time {
  color: var(--text-secondary);
  font-size: 12px;
}

.task-status {
  display: flex;
  gap: var(--space-2);
}

/* Result Section */
.result-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.back-link {
  align-self: flex-start;
  margin-bottom: 8px;
}

/* ===== Vertical Timeline Pipeline ===== */
.pipeline-title-bar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
  gap: 16px;
}

.pipeline-title-bar h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.pipeline-title-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.vtimeline {
  position: relative;
  padding-left: 0;
}

.vt-step {
  display: flex;
  gap: 16px;
  min-height: 60px;
}

.vt-rail {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 36px;
  flex-shrink: 0;
}

.vt-icon {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  flex-shrink: 0;
  transition: all 0.3s ease;
}

.vt-pending .vt-icon {
  background: var(--bg-card-hover);
  color: var(--text-muted);
  border: 2px solid var(--border-primary);
}

.vt-active .vt-icon {
  background: var(--accent-light);
  color: var(--accent);
  border: 2px solid var(--accent);
}

.vt-done .vt-icon {
  background: var(--success-light);
  color: var(--success);
  border: 2px solid var(--success);
}

.vt-failed .vt-icon {
  background: var(--error-light);
  color: var(--error);
  border: 2px solid var(--error);
}

@keyframes vt-pulse {
  0% { box-shadow: 0 0 0 0 rgba(14, 165, 233, 0.4); }
  70% { box-shadow: 0 0 0 6px transparent; }
  100% { box-shadow: 0 0 0 0 transparent; }
}

.vt-pulse {
  animation: vt-pulse 1.5s ease infinite;
}

.vt-active .vt-icon {
  animation: vt-pulse 1.5s ease infinite;
}

.vt-line {
  width: 2px;
  flex: 1;
  min-height: 16px;
  background: var(--border-primary);
  transition: background var(--transition-slow);
}

.vt-line.filled {
  background: var(--success);
}

.vt-content {
  flex: 1;
  padding-bottom: 20px;
  min-width: 0;
}

.vt-header {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  cursor: pointer;
  user-select: none;
  flex-wrap: wrap;
}

.vt-header-actions {
  display: flex;
  gap: 6px;
  margin-left: auto;
  flex-shrink: 0;
}

.header-action-btn {
  font-size: 12px;
}

.vt-label {
  font-size: 15px;
  font-weight: 600;
}

.vt-pending .vt-label {
  color: var(--text-muted);
}

.vt-active .vt-label {
  color: var(--accent);
}

.vt-done .vt-label {
  color: var(--text-primary);
}

.vt-failed .vt-label {
  color: var(--error);
}

.vt-fold-icon {
  font-size: 12px;
  color: var(--text-secondary);
  transition: transform var(--transition-slow);
  margin-left: auto;
}

.vt-fold-icon.rotated {
  transform: rotate(180deg);
}

.vt-progress {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 6px;
}

.vt-progress .el-progress {
  flex: 1;
}

.vt-progress-text {
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
}

/* Step result panel - glass morphism */
.vt-panel {
  margin-top: 10px;
  padding: var(--space-4);
  background: var(--glass-bg);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-radius: var(--radius-lg);
  border: 1px solid var(--glass-border);
  border-left: 3px solid var(--accent);
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-primary);
}

.vt-inline-progress {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--accent);
  font-size: 13px;
  padding: 4px 0;
}
.frame-pending {
  opacity: 0.5;
}
.frame-waiting {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--text-muted);
  font-style: italic;
}

.vt-kv {
  display: flex;
  gap: var(--space-3);
  padding: 2px 0;
}

.vt-kv > span:first-child {
  color: var(--text-secondary);
  min-width: 56px;
  flex-shrink: 0;
}

.vt-hint {
  color: var(--text-secondary);
  margin: var(--space-1) 0;
}

/* Synthesize result section */
.synth-result {
  margin-top: 4px;
}

.synth-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.synth-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.synth-title-row h4 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.synth-summary {
  color: var(--text-secondary);
  font-size: 13px;
  margin: 6px 0 var(--space-4);
}

.synth-section {
  margin-top: 16px;
}

.synth-section h5 {
  font-size: 14px;
  font-weight: 600;
  margin: 0 0 8px;
}

.character-card-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 10px;
}

.character-card-item {
  background: var(--el-fill-color-lighter, #f5f7fa);
  border-radius: 8px;
  padding: 10px 12px;
  border: 1px solid var(--el-border-color-lighter, #e4e7ed);
}

.char-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.char-card-name {
  font-weight: 600;
  font-size: 14px;
}

.char-card-desc {
  font-size: 13px;
  color: var(--el-text-color-regular);
  line-height: 1.6;
  margin: 0;
  white-space: pre-line;
  word-break: break-word;
}

/* Shots */
.shots-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.shot-item {
  padding: var(--space-4);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-bg);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  transition: all var(--transition-fast);
}

.shot-item:hover {
  border-color: var(--accent);
  box-shadow: var(--shadow-sm);
}

.shot-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.shot-index {
  font-weight: 700;
  color: var(--accent);
  font-size: 14px;
  min-width: 32px;
}

.shot-time {
  color: var(--text-secondary);
  font-size: 12px;
  font-family: monospace;
}

.shot-body {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.shot-frame-thumb {
  flex-shrink: 0;
  width: 120px;
  border-radius: var(--radius-md);
  overflow: hidden;
}

.shot-frame-img {
  width: 120px;
  height: 80px;
  display: block;
}

.shot-text-content {
  flex: 1;
  min-width: 0;
}

.shot-title-text {
  font-weight: 600;
  font-size: 13px;
  color: var(--text-primary);
}

.shot-desc {
  margin: 0 0 8px;
  font-size: 14px;
  line-height: 1.6;
}

.shot-frame-desc {
  margin: 0 0 6px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary);
}
.shot-frame-desc strong {
  color: var(--text-primary);
  font-size: 12px;
}
.shot-middle {
  color: var(--text-primary);
  font-size: 14px;
}

.resynth-popover {
  padding: 4px 0;
}
.resynth-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 0;
}
.resynth-hint {
  color: var(--text-muted);
  font-size: 12px;
  margin: 4px 0 0;
}

.shot-dialogue {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  color: var(--warning);
  font-size: 13px;
  margin: var(--space-2) 0;
  padding: var(--space-2) var(--space-3);
  background: var(--warning-light);
  border-radius: var(--radius-md);
}

.shot-characters {
  display: flex;
  gap: 6px;
  margin-top: 8px;
}



/* Detect shots grid with thumbnails */
.detect-shots-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 8px;
  margin-top: 10px;
}

.detect-shot-card {
  border-radius: var(--radius-lg);
  overflow: hidden;
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  transition: all var(--transition-fast);
}

.detect-shot-card:hover {
  border-color: var(--accent);
  box-shadow: var(--shadow-sm);
}

.detect-shot-img {
  width: 100%;
  height: 80px;
  display: block;
  cursor: pointer;
}

.detect-shot-info {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  font-size: 11px;
}

.shot-badge {
  font-weight: 700;
  color: var(--accent);
  font-size: 12px;
}

.shot-range {
  font-family: monospace;
  color: var(--text-secondary);
  font-size: 11px;
}

.transcript-shots-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 500px;
  overflow-y: auto;
}

.transcript-shot-item {
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-lg);
  background: var(--bg-secondary);
  transition: all var(--transition-fast);
}

.transcript-shot-item:hover {
  background: var(--bg-card-hover);
}

.transcript-pending {
  opacity: 0.5;
}

.transcript-silent {
  opacity: 0.6;
}

.transcript-shot-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.transcript-index {
  font-size: 12px;
  color: var(--accent);
  font-weight: 600;
  min-width: 24px;
  flex-shrink: 0;
}

.transcript-time {
  font-family: monospace;
  color: var(--text-secondary);
  font-size: 12px;
  white-space: nowrap;
}

.transcript-waiting {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--text-muted);
  font-size: 12px;
}

.transcript-shot-text {
  margin: 6px 0 0 32px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-primary);
}

/* Legacy flat segments list */
.transcript-list {
  margin-top: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.transcript-row {
  display: flex;
  align-items: baseline;
  gap: 10px;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-md);
  transition: background var(--transition-fast);
}

.transcript-row:hover {
  background: var(--bg-card-hover);
}

.transcript-row:not(:last-child) {
  border-bottom: 1px solid var(--border-primary);
}

.transcript-text {
  font-size: 13px;
  line-height: 1.6;
  flex: 1;
}

.frame-analysis-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 600px;
  overflow-y: auto;
}

.frame-item {
  display: flex;
  gap: 14px;
  padding: var(--space-3);
  background: var(--glass-bg);
  border-radius: var(--radius-lg);
  border: 1px solid var(--glass-border);
  transition: all var(--transition-fast);
}

.frame-item:hover {
  border-color: var(--accent);
}

.frame-thumb {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.frame-img {
  width: 160px;
  height: 90px;
  border-radius: var(--radius-md);
  cursor: pointer;
  border: 1px solid var(--glass-border);
}

.frame-badge {
  font-weight: 700;
  color: var(--accent);
  min-width: 32px;
  flex-shrink: 0;
  font-size: 12px;
  text-align: center;
}

.frame-desc {
  margin: 0;
  font-size: 13px;
  line-height: 1.7;
  flex: 1;
  min-width: 0;
}

.frame-desc :deep(h1),
.frame-desc :deep(h2),
.frame-desc :deep(h3),
.frame-desc :deep(h4),
.frame-desc :deep(h5) {
  font-size: 13px;
  font-weight: 600;
  margin: 6px 0 2px;
}

.frame-desc :deep(p) {
  margin: 2px 0;
}

.frame-desc :deep(ul),
.frame-desc :deep(ol) {
  margin: 2px 0;
  padding-left: 18px;
}

.frame-desc :deep(li) {
  margin: 1px 0;
}

.frame-desc :deep(strong) {
  color: var(--text-primary);
}

.frame-failed {
  color: var(--error);
}

.text-danger {
  color: var(--error);
}

.download-tip {
  margin-bottom: var(--space-4);
}

.download-tip a {
  color: var(--accent);
  font-weight: 600;
  text-decoration: none;
}

.download-tip a:hover {
  text-decoration: underline;
}

/* Pipeline title tags */
.pipeline-title-left {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.pipeline-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.synth-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 12px;
}

/* Download video player */
.download-video-player {
  margin-top: 12px;
}

.video-player {
  width: 100%;
  max-width: 600px;
  border-radius: var(--radius-lg);
  background: #000;
}

</style>

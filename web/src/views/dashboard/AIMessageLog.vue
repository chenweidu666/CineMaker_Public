<template>
  <div class="ai-log-container">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <div class="header-left">
            <el-button text @click="$router.back()">
              <el-icon><ArrowLeft /></el-icon>
            </el-button>
            <h2>AI 日志</h2>
          </div>
          <el-button v-if="activeTab === 'logs'" type="primary" @click="loadData">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </el-header>

      <el-main>
        <!-- 主 Tab：AI 日志 | AI 配置 -->
        <el-tabs v-model="activeTab" class="main-tabs">
          <el-tab-pane label="AI 日志" name="logs">
        <!-- 统计卡片 -->
        <el-row :gutter="16" class="stats-row">
          <el-col :span="6">
            <el-card shadow="hover" class="stat-card">
              <div class="stat-item">
                <el-icon :size="36" color="#409eff"><ChatDotRound /></el-icon>
                <div class="stat-value">{{ stats.total || 0 }}</div>
                <div class="stat-label">总调用次数</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover" class="stat-card">
              <div class="stat-item">
                <el-icon :size="36" color="#67c23a"><SuccessFilled /></el-icon>
                <div class="stat-value">{{ stats.success || 0 }}</div>
                <div class="stat-label">成功</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover" class="stat-card">
              <div class="stat-item">
                <el-icon :size="36" color="#f56c6c"><CircleCloseFilled /></el-icon>
                <div class="stat-value">{{ stats.failed || 0 }}</div>
                <div class="stat-label">失败</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover" class="stat-card">
              <div class="stat-item">
                <el-icon :size="36" color="#e6a23c"><Timer /></el-icon>
                <div class="stat-value">{{ stats.today || 0 }}</div>
                <div class="stat-label">今日调用</div>
              </div>
            </el-card>
          </el-col>
        </el-row>

        <!-- 筛选条件 -->
        <el-card shadow="never" class="filter-card">
          <el-form :inline="true" :model="filters" class="filter-form">
            <el-form-item label="类型">
              <el-select v-model="filters.service_type" placeholder="全部" clearable style="width: 120px">
                <el-option label="文本" value="text" />
                <el-option label="图片" value="image" />
                <el-option label="视频" value="video" />
              </el-select>
            </el-form-item>
            <el-form-item label="用途">
              <el-select v-model="filters.purpose" placeholder="全部" clearable filterable style="width: 200px">
                <el-option v-for="p in purposeOptions" :key="p" :label="purposeLabel(p)" :value="p" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px">
                <el-option label="等待中" value="pending" />
                <el-option label="成功" value="success" />
                <el-option label="失败" value="failed" />
              </el-select>
            </el-form-item>
            <el-form-item label="关键词">
              <el-input v-model="filters.keyword" placeholder="搜索提示词/响应" clearable style="width: 200px" @keyup.enter="handleSearch" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleSearch">搜索</el-button>
              <el-button @click="resetFilters">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 日志列表 -->
        <el-card shadow="never" class="table-card">
          <el-table :data="logs" v-loading="loading" stripe style="width: 100%" @row-click="showDetail" row-class-name="clickable-row">
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="service_type" label="类型" width="80">
              <template #default="{ row }">
                <el-tag :type="serviceTypeTag(row.service_type)" size="small">
                  {{ serviceTypeLabel(row.service_type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="purpose" label="用途" width="160">
              <template #default="{ row }">
                <span class="purpose-text">{{ purposeLabel(row.purpose) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="provider" label="供应商" width="100" />
            <el-table-column prop="model" label="模型" width="160" show-overflow-tooltip />
            <el-table-column label="提示词" min-width="300">
              <template #default="{ row }">
                <div class="prompt-preview">{{ truncate(row.user_prompt, 120) }}</div>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="statusTag(row.status)" size="small">
                  {{ statusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="duration_ms" label="耗时" width="90">
              <template #default="{ row }">
                {{ row.duration_ms > 0 ? formatDuration(row.duration_ms) : '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="时间" width="170">
              <template #default="{ row }">
                {{ formatTime(row.created_at) }}
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-pagination
              v-model:current-page="pagination.page"
              v-model:page-size="pagination.page_size"
              :total="pagination.total"
              :page-sizes="[20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="loadLogs"
              @current-change="loadLogs"
            />
          </div>
        </el-card>
          </el-tab-pane>

          <el-tab-pane label="配置 API" name="config">
            <AIConfig embedded />
          </el-tab-pane>
        </el-tabs>
      </el-main>
    </el-container>

    <!-- 详情弹窗 -->
    <el-drawer v-model="detailVisible" title="消息详情" size="60%" direction="rtl">
      <div v-if="currentLog" class="detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ currentLog.id }}</el-descriptions-item>
          <el-descriptions-item label="请求ID">
            <span class="mono-text">{{ currentLog.request_id }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="服务类型">
            <el-tag :type="serviceTypeTag(currentLog.service_type)" size="small">
              {{ serviceTypeLabel(currentLog.service_type) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="用途">{{ purposeLabel(currentLog.purpose) }}</el-descriptions-item>
          <el-descriptions-item label="供应商">{{ currentLog.provider || '-' }}</el-descriptions-item>
          <el-descriptions-item label="模型">{{ currentLog.model || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTag(currentLog.status)" size="small">
              {{ statusLabel(currentLog.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="耗时">
            {{ currentLog.duration_ms > 0 ? formatDuration(currentLog.duration_ms) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="时间" :span="2">{{ formatTime(currentLog.created_at) }}</el-descriptions-item>
        </el-descriptions>

        <div v-if="currentLog.system_prompt" class="prompt-section">
          <h4>系统提示词 (System Prompt)</h4>
          <el-input
            type="textarea"
            :model-value="currentLog.system_prompt"
            readonly
            :autosize="{ minRows: 3, maxRows: 15 }"
          />
        </div>

        <div class="prompt-section">
          <h4>用户提示词 (User Prompt)</h4>
          <el-input
            type="textarea"
            :model-value="currentLog.user_prompt"
            readonly
            :autosize="{ minRows: 3, maxRows: 20 }"
          />
        </div>

        <div v-if="currentLog.full_request" class="prompt-section">
          <h4>完整请求参数</h4>
          <el-input
            type="textarea"
            :model-value="formatJSON(currentLog.full_request)"
            readonly
            :autosize="{ minRows: 2, maxRows: 10 }"
          />
        </div>

        <div v-if="currentLog.response" class="prompt-section">
          <h4>AI 响应</h4>
          <el-input
            type="textarea"
            :model-value="currentLog.response"
            readonly
            :autosize="{ minRows: 3, maxRows: 20 }"
          />
        </div>

        <div v-if="currentLog.error_message" class="prompt-section error-section">
          <h4>错误信息</h4>
          <el-alert :title="currentLog.error_message" type="error" :closable="false" show-icon />
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft, Refresh, ChatDotRound, SuccessFilled, CircleCloseFilled, Timer } from '@element-plus/icons-vue'
import { aiLogAPI, type AIMessageLog, type AILogStats } from '../../api/ai-log'
import AIConfig from '@/views/settings/AIConfig.vue'

const route = useRoute()
const activeTab = ref<'logs' | 'config'>(
  (route.query.tab as string) === 'config' ? 'config' : 'logs'
)

const loading = ref(false)
const logs = ref<AIMessageLog[]>([])
const stats = ref<AILogStats>({ total: 0, success: 0, failed: 0, today: 0, by_type: [], by_purpose_top: [] })
const detailVisible = ref(false)
const currentLog = ref<AIMessageLog | null>(null)

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const filters = reactive({
  service_type: '',
  purpose: '',
  status: '',
  keyword: ''
})

const purposeOptions = ref<string[]>([])

const purposeMap: Record<string, string> = {
  general: '通用文本',
  translate_prompt: '提示词翻译',
  generate_image: '图片生成',
  generate_video: '视频生成',
  generate_storyboard: '分镜生成',
  generate_script: '剧本生成',
  generate_characters: '角色生成',
  generate_first_frame: '首帧提示词',
  generate_key_frame: '关键帧提示词',
  generate_last_frame: '尾帧提示词',
  generate_action_sequence: '动作序列提示词',
  generate_video_prompt: '视频提示词',
  extract_backgrounds_from_script: '场景提取',
  extract_characters: '角色提取',
  extract_props: '道具提取',
  polish_prompt: '提示词优化'
}

const purposeLabel = (purpose: string) => purposeMap[purpose] || purpose

const serviceTypeLabel = (t: string) => {
  const m: Record<string, string> = { text: '文本', image: '图片', video: '视频' }
  return m[t] || t
}

const serviceTypeTag = (t: string): '' | 'success' | 'warning' | 'danger' | 'info' => {
  const m: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = { text: '', image: 'success', video: 'warning' }
  return m[t] || 'info'
}

const statusLabel = (s: string) => {
  const m: Record<string, string> = { pending: '等待中', success: '成功', failed: '失败' }
  return m[s] || s
}

const statusTag = (s: string): '' | 'success' | 'warning' | 'danger' | 'info' => {
  const m: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = { pending: 'warning', success: 'success', failed: 'danger' }
  return m[s] || 'info'
}

const truncate = (text: string, len: number) => {
  if (!text) return '-'
  return text.length > len ? text.slice(0, len) + '...' : text
}

const formatDuration = (ms: number) => {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(1)}s`
}

const formatTime = (timeStr: string) => {
  if (!timeStr) return ''
  return new Date(timeStr).toLocaleString('zh-CN')
}

const formatJSON = (obj: unknown) => {
  try {
    return JSON.stringify(obj, null, 2)
  } catch {
    return String(obj)
  }
}

const loadStats = async () => {
  try {
    const data = await aiLogAPI.getStats() as unknown as AILogStats
    stats.value = data
    if (data.by_purpose_top) {
      purposeOptions.value = data.by_purpose_top.map(p => p.purpose)
    }
  } catch (e) {
    console.error('Failed to load stats', e)
  }
}

const loadLogs = async () => {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      page_size: pagination.page_size
    }
    if (filters.service_type) params.service_type = filters.service_type
    if (filters.purpose) params.purpose = filters.purpose
    if (filters.status) params.status = filters.status
    if (filters.keyword) params.keyword = filters.keyword

    const data = await aiLogAPI.list(params as any) as unknown as { items: AIMessageLog[]; pagination: { total: number; page: number; page_size: number } }
    logs.value = data.items || []
    pagination.total = data.pagination.total
  } catch (e) {
    console.error('Failed to load logs', e)
  } finally {
    loading.value = false
  }
}

const loadData = async () => {
  await Promise.all([loadStats(), loadLogs()])
}

const handleSearch = () => {
  pagination.page = 1
  loadLogs()
}

const resetFilters = () => {
  filters.service_type = ''
  filters.purpose = ''
  filters.status = ''
  filters.keyword = ''
  pagination.page = 1
  loadLogs()
}

const showDetail = (row: AIMessageLog) => {
  currentLog.value = row
  detailVisible.value = true
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.ai-log-container {
  min-height: 100vh;
  background: var(--bg-primary);
}

.header {
  background: var(--bg-card);
  box-shadow: var(--shadow-sm);
  height: 60px !important;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.header-left h2 {
  margin: 0;
  font-size: 18px;
  color: var(--text-primary);
}

.stats-row {
  margin-bottom: var(--space-4);
}

.stat-card {
  cursor: default;
}

.stat-item {
  text-align: center;
  padding: var(--space-3) 0;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  margin: var(--space-2) 0 var(--space-1);
  color: var(--text-primary);
}

.stat-label {
  color: var(--text-secondary);
  font-size: 13px;
}

.filter-card {
  margin-bottom: var(--space-4);
}

.filter-form {
  display: flex;
  flex-wrap: wrap;
  gap: 0;
}

.main-tabs {
  background: var(--bg-card, #fff);
  border-radius: var(--radius-lg, 8px);
  padding: var(--space-4);
}

.table-card {
  margin-bottom: var(--space-4);
}

.clickable-row {
  cursor: pointer;
}

.prompt-preview {
  color: var(--text-secondary);
  font-size: 13px;
  line-height: 1.5;
  word-break: break-all;
}

.purpose-text {
  font-size: 13px;
}

.pagination {
  margin-top: var(--space-4);
  display: flex;
  justify-content: center;
}

.detail-content {
  padding: 0 var(--space-1);
}

.prompt-section {
  margin-top: var(--space-5);
}

.prompt-section h4 {
  margin: 0 0 var(--space-2);
  font-size: 14px;
  color: var(--text-secondary);
}

.error-section {
  margin-top: var(--space-4);
}

.mono-text {
  font-family: 'Courier New', monospace;
  font-size: 12px;
  color: var(--text-muted);
}

:deep(.el-textarea__inner) {
  font-family: 'Courier New', 'Menlo', monospace;
  font-size: 13px;
  line-height: 1.6;
  background: var(--bg-card-hover);
}
</style>

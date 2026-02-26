<template>
  <div class="team-page">
    <div class="team-container">
      <div class="page-header">
        <el-button @click="$router.push('/')" text>
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h2>团队管理</h2>
      </div>

      <el-tabs v-model="activeMainTab" class="main-tabs" @tab-change="handleMainTabChange">
        <!-- ===== 团队信息 Tab ===== -->
        <el-tab-pane label="团队信息" name="team">
          <el-card class="team-info-card" v-loading="loading">
            <template #header>
              <div class="card-header">
                <span>团队信息</span>
                <el-button v-if="isOwner" type="primary" text @click="showEditDialog = true">
                  <el-icon><Edit /></el-icon>
                  编辑
                </el-button>
              </div>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="团队名称">{{ team?.name }}</el-descriptions-item>
              <el-descriptions-item label="创建者">{{ team?.owner?.username }}</el-descriptions-item>
              <el-descriptions-item label="成员数">{{ team?.members?.length || 0 }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatDate(team?.created_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="members-card">
            <template #header>
              <div class="card-header">
                <span>团队成员</span>
                <el-button v-if="isOwner" type="primary" size="small" @click="showInviteDialog = true">
                  <el-icon><Plus /></el-icon>
                  邀请成员
                </el-button>
              </div>
            </template>
            <el-table :data="team?.members || []" style="width: 100%">
              <el-table-column prop="username" label="用户名" />
              <el-table-column prop="email" label="邮箱" />
              <el-table-column prop="role" label="角色">
                <template #default="{ row }">
                  <el-tag :type="row.role === 'owner' ? 'danger' : row.role === 'admin' ? 'warning' : 'info'" size="small">
                    {{ row.role === 'owner' ? '所有者' : row.role === 'admin' ? '管理员' : '成员' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="100" v-if="isOwner">
                <template #default="{ row }">
                  <el-button
                    v-if="row.role !== 'owner'"
                    type="danger"
                    text
                    size="small"
                    @click="handleRemove(row)"
                  >
                    移除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- ===== AI 配置 Tab ===== -->
        <el-tab-pane label="AI 配置" name="ai-config">
          <div class="ai-config-section">
            <div class="section-toolbar">
              <el-button type="primary" @click="showCreateDialog">
                <el-icon><Plus /></el-icon>
                新增配置
              </el-button>
            </div>

            <el-tabs v-model="aiActiveTab" @tab-change="handleAITabChange" class="ai-tabs">
              <el-tab-pane label="文本 LLM" name="text">
                <ConfigList
                  :configs="aiConfigs"
                  :loading="aiLoading"
                  :show-test-button="true"
                  @edit="handleAIEdit"
                  @delete="handleAIDelete"
                  @toggle-active="handleAIToggleActive"
                  @test="handleAITest"
                />
              </el-tab-pane>
              <el-tab-pane label="图片生成" name="image">
                <ConfigList
                  :configs="aiConfigs"
                  :loading="aiLoading"
                  :show-test-button="false"
                  @edit="handleAIEdit"
                  @delete="handleAIDelete"
                  @toggle-active="handleAIToggleActive"
                />
              </el-tab-pane>
              <el-tab-pane label="视频生成" name="video">
                <ConfigList
                  :configs="aiConfigs"
                  :loading="aiLoading"
                  :show-test-button="false"
                  @edit="handleAIEdit"
                  @delete="handleAIDelete"
                  @toggle-active="handleAIToggleActive"
                />
              </el-tab-pane>
            </el-tabs>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- ===== Team Edit Dialog ===== -->
    <el-dialog v-model="showEditDialog" title="编辑团队" width="400px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="团队名称">
          <el-input v-model="editForm.name" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleUpdateTeam">保存</el-button>
      </template>
    </el-dialog>

    <!-- ===== Invite Dialog ===== -->
    <el-dialog v-model="showInviteDialog" title="邀请成员" width="400px">
      <el-form :model="inviteForm" label-width="80px">
        <el-form-item label="邮箱">
          <el-input v-model="inviteForm.email" type="email" placeholder="输入被邀请人的邮箱" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showInviteDialog = false">取消</el-button>
        <el-button type="primary" :loading="inviting" @click="handleInvite">发送邀请</el-button>
      </template>
    </el-dialog>

    <!-- ===== AI Config Create/Edit Dialog ===== -->
    <el-dialog
      v-model="aiDialogVisible"
      :title="aiIsEdit ? '编辑配置' : '新增配置'"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form ref="aiFormRef" :model="aiForm" :rules="aiRules" label-width="100px">
        <el-form-item label="配置名称" prop="name">
          <el-input v-model="aiForm.name" placeholder="输入配置名称" />
        </el-form-item>

        <el-form-item label="厂商" prop="provider">
          <el-select
            v-model="aiForm.provider"
            placeholder="选择厂商"
            @change="handleProviderChange"
            style="width: 100%"
          >
            <el-option
              v-for="provider in allProviders"
              :key="provider.id"
              :label="provider.name"
              :value="provider.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="优先级" prop="priority">
          <el-input-number v-model="aiForm.priority" :min="0" :max="100" :step="1" style="width: 100%" />
          <div class="form-tip">数值越大，优先级越高</div>
        </el-form-item>

        <el-form-item label="模型" prop="model">
          <el-select
            v-model="aiForm.model"
            placeholder="选择或输入模型名称"
            multiple
            filterable
            allow-create
            default-first-option
            collapse-tags
            collapse-tags-tooltip
            style="width: 100%"
          >
            <el-option
              v-for="model in currentProviderModels"
              :key="model"
              :label="model"
              :value="model"
            />
          </el-select>
          <div class="form-tip">可多选或手动输入自定义模型名</div>
        </el-form-item>

        <el-form-item label="Base URL" prop="base_url">
          <el-input v-model="aiForm.base_url" placeholder="https://api.openai.com/v1" />
          <div class="form-tip">
            完整请求端点：{{ fullEndpointExample }}
          </div>
        </el-form-item>

        <el-form-item label="API Key" prop="api_key">
          <el-input
            v-model="aiForm.api_key"
            type="password"
            show-password
            placeholder="输入 API Key"
          />
        </el-form-item>

        <el-form-item v-if="aiIsEdit" label="启用">
          <el-switch v-model="aiForm.is_active" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="aiDialogVisible = false">取消</el-button>
        <el-button
          v-if="aiForm.service_type === 'text'"
          @click="handleAITestConnection"
          :loading="aiTesting"
        >连接测试</el-button>
        <el-button type="primary" @click="handleAISubmit" :loading="aiSubmitting">
          {{ aiIsEdit ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft, Edit, Plus } from '@element-plus/icons-vue'
import { getTeam, updateTeam, inviteMember, removeMember } from '@/api/team'
import { aiAPI } from '@/api/ai'
import { useUserStore } from '@/stores/user'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import type { TeamInfo } from '@/api/auth'
import type { AIServiceConfig, AIServiceType, CreateAIConfigRequest, UpdateAIConfigRequest } from '@/types/ai'
import ConfigList from '@/views/settings/components/ConfigList.vue'

const route = useRoute()
const userStore = useUserStore()

// =============================================
// Team Management State
// =============================================
const team = ref<TeamInfo | null>(null)
const loading = ref(false)
const saving = ref(false)
const inviting = ref(false)
const showEditDialog = ref(false)
const showInviteDialog = ref(false)
const editForm = reactive({ name: '' })
const inviteForm = reactive({ email: '' })

const isOwner = computed(() => userStore.user?.role === 'owner')

const activeMainTab = ref<string>(
  (route.query.tab as string) === 'ai-config' ? 'ai-config' : 'team'
)

function formatDate(dateStr?: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

async function loadTeam() {
  loading.value = true
  try {
    team.value = await getTeam()
    editForm.name = team.value?.name || ''
  } catch {
    ElMessage.error('获取团队信息失败')
  } finally {
    loading.value = false
  }
}

async function handleUpdateTeam() {
  if (!editForm.name.trim()) {
    ElMessage.warning('请输入团队名称')
    return
  }
  saving.value = true
  try {
    team.value = await updateTeam({ name: editForm.name })
    showEditDialog.value = false
    ElMessage.success('更新成功')
  } catch (e: any) {
    ElMessage.error(e?.message || '更新失败')
  } finally {
    saving.value = false
  }
}

async function handleInvite() {
  if (!inviteForm.email.trim()) {
    ElMessage.warning('请输入邮箱')
    return
  }
  inviting.value = true
  try {
    await inviteMember({ email: inviteForm.email })
    showInviteDialog.value = false
    inviteForm.email = ''
    ElMessage.success('邀请已发送')
    loadTeam()
  } catch (e: any) {
    ElMessage.error(e?.message || '邀请失败')
  } finally {
    inviting.value = false
  }
}

async function handleRemove(member: any) {
  try {
    await ElMessageBox.confirm(`确定要移除成员 ${member.username} 吗？`, '确认', { type: 'warning' })
    await removeMember(member.id)
    ElMessage.success('已移除')
    loadTeam()
  } catch { /* cancelled */ }
}

// =============================================
// AI Config State
// =============================================
const aiActiveTab = ref<AIServiceType>('text')
const aiLoading = ref(false)
const aiConfigs = ref<AIServiceConfig[]>([])
const aiDialogVisible = ref(false)
const aiIsEdit = ref(false)
const aiEditingId = ref<number>()
const aiFormRef = ref<FormInstance>()
const aiSubmitting = ref(false)
const aiTesting = ref(false)

const aiForm = reactive<CreateAIConfigRequest & { is_active?: boolean; provider?: string }>({
  service_type: 'text',
  provider: '',
  name: '',
  base_url: '',
  api_key: '',
  model: [],
  priority: 0,
  is_active: true,
})

interface ProviderConfig {
  id: string
  name: string
  models: string[]
}

const providerConfigs: Record<AIServiceType, ProviderConfig[]> = {
  text: [
    { id: 'openai', name: 'OpenAI', models: ['gpt-5.2', 'gemini-3-flash-preview'] },
    { id: 'gemini', name: 'Google Gemini', models: ['gemini-2.5-pro', 'gemini-3-flash-preview'] },
  ],
  image: [
    { id: 'volcengine', name: '火山引擎 Seedream 4.0', models: ['ep-20260217022745-6lcxv'] },
  ],
  video: [
    { id: 'volces', name: '火山引擎 Seedance 1.5 Pro', models: ['ep-20260218045808-4lb26'] },
  ],
}

const allProviders = computed(() => providerConfigs[aiForm.service_type] || [])

const currentProviderModels = computed(() => {
  if (!aiForm.provider) return []
  const provider = (providerConfigs[aiForm.service_type] || []).find(p => p.id === aiForm.provider)
  return provider?.models || []
})

const fullEndpointExample = computed(() => {
  const baseUrl = aiForm.base_url || 'https://api.example.com'
  const provider = aiForm.provider
  const serviceType = aiForm.service_type

  let endpoint = ''
  if (serviceType === 'text') {
    endpoint = (provider === 'gemini' || provider === 'google')
      ? '/v1beta/models/{model}:generateContent'
      : '/chat/completions'
  } else if (serviceType === 'image') {
    endpoint = (provider === 'gemini' || provider === 'google')
      ? '/v1beta/models/{model}:generateContent'
      : '/images/generations'
  } else if (serviceType === 'video') {
    endpoint = (['doubao', 'volcengine', 'volces'].includes(provider || ''))
      ? '/contents/generations/tasks'
      : '/video/generations'
  }
  return baseUrl + endpoint
})

const aiRules: FormRules = {
  name: [{ required: true, message: '请输入配置名称', trigger: 'blur' }],
  provider: [{ required: true, message: '请选择厂商', trigger: 'change' }],
  base_url: [
    { required: true, message: '请输入 Base URL', trigger: 'blur' },
    { type: 'url', message: '请输入正确的 URL 格式', trigger: 'blur' },
  ],
  api_key: [{ required: true, message: '请输入 API Key', trigger: 'blur' }],
  model: [{
    required: true,
    message: '请至少选择一个模型',
    trigger: 'change',
    validator: (_rule: any, value: any, callback: any) => {
      if ((Array.isArray(value) && value.length > 0) || (typeof value === 'string' && value.length > 0)) {
        callback()
      } else {
        callback(new Error('请至少选择一个模型'))
      }
    },
  }],
}

async function loadAIConfigs() {
  aiLoading.value = true
  try {
    aiConfigs.value = await aiAPI.list(aiActiveTab.value)
  } catch (error: any) {
    ElMessage.error(error.message || '加载 AI 配置失败')
  } finally {
    aiLoading.value = false
  }
}

function generateConfigName(provider: string, serviceType: AIServiceType): string {
  const providerNames: Record<string, string> = { openai: 'OpenAI', gemini: 'Gemini', google: 'Google', volcengine: '火山', volces: '火山' }
  const serviceNames: Record<AIServiceType, string> = { text: '文本', image: '图片', video: '视频' }
  const num = Math.floor(Math.random() * 10000).toString().padStart(4, '0')
  return `${providerNames[provider] || provider}-${serviceNames[serviceType] || serviceType}-${num}`
}

function showCreateDialog() {
  aiIsEdit.value = false
  aiEditingId.value = undefined
  resetAIForm()
  aiForm.service_type = aiActiveTab.value
  const providers = providerConfigs[aiActiveTab.value] || []
  const defaultProvider = providers[0]
  if (defaultProvider) {
    aiForm.provider = defaultProvider.id
    if (defaultProvider.id === 'gemini' || defaultProvider.id === 'google') {
      aiForm.base_url = 'https://generativelanguage.googleapis.com'
    } else if (defaultProvider.id === 'volcengine' || defaultProvider.id === 'volces') {
      aiForm.base_url = 'https://ark.cn-beijing.volces.com/api/v3'
    } else {
      aiForm.base_url = 'https://api.openai.com/v1'
    }
    aiForm.name = generateConfigName(defaultProvider.id, aiActiveTab.value)
  }
  aiDialogVisible.value = true
}

function handleAIEdit(config: AIServiceConfig) {
  aiIsEdit.value = true
  aiEditingId.value = config.id
  Object.assign(aiForm, {
    service_type: config.service_type,
    provider: config.provider || 'openai',
    name: config.name,
    base_url: config.base_url,
    api_key: config.api_key,
    model: Array.isArray(config.model) ? config.model : [config.model],
    priority: config.priority || 0,
    is_active: config.is_active,
  })
  aiDialogVisible.value = true
}

async function handleAIDelete(config: AIServiceConfig) {
  try {
    await ElMessageBox.confirm('确定要删除该配置吗？', '警告', {
      confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning',
    })
    await aiAPI.delete(config.id)
    ElMessage.success('删除成功')
    loadAIConfigs()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '删除失败')
  }
}

async function handleAIToggleActive(config: AIServiceConfig) {
  try {
    const newState = !config.is_active
    await aiAPI.update(config.id, { is_active: newState })
    ElMessage.success(newState ? '已启用配置' : '已禁用配置')
    await loadAIConfigs()
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  }
}

async function handleAITestConnection() {
  if (!aiFormRef.value) return
  const valid = await aiFormRef.value.validate().catch(() => false)
  if (!valid) return

  aiTesting.value = true
  try {
    await aiAPI.testConnection({
      base_url: aiForm.base_url,
      api_key: aiForm.api_key,
      model: aiForm.model,
      provider: aiForm.provider,
    })
    ElMessage.success('连接测试成功！')
  } catch (error: any) {
    ElMessage.error(error.message || '连接测试失败')
  } finally {
    aiTesting.value = false
  }
}

async function handleAITest(config: AIServiceConfig) {
  aiTesting.value = true
  try {
    await aiAPI.testConnection({
      base_url: config.base_url,
      api_key: config.api_key,
      model: config.model,
      provider: config.provider,
    })
    ElMessage.success('连接测试成功！')
  } catch (error: any) {
    ElMessage.error(error.message || '连接测试失败')
  } finally {
    aiTesting.value = false
  }
}

async function handleAISubmit() {
  if (!aiFormRef.value) return
  await aiFormRef.value.validate(async (valid) => {
    if (!valid) return
    aiSubmitting.value = true
    try {
      if (aiIsEdit.value && aiEditingId.value) {
        const updateData: UpdateAIConfigRequest = {
          name: aiForm.name, provider: aiForm.provider,
          base_url: aiForm.base_url, api_key: aiForm.api_key,
          model: aiForm.model, priority: aiForm.priority,
          is_active: aiForm.is_active,
        }
        await aiAPI.update(aiEditingId.value, updateData)
        ElMessage.success('更新成功')
      } else {
        await aiAPI.create(aiForm)
        ElMessage.success('创建成功')
      }
      aiDialogVisible.value = false
      loadAIConfigs()
    } catch (error: any) {
      ElMessage.error(error.message || '操作失败')
    } finally {
      aiSubmitting.value = false
    }
  })
}

function handleProviderChange() {
  aiForm.model = []
  const p = aiForm.provider
  if (p === 'gemini' || p === 'google') {
    aiForm.base_url = 'https://generativelanguage.googleapis.com'
  } else if (p === 'volcengine' || p === 'volces') {
    aiForm.base_url = 'https://ark.cn-beijing.volces.com/api/v3'
  } else {
    aiForm.base_url = 'https://api.openai.com/v1'
  }
  if (!aiIsEdit.value) {
    aiForm.name = generateConfigName(p || '', aiForm.service_type)
  }
}

function handleAITabChange(tabName: string | number) {
  aiActiveTab.value = tabName as AIServiceType
  loadAIConfigs()
}

function resetAIForm() {
  Object.assign(aiForm, {
    service_type: aiActiveTab.value || 'text',
    provider: '', name: '', base_url: '', api_key: '',
    model: [], priority: 0, is_active: true,
  })
  aiFormRef.value?.resetFields()
}

// =============================================
// Main Tab Switch
// =============================================
function handleMainTabChange(tabName: string | number) {
  if (tabName === 'ai-config' && aiConfigs.value.length === 0) {
    loadAIConfigs()
  }
}

// =============================================
// Init
// =============================================
onMounted(() => {
  loadTeam()
  if (activeMainTab.value === 'ai-config') {
    loadAIConfigs()
  }
})
</script>

<style scoped>
.team-page {
  min-height: 100vh;
  background: var(--bg-primary, #f5f7fa);
  padding: 24px;
}

.team-container {
  max-width: 960px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  color: var(--text-primary, #1d2129);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.team-info-card {
  margin-bottom: 20px;
}

.members-card {
  margin-bottom: 20px;
}

/* ===== Main Tabs ===== */
.main-tabs {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-primary, #e5e6eb);
  border-radius: var(--radius-lg, 12px);
  padding: 20px;
  box-shadow: var(--shadow-card, 0 1px 4px rgba(0,0,0,0.04));
}

/* ===== AI Config Section ===== */
.ai-config-section {
  padding-top: 4px;
}

.section-toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}

.ai-tabs {
  background: var(--bg-secondary, #f7f8fa);
  border: 1px solid var(--border-primary, #e5e6eb);
  border-radius: var(--radius-md, 8px);
  padding: 16px;
}

.form-tip {
  font-size: 0.75rem;
  color: var(--text-muted, #86909c);
  margin-top: 4px;
}

/* ===== Dialog glass style ===== */
:deep(.el-dialog) {
  background: var(--glass-bg-heavy, rgba(255,255,255,0.95));
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border: 1px solid var(--glass-border, rgba(0,0,0,0.06));
  border-radius: var(--radius-2xl, 16px);
}

:deep(.el-dialog__header) {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-primary, #e5e6eb);
  margin-right: 0;
}

:deep(.el-dialog__title) {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

:deep(.el-dialog__body) {
  padding: 24px;
}

:deep(.el-dialog__footer) {
  padding: 16px 24px;
  border-top: 1px solid var(--border-primary, #e5e6eb);
}

/* ===== Dark Mode ===== */
.dark .main-tabs {
  background: var(--bg-card);
}

.dark .ai-tabs {
  background: var(--bg-card);
}

.dark :deep(.el-dialog) {
  background: var(--glass-bg-heavy);
  border-color: var(--glass-border);
}

.dark :deep(.el-input__wrapper) {
  background: var(--bg-secondary);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
}

.dark :deep(.el-input__inner) {
  color: var(--text-primary);
}

.dark :deep(.el-select .el-input__wrapper) {
  background: var(--bg-secondary);
}
</style>

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

      <div class="main-content">
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
      </div>
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ArrowLeft, Edit, Plus } from '@element-plus/icons-vue'
import { getTeam, updateTeam, inviteMember, removeMember } from '@/api/team'
import { useUserStore } from '@/stores/user'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { TeamInfo } from '@/api/auth'

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
// Init
// =============================================
onMounted(() => {
  loadTeam()
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

/* ===== Main Content ===== */
.main-content {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-primary, #e5e6eb);
  border-radius: var(--radius-lg, 12px);
  padding: 20px;
  box-shadow: var(--shadow-card, 0 1px 4px rgba(0,0,0,0.04));
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
.dark .main-content {
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

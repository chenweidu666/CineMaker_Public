<template>
  <div class="auth-container">
    <div class="auth-card">
      <div class="auth-header">
        <h1 class="auth-title">CineMaker</h1>
        <p class="auth-subtitle">创建你的账号</p>
      </div>

      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent="handleRegister">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" size="large" prefix-icon="User" />
        </el-form-item>

        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" type="email" placeholder="请输入邮箱" size="large" prefix-icon="Message" />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" placeholder="请输入密码（至少6位）" size="large" prefix-icon="Lock" show-password />
        </el-form-item>

        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="form.confirmPassword" type="password" placeholder="请再次输入密码" size="large" prefix-icon="Lock" show-password />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" size="large" class="auth-btn" :loading="loading" native-type="submit">
            注册
          </el-button>
        </el-form-item>
      </el-form>

      <div class="auth-footer">
        <span>已有账号？</span>
        <router-link to="/login" class="auth-link">去登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 50, message: '用户名长度2-50个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 128, message: '密码长度6-128个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

async function handleRegister() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await userStore.register({
      username: form.username,
      email: form.email,
      password: form.password
    })
    ElMessage.success('注册成功')
    router.push('/')
  } catch (e: any) {
    const msg = e?.response?.data?.error?.message || e?.message || '注册失败'
    ElMessage.error(msg)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary, #f5f7fa);
  padding: 20px;
}

.auth-card {
  width: 100%;
  max-width: 420px;
  background: var(--bg-card, #ffffff);
  border-radius: 16px;
  padding: 48px 40px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
}

.auth-header {
  text-align: center;
  margin-bottom: 36px;
}

.auth-title {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary, #1d2129);
  margin: 0 0 8px 0;
}

.auth-subtitle {
  font-size: 14px;
  color: var(--text-secondary, #86909c);
  margin: 0;
}

.auth-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
}

.auth-footer {
  text-align: center;
  margin-top: 20px;
  font-size: 14px;
  color: var(--text-secondary, #86909c);
}

.auth-link {
  color: var(--el-color-primary);
  text-decoration: none;
  margin-left: 4px;
}

.auth-link:hover {
  text-decoration: underline;
}
</style>

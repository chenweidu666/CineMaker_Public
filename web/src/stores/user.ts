import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, register as apiRegister, getMe, refreshToken as apiRefresh } from '@/api/auth'
import type { UserInfo, LoginParams, RegisterParams } from '@/api/auth'
import { ElMessage } from 'element-plus'
import router from '@/router'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const refreshTokenVal = ref(localStorage.getItem('refreshToken') || '')
  const user = ref<UserInfo | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const username = computed(() => user.value?.username || '')
  const teamId = computed(() => user.value?.team_id)

  function setTokens(accessToken: string, refreshTk: string) {
    token.value = accessToken
    refreshTokenVal.value = refreshTk
    localStorage.setItem('token', accessToken)
    localStorage.setItem('refreshToken', refreshTk)
  }

  function clearAuth() {
    token.value = ''
    refreshTokenVal.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
  }

  async function login(params: LoginParams) {
    const res = await apiLogin(params)
    setTokens(res.access_token, res.refresh_token)
    user.value = res.user
    return res
  }

  async function register(params: RegisterParams) {
    const res = await apiRegister(params)
    setTokens(res.access_token, res.refresh_token)
    user.value = res.user
    return res
  }

  async function fetchUser() {
    try {
      const res = await getMe()
      user.value = res
      return res
    } catch {
      clearAuth()
      throw new Error('获取用户信息失败')
    }
  }

  async function refresh() {
    if (!refreshTokenVal.value) {
      clearAuth()
      return
    }
    try {
      const res = await apiRefresh(refreshTokenVal.value)
      setTokens(res.access_token, res.refresh_token)
      user.value = res.user
    } catch {
      clearAuth()
    }
  }

  function logout() {
    clearAuth()
    router.push('/login')
    ElMessage.success('已退出登录')
  }

  return {
    token,
    refreshTokenVal,
    user,
    isLoggedIn,
    username,
    teamId,
    login,
    register,
    fetchUser,
    refresh,
    logout,
    clearAuth,
  }
})

import type { AxiosError, AxiosInstance, AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'
import axios from 'axios'
import { ElMessage } from 'element-plus'

interface CustomAxiosInstance extends Omit<AxiosInstance, 'get' | 'post' | 'put' | 'patch' | 'delete'> {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
}

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 600000,
  headers: {
    'Content-Type': 'application/json'
  }
}) as CustomAxiosInstance

let isRefreshing = false
let pendingRequests: Array<{
  resolve: (token: string) => void
  reject: (error: any) => void
}> = []

request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.success) {
      return res.data
    } else {
      return Promise.reject(new Error(res.error?.message || '请求失败'))
    }
  },
  async (error: AxiosError<any>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    if (error.response?.status === 401 && !originalRequest._retry) {
      if (originalRequest.url?.includes('/auth/refresh') || originalRequest.url?.includes('/auth/login')) {
        return Promise.reject(error)
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          pendingRequests.push({
            resolve: (newToken: string) => {
              originalRequest.headers.Authorization = `Bearer ${newToken}`
              resolve(axios(originalRequest))
            },
            reject
          })
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      const refreshTokenVal = localStorage.getItem('refreshToken')
      if (!refreshTokenVal) {
        isRefreshing = false
        rejectPendingRequests(error)
        handleLogout()
        return Promise.reject(error)
      }

      try {
        const { data } = await axios.post('/api/v1/auth/refresh', { refresh_token: refreshTokenVal })
        const newToken = data.data.access_token
        const newRefresh = data.data.refresh_token
        localStorage.setItem('token', newToken)
        localStorage.setItem('refreshToken', newRefresh)

        pendingRequests.forEach(({ resolve }) => resolve(newToken))
        pendingRequests = []
        isRefreshing = false

        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return axios(originalRequest)
      } catch (refreshError) {
        isRefreshing = false
        rejectPendingRequests(refreshError)
        handleLogout()
        return Promise.reject(error)
      }
    }

    return Promise.reject(error)
  }
)

function rejectPendingRequests(error: any) {
  pendingRequests.forEach(({ reject }) => reject(error))
  pendingRequests = []
}

function handleLogout() {
  localStorage.removeItem('token')
  localStorage.removeItem('refreshToken')
  // Lazy import to avoid circular dependency with Pinia store
  import('@/stores/user').then(({ useUserStore }) => {
    try {
      const userStore = useUserStore()
      userStore.clearAuth()
    } catch {
      // Pinia may not be initialized yet during early 401s
    }
  })
  if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
    ElMessage.warning('登录已过期，请重新登录')
    window.location.href = '/login'
  }
}

export default request

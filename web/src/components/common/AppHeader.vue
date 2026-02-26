<template>
  <div class="app-header-wrapper">
    <header class="app-header" :class="{ 'header-fixed': fixed }">
      <div class="header-content">
        <!-- Left section: Logo + Left slot -->
        <div class="header-left">
          <router-link v-if="showLogo" to="/" class="logo">
            <img src="/cinemaker-favicon.svg" alt="CineMaker" class="logo-icon" />
            <span class="logo-text">CineMaker</span>
          </router-link>
          <!-- Left slot for business content | 左侧插槽用于业务内容 -->
          <slot name="left" />
        </div>

        <!-- Center section: Center slot -->
        <div class="header-center">
          <slot name="center" />
        </div>

        <!-- Right section: Actions + Right slot -->
        <div class="header-right">
          <!-- Right slot for business content (before actions) | 右侧插槽（在操作按钮前） -->
          <slot name="right" />
          <!-- Theme Toggle | 主题切换 -->
          <el-button @click="toggleTheme" class="header-btn theme-toggle-btn" :title="isDark ? '切换到浅色模式' : '切换到深色模式'">
            <el-icon v-if="isDark"><Sunny /></el-icon>
            <el-icon v-else><Moon /></el-icon>
          </el-button>

          <!-- User Dropdown | 用户菜单 -->
          <el-dropdown v-if="userStore.isLoggedIn" trigger="click" @command="handleUserCommand">
            <el-button class="header-btn user-btn">
              <el-icon><UserFilled /></el-icon>
              <span class="btn-text">{{ userStore.username || '用户' }}</span>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="ai-logs">
                  <el-icon><ChatDotRound /></el-icon>
                  AI 日志
                </el-dropdown-item>
                <el-dropdown-item command="team">
                  <el-icon><OfficeBuilding /></el-icon>
                  团队管理
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </header>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Sunny, Moon, UserFilled, OfficeBuilding, SwitchButton, ChatDotRound } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

/**
 * AppHeader - Global application header component
 * 应用顶部头组件
 * 
 * Features | 功能:
 * - Fixed position at top | 固定在顶部
 * - Model switch | 模型切换
 * - Slots support for business content | 支持插槽放置业务内容
 * 
 * Slots | 插槽:
 * - left: Content after logo | logo 右侧内容
 * - center: Center content | 中间内容
 * - right: Content before actions | 操作按钮左侧内容
 */

interface Props {
  /** Fixed position at top | 是否固定在顶部 */
  fixed?: boolean
  /** Show logo | 是否显示 logo */
  showLogo?: boolean
}

withDefaults(defineProps<Props>(), {
  fixed: true,
  showLogo: true,
})

const appRouter = useRouter()
const userStore = useUserStore()

const handleUserCommand = (command: string) => {
  if (command === 'ai-logs') {
    appRouter.push('/ai-logs')
  } else if (command === 'team') {
    appRouter.push('/team')
  } else if (command === 'logout') {
    userStore.logout()
  }
}

const isDark = ref(document.documentElement.classList.contains('dark'))

const toggleTheme = () => {
  isDark.value = !isDark.value
  if (isDark.value) {
    document.documentElement.classList.add('dark')
    localStorage.setItem('theme', 'dark')
  } else {
    document.documentElement.classList.remove('dark')
    localStorage.setItem('theme', 'light')
  }
}
</script>

<style scoped>
.app-header {
  background: var(--glass-bg-heavy);
  border-bottom: 1px solid var(--glass-border);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));
  z-index: 1000;
  position: relative;
}

.app-header::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 10%;
  right: 10%;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--accent), transparent);
  opacity: 0.15;
}

.app-header.header-fixed {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-4);
  height: 70px;
  max-width: 100%;
  margin: 0 auto;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex-shrink: 0;
}

.header-center {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-width: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-shrink: 0;
}

.logo {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  text-decoration: none;
  color: var(--text-primary);
  font-weight: 700;
  font-size: 1.125rem;
  transition: opacity var(--transition-fast);
}

.logo:hover {
  opacity: 0.8;
}

.logo-text {
  background: linear-gradient(135deg, var(--accent) 0%, #06b6d4 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.logo-icon {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.12);
}

:deep(.header-btn) {
  border: 1px solid var(--glass-border, rgba(0,0,0,0.08));
  border-radius: 999px;
  font-weight: 500;
  font-size: 13px;
  padding: 7px 14px;
  background: var(--glass-bg, rgba(255,255,255,0.65));
  color: var(--text-secondary);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  transition: all var(--transition-fast);
}

:deep(.header-btn:hover),
:deep(.header-btn:focus) {
  background: var(--glass-bg-heavy, rgba(255,255,255,0.82));
  color: var(--text-primary);
  border-color: var(--accent, #0ea5e9);
  box-shadow: 0 2px 8px rgba(14, 165, 233, 0.1);
}

:deep(.header-btn .el-icon) {
  font-size: 15px;
}

:deep(.header-btn .btn-text) {
  margin-left: 4px;
}

.theme-toggle-btn {
  width: 36px;
  height: 36px;
  padding: 0 !important;
  border-radius: 50% !important;
  display: flex;
  align-items: center;
  justify-content: center;
}

.theme-toggle-btn .el-icon {
  font-size: 16px;
}

:deep(.user-btn) {
  gap: 4px;
}

:deep(.user-btn .el-icon) {
  font-size: 16px;
}

/* Dark mode adjustments | 深色模式适配 */
.dark .app-header {
  background: var(--glass-bg-heavy);
}

.dark .app-header::after {
  opacity: 0.25;
}

/* ========================================
   Common Slot Styles / 插槽通用样式
   ======================================== */

/* Back Button | 返回按钮 */
:deep(.back-btn) {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

:deep(.back-btn:hover) {
  color: var(--text-primary);
  background: var(--bg-hover);
}

:deep(.back-btn .el-icon) {
  font-size: 1rem;
}

/* Page Title | 页面标题 */
:deep(.page-title) {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

:deep(.page-title h1),
:deep(.header-title),
:deep(.drama-title) {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.3;
}

:deep(.page-title .subtitle) {
  font-size: 0.8125rem;
  color: var(--text-muted);
}

/* Episode Title | 章节标题 */
:deep(.episode-title) {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

/* Responsive | 响应式 */
@media (max-width: 768px) {
  .header-content {
    padding: 0 var(--space-3);
  }
  
  :deep(.btn-text) {
    display: none;
  }
  
  :deep(.header-btn) {
    padding: 8px;
  }

  :deep(.page-title h1),
  :deep(.header-title),
  :deep(.drama-title) {
    font-size: 1rem;
  }

  :deep(.back-btn span) {
    display: none;
  }
}
</style>

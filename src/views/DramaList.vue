<template>
  <div class="max-w-7xl mx-auto px-6 py-8">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-800">项目列表</h1>
      <p class="text-sm text-gray-500 mt-1">管理你的 AI 短剧项目</p>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden cursor-pointer hover:shadow-md transition-shadow"
        @click="$router.push(`/drama/${drama.id}`)"
      >
        <div class="aspect-video bg-gray-100 relative overflow-hidden">
          <img :src="drama.coverImage" class="w-full h-full object-cover" />
          <div class="absolute top-3 right-3">
            <el-tag size="small" effect="dark" type="success">制作中</el-tag>
          </div>
        </div>
        <div class="p-4">
          <h3 class="text-lg font-semibold text-gray-800 mb-1">{{ drama.title }}</h3>
          <p class="text-sm text-gray-500 line-clamp-2 mb-3">{{ drama.description }}</p>
          <div class="flex items-center gap-2">
            <el-tag size="small" effect="plain">{{ drama.style }}</el-tag>
            <el-tag size="small" effect="plain" type="info">{{ drama.episodeCount }} 集</el-tag>
          </div>
        </div>
      </div>

      <div
        class="bg-white rounded-xl shadow-sm border-2 border-dashed border-gray-200 overflow-hidden flex items-center justify-center min-h-[280px] cursor-pointer hover:border-blue-300 transition-colors"
        @click="showNewDialog"
      >
        <div class="text-center text-gray-400">
          <el-icon size="40"><Plus /></el-icon>
          <p class="mt-2 text-sm">创建新项目</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { drama } from '../data/mockData.js'

function showNewDialog() {
  ElMessage.info('演示模式：暂不支持创建新项目')
}
</script>

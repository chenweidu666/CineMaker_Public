<template>
  <div class="max-w-7xl mx-auto px-6 py-6">
    <div class="flex items-center gap-3 mb-6">
      <el-button text @click="$router.push('/')">
        <el-icon><ArrowLeft /></el-icon>
      </el-button>
      <div>
        <h1 class="text-xl font-bold text-gray-800">{{ drama.title }}</h1>
        <p class="text-sm text-gray-500">{{ drama.description }}</p>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="bg-white rounded-xl shadow-sm p-4">
      <el-tab-pane label="角色管理" name="characters">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4 p-2">
          <div
            v-for="char in characters"
            :key="char.id"
            class="bg-gray-50 rounded-lg overflow-hidden border border-gray-100"
          >
            <div class="aspect-square bg-gray-100 relative overflow-hidden">
              <img :src="char.imageUrl" class="w-full h-full object-cover" />
            </div>
            <div class="p-3">
              <div class="flex items-center gap-2 mb-1">
                <span class="font-medium text-gray-800">{{ char.name }}</span>
                <el-tag size="small" effect="plain">{{ char.outfitName }}</el-tag>
              </div>
              <p class="text-xs text-gray-500 line-clamp-3">{{ char.appearance }}</p>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="场景管理" name="scenes">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4 p-2">
          <div
            v-for="scene in scenes"
            :key="scene.id"
            class="bg-gray-50 rounded-lg overflow-hidden border border-gray-100"
          >
            <div class="aspect-video bg-gray-100 relative overflow-hidden">
              <img :src="scene.imageUrl" class="w-full h-full object-cover" />
              <div class="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/60 to-transparent p-2">
                <span class="text-white text-xs">{{ scene.atmosphere }}</span>
              </div>
            </div>
            <div class="p-3">
              <h4 class="font-medium text-gray-800 text-sm mb-1">{{ scene.name }}</h4>
              <p class="text-xs text-gray-500">{{ scene.location }}</p>
              <p class="text-xs text-gray-400 line-clamp-2 mt-1">{{ scene.description }}</p>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="章节管理" name="episodes">
        <div class="p-2">
          <div
            class="flex items-center justify-between bg-gray-50 rounded-lg p-4 border border-gray-100 cursor-pointer hover:bg-gray-100 transition-colors"
            @click="$router.push(`/drama/${drama.id}/episode/${episode.id}`)"
          >
            <div class="flex items-center gap-4">
              <div class="w-32 aspect-video rounded-md overflow-hidden bg-gray-200 flex-shrink-0">
                <img :src="storyboards[0].composedImage" class="w-full h-full object-cover" />
              </div>
              <div>
                <div class="flex items-center gap-2 mb-1">
                  <el-tag size="small" effect="plain">EP01</el-tag>
                  <h4 class="font-medium text-gray-800">{{ episode.title }}</h4>
                </div>
                <p class="text-sm text-gray-500">{{ storyboards.length }} 个分镜 · {{ totalDuration }}s</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <el-tag size="small" :type="episode.status === 'draft' ? 'warning' : 'success'">
                {{ episode.status === 'draft' ? '草稿' : '已完成' }}
              </el-tag>
              <el-icon class="text-gray-400"><ArrowRight /></el-icon>
            </div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { drama, episode, characters, scenes, storyboards } from '../data/mockData.js'

const activeTab = ref('characters')
const totalDuration = computed(() => storyboards.reduce((s, b) => s + b.duration, 0))
</script>

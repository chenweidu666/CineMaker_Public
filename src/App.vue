<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-white shadow-sm border-b border-gray-200 sticky top-0 z-50">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-14">
          <router-link to="/" class="flex items-center gap-2 text-decoration-none">
            <span class="text-xl">🎬</span>
            <span class="text-lg font-bold text-gray-800">CineMaker</span>
            <el-tag size="small" type="warning" effect="plain" class="ml-2">Demo</el-tag>
          </router-link>
          <div class="flex items-center gap-3 text-sm text-gray-500">
            <span>AI短剧生成系统演示</span>
            <el-tag size="small" effect="plain">v2.0</el-tag>
          </div>
        </div>
      </div>
    </header>
    <main>
      <router-view />
    </main>
  </div>
</template>

<script setup>
</script>

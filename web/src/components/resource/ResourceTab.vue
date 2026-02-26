<template>
  <div>
    <!-- Tab Header -->
    <div class="tab-header">
      <h2>{{ $t(config.i18n.tabLabel) }}</h2>
      <div class="tab-actions">
        <el-button :icon="Document" @click="manager.openExtractDialog(episodes)" disabled
          >自动导入（待开发）</el-button>
        <el-button :icon="MagicStick" @click="manager.openAiDialog()"
          >AI生成</el-button>
        <el-button type="primary" :icon="Plus" @click="manager.openAddDialog()"
          >{{ config.i18n.addButton }}</el-button>
      </div>
    </div>

    <!-- Card Grid -->
    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="6" v-for="item in manager.items.value" :key="item.id">
        <el-card shadow="hover" :class="['resource-card', `resource-card--${config.type}`]">
          <div :class="['resource-preview', `resource-preview--${config.type}`]">
            <ImagePreview
              :image-url="manager.getImageUrl(item)"
              :alt="config.card.imageAlt(item)"
              :size="config.type === 'character' ? 200 : 120"
              :show-placeholder-text="false"
            />
          </div>

          <div class="resource-info">
            <div class="resource-name">
              <h4>{{ config.card.title(item) }}</h4>
              <el-tag
                v-if="config.card.subtitleTag?.(item)"
                :type="config.card.subtitleTag(item)?.type || undefined"
                size="small"
              >
                {{ config.card.subtitleTag(item)?.text }}
              </el-tag>
            </div>
            <p class="desc">{{ config.card.description(item) }}</p>
          </div>

          <div class="resource-actions">
            <el-button size="small" @click="manager.openEditDialog(item)">
              {{ $t('common.edit') }}
            </el-button>
            <el-button
              size="small"
              @click="manager.openImageDialog(item)"
            >
              {{ $t('prop.generateImage') }}
            </el-button>
            <el-button size="small" type="danger" @click="manager.deleteItem(item)">
              {{ $t('common.delete') }}
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty
      v-if="!manager.items.value || manager.items.value.length === 0"
      :description="$t(config.i18n.emptyText)"
    />

    <!-- ===== 添加/编辑对话框 ===== -->
    <el-dialog
      v-model="manager.formDialogVisible.value"
      :title="manager.editingItem.value ? $t('common.edit') : $t('common.add')"
      width="700px"
    >
      <el-form :model="manager.formData.value" label-width="100px">
        <!-- 图片上传 -->
        <el-form-item :label="$t('common.image')">
          <el-upload
            class="avatar-uploader"
            :action="`/api/v1/upload/image`"
            :show-file-list="false"
            :on-success="manager.handleImageUploadSuccess"
            :before-upload="manager.beforeAvatarUpload"
          >
            <div
              v-if="manager.hasImage(manager.formData.value)"
              class="avatar-wrapper"
              :style="{ width: config.imageUploadSize.width + 'px', height: config.imageUploadSize.height + 'px', position: 'relative', overflow: 'hidden', borderRadius: '6px' }"
            >
              <img
                :src="manager.getImageUrl(manager.formData.value)"
                class="avatar"
                style="width: 100%; height: 100%; object-fit: cover"
              />
              <div class="avatar-overlay">
                <el-icon style="color: white; font-size: 24px;"><Plus /></el-icon>
              </div>
            </div>
            <div
              v-else
              class="avatar-uploader-icon"
              :style="{ border: '1px dashed #d9d9d9', borderRadius: '6px', cursor: 'pointer', overflow: 'hidden', width: config.imageUploadSize.width + 'px', height: config.imageUploadSize.height + 'px', fontSize: '28px', color: '#8c939d', textAlign: 'center', lineHeight: config.imageUploadSize.height + 'px' }"
            >
              <el-icon><Plus /></el-icon>
            </div>
          </el-upload>
        </el-form-item>

        <!-- 动态表单字段 -->
        <el-form-item
          v-for="field in config.formFields"
          :key="field.key"
          :label="field.label.includes('.') ? $t(field.label) : field.label"
        >
          <el-select
            v-if="field.type === 'select'"
            v-model="manager.formData.value[field.key]"
            :placeholder="$t('common.pleaseSelect')"
          >
            <el-option
              v-for="opt in field.options"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
          <el-input
            v-else-if="field.type === 'textarea'"
            v-model="manager.formData.value[field.key]"
            type="textarea"
            :rows="field.rows || 3"
            :placeholder="field.placeholder || (field.label.includes('.') ? $t(field.label) : field.label)"
          />
          <el-input
            v-else
            v-model="manager.formData.value[field.key]"
            :placeholder="field.placeholder || (field.label.includes('.') ? $t(field.label) : field.label)"
          />
        </el-form-item>

      </el-form>
      <template #footer>
        <el-button @click="manager.formDialogVisible.value = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="manager.saveItem()">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- ===== AI生成对话框 ===== -->
    <el-dialog
      v-model="manager.aiDialogVisible.value"
      :title="config.aiDialogTitle"
      width="700px"
    >
      <div style="margin-bottom: 16px;">
        <h4>输入原始信息</h4>
        <el-input
          v-model="manager.aiInput.value"
          :rows="6"
          type="textarea"
          :placeholder="config.aiInputPlaceholder"
        />
      </div>

      <div style="margin-bottom: 16px;">
        <h4>AI处理结果</h4>
        <el-form label-width="100px">
          <el-form-item
            v-for="field in config.aiFields"
            :key="field.key"
            :label="field.label"
          >
            <el-input
              v-if="field.type === 'textarea'"
              :model-value="manager.aiResult.value[field.key] || ''"
              type="textarea"
              :rows="field.rows || 3"
              :placeholder="field.label"
              readonly
            />
            <el-input
              v-else
              :model-value="manager.aiResult.value[field.key] || ''"
              :placeholder="field.label"
              readonly
            />
          </el-form-item>
        </el-form>
      </div>

      <template #footer>
        <el-button @click="manager.aiDialogVisible.value = false">{{ $t('common.cancel') }}</el-button>
        <el-button @click="manager.aiProcess()" type="primary" plain :loading="manager.aiProcessing.value">AI处理</el-button>
        <el-button @click="manager.saveAiToForm()" type="primary">导入到表单</el-button>
      </template>
    </el-dialog>

    <!-- ===== 图片生成对话框 ===== -->
    <el-dialog
      v-model="manager.imageDialogVisible.value"
      title="生成图片"
      width="600px"
    >
      <el-form label-width="100px">
        <el-form-item label="生成提示词">
          <el-input
            v-model="manager.imagePrompt.value"
            type="textarea"
            :rows="6"
            placeholder="用于生成图片的提示词"
          />
          <div style="display: flex; justify-content: space-between; align-items: center; margin-top: 5px;">
            <div style="color: #999; font-size: 12px;">可以编辑提示词来优化生成效果</div>
            <el-button
              type="primary" plain size="small"
              @click="manager.aiGeneratePrompt()"
              :loading="manager.generatingPrompt.value"
            >AI智能生成提示词</el-button>
          </div>
        </el-form-item>
        <el-form-item label="参考图片">
          <el-upload
            v-model:file-list="manager.referenceImageList.value"
            :action="`/api/v1/upload/image`"
            list-type="picture-card"
            :on-success="manager.handleImageRefSuccess"
            :before-upload="manager.beforeAvatarUpload"
            :limit="5"
            multiple
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
          <div style="color: #999; font-size: 12px; margin-top: 5px;">可以上传最多5张参考图片</div>
        </el-form-item>
        <el-form-item label="图片方向">
          <el-radio-group v-model="manager.imageOrientation.value">
            <el-radio value="horizontal">横屏</el-radio>
            <el-radio value="vertical">竖屏</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="manager.imageDialogVisible.value = false">取消</el-button>
        <el-button @click="manager.saveImageConfig()" type="success" plain>保存配置</el-button>
        <el-button @click="manager.confirmGenerateImage()" type="primary" :loading="manager.generatingImage.value">开始生成</el-button>
      </template>
    </el-dialog>

    <!-- ===== 从剧本提取对话框 ===== -->
    <el-dialog
      v-model="manager.extractDialogVisible.value"
      :title="$t(config.i18n.extractTitle)"
      width="500px"
    >
      <el-form label-width="100px">
        <el-form-item :label="$t('prop.selectEpisode')">
          <el-select
            v-model="manager.selectedEpisodeId.value"
            :placeholder="$t('common.pleaseSelect')"
            style="width: 100%"
          >
            <el-option
              v-for="ep in episodes"
              :key="ep.id"
              :label="ep.title"
              :value="ep.id"
            />
          </el-select>
        </el-form-item>
        <el-alert
          :title="$t(config.i18n.extractTip)"
          type="info"
          :closable="false"
          show-icon
        />
      </el-form>
      <template #footer>
        <el-button @click="manager.extractDialogVisible.value = false">{{ $t('common.cancel') }}</el-button>
        <el-button
          type="primary"
          @click="manager.handleExtract()"
          :disabled="!manager.selectedEpisodeId.value"
        >{{ $t(config.i18n.startExtract) }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'
import { Document, Plus, MagicStick } from '@element-plus/icons-vue'
import { ImagePreview } from '@/components/common'
import { useResourceManager } from '@/composables/useResourceManager'
import type { ResourceConfig } from '@/composables/resourceConfig'

const props = defineProps<{
  config: ResourceConfig
  dramaId: string | number
  dramaStyle?: string
  episodes: any[]
}>()

const emit = defineEmits<{
  (e: 'refresh'): void
}>()

import { ref, toRef, computed } from 'vue'

const dramaIdRef = toRef(props, 'dramaId')
const dramaStyleRef = toRef(props, 'dramaStyle')

const manager = useResourceManager(props.config, dramaIdRef, dramaStyleRef)

watch(() => props.dramaId, (newId) => {
  if (newId) manager.loadItems()
}, { immediate: true })

onUnmounted(() => {
  manager.stopPolling()
})
</script>

<style scoped>
/* ===== Tab Header ===== */
.tab-header {
  display: flex;
  flex-direction: column;
  gap: var(--space-3, 12px);
  margin-bottom: var(--space-4, 16px);
}

@media (min-width: 640px) {
  .tab-header {
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }
}

.tab-header h2 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.tab-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.tab-actions .el-button {
  min-width: 100px;
}

/* ===== Resource Card ===== */
.resource-card {
  margin-bottom: var(--space-4);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-xl);
  overflow: hidden;
  transition: all var(--transition-normal);
  background: var(--glass-bg-heavy);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));
}

.resource-card:hover {
  border-color: var(--accent);
  box-shadow: var(--shadow-card-hover);
  transform: translateY(-2px);
}

.resource-card :deep(.el-card__body) {
  padding: 0;
}

/* ===== Preview Area ===== */
.resource-preview {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 160px;
  background: linear-gradient(135deg, var(--accent) 0%, #06b6d4 100%);
  overflow: hidden;
}

.resource-preview--character {
  height: 520px;
}

.resource-preview--scene {
  height: 160px;
}

.resource-preview--prop {
  height: 160px;
}

.resource-preview img,
.resource-preview :deep(img) {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s;
}

.resource-card:hover .resource-preview img,
.resource-card:hover .resource-preview :deep(img) {
  transform: scale(1.05);
}

/* ===== Info Area ===== */
.resource-info {
  text-align: center;
  padding: var(--space-4, 16px);
}

.resource-name {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: var(--space-2, 8px);
}

.resource-info h4 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.desc {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: var(--space-2, 8px) 0;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
}

/* ===== Actions ===== */
.resource-actions {
  display: flex;
  gap: var(--space-2, 8px);
  justify-content: center;
  padding: 0 var(--space-4, 16px) var(--space-4, 16px);
}

/* ===== Avatar Upload ===== */
.avatar-wrapper:hover .avatar-overlay {
  opacity: 1;
}

.avatar-overlay {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.3s;
  cursor: pointer;
}
</style>

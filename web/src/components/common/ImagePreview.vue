<template>
  <div class="image-preview-wrapper">
    <!-- 缩略图 -->
    <div
      class="thumbnail-container"
      @click="handlePreview"
      :class="{ 'has-image': hasImage }"
    >
      <img v-if="hasImage" :src="imageUrl" :alt="alt" class="thumbnail-image" />
      <div v-else class="no-image-placeholder">
        <el-icon :size="size / 2"><Picture /></el-icon>
        <span v-if="showPlaceholderText">{{ placeholderText }}</span>
      </div>
    </div>

    <!-- 放大预览对话框 -->
    <el-dialog
      v-model="previewVisible"
      :width="dialogWidth"
      align-center
      append-to-body
      :show-close="true"
      class="image-preview-dialog"
      destroy-on-close
      @close="handleClose"
    >
      <template #header>
        <div class="preview-header">
          <span class="preview-title">{{ alt || "图片预览" }}</span>
        </div>
      </template>

      <div class="preview-body" :class="{ 'has-details': hasDetails, 'layout-bottom': detailLayout === 'bottom' }">
        <div class="preview-img-area">
          <img v-if="imageUrl" :src="imageUrl" :alt="alt" class="preview-img" @error="handleImgError" />
          <div v-if="imgError" class="preview-error">
            <el-icon :size="48"><Picture /></el-icon>
            <span>图片加载失败</span>
          </div>
        </div>
        <div v-if="hasDetails" class="preview-details">
          <slot name="details" />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, useSlots } from "vue";
import { Picture } from "@element-plus/icons-vue";

const slots = useSlots();
const hasDetails = computed(() => !!slots.details);

interface Props {
  imageUrl?: string;
  alt?: string;
  size?: number;
  placeholderText?: string;
  showPlaceholderText?: boolean;
  dialogWidth?: string;
  detailLayout?: 'side' | 'bottom';
}

const props = withDefaults(defineProps<Props>(), {
  imageUrl: "",
  alt: "",
  size: 120,
  placeholderText: "暂无图片",
  showPlaceholderText: true,
  dialogWidth: "800px",
  detailLayout: "side",
});

const previewVisible = ref(false);
const imgError = ref(false);

const hasImage = computed(() => {
  return props.imageUrl && props.imageUrl.trim() !== "";
});

const handlePreview = () => {
  if (hasImage.value) {
    imgError.value = false;
    previewVisible.value = true;
  }
};

const handleClose = () => {
  previewVisible.value = false;
};

const handleImgError = () => {
  imgError.value = true;
};
</script>

<style scoped>
.image-preview-wrapper {
  display: inline-block;
  width: 100%;
  height: 100%;
}

.thumbnail-container {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  transition: all var(--transition-fast);
}

.thumbnail-container.has-image {
  cursor: pointer;
}

.thumbnail-container.has-image:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.thumbnail-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
}

.no-image-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--text-muted);
  padding: 16px;
  text-align: center;
}

.no-image-placeholder span {
  font-size: 12px;
}


</style>

<!-- Unscoped: el-dialog teleports to body, scoped styles won't match -->
<style>
.image-preview-dialog .el-dialog {
  border-radius: var(--radius-xl, 16px);
  overflow: hidden;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.image-preview-dialog .el-dialog__body {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  padding: 0;
}

.image-preview-dialog .el-dialog__header {
  padding: 14px 20px;
  border-bottom: 1px solid var(--el-border-color-lighter, #ebeef5);
  margin-right: 0;
}

.image-preview-dialog .preview-title {
  font-size: 16px;
  font-weight: 600;
}

/* --- Pure image mode (no details slot) --- */
.image-preview-dialog .preview-body {
  width: 100%;
  display: flex;
  background: #f5f5f5;
  max-height: calc(90vh - 56px);
  overflow: hidden;
}

.image-preview-dialog .preview-body .preview-img-area {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 300px;
  max-height: 80vh;
}

.image-preview-dialog .preview-img {
  max-width: 100%;
  max-height: 80vh;
  object-fit: contain;
  display: block;
}

/* --- Two-column mode (has details slot) --- */
.image-preview-dialog .preview-body.has-details {
  align-items: stretch;
  background: var(--el-bg-color, #fff);
}

.image-preview-dialog .preview-body.has-details .preview-img-area {
  background: #f5f5f5;
  min-height: 0;
  max-height: none;
  padding: 12px;
}

.image-preview-dialog .preview-body.has-details .preview-img {
  max-height: 70vh;
  border-radius: 8px;
}

.image-preview-dialog .preview-details {
  width: 300px;
  flex-shrink: 0;
  padding: 20px 24px;
  overflow-y: auto;
  border-left: 1px solid var(--el-border-color-lighter, #ebeef5);
}

/* --- Bottom layout mode --- */
.image-preview-dialog .preview-body.layout-bottom {
  flex-direction: column;
}

.image-preview-dialog .preview-body.layout-bottom .preview-img-area {
  min-height: 0;
  max-height: none;
  padding: 16px;
}

.image-preview-dialog .preview-body.layout-bottom .preview-img {
  max-height: 60vh;
  border-radius: 8px;
}

.image-preview-dialog .preview-body.layout-bottom .preview-details {
  width: 100%;
  max-height: 30vh;
  border-left: none;
  border-top: 1px solid var(--el-border-color-lighter, #ebeef5);
  padding: 16px 20px;
}

/* --- Detail sections --- */
.image-preview-dialog .detail-section {
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px dashed var(--el-border-color-extra-light, #f2f6fc);
}

.image-preview-dialog .detail-section:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

.image-preview-dialog .detail-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--el-text-color-secondary, #909399);
  margin-bottom: 6px;
  letter-spacing: 0.5px;
  text-transform: uppercase;
}

.image-preview-dialog .detail-value {
  font-size: 13px;
  line-height: 1.7;
  color: var(--el-text-color-primary, #303133);
  word-break: break-word;
}

.image-preview-dialog .detail-value .el-tag {
  margin-top: 2px;
}

/* --- Error state --- */
.image-preview-dialog .preview-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #909399;
  font-size: 14px;
  padding: 40px;
}

/* --- Dark mode --- */
.dark .image-preview-dialog .preview-body,
.dark .image-preview-dialog .preview-body.has-details .preview-img-area {
  background: #1a1a1a;
}
</style>

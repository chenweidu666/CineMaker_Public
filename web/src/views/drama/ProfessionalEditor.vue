<template>
  <div class="professional-editor">
    <!-- 顶部工具栏 -->
    <AppHeader
      :fixed="false"
      :show-logo="false"
    >
      <template #left>
        <el-button text @click="goBack" class="back-btn">
          <el-icon>
            <ArrowLeft />
          </el-icon>
          <span>{{ $t("editor.backToEpisode") }}</span>
        </el-button>
        <span class="episode-title"
          >{{ drama?.title }} -
          {{ $t("editor.episode", { number: episodeNumber }) }}</span
        >
      </template>
    </AppHeader>

    <!-- 主编辑区域 -->
    <div class="editor-main">
      <!-- 左侧分镜列表 -->
      <div class="storyboard-panel">
        <div class="panel-header">
          <h3>{{ $t("storyboard.scriptStructure") }}</h3>
          <el-button text :icon="Plus" @click="handleAddStoryboard">{{
            $t("storyboard.add")
          }}</el-button>
        </div>

        <div class="storyboard-list">
          <div
            v-for="(shot, index) in storyboards"
            :key="shot.id"
            class="storyboard-item"
            :class="{ active: currentStoryboardId === shot.id }"
            @click="selectStoryboard(shot.id)"
          >
            <div class="shot-content">
              <div class="shot-header">
                <div class="shot-title-row">
                  <span class="shot-number">{{
                    $t("storyboard.shotNumber", {
                      number: shot.storyboard_number,
                    })
                  }}</span>
                  <span class="shot-title">{{
                    shot.title || $t("storyboard.untitled")
                  }}</span>
                </div>
                <div class="shot-actions">
                  <span class="shot-duration">{{ shot.duration }}s</span>
                  <el-button
                    link
                    type="danger"
                    :icon="Delete"
                    @click.stop="handleDeleteStoryboard(shot)"
                    class="delete-btn"
                  />
                </div>
              </div>
              <div class="shot-action" v-if="shot.action">
                {{ shot.action }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 中间时间线编辑区域 -->
      <div class="timeline-area">
        <VideoTimelineEditor
          ref="timelineEditorRef"
          v-if="storyboards.length > 0"
          :scenes="storyboards"
          :episode-id="episodeId.toString()"
          :drama-id="dramaId.toString()"
          :assets="videoAssets"
          :batch-prompt-generating="batchPromptGenerating"
          :batch-image-generating="batchImageGenerating"
          @select-scene="handleTimelineSelect"
          @asset-deleted="loadVideoAssets"
          @merge-completed="handleMergeCompleted"
          @batch-generate-prompts="handleBatchGeneratePrompts"
          @batch-generate-images="handleBatchGenerateImages"
          :batch-video-generating="batchVideoGenerating"
        />
        <el-empty
          v-else
          :description="$t('storyboard.noStoryboard')"
          class="empty-timeline"
        />
      </div>

      <!-- 右侧编辑面板 -->
      <div class="edit-panel">
        <el-tabs v-model="activeTab" class="edit-tabs">
          <!-- 镜头属性标签 -->
          <el-tab-pane
            :label="$t('storyboard.shotProperties')"
            name="shot"
            v-if="currentStoryboard"
          >
            <div v-if="currentStoryboard" class="shot-editor-new">
              <!-- 场景(Scene) -->
              <div class="scene-section">
                <div class="section-label">
                  {{ $t("storyboard.scene") }} (Scene)
                  <el-button
                    size="small"
                    text
                    @click="showSceneSelector = true"
                    >{{ $t("storyboard.selectScene") }}</el-button
                  >
                </div>
                <div
                  class="scene-preview"
                  v-if="hasImage(currentStoryboard.background)"
                  @click="showSceneImage"
                >
                  <img
                    :src="getImageUrl(currentStoryboard.background)"
                    alt="场景"
                    style="cursor: pointer"
                  />
                  <div class="scene-info">
                    <div>
                      {{ currentStoryboard.background.location }} ·
                      {{ currentStoryboard.background.time }}
                    </div>
                    <div class="scene-id">
                      {{ $t("editor.sceneId") }}:
                      {{ currentStoryboard.scene_id || "N/A" }}
                    </div>
                  </div>
                </div>
                <div class="scene-preview-empty" v-else>
                  <el-icon :size="48" color="#666">
                    <Picture />
                  </el-icon>
                  <div>
                    {{
                      currentStoryboard.background
                        ? $t("editor.sceneGenerating")
                        : $t("editor.noBackground")
                    }}
                  </div>
                </div>
              </div>

              <!-- 登场角色(Cast) -->
              <div class="cast-section">
                <div class="section-label">
                  {{ $t("editor.cast") }} (Cast)
                  <el-button
                    size="small"
                    text
                    :icon="Plus"
                    @click="showCharacterSelector = true"
                    >{{ $t("editor.addCharacter") }}</el-button
                  >
                </div>
                <div class="cast-list">
                  <div
                    v-for="char in currentStoryboardCharacters"
                    :key="char.id"
                    class="cast-item active"
                  >
                    <div class="cast-avatar" @click="showCharacterImage(char)">
                      <img
                        v-if="hasImage(char)"
                        :src="getImageUrl(char)"
                        :alt="char.name"
                      />
                      <span v-else>{{ char.name?.[0] || "?" }}</span>
                    </div>
                    <div class="cast-name">{{ char.name }}</div>
                    <div
                      class="cast-remove"
                      @click.stop="toggleCharacterInShot(char.id)"
                      :title="$t('editor.removeCharacter')"
                    >
                      <el-icon :size="14">
                        <Close />
                      </el-icon>
                    </div>
                  </div>
                  <div
                    v-if="
                      !currentStoryboard?.characters ||
                      currentStoryboard.characters.length === 0
                    "
                    class="cast-empty"
                  >
                    {{ $t("editor.noCharacters") }}
                  </div>
                </div>
              </div>

              <!-- 道具(Props) -->
              <div class="cast-section">
                <div class="section-label">
                  {{ $t("editor.props") }} (Props)
                  <el-button
                    size="small"
                    text
                    :icon="Plus"
                    @click="showPropSelector = true"
                    >{{ $t("editor.addProp") }}</el-button
                  >
                </div>
                <div class="cast-list">
                  <div
                    v-for="prop in currentStoryboardProps"
                    :key="prop.id"
                    class="cast-item active"
                  >
                    <div class="cast-avatar">
                      <img
                        v-if="hasImage(prop)"
                        :src="getImageUrl(prop)"
                        :alt="prop.name"
                      />
                      <el-icon v-else>
                        <Box />
                      </el-icon>
                    </div>
                    <div class="cast-name">{{ prop.name }}</div>
                    <div
                      class="cast-remove"
                      @click.stop="togglePropInShot(prop.id)"
                      title="移除道具"
                    >
                      <el-icon :size="14">
                        <Close />
                      </el-icon>
                    </div>
                  </div>
                  <div
                    v-if="
                      !currentStoryboardProps ||
                      currentStoryboardProps.length === 0
                    "
                    class="cast-empty"
                  >
                    {{ $t("editor.noProps") }}
                  </div>
                </div>
              </div>

              <!-- 首帧/中间过程/尾帧画面描述 -->
              <div class="frame-desc-section">
                <div style="display: flex; justify-content: flex-end; margin-bottom: 6px; gap: 6px;">
                  <el-button v-if="!descEditing" size="small" text type="primary" @click="descEditing = true">
                    <el-icon style="margin-right: 4px;"><Edit /></el-icon>编辑描述
                  </el-button>
                  <template v-else>
                    <el-button size="small" @click="cancelDescEdit">取消</el-button>
                    <el-button size="small" type="primary" :loading="descSaving" @click="saveAllDescFields">保存</el-button>
                  </template>
                </div>
                <div class="frame-desc-row">
                  <div class="frame-desc-item">
                    <div class="frame-desc-label">
                      <span class="frame-dot first"></span>
                      首帧画面
                    </div>
                    <el-input
                      v-model="currentStoryboard.first_frame_desc"
                      type="textarea"
                      :autosize="{ minRows: 2, maxRows: 6 }"
                      :readonly="!descEditing"
                      placeholder="镜头起始的静态画面：角色初始位置、姿态、表情、环境状态（可选，分镜拆分时自动生成）"
                    />
                  </div>
                  <div class="frame-desc-item">
                    <div class="frame-desc-label">
                      <span class="frame-dot middle"></span>
                      中间过程
                    </div>
                    <el-input
                      v-model="currentStoryboard.middle_action_desc"
                      type="textarea"
                      :autosize="{ minRows: 2, maxRows: 8 }"
                      :readonly="!descEditing"
                      placeholder="从首帧到尾帧之间的动态过程：角色位移轨迹、动作变化、对白内容及语气、表情变化、环境动态（可选，分镜拆分时自动生成）"
                    />
                  </div>
                  <div class="frame-desc-item">
                    <div class="frame-desc-label">
                      <span class="frame-dot last"></span>
                      尾帧画面
                    </div>
                    <el-input
                      v-model="currentStoryboard.last_frame_desc"
                      type="textarea"
                      :autosize="{ minRows: 2, maxRows: 6 }"
                      :readonly="!descEditing"
                      placeholder="镜头结束的静态画面：动作完成后的角色姿态、表情、环境变化（可选，分镜拆分时自动生成）"
                    />
                  </div>
                </div>
              </div>
            </div>
            <el-empty v-else :description="$t('editor.noShotSelected')" />
          </el-tab-pane>

          <!-- 图片生成标签 -->
          <el-tab-pane :label="$t('editor.shotImage')" name="image">
            <div class="tab-content" v-if="currentStoryboard">
              <div class="image-generation-section">

                <!-- ① 首帧/尾帧分页切换 -->
                <div class="frame-type-tabs">
                  <div
                    class="frame-type-tab"
                    :class="{ active: selectedFrameType === 'first' }"
                    @click="selectedFrameType = 'first'"
                  >
                    <span class="frame-dot first"></span>
                    首帧
                    <el-badge
                      v-if="firstFrameCount > 0"
                      :value="firstFrameCount"
                      type="primary"
                      class="frame-count-badge"
                    />
                  </div>
                  <div
                    class="frame-type-tab"
                    :class="{ active: selectedFrameType === 'last' }"
                    @click="selectedFrameType = 'last'"
                  >
                    <span class="frame-dot last"></span>
                    尾帧
                    <el-badge
                      v-if="lastFrameCount > 0"
                      :value="lastFrameCount"
                      type="warning"
                      class="frame-count-badge"
                    />
                  </div>
                </div>

                <!-- ② 参考资源（可勾选） -->
                <div class="ref-assets-bar" v-if="currentStoryboard">
                  <div class="ref-assets-group" v-if="currentStoryboard.background && hasImage(currentStoryboard.background)">
                    <span class="ref-assets-label">场景</span>
                    <div
                      class="ref-asset-circle selectable"
                      :class="{ selected: selectedRefScene }"
                      :title="(currentStoryboard.background.location || '场景') + (selectedRefScene ? ' (已选)' : ' (未选)')"
                      @click="selectedRefScene = !selectedRefScene"
                    >
                      <img :src="getImageUrl(currentStoryboard.background)" />
                      <div v-if="selectedRefScene" class="ref-check-badge">✓</div>
                    </div>
                  </div>
                  <div class="ref-assets-group" v-if="currentStoryboardCharacters.length > 0">
                    <span class="ref-assets-label">角色</span>
                    <div
                      v-for="char in currentStoryboardCharacters"
                      :key="'ref-char-' + char.id"
                      class="ref-asset-circle selectable"
                      :class="{ selected: selectedRefCharIds.has(char.id), disabled: !hasImage(char) }"
                      :title="(char.name || '角色') + (selectedRefCharIds.has(char.id) ? ' (已选)' : ' (未选)')"
                      @click="hasImage(char) && toggleRefChar(char.id)"
                    >
                      <img v-if="hasImage(char)" :src="getImageUrl(char)" />
                      <span v-else class="ref-asset-initial">{{ char.name?.[0] || '?' }}</span>
                      <div v-if="selectedRefCharIds.has(char.id)" class="ref-check-badge">✓</div>
                    </div>
                  </div>
                  <div class="ref-assets-group" v-if="currentStoryboardProps.length > 0">
                    <span class="ref-assets-label">道具</span>
                    <div
                      v-for="prop in currentStoryboardProps"
                      :key="'ref-prop-' + prop.id"
                      class="ref-asset-circle selectable"
                      :class="{ selected: selectedRefPropIds.has(prop.id), disabled: !hasImage(prop) }"
                      :title="(prop.name || '道具') + (selectedRefPropIds.has(prop.id) ? ' (已选)' : ' (未选)')"
                      @click="hasImage(prop) && toggleRefProp(prop.id)"
                    >
                      <img v-if="hasImage(prop)" :src="getImageUrl(prop)" />
                      <span v-else class="ref-asset-initial">{{ prop.name?.[0] || '?' }}</span>
                      <div v-if="selectedRefPropIds.has(prop.id)" class="ref-check-badge">✓</div>
                    </div>
                  </div>
                  <div
                    v-if="!currentStoryboard.background && currentStoryboardCharacters.length === 0 && currentStoryboardProps.length === 0"
                    class="ref-assets-empty"
                  >
                    暂无参考资源，请在「镜头属性」中添加
                  </div>
                </div>

                <!-- ③ 参数设置 -->
                <div class="gen-params-card">
                  <div class="param-row">
                    <span class="param-label">模型</span>
                    <el-select
                      v-model="selectedImageToImageModel"
                      placeholder="选择图片生成模型"
                      size="small"
                      style="flex: 1;"
                    >
                      <el-option
                        v-for="model in imageModels"
                        :key="model.modelName"
                        :label="`${model.configName || model.modelName}`"
                        :value="model.modelName"
                      />
                    </el-select>
                  </div>
                  <div class="param-row">
                    <span class="param-label">方向</span>
                    <el-radio-group v-model="imageOrientation" size="small">
                      <el-radio-button label="landscape">横屏 16:9</el-radio-button>
                      <el-radio-button label="portrait">竖屏 9:16</el-radio-button>
                    </el-radio-group>
                  </div>

                  <!-- V3: 非第1个镜头 + 首帧模式 - 引用其他镜头视频尾帧 -->
                  <div v-if="!isFirstStoryboard && selectedFrameType === 'first'" style="margin-top: 12px;">
                    <div class="param-row" style="margin-bottom: 8px; flex-wrap: wrap; gap: 6px;">
                      <span class="param-label">参考帧</span>
                      <el-switch v-model="prevFrameEnabled" size="small" style="margin-right: 4px;" />
                      <template v-if="prevFrameEnabled">
                        <el-select
                          v-model="refFrameState.selectedStoryboardId"
                          size="small"
                          style="width: 130px;"
                          placeholder="选择镜头"
                          @change="onRefStoryboardChange"
                        >
                          <el-option
                            v-for="sb in otherStoryboards"
                            :key="sb.id"
                            :value="sb.id"
                            :label="`镜头 #${sb.storyboard_number}`"
                          />
                        </el-select>
                        <el-tag v-if="refFrameState.loading" type="info" size="small">
                          <el-icon class="is-loading" style="margin-right: 4px;"><Loading /></el-icon>加载中...
                        </el-tag>
                      </template>
                    </div>
                    <!-- 帧来源选择 -->
                    <div v-if="prevFrameEnabled && refFrameState.selectedStoryboardId && !refFrameState.loading" class="param-row" style="margin-bottom: 8px; flex-wrap: wrap; gap: 6px;">
                      <span class="param-label">帧来源</span>
                      <el-radio-group :model-value="refFrameState.sourceType" @change="(v: any) => applyRefSourceType(v)" size="small">
                        <el-radio-button value="first">
                          首帧图{{ refFrameState.refImages.first.length ? ` (${refFrameState.refImages.first.length})` : '' }}
                        </el-radio-button>
                        <el-radio-button value="last">
                          尾帧图{{ refFrameState.refImages.last.length ? ` (${refFrameState.refImages.last.length})` : '' }}
                        </el-radio-button>
                        <el-radio-button value="video_last">视频尾帧截图</el-radio-button>
                      </el-radio-group>
                      <!-- 视频尾帧模式下的视频选择 -->
                      <template v-if="refFrameState.sourceType === 'video_last'">
                        <el-select
                          v-if="refFrameState.videos.length > 1"
                          :model-value="refFrameState.selectedVideoId"
                          @change="onRefVideoChange"
                          size="small"
                          style="width: 150px;"
                          placeholder="选择视频"
                        >
                          <el-option
                            v-for="(vid, idx) in refFrameState.videos"
                            :key="vid.id"
                            :value="vid.id"
                            :label="`视频${idx + 1} (${vid.duration || '?'}s)`"
                          />
                        </el-select>
                        <el-button text size="small" @click="fetchRefVideoLastFrame()" style="margin-left: 4px;">
                          <el-icon><Refresh /></el-icon>
                        </el-button>
                      </template>
                      <el-tag v-if="!refFrameState.loading && refFrameState.framePath" type="success" size="small">已就绪</el-tag>
                      <el-tag v-else-if="!refFrameState.loading && refFrameState.errorMsg" type="warning" size="small">{{ refFrameState.errorMsg }}</el-tag>
                    </div>
                    <!-- 首帧图/尾帧图多图缩略选择 -->
                    <div
                      v-if="prevFrameEnabled && refFrameState.selectedStoryboardId && !refFrameState.loading && (refFrameState.sourceType === 'first' || refFrameState.sourceType === 'last')"
                      style="margin-bottom: 8px;"
                    >
                      <div
                        v-if="(refFrameState.sourceType === 'first' ? refFrameState.refImages.first : refFrameState.refImages.last).length > 1"
                        style="display: flex; gap: 6px; flex-wrap: wrap; margin-bottom: 6px;"
                      >
                        <div
                          v-for="(img, idx) in (refFrameState.sourceType === 'first' ? refFrameState.refImages.first : refFrameState.refImages.last)"
                          :key="img.id"
                          style="width: 56px; height: 56px; border-radius: 6px; overflow: hidden; cursor: pointer; border: 2px solid transparent; flex-shrink: 0;"
                          :style="{ borderColor: refFrameState.framePath === (img.local_path || img.image_url) ? 'var(--accent)' : 'transparent' }"
                          @click="refFrameState.framePath = img.local_path || img.image_url || ''; refFrameState.errorMsg = ''"
                        >
                          <el-image :src="getImageUrl(img)" fit="cover" lazy loading="lazy" style="width: 100%; height: 100%;" />
                        </div>
                      </div>
                    </div>
                    <div v-if="prevFrameEnabled && refFrameState.framePath" style="background: #f5f7fa; border-radius: 8px; padding: 12px; border: 1px solid #ebeef5;">
                      <div style="text-align: center; margin-bottom: 10px;">
                        <el-image
                          :src="`/static/${refFrameState.framePath}`"
                          fit="contain"
                          style="max-width: 100%; max-height: 200px; border-radius: 6px;"
                          :preview-src-list="[`/static/${refFrameState.framePath}`]"
                          preview-teleported
                        />
                      </div>
                      <div style="display: flex; gap: 6px; flex-wrap: wrap;">
                        <el-button
                          :type="refFrameState.mode === 'reference' ? 'primary' : 'default'"
                          size="small"
                          @click="useAsReferenceForFirstFrame"
                        >
                          参考图生成首帧
                        </el-button>
                        <el-button
                          :type="refFrameState.mode === 'direct' ? 'primary' : 'default'"
                          size="small"
                          @click="usePrevFrameAsFirstFrame"
                          :loading="refFrameState.importing"
                        >
                          直接用作首帧
                        </el-button>
                      </div>
                      <p style="margin-top: 6px; font-size: 11px; color: #909399; line-height: 1.5;">
                        「参考图生成」推荐：镜头切换（切特写、换角度）；「直接用作首帧」：续拍同一镜头，需手动调整描述。
                      </p>
                    </div>
                  </div>

                  <!-- 尾帧模式：选择首帧参考图 -->
                  <div v-if="selectedFrameType === 'last'" class="first-frame-selector">
                    <div class="param-row" style="margin-bottom: 6px;">
                      <span class="param-label">首帧参考</span>
                      <el-tag v-if="firstFrameImagesForLast.length === 0" type="warning" size="small">暂无首帧图片，请先生成</el-tag>
                    </div>
                    <div v-if="firstFrameImagesForLast.length > 0" class="first-frame-grid">
                      <div
                        v-for="img in firstFrameImagesForLast"
                        :key="img.id"
                        class="first-frame-thumb"
                        :class="{ selected: selectedFirstFrameId === img.id }"
                        @click="selectedFirstFrameId = img.id"
                      >
                        <el-image
                          :src="getImageUrl(img)"
                          fit="cover"
                          lazy
                          style="width: 100%; height: 100%; border-radius: 4px;"
                        />
                        <div v-if="selectedFirstFrameId === img.id" class="selected-badge">✓</div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- ③ 提示词 + 生成按钮 -->
                <div class="prompt-and-action">
                  <div class="prompt-header">
                    <span class="prompt-title">{{ $t("editor.prompt") }}</span>
                    <el-button
                      size="small"
                      type="primary"
                      :disabled="isGeneratingPrompt(currentStoryboard?.id, selectedFrameType)"
                      :loading="isGeneratingPrompt(currentStoryboard?.id, selectedFrameType)"
                      @click="extractFramePrompt"
                    >
                      {{ $t("editor.extractPrompt") }}
                    </el-button>
                  </div>
                  <el-input
                    v-model="currentFramePrompt"
                    type="textarea"
                    :rows="7"
                    :placeholder="$t('editor.promptPlaceholder')"
                  />
                  <div class="action-bar">
                    <el-button
                      type="success"
                      :icon="MagicStick"
                      :loading="generatingImage"
                      :disabled="!currentFramePrompt"
                      @click="generateFrameImage"
                    >
                      {{ generatingImage ? $t("editor.generating") : $t("editor.generateImage") }}
                    </el-button>
                    <el-button :icon="Upload" @click="uploadImage">{{ $t("editor.uploadImage") }}</el-button>
                  </div>
                </div>

                <!-- 生成结果（所有镜头都可查看已有图片） -->
                <div
                  class="generation-result"
                  v-if="generatedImages.length > 0"
                >
                  <div class="section-label" style="display: flex; align-items: center; justify-content: space-between;">
                    <span>{{ $t("editor.generationResult") }} ({{ generatedImages.length }})</span>
                    <el-button
                      v-if="generatedImages.some(img => img.status === 'failed')"
                      type="danger"
                      size="small"
                      text
                      @click="clearFailedImages"
                    >
                      清除失败
                    </el-button>
                  </div>
                  <div class="image-grid">
                    <!-- 尾帧提示：先生成首帧 -->
                    <div
                      v-if="selectedFrameType === 'last' && generatedImages.length === 0"
                      class="last-frame-hint"
                      style="width: 100%; padding: 12px; color: #909399; font-size: 12px; text-align: center;"
                    >
                      尾帧会自动引用首帧图片作为参考，请先生成首帧图片
                    </div>
                    <div
                      v-for="img in generatedImages"
                      :key="img.id"
                      class="image-item-wrapper"
                    >
                      <div
                        class="image-item"
                      >
                        <el-image
                          v-if="hasImage(img)"
                          :src="getImageUrl(img)"
                          :preview-src-list="
                            generatedImages
                              .filter((i) => hasImage(i))
                              .map((i) => getImageUrl(i)!)
                          "
                          :initial-index="
                            generatedImages
                              .filter((i) => i.image_url)
                              .findIndex((i) => i.id === img.id)
                          "
                          fit="cover"
                          lazy
                          loading="lazy"
                          preview-teleported
                        />
                        <div v-else class="image-placeholder">
                          <el-icon :size="32">
                            <Picture />
                          </el-icon>
                          <p>{{ getStatusText(img.status) }}</p>
                        </div>
                        <div class="image-actions" v-if="hasImage(img)">
                          <div></div>
                          <div
                            class="delete-icon-overlay"
                            @click.stop="handleDeleteImage(img)"
                          >
                            <el-icon :size="18" color="red">
                              <DeleteFilled />
                            </el-icon>
                          </div>
                        </div>
                      </div>
                      <div class="image-meta">
                        <el-tag
                          v-if="img.frame_type"
                          :type="{ first: 'primary', last: 'warning', key: 'success', action: 'info', panel: '' }[img.frame_type] || 'info'"
                          size="small"
                          effect="plain"
                        >{{ getFrameTypeLabel(img.frame_type) }}</el-tag>
                        <el-popover
                          v-if="img.prompt"
                          placement="bottom"
                          :width="320"
                          trigger="click"
                        >
                          <template #reference>
                            <el-button size="small" text type="info" style="padding: 0 4px; height: 20px; font-size: 11px;">
                              提示词
                            </el-button>
                          </template>
                          <div style="font-size: 12px; color: #606266; line-height: 1.6; max-height: 200px; overflow-y: auto; word-break: break-all;">
                            {{ img.prompt }}
                          </div>
                        </el-popover>
                        <el-button
                          v-if="hasImage(img)"
                          size="small"
                          text
                          type="primary"
                          :icon="Edit"
                          style="padding: 0 4px; height: 20px; font-size: 11px;"
                          @click="openImageEditor(img)"
                        >
                          编辑
                        </el-button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <el-empty v-else description="未选择镜头" />
          </el-tab-pane>

          <!-- 视频生成标签 -->
          <el-tab-pane :label="$t('video.videoGeneration')" name="video">
            <div class="tab-content" v-if="currentStoryboard">
              <div class="video-generation-section">

                <!-- ① 参考资源 -->
                <div class="ref-assets-bar" v-if="currentStoryboard">
                  <div class="ref-assets-group" v-if="currentStoryboard.background && hasImage(currentStoryboard.background)">
                    <span class="ref-assets-label">场景</span>
                    <div class="ref-asset-circle" :title="currentStoryboard.background.location || '场景'">
                      <img :src="getImageUrl(currentStoryboard.background)" />
                    </div>
                  </div>
                  <div class="ref-assets-group" v-if="currentStoryboardCharacters.length > 0">
                    <span class="ref-assets-label">角色</span>
                    <div
                      v-for="char in currentStoryboardCharacters"
                      :key="'vref-char-' + char.id"
                      class="ref-asset-circle"
                      :title="char.name"
                    >
                      <img v-if="hasImage(char)" :src="getImageUrl(char)" />
                      <span v-else class="ref-asset-initial">{{ char.name?.[0] || '?' }}</span>
                    </div>
                  </div>
                  <div class="ref-assets-group" v-if="currentStoryboardProps.length > 0">
                    <span class="ref-assets-label">道具</span>
                    <div
                      v-for="prop in currentStoryboardProps"
                      :key="'vref-prop-' + prop.id"
                      class="ref-asset-circle"
                      :title="prop.name"
                    >
                      <img v-if="hasImage(prop)" :src="getImageUrl(prop)" />
                      <span v-else class="ref-asset-initial">{{ prop.name?.[0] || '?' }}</span>
                    </div>
                  </div>
                </div>

                <!-- ② 首帧/尾帧对比选择 -->
                <div class="video-frame-compare">
                  <div class="video-frame-card">
                    <div class="video-frame-header">
                      <span class="frame-dot first"></span>
                      <span>首帧</span>
                      <el-select
                        v-if="videoFirstFrameImages.length > 1"
                        v-model="videoFirstFrameId"
                        size="small"
                        style="width: 90px; margin-left: auto;"
                      >
                        <el-option
                          v-for="img in videoFirstFrameImages"
                          :key="img.id"
                          :label="`#${img.id}`"
                          :value="img.id"
                        />
                      </el-select>
                    </div>
                    <div class="video-frame-preview" v-if="selectedVideoFirstFrame">
                      <el-image
                        :src="getImageUrl(selectedVideoFirstFrame)"
                        fit="contain"
                        :preview-src-list="[getImageUrl(selectedVideoFirstFrame)!]"
                        preview-teleported
                      />
                    </div>
                    <div class="video-frame-placeholder" v-else>
                      请先生成首帧图片
                    </div>
                  </div>
                  <div class="video-frame-arrow-center">→</div>
                  <div class="video-frame-card">
                    <div class="video-frame-header">
                      <span class="frame-dot last"></span>
                      <span>尾帧</span>
                      <el-tag size="small" type="info" effect="plain" style="margin-left: 2px;">可选</el-tag>
                      <el-select
                        v-if="videoLastFrameImages.length > 0"
                        v-model="videoLastFrameId"
                        size="small"
                        clearable
                        placeholder="选择"
                        style="width: 90px; margin-left: auto;"
                      >
                        <el-option
                          v-for="img in videoLastFrameImages"
                          :key="img.id"
                          :label="`#${img.id}`"
                          :value="img.id"
                        />
                      </el-select>
                    </div>
                    <div class="video-frame-preview" v-if="selectedVideoLastFrame">
                      <el-image
                        :src="getImageUrl(selectedVideoLastFrame)"
                        fit="contain"
                        :preview-src-list="[getImageUrl(selectedVideoLastFrame)!]"
                        preview-teleported
                      />
                    </div>
                    <div class="video-frame-placeholder" v-else>
                      {{ videoLastFrameImages.length > 0 ? '请选择尾帧' : '请先生成尾帧图片' }}
                    </div>
                  </div>
                </div>

                <!-- ③ 视频参数 -->
                <div class="gen-params-card">
                  <div class="param-row">
                    <span class="param-label">时长</span>
                    <el-slider
                      v-model="videoDuration"
                      :min="4" :max="20" :step="1"
                      show-stops
                      style="flex: 1;"
                    />
                    <span class="param-value">{{ videoDuration }}s</span>
                  </div>
                  <div class="param-inline-group">
                    <div class="param-row">
                      <span class="param-label">清晰度</span>
                      <el-select v-model="videoResolution" size="small" style="width: 90px;">
                        <el-option label="480p" value="480p" />
                        <el-option label="720p" value="720p" />
                      </el-select>
                    </div>
                    <div class="param-row">
                      <span class="param-label">比例</span>
                      <el-select v-model="videoAspectRatio" size="small" style="width: 110px;">
                        <el-option label="16:9 横屏" value="16:9" />
                        <el-option label="9:16 竖屏" value="9:16" />
                      </el-select>
                    </div>
                    <div class="param-row">
                      <span class="param-label">音频</span>
                      <el-switch v-model="videoGenerateAudio" size="small" />
                    </div>
                    <div class="param-row">
                      <span class="param-label">人物对话</span>
                      <el-switch v-model="videoIncludeDialogue" size="small" />
                      <el-tooltip content="开启后视频提示词会包含角色台词和嘴部动作描述，用于后续配音对口型" placement="top">
                        <el-icon style="margin-left: 4px; color: #909399; cursor: help;"><QuestionFilled /></el-icon>
                      </el-tooltip>
                    </div>
                  </div>
                </div>

                <!-- ③ 提示词 + 生成按钮 -->
                <div class="prompt-and-action">
                  <div class="prompt-header">
                    <span class="prompt-title">视频提示词</span>
                    <el-button
                      size="small"
                      type="primary"
                      :icon="MagicStick"
                      @click="autoGenerateVideoPrompt"
                    >
                      AI提取提示词
                    </el-button>
                  </div>
                  <el-input
                    v-model="videoPromptText"
                    type="textarea"
                    :rows="6"
                    placeholder="点击「AI提取提示词」按钮，根据分镜信息自动生成视频提示词"
                    resize="vertical"
                    @input="onVideoPromptInput"
                    @blur="saveVideoPrompt"
                  />
                  <div class="action-bar">
                    <el-button
                      type="success"
                      :icon="VideoCamera"
                      :loading="generatingVideo"
                      :disabled="!selectedVideoModel || !videoPromptText"
                      @click="generateVideoSimple"
                    >
                      {{ generatingVideo ? "生成中..." : "生成视频" }}
                    </el-button>
                  </div>
                </div>

                <!-- 生成的视频列表 -->
                <div
                  class="generation-result"
                  v-if="generatedVideos.length > 0"
                >
                  <div class="section-label">
                    生成结果 ({{ generatedVideos.length }})
                  </div>
                  <div class="image-grid">
                    <div
                      v-for="video in generatedVideos"
                      :key="video.id"
                      class="image-item-wrapper"
                    >
                      <div class="image-item video-item">
                        <div
                          v-if="video.video_url"
                          class="video-thumbnail"
                          @click="playVideo(video)"
                        >
                          <video
                            :src="getVideoUrl(video) + '#t=0.1'"
                            :poster="video.first_frame_url ? fixImageUrl(video.first_frame_url) : undefined"
                            preload="metadata"
                          />
                          <div class="play-overlay">
                            <el-icon :size="40" color="#fff">
                              <VideoPlay />
                            </el-icon>
                          </div>
                        </div>
                        <div v-else class="image-placeholder">
                          <el-icon :size="32">
                            <VideoCamera />
                          </el-icon>
                          <p>{{ getStatusText(video.status) }}</p>
                          <div
                            v-if="video.status === 'failed' && video.error_msg"
                            class="error-message"
                          >
                            <el-alert
                              type="error"
                              :closable="false"
                              show-icon
                              style="margin-top: 8px"
                            >
                              <template #title>
                                <div style="font-size: 12px; line-height: 1.4; word-break: break-all">
                                  {{ video.error_msg }}
                                </div>
                              </template>
                            </el-alert>
                          </div>
                        </div>
                        <!-- 视频操作按钮 -->
                        <div class="video-actions">
                          <div
                            v-if="video.status === 'completed'"
                            class="add-to-assets-button"
                            @click.stop="addVideoToAssets(video)"
                          >
                            <el-icon
                              :size="18"
                              color="var(--text-primary)"
                              v-if="!addingToAssets.has(video.id)"
                            >
                              <FolderAdd />
                            </el-icon>
                            <el-icon
                              :size="18"
                              color="var(--text-primary)"
                              v-else
                              class="is-loading"
                            >
                              <Loading />
                            </el-icon>
                          </div>
                          <div v-else></div>
                          <!-- 删除按钮 -->
                          <div
                            class="delete-video-button"
                            @click.stop="handleDeleteVideo(video)"
                          >
                            <el-icon :size="18" color="red">
                              <DeleteFilled />
                            </el-icon>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <el-empty v-else description="未选择镜头" />
          </el-tab-pane>

          <!-- 音效与配乐标签 -->
          <el-tab-pane :label="$t('video.soundAndMusicTab')" name="audio">
            <div class="tab-content">
              <el-empty :description="$t('video.soundMusicInDev')" />
            </div>
          </el-tab-pane>

          <!-- 视频合成列表标签 -->
          <el-tab-pane :label="$t('video.videoMerge')" name="merges">
            <div class="tab-content">
              <div class="merges-list" v-loading="loadingMerges">
                <el-empty
                  v-if="videoMerges.length === 0"
                  :description="$t('video.noMergeRecords')"
                  :image-size="120"
                >
                  <template #description>
                    <div
                      style="color: #909399; font-size: 14px; margin-top: 12px"
                    >
                      <p style="margin: 0">{{ $t("video.noMergeYet") }}</p>
                      <p style="margin: 8px 0 0 0; font-size: 12px">
                        {{ $t("video.mergeInstructions") }}
                      </p>
                    </div>
                  </template>
                </el-empty>
                <div v-else class="merge-items">
                  <div
                    v-for="merge in videoMerges"
                    :key="merge.id"
                    class="merge-item"
                    :class="'merge-status-' + merge.status"
                  >
                    <!-- 状态指示条 -->
                    <div class="status-indicator"></div>

                    <!-- 主要内容区域 -->
                    <div class="merge-content">
                      <!-- 标题和状态 -->
                      <div class="merge-header">
                        <div class="title-section">
                          <el-icon :size="20" class="title-icon">
                            <VideoCamera v-if="merge.status === 'completed'" />
                            <Loading
                              v-else-if="merge.status === 'processing'"
                              class="rotating"
                            />
                            <WarningFilled
                              v-else-if="merge.status === 'failed'"
                            />
                            <Clock v-else />
                          </el-icon>
                          <h3 class="merge-title">{{ merge.title }}</h3>
                        </div>
                        <el-tag
                          :type="
                            merge.status === 'completed'
                              ? 'success'
                              : merge.status === 'failed'
                                ? 'danger'
                                : 'warning'
                          "
                          effect="dark"
                          size="large"
                          round
                        >
                          {{
                            merge.status === "pending"
                              ? "等待中"
                              : merge.status === "processing"
                                ? "合成中"
                                : merge.status === "completed"
                                  ? "已完成"
                                  : "失败"
                          }}
                        </el-tag>
                      </div>

                      <!-- 详细信息网格 -->
                      <div class="merge-details">
                        <div class="detail-item">
                          <div class="detail-icon">
                            <el-icon :size="16">
                              <Timer />
                            </el-icon>
                          </div>
                          <div class="detail-content">
                            <div class="detail-label">
                              {{ $t("professionalEditor.videoDuration") }}
                            </div>
                            <div class="detail-value">
                              {{
                                merge.duration
                                  ? `${merge.duration}
                              ${$t("professionalEditor.seconds")}`
                                  : "-"
                              }}
                            </div>
                          </div>
                        </div>
                        <div class="detail-item">
                          <div class="detail-icon">
                            <el-icon :size="16">
                              <Calendar />
                            </el-icon>
                          </div>
                          <div class="detail-content">
                            <div class="detail-label">创建时间</div>
                            <div class="detail-value">
                              {{ formatDateTime(merge.created_at) }}
                            </div>
                          </div>
                        </div>
                        <div class="detail-item" v-if="merge.completed_at">
                          <div class="detail-icon">
                            <el-icon :size="16">
                              <Check />
                            </el-icon>
                          </div>
                          <div class="detail-content">
                            <div class="detail-label">完成时间</div>
                            <div class="detail-value">
                              {{ formatDateTime(merge.completed_at) }}
                            </div>
                          </div>
                        </div>
                      </div>

                      <!-- 错误提示 -->
                      <div
                        class="merge-error"
                        v-if="merge.status === 'failed' && merge.error_msg"
                      >
                        <el-alert type="error" :closable="false" show-icon>
                          <template #title>
                            <div style="font-size: 13px; line-height: 1.5">
                              {{ merge.error_msg }}
                            </div>
                          </template>
                        </el-alert>
                      </div>

                      <!-- 操作按钮 -->
                      <div class="merge-actions">
                        <template
                          v-if="
                            merge.status === 'completed' && merge.merged_url
                          "
                        >
                          <el-button
                            type="primary"
                            :icon="VideoCamera"
                            @click="
                              downloadVideo(merge.merged_url, merge.title)
                            "
                            round
                          >
                            下载视频
                          </el-button>
                          <el-button
                            :icon="View"
                            @click="previewMergedVideo(merge.merged_url)"
                            round
                          >
                            在线预览
                          </el-button>
                        </template>
                        <el-button
                          type="danger"
                          :icon="Delete"
                          @click="deleteMerge(merge.id)"
                          round
                        >
                          删除
                        </el-button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>

    <!-- 角色选择器对话框 -->
    <el-dialog
      v-model="showCharacterImagePreview"
      :title="previewCharacter?.name"
      width="600px"
    >
      <div class="character-image-preview" v-if="previewCharacter">
        <img
          v-if="previewCharacter.local_path"
          :src="getImageUrl(previewCharacter)"
          :alt="previewCharacter.name"
        />
        <el-empty v-else description="暂无图片" />
      </div>
      <!-- ... -->
    </el-dialog>

    <!-- 场景大图预览对话框 -->
    <el-dialog
      v-model="showSceneImagePreview"
      :title="
        currentStoryboard?.background
          ? `${currentStoryboard.background.location} · ${currentStoryboard.background.time}`
          : '场景预览'
      "
      width="800px"
    >
      <div
        class="scene-image-preview"
        v-if="currentStoryboard?.background?.image_url"
      >
        <img :src="currentStoryboard.background.image_url" alt="场景" />
      </div>
    </el-dialog>

    <!-- 角色选择对话框 -->
    <el-dialog
      v-model="showCharacterSelector"
      title="添加角色到镜头"
      width="800px"
    >
      <div class="character-selector-grid">
        <div
          v-for="char in availableCharacters"
          :key="char.id"
          class="character-card"
          :class="{ selected: isCharacterInCurrentShot(char.id) }"
          @click="toggleCharacterInShot(char.id)"
        >
          <div class="character-avatar-large">
            <img
              v-if="char.local_path"
              :src="getImageUrl(char)"
              :alt="char.name"
            />
            <span v-else>{{ char.name?.[0] || "?" }}</span>
          </div>
          <div class="character-info">
            <div class="character-name">{{ char.outfit_name || char.name }}</div>
            <div class="character-role">{{ char._parentName ? char._parentName + ' · ' : '' }}{{ char.role || "角色" }}</div>
          </div>
          <div class="character-check" v-if="isCharacterInCurrentShot(char.id)">
            <el-icon color="#409eff" :size="24">
              <Check />
            </el-icon>
          </div>
        </div>
        <div v-if="availableCharacters.length === 0" class="empty-characters">
          <el-empty description="暂无角色，请先在剧集中创建角色" />
        </div>
      </div>
      <template #footer>
        <el-button @click="showCharacterSelector = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 道具选择对话框 -->
    <el-dialog
      v-model="showPropSelector"
      :title="$t('editor.addPropToShot')"
      width="800px"
    >
      <div class="character-selector-grid">
        <div
          v-for="prop in availableProps"
          :key="prop.id"
          class="character-card"
          :class="{ selected: isPropInCurrentShot(prop.id) }"
          @click="togglePropInShot(prop.id)"
        >
          <div class="character-avatar-large">
            <img
              v-if="prop.local_path"
              :src="getImageUrl(prop)"
              :alt="prop.name"
            />
            <el-icon v-else :size="32">
              <Box />
            </el-icon>
          </div>
          <div class="character-info">
            <div class="character-name">{{ prop.name }}</div>
            <div class="character-role">
              {{ prop.type || $t("editor.props") }}
            </div>
          </div>
          <div class="character-check" v-if="isPropInCurrentShot(prop.id)">
            <el-icon color="#409eff" :size="24">
              <Check />
            </el-icon>
          </div>
        </div>
        <div v-if="availableProps.length === 0" class="empty-characters">
          <el-empty :description="$t('editor.noPropsAvailable')" />
        </div>
      </div>
      <template #footer>
        <el-button @click="showPropSelector = false">{{
          $t("common.close")
        }}</el-button>
      </template>
    </el-dialog>

    <!-- 场景选择对话框 -->
    <el-dialog v-model="showSceneSelector" title="选择场景背景" width="800px">
      <div class="scene-selector-grid">
        <div
          v-for="scene in availableScenes"
          :key="scene.id"
          class="scene-card"
          :class="{ selected: currentStoryboard?.scene_id === scene.id }"
          @click="selectScene(scene.id)"
        >
          <div class="scene-image">
            <img
              v-if="hasImage(scene)"
              :src="getImageUrl(scene)"
              :alt="scene.location"
            />
            <el-icon v-else :size="48" color="#ccc">
              <Picture />
            </el-icon>
          </div>
          <div class="scene-info">
            <div class="scene-location">{{ scene.location }}</div>
            <div class="scene-time">{{ scene.time }}</div>
          </div>
        </div>
        <div v-if="availableScenes.length === 0" class="empty-scenes">
          <el-empty description="暂无可用场景" />
        </div>
      </div>
    </el-dialog>

    <!-- 视频预览对话框 -->
    <el-dialog
      v-model="showVideoPreview"
      title="视频预览"
      width="800px"
      :close-on-click-modal="true"
      destroy-on-close
    >
      <div class="video-preview-container" v-if="previewVideo">
        <video
          v-if="previewVideo.video_url"
          :src="getVideoUrl(previewVideo)"
          controls
          autoplay
          style="
            width: 100%;
            max-height: 70vh;
            display: block;
            background: #000;
            border-radius: 8px;
          "
        />
        <div v-else style="text-align: center; padding: 40px">
          <el-icon :size="48" color="#ccc">
            <VideoCamera />
          </el-icon>
          <p style="margin-top: 16px; color: #909399">视频生成中...</p>
        </div>
        <div class="video-meta">
          <div
            style="
              display: flex;
              justify-content: space-between;
              align-items: center;
            "
          >
            <div>
              <el-tag :type="getStatusType(previewVideo.status)" size="small">{{
                getStatusText(previewVideo.status)
              }}</el-tag>
              <span
                v-if="previewVideo.duration"
                style="margin-left: 12px; color: #606266; font-size: 14px"
                >{{ $t("professionalEditor.duration") }}:
                {{ previewVideo.duration
                }}{{ $t("professionalEditor.seconds") }}</span
              >
            </div>
            <el-button
              v-if="previewVideo.video_url"
              size="small"
              @click="window.open(previewVideo.video_url, '_blank')"
            >
              {{ $t("professionalEditor.downloadVideo") }}
            </el-button>
          </div>
          <div
            v-if="previewVideo.prompt"
            style="
              margin-top: 12px;
              font-size: 12px;
              color: #606266;
              line-height: 1.6;
            "
          >
            <strong>提示词：</strong>{{ previewVideo.prompt }}
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 宫格图片编辑器组件 -->
    <GridImageEditor
      v-model="showGridEditor"
      :storyboard-id="currentStoryboard?.id || 0"
      :drama-id="dramaId"
      :all-images="allGeneratedImages"
      @success="handleGridImageSuccess"
    />

    <!-- 图片裁剪对话框 -->
    <ImageCropDialog
      v-model="showCropDialog"
      :image-url="cropImageUrl"
      @save="handleCropSave"
    />

    <!-- V3: 批量任务进度弹窗已注释，分镜结果直接作为视频提示词，不再需要批量生成提示词/图片 -->

    <!-- 单张图片生成确认/进度弹窗 -->
    <el-dialog
      v-model="imageGenDialog.visible"
      :title="imageGenDialog.phase === 'confirm' ? '确认生成图片' : imageGenDialog.phase === 'generating' ? '图片生成中...' : '生成结果'"
      width="560px"
      :close-on-click-modal="true"
      :close-on-press-escape="true"
      :show-close="true"
      :before-close="handleImageGenDialogClose"
    >
      <template #default>
        <!-- 确认阶段 -->
        <template v-if="imageGenDialog.phase === 'confirm'">
          <div style="margin-bottom: 16px;">
            <div style="display: flex; align-items: center; margin-bottom: 10px;">
              <span style="font-weight: 600; color: #303133; width: 80px;">帧类型：</span>
              <el-tag :type="imageGenDialog.frameType === 'first' ? 'success' : imageGenDialog.frameType === 'last' ? 'warning' : 'primary'" size="default">
                {{ getFrameTypeLabel(imageGenDialog.frameType) }}
              </el-tag>
            </div>
            <div style="display: flex; align-items: center; margin-bottom: 10px;">
              <span style="font-weight: 600; color: #303133; width: 80px;">AI 模型：</span>
              <span style="color: #606266; font-size: 13px;">{{ imageGenDialog.model }}</span>
            </div>
            <div style="display: flex; align-items: center; margin-bottom: 10px;">
              <span style="font-weight: 600; color: #303133; width: 80px;">图片尺寸：</span>
              <el-tag
                :type="imageGenDialog.orientation === 'landscape' ? 'success' : 'warning'"
                size="default"
              >{{ imageGenDialog.orientation === 'landscape' ? '横屏 16:9 (2560×1440)' : '竖屏 9:16 (1440×2560)' }}</el-tag>
            </div>
          </div>

          <div style="margin-bottom: 16px;">
            <p style="font-weight: 600; color: #303133; margin-bottom: 8px;">📝 AI提取提示词：</p>
            <div style="background: #f5f7fa; border-radius: 8px; padding: 12px; max-height: 150px; overflow-y: auto; font-size: 13px; color: #606266; line-height: 1.6; border: 1px solid #ebeef5;">
              {{ imageGenDialog.prompt }}
            </div>
          </div>

          <div>
            <p style="font-weight: 600; color: #303133; margin-bottom: 8px;">🖼️ 参考图片（{{ imageGenDialog.referenceImages.length }} 张）：</p>
            <div v-if="imageGenDialog.referenceImages.length > 0" style="background: #f5f7fa; border-radius: 8px; padding: 12px; border: 1px solid #ebeef5;">
              <div v-for="(ref, idx) in imageGenDialog.referenceImages" :key="idx" style="display: flex; align-items: center; padding: 4px 0; font-size: 13px; color: #606266;">
                <span style="margin-right: 8px;">{{ ref.name }}</span>
                <span style="color: #909399; font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 300px;">{{ ref.path }}</span>
              </div>
            </div>
            <div v-else style="color: #909399; font-size: 13px;">
              无参考图片
            </div>
          </div>

          <div v-if="imageGenDialog.frameType === 'last'" style="margin-top: 12px; padding: 10px; background: #fdf6ec; border-radius: 6px; border: 1px solid #faecd8;">
            <span style="color: #e6a23c; font-size: 13px;">⚠ 尾帧生成会自动添加首帧图片作为参考，确保首尾帧一致性</span>
          </div>
        </template>

        <!-- 生成中阶段 -->
        <template v-if="imageGenDialog.phase === 'generating'">
          <div style="text-align: center; padding: 20px 0;">
            <el-progress :percentage="imageGenDialog.progress" :stroke-width="12" style="margin-bottom: 16px;" />
            <p style="color: #606266; font-size: 14px;">{{ imageGenDialog.statusText }}</p>
            <div style="margin-top: 16px;">
              <div style="display: flex; align-items: center; justify-content: center; margin-bottom: 8px;">
                <el-tag size="small" type="info" style="margin-right: 8px;">{{ getFrameTypeLabel(imageGenDialog.frameType) }}</el-tag>
                <span style="color: #909399; font-size: 12px;">参考图 {{ imageGenDialog.referenceImages.length }} 张</span>
              </div>
            </div>
          </div>
        </template>

        <!-- 完成阶段 -->
        <template v-if="imageGenDialog.phase === 'done'">
          <div style="text-align: center; padding: 20px 0;">
            <div v-if="imageGenDialog.error" style="color: #f56c6c; margin-bottom: 12px;">
              <el-icon :size="48" style="margin-bottom: 8px;"><WarningFilled /></el-icon>
              <p style="font-size: 15px; font-weight: 600;">生成失败</p>
              <p style="font-size: 13px; margin-top: 8px;">{{ imageGenDialog.error }}</p>
            </div>
            <div v-else-if="imageGenDialog.aborted" style="color: #e6a23c; margin-bottom: 12px;">
              <p style="font-size: 15px; font-weight: 600;">已终止生成</p>
            </div>
            <div v-else style="color: #67c23a; margin-bottom: 12px;">
              <el-icon :size="48" style="margin-bottom: 8px;"><CircleCheckFilled /></el-icon>
              <p style="font-size: 15px; font-weight: 600;">图片生成完成！</p>
              <p style="font-size: 13px; color: #909399; margin-top: 8px;">弹窗将自动关闭</p>
            </div>
          </div>
        </template>
      </template>

      <template #footer>
        <template v-if="imageGenDialog.phase === 'confirm'">
          <el-button @click="imageGenDialog.visible = false">取消</el-button>
          <el-button @click="debugGenerateImage" title="仅展示 curl 命令，不调用 API">Debug</el-button>
          <el-button type="primary" @click="confirmGenerateImage">确认生成</el-button>
        </template>
        <template v-else-if="imageGenDialog.phase === 'generating'">
          <el-button @click="imageGenDialog.visible = false">后台运行</el-button>
          <el-button type="danger" @click="abortImageGeneration">终止生成</el-button>
        </template>
        <template v-else-if="imageGenDialog.phase === 'done'">
          <el-button type="primary" @click="imageGenDialog.visible = false">关闭</el-button>
        </template>
      </template>
    </el-dialog>

    <!-- Debug - 图片生成请求参数弹窗 -->
    <el-dialog
      v-model="imageDebugDialog.visible"
      title="Debug - 图片生成请求参数"
      width="700px"
    >
      <div style="margin-bottom: 12px;">
        <pre style="background: #1e1e1e; color: #d4d4d4; padding: 12px; border-radius: 8px; font-size: 12px; line-height: 1.5; white-space: pre-wrap; word-break: break-all; max-height: 500px; overflow-y: auto;">{{ imageDebugDialog.curlCommand }}</pre>
      </div>
      <template #footer>
        <el-button @click="copyImageDebugCommand" type="primary" plain>复制</el-button>
        <el-button @click="imageDebugDialog.visible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 视频生成确认/进度弹窗 -->
    <el-dialog
      v-model="videoGenDialog.visible"
      :title="videoGenDialog.phase === 'confirm' ? '确认生成视频' : videoGenDialog.phase === 'generating' ? '视频生成中...' : '生成结果'"
      width="560px"
      :close-on-click-modal="true"
      :close-on-press-escape="true"
      :show-close="true"
      :before-close="handleVideoGenDialogClose"
    >
      <template #default>
        <!-- 确认阶段 -->
        <template v-if="videoGenDialog.phase === 'confirm'">
          <div style="margin-bottom: 14px;">
            <div style="display: flex; align-items: center; margin-bottom: 8px;">
              <span style="font-weight: 600; color: #303133; width: 80px;">模型：</span>
              <span style="color: #606266; font-size: 13px;">{{ videoGenDialog.model }}</span>
            </div>
            <div style="display: flex; align-items: center; margin-bottom: 8px;">
              <span style="font-weight: 600; color: #303133; width: 80px;">预估价格：</span>
              <el-tag size="small" type="warning">¥{{ videoGenDialog.estimatedPrice.toFixed(2) }}</el-tag>
            </div>
            <div style="display: flex; align-items: center; margin-bottom: 8px;">
              <span style="font-weight: 600; color: #303133; width: 80px;">价格：</span>
              <el-tag size="small" type="warning">{{ videoGenDialog.pricing }}</el-tag>
            </div>
            <div style="display: flex; align-items: center; margin-bottom: 8px;">
              <span style="font-weight: 600; color: #303133; width: 80px;">时长：</span>
              <el-tag size="small" type="success">{{ videoDuration }}秒</el-tag>
            </div>
            <div style="display: flex; align-items: center; margin-bottom: 8px;">
              <span style="font-weight: 600; color: #303133; width: 80px;">生成音频：</span>
              <el-tag size="small" :type="videoGenerateAudio ? 'success' : 'info'">{{ videoGenerateAudio ? '是' : '否' }}</el-tag>
            </div>
            <div style="display: flex; align-items: center; margin-bottom: 8px;">
              <span style="font-weight: 600; color: #303133; width: 80px;">人物对话：</span>
              <el-tag size="small" :type="videoIncludeDialogue ? 'success' : 'info'">{{ videoIncludeDialogue ? '是（含嘴部动作）' : '否' }}</el-tag>
            </div>
            <div style="display: flex; align-items: center; margin-bottom: 8px;">
              <span style="font-weight: 600; color: #303133; width: 80px;">参考模式：</span>
              <el-tag size="small" type="primary">{{ videoGenDialog.referenceMode }}</el-tag>
            </div>
            <div style="display: flex; align-items: center; margin-bottom: 8px;">
              <span style="font-weight: 600; color: #303133; width: 80px;">画面比例：</span>
              <el-tag size="small" :type="videoAspectRatio === '9:16' ? 'warning' : 'success'">{{ videoAspectRatio === '9:16' ? '9:16 (竖屏)' : '16:9 (横屏)' }}</el-tag>
            </div>
            <div style="display: flex; align-items: center; margin-bottom: 8px;">
              <span style="font-weight: 600; color: #303133; width: 80px;">视频清晰度：</span>
              <el-radio-group v-model="videoGenDialog.resolution" size="small">
                <el-radio-button label="480p">480p (512x288)</el-radio-button>
                <el-radio-button label="720p">720p (640x360)</el-radio-button>
              </el-radio-group>
            </div>
          </div>

          <div style="margin-bottom: 14px;">
            <p style="font-weight: 600; color: #303133; margin-bottom: 6px;">📝 视频提示词：</p>
            <div style="background: #f5f7fa; border-radius: 8px; padding: 10px; max-height: 120px; overflow-y: auto; font-size: 13px; color: #606266; line-height: 1.6; border: 1px solid #ebeef5;">
              {{ videoGenDialog.prompt || '(无提示词)' }}
            </div>
          </div>

          <div>
            <p style="font-weight: 600; color: #303133; margin-bottom: 6px;">🖼️ 参考图片（{{ videoGenDialog.referenceImages.length }} 张）：</p>
            <div v-if="videoGenDialog.referenceImages.length > 0" style="background: #f5f7fa; border-radius: 8px; padding: 10px; border: 1px solid #ebeef5;">
              <div v-for="(ref, idx) in videoGenDialog.referenceImages" :key="idx" style="display: flex; align-items: center; padding: 3px 0; font-size: 13px; color: #606266;">
                <span style="margin-right: 8px;">{{ ref.name }}</span>
                <span style="color: #909399; font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 300px;">{{ ref.path }}</span>
              </div>
            </div>
            <div v-else style="color: #909399; font-size: 13px;">无参考图片（纯文本生成）</div>
          </div>
        </template>

        <!-- 生成中阶段 -->
        <template v-if="videoGenDialog.phase === 'generating'">
          <div style="text-align: center; padding: 20px 0;">
            <el-progress :percentage="videoGenDialog.progress" :stroke-width="12" style="margin-bottom: 16px;" />
            <p style="color: #606266; font-size: 14px;">{{ videoGenDialog.statusText }}</p>
            <p style="color: #909399; font-size: 12px; margin-top: 8px;">视频生成通常需要1-5分钟，请耐心等待</p>
          </div>
        </template>

        <!-- 完成阶段 -->
        <template v-if="videoGenDialog.phase === 'done'">
          <div style="padding: 12px 0;">
            <!-- DEBUG: curl 命令展示 -->
            <div v-if="videoGenDialog.curlCommand" style="margin-bottom: 12px;">
              <p style="font-size: 13px; font-weight: 600; color: #409eff; margin-bottom: 8px;">📋 curl 命令（未实际调用 API）：</p>
              <pre style="background: #1e1e1e; color: #d4d4d4; padding: 12px; border-radius: 8px; font-size: 12px; line-height: 1.5; white-space: pre-wrap; word-break: break-all; max-height: 400px; overflow-y: auto;">{{ videoGenDialog.curlCommand }}</pre>
            </div>
            <div v-else-if="videoGenDialog.error" style="text-align: center; color: #f56c6c; margin-bottom: 12px;">
              <el-icon :size="48" style="margin-bottom: 8px;"><WarningFilled /></el-icon>
              <p style="font-size: 15px; font-weight: 600;">生成失败</p>
              <p style="font-size: 13px; margin-top: 8px;">{{ videoGenDialog.error }}</p>
            </div>
            <div v-else-if="videoGenDialog.aborted" style="color: #e6a23c; margin-bottom: 12px;">
              <p style="font-size: 15px; font-weight: 600;">已终止生成</p>
            </div>
            <div v-else style="color: #67c23a; margin-bottom: 12px;">
              <el-icon :size="48" style="margin-bottom: 8px;"><CircleCheckFilled /></el-icon>
              <p style="font-size: 15px; font-weight: 600;">视频生成完成！</p>
              <p style="font-size: 13px; color: #909399; margin-top: 8px;">弹窗将自动关闭</p>
            </div>
          </div>
        </template>
      </template>

      <template #footer>
        <template v-if="videoGenDialog.phase === 'confirm'">
          <el-button @click="videoGenDialog.visible = false">取消</el-button>
          <el-button @click="debugGenerateVideo" title="仅展示 curl 命令，不调用 API">Debug</el-button>
          <el-button type="primary" @click="confirmGenerateVideo">确认生成</el-button>
        </template>
        <template v-else-if="videoGenDialog.phase === 'generating'">
          <el-button @click="videoGenDialog.visible = false">后台运行</el-button>
          <el-button type="danger" @click="abortVideoGeneration">终止生成</el-button>
        </template>
        <template v-else-if="videoGenDialog.phase === 'done'">
          <el-button type="primary" @click="videoGenDialog.visible = false">关闭</el-button>
        </template>
      </template>
    </el-dialog>

    <!-- V3 链式视频生成弹窗（已禁用） -->
  </div>
</template>

<script setup lang="ts">
import {
  ref,
  reactive,
  computed,
  watch,
  onMounted,
  onBeforeUnmount,
  nextTick,
} from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  ArrowLeft,
  Plus,
  Picture,
  VideoPlay,
  VideoPause,
  View,
  Setting,
  Upload,
  MagicStick,
  VideoCamera,
  ZoomIn,
  ZoomOut,
  Top,
  Bottom,
  Check,
  Close,
  Right,
  Timer,
  Calendar,
  Clock,
  Loading,
  WarningFilled,
  CircleCheckFilled,
  Delete,
  Connection,
  Box,
  Crop,
  FolderAdd,
  Refresh,
  DeleteFilled,
  Edit,
  QuestionFilled,
} from "@element-plus/icons-vue";
import { dramaAPI } from "@/api/drama";
import { propAPI } from "@/api/prop";
import { generateFramePrompt, batchGenerateFramePrompts, getStoryboardFramePrompts, type FrameType } from "@/api/frame";
import { imageAPI } from "@/api/image";
import { videoAPI } from "@/api/video";
import { aiAPI } from "@/api/ai";
import { assetAPI } from "@/api/asset";
import { videoMergeAPI } from "@/api/videoMerge";
import { taskAPI } from "@/api/task";
import type { ImageGeneration } from "@/types/image";
import type { VideoGeneration } from "@/types/video";
import type { AIServiceConfig } from "@/types/ai";
import type { Asset } from "@/types/asset";
import type { VideoMerge } from "@/api/videoMerge";
import VideoTimelineEditor from "@/components/editor/VideoTimelineEditor.vue";
import GridImageEditor from "@/components/editor/GridImageEditor.vue";
import type { Drama, Episode, Storyboard } from "@/types/drama";
import { AppHeader, ImageCropDialog } from "@/components/common";
import { getImageUrl, hasImage, getVideoUrl, fixImageUrl } from "@/utils/image";

// AI模型配置
interface ModelOption {
  modelName: string;
  configName: string;
  configId: number;
  priority: number;
  price?: number;
}

const route = useRoute();
const router = useRouter();
const { t: $t } = useI18n();

const dramaId = Number(route.params.dramaId);
const episodeNumber = Number(route.params.episodeNumber);
const episodeId = ref<number>(0);

const drama = ref<Drama | null>(null);
const episode = ref<Episode | null>(null);
const storyboards = ref<Storyboard[]>([]);
const characters = ref<any[]>([]);
const availableScenes = ref<any[]>([]);
const props = ref<any[]>([]);
const showPropSelector = ref(false);

const currentStoryboardId = ref<string | null>(null);
const activeTab = ref("shot");
const showSceneSelector = ref(false);
const showCharacterSelector = ref(false);
const showCharacterImagePreview = ref(false);
const previewCharacter = ref<any>(null);
const showSceneImagePreview = ref(false);
const showSettings = ref(false);
const showVideoPreview = ref(false);
const previewVideo = ref<VideoGeneration | null>(null);
const addingToAssets = ref<Set<number>>(new Set());

const currentPlayState = ref<"playing" | "paused">("paused");
const currentTime = ref(0);
const totalDuration = computed(() => {
  if (!Array.isArray(storyboards.value)) return 0;
  return storyboards.value.reduce((sum, s) => sum + (s.duration || 0), 0);
});

const selectedCharacters = ref<number[]>([]);
const narrativeTab = ref("shot-prompt");

// 图片生成相关状态
const selectedFrameType = ref<FrameType>("first");
const panelCount = ref(3);
const firstFrameImagesForLast = ref<any[]>([]);
const selectedFirstFrameId = ref<number | null>(null);
const selectedRefScene = ref(true);
const selectedRefCharIds = ref<Set<number>>(new Set());
const selectedRefPropIds = ref<Set<number>>(new Set());
const selectedRefPrevFrame = ref(false);
const generatingPromptStates = ref<Record<string, boolean>>({}); // 按 "镜头ID_帧类型" 记录生成状态
const framePrompts = ref<Record<string, string>>({
  key: "",
  first: "",
  last: "",
  panel: "",
});
const currentFramePrompt = ref("");
const generatingImageIds = ref<Set<number>>(new Set());
const generatingImage = computed(() => {
  if (!currentStoryboard.value) return false;
  return generatingImageIds.value.has(currentStoryboard.value.id);
});
const generatedImages = ref<ImageGeneration[]>([]);

// 图片模型选择
const imageModels = ref<ModelOption[]>([]);
const selectedImageModel = ref<string>("");
const selectedImageToImageModel = ref<string>("");
const imageOrientation = ref<"landscape" | "portrait">("portrait"); // 图片方向：默认竖屏

// 图片生成弹窗状态
const imageGenDialog = reactive({
  visible: false,
  phase: "confirm" as "confirm" | "generating" | "done",
  prompt: "",
  originalPrompt: "", // 保存原始提示词（不含比例要求）
  frameType: "",
  referenceImages: [] as { name: string; path: string }[],
  model: "使用默认配置模型",
  imageGenId: null as number | null,
  progress: 0,
  statusText: "",
  error: "",
  aborted: false,
  orientation: "landscape" as "landscape" | "portrait", // 横竖屏：landscape（横屏）或 portrait（竖屏）
  width: 2560, // 图片宽度（根据 orientation 自动设置）
  height: 1440, // 图片高度（根据 orientation 自动设置）
});

const imageDebugDialog = reactive({
  visible: false,
  curlCommand: "",
});

// 根据横竖屏生成带比例要求的提示词
const getPromptWithOrientation = (basePrompt: string, orientation: string) => {
  const orientationText = orientation === "landscape" ? "（16:9横屏）" : "（9:16竖屏）";
  const cleanPrompt = basePrompt.replace(/（\d+:\d+[横竖]屏）/g, "").replace(/，\s*$/g, "").replace(/\s*$/g, "");
  return cleanPrompt + orientationText;
};

// 监听 orientation 变化，自动设置 width、height 和更新提示词中的比例要求
watch(
  () => imageGenDialog.orientation,
  (newOrientation) => {
    if (newOrientation === "landscape") {
      imageGenDialog.width = 2560;
      imageGenDialog.height = 1440;
    } else {
      imageGenDialog.width = 1440;
      imageGenDialog.height = 2560;
    }
    // 更新提示词中的比例要求
    if (imageGenDialog.originalPrompt) {
      imageGenDialog.prompt = getPromptWithOrientation(imageGenDialog.originalPrompt, newOrientation);
    }
  }
);

const batchExtractingPrompts = ref(false);
const batchExtractProgress = ref("");
const batchPromptGenerating = ref(false);
const batchImageGenerating = ref(false);
const batchVideoGenerating = ref(false);
const isSwitchingFrameType = ref(false); // 标志位：是否正在切换帧类型

// V3 链式视频生成弹窗状态（已禁用）
// const batchVideoDialog = reactive({
//   visible: false,
//   phase: "config" as "config" | "running" | "done",
//   skipExisting: true,
//   model: "",
//   pricing: "",
//   firstFrameStatus: "" as "" | "ok" | "missing",
//   resolution: "480p" as "480p" | "720p",
//   generateAudio: true,
//   enableSubtitle: false,
//   progress: 0,
//   statusText: "",
//   total: 0,
//   completed: 0,
//   failed: 0,
//   aborted: false,
//   submitted: false,
//   error: "",
// });

// 批量任务进度弹窗状态
const batchTaskDialog = reactive({
  visible: false,
  type: "prompt" as "prompt" | "image",
  title: "",
  phase: "config" as "config" | "running" | "done",
  skipExisting: true,
  frameType: "first",
  progress: 0,
  statusText: "",
  detail: "",
  total: 0,
  completed: 0,
  failed: 0,
});
const loadingImages = ref(false);
let pollingTimer: any = null;
let pollingFrameType: FrameType | null = null; // 记录正在轮询的帧类型

// 宫格图片编辑器状态
const showGridEditor = ref(false);

// 所有已生成的图片（用于宫格编辑器选择）
const allGeneratedImages = ref<ImageGeneration[]>([]);

// 首帧/尾帧图片数量（用于分页标签上的角标）
const firstFrameCount = computed(() =>
  allGeneratedImages.value.filter(
    (img) => img.frame_type === "first" && img.status === "completed"
  ).length
);
const lastFrameCount = computed(() =>
  allGeneratedImages.value.filter(
    (img) => img.frame_type === "last" && img.status === "completed"
  ).length
);

// 图片裁剪对话框状态
const showCropDialog = ref(false);
const cropImageUrl = ref<string>("");
const cropImageData = ref<ImageGeneration | null>(null);

// 视频生成相关状态
const videoDuration = ref(4); // 默认4秒（最短时长），会根据镜头duration自动更新
const videoGenerateAudio = ref(false); // 是否生成音频（默认：false，需手动开启）
const videoIncludeDialogue = ref(false); // 是否包含人物对话/嘴部动作（独立于音频）
const videoEnableSubtitle = ref(false); // 是否生成字幕（默认：false）
const videoAspectRatio = ref("9:16"); // 画面比例：默认竖屏 9:16
const selectedVideoFrameType = ref<FrameType>("first");
const selectedImagesForVideo = ref<number[]>([]);
const selectedLastImageForVideo = ref<number | null>(null);
const generatingVideoIds = ref<Set<number>>(new Set());
const generatingVideo = computed(() => {
  if (!currentStoryboard.value) return false;
  return generatingVideoIds.value.has(currentStoryboard.value.id);
});
const generatedVideos = ref<VideoGeneration[]>([]);

// 手动选择的视频输入图片
const manuallySelectedVideoImage = ref<any>(null);
const useGeneratedImageForVideo = ref(true); // 是否使用已生成的镜头图片作为视频输入

// 视频生成：首帧/尾帧分别选择
const videoFirstFrameId = ref<number | null>(null);
const videoLastFrameId = ref<number | null>(null);

// 视频生成弹窗状态
const videoGenDialog = reactive({
  visible: false,
  phase: "confirm" as "confirm" | "generating" | "done",
  prompt: "",
  model: "",
  pricing: "",
  estimatedPrice: 0,
  referenceMode: "",
  referenceImages: [] as { name: string; path: string }[],
  videoGenId: null as number | null,
  progress: 0,
  statusText: "",
  error: "",
  curlCommand: "", // DEBUG: 用于展示 curl 命令
  aborted: false,
  resolution: "480p" as "480p" | "720p",
  requestParams: null as any,
});
const videoAssets = ref<Asset[]>([]);
const loadingVideos = ref(false);
const timelineEditorRef = ref<InstanceType<typeof VideoTimelineEditor> | null>(
  null,
);
const videoReferenceImages = ref<ImageGeneration[]>([]);
const selectedVideoModel = ref<string>("");
const selectedReferenceMode = ref<string>(""); // 参考图模式：single, first_last, multiple, none
const videoPromptText = ref<string>(""); // 可编辑的视频提示词
const videoPromptHasReference = ref<boolean>(false); // 标记：提示词是否包含参考图片描述
const videoResolution = ref<string>("480p"); // 视频清晰度：480p 或 720p（默认480p对应512x288）
const previewImageUrl = ref<string>(""); // 预览大图的URL
const videoModelCapabilities = ref<VideoModelCapability[]>([]);
let videoPollingTimer: any = null;
let mergePollingTimer: any = null; // 视频合成列表轮询定时器

// 视频合成列表
const videoMerges = ref<VideoMerge[]>([]);
const loadingMerges = ref(false);

// 视频模型能力配置
interface VideoModelCapability {
  id: string;
  name: string;
  provider: string; // 提供方名称
  pricing: string; // 价格信息
  supportMultipleImages: boolean; // 支持多张图片
  supportFirstLastFrame: boolean; // 支持首尾帧
  supportSingleImage: boolean; // 支持单图
  supportTextOnly: boolean; // 支持纯文本
  supportAudio: boolean; // 支持音频生成
  maxImages: number; // 最多支持几张图片
}

// 模型能力默认配置（作为后备）
const defaultModelCapabilities: Record<
  string,
  Omit<VideoModelCapability, "id" | "name" | "provider">
> = {
  "doubao-seedance-1-5-pro-251215": {
    pricing: "有声¥0.016/无声¥0.008/千tokens",
    supportSingleImage: true,
    supportMultipleImages: true,
    supportFirstLastFrame: true,
    supportTextOnly: true,
    supportAudio: true,
    maxImages: 2,
  },
};

const extractProviderFromModel = (_modelName: string): string => {
  return "doubao";
};

// 加载视频AI配置
const loadVideoModels = async () => {
  try {
    const configs = await aiAPI.list("video");

    // 只显示启用的配置
    const activeConfigs = configs.filter((c) => c.is_active);

    // 展开模型列表并去重
    const allModels = activeConfigs
      .flatMap((config) => {
        const models = Array.isArray(config.model)
          ? config.model
          : [config.model];
        // 解析 settings 中的价格信息
        let pricePerImage = 0;
        try {
          if (config.settings) {
            const settings = JSON.parse(config.settings);
            pricePerImage = settings.price_per_image || 0;
          }
        } catch (e) {
        }
        return models.map((modelName) => ({
          modelName,
          configName: config.name,
          provider: config.provider || config.name,
          priority: config.priority || 0,
          price: pricePerImage
        }));
      })
      .sort((a, b) => b.priority - a.priority);

    // 按模型名称去重
    const modelMap = new Map<
      string,
      { configName: string; provider: string; priority: number }
    >();
    allModels.forEach((model) => {
      if (!modelMap.has(model.modelName)) {
        modelMap.set(model.modelName, {
          configName: model.configName,
          provider: model.provider,
          priority: model.priority,
        });
      }
    });

    // 构建模型能力列表
    videoModelCapabilities.value = Array.from(modelMap.entries()).map(
      ([modelName, info]) => {
        const capability = defaultModelCapabilities[modelName] || {
          pricing: "-",
          supportSingleImage: true,
          supportMultipleImages: false,
          supportFirstLastFrame: false,
          supportTextOnly: true,
          maxImages: 1,
        };

        return {
          id: modelName,
          name: modelName,
          provider: info.provider || info.configName,
          ...capability,
        };
      },
    );

    // 过滤掉纯文生视频模型，只保留支持图生视频的模型
    videoModelCapabilities.value = videoModelCapabilities.value.filter(
      (model) => model.supportSingleImage || model.supportMultipleImages || model.supportFirstLastFrame
    );

    // 从localStorage加载已保存的模型配置
    const savedVideoModel = localStorage.getItem(`ai_video_model_${dramaId}`);
    if (savedVideoModel) {
      const availableModelNames = videoModelCapabilities.value.map((m) => m.id);
      if (availableModelNames.includes(savedVideoModel)) {
        selectedVideoModel.value = savedVideoModel;
      }
    }

    if (!selectedVideoModel.value && videoModelCapabilities.value.length > 0) {
      selectedVideoModel.value = videoModelCapabilities.value[0].id;
    }
  } catch (error: any) {
    console.error("加载视频模型配置失败:", error);
    ElMessage.error("加载视频模型失败");
  }
};

// 加载图片AI配置
const loadImageModels = async () => {
  try {
    const configs = await aiAPI.list("image");

    // 只显示启用的配置
    const activeConfigs = configs.filter((c) => c.is_active);

    // 展开模型列表并去重
    const allModels = activeConfigs
      .flatMap((config) => {
        const models = Array.isArray(config.model)
          ? config.model
          : [config.model];
        // 解析 settings 中的价格信息
        let pricePerImage = 0;
        try {
          if (config.settings) {
            const settings = JSON.parse(config.settings);
            pricePerImage = settings.price_per_image || 0;
          }
        } catch (e) {
        }
        return models.map((modelName) => ({
          modelName,
          configName: config.name,
          provider: config.provider || config.name,
          priority: config.priority || 0,
          price: pricePerImage
        }));
      })
      .sort((a, b) => b.priority - a.priority);

    // 按模型名称去重
    const modelMap = new Map<string, ModelOption>();
    allModels.forEach((model) => {
      if (!modelMap.has(model.modelName)) {
        modelMap.set(model.modelName, {
          modelName: model.modelName,
          configName: model.configName,
          configId: 0,
          priority: model.priority,
          price: model.price
        });
      }
    });

    // 显示所有图片模型
    imageModels.value = Array.from(modelMap.values());

    // 从localStorage加载已保存的模型配置
    const savedImageToImageModel = localStorage.getItem(`ai_image_to_image_model_${dramaId}`);
    if (savedImageToImageModel) {
      const availableModelNames = imageModels.value.map((m) => m.modelName);
      if (availableModelNames.includes(savedImageToImageModel)) {
        selectedImageToImageModel.value = savedImageToImageModel;
      }
    }

    // 自动选择默认图片模型：优先选数据库中标记为默认的模型
    if (!selectedImageToImageModel.value && imageModels.value.length > 0) {
      // 先从配置中找到标记为默认的模型
      const defaultConfig = activeConfigs.find(c => c.is_default);
      let defaultModelId: string | undefined;
      if (defaultConfig) {
        const defaultModels = Array.isArray(defaultConfig.model) ? defaultConfig.model : [defaultConfig.model];
        defaultModelId = defaultModels[0];
      }
      
      // 找到对应的模型
      const defaultModel = defaultModelId ? imageModels.value.find(m => m.modelName === defaultModelId) : undefined;
      
      selectedImageToImageModel.value = defaultModel?.modelName || imageModels.value[0].modelName;
    }
  } catch (error: any) {
    console.error("加载图片模型配置失败:", error);
    ElMessage.error("加载图片模型失败");
  }
};

// 加载视频素材库
const loadVideoAssets = async () => {
  try {
    const result = await assetAPI.listAssets({
      drama_id: dramaId.toString(),
      episode_id: episodeId.value,
      type: "video",
      page: 1,
      page_size: 100,
    });
    // 检查数据结构并正确赋值
    videoAssets.value = result.items || [];
  } catch (error: any) {
    console.error("加载视频素材库失败:", error);
  }
};

// 当前模型能力
const currentModelCapability = computed(() => {
  return videoModelCapabilities.value.find(
    (m) => m.id === selectedVideoModel.value,
  );
});

// Seedance 1.5 Pro 始终支持音频
const selectedModelSupportsAudio = computed(() => true);

// 计算输入的 token 数量（简单估算：中文字符数 + 英文单词数）
const calculateInputTokens = (text: string): number => {
  if (!text) return 0;
  const chineseChars = (text.match(/[\u4e00-\u9fa5]/g) || []).length;
  const englishWords = (text.match(/[a-zA-Z]+/g) || []).length;
  return chineseChars + englishWords;
};

// 计算预估价格（基于实际测试数据）
const calculateEstimatedPrice = (tokens: number, imageCount: number, modelId: string, generateAudio: boolean, duration: number): number => {
  // 基于实际测试数据发现的规律：
  // - tokens与视频时长成正比
  // - 平均每秒约 21,000 tokens
  // - 费用 = 时长 × 0.21 (¥0.01/千tokens × 21,000 tokens/秒 / 1000)
  return duration * 0.21;
};

// 当前模型支持的参考图模式
const availableReferenceModes = computed(() => {
  const capability = currentModelCapability.value;
  if (!capability) return [];

  const modes: Array<{ value: string; label: string; description?: string }> =
    [];

  if (capability.supportTextOnly) {
    modes.push({ value: "none", label: "纯文本", description: "不使用参考图" });
  }
  if (capability.supportSingleImage) {
    modes.push({
      value: "single",
      label: "单图",
      description: "使用单张参考图",
    });
  }
  if (capability.supportFirstLastFrame) {
    modes.push({
      value: "first_last",
      label: "首尾帧",
      description: "使用首帧和尾帧",
    });
  }
  if (capability.supportMultipleImages) {
    modes.push({
      value: "multiple",
      label: "多图",
      description: `最多${capability.maxImages}张`,
    });
  }

  return modes;
});

// 帧提示词存储key生成函数
const getPromptStorageKey = (
  storyboardId: number | undefined,
  frameType: FrameType,
) => {
  if (!storyboardId) return null;
  return `frame_prompt_${storyboardId}_${frameType}`;
};

const isCharacterSelected = (charId: number) => {
  return selectedCharacters.value.includes(charId);
};

const toggleCharacter = (charId: number) => {
  const index = selectedCharacters.value.indexOf(charId);
  if (index > -1) {
    selectedCharacters.value.splice(index, 1);
  } else {
    selectedCharacters.value.push(charId);
  }
};

const currentStoryboard = computed(() => {
  if (!currentStoryboardId.value) return null;
  return (
    storyboards.value.find(
      (s) => String(s.id) === String(currentStoryboardId.value),
    ) || null
  );
});

// 获取上一个镜头
const previousStoryboard = computed(() => {
  if (!currentStoryboardId.value || storyboards.value.length < 2) return null;
  const currentIndex = storyboards.value.findIndex(
    (s) => String(s.id) === String(currentStoryboardId.value),
  );
  if (currentIndex <= 0) return null;
  return storyboards.value[currentIndex - 1];
});

// V3: 判断当前是否为第一个镜头（只有第一个镜头需要手动生成首帧图片）
const isFirstStoryboard = computed(() => {
  if (!currentStoryboardId.value || storyboards.value.length === 0) return false;
  const currentIndex = storyboards.value.findIndex(
    (s) => String(s.id) === String(currentStoryboardId.value),
  );
  return currentIndex === 0;
});

// V3: 上一镜头视频尾帧开关
const prevFrameEnabled = ref(false);

// V3: 可选择的其他镜头列表（排除当前镜头）
const otherStoryboards = computed(() => {
  if (!currentStoryboardId.value) return [];
  return storyboards.value.filter(
    (s) => String(s.id) !== String(currentStoryboardId.value),
  );
});

// V3: 参考帧状态（可引用任意镜头的首帧图/尾帧图/视频尾帧）
const refFrameState = reactive({
  loading: false,
  importing: false,
  framePath: "" as string,
  errorMsg: "" as string,
  videos: [] as { id: number; created_at: string; duration?: number; model: string }[],
  selectedStoryboardId: null as number | null,
  selectedVideoId: null as number | null,
  mode: "" as "" | "reference" | "direct",
  sourceType: "first" as "first" | "last" | "video_last",
  refImages: { first: [] as any[], last: [] as any[] },
});

// V3: 加载指定镜头的首帧/尾帧生成图
const loadRefStoryboardImages = async (storyboardId: number) => {
  refFrameState.refImages = { first: [], last: [] };
  try {
    const [firstRes, lastRes] = await Promise.all([
      imageAPI.listImages({ storyboard_id: storyboardId, frame_type: "first", page: 1, page_size: 20 }),
      imageAPI.listImages({ storyboard_id: storyboardId, frame_type: "last", page: 1, page_size: 20 }),
    ]);
    refFrameState.refImages.first = (firstRes.items || []).filter((img: any) => img.status === "completed");
    refFrameState.refImages.last = (lastRes.items || []).filter((img: any) => img.status === "completed");
  } catch (e) {
    console.error("加载参考镜头图片失败:", e);
  }
};

// V3: 从指定镜头的已完成视频截取尾帧
const fetchRefVideoLastFrame = async (videoId?: number) => {
  if (!refFrameState.selectedStoryboardId) return;

  const targetSb = storyboards.value.find(
    (s) => Number(s.id) === refFrameState.selectedStoryboardId,
  );

  refFrameState.loading = true;
  refFrameState.framePath = "";
  refFrameState.errorMsg = "";
  refFrameState.videos = [];
  refFrameState.selectedVideoId = null;

  try {
    const result = await videoAPI.extractLastFrame(refFrameState.selectedStoryboardId, videoId);
    if (result.videos && result.videos.length > 0) {
      refFrameState.videos = result.videos;
    }
    if (result.success && result.frame_path) {
      refFrameState.framePath = result.frame_path;
      refFrameState.errorMsg = "";
      if (result.video_id) {
        refFrameState.selectedVideoId = result.video_id;
      }
    } else if (!result.has_video) {
      refFrameState.framePath = "";
      refFrameState.errorMsg = `镜头 #${targetSb?.storyboard_number || '?'} 还没有已完成的视频。`;
    } else {
      refFrameState.framePath = "";
      refFrameState.errorMsg = result.message || "截取尾帧失败";
    }
  } catch (error: any) {
    refFrameState.framePath = "";
    refFrameState.errorMsg = error.message || "网络错误";
  } finally {
    refFrameState.loading = false;
  }
};

// V3: 根据帧来源类型应用选中的参考帧
const applyRefSourceType = (type: "first" | "last" | "video_last") => {
  refFrameState.sourceType = type;
  refFrameState.framePath = "";
  refFrameState.errorMsg = "";
  refFrameState.mode = "";
  selectedRefPrevFrame.value = false;

  if (type === "video_last") {
    fetchRefVideoLastFrame();
  } else {
    const imgs = type === "first" ? refFrameState.refImages.first : refFrameState.refImages.last;
    if (imgs.length > 0) {
      const img = imgs[0];
      refFrameState.framePath = img.local_path || img.image_url || "";
      if (!refFrameState.framePath) {
        refFrameState.errorMsg = "该图片无可用路径";
      }
    } else {
      const targetSb = storyboards.value.find(
        (s) => Number(s.id) === refFrameState.selectedStoryboardId,
      );
      const label = type === "first" ? "首帧" : "尾帧";
      refFrameState.errorMsg = `镜头 #${targetSb?.storyboard_number || '?'} 还没有已完成的${label}图。`;
    }
  }
};

const onRefStoryboardChange = async (storyboardId: number) => {
  refFrameState.selectedStoryboardId = storyboardId;
  refFrameState.mode = "";
  refFrameState.framePath = "";
  refFrameState.errorMsg = "";
  selectedRefPrevFrame.value = false;
  refFrameState.sourceType = "first";
  refFrameState.loading = true;

  await loadRefStoryboardImages(storyboardId);

  // 自动选最优来源：优先首帧 > 尾帧 > 视频尾帧
  if (refFrameState.refImages.first.length > 0) {
    refFrameState.sourceType = "first";
    applyRefSourceType("first");
  } else if (refFrameState.refImages.last.length > 0) {
    refFrameState.sourceType = "last";
    applyRefSourceType("last");
  } else {
    refFrameState.sourceType = "video_last";
    await fetchRefVideoLastFrame();
  }
  refFrameState.loading = false;
};

const onRefVideoChange = (videoId: number) => {
  refFrameState.selectedVideoId = videoId;
  fetchRefVideoLastFrame(videoId);
};

// V3: 将参考镜头视频尾帧导入为当前镜头的首帧图片
const usePrevFrameAsFirstFrame = async () => {
  if (!refFrameState.framePath || !currentStoryboard.value) return;

  refFrameState.mode = "direct";
  const sb = currentStoryboard.value;
  const targetSb = storyboards.value.find(
    (s) => Number(s.id) === refFrameState.selectedStoryboardId,
  );
  refFrameState.importing = true;
  try {
    await imageAPI.uploadImage({
      storyboard_id: Number(sb.id),
      drama_id: Number(sb.episode?.drama_id || dramaId),
      frame_type: "first",
      image_url: refFrameState.framePath,
      prompt: `镜头 #${targetSb?.storyboard_number || '?'} ${refFrameState.sourceType === 'first' ? '首帧图' : refFrameState.sourceType === 'last' ? '尾帧图' : '视频尾帧截取'}`,
    });
    ElMessage.success("已导入为当前镜头首帧");
    await loadStoryboardImages(sb.id, "first");
    selectedFrameType.value = "first";
  } catch (error: any) {
    ElMessage.error("导入失败: " + (error.message || "未知错误"));
  } finally {
    refFrameState.importing = false;
  }
};

// V3: 用参考镜头尾帧作为参考图来生成当前镜头首帧
const useAsReferenceForFirstFrame = () => {
  if (!refFrameState.framePath || !currentStoryboard.value) return;
  refFrameState.mode = "reference";
  selectedFrameType.value = "first";
  selectedRefPrevFrame.value = true;
  const targetSb = storyboards.value.find(
    (s) => Number(s.id) === refFrameState.selectedStoryboardId,
  );
  const srcLabel = refFrameState.sourceType === 'first' ? '首帧图' : refFrameState.sourceType === 'last' ? '尾帧图' : '视频尾帧';
  ElMessage.info(`已将镜头 #${targetSb?.storyboard_number || '?'} ${srcLabel}加入参考图，请编辑提示词后点击生成`);
};

// V3: 切换镜头时重置参考帧状态
watch(
  [currentStoryboardId, () => isFirstStoryboard.value],
  () => {
    prevFrameEnabled.value = false;
    refFrameState.mode = "";
    refFrameState.framePath = "";
    refFrameState.errorMsg = "";
    refFrameState.videos = [];
    refFrameState.selectedStoryboardId = null;
    refFrameState.selectedVideoId = null;
    refFrameState.sourceType = "first";
    refFrameState.refImages = { first: [], last: [] };
  },
);

// 打开参考帧开关时，默认选中上一个镜头并自动加载
watch(prevFrameEnabled, (enabled) => {
  if (enabled && !isFirstStoryboard.value && previousStoryboard.value) {
    if (!refFrameState.selectedStoryboardId) {
      refFrameState.selectedStoryboardId = Number(previousStoryboard.value.id);
      onRefStoryboardChange(refFrameState.selectedStoryboardId);
    }
  } else if (!enabled) {
    refFrameState.mode = "";
    selectedRefPrevFrame.value = false;
  }
});

// 上一个镜头的尾帧图片列表（支持多个）
const previousStoryboardLastFrames = ref<any[]>([]);

// 加载上一个镜头的尾帧
const loadPreviousStoryboardLastFrame = async () => {
  if (!previousStoryboard.value) {
    previousStoryboardLastFrames.value = [];
    return;
  }
  try {
    const result = await imageAPI.listImages({
      storyboard_id: previousStoryboard.value.id,
      frame_type: "last",
      page: 1,
      page_size: 10,
    });
    const images = result.items || [];
    previousStoryboardLastFrames.value = images.filter(
      (img: any) => img.status === "completed" && img.image_url,
    );
  } catch (error) {
    console.error("加载上一镜头尾帧失败:", error);
    previousStoryboardLastFrames.value = [];
  }
};

// 选择上一镜头尾帧作为首帧参考
const selectPreviousLastFrame = (img: any) => {
  // 检查是否已选中，已选中则取消
  const currentIndex = selectedImagesForVideo.value.indexOf(img.id);
  if (currentIndex > -1) {
    selectedImagesForVideo.value.splice(currentIndex, 1);
    ElMessage.success("已取消首帧参考");
    return;
  }

  // 参考handleImageSelect的逻辑，根据模式处理
  if (
    !selectedReferenceMode.value ||
    selectedReferenceMode.value === "single"
  ) {
    // 单图模式或未选模式：直接替换
    selectedImagesForVideo.value = [img.id];
  } else if (selectedReferenceMode.value === "first_last") {
    // 首尾帧模式：作为首帧参考
    selectedImagesForVideo.value = [img.id];
  } else if (selectedReferenceMode.value === "multiple") {
    // 多图模式：添加到列表
    const capability = currentModelCapability.value;
    if (
      capability &&
      selectedImagesForVideo.value.length >= capability.maxImages
    ) {
      ElMessage.warning(`最多只能选择${capability.maxImages}张图片`);
      return;
    }
    selectedImagesForVideo.value.push(img.id);
  }
  ElMessage.success("已添加为首帧参考");
};

// 监听帧类型切换，从存储中加载或清空
watch(selectedFrameType, async (newType) => {
  // 切换帧类型时，停止之前的轮询，避免旧结果覆盖新帧类型
  stopPolling();

  if (!currentStoryboard.value) {
    currentFramePrompt.value = "";
    generatedImages.value = [];
    return;
  }

  // 设置切换标志，防止watch(currentFramePrompt)错误保存
  isSwitchingFrameType.value = true;

  // 优先从 sessionStorage 中加载该帧类型的提示词（确保数据准确）
  const storageKey = `frame_prompt_${currentStoryboard.value.id}_${newType}`;
  const stored = sessionStorage.getItem(storageKey);

  if (stored) {
    currentFramePrompt.value = stored;
    framePrompts.value[newType] = stored;
  } else {
    // 如果 sessionStorage 中没有，从数据库加载
    try {
      const prompts = await getStoryboardFramePrompts(currentStoryboard.value.id);
      if (prompts?.frame_prompts) {
        const matched = prompts.frame_prompts.find((p: any) => p.frame_type === newType);
        if (matched) {
          currentFramePrompt.value = matched.prompt;
          framePrompts.value[newType] = matched.prompt;
          sessionStorage.setItem(storageKey, matched.prompt);
        } else {
          currentFramePrompt.value = framePrompts.value[newType] || "";
        }
      } else {
        currentFramePrompt.value = framePrompts.value[newType] || "";
      }
    } catch (e) {
      currentFramePrompt.value = framePrompts.value[newType] || "";
    }
  }

  // 重新加载该帧类型的图片
  loadStoryboardImages(currentStoryboard.value.id, newType);

  // 如果切换到尾帧，加载首帧图片列表供用户选择
  if (newType === "last" && currentStoryboard.value) {
    await loadFirstFrameImagesForLast(currentStoryboard.value.id);
  } else {
    firstFrameImagesForLast.value = [];
    selectedFirstFrameId.value = null;
  }

  // 重置切换标志
  setTimeout(() => {
    isSwitchingFrameType.value = false;
  }, 0);
});

// 监听当前分镜切换，重置提示词
watch(currentStoryboard, async (newStoryboard) => {
  descEditing.value = false;
  if (!newStoryboard) {
    currentFramePrompt.value = "";
    generatedImages.value = [];
    generatedVideos.value = [];
    videoReferenceImages.value = [];
    previousStoryboardLastFrames.value = [];
    videoPromptText.value = "";
    videoFirstFrameId.value = null;
    videoLastFrameId.value = null;
    stopVideoPolling();
    return;
  }

  stopVideoPolling();

  generatedVideos.value = [];
  videoFirstFrameId.value = null;
  videoLastFrameId.value = null;

  videoPromptText.value = newStoryboard.video_prompt || "";

  // 检查提示词是否包含参考图片描述
  videoPromptHasReference.value = checkPromptHasReference(newStoryboard.video_prompt || "");

  // 设置切换标志
  isSwitchingFrameType.value = true;

  // 清空 framePrompts 对象，避免显示上一个镜头的提示词
  framePrompts.value = {
    key: "",
    first: "",
    last: "",
    panel: "",
  };

  // 加载当前帧类型的提示词
  const storageKey = getPromptStorageKey(
    newStoryboard.id,
    selectedFrameType.value,
  );

  let promptLoaded = false;
  if (storageKey) {
    const stored = sessionStorage.getItem(storageKey);
    if (stored) {
      currentFramePrompt.value = stored;
      framePrompts.value[selectedFrameType.value] = stored;
      promptLoaded = true;
    }
  }

  // 如果 sessionStorage 中没有提示词，从数据库加载
  if (!promptLoaded) {
    try {
      const prompts = await getStoryboardFramePrompts(newStoryboard.id);
      if (prompts?.frame_prompts) {
        const matched = prompts.frame_prompts.find((p: any) => p.frame_type === selectedFrameType.value);
        if (matched) {
          currentFramePrompt.value = matched.prompt;
          framePrompts.value[selectedFrameType.value] = matched.prompt;
          // 同步到 sessionStorage 以便后续快速访问
          if (storageKey) {
            sessionStorage.setItem(storageKey, matched.prompt);
          }
        } else {
          currentFramePrompt.value = "";
        }
      } else {
        currentFramePrompt.value = "";
      }
    } catch (e) {
      currentFramePrompt.value = "";
    }
  }

  // 重置切换标志
  setTimeout(() => {
    isSwitchingFrameType.value = false;
  }, 0);

  // 初始化参考资源勾选：默认全选
  selectedRefScene.value = true;
  selectedRefCharIds.value = new Set(
    (currentStoryboardCharacters.value || []).filter((c: any) => c.local_path).map((c: any) => c.id)
  );
  selectedRefPropIds.value = new Set(
    (currentStoryboardProps.value || []).filter((p: any) => p.local_path).map((p: any) => p.id)
  );
  selectedRefPrevFrame.value = false;

  // 加载该分镜的图片列表（根据当前选择的帧类型）
  await loadStoryboardImages(newStoryboard.id, selectedFrameType.value);

  // 如果当前是尾帧模式，加载首帧图片列表
  if (selectedFrameType.value === "last") {
    await loadFirstFrameImagesForLast(newStoryboard.id);
  }

  // 加载所有已生成的图片（用于宫格编辑器）
  await loadAllGeneratedImages();

  // 加载视频参考图片（所有帧类型）
  await loadVideoReferenceImages(newStoryboard.id);

  // 加载该分镜的视频列表
  await loadStoryboardVideos(newStoryboard.id);

  // 加载上一镜头的尾帧
  await loadPreviousStoryboardLastFrame();
});

// 监听提示词变化，自动保存到sessionStorage
watch(currentFramePrompt, (newPrompt) => {
  // 如果正在切换帧类型或分镜，不要保存（避免错误保存到新帧类型）
  if (isSwitchingFrameType.value) return;
  if (!currentStoryboard.value) return;

  const storageKey = getPromptStorageKey(
    currentStoryboard.value.id,
    selectedFrameType.value,
  );
  if (storageKey) {
    if (newPrompt) {
      sessionStorage.setItem(storageKey, newPrompt);
    } else {
      sessionStorage.removeItem(storageKey);
    }
  }
});

// 监听视频模型切换，清空已选图片并自动选择参考图模式
watch(selectedVideoModel, (newModel) => {
  if (newModel) {
    localStorage.setItem(`ai_video_model_${dramaId}`, newModel);
  }
  selectedImagesForVideo.value = [];
  selectedLastImageForVideo.value = null;
  // 自动选择最佳参考图模式：优先首尾帧 > 单图 > 多图 > 纯文本
  const modes = availableReferenceModes.value;
  const preferredOrder = ["first_last", "single", "multiple", "none"];
  const preferredMode = preferredOrder
    .map(pref => modes.find(m => m.value === pref))
    .find(m => m !== undefined);
  selectedReferenceMode.value = preferredMode?.value || modes[0]?.value || "";
});

// 监听图片模型切换，保存到localStorage
watch(selectedImageModel, (newModel) => {
  if (newModel) {
    localStorage.setItem(`ai_image_model_${dramaId}`, newModel);
  }
});

// 监听图文生图模型切换，保存到localStorage
watch(selectedImageToImageModel, (newModel) => {
  if (newModel) {
    localStorage.setItem(`ai_image_to_image_model_${dramaId}`, newModel);
  }
});

// 监听镜头切换，自动更新视频时长并重置音频/字幕开关
const isLoadingDuration = ref(false);
watch(currentStoryboard, (newStoryboard) => {
  isLoadingDuration.value = true;
  if (newStoryboard?.duration) {
    videoDuration.value = Math.round(newStoryboard.duration);
  } else {
    videoDuration.value = 5;
  }
  videoGenerateAudio.value = false;
  videoEnableSubtitle.value = false;
  nextTick(() => { isLoadingDuration.value = false; });
});

// 滑块调整时自动保存时长到数据库
watch(videoDuration, async (newDuration) => {
  if (isLoadingDuration.value || !currentStoryboard.value) return;
  try {
    await dramaAPI.updateStoryboard(
      currentStoryboard.value.id.toString(),
      { duration: newDuration },
    );
    currentStoryboard.value.duration = newDuration;
  } catch (error: any) {
  }
});

// 监听参考图模式切换，清空已选图片
watch(selectedReferenceMode, () => {
  selectedImagesForVideo.value = [];
  selectedLastImageForVideo.value = null;
});

// 当前分镜的角色列表
const currentStoryboardCharacters = computed(() => {
  if (!currentStoryboard.value?.characters) return [];

  const characterIds: number[] = [];

  if (
    Array.isArray(currentStoryboard.value.characters) &&
    currentStoryboard.value.characters.length > 0
  ) {
    const firstItem = currentStoryboard.value.characters[0];
    if (typeof firstItem === "object" && firstItem.id) {
      currentStoryboard.value.characters.forEach((char: any) => {
        if (char.id) characterIds.push(char.id);
      });
    } else if (typeof firstItem === "number") {
      characterIds.push(...currentStoryboard.value.characters);
    }
  }

  // Search in both top-level and children (flattened)
  const allChars: any[] = [];
  for (const char of (characters.value || [])) {
    allChars.push(char);
    if (char.children?.length) {
      allChars.push(...char.children);
    }
  }
  return allChars.filter((c) => characterIds.includes(c.id));
});

// 可选择的角色列表（展平：有子造型时显示子造型，否则显示角色本身）
const availableCharacters = computed(() => {
  const result: any[] = [];
  for (const char of (characters.value || [])) {
    if (char.children?.length) {
      for (const child of char.children) {
        result.push({ ...child, _parentName: char.name });
      }
    } else {
      result.push(char);
    }
  }
  return result;
});

// 可选择的道具列表
const availableProps = computed(() => {
  return props.value || [];
});

// 当前分镜的道具列表
const currentStoryboardProps = computed(() => {
  if (!currentStoryboard.value?.props) return [];
  return currentStoryboard.value.props;
});

// 检查道具是否在当前镜头中
const isPropInCurrentShot = (propId: number) => {
  if (!currentStoryboard.value?.props) return false;
  return currentStoryboard.value.props.some((p: any) => p.id === propId);
};

// 切换道具在镜头中的状态
const togglePropInShot = async (propId: number) => {
  if (!currentStoryboard.value) return;

  let newProps = [...(currentStoryboard.value.props || [])];
  if (isPropInCurrentShot(propId)) {
    newProps = newProps.filter((p: any) => p.id !== propId);
  } else {
    const prop = props.value.find((p) => p.id === propId);
    if (prop) {
      newProps.push(prop);
    }
  }

  // 乐观更新
  currentStoryboard.value.props = newProps;

  try {
    const propIds = newProps.map((p: any) => p.id);
    await propAPI.associateWithStoryboard(
      Number(currentStoryboard.value.id),
      propIds,
    );
  } catch (error) {
    ElMessage.error($t("editor.updatePropFailed"));
  }
};

// 检查角色是否已在当前镜头中
const isCharacterInCurrentShot = (charId: number) => {
  if (!currentStoryboard.value?.characters) return false;

  if (
    Array.isArray(currentStoryboard.value.characters) &&
    currentStoryboard.value.characters.length > 0
  ) {
    const firstItem = currentStoryboard.value.characters[0];
    if (typeof firstItem === "object" && firstItem.id) {
      return currentStoryboard.value.characters.some((c) => c.id === charId);
    }
    if (typeof firstItem === "number") {
      return currentStoryboard.value.characters.includes(charId);
    }
  }

  return false;
};

// 切换角色在镜头中的状态
const showCharacterImage = (char: any) => {
  previewCharacter.value = char;
  showCharacterImagePreview.value = true;
};

// 展示场景大图
const showSceneImage = () => {
  if (currentStoryboard.value?.background?.image_url) {
    showSceneImagePreview.value = true;
  }
};

// 描述编辑状态
const descEditing = ref(false);
const descSaving = ref(false);
const descBackup = ref<{ first: string; middle: string; last: string }>({ first: "", middle: "", last: "" });

watch(descEditing, (editing) => {
  if (editing && currentStoryboard.value) {
    descBackup.value = {
      first: currentStoryboard.value.first_frame_desc || "",
      middle: currentStoryboard.value.middle_action_desc || "",
      last: currentStoryboard.value.last_frame_desc || "",
    };
  }
});

const cancelDescEdit = () => {
  if (currentStoryboard.value) {
    currentStoryboard.value.first_frame_desc = descBackup.value.first;
    currentStoryboard.value.middle_action_desc = descBackup.value.middle;
    currentStoryboard.value.last_frame_desc = descBackup.value.last;
  }
  descEditing.value = false;
};

const saveAllDescFields = async () => {
  if (!currentStoryboard.value) return;
  descSaving.value = true;
  try {
    await dramaAPI.updateStoryboard(
      currentStoryboard.value.id.toString(),
      {
        first_frame_desc: currentStoryboard.value.first_frame_desc,
        middle_action_desc: currentStoryboard.value.middle_action_desc,
        last_frame_desc: currentStoryboard.value.last_frame_desc,
      },
    );
    ElMessage.success("描述已保存");
    descEditing.value = false;
  } catch (error: any) {
    ElMessage.error("保存失败: " + (error.message || "未知错误"));
  } finally {
    descSaving.value = false;
  }
};

// 保存分镜字段（单字段，其他功能使用）
const saveStoryboardField = async (fieldName: string) => {
  if (!currentStoryboard.value) return;
  try {
    const updateData: any = {};
    updateData[fieldName] = currentStoryboard.value[fieldName];

    await dramaAPI.updateStoryboard(
      currentStoryboard.value.id.toString(),
      updateData,
    );
  } catch (error: any) {
    ElMessage.error("保存失败: " + (error.message || "未知错误"));
  }
};

// 提取帧提示词
// 提取帧提示词
const extractFramePrompt = async () => {
  if (!currentStoryboard.value) return;

  const storyboardId = currentStoryboard.value.id;
  // 记录点击时的帧类型，后续任务完成时用于判断是否需要更新当前显示
  const targetFrameType = selectedFrameType.value;

  if (targetFrameType === "panel") {
    // 如果是分镜板模式，还需要捕获当前的panelCount
    // 注意：这里简单起见使用当前的panelCount，理想情况下应该传递参数或锁定UI
  }

  // 设置当前镜头的生成状态为true
  const stateKey = `${storyboardId}_${targetFrameType}`;
  generatingPromptStates.value[stateKey] = true;

  try {
    const params: any = {
      frame_type: targetFrameType,
      image_ratio: imageOrientation.value === "landscape" ? "16:9" : "9:16",
    };
    if (targetFrameType === "panel") {
      params.panel_count = panelCount.value;
    }

    const { task_id } = await generateFramePrompt(storyboardId, params);

    // 轮询任务状态（独立函数，不依赖组件当前状态）
    const pollTask = async () => {
      while (true) {
        const task = await taskAPI.getStatus(task_id);
        if (task.status === "completed") {
          let result = task.result;
          if (typeof result === "string") {
            try {
              result = JSON.parse(result);
            } catch (e) {
              console.error("Failed to parse task result", e);
              throw new Error("解析任务结果失败");
            }
          }
          return result.response;
        } else if (task.status === "failed") {
          throw new Error(task.message || task.error || "生成失败");
        }
        // 等待1秒后继续轮询
        await new Promise((resolve) => setTimeout(resolve, 1000));
      }
    };

    const result = await pollTask();

    // 根据返回结果构建提示词字符串
    let extractedPrompt = "";
    if (result.single_frame) {
      extractedPrompt = result.single_frame.prompt;
    } else if (result.multi_frame && result.multi_frame.frames) {
      // 多帧情况，将所有帧的prompt合并
      extractedPrompt = result.multi_frame.frames
        .map((frame: any) => frame.prompt)
        .join("\n\n");
    }

    // 更新存储（这一步必须做，无论用户是否还在当前页面）
    // 更新 session storage
    const storageKey = getPromptStorageKey(storyboardId, targetFrameType);
    if (storageKey) {
      sessionStorage.setItem(storageKey, extractedPrompt);
    }

    // 如果任务完成时，用户当前的选中状态正好是该镜头+该类型，则立即更新显示
    if (
      currentStoryboard.value &&
      currentStoryboard.value.id === storyboardId &&
      selectedFrameType.value === targetFrameType
    ) {
      currentFramePrompt.value = extractedPrompt;
      framePrompts.value[targetFrameType] = extractedPrompt;
    }

    // 更新内存缓存（稍微复杂点，framePrompts 是响应式的且绑定当前镜头，这里只做sessionStorage持久化即可，
    // 因为切换镜头时会重新读取sessionStorage。
    // 但为了确保如果用户没切走也能看到，上面已经更新了 currentFramePrompt

    ElMessage.success(`${getFrameTypeLabel(targetFrameType)}提示词提取成功`);
  } catch (error: any) {
    ElMessage.error("提取失败: " + (error.message || "未知错误"));
  } finally {
    // 清除该镜头的生成状态
    const stateKey = `${storyboardId}_${targetFrameType}`;
    if (generatingPromptStates.value[stateKey]) {
      generatingPromptStates.value[stateKey] = false;
    }
  }
};

// 检查是否正在生成提示词
const isGeneratingPrompt = (
  storyboardId: number | undefined,
  frameType: string,
) => {
  if (!storyboardId) return false;
  return !!generatingPromptStates.value[`${storyboardId}_${frameType}`];
};

// 获取帧类型的中文标签
const getFrameTypeLabel = (frameType: string): string => {
  const labels: Record<string, string> = {
    key: "关键帧",
    first: "首帧",
    last: "尾帧",
    panel: "分镜版",
    action: "动作序列",
  };
  return labels[frameType] || frameType;
};

// 批量提取全部分镜的提示词
const batchExtractAllPrompts = async () => {
  if (!episodeId.value) {
    ElMessage.warning("未找到章节信息");
    return;
  }

  const frameType = selectedFrameType.value;

  try {
    await ElMessageBox.confirm(
      `将为当前章节的所有分镜批量AI提取「${getFrameTypeLabel(frameType)}」提示词，已有的提示词会被覆盖。是否继续？`,
      "批量AI提取提示词",
      { confirmButtonText: "开始", cancelButtonText: "取消", type: "warning" }
    );
  } catch {
    return; // 用户取消
  }

  batchExtractingPrompts.value = true;
  batchExtractProgress.value = "任务创建中...";

  try {
    const { task_id } = await batchGenerateFramePrompts(episodeId.value, {
      frame_type: frameType,
    });

    // 轮询任务状态
    const pollTask = async () => {
      while (true) {
        const task = await taskAPI.getStatus(task_id);
        if (task.status === "completed") {
          let result = task.result;
          if (typeof result === "string") {
            try { result = JSON.parse(result); } catch { /* ignore */ }
          }
          const total = result?.total || 0;
          const completed = result?.completed || 0;
          const failed = result?.failed || 0;
          ElMessage.success(`批量提取完成！共 ${total} 个分镜，成功 ${completed} 个${failed > 0 ? `，失败 ${failed} 个` : ""}`);

          // 并行刷新所有分镜的提示词到 sessionStorage
          await Promise.allSettled(
            storyboards.value.map(async (sb) => {
              const prompts = await getStoryboardFramePrompts(sb.id);
              if (prompts?.frame_prompts) {
                const matched = prompts.frame_prompts.find((p: any) => p.frame_type === frameType);
                if (matched) {
                  const storageKey = getPromptStorageKey(sb.id, frameType);
                  if (storageKey) {
                    sessionStorage.setItem(storageKey, matched.prompt);
                  }
                  if (currentStoryboard.value && currentStoryboard.value.id === sb.id && selectedFrameType.value === frameType) {
                    currentFramePrompt.value = matched.prompt;
                    framePrompts.value[frameType] = matched.prompt;
                  }
                }
              }
            })
          );
          break;
        } else if (task.status === "failed") {
          ElMessage.error("批量提取失败: " + (task.message || "未知错误"));
          break;
        } else {
          // processing
          batchExtractProgress.value = task.message || "处理中...";
          await new Promise((resolve) => setTimeout(resolve, 2000));
        }
      }
    };

    await pollTask();
  } catch (error: any) {
    ElMessage.error("批量提取失败: " + (error.message || "未知错误"));
  } finally {
    batchExtractingPrompts.value = false;
    batchExtractProgress.value = "";
  }
};

// 加载分镜的图片列表
const loadStoryboardImages = async (
  storyboardId: string | number,
  frameType?: string,
) => {
  loadingImages.value = true;
  try {
    const params: any = {
      storyboard_id: storyboardId,
      page: 1,
      page_size: 50,
    };
    // 如果指定了帧类型，添加过滤
    if (frameType) {
      params.frame_type = frameType;
    }
    const result = await imageAPI.listImages(params);
    generatedImages.value = result.items || [];

    // 同时加载所有帧类型的图片（用于视频输入选择）
    await loadAllStoryboardImages(storyboardId);

    // 如果有进行中的任务，启动轮询
    const hasPendingOrProcessing = generatedImages.value.some(
      (img) => img.status === "pending" || img.status === "processing",
    );
    if (hasPendingOrProcessing) {
      startPolling();
    }
  } catch (error: any) {
    console.error("加载图片列表失败:", error);
  } finally {
    loadingImages.value = false;
  }
};

// 加载分镜的所有图片（所有帧类型）
const loadAllStoryboardImages = async (storyboardId: string | number) => {
  try {
    const params: any = {
      storyboard_id: storyboardId,
      page: 1,
      page_size: 100,
    };
    const result = await imageAPI.listImages(params);
    allGeneratedImages.value = result.items || [];
  } catch (error: any) {
    console.error("加载所有图片列表失败:", error);
  }
};

// 切换参考资源的勾选状态
const toggleRefChar = (charId: number) => {
  const s = new Set(selectedRefCharIds.value);
  if (s.has(charId)) s.delete(charId); else s.add(charId);
  selectedRefCharIds.value = s;
};
const toggleRefProp = (propId: number) => {
  const s = new Set(selectedRefPropIds.value);
  if (s.has(propId)) s.delete(propId); else s.add(propId);
  selectedRefPropIds.value = s;
};

// 加载当前分镜的首帧图片（供尾帧选择参考）
const loadFirstFrameImagesForLast = async (storyboardId: string | number) => {
  try {
    const result = await imageAPI.listImages({
      storyboard_id: storyboardId,
      frame_type: "first",
      page: 1,
      page_size: 20,
    } as any);
    const completed = (result.items || []).filter(
      (img: any) => img.status === "completed" && (img.local_path || img.image_url)
    );
    firstFrameImagesForLast.value = completed;
    if (completed.length > 0 && !selectedFirstFrameId.value) {
      selectedFirstFrameId.value = completed[0].id;
    }
  } catch (e) {
    firstFrameImagesForLast.value = [];
  }
};

// 启动状态轮询
const startPolling = () => {
  if (pollingTimer) return;

  // 记录开始轮询时的帧类型
  pollingFrameType = selectedFrameType.value;

  pollingTimer = setInterval(async () => {
    if (!currentStoryboard.value) {
      stopPolling();
      return;
    }

    // 如果帧类型已切换，停止轮询（防止更新到错误的帧类型）
    if (selectedFrameType.value !== pollingFrameType) {
      stopPolling();
      return;
    }

    try {
      const params: any = {
        storyboard_id: currentStoryboard.value.id,
        page: 1,
        page_size: 50,
      };
      // 使用轮询开始时记录的帧类型
      if (pollingFrameType) {
        params.frame_type = pollingFrameType;
      }
      const result = await imageAPI.listImages(params);

      // 再次检查帧类型是否仍然匹配，避免竞态条件
      if (selectedFrameType.value === pollingFrameType) {
        generatedImages.value = result.items || [];
      }

      // 如果没有进行中的任务，停止轮询并刷新视频参考图片
      const hasPendingOrProcessing = (result.items || []).some(
        (img: any) => img.status === "pending" || img.status === "processing",
      );
      if (!hasPendingOrProcessing) {
        stopPolling();
        // 刷新视频参考图片列表
        if (currentStoryboard.value) {
          loadVideoReferenceImages(currentStoryboard.value.id);
        }
      }
    } catch (error) {
      console.error("轮询图片状态失败:", error);
    }
  }, 3000); // 每3秒轮询一次
};

// 停止轮询
const stopPolling = () => {
  if (pollingTimer) {
    clearInterval(pollingTimer);
    pollingTimer = null;
  }
  pollingFrameType = null;
};

// 生成图片 - 弹出确认弹窗
const generateFrameImage = async () => {
  if (!currentStoryboard.value || !currentFramePrompt.value) return;

  // 收集参考图片信息（仅用户勾选的资源）
  const refImages: { name: string; path: string }[] = [];

  // 1. 场景图片（仅勾选时）
  if (selectedRefScene.value && currentStoryboard.value.background?.local_path) {
    const bgName = currentStoryboard.value.background?.location || "场景背景";
    refImages.push({ name: `🏠 ${bgName}`, path: currentStoryboard.value.background.local_path });
  }

  // 2. 角色图片（仅勾选的角色）
  const storyboardCharacters = currentStoryboardCharacters.value;
  if (storyboardCharacters && storyboardCharacters.length > 0) {
    storyboardCharacters.forEach((char: any) => {
      if (char.local_path && selectedRefCharIds.value.has(char.id)) {
        refImages.push({ name: `👤 ${char.name || "角色"}`, path: char.local_path });
      }
    });
  }

  // 3. 道具图片（仅勾选的道具）
  const storyboardProps = currentStoryboardProps.value;
  if (storyboardProps && storyboardProps.length > 0) {
    storyboardProps.forEach((prop: any) => {
      if (prop.local_path && selectedRefPropIds.value.has(prop.id)) {
        refImages.push({ name: `🔧 ${prop.name || "道具"}`, path: prop.local_path });
      }
    });
  }

  // 4. 如果是尾帧，使用用户选择的首帧图片
  if (selectedFrameType.value === "last" && selectedFirstFrameId.value) {
    const selectedImg = firstFrameImagesForLast.value.find(
      (img: any) => img.id === selectedFirstFrameId.value
    );
    if (selectedImg) {
      refImages.push({
        name: "🖼️ 首帧参考图",
        path: selectedImg.local_path || selectedImg.image_url,
      });
    }
  }

  // 5. 如果是首帧 + 勾选了参考帧
  if (selectedFrameType.value === "first" && selectedRefPrevFrame.value && refFrameState.framePath) {
    const refSb = storyboards.value.find((s) => Number(s.id) === refFrameState.selectedStoryboardId);
    refImages.push({
      name: `🎬 镜头 #${refSb?.storyboard_number || '?'} ${refFrameState.sourceType === 'first' ? '首帧图' : refFrameState.sourceType === 'last' ? '尾帧图' : '视频尾帧'}`,
      path: refFrameState.framePath,
    });
  }

  // 获取当前选择的图片模型名称
  const modelDisplay = selectedImageModel.value || "使用默认配置模型";
  
  // 统一使用图文生图模型
  const modelForGeneration = selectedImageToImageModel.value;

  // 提示词直接使用 AI 生成的内容，参考图片描述由后端 buildReferenceImageDescriptions 统一拼接
  let promptWithRefs = currentFramePrompt.value;

  imageGenDialog.orientation = imageOrientation.value;

  // 初始化弹窗
  imageGenDialog.visible = true;
  imageGenDialog.phase = "confirm";
  imageGenDialog.originalPrompt = promptWithRefs;
  imageGenDialog.prompt = getPromptWithOrientation(promptWithRefs, imageOrientation.value);
  imageGenDialog.frameType = selectedFrameType.value;
  imageGenDialog.referenceImages = refImages;
  imageGenDialog.model = modelForGeneration;
  imageGenDialog.imageGenId = null;
  imageGenDialog.progress = 0;
  imageGenDialog.statusText = "";
  imageGenDialog.error = "";
  imageGenDialog.aborted = false;
};

// Debug: 调用后端 preview 接口，展示 curl 请求 + 真正会发给模型的完整 prompt
const debugGenerateImage = async () => {
  if (!currentStoryboard.value) return;
  const referenceImagePaths = imageGenDialog.referenceImages.map((r) => r.path);
  const model = selectedImageToImageModel.value;
  const apiBase = window.location.origin + '/api/v1';

  const reqBody = {
    drama_id: dramaId.toString(),
    prompt: imageGenDialog.prompt,
    storyboard_id: currentStoryboard.value.id,
    image_type: "storyboard",
    frame_type: imageGenDialog.frameType,
    reference_images: referenceImagePaths.length > 0 ? referenceImagePaths : undefined,
    width: imageGenDialog.width,
    height: imageGenDialog.height,
    model,
  };

  try {
    const result = await imageAPI.previewImagePrompt(reqBody as any);
    const finalPrompt = result.final_prompt || '(无)';

    // curl 里的 prompt 用后端最终版本，这样复制出去就是真正发给模型的
    const curlReqBody = { ...reqBody, prompt: finalPrompt };
    const curlBody = JSON.stringify(curlReqBody, null, 2);
    const curlCmd = `curl -X POST '${apiBase}/images' \\\n  -H 'Content-Type: application/json' \\\n  -H 'Authorization: Bearer <YOUR_TOKEN>' \\\n  -d '${curlBody.replace(/'/g, "'\\''")}'`;

    imageDebugDialog.curlCommand =
      `=== 最终发送给模型的完整请求 ===\n${curlCmd}`;
  } catch (e: any) {
    const curlBody = JSON.stringify(reqBody, null, 2);
    const curlCmd = `curl -X POST '${apiBase}/images' \\\n  -H 'Content-Type: application/json' \\\n  -H 'Authorization: Bearer <YOUR_TOKEN>' \\\n  -d '${curlBody.replace(/'/g, "'\\''")}'`;
    imageDebugDialog.curlCommand = `=== API 请求（原始，preview 失败） ===\n${curlCmd}\n\n=== 错误 ===\n${e.message || e}`;
  }

  imageDebugDialog.visible = true;
};

const copyImageDebugCommand = async () => {
  try {
    await navigator.clipboard.writeText(imageDebugDialog.curlCommand);
    ElMessage.success("已复制到剪贴板");
  } catch {
    const textarea = document.createElement('textarea');
    textarea.value = imageDebugDialog.curlCommand;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand('copy');
    document.body.removeChild(textarea);
    ElMessage.success("已复制到剪贴板");
  }
};

// 确认并执行图片生成
const confirmGenerateImage = async () => {
  if (!currentStoryboard.value) return;
  const genStoryboardId = currentStoryboard.value.id;

  imageGenDialog.phase = "generating";
  imageGenDialog.progress = 10;
  imageGenDialog.statusText = "正在提交生成任务...";
  generatingImageIds.value.add(genStoryboardId);

  try {
    const referenceImagePaths = imageGenDialog.referenceImages.map((r) => r.path);

    // 统一使用图文生图模型
    const model = selectedImageToImageModel.value;

    const result = await imageAPI.generateImage({
      drama_id: dramaId.toString(),
      prompt: imageGenDialog.prompt,
      storyboard_id: currentStoryboard.value.id,
      image_type: "storyboard",
      frame_type: imageGenDialog.frameType,
      reference_images: referenceImagePaths.length > 0 ? referenceImagePaths : undefined,
      width: imageGenDialog.width,
      height: imageGenDialog.height,
      model,
    });

    imageGenDialog.imageGenId = result.id || result.ID;
    imageGenDialog.progress = 30;
    imageGenDialog.statusText = "任务已提交，等待生成中...";

    generatedImages.value.unshift(result);

    // 轮询等待这张图片完成
    const maxWaitMs = 5 * 60 * 1000;
    const startTime = Date.now();
    const interval = 3000;

    while (Date.now() - startTime < maxWaitMs) {
      if (imageGenDialog.aborted) {
        imageGenDialog.statusText = "已终止";
        imageGenDialog.phase = "done";
        break;
      }

      await new Promise((resolve) => setTimeout(resolve, interval));

      try {
        const img = await imageAPI.getImage(imageGenDialog.imageGenId!);

        if (img.status === "completed") {
          imageGenDialog.progress = 100;
          imageGenDialog.statusText = "生成完成！";
          imageGenDialog.phase = "done";

          // 更新列表中的图片
          const idx = generatedImages.value.findIndex((i) => i.id === imageGenDialog.imageGenId);
          if (idx >= 0) {
            generatedImages.value[idx] = img;
          }

          // 更新所有图片列表
          const allIdx = allGeneratedImages.value.findIndex((i) => i.id === imageGenDialog.imageGenId);
          if (allIdx >= 0) {
            allGeneratedImages.value[allIdx] = img;
          } else {
            allGeneratedImages.value.unshift(img);
          }

          // 刷新视频参考图片
          if (currentStoryboard.value) {
            loadVideoReferenceImages(currentStoryboard.value.id);
          }

          // 2秒后自动关闭弹窗
          setTimeout(() => {
            if (imageGenDialog.phase === "done" && !imageGenDialog.error) {
              imageGenDialog.visible = false;
            }
          }, 2000);
          break;
        } else if (img.status === "failed") {
          imageGenDialog.progress = 100;
          imageGenDialog.error = img.error_msg || "生成失败";
          imageGenDialog.statusText = "生成失败";
          imageGenDialog.phase = "done";

          const idx = generatedImages.value.findIndex((i) => i.id === imageGenDialog.imageGenId);
          if (idx >= 0) {
            generatedImages.value[idx] = img;
          }

          // 更新所有图片列表
          const allIdx = allGeneratedImages.value.findIndex((i) => i.id === imageGenDialog.imageGenId);
          if (allIdx >= 0) {
            allGeneratedImages.value[allIdx] = img;
          } else {
            allGeneratedImages.value.unshift(img);
          }
          break;
        } else {
          // 处理中 - 逐步增加进度
          const elapsed = Date.now() - startTime;
          imageGenDialog.progress = Math.min(30 + Math.round((elapsed / maxWaitMs) * 60), 90);
          imageGenDialog.statusText = "AI 正在绘制图片...";
        }
      } catch {
        // 查询失败，继续轮询
      }
    }

    // 超时
    if (imageGenDialog.phase === "generating") {
      imageGenDialog.statusText = "生成时间较长，已在后台继续处理";
      imageGenDialog.phase = "done";
      startPolling();
    }
  } catch (error: any) {
    imageGenDialog.error = error.message || "未知错误";
    imageGenDialog.statusText = "提交失败";
    imageGenDialog.phase = "done";
  } finally {
    generatingImageIds.value.delete(genStoryboardId);
  }
};

// 终止图片生成
const abortImageGeneration = () => {
  imageGenDialog.aborted = true;
  imageGenDialog.statusText = "正在终止...";
  // 轮询停止后会自动更新状态
  startPolling(); // 让后台轮询继续跟踪
};

// 处理图片生成弹窗关闭
const handleImageGenDialogClose = (done: () => void) => {
  if (imageGenDialog.phase === 'generating') {
    // 生成过程中关闭，只关闭弹窗，后台继续运行
    ElMessage.info('图片生成继续在后台进行，您可以随时查看进度');
    done();
  } else {
    // 其他阶段直接关闭
    done();
  }
};

// 获取状态标签类型
const getStatusType = (status: string) => {
  const statusMap: Record<string, any> = {
    pending: "info",
    processing: "warning",
    completed: "success",
    failed: "danger",
  };
  return statusMap[status] || "info";
};

// 播放视频
const playVideo = (video: VideoGeneration) => {
  previewVideo.value = video;
  showVideoPreview.value = true;
};

// 添加视频到素材库
const addVideoToAssets = async (video: VideoGeneration) => {
  if (video.status !== "completed" || !video.video_url) {
    ElMessage.warning("只能添加已完成的视频到素材库");
    return;
  }

  addingToAssets.value.add(video.id);

  try {
    // 检查该镜头是否已存在素材
    let isReplacing = false;
    if (video.storyboard_id) {
      const existingAsset = videoAssets.value.find(
        (asset: any) => asset.storyboard_id === video.storyboard_id,
      );

      if (existingAsset) {
        isReplacing = true;
        // 自动替换：先删除旧素材
        try {
          await assetAPI.deleteAsset(existingAsset.id);
        } catch (error) {
          console.error("删除旧素材失败:", error);
        }
      }
    }

    // 添加新素材
    await assetAPI.importFromVideo(video.id);
    ElMessage.success("已添加到素材库");

    // 重新加载素材库列表
    await loadVideoAssets();

    // 如果是替换操作，更新时间线中使用该分镜的所有视频片段
    if (isReplacing && video.storyboard_id && video.video_url) {
      if (timelineEditorRef.value) {
        timelineEditorRef.value.updateClipsByStoryboardId(
          video.storyboard_id,
          video.video_url,
        );
      }
    }
  } catch (error: any) {
    ElMessage.error(error.message || "添加失败");
  } finally {
    addingToAssets.value.delete(video.id);
  }
};

// 删除视频
const handleDeleteVideo = async (video: VideoGeneration) => {
  if (!currentStoryboard.value) return;

  try {
    await ElMessageBox.confirm(
      "确定要删除这个视频吗？删除后无法恢复。",
      "确认删除",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      },
    );

    await videoAPI.deleteVideo(video.id);
    ElMessage.success("删除成功");

    // 重新加载当前镜头的视频列表
    await loadStoryboardVideos(Number(currentStoryboard.value.id));
  } catch (error: any) {
    if (error !== "cancel") {
      console.error("删除视频失败:", error);
      ElMessage.error(error.message || "删除失败");
    }
  }
};

// 获取状态中文文本
const getStatusText = (status: string) => {
  const statusTextMap: Record<string, string> = {
    pending: "等待中",
    processing: "生成中",
    completed: "已完成",
    failed: "失败",
  };
  return statusTextMap[status] || status;
};

// 获取帧类型中文文本
const getFrameTypeText = (frameType?: string) => {
  if (!frameType) return "";
  const frameTypeMap: Record<string, string> = {
    first: "首帧",
    key: "关键帧",
    last: "尾帧",
    panel: "分镜板",
    action: "动作序列",
  };
  return frameTypeMap[frameType] || frameType;
};

// 获取分镜缩略图
const getStoryboardThumbnail = (storyboard: any) => {
  // 优先使用composed_image
  if (storyboard.composed_image) {
    return storyboard.composed_image;
  }

  // 如果没有composed_image，从image_url字段获取
  if (storyboard.image_url) {
    return storyboard.image_url;
  }

  return null;
};

// 处理图片选择（根据模型能力）
const handleImageSelect = (imageId: number) => {
  if (!selectedReferenceMode.value) {
    ElMessage.warning("请先选择参考图模式");
    return;
  }

  if (!currentModelCapability.value) {
    ElMessage.warning("请先选择视频生成模型");
    return;
  }

  const capability = currentModelCapability.value;
  const currentIndex = selectedImagesForVideo.value.indexOf(imageId);

  // 已选中，则取消选择
  if (currentIndex > -1) {
    selectedImagesForVideo.value.splice(currentIndex, 1);
    return;
  }

  // 获取当前点击的图片对象
  const clickedImage = videoReferenceImages.value.find(
    (img) => img.id === imageId,
  );
  if (!clickedImage) return;

  // 根据选择的参考图模式处理
  switch (selectedReferenceMode.value) {
    case "single":
      // 单图模式：只能选1张，直接替换
      selectedImagesForVideo.value = [imageId];
      break;

    case "first_last":
      // 首尾帧模式：根据图片类型分别处理
      const frameType = clickedImage.frame_type;

      if (
        frameType === "first" ||
        frameType === "panel" ||
        frameType === "key"
      ) {
        // 首帧：直接替换
        selectedImagesForVideo.value = [imageId];
      } else if (frameType === "last") {
        // 尾帧：设置到单独的变量
        selectedLastImageForVideo.value = imageId;
      } else {
        ElMessage.warning("首尾帧模式下，请选择首帧或尾帧类型的图片");
      }
      break;

    case "multiple":
      // 多图模式：检查是否超出最大数量
      if (selectedImagesForVideo.value.length >= capability.maxImages) {
        ElMessage.warning(`最多只能选择${capability.maxImages}张图片`);
        return;
      }
      selectedImagesForVideo.value.push(imageId);
      break;

    default:
      ElMessage.warning("未知的参考图模式");
  }
};

// 预览图片（使用已导入的 getImageUrl 工具函数来获取正确的图片URL）
const previewImage = (url: string) => {
  // 使用Element Plus的图片预览
  const viewer = document.createElement("div");
  viewer.innerHTML = `
    <div style="position: fixed; top: 0; left: 0; right: 0; bottom: 0; z-index: 9999; background: rgba(0,0,0,0.8); display: flex; align-items: center; justify-content: center;" onclick="this.remove()">
      <img src="${url}" style="max-width: 90vw; max-height: 90vh; object-fit: contain;" onclick="event.stopPropagation();" />
    </div>
  `;
  document.body.appendChild(viewer);
};

// 获取已选图片对象列表
const selectedImageObjects = computed(() => {
  return selectedImagesForVideo.value
    .map((id) => videoReferenceImages.value.find((img) => img.id === id))
    .filter((img) => img && img.image_url);
});

// 首尾帧模式：获取首帧图片
const firstFrameSlotImage = computed(() => {
  if (selectedImagesForVideo.value.length === 0) return null;
  const firstImageId = selectedImagesForVideo.value[0];
  // 同时搜索当前镜头图片和上一镜头尾帧
  return (
    videoReferenceImages.value.find((img) => img.id === firstImageId) ||
    previousStoryboardLastFrames.value.find((img) => img.id === firstImageId)
  );
});

// 首尾帧模式：获取尾帧图片
const lastFrameSlotImage = computed(() => {
  if (!selectedLastImageForVideo.value) return null;
  // 同时搜索当前镜头图片和上一镜头尾帧
  return (
    videoReferenceImages.value.find(
      (img) => img.id === selectedLastImageForVideo.value,
    ) ||
    previousStoryboardLastFrames.value.find(
      (img) => img.id === selectedLastImageForVideo.value,
    )
  );
});

// 移除已选择的图片
const removeSelectedImage = (imageId: number) => {
  // 检查是否是尾帧
  if (selectedLastImageForVideo.value === imageId) {
    selectedLastImageForVideo.value = null;
    return;
  }

  // 检查是否是首帧或其他图片
  const index = selectedImagesForVideo.value.indexOf(imageId);
  if (index > -1) {
    selectedImagesForVideo.value.splice(index, 1);
  }
};

// 检查提示词是否包含参考图片描述
const checkPromptHasReference = (prompt: string): boolean => {
  if (!prompt) return false;

  // 检查是否包含参考图片相关的关键词
  const referenceKeywords = [
    '参考图片',
    '参考图',
    '参考图片说明',
    '传入了一张',
    '首帧图片',
    '输入图片',
    '参考图像',
  ];

  return referenceKeywords.some(keyword => prompt.includes(keyword));
};

// 保存视频提示词到后端
const saveVideoPrompt = async () => {
  if (!currentStoryboard.value) return;
  const newPrompt = videoPromptText.value.trim();
  if (newPrompt === (currentStoryboard.value.video_prompt || "")) return;
  try {
    await dramaAPI.updateStoryboard(String(currentStoryboard.value.id), { video_prompt: newPrompt });
    currentStoryboard.value.video_prompt = newPrompt;

    // 更新标记：提示词是否包含参考图片描述
    videoPromptHasReference.value = checkPromptHasReference(newPrompt);
  } catch (e: any) {
  }
};

// 监听视频提示词输入，实时更新标记
const onVideoPromptInput = () => {
  videoPromptHasReference.value = checkPromptHasReference(videoPromptText.value);
};

// V3: 当前镜头的输入图片（用于视频生成）
// 来源优先级：1. 手动选择的图片  2. 当前镜头已生成的首帧图片  3. 上一镜头视频尾帧截取
const currentShotImageForVideo = computed(() => {
  // 1. 首先检查手动选择的图片
  if (manuallySelectedVideoImage.value) {
    return manuallySelectedVideoImage.value;
  }

  // 2. 检查当前镜头的已生成首帧图片
  const firstFrameImg = allGeneratedImages.value.find(
    (img) => img.status === "completed" && img.image_url && img.frame_type === "first"
  );
  if (firstFrameImg) {
    return {
      url: getImageUrl(firstFrameImg) || "",
      path: firstFrameImg.local_path || firstFrameImg.image_url || "",
      localPath: firstFrameImg.local_path || "",
      imageUrl: firstFrameImg.image_url || "",
      source: `当前镜头首帧图片 (ID:${firstFrameImg.id})`,
      imageGenId: firstFrameImg.id,
    };
  }

  // 3. 检查已生成的任何类型图片
  const anyImg = allGeneratedImages.value.find(
    (img) => img.status === "completed" && img.image_url
  );
  if (anyImg) {
    return {
      url: getImageUrl(anyImg) || "",
      path: anyImg.local_path || anyImg.image_url || "",
      localPath: anyImg.local_path || "",
      imageUrl: anyImg.image_url || "",
      source: `当前镜头图片 (${anyImg.frame_type || "unknown"}, ID:${anyImg.id})`,
      imageGenId: anyImg.id,
    };
  }

  // 4. 检查参考镜头帧（首帧图/尾帧图/视频尾帧截取）
  if (refFrameState.framePath) {
    const refSb = storyboards.value.find((s) => Number(s.id) === refFrameState.selectedStoryboardId);
    const srcLabel = refFrameState.sourceType === 'first' ? '首帧图' : refFrameState.sourceType === 'last' ? '尾帧图' : '视频尾帧截取';
    return {
      url: `/static/${refFrameState.framePath}`,
      path: refFrameState.framePath,
      localPath: refFrameState.framePath,
      imageUrl: "",
      source: `镜头 #${refSb?.storyboard_number || '?'} ${srcLabel}`,
      imageGenId: null,
    };
  }

  return null;
});

// 可用的视频输入图片列表
const availableVideoInputImages = computed(() => {
  const options: { label: string; value: any }[] = [];

  // 1. 添加当前镜头的所有已生成图片（所有帧类型）
  const completedImages = allGeneratedImages.value.filter(img => img.status === "completed" && img.image_url);
  completedImages.forEach(img => {
    const frameTypeLabel = img.frame_type === "first" ? "首帧" : 
                          img.frame_type === "last" ? "尾帧" : 
                          img.frame_type === "key" ? "关键帧" : 
                          img.frame_type === "panel" ? "分镜板" : 
                          img.frame_type === "action" ? "动作序列" : img.frame_type || "未知";
    options.push({
      label: `当前镜头${frameTypeLabel} (ID:${img.id})`,
      value: {
        url: getImageUrl(img) || "",
        path: img.local_path || img.image_url || "",
        localPath: img.local_path || "",
        imageUrl: img.image_url || "",
        source: `当前镜头${frameTypeLabel} (ID:${img.id})`,
        imageGenId: img.id,
      }
    });
  });

  // 2. 添加参考镜头帧（首帧图/尾帧图/视频尾帧截取）
  if (refFrameState.framePath) {
    const refSb = storyboards.value.find((s) => Number(s.id) === refFrameState.selectedStoryboardId);
    const srcLabel = refFrameState.sourceType === 'first' ? '首帧图' : refFrameState.sourceType === 'last' ? '尾帧图' : '视频尾帧截取';
    const refLabel = `镜头 #${refSb?.storyboard_number || '?'} ${srcLabel}`;
    options.push({
      label: refLabel,
      value: {
        url: `/static/${refFrameState.framePath}`,
        path: refFrameState.framePath,
        localPath: refFrameState.framePath,
        imageUrl: "",
        source: refLabel,
        imageGenId: null,
      }
    });
  }

  return options;
});

// 视频生成：首帧图片列表
const videoFirstFrameImages = computed(() => {
  return allGeneratedImages.value.filter(
    img => img.status === "completed" && img.image_url && img.frame_type === "first"
  );
});

// 视频生成：尾帧图片列表
const videoLastFrameImages = computed(() => {
  return allGeneratedImages.value.filter(
    img => img.status === "completed" && img.image_url && img.frame_type === "last"
  );
});

// 视频生成：选中的首帧图片对象
const selectedVideoFirstFrame = computed(() => {
  if (!videoFirstFrameId.value) return videoFirstFrameImages.value[0] || null;
  return videoFirstFrameImages.value.find(img => img.id === videoFirstFrameId.value) || null;
});

// 视频生成：选中的尾帧图片对象（默认选最新的）
const selectedVideoLastFrame = computed(() => {
  if (!videoLastFrameId.value) return videoLastFrameImages.value[0] || null;
  return videoLastFrameImages.value.find(img => img.id === videoLastFrameId.value) || null;
});

// 处理视频输入图片选择变化
const onVideoImageSelect = (value: any) => {
  if (!value) {
    manuallySelectedVideoImage.value = null;
  }
};

// V3: 自动从分镜信息生成视频提示词（使用AI生成详细格式）
const autoGenerateVideoPrompt = async () => {
  if (!currentStoryboard.value) {
    ElMessage.warning("请先选择分镜");
    return;
  }

  try {
    ElMessage.info("正在生成视频提示词...");

    const response = await dramaAPI.generateVideoPrompt(currentStoryboard.value.id, undefined, videoDuration.value, videoEnableSubtitle.value, videoGenerateAudio.value, videoAspectRatio.value, videoIncludeDialogue.value);
    const prompt = response.video_prompt;

    if (!prompt) {
      ElMessage.error("生成视频提示词失败：返回结果为空");
      return;
    }

    videoPromptText.value = prompt;

    const shotImage = currentShotImageForVideo.value;
    videoPromptHasReference.value = !!shotImage;

    saveVideoPrompt();

    ElMessage.success("已使用AI生成视频提示词");
  } catch (error) {
    console.error("生成视频提示词失败:", error);
    ElMessage.error("生成视频提示词失败：" + (error as Error).message);
  }
};

// V3: 简化的视频生成（支持首帧+尾帧双图模式）
const generateVideoSimple = async () => {
  if (!selectedVideoModel.value) {
    ElMessage.warning("请先选择视频生成模型");
    return;
  }
  if (!currentStoryboard.value) {
    ElMessage.warning("请先选择分镜");
    return;
  }

  const firstFrame = selectedVideoFirstFrame.value;
  const lastFrame = selectedVideoLastFrame.value;

  const provider = extractProviderFromModel(selectedVideoModel.value);
  const prompt =
    videoPromptText.value ||
    currentStoryboard.value.video_prompt ||
    currentStoryboard.value.action ||
    currentStoryboard.value.description ||
    "";

  if (!prompt.trim()) {
    ElMessage.warning("请先生成或填写视频提示词");
    return;
  }

  const hasFirstFrame = !!firstFrame;
  const hasLastFrame = !!lastFrame;

  const requestParams: any = {
    drama_id: dramaId.toString(),
    storyboard_id: currentStoryboard.value.id,
    prompt,
    duration: videoDuration.value,
    provider,
    model: selectedVideoModel.value,
    resolution: videoResolution.value,
    generate_audio: videoGenerateAudio.value,
    enable_subtitle: videoEnableSubtitle.value,
    aspect_ratio: videoAspectRatio.value,
  };

  // 参考图展示信息
  const refImages: { name: string; path: string }[] = [];

  if (hasFirstFrame && hasLastFrame) {
    requestParams.reference_mode = "first_last";
    requestParams.first_frame_local_path = firstFrame!.local_path || undefined;
    requestParams.first_frame_url = firstFrame!.image_url || undefined;
    requestParams.last_frame_local_path = lastFrame!.local_path || undefined;
    requestParams.last_frame_url = lastFrame!.image_url || undefined;
    refImages.push({ name: "🎬 首帧", path: firstFrame!.local_path || firstFrame!.image_url || "" });
    refImages.push({ name: "🎬 尾帧", path: lastFrame!.local_path || lastFrame!.image_url || "" });
  } else if (hasFirstFrame) {
    requestParams.reference_mode = "single";
    if (firstFrame!.local_path) {
      requestParams.image_local_path = firstFrame!.local_path;
    } else if (firstFrame!.image_url) {
      requestParams.image_url = firstFrame!.image_url;
    }
    requestParams.image_gen_id = firstFrame!.id;
    refImages.push({ name: "🎬 首帧", path: firstFrame!.local_path || firstFrame!.image_url || "" });
  } else {
    requestParams.reference_mode = "text";
  }

  // 获取模型价格
  const modelCap = currentModelCapability.value;
  const estimatedPrice = calculateEstimatedPrice(0, 0, selectedVideoModel.value, videoGenerateAudio.value, videoDuration.value);

  // 参考图模式显示文字
  const modeText = hasFirstFrame && hasLastFrame
    ? "首帧 + 尾帧"
    : hasFirstFrame
      ? "仅首帧"
      : "纯文本生成";

  // 初始化弹窗
  videoGenDialog.visible = true;
  videoGenDialog.phase = "confirm";
  videoGenDialog.prompt = prompt;
  videoGenDialog.model = selectedVideoModel.value;
  videoGenDialog.pricing = modelCap?.pricing || "-";
  videoGenDialog.estimatedPrice = estimatedPrice;
  videoGenDialog.referenceMode = modeText;
  videoGenDialog.referenceImages = refImages;
  videoGenDialog.videoGenId = null;
  videoGenDialog.progress = 0;
  videoGenDialog.statusText = "";
  videoGenDialog.error = "";
  videoGenDialog.curlCommand = "";
  videoGenDialog.aborted = false;
  videoGenDialog.requestParams = requestParams;
};

// 生成视频 - 弹出确认弹窗（保留旧逻辑用于兼容）
const generateVideo = async () => {
  if (!selectedVideoModel.value) {
    ElMessage.warning("请先选择视频生成模型");
    return;
  }
  if (!currentStoryboard.value) {
    ElMessage.warning("请先选择分镜");
    return;
  }
  if (
    selectedReferenceMode.value !== "none" &&
    selectedImagesForVideo.value.length === 0
  ) {
    ElMessage.warning("请选择参考图片");
    return;
  }

  // 获取选中的图片信息
  let selectedImage: any = null;
  if (
    selectedReferenceMode.value !== "none" &&
    selectedImagesForVideo.value.length > 0
  ) {
    selectedImage =
      videoReferenceImages.value.find(
        (img) => img.id === selectedImagesForVideo.value[0],
      ) ||
      previousStoryboardLastFrames.value.find(
        (img) => img.id === selectedImagesForVideo.value[0],
      );
    if (!selectedImage || !selectedImage.image_url) {
      ElMessage.error("请选择有效的参考图片");
      return;
    }
  }

  // 构建请求参数
  const provider = extractProviderFromModel(selectedVideoModel.value);
  const prompt =
    videoPromptText.value ||
    currentStoryboard.value.video_prompt ||
    currentStoryboard.value.action ||
    currentStoryboard.value.description ||
    "";

  const requestParams: any = {
    drama_id: dramaId.toString(),
    storyboard_id: currentStoryboard.value.id,
    prompt,
    duration: videoDuration.value,
    provider,
    model: selectedVideoModel.value,
    reference_mode: selectedReferenceMode.value,
    resolution: videoResolution.value,
    generate_audio: videoGenerateAudio.value,
    enable_subtitle: videoEnableSubtitle.value,
    aspect_ratio: videoAspectRatio.value,
  };

  // 收集参考图展示信息
  const refImages: { name: string; path: string }[] = [];

  switch (selectedReferenceMode.value) {
    case "single":
      if (selectedImage.local_path) {
        requestParams.image_local_path = selectedImage.local_path;
      } else if (selectedImage.image_url) {
        requestParams.image_url = selectedImage.image_url;
      }
      requestParams.image_gen_id = selectedImage.id;
      refImages.push({
        name: `🖼️ 参考图 (${selectedImage.frame_type || "single"})`,
        path: selectedImage.local_path || selectedImage.image_url,
      });
      break;

    case "first_last": {
      const firstImage =
        videoReferenceImages.value.find(
          (img) => img.id === selectedImagesForVideo.value[0],
        ) ||
        previousStoryboardLastFrames.value.find(
          (img) => img.id === selectedImagesForVideo.value[0],
        );
      const lastImage =
        videoReferenceImages.value.find(
          (img) => img.id === selectedLastImageForVideo.value,
        ) ||
        previousStoryboardLastFrames.value.find(
          (img) => img.id === selectedLastImageForVideo.value,
        );

      if (firstImage?.local_path) {
        requestParams.first_frame_local_path = firstImage.local_path;
      } else if (firstImage?.image_url) {
        requestParams.first_frame_url = firstImage.image_url;
      }
      if (lastImage?.local_path) {
        requestParams.last_frame_local_path = lastImage.local_path;
      } else if (lastImage?.image_url) {
        requestParams.last_frame_url = lastImage.image_url;
      }
      if (firstImage) {
        refImages.push({ name: "🎬 首帧", path: firstImage.local_path || firstImage.image_url || "" });
      }
      if (lastImage) {
        refImages.push({ name: "🎬 尾帧", path: lastImage.local_path || lastImage.image_url || "" });
      }
      break;
    }

    case "multiple": {
      const selectedImages = selectedImagesForVideo.value
        .map((id) => videoReferenceImages.value.find((img) => img.id === id))
        .filter((img) => img?.local_path || img?.image_url);
      requestParams.reference_image_urls = selectedImages.map((img) => img!.local_path || img!.image_url);
      selectedImages.forEach((img, idx) => {
        refImages.push({ name: `🖼️ 参考图${idx + 1}`, path: img!.local_path || img!.image_url || "" });
      });
      break;
    }

    case "none":
      break;
  }

  // 获取模型价格
  const modelCap = currentModelCapability.value;
  const modeLabels: Record<string, string> = { single: "单图", first_last: "首尾帧", multiple: "多图", none: "纯文本" };

  // 计算预估价格（基于实际测试数据）
  const estimatedPrice = calculateEstimatedPrice(0, 0, selectedVideoModel.value, videoGenerateAudio.value, videoDuration.value);

  // 初始化弹窗
  videoGenDialog.visible = true;
  videoGenDialog.phase = "confirm";
  videoGenDialog.prompt = fullPrompt;
  videoGenDialog.model = selectedVideoModel.value;
  videoGenDialog.pricing = modelCap?.pricing || "-";
  videoGenDialog.estimatedPrice = estimatedPrice;
  videoGenDialog.referenceMode = modeLabels[selectedReferenceMode.value] || selectedReferenceMode.value;
  videoGenDialog.referenceImages = refImages;
  videoGenDialog.videoGenId = null;
  videoGenDialog.progress = 0;
  videoGenDialog.statusText = "";
  videoGenDialog.error = "";
  videoGenDialog.curlCommand = "";
  videoGenDialog.aborted = false;
  videoGenDialog.resolution = videoResolution.value;
  videoGenDialog.requestParams = requestParams;
};

// Debug: 展示 curl 命令，不调用 API
const debugGenerateVideo = () => {
  if (!videoGenDialog.requestParams) return;
  videoGenDialog.requestParams.resolution = videoGenDialog.resolution;
  const apiBase = window.location.origin + '/api/v1';
  const curlBody = JSON.stringify(videoGenDialog.requestParams, null, 2);
  const curlCmd = `curl -X POST '${apiBase}/videos' \\\n  -H 'Content-Type: application/json' \\\n  -d '${curlBody.replace(/'/g, "'\\''")}'`;
  videoGenDialog.phase = "done";
  videoGenDialog.statusText = "";
  videoGenDialog.error = "";
  videoGenDialog.curlCommand = curlCmd;
};

// 确认并执行视频生成
const confirmGenerateVideo = async () => {
  if (!videoGenDialog.requestParams) return;
  const genStoryboardId = currentStoryboard.value?.id;

  // 使用对话框中选择的分辨率更新请求参数
  videoGenDialog.requestParams.resolution = videoGenDialog.resolution;

  videoGenDialog.phase = "generating";
  videoGenDialog.progress = 10;
  videoGenDialog.statusText = "正在提交生成任务...";
  if (genStoryboardId) generatingVideoIds.value.add(genStoryboardId);

  try {
    const result = await videoAPI.generateVideo(videoGenDialog.requestParams);

    videoGenDialog.videoGenId = result.id || result.ID;
    videoGenDialog.progress = 20;
    videoGenDialog.statusText = "任务已提交，等待生成中...";

    generatedVideos.value.unshift(result);

    // 轮询等待视频完成（视频比图片慢，最多10分钟）
    const maxWaitMs = 10 * 60 * 1000;
    const startTime = Date.now();
    const interval = 5000;

    while (Date.now() - startTime < maxWaitMs) {
      if (videoGenDialog.aborted) {
        videoGenDialog.statusText = "已终止";
        videoGenDialog.phase = "done";
        break;
      }

      // 检测弹窗是否关闭，如果关闭则切换到全局轮询
      if (!videoGenDialog.visible && videoGenDialog.phase === 'generating') {
        videoGenDialog.statusText = "生成时间较长，已在后台继续处理";
        startVideoPolling();
        break;
      }

      await new Promise((resolve) => setTimeout(resolve, interval));

      try {
        const vid = await videoAPI.getVideo(videoGenDialog.videoGenId!);

        if (vid.status === "completed") {
          videoGenDialog.progress = 100;
          videoGenDialog.statusText = "视频生成完成！";
          videoGenDialog.phase = "done";

          const idx = generatedVideos.value.findIndex((v) => v.id === videoGenDialog.videoGenId);
          if (idx >= 0) {
            generatedVideos.value[idx] = vid;
          }

          // 刷新素材库（后端自动创建了asset记录）
          await loadVideoAssets();

          setTimeout(() => {
            if (videoGenDialog.phase === "done" && !videoGenDialog.error) {
              videoGenDialog.visible = false;
            }
          }, 2000);
          break;
        } else if (vid.status === "failed") {
          videoGenDialog.progress = 100;
          videoGenDialog.error = vid.error_msg || "生成失败";
          videoGenDialog.statusText = "生成失败";
          videoGenDialog.phase = "done";

          const idx = generatedVideos.value.findIndex((v) => v.id === videoGenDialog.videoGenId);
          if (idx >= 0) {
            generatedVideos.value[idx] = vid;
          }
          break;
        } else {
          const elapsed = Date.now() - startTime;
          videoGenDialog.progress = Math.min(20 + Math.round((elapsed / maxWaitMs) * 70), 90);
          videoGenDialog.statusText = "AI 正在生成视频...";
        }
      } catch {
        // 查询失败，继续轮询
      }
    }

    // 如果超时或弹窗关闭，启动全局轮询
    if (videoGenDialog.phase === "generating") {
      videoGenDialog.statusText = "生成时间较长，已在后台继续处理";
      videoGenDialog.phase = "done";
      startVideoPolling();
    }
  } catch (error: any) {
    videoGenDialog.error = error.message || "未知错误";
    videoGenDialog.statusText = "提交失败";
    videoGenDialog.phase = "done";
  } finally {
    if (genStoryboardId) generatingVideoIds.value.delete(genStoryboardId);
  }
};

// 终止视频生成
const abortVideoGeneration = () => {
  videoGenDialog.aborted = true;
  videoGenDialog.statusText = "正在终止...";
  startVideoPolling();
};

// 处理视频生成弹窗关闭
const handleVideoGenDialogClose = (done: () => void) => {
  if (videoGenDialog.phase === 'generating') {
    // 生成过程中关闭，只关闭弹窗，后台继续运行
    ElMessage.info('视频生成继续在后台进行，您可以随时查看进度');
    done();
  } else {
    // 其他阶段直接关闭
    done();
  }
};

// 加载分镜的视频参考图片（所有帧类型）
const loadVideoReferenceImages = async (storyboardId: number) => {
  try {
    const result = await imageAPI.listImages({
      storyboard_id: storyboardId,
      page: 1,
      page_size: 100,
    });
    videoReferenceImages.value = result.items || [];
  } catch (error: any) {
    console.error("加载视频参考图片失败:", error);
  }
};

// 加载分镜的视频列表
const loadStoryboardVideos = async (storyboardId: number) => {
  loadingVideos.value = true;
  try {
    const result = await videoAPI.listVideos({
      storyboard_id: storyboardId.toString(),
      page: 1,
      page_size: 50,
    });
    generatedVideos.value = result.items || [];

    // 如果有进行中的任务，启动轮询
    const hasPendingOrProcessing = generatedVideos.value.some(
      (v) => v.status === "pending" || v.status === "processing",
    );
    if (hasPendingOrProcessing) {
      startVideoPolling();
    }
  } catch (error: any) {
    console.error("加载视频列表失败:", error);
  } finally {
    loadingVideos.value = false;
  }
};

// 启动视频状态轮询
const startVideoPolling = () => {
  if (videoPollingTimer) return;

  videoPollingTimer = setInterval(async () => {
    if (!currentStoryboard.value) {
      stopVideoPolling();
      return;
    }

    try {
      // 保存旧的视频列表用于对比
      const oldVideos = [...generatedVideos.value];

      const result = await videoAPI.listVideos({
        storyboard_id: currentStoryboard.value.id.toString(),
        page: 1,
        page_size: 50,
      });
      generatedVideos.value = result.items || [];

      // 更新弹窗进度（如果弹窗正在显示生成中）
      if (videoGenDialog.phase === 'generating' && videoGenDialog.videoGenId) {
        const currentVideo = generatedVideos.value.find(v => v.id === videoGenDialog.videoGenId);
        if (currentVideo) {
          if (currentVideo.status === 'completed') {
            videoGenDialog.progress = 100;
            videoGenDialog.statusText = '生成完成';
            videoGenDialog.phase = 'done';
            stopVideoPolling();
          } else if (currentVideo.status === 'failed') {
            videoGenDialog.progress = 100;
            videoGenDialog.error = currentVideo.error_msg || '生成失败';
            videoGenDialog.statusText = '生成失败';
            videoGenDialog.phase = 'done';
            stopVideoPolling();
          } else if (currentVideo.status === 'processing') {
            // 根据时间估算进度（每5秒增加5%，最高到95%）
            if (videoGenDialog.progress < 95) {
              videoGenDialog.progress = Math.min(95, videoGenDialog.progress + 5);
            }
            videoGenDialog.statusText = 'AI 正在生成视频...';
          }
        }
      }

      // 检测是否有视频从 processing 变为 completed
      const hasNewlyCompleted = generatedVideos.value.some((newVideo) => {
        const oldVideo = oldVideos.find((v) => v.id === newVideo.id);
        return (
          oldVideo &&
          (oldVideo.status === "pending" || oldVideo.status === "processing") &&
          newVideo.status === "completed"
        );
      });

      // 如果有视频完成，重新加载分镜列表和素材库
      if (hasNewlyCompleted) {
        if (episodeId.value) {
          try {
            const storyboardsRes = await dramaAPI.getStoryboards(
              episodeId.value.toString(),
            );
            storyboards.value = storyboardsRes?.storyboards || [];
          } catch (error) {
            console.error("重新加载分镜列表失败:", error);
          }
        }
        // 刷新素材库（后端视频完成后自动创建asset记录）
        await loadVideoAssets();
      }

      // 如果没有进行中的任务，停止轮询
      const hasPendingOrProcessing = generatedVideos.value.some(
        (v) => v.status === "pending" || v.status === "processing",
      );
      if (!hasPendingOrProcessing) {
        stopVideoPolling();
      }
    } catch (error) {
      console.error("轮询视频状态失败:", error);
    }
  }, 5000); // 每5秒轮询一次
};

// 停止视频轮询
const stopVideoPolling = () => {
  if (videoPollingTimer) {
    clearInterval(videoPollingTimer);
    videoPollingTimer = null;
  }
};

const toggleCharacterInShot = async (charId: number) => {
  if (!currentStoryboard.value) return;

  // 初始化characters数组
  if (!currentStoryboard.value.characters) {
    currentStoryboard.value.characters = [];
  }

  const char = characters.value.find((c) => c.id === charId);
  if (!char) return;

  // 检查是否已存在
  const existIndex = currentStoryboard.value.characters.findIndex((c) =>
    typeof c === "object" ? c.id === charId : c === charId,
  );

  if (existIndex > -1) {
    // 移除角色
    currentStoryboard.value.characters.splice(existIndex, 1);
  } else {
    // 添加角色（作为对象）
    currentStoryboard.value.characters.push(char);
  }

  // 保存到后端
  try {
    const characterIds = currentStoryboard.value.characters.map((c) =>
      typeof c === "object" ? c.id : c,
    );

    await dramaAPI.updateStoryboard(currentStoryboard.value.id.toString(), {
      character_ids: characterIds,
    });

    if (existIndex > -1) {
      ElMessage.success(`已移除角色: ${char.name}`);
    } else {
      ElMessage.success(`已添加角色: ${char.name}`);
    }
  } catch (error: any) {
    ElMessage.error("保存失败: " + (error.message || "未知错误"));
    // 回滚操作
    if (existIndex > -1) {
      currentStoryboard.value.characters.push(char);
    } else {
      currentStoryboard.value.characters.splice(
        currentStoryboard.value.characters.length - 1,
        1,
      );
    }
  }
};

const removeCharacterFromShot = async (charId: number) => {
  if (!currentStoryboard.value) return;

  // 初始化characters数组
  if (!currentStoryboard.value.characters) {
    currentStoryboard.value.characters = [];
  }

  const char = characters.value.find((c) => c.id === charId);
  if (!char) return;

  // 检查是否已存在
  const existIndex = currentStoryboard.value.characters.findIndex((c) =>
    typeof c === "object" ? c.id === charId : c === charId,
  );

  if (existIndex > -1) {
    // 移除角色
    currentStoryboard.value.characters.splice(existIndex, 1);
  }

  // 保存到后端
  try {
    const characterIds = currentStoryboard.value.characters.map((c) =>
      typeof c === "object" ? c.id : c,
    );

    await dramaAPI.updateStoryboard(currentStoryboard.value.id.toString(), {
      character_ids: characterIds,
    });

    ElMessage.success(`已移除角色: ${char.name}`);
  } catch (error: any) {
    ElMessage.error("保存失败: " + (error.message || "未知错误"));
    // 回滚操作
    currentStoryboard.value.characters.push(char);
  }
};

const loadData = async () => {
  try {
    // 加载剧集信息
    const dramaRes = await dramaAPI.get(dramaId.toString());
    drama.value = dramaRes;

    // 找到当前章节
    const ep = dramaRes.episodes?.find(
      (e) => e.episode_number === episodeNumber,
    );
    if (!ep) {
      ElMessage.error("章节不存在");
      router.back();
      return;
    }

    episode.value = ep;
    episodeId.value = ep.id;

    // 加载分镜列表
    const storyboardsRes = await dramaAPI.getStoryboards(ep.id.toString());

    // API返回格式: {storyboards: [...], total: number}
    storyboards.value = storyboardsRes?.storyboards || [];

    // 默认选中第一个分镜
    if (storyboards.value.length > 0 && !currentStoryboardId.value) {
      currentStoryboardId.value = storyboards.value[0].id;
    }

    // 加载角色列表
    characters.value = dramaRes.characters || [];

    // 加载可用场景列表
    availableScenes.value = dramaRes.scenes || [];

    // 加载道具列表
    props.value = dramaRes.props || [];

    // 加载视频素材库
    await loadVideoAssets();
  } catch (error: any) {
    ElMessage.error("加载数据失败: " + (error.message || "未知错误"));
  }
};

const selectScene = async (sceneId: number) => {
  if (!currentStoryboard.value) return;

  try {
    // TODO: 调用API更新分镜的scene_id
    await dramaAPI.updateStoryboard(currentStoryboard.value.id.toString(), {
      scene_id: sceneId,
    });

    // 重新加载数据
    await loadData();
    showSceneSelector.value = false;
    ElMessage.success("场景关联成功");
  } catch (error: any) {
    ElMessage.error(error.message || "场景关联失败");
  }
};

const selectStoryboard = (id: string) => {
  currentStoryboardId.value = id;
};

const handleTimelineSelect = (sceneId: number) => {
  selectStoryboard(String(sceneId));
};

const togglePlay = () => {
  if (currentPlayState.value === "playing") {
    currentPlayState.value = "paused";
  } else {
    currentPlayState.value = "playing";
  }
};

const formatTime = (seconds: number) => {
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
};

const zoomIn = () => {
  ElMessage.info("时间线缩放功能开发中");
};

const zoomOut = () => {
  ElMessage.info("时间线缩放功能开发中");
};

const generateImage = async () => {
  if (!currentStoryboard.value) return;

  try {
    ElMessage.info("图片生成功能开发中");
  } catch (error: any) {
    ElMessage.error(error.message || "生成失败");
  }
};

const uploadImage = () => {
  if (!currentStoryboard.value) {
    ElMessage.warning("请先选择镜头");
    return;
  }

  // 创建隐藏的文件输入
  const input = document.createElement("input");
  input.type = "file";
  input.accept = "image/*";
  input.onchange = async (e: Event) => {
    const target = e.target as HTMLInputElement;
    const file = target.files?.[0];
    if (!file) return;

    // 验证文件大小 (10MB)
    if (file.size > 10 * 1024 * 1024) {
      ElMessage.error("图片大小不能超过 10MB");
      return;
    }

    try {
      // 创建 FormData
      const formData = new FormData();
      formData.append("file", file);

      // 上传到服务器
      const response = await fetch("/api/v1/upload/image", {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        throw new Error("上传失败");
      }

      const result = await response.json();
      const imageUrl = result.data?.url;

      if (imageUrl && currentStoryboard.value) {
        // 创建图片生成记录（关联到当前镜头和帧类型）
        await imageAPI.uploadImage({
          storyboard_id: currentStoryboard.value.id,
          drama_id: parseInt(dramaId),
          frame_type: selectedFrameType.value || "first",
          image_url: imageUrl,
          prompt: currentFramePrompt.value || "用户上传图片",
        });

        // 刷新图片列表
        await loadStoryboardImages(
          currentStoryboard.value.id,
          selectedFrameType.value,
        );

        ElMessage.success("图片上传成功");
      }
    } catch (error: any) {
      console.error("上传图片失败:", error);
      ElMessage.error(error.message || "上传失败");
    }
  };
  input.click();
};

// 一键清除失败图片
const clearFailedImages = async () => {
  const failedImages = generatedImages.value.filter(img => img.status === 'failed');
  if (failedImages.length === 0) {
    ElMessage.info("没有失败的图片");
    return;
  }

  try {
    await ElMessageBox.confirm(
      `确定要清除 ${failedImages.length} 张失败的图片记录吗？`,
      "清除失败图片",
      { confirmButtonText: "清除", cancelButtonText: "取消", type: "warning" }
    );

    let deleted = 0;
    for (const img of failedImages) {
      try {
        await imageAPI.deleteImage(img.id);
        deleted++;
      } catch (e) {
      }
    }

    ElMessage.success(`已清除 ${deleted} 张失败图片`);

    if (currentStoryboard.value) {
      await loadStoryboardImages(currentStoryboard.value.id, selectedFrameType.value);
    }
  } catch {
    // 用户取消
  }
};

// 打开图片编辑器（Seededit）
const openImageEditor = (img: ImageGeneration) => {
  router.push({ name: 'ImageEditor', params: { id: img.id } })
}

// 删除图片
const handleDeleteImage = async (img: ImageGeneration) => {
  if (!currentStoryboard.value) return;

  try {
    await ElMessageBox.confirm("确定要删除这张图片吗？", "确认删除", {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    });

    await imageAPI.deleteImage(img.id);
    ElMessage.success("删除成功");

    // 重新加载当前帧类型的图片列表
    await loadStoryboardImages(
      currentStoryboard.value.id,
      selectedFrameType.value,
    );
  } catch (error: any) {
    if (error !== "cancel") {
      console.error("删除图片失败:", error);
      ElMessage.error(error.message || "删除失败");
    }
  }
};

// 加载所有已生成的图片（用于宫格编辑器）
const loadAllGeneratedImages = async () => {
  if (!currentStoryboard.value) return;

  try {
    const result = await imageAPI.listImages({
      storyboard_id: currentStoryboard.value.id,
      page: 1,
      page_size: 100,
    });
    allGeneratedImages.value = result.items || [];
  } catch (error: any) {
    console.error("加载所有图片失败:", error);
  }
};

// 处理宫格图片创建成功
const handleGridImageSuccess = async () => {
  if (currentStoryboard.value) {
    // 刷新动作序列图片列表
    await loadStoryboardImages(currentStoryboard.value.id, "action");
    // 重新加载所有图片
    await loadAllGeneratedImages();
  }
};

// 打开裁剪对话框
const openCropDialog = (img: ImageGeneration) => {
  cropImageData.value = img;
  cropImageUrl.value = getImageUrl(img) || "";
  showCropDialog.value = true;
};

// 处理裁剪保存
const handleCropSave = async (images: { blob: Blob; frameType: string }[]) => {
  if (!currentStoryboard.value || !cropImageData.value) return;

  try {
    // 将 Blob 转换为 base64 data URL
    const convertBlobToBase64 = (blob: Blob): Promise<string> => {
      return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onloadend = () => resolve(reader.result as string);
        reader.onerror = reject;
        reader.readAsDataURL(blob);
      });
    };

    // 上传裁剪后的图片并创建新的图片生成记录
    for (const img of images) {
      // 将 Blob 转换为 base64
      const imageUrl = await convertBlobToBase64(img.blob);

      // 调用上传接口
      await imageAPI.uploadImage({
        storyboard_id: currentStoryboard.value.id,
        drama_id: Number(dramaId),
        frame_type: img.frameType,
        image_url: imageUrl,
        prompt: cropImageData.value.prompt || "",
      });
    }

    ElMessage.success("裁剪图片保存成功");

    // 刷新图片列表 - 刷新所有帧类型的图片，确保新裁剪的图片能在视频生成tab中被选择到
    if (currentStoryboard.value) {
      // 刷新当前镜头的所有图片（不限制帧类型）
      await loadStoryboardImages(currentStoryboard.value.id);
      // 刷新所有生成的图片列表
      await loadAllGeneratedImages();
    }
  } catch (error) {
    console.error("Failed to save cropped images:", error);
    ElMessage.error("保存裁剪图片失败");
  }
};

const goBack = () => {
  router.replace({
    name: "EpisodeWorkflowNew",
    params: { id: dramaId, episodeNumber },
  });
};

const handleAddStoryboard = async () => {
  if (!episodeId.value) return;

  try {
    const nextShotNumber =
      storyboards.value.length > 0
        ? Math.max(...storyboards.value.map((s) => s.storyboard_number)) + 1
        : 1;

    await dramaAPI.createStoryboard({
      episode_id: parseInt(episodeId.value),
      storyboard_number: nextShotNumber,
      title: `镜头 ${nextShotNumber}`,
      description: "新镜头描述",
      action: "动作描述",
      dialogue: "",
      duration: 5,
      scene_id:
        storyboards.value.length > 0
          ? storyboards.value[storyboards.value.length - 1].scene_id
          : undefined,
    });

    ElMessage.success("添加分镜成功");
    await loadData(); // Refresh list

    // Select the new storyboard (the last one)
    if (storyboards.value.length > 0) {
      selectStoryboard(storyboards.value[storyboards.value.length - 1].id);
    }
  } catch (error: any) {
    console.error("添加分镜失败:", error);
    ElMessage.error(error.message || "添加分镜失败");
  }
};

const handleDeleteStoryboard = async (storyboard: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除镜头 ${storyboard.storyboard_number} 吗？此操作不可恢复。`,
      "删除确认",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      },
    );

    await dramaAPI.deleteStoryboard(storyboard.id);
    ElMessage.success("删除分镜成功");

    // If deleted current storyboard, clear selection or select another
    if (currentStoryboardId.value === storyboard.id) {
      currentStoryboardId.value = undefined;
      currentStoryboard.value = undefined;
    }

    await loadData();
  } catch (error: any) {
    if (error !== "cancel") {
      console.error("删除分镜失败:", error);
      ElMessage.error(error.message || "删除分镜失败");
    }
  }
};

// 加载视频合成列表
const loadVideoMerges = async () => {
  if (!episodeId.value) return;

  try {
    loadingMerges.value = true;
    const result = await videoMergeAPI.listMerges({
      episode_id: episodeId.value.toString(),
      page: 1,
      page_size: 20,
    });
    videoMerges.value = result.merges;

    // 检查是否有进行中的任务
    const hasProcessingTasks = result.merges.some(
      (merge: any) =>
        merge.status === "pending" || merge.status === "processing",
    );

    if (hasProcessingTasks) {
      startMergePolling();
    } else {
      stopMergePolling();
    }
  } catch (error: any) {
    console.error("加载视频合成列表失败:", error);
    ElMessage.error("加载视频合成列表失败");
  } finally {
    loadingMerges.value = false;
  }
};

// 启动视频合成列表轮询
const startMergePolling = () => {
  if (mergePollingTimer) return;

  mergePollingTimer = setInterval(async () => {
    if (!episodeId.value) {
      stopMergePolling();
      return;
    }

    try {
      const result = await videoMergeAPI.listMerges({
        episode_id: episodeId.value.toString(),
        page: 1,
        page_size: 20,
      });
      videoMerges.value = result.merges;

      // 检查是否还有进行中的任务
      const hasProcessingTasks = result.merges.some(
        (merge: any) =>
          merge.status === "pending" || merge.status === "processing",
      );

      if (!hasProcessingTasks) {
        stopMergePolling();
      }
    } catch (error) {
      console.error('合成列表轮询失败:', error)
      stopMergePolling()
    }
  }, 3000); // 每3秒轮询一次
};

// 停止视频合成列表轮询
const stopMergePolling = () => {
  if (mergePollingTimer) {
    clearInterval(mergePollingTimer);
    mergePollingTimer = null;
  }
};

// ========== 批量任务弹窗 ==========
const handleBatchDialogClose = () => {
  if (batchTaskDialog.phase === 'running') return; // 运行中不允许关闭
};

const resetBatchDialog = (type: "prompt" | "image") => {
  batchTaskDialog.type = type;
  batchTaskDialog.title = type === "prompt" ? "一键AI提取提示词" : "一键生成图片";
  batchTaskDialog.phase = "config";
  batchTaskDialog.skipExisting = true;
  batchTaskDialog.frameType = "first";
  batchTaskDialog.progress = 0;
  batchTaskDialog.statusText = "";
  batchTaskDialog.detail = "";
  batchTaskDialog.total = 0;
  batchTaskDialog.completed = 0;
  batchTaskDialog.failed = 0;
};

// ========== 批量生成提示词（固定首帧类型） ==========
const handleBatchGeneratePrompts = () => {
  if (!episodeId.value) {
    ElMessage.warning("未找到章节信息");
    return;
  }
  if (storyboards.value.length === 0) {
    ElMessage.warning("当前章节没有分镜数据，请先拆分分镜");
    return;
  }
  resetBatchDialog("prompt");
  batchTaskDialog.visible = true;
};

// ========== 批量生成图片 ==========
const handleBatchGenerateImages = () => {
  if (!episodeId.value) {
    ElMessage.warning("未找到章节信息");
    return;
  }
  if (storyboards.value.length === 0) {
    ElMessage.warning("当前章节没有分镜数据，请先拆分分镜");
    return;
  }
  resetBatchDialog("image");
  batchTaskDialog.visible = true;
};

// ========== 开始执行批量任务 ==========
const startBatchTask = async () => {
  if (batchTaskDialog.type === "prompt") {
    await executeBatchPromptGeneration();
  } else {
    await executeBatchImageGeneration();
  }
};

// 执行批量生成提示词
const executeBatchPromptGeneration = async () => {
  batchTaskDialog.phase = "running";
  batchTaskDialog.progress = 0;
  batchTaskDialog.statusText = "正在提交任务...";
  batchTaskDialog.detail = "";
  batchPromptGenerating.value = true;

  const selectedFrameTypeForPrompt = batchTaskDialog.frameType as any;

  try {
    const { task_id } = await batchGenerateFramePrompts(episodeId.value!, {
      frame_type: selectedFrameTypeForPrompt,
      skip_existing: batchTaskDialog.skipExisting,
    });

    batchTaskDialog.statusText = `任务已提交，正在生成「${getFrameTypeLabel(selectedFrameTypeForPrompt)}」提示词...`;

    // 轮询任务状态
    while (true) {
      await new Promise((resolve) => setTimeout(resolve, 2000));
      const task = await taskAPI.getStatus(task_id);

      if (task.status === "completed") {
        let result = task.result;
        if (typeof result === "string") {
          try { result = JSON.parse(result); } catch { /* ignore */ }
        }
        batchTaskDialog.total = result?.total || 0;
        batchTaskDialog.completed = result?.completed || 0;
        batchTaskDialog.failed = result?.failed || 0;
        batchTaskDialog.progress = 100;
        batchTaskDialog.statusText = "正在同步提示词数据...";

        // 并行刷新所有分镜的提示词到 sessionStorage
        const frameType = selectedFrameTypeForPrompt;
        await Promise.allSettled(
          storyboards.value.map(async (sb) => {
            const prompts = await getStoryboardFramePrompts(sb.id);
            if (prompts?.frame_prompts) {
              const matched = prompts.frame_prompts.find((p: any) => p.frame_type === frameType);
              if (matched) {
                const storageKey = getPromptStorageKey(sb.id, frameType);
                if (storageKey) {
                  sessionStorage.setItem(storageKey, matched.prompt);
                }
                if (currentStoryboard.value && currentStoryboard.value.id === sb.id && selectedFrameType.value === frameType) {
                  currentFramePrompt.value = matched.prompt;
                  framePrompts.value[frameType] = matched.prompt;
                }
              }
            }
          })
        );

        batchTaskDialog.phase = "done";
        break;
      } else if (task.status === "failed") {
        batchTaskDialog.total = 0;
        batchTaskDialog.completed = 0;
        batchTaskDialog.failed = 1;
        batchTaskDialog.progress = 100;
        batchTaskDialog.phase = "done";
        batchTaskDialog.statusText = task.message || "未知错误";
        ElMessage.error("提示词生成失败: " + (task.message || "未知错误"));
        break;
      } else {
        // processing - 解析进度
        const msg = task.message || "";
        const progressMatch = msg.match(/第\s*(\d+)\s*\/\s*(\d+)/);
        if (progressMatch) {
          const current = parseInt(progressMatch[1]);
          const total = parseInt(progressMatch[2]);
          batchTaskDialog.progress = Math.min(Math.round((current / total) * 95), 95);
          batchTaskDialog.statusText = `正在AI提取提示词 (${current}/${total})...`;
        } else {
          batchTaskDialog.statusText = msg || "处理中...";
        }
        // 解析已完成/失败数
        const completedMatch = msg.match(/已完成:\s*(\d+)/);
        const failedMatch = msg.match(/失败:\s*(\d+)/);
        if (completedMatch) batchTaskDialog.detail = `已完成: ${completedMatch[1]}${failedMatch ? `，失败: ${failedMatch[1]}` : ""}`;
      }
    }
  } catch (error: any) {
    batchTaskDialog.phase = "done";
    batchTaskDialog.total = 0;
    batchTaskDialog.completed = 0;
    batchTaskDialog.failed = 1;
    batchTaskDialog.statusText = error.message || "未知错误";
    ElMessage.error("批量AI提取提示词失败: " + (error.message || "未知错误"));
  } finally {
    batchPromptGenerating.value = false;
  }
};

// 执行批量生成图片
const executeBatchImageGeneration = async () => {
  batchTaskDialog.phase = "running";
  batchTaskDialog.progress = 0;
  batchTaskDialog.statusText = "正在提交图片生成任务...";
  batchTaskDialog.detail = "";
  batchImageGenerating.value = true;

  try {
    const imageGens = await imageAPI.batchGenerateForEpisode(episodeId.value!, {
      skipExisting: batchTaskDialog.skipExisting,
      frameType: batchTaskDialog.frameType,
    });
    const imageCount = Array.isArray(imageGens) ? imageGens.length : 0;

    if (imageCount === 0) {
      batchTaskDialog.phase = "done";
      batchTaskDialog.total = 0;
      batchTaskDialog.completed = 0;
      batchTaskDialog.failed = 0;
      batchTaskDialog.progress = 100;
      batchTaskDialog.statusText = batchTaskDialog.skipExisting
        ? "所有分镜已有图片或缺少提示词，无需生成"
        : "没有可生成图片的分镜（缺少提示词）";
      return;
    }

    batchTaskDialog.total = imageCount;
    batchTaskDialog.statusText = `已提交 ${imageCount} 个图片生成任务，等待完成...`;

    // 轮询等待所有图片完成
    const imageIds = (imageGens as any[]).map((img: any) => img.id || img.ID);
    const maxWaitMs = 10 * 60 * 1000;
    const startTime = Date.now();
    const interval = 5000;

    while (Date.now() - startTime < maxWaitMs) {
      await new Promise(resolve => setTimeout(resolve, interval));

      let allDone = true;
      let completedCount = 0;
      let failedCount = 0;

      for (const imgId of imageIds) {
        try {
          const img = await imageAPI.getImage(imgId);
          if (img.status === "completed") {
            completedCount++;
          } else if (img.status === "failed") {
            failedCount++;
          } else {
            allDone = false;
          }
        } catch {
          allDone = false;
        }
      }

      batchTaskDialog.completed = completedCount;
      batchTaskDialog.failed = failedCount;
      const doneCount = completedCount + failedCount;
      batchTaskDialog.progress = Math.min(Math.round((doneCount / imageCount) * 100), 100);
      batchTaskDialog.statusText = `图片生成中 (${doneCount}/${imageCount})...`;
      batchTaskDialog.detail = `完成: ${completedCount}${failedCount > 0 ? `，失败: ${failedCount}` : ""}，处理中: ${imageCount - doneCount}`;

      if (allDone || doneCount >= imageCount) {
        batchTaskDialog.progress = 100;
        batchTaskDialog.phase = "done";

        // 刷新数据
        if (currentStoryboard.value) {
          await loadStoryboardImages(currentStoryboard.value.id, selectedFrameType.value);
        }
        return;
      }
    }

    // 超时
    batchTaskDialog.phase = "done";
    batchTaskDialog.statusText = "部分图片可能仍在生成中（已超时）";
  } catch (error: any) {
    batchTaskDialog.phase = "done";
    batchTaskDialog.total = 0;
    batchTaskDialog.completed = 0;
    batchTaskDialog.failed = 1;
    batchTaskDialog.statusText = error.message || "未知错误";
    ElMessage.error("批量生成图片失败: " + (error.message || "未知错误"));
  } finally {
    batchImageGenerating.value = false;
  }
};

// 轮询图片列表，等待指定的图片全部完成
const pollImagesUntilDone = async (imageIds: number[]) => {
  const maxWaitMs = 10 * 60 * 1000; // 最多等10分钟
  const startTime = Date.now();
  const interval = 5000; // 每5秒检查一次

  while (Date.now() - startTime < maxWaitMs) {
    await new Promise(resolve => setTimeout(resolve, interval));

    let allDone = true;
    let completedCount = 0;
    let failedCount = 0;

    for (const imgId of imageIds) {
      try {
        const img = await imageAPI.getImage(imgId);
        if (img.status === "completed") {
          completedCount++;
        } else if (img.status === "failed") {
          failedCount++;
        } else {
          allDone = false;
        }
      } catch {
        // 如果查询失败，视为未完成
        allDone = false;
      }
    }

    const pendingCount = imageIds.length - completedCount - failedCount;
    ElMessage.info({
      message: `图片进度：完成 ${completedCount}/${imageIds.length}${failedCount > 0 ? `，失败 ${failedCount}` : ""}${pendingCount > 0 ? `，处理中 ${pendingCount}` : ""}`,
      duration: 4000,
    });

    if (allDone || (completedCount + failedCount >= imageIds.length)) {
      return;
    }
  }

  ElMessage.warning("图片生成等待超时，将继续执行后续步骤");
};

// 通用的轮询任务直到完成
const pollTaskUntilDone = async (taskId: string, stepName: string): Promise<any> => {
  while (true) {
    const task = await taskAPI.getStatus(taskId);
    if (task.status === "completed") {
      ElMessage.success(`${stepName}完成！`);
      return task.result;
    } else if (task.status === "failed") {
      throw new Error(`${stepName}失败: ${task.message || "未知错误"}`);
    }
    await new Promise(resolve => setTimeout(resolve, 3000));
  }
};

// 处理视频合成完成事件
const handleMergeCompleted = async (mergeId: number) => {
  // 刷新视频合成列表
  await loadVideoMerges();
  // 切换到视频合成标签页
  activeTab.value = "merges";
};

// 下载视频
const downloadVideo = async (url: string, title: string) => {
  try {
    const loadingMsg = ElMessage.info({
      message: "正在准备下载...",
      duration: 0,
    });

    // 处理相对路径，添加 /static/ 前缀
    const videoUrl = url.startsWith("http") ? url : `/static/${url}`;

    // 使用fetch获取视频blob
    const response = await fetch(videoUrl);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const blob = await response.blob();
    const blobUrl = window.URL.createObjectURL(blob);

    // 创建下载链接
    const link = document.createElement("a");
    link.href = blobUrl;
    link.download = `${title}.mp4`;
    link.style.display = "none";
    document.body.appendChild(link);
    link.click();

    // 清理
    setTimeout(() => {
      document.body.removeChild(link);
      window.URL.revokeObjectURL(blobUrl);
    }, 100);

    loadingMsg.close();
    ElMessage.success("视频下载已开始");
  } catch (error) {
    console.error("下载视频失败:", error);
    ElMessage.error("视频下载失败，请稍后重试");
  }
};

// ========== V3 链式生成视频 ==========
// 链式视频生成已禁用
const handleBatchGenerateVideos = async () => {
  ElMessage.info("链式视频生成功能已暂时禁用");
};

// V3 链式视频生成已禁用
// const startChainVideoGeneration = async () => { ... };

// 链式生成进度轮询已禁用
// const startChainProgressPolling = () => { ... };

// 预览合成视频
const previewMergedVideo = (url: string) => {
  // 处理相对路径，添加 /static/ 前缀
  const videoUrl = url.startsWith("http") ? url : `/static/${url}`;
  window.open(videoUrl, "_blank");
};

// 删除视频合成记录
const deleteMerge = async (mergeId: number) => {
  try {
    await ElMessageBox.confirm(
      "确定要删除此合成记录吗？此操作不可恢复。",
      "删除确认",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      },
    );

    await videoMergeAPI.deleteMerge(mergeId);
    ElMessage.success("删除成功");
    // 刷新列表
    await loadVideoMerges();
  } catch (error: any) {
    if (error !== "cancel") {
      console.error("删除失败:", error);
      ElMessage.error(error.response?.data?.message || "删除失败");
    }
  }
};

// 格式化日期时间
const formatDateTime = (dateStr: string) => {
  const date = new Date(dateStr);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  if (minutes < 1) return "刚刚";
  if (minutes < 60) return `${minutes}分钟前`;
  if (hours < 24) return `${hours}小时前`;
  if (days < 7) return `${days}天前`;

  // 超过7天显示完整日期
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  const hour = String(date.getHours()).padStart(2, "0");
  const minute = String(date.getMinutes()).padStart(2, "0");
  return `${month}-${day} ${hour}:${minute}`;
};

onMounted(async () => {
  await loadData();
  await loadVideoModels();
  await loadImageModels();
  await loadVideoMerges();

  document.addEventListener('visibilitychange', handleVisibilityChange);
});

const handleVisibilityChange = async () => {
  if (!document.hidden && dramaId.value) {
    try {
      const dramaRes = await dramaAPI.getDrama(dramaId.value.toString());
      characters.value = dramaRes.characters || [];
    } catch (error) {
      console.error('重新加载角色列表失败:', error);
    }
  }
};

onBeforeUnmount(() => {
  stopPolling();
  stopVideoPolling();
  stopMergePolling();
  document.removeEventListener('visibilitychange', handleVisibilityChange);
});
</script>

<style scoped lang="scss">
// 镜头列表项样式
.storyboard-item {
  padding: 8px;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
  border: 1px solid var(--border-primary);
  margin-bottom: 8px;
  display: flex;
  gap: 10px;
  align-items: center;
  background: var(--bg-card);

  &:hover {
    background: var(--bg-card-hover);
    border-color: var(--border-secondary);
  }

  &.active {
    background: var(--accent);
    border-color: var(--accent);

    .shot-number,
    .shot-title {
      color: var(--text-inverse) !important;
    }

    .shot-duration {
      background: rgba(255, 255, 255, 0.2);
      color: var(--text-inverse);
    }
  }

  .shot-thumbnail {
    width: 80px;
    height: 50px;
    border-radius: 4px;
    overflow: hidden;
    background: var(--bg-secondary);
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  .shot-content {
    flex: 1;
    min-width: 0;

    .shot-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 4px;

      .shot-number {
        font-size: 11px;
        color: var(--text-secondary);
        font-weight: 500;
      }

      .shot-duration {
        font-size: 11px;
        color: var(--text-secondary);
        background: var(--bg-secondary);
        padding: 2px 6px;
        border-radius: 3px;
      }
    }

    .shot-title {
      font-size: 13px;
      color: var(--text-primary);
      font-weight: 500;
      line-height: 1.3;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
}

// 视频合成列表样式
.merges-list {
  padding: 16px;
  max-height: calc(100vh - 200px);
  overflow-y: auto;
  background: var(--bg-secondary);

  .merge-items {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .merge-item {
    position: relative;
    background: var(--bg-card);
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: 1px solid var(--border-primary);

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
      border-color: var(--accent);
    }

    .status-indicator {
      position: absolute;
      left: 0;
      top: 0;
      bottom: 0;
      width: 4px;
      transition: all 0.3s;
    }

    &.merge-status-completed .status-indicator {
      background: linear-gradient(to bottom, #67c23a, #85ce61);
    }

    &.merge-status-processing .status-indicator {
      background: linear-gradient(to bottom, #e6a23c, #f0c78a);
      animation: pulse 2s ease-in-out infinite;
    }

    &.merge-status-failed .status-indicator {
      background: linear-gradient(to bottom, #f56c6c, #f89898);
    }

    &.merge-status-pending .status-indicator {
      background: linear-gradient(to bottom, #909399, #b1b3b8);
    }

    .merge-content {
      padding: 20px 24px;
      padding-left: 28px;
    }

    .merge-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
      padding-bottom: 14px;
      border-bottom: 1px solid var(--border-primary);

      .title-section {
        display: flex;
        align-items: center;
        gap: 12px;
        flex: 1;

        .title-icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 38px;
          height: 38px;
          border-radius: 10px;
          background: var(--bg-secondary);
          color: var(--text-secondary);
          transition: all 0.3s;
        }

        .merge-title {
          margin: 0;
          font-size: 15px;
          font-weight: 500;
          color: var(--text-secondary);
          line-height: 1.4;
        }
      }

      :deep(.el-tag) {
        font-weight: 500;
        padding: 4px 12px;
        font-size: 12px;
      }
    }

    &.merge-status-completed .title-icon {
      background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
      color: #fff;
    }

    &.merge-status-processing .title-icon {
      background: linear-gradient(135deg, #e6a23c 0%, #f0c78a 100%);
      color: #fff;
    }

    &.merge-status-failed .title-icon {
      background: linear-gradient(135deg, #f56c6c 0%, #f89898 100%);
      color: #fff;
    }

    .merge-details {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 12px;
      margin-bottom: 16px;

      .detail-item {
        display: flex;
        gap: 10px;
        padding: 12px 14px;
        background: var(--bg-secondary);
        border-radius: 8px;
        border: 1px solid var(--border-primary);
        transition: all 0.3s;

        &:hover {
          border-color: var(--accent);
          transform: translateY(-1px);
        }

        .detail-icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 28px;
          height: 28px;
          border-radius: 6px;
          background: var(--bg-card);
          color: var(--accent);
          flex-shrink: 0;
        }

        .detail-content {
          flex: 1;
          min-width: 0;

          .detail-label {
            font-size: 11px;
            color: var(--text-muted);
            margin-bottom: 3px;
            font-weight: 500;
          }

          .detail-value {
            font-size: 13px;
            color: var(--text-primary);
            font-weight: 500;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
          }
        }
      }
    }

    .merge-error {
      margin-bottom: 12px;

      :deep(.el-alert) {
        border-radius: 8px;
        border: none;
        padding: 8px 12px;
        font-size: 12px;
      }
    }

    .merge-actions {
      display: flex;
      gap: 8px;
      margin-top: 12px;

      :deep(.el-button) {
        flex: 1;
        max-width: 160px;
        font-weight: 500;
        padding: 8px 15px;
        font-size: 13px;
      }
    }
  }
}

// 旋转动画
@keyframes rotating {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

.rotating {
  animation: rotating 2s linear infinite;
}

// 脉冲动画
@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.6;
  }
}

// 白色主题样式
.shot-editor-new {
  padding: 16px;
  height: 100%;
  overflow-y: auto;
  // background: #fff;

  .section-label {
    font-size: 12px;
    color: #666;
    margin-bottom: 8px;
  }

  // 场景预览
  .scene-section {
    margin-bottom: 20px;
  }

  .scene-preview {
    width: 100%;
    height: 80px;
    border-radius: 6px;
    overflow: hidden;
    position: relative;
    background: #f5f5f5;
    border: 1px solid var(--border-primary);

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .scene-info {
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      padding: 6px 8px;
      background: linear-gradient(to top, rgba(0, 0, 0, 0.7), transparent);
      font-size: 11px;
      color: #fff;

      .scene-id {
        font-size: 10px;
        color: #e0e0e0;
        margin-top: 2px;
      }
    }
  }

  .scene-preview-empty {
    width: 100%;
    height: 80px;
    border-radius: 6px;
    border: 1px dashed #d0d0d0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    background: #fafafa;

    .el-icon {
      font-size: 32px !important;
      color: #c0c0c0;
    }

    div {
      font-size: 11px;
      color: #999;
    }
  }

  // 角色列表
  .cast-section {
    margin-bottom: 20px;
  }

  .cast-list {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 8px;

    .cast-item {
      position: relative;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 4px;
      cursor: pointer;
      transition: all 0.2s;

      &:hover {
        .cast-avatar {
          border-color: #409eff;
        }

        .cast-remove {
          opacity: 1;
          visibility: visible;
        }
      }

      &.active {
        .cast-avatar {
          border-color: #409eff;
          background: #409eff;
        }
      }

      .cast-avatar {
        width: 36px;
        height: 36px;
        border-radius: 50%;
        border: 2px solid #e0e0e0;
        overflow: hidden;
        display: flex;
        align-items: center;
        justify-content: center;
        background: #f5f5f5;
        font-size: 14px;
        font-weight: 500;
        color: #666;
        transition: all 0.2s;

        img {
          width: 100%;
          height: 100%;
          object-fit: cover;
        }
      }

      .cast-name {
        font-size: 10px;
        color: #666;
        max-width: 36px;
        text-align: center;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .cast-remove {
        position: absolute;
        top: -3px;
        right: -3px;
        width: 16px;
        height: 16px;
        border-radius: 50%;
        background: #f56c6c;
        color: white;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        transition: all 0.2s;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
        z-index: 10;
        opacity: 0;
        visibility: hidden;
        font-size: 12px;

        &:hover {
          background: #f23030;
          transform: scale(1.1);
        }
      }
    }

    .cast-empty {
      width: 100%;
      text-align: center;
      padding: 15px;
      color: var(--text-muted);
      font-size: 11px;
    }
  }

  // 视效设置
  .settings-section {
    margin-bottom: 16px;

    .settings-grid {
      display: grid;
      grid-template-columns: 1fr 1fr 1fr;
      gap: 10px;

      .setting-item {
        label {
          display: block;
          font-size: 11px;
          color: var(--text-secondary);
          margin-bottom: 6px;
        }
      }
    }

    .audio-controls {
      margin-top: 8px;
    }
  }

  // 叙事内容
  .narrative-section {
    margin-bottom: 14px;
  }

  .frame-desc-section {
    margin-bottom: 14px;

    .frame-desc-row {
      display: flex;
      flex-direction: column;
      gap: 10px;
    }

    .frame-desc-item {
      .frame-desc-label {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 12px;
        font-weight: 500;
        color: var(--text-secondary);
        margin-bottom: 6px;

        .frame-dot {
          width: 8px;
          height: 8px;
          border-radius: 50%;

          &.first { background: #409eff; }
          &.middle { background: #67c23a; }
          &.last { background: #e6a23c; }
        }
      }

      :deep(.el-textarea__inner) {
        font-size: 12px;
        line-height: 1.5;
      }
    }
  }

  .shot-description-editable {
    .desc-field {
      margin-bottom: 10px;
      &:last-child {
        margin-bottom: 0;
      }

      > label {
        display: block;
        font-size: 12px;
        color: var(--text-secondary);
        margin-bottom: 4px;
        font-weight: 500;
      }
    }
  }

  .dialogue-section {
    margin-bottom: 14px;
  }
}

// 场景选择对话框样式
.scene-selector-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  max-height: 500px;
  overflow-y: auto;
  padding: 10px;

  .scene-card {
    border: 2px solid var(--border-primary);
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      border-color: var(--accent);
      transform: translateY(-2px);
      box-shadow: var(--shadow-md);
    }

    &.selected {
      border-color: var(--accent);
      background: var(--accent-light);
    }

    .scene-image {
      width: 100%;
      height: 150px;
      background: var(--bg-secondary);
      display: flex;
      align-items: center;
      justify-content: center;

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }
    }

    .scene-info {
      padding: 12px;
      background: var(--bg-card);

      .scene-location {
        font-size: 14px;
        font-weight: 600;
        color: var(--text-primary);
        margin-bottom: 4px;
      }

      .scene-time {
        font-size: 12px;
        color: var(--text-muted);
      }
    }
  }

  .empty-scenes {
    grid-column: 1 / -1;
    padding: 40px 0;
  }
}

// 更新section-label样式以支持按钮
.section-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

// 角色选择对话框样式
.character-selector-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  max-height: 500px;
  overflow-y: auto;
  padding: 12px;

  .character-card {
    position: relative;
    border: 2px solid var(--border-primary);
    border-radius: 8px;
    padding: 16px;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;

    &:hover {
      border-color: var(--accent);
      transform: translateY(-2px);
      box-shadow: var(--shadow-md);
    }

    &.selected {
      border-color: var(--accent);
      background: var(--accent-light);
    }

    .character-avatar-large {
      width: 80px;
      height: 80px;
      border-radius: 50%;
      overflow: hidden;
      display: flex;
      align-items: center;
      justify-content: center;
      background: var(--bg-secondary);
      font-size: 32px;
      font-weight: 600;
      color: var(--accent);

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }
    }

    .character-info {
      text-align: center;

      .character-name {
        font-size: 14px;
        font-weight: 600;
        color: var(--text-primary);
        margin-bottom: 4px;
      }

      .character-role {
        font-size: 12px;
        color: var(--text-muted);
      }
    }

    .character-check {
      position: absolute;
      top: 8px;
      right: 8px;
    }
  }

  .empty-characters {
    grid-column: 1 / -1;
    padding: 40px 0;
  }
}

// 角色大图预览样式
.character-image-preview {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;

  img {
    max-width: 100%;
    max-height: 500px;
    border-radius: 8px;
    object-fit: contain;
  }
}

// 场景大图预览样式
.scene-image-preview {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 450px;
  background: var(--bg-secondary);
  border-radius: 8px;

  img {
    max-width: 100%;
    max-height: 600px;
    border-radius: 8px;
    object-fit: contain;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }
}

// 设置部分样式
.settings-section {
  margin-bottom: 20px;

  .section-label {
    font-size: 12px;
    color: var(--text-secondary);
    margin-bottom: 12px;
  }

  .settings-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;

    .setting-item {
      label {
        display: block;
        font-size: 11px;
        color: var(--text-secondary);
        margin-bottom: 6px;
      }
    }
  }

  .audio-controls {
    :deep(.el-textarea__inner) {
      background: var(--bg-card);
      border-color: var(--border-primary);
      color: var(--text-primary);

      &::placeholder {
        color: var(--text-muted);
      }
    }

    :deep(.el-select) {
      width: 100%;
    }

    :deep(.el-slider__runway) {
      background: #e4e7ed;
    }

    :deep(.el-slider__bar) {
      background: #409eff;
    }

    :deep(.el-slider__button) {
      border-color: #409eff;
    }
  }
}

.professional-editor {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  color: var(--text-primary);

  .editor-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 20px;
    background: var(--bg-card);
    border-bottom: 1px solid var(--border-primary);

    .toolbar-left {
      display: flex;
      align-items: center;
      gap: 12px;

      .back-btn {
        color: var(--text-secondary);

        &:hover {
          color: var(--accent);
        }
      }

      .episode-title {
        font-size: 14px;
        color: var(--text-primary);
      }
    }

    .toolbar-right {
      display: flex;
      gap: 8px;
    }
  }

  .editor-main {
    flex: 1;
    display: flex;
    overflow: hidden;
    height: calc(100vh - 60px);

    .storyboard-panel {
      width: 280px;
      background: var(--bg-card);
      border-right: 1px solid var(--border-primary);
      display: flex;
      flex-direction: column;

      .panel-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 16px;
        border-bottom: 1px solid var(--border-primary);

        h3 {
          margin: 0;
          font-size: 16px;
          font-weight: 500;
        }
      }

      .storyboard-list {
        flex: 1;
        overflow-y: auto;
        padding: 8px;

        .storyboard-item {
          display: flex;
          flex-direction: column;
          padding: 12px;
          margin-bottom: 8px;
          background: var(--bg-secondary);
          border-radius: 8px;
          cursor: pointer;
          transition: all 0.2s;

          &:hover {
            background: var(--bg-card-hover);
          }

          &.active {
            background: var(--accent-light);
            border-left: 3px solid var(--accent);

            .shot-content {
              .shot-number,
              .shot-title {
                color: var(--accent) !important;
              }

              .shot-action {
                color: var(--text-primary) !important;
              }

              .shot-duration {
                background: var(--accent-light);
                color: var(--accent);
              }
            }
          }

          .shot-content {
            width: 100%;

            .shot-header {
              display: flex;
              justify-content: space-between;
              align-items: center;
              margin-bottom: 6px;
              gap: 8px;

              .shot-title-row {
                display: flex;
                align-items: baseline;
                gap: 8px;
                flex: 1;
                min-width: 0;

                .shot-number {
                  font-size: 12px;
                  font-weight: 600;
                  color: var(--text-secondary);
                  flex-shrink: 0;
                }

                .shot-title {
                  font-size: 13px;
                  font-weight: 500;
                  color: var(--text-primary);
                  overflow: hidden;
                  text-overflow: ellipsis;
                  white-space: nowrap;
                }
              }

              .shot-duration {
                font-size: 11px;
                color: var(--text-muted);
                background: var(--bg-card-hover);
                padding: 2px 8px;
                border-radius: 4px;
                flex-shrink: 0;
              }
            }

            .shot-action {
              font-size: 11px;
              color: var(--text-secondary);
              line-height: 1.5;
              overflow: hidden;
              text-overflow: ellipsis;
              display: -webkit-box;
              -webkit-line-clamp: 2;
              -webkit-box-orient: vertical;
            }
          }
        }
      }
    }

    .timeline-area {
      flex: 1;
      display: flex;
      flex-direction: column;
      background: var(--bg-secondary);
      overflow: hidden;

      .empty-timeline {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: center;
      }
    }

    .edit-panel {
      width: 520px;
      background: var(--bg-card);
      border-left: 1px solid var(--border-primary);
      overflow: hidden;
      flex-shrink: 0;

      .edit-tabs {
        height: 100%;

        :deep(.el-tabs__header) {
          margin: 0;
          background: var(--bg-secondary);
          padding: 0 16px;
          border-bottom: 1px solid var(--border-primary);
        }

        :deep(.el-tabs__content) {
          height: calc(100% - 55px);
          overflow-y: auto;
        }

        .tab-content {
          padding: 16px;
        }

        .scene-editor,
        .shot-editor {
          .el-form-item {
            margin-bottom: 16px;
          }
        }
      }
    }
  }
}

// 通用参数行样式
.param-row {
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 8px;

  &:last-child {
    margin-bottom: 0;
  }
}

.param-label {
  min-width: 50px;
  font-size: 12px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

// 首帧/尾帧分页标签
.frame-type-tabs {
  display: flex;
  gap: 0;
  margin-bottom: 14px;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #e4e7ed;

  .frame-type-tab {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 10px 0;
    font-size: 13px;
    font-weight: 500;
    color: #606266;
    background: #fafafa;
    cursor: pointer;
    transition: all 0.2s;
    position: relative;
    user-select: none;

    &:not(:last-child) {
      border-right: 1px solid #e4e7ed;
    }

    &:hover {
      background: #f0f2f5;
    }

    &.active {
      color: #303133;
      background: #fff;
      font-weight: 600;
      box-shadow: 0 -2px 0 0 var(--accent) inset;
    }

    .frame-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      &.first { background: #409eff; }
      &.last { background: #e6a23c; }
    }

    .frame-count-badge {
      :deep(.el-badge__content) {
        font-size: 10px;
        height: 16px;
        line-height: 16px;
        padding: 0 5px;
      }
    }
  }
}

// 图片生成界面样式
.image-generation-section {
  .ref-assets-bar {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 8px 12px;
    margin-bottom: 12px;
    background: var(--bg-secondary, #f5f7fa);
    border-radius: 8px;
    flex-wrap: wrap;

    .ref-assets-group {
      display: flex;
      align-items: center;
      gap: 6px;

      .ref-assets-label {
        font-size: 11px;
        color: #909399;
        white-space: nowrap;
        margin-right: 2px;
      }
    }

    .ref-asset-circle {
      width: 30px;
      height: 30px;
      border-radius: 50%;
      border: 2px solid #dcdfe6;
      overflow: hidden;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #e8e8e8;
      cursor: default;
      transition: border-color 0.2s, opacity 0.2s;
      flex-shrink: 0;
      position: relative;

      &.selectable {
        cursor: pointer;
      }

      &.selectable:hover { border-color: #409eff; }

      &.selected {
        border-color: #409eff;
      }

      &:not(.selected) {
        opacity: 0.45;
      }

      &.disabled {
        cursor: not-allowed;
        opacity: 0.3;
      }

      img { width: 100%; height: 100%; object-fit: cover; }

      .ref-asset-initial {
        font-size: 11px;
        font-weight: 600;
        color: #606266;
      }

      .ref-check-badge {
        position: absolute;
        bottom: -2px;
        right: -2px;
        width: 14px;
        height: 14px;
        background: #409eff;
        color: #fff;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 9px;
        font-weight: bold;
        line-height: 1;
      }
    }

    .ref-assets-empty {
      font-size: 12px;
      color: #c0c4cc;
    }
  }

  .gen-params-card {
    margin-bottom: 12px;
    padding: 10px 14px;
    background: var(--bg-secondary, #f5f7fa);
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    gap: 8px;

    .param-row {
      display: flex;
      align-items: center;
      gap: 10px;

      .param-label {
        font-size: 12px;
        color: #909399;
        white-space: nowrap;
        min-width: 36px;
      }

      .param-value {
        font-size: 13px;
        color: #303133;
        min-width: 28px;
        text-align: right;
      }

      .param-hint {
        font-size: 12px;
        color: #c0c4cc;
      }

      .param-input-thumb {
        width: 40px;
        height: 28px;
        border-radius: 4px;
        border: 1.5px solid #67c23a;
        flex-shrink: 0;
        cursor: pointer;
      }
    }

    .param-inline-group {
      display: flex;
      gap: 14px;
      flex-wrap: wrap;
    }
  }

  .prompt-and-action {
    margin-bottom: 16px;

    .prompt-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 8px;

      .prompt-title {
        font-size: 13px;
        font-weight: 500;
        color: var(--text-primary);
      }
    }

    :deep(.el-textarea__inner) {
      font-family: "Monaco", "Menlo", "Consolas", monospace;
      font-size: 12px;
      line-height: 1.6;
    }

    .action-bar {
      display: flex;
      gap: 10px;
      margin-top: 10px;
    }
  }

  .generation-result {
    .section-label {
      font-size: 13px;
      color: var(--text-primary);
      font-weight: 600;
      margin-bottom: 12px;
      display: flex;
      align-items: center;
      gap: 6px;

      &::before {
        content: "";
        width: 3px;
        height: 14px;
        background: linear-gradient(
          to bottom,
          var(--accent),
          var(--accent-hover)
        );
        border-radius: 2px;
      }
    }

    .image-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
      gap: 10px;

      .image-item-wrapper {
        display: flex;
        flex-direction: column;
        gap: 4px;

        .image-meta {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 6px;
        }
      }

      .image-item {
        position: relative;
        border-radius: 8px;
        overflow: hidden;
        background: var(--bg-card);
        border: 1px solid var(--border-primary);
        transition: all 0.2s ease;
        cursor: pointer;
        box-shadow: var(--shadow-sm);
        height: 150px;

        &:hover {
          transform: translateY(-2px);
          box-shadow: var(--shadow-md);
          // border-color: var(--accent);

          .image-actions {
            transform: translateY(0);
            opacity: 0.7;
          }
        }

        :deep(.el-image) {
          width: 100%;
          aspect-ratio: 16 / 9;
          background: var(--bg-secondary);
          display: block;
          height: 100%;
        }

        .image-placeholder {
          width: 100%;
          aspect-ratio: 16 / 9;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          gap: 8px;
          background: var(--bg-secondary);
          color: var(--text-muted);
          position: relative;
          overflow: hidden;
          height: 100%;

          &::before {
            content: "";
            position: absolute;
            width: 200%;
            height: 200%;
            background: linear-gradient(
              45deg,
              transparent 30%,
              var(--border-secondary) 50%,
              transparent 70%
            );
            animation: shimmer 2s infinite;
            top: -50%;
            left: -50%;
          }

          .el-icon {
            position: relative;
            z-index: 1;
            font-size: 24px !important;
          }

          p {
            margin: 0;
            font-size: 11px;
            font-weight: 500;
            position: relative;
            z-index: 1;
          }
        }

        .image-actions {
          position: absolute;
          bottom: 0;
          left: 0;
          width: 100%;
          display: flex;
          justify-content: space-between;
          align-items: center;
          background-color: var(--bg-primary);
          opacity: 0;
          padding: 0 8px;
          height: 32px;
          transform: translateY(100%);
          transition:
            transform 0.3s ease,
            opacity 0.2s ease;

          .crop-icon-overlay,
          .delete-icon-overlay {
            width: 28px;
            height: 28px;
            display: flex;
            align-items: center;
            justify-content: center;
            border-radius: 4px;
            transition: all 0.2s ease;

            &:hover {
              cursor: pointer;
              background: var(--bg-secondary);
              transform: scale(1.1);
            }
          }
        }
      }

      .image-status {
        display: flex;
        justify-content: center;
        align-items: center;
        padding: 2px 0;

        :deep(.el-tag) {
          font-size: 10px;
          height: 20px;
          padding: 0 6px;
        }
      }
    }

    @keyframes shimmer {
      0% {
        transform: translateX(-100%) translateY(-100%) rotate(45deg);
      }

      100% {
        transform: translateX(100%) translateY(100%) rotate(45deg);
      }
    }
  }

  .panel-count-label {
    margin-left: 5px;
    font-size: 12px;
    color: var(--text-muted);
  }

  .model-tags {
    font-size: 12px;
    color: var(--text-muted);
  }

  .mode-description {
    font-size: 12px;
    color: var(--text-muted);
  }
}

// 视频生成样式
.video-generation-section {
  .ref-assets-bar {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 8px 12px;
    margin-bottom: 12px;
    background: var(--bg-secondary, #f5f7fa);
    border-radius: 8px;
    flex-wrap: wrap;

    .ref-assets-group {
      display: flex;
      align-items: center;
      gap: 6px;

      .ref-assets-label {
        font-size: 11px;
        color: #909399;
        white-space: nowrap;
        margin-right: 2px;
      }
    }

    .ref-asset-circle {
      width: 30px;
      height: 30px;
      border-radius: 50%;
      border: 2px solid #dcdfe6;
      overflow: hidden;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #e8e8e8;
      flex-shrink: 0;
      transition: border-color 0.2s;

      &:hover { border-color: #409eff; }

      img { width: 100%; height: 100%; object-fit: cover; }

      .ref-asset-initial {
        font-size: 11px;
        font-weight: 600;
        color: #606266;
      }
    }
  }

  .gen-params-card {
    margin-bottom: 12px;
    padding: 10px 14px;
    background: var(--bg-secondary, #f5f7fa);
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    gap: 8px;

    .param-row {
      display: flex;
      align-items: center;
      gap: 10px;

      .param-label {
        font-size: 12px;
        color: #909399;
        white-space: nowrap;
        min-width: 36px;
      }

      .param-value {
        font-size: 13px;
        color: #303133;
        min-width: 28px;
        text-align: right;
      }

      .param-hint {
        font-size: 12px;
        color: #c0c4cc;
      }

      .param-input-thumb {
        width: 40px;
        height: 28px;
        border-radius: 4px;
        border: 1.5px solid #67c23a;
        flex-shrink: 0;
        cursor: pointer;
      }
    }

    .param-inline-group {
      display: flex;
      gap: 14px;
      flex-wrap: wrap;
    }

  }

  .video-frame-compare {
    display: flex;
    align-items: stretch;
    gap: 6px;
    margin-bottom: 12px;

    .video-frame-card {
      flex: 1;
      min-width: 0;
      background: var(--bg-secondary, #f5f7fa);
      border-radius: 8px;
      padding: 8px;
      display: flex;
      flex-direction: column;
      gap: 6px;

      .video-frame-header {
        display: flex;
        align-items: center;
        gap: 4px;
        font-size: 12px;
        font-weight: 500;
        color: #606266;
      }

      .video-frame-preview {
        width: 100%;
        aspect-ratio: 16 / 9;
        border-radius: 6px;
        overflow: hidden;
        border: 1.5px solid #dcdfe6;
        background: #000;

        :deep(.el-image) {
          width: 100%;
          height: 100%;
        }
      }

      .video-frame-placeholder {
        width: 100%;
        aspect-ratio: 16 / 9;
        border-radius: 6px;
        border: 1.5px dashed #dcdfe6;
        background: #fafafa;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 12px;
        color: #c0c4cc;
      }
    }

    .video-frame-arrow-center {
      display: flex;
      align-items: center;
      font-size: 18px;
      color: #c0c4cc;
      flex-shrink: 0;
      padding-top: 20px;
    }
  }

  .prompt-and-action {
    margin-bottom: 16px;

    .prompt-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 8px;

      .prompt-title {
        font-size: 13px;
        font-weight: 500;
        color: var(--text-primary);
      }
    }

    :deep(.el-textarea__inner) {
      font-family: "Monaco", "Menlo", "Consolas", monospace;
      font-size: 12px;
      line-height: 1.6;
    }

    .action-bar {
      display: flex;
      gap: 10px;
      margin-top: 10px;
    }
  }

  .section-label {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 12px;
  }

  .generation-result {
    margin-top: 16px;

    .section-label {
      font-size: 13px;
      color: #303133;
      font-weight: 600;
      margin-bottom: 12px;
      display: flex;
      align-items: center;
      gap: 6px;

      // &::before {
      //   content: '';
      //   width: 3px;
      //   height: 14px;
      //   background: linear-gradient(to bottom, #409eff, #66b1ff);
      //   border-radius: 2px;
      // }
    }

    .image-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
      gap: 10px;

      .image-item {
        position: relative;
        border-radius: 8px;
        overflow: hidden;
        background: #fff;
        border: 1px solid #e8e8e8;
        transition: all 0.2s ease;
        cursor: pointer;
        height: 150px;
        box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);

        &:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);

          .video-actions {
            transform: translateY(0);
            opacity: 0.7;
          }
        }

        .image-placeholder {
          width: 100%;
          aspect-ratio: 16 / 9;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          gap: 8px;
          background: linear-gradient(135deg, #f5f7fa 0%, #e8ecf0 100%);
          color: #909399;
          position: relative;
          overflow: hidden;
          height: 100%;

          &::before {
            content: "";
            position: absolute;
            width: 200%;
            height: 200%;
            background: linear-gradient(
              45deg,
              transparent 30%,
              var(--border-secondary) 50%,
              transparent 70%
            );
            animation: shimmer 2s infinite;
            top: -50%;
            left: -50%;
          }

          .el-icon {
            position: relative;
            z-index: 1;
            font-size: 24px !important;
          }

          p {
            margin: 0;
            font-size: 11px;
            font-weight: 500;
            position: relative;
            z-index: 1;
          }
        }

        .image-info {
          position: absolute;
          bottom: 0;
          left: 0;
          right: 0;
          padding: 6px 8px;
          background: linear-gradient(
            to top,
            rgba(0, 0, 0, 0.75),
            rgba(0, 0, 0, 0.2) 70%,
            transparent
          );
          display: flex;
          justify-content: space-between;
          align-items: center;
          gap: 4px;

          :deep(.el-tag) {
            backdrop-filter: blur(8px);
            font-size: 10px;
            height: 20px;
            padding: 0 6px;
          }

          .frame-type-tag {
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 10px;
            font-weight: 500;
            background: rgba(255, 255, 255, 0.25);
            color: white;
            backdrop-filter: blur(8px);
            border: 1px solid rgba(255, 255, 255, 0.3);
            text-transform: uppercase;
            letter-spacing: 0.3px;
          }
        }

        // 视频缩略图特殊样式
        &.video-item .video-thumbnail {
          position: relative;
          width: 100%;
          height: 100%;
          overflow: hidden;
          cursor: pointer;

          video {
            width: 100%;
            height: 100%;
            object-fit: cover;
            display: block;
            pointer-events: none;
          }

          .play-overlay {
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            background: rgba(0, 0, 0, 0.3);
            opacity: 0;
            transition: opacity 0.2s ease;

            .el-icon {
              filter: drop-shadow(0 2px 8px rgba(0, 0, 0, 0.3));
            }
          }

          &:hover .play-overlay {
            opacity: 1;
          }
        }

        .video-actions {
          position: absolute;
          bottom: 0;
          left: 0;
          width: 100%;
          display: flex;
          justify-content: space-between;
          align-items: center;
          background-color: var(--bg-primary);
          opacity: 0;
          padding: 0 8px;
          height: 32px;
          transform: translateY(100%);
          transition:
            transform 0.3s ease,
            opacity 0.2s ease;

          .add-to-assets-button,
          .delete-video-button {
            width: 28px;
            height: 28px;
            display: flex;
            align-items: center;
            justify-content: center;
            cursor: pointer;
            border-radius: 4px;
            transition: all 0.2s ease;

            &:hover {
              background: var(--bg-secondary);
              transform: scale(1.1);
            }

            .is-loading {
              animation: rotate 1s linear infinite;
            }
          }
        }
      }
    }
  }

  .reference-mode-title {
    margin-bottom: 12px;
    font-size: 13px;
    color: var(--text-primary);
    font-weight: 500;
  }

  .frame-label {
    margin-bottom: 8px;
    font-size: 12px;
    color: var(--text-muted);
  }

  .slot-hint {
    margin-top: 8px;
    font-size: 12px;
    color: var(--text-muted);
  }

  .image-slot {
    position: relative;
    width: 140px;
    height: 90px;
    border: 2px dashed var(--border-primary);
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    background: var(--bg-card);
    // display: flex;
    // align-items: center;
    // justify-content: center;

    &:hover {
      border-color: var(--accent);
    }
  }

  .video-params-section {
    margin-bottom: 16px;
    padding: 12px 16px;
    background: var(--bg-secondary);
    border-radius: 8px;
    border: 1px solid var(--border-primary);
  }

  .image-slots-container {
    padding: 12px;
    background: var(--bg-secondary);
    border-radius: 8px;
    border: 1px dashed var(--border-primary);
  }

  .image-slot {
    position: relative;
    width: 140px;
    height: 90px;
    border: 2px dashed var(--border-primary);
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    transition: all 0.3s;
    background: var(--bg-card);

    &:hover {
      border-color: var(--accent);
      box-shadow: var(--shadow-md);
    }

    &.image-slot-small {
      width: 80px;
      height: 52px;
    }
  }

  .image-slot-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
  }

  .image-slot-remove {
    position: absolute;
    top: 4px;
    right: 4px;
    width: 24px;
    height: 24px;
    background: rgba(0, 0, 0, 0.6);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      background: rgba(255, 73, 73, 0.9);
      transform: scale(1.1);
    }
  }

  .reference-images-section {
    margin-top: 12px;

    .frame-type-buttons {
      margin-bottom: 12px;
      text-align: center;

      :deep(.el-radio-group) {
        display: inline-flex;
        flex-wrap: wrap;
        gap: 0;
      }

      :deep(.el-radio-button) {
        overflow: hidden;

        &:first-child .el-radio-button__inner {
          border-radius: 6px 0 0 6px;
        }

        &:last-child .el-radio-button__inner {
          border-radius: 0 6px 6px 0;
        }
      }

      :deep(.el-radio-button__inner) {
        padding: 6px 12px;
        font-size: 12px;
        font-weight: 500;
        border-color: var(--border-primary);
        transition: all 0.2s;

        &:hover {
          // color: var(--accent);
          border-color: var(--accent);
        }
      }

      :deep(.el-radio-button.is-active .el-radio-button__inner) {
        background: var(--accent);
        border-color: var(--accent);
        box-shadow: 0 2px 6px rgba(14, 165, 233, 0.25);
      }
    }

    .frame-type-content {
      padding: 4px 10px;
      background: var(--bg-card);
      border-radius: 8px;
      border: 1px solid var(--border-primary);
      min-height: 160px;
    }

    .image-scroll-container {
      max-height: 220px;
      overflow-y: auto;
      overflow-x: hidden;
      padding-right: 4px;

      &::-webkit-scrollbar {
        width: 6px;
      }

      &::-webkit-scrollbar-track {
        background: #f1f1f1;
        border-radius: 3px;
      }

      &::-webkit-scrollbar-thumb {
        background: #c1c1c1;
        border-radius: 3px;

        &:hover {
          background: #a8a8a8;
        }
      }
    }

    .previous-frame-section {
      margin-bottom: 12px;
      padding: 8px;
      background: var(--bg-secondary);
      border: 1px solid var(--border-primary);
      border-radius: 6px;

      .hint-text {
        color: var(--text-muted);
        font-size: 11px;
      }
    }

    .reference-grid {
      display: grid !important;
      grid-template-columns: repeat(4, 1fr) !important;
      gap: 8px !important;

      .reference-item {
        // padding-top: 4px;
        margin-top: 6px;
        position: relative;
        border-radius: 6px;
        overflow: hidden;
        cursor: pointer;
        border: 2px solid transparent;
        transition: all 0.2s ease;
        width: 100% !important;
        max-width: 120px !important;
        background: var(--bg-card);

        &:hover {
          transform: translateY(-4px) scale(1.02);
          box-shadow: var(--shadow-lg);
          border-color: var(--accent);
        }

        &.selected {
          border-color: var(--accent);
          box-shadow: var(--shadow-glow);
        }

        img {
          width: 100%;
          max-width: 180px;
          aspect-ratio: 16 / 9;
          object-fit: cover;
          display: block;
          transition: transform 0.3s;
        }

        &:hover img {
          transform: scale(1.05);
        }

        .reference-label {
          position: absolute;
          bottom: 0;
          left: 0;
          right: 0;
          padding: 4px 8px;
          background: linear-gradient(to top, rgba(0, 0, 0, 0.7), transparent);
          color: white;
          font-size: 10px;
          text-align: center;
        }
      }
    }
  }

  .generation-controls {
    margin-top: 40px;
    padding-top: 0;
    text-align: center;

    :deep(.el-button) {
      padding: 12px 32px;
      font-size: 14px;
      font-weight: 500;
      border-radius: 8px;
      transition: all 0.3s;

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
      }
    }
  }

  @keyframes shimmer {
    0% {
      transform: translateX(-100%) translateY(-100%) rotate(45deg);
    }

    100% {
      transform: translateX(100%) translateY(100%) rotate(45deg);
    }
  }
}

// 视频合成列表样式
.merges-list {
  min-height: 300px;

  .merge-items {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .merge-item {
    position: relative;
    background: var(--bg-card);
    border-radius: 12px;
    overflow: hidden;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: 1px solid var(--border-primary);
    box-shadow: var(--shadow-sm);

    &:hover {
      transform: translateY(-2px);
      box-shadow: var(--shadow-md);
      border-color: var(--accent-light);
    }

    // 状态指示条
    .status-indicator {
      position: absolute;
      left: 0;
      top: 0;
      bottom: 0;
      width: 4px;
      transition: all 0.3s ease;
    }

    &.merge-status-pending .status-indicator {
      background: linear-gradient(to bottom, #909399, #b1b3b8);
    }

    &.merge-status-processing .status-indicator {
      background: linear-gradient(to bottom, #e6a23c, #f0c78a);
      animation: pulse 2s ease-in-out infinite;
    }

    &.merge-status-completed .status-indicator {
      background: linear-gradient(to bottom, #67c23a, #95d475);
    }

    &.merge-status-failed .status-indicator {
      background: linear-gradient(to bottom, #f56c6c, #f89898);
    }

    .merge-content {
      padding: 20px 20px 20px 24px;
    }

    .merge-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 16px;
      gap: 12px;

      .title-section {
        display: flex;
        align-items: center;
        gap: 12px;
        flex: 1;
        min-width: 0;

        .title-icon {
          color: #409eff;
          flex-shrink: 0;

          &.rotating {
            animation: rotate 1.5s linear infinite;
          }
        }

        .merge-title {
          margin: 0;
          font-size: 16px;
          font-weight: 600;
          color: var(--text-secondary);
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
      }

      :deep(.el-tag) {
        flex-shrink: 0;
        font-weight: 500;
        letter-spacing: 0.3px;
      }
    }

    .merge-details {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
      gap: 16px;
      margin-bottom: 16px;
      padding: 16px;
      background: var(--bg-secondary);
      border-radius: 8px;
      border: 1px solid var(--border-primary);

      .detail-item {
        display: flex;
        align-items: flex-start;
        gap: 10px;

        .detail-icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 32px;
          height: 32px;
          background: var(--bg-card);
          border-radius: 8px;
          color: var(--accent);
          flex-shrink: 0;
          box-shadow: var(--shadow-xs);
        }

        .detail-content {
          flex: 1;
          min-width: 0;

          .detail-label {
            font-size: 12px;
            color: var(--text-muted);
            margin-bottom: 4px;
            font-weight: 500;
          }

          .detail-value {
            font-size: 14px;
            color: var(--text-primary);
            font-weight: 500;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
          }
        }
      }
    }

    .merge-error {
      margin-bottom: 16px;

      :deep(.el-alert) {
        border-radius: 8px;
        border-left: 4px solid #f56c6c;
      }
    }

    .merge-actions {
      display: flex;
      gap: 10px;
      flex-wrap: wrap;
      padding-top: 16px;
      border-top: 1px solid var(--border-primary);

      :deep(.el-button) {
        font-weight: 500;
        transition: all 0.3s ease;

        &:hover {
          transform: translateY(-1px);
        }

        &.el-button--primary {
          box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);

          &:hover {
            box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
          }
        }
      }
    }
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
    }

    50% {
      opacity: 0.6;
    }
  }

  @keyframes rotate {
    from {
      transform: rotate(0deg);
    }

    to {
      transform: rotate(360deg);
    }
  }
}

.video-meta {
  margin-top: 16px;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid var(--border-primary);
  background: var(--bg-secondary);
}

.first-frame-selector {
  border-top: 1px solid #e4e7ed;
  padding-top: 8px;

  .first-frame-grid {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;

    .first-frame-thumb {
      width: 100px;
      height: 100px;
      border-radius: 4px;
      border: 2px solid transparent;
      cursor: pointer;
      position: relative;
      overflow: hidden;
      transition: border-color 0.2s;

      &:hover {
        border-color: #409eff80;
      }

      &.selected {
        border-color: #409eff;
      }

      .selected-badge {
        position: absolute;
        top: 1px;
        right: 1px;
        width: 14px;
        height: 14px;
        background: #409eff;
        color: #fff;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 9px;
        font-weight: bold;
      }
    }
  }
}
</style>
<style>
.video-prompt-box {
  margin-bottom: 10px;
  padding: 8px 10px;
  background: var(--bg-secondary);
  border-radius: 6px;
  border: 1px solid var(--border-primary);
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-secondary);
  word-break: break-word;
  max-height: 300px;
  overflow-y: auto;
}

/* 动作序列图片裁剪图标样式 */
.action-image-item {
  position: relative;
}

/* .crop-icon-overlay {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 28px;
  height: 28px;
  background: rgba(0, 0, 0, 0.7);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  opacity: 0;
  transition: all 0.3s ease;
  z-index: 10;
} */

.action-image-item:hover .crop-icon-overlay {
  opacity: 1;
}

.crop-icon-overlay:hover {
  /* background: var(--accent); */
  transform: scale(1.1);
}

/* 删除按钮样式 */
/* .delete-icon-overlay {
  position: absolute;
  bottom: 4px;
  left: 4px;
  width: 28px;
  height: 28px;
  background: rgba(220, 38, 38, 0.9);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  opacity: 0;
  transition: all 0.3s ease;
  z-index: 10;
} */

.image-item:hover .delete-icon-overlay {
  opacity: 1;
}

.delete-icon-overlay:hover {
  /* background: rgba(220, 38, 38, 1); */
  transform: scale(1.1);
}

/* 宫格图片入口按钮样式 */
.grid-entry-button {
  cursor: pointer;
  transition: all 0.3s ease;
  border: none !important;
  min-height: 88px;
}

.grid-entry-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(14, 165, 233, 0.3);
}

.grid-entry-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  border: 2px dashed var(--border-primary);
  border-radius: 8px;
  background: var(--bg-secondary);
  transition: all 0.3s ease;
}

.grid-entry-button:hover .grid-entry-placeholder {
  border-color: var(--accent);
  background: var(--bg-card);
}

/* 宫格图片编辑器样式 */
.creation-mode-selector,
.grid-type-selector {
  margin-bottom: 16px;
}

.grid-editor {
  margin-bottom: 20px;
}

.grid-container {
  display: grid;
  gap: 12px;
  margin-bottom: 16px;
  padding: 16px;
  background: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-primary);
}

.grid-container.grid-4 {
  grid-template-columns: repeat(2, 1fr);
}

.grid-container.grid-6 {
  grid-template-columns: repeat(3, 1fr);
}

.grid-container.grid-9 {
  grid-template-columns: repeat(3, 1fr);
}

.grid-cell {
  position: relative;
  aspect-ratio: 1;
  border: 2px dashed var(--border-primary);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  background: var(--bg-card);
}

.grid-cell:hover {
  border-color: var(--accent);
  box-shadow: 0 2px 8px rgba(14, 165, 233, 0.2);
}

.grid-cell img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.grid-cell-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-secondary);
}

.grid-cell-placeholder p {
  margin-top: 8px;
  font-size: 12px;
}

.grid-cell-actions {
  position: absolute;
  top: 8px;
  right: 8px;
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.grid-cell:hover .grid-cell-actions {
  opacity: 1;
}

.grid-cell-actions .el-icon {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.6);
  border-radius: 4px;
  color: white;
  cursor: pointer;
  transition: all 0.3s ease;
}

.grid-cell-actions .el-icon:hover {
  background: rgba(0, 0, 0, 0.8);
  transform: scale(1.1);
}

.grid-controls {
  display: flex;
  gap: 12px;
}

/* 图片选择器样式 */
.image-selector-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 12px;
  max-height: 500px;
  overflow-y: auto;
  padding: 12px;
}

.image-selector-item {
  position: relative;
  cursor: pointer;
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.3s ease;
  border: 2px solid transparent;
}

.image-selector-item:hover {
  border-color: var(--accent);
  box-shadow: 0 2px 8px rgba(14, 165, 233, 0.3);
  transform: translateY(-2px);
}

.image-selector-label {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 4px 8px;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  font-size: 12px;
  text-align: center;
}

.grid-preview-container {
  text-align: center;
}

.grid-preview-container img {
  max-width: 100%;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
</style>

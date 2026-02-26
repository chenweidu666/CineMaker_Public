<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <AppHeader :fixed="false" :show-logo="false">
        <template #left>
          <el-button text @click="goBack" class="back-btn">
            <el-icon><ArrowLeft /></el-icon>
            <span>{{ $t("workflow.backToProject") }}</span>
          </el-button>
          <h1 class="header-title">
            {{ $t("workflow.episodeProduction", { number: episodeNumber }) }}
          </h1>
        </template>
        <template #center>
          <div class="custom-steps">
            <div
              class="step-item clickable"
              :class="{ active: currentStep >= 0, current: currentStep === 0 }"
              @click="currentStep = 0"
            >
              <div class="step-circle">1</div>
              <span class="step-text">{{ $t("workflow.steps.content") }}</span>
            </div>
            <el-icon class="step-arrow"><ArrowRight /></el-icon>
            <div
              class="step-item clickable"
              :class="{ active: currentStep >= 1, current: currentStep === 1 }"
              @click="currentStep = 1"
            >
              <div class="step-circle">2</div>
              <span class="step-text">{{ $t("workflow.steps.scriptDesign") }}</span>
            </div>
            <el-icon class="step-arrow"><ArrowRight /></el-icon>
            <div
              class="step-item clickable"
              :class="{ active: currentStep >= 2, current: currentStep === 2 }"
              @click="currentStep = 2"
            >
              <div class="step-circle">3</div>
              <span class="step-text">{{ $t("workflow.steps.generateStoryboard") }}</span>
            </div>
          </div>
        </template>
      </AppHeader>

      <div class="content-container">
        <!-- 阶段 0: 资源定义 -->
        <el-card
          v-show="currentStep === 0"
          shadow="never"
          class="stage-card stage-card-fullscreen"
        >
          <div class="stage-body stage-body-fullscreen">
            <!-- 角色定义 -->
            <div class="resource-section">
              <div class="section-header">
                <div class="section-title">
                  <h3>
                    <el-icon><User /></el-icon>
                    角色定义
                  </h3>
                  <el-alert type="info" :closable="false" style="margin: 0">
                    共 {{ charactersCount }} 个角色
                  </el-alert>
                </div>
                <div class="section-actions">
                  <el-button
                    :icon="FolderOpened"
                    @click="openCharacterLibraryDialog"
                    size="default"
                  >
                    从角色库选择
                  </el-button>
                </div>
              </div>

              <div class="character-image-list">
                <div
                  v-for="char in currentEpisode?.characters"
                  :key="char.id"
                  class="character-item"
                >
                  <el-card shadow="hover" class="fixed-card character-card">
                    <div class="card-header">
                      <div class="header-left">
                        <h4>{{ char.name }}</h4>
                        <el-tag size="small">{{ char.role }}</el-tag>
                      </div>
                      <el-button
                        type="warning"
                        size="small"
                        :icon="Close"
                        circle
                        @click="removeCharacterFromEpisode(char.id)"
                        title="从章节移除"
                      />
                    </div>

                    <div class="card-image-container">
                      <ImagePreview
                        :image-url="hasImage(char) ? getImageUrl(char) : ''"
                        :alt="char.name"
                        :size="120"
                        dialog-width="900px"
                      >
                        <template #details>
                          <div v-if="char.role" class="detail-section">
                            <div class="detail-label">身份</div>
                            <div class="detail-value">
                              <el-tag :type="char.role === 'main' ? 'danger' : 'info'" size="small">
                                {{ char.role === 'main' ? '主角' : char.role === 'supporting' ? '配角' : char.role }}
                              </el-tag>
                            </div>
                          </div>
                          <CollapsibleText v-if="char.appearance" label="外貌" :text="char.appearance" default-collapsed />
                          <CollapsibleText v-if="char.personality" label="性格" :text="char.personality" />
                          <CollapsibleText v-if="char.voice_style" label="声音" :text="char.voice_style" />
                          <CollapsibleText v-if="char.description" label="描述" :text="char.description" />
                        </template>
                      </ImagePreview>
                    </div>
                  </el-card>
                </div>
              </div>
            </div>

            <el-divider />

            <!-- 场景定义 -->
            <div class="resource-section">
              <div class="section-header">
                <div class="section-title">
                  <h3>
                    <el-icon><Location /></el-icon>
                    场景定义
                  </h3>
                  <el-alert type="info" :closable="false" style="margin: 0">
                    共 {{ currentEpisode?.scenes?.length || 0 }} 个场景
                  </el-alert>
                </div>
                <div class="section-actions">
                  <el-button
                    :icon="FolderOpened"
                    @click="openSceneLibraryDialog"
                    size="default"
                  >
                    从场景库选择
                  </el-button>
                </div>
              </div>

              <div class="scene-image-list">
                <div
                  v-for="scene in currentEpisode?.scenes"
                  :key="scene.id"
                  class="scene-item"
                >
                  <el-card shadow="hover" class="fixed-card scene-card">
                    <div class="card-header">
                      <div class="header-left">
                        <h4>{{ scene.name || scene.location || '未命名场景' }}</h4>
                      </div>
                      <el-button
                        type="warning"
                        size="small"
                        :icon="Close"
                        circle
                        @click="removeSceneFromEpisode(scene.id)"
                        title="从章节移除"
                      />
                    </div>

                    <div class="card-image-container">
                      <ImagePreview
                        :image-url="scene.image_url || getFirstRefImage(scene) || ''"
                        :alt="scene.name || scene.location || '场景'"
                        :size="120"
                        dialog-width="900px"
                      >
                        <template #details>
                          <div v-if="scene.location" class="detail-section">
                            <div class="detail-label">位置</div>
                            <div class="detail-value">{{ scene.location }}</div>
                          </div>
                          <div v-if="scene.time" class="detail-section">
                            <div class="detail-label">时间</div>
                            <div class="detail-value">{{ scene.time }}</div>
                          </div>
                          <CollapsibleText v-if="scene.description" label="描述" :text="scene.description" />
                        </template>
                      </ImagePreview>
                      <span v-if="!scene.image_url && getFirstRefImage(scene)" class="ref-badge">参考</span>
                    </div>
                  </el-card>
                </div>
              </div>
            </div>

            <el-divider />

            <!-- 道具定义 -->
            <div class="resource-section">
              <div class="section-header">
                <div class="section-title">
                  <h3>
                    <el-icon><Box /></el-icon>
                    道具定义
                  </h3>
                  <el-alert type="info" :closable="false" style="margin: 0">
                    共 {{ episodeProps?.length || 0 }} 个道具
                  </el-alert>
                </div>
                <div class="section-actions">
                  <el-button
                    :icon="FolderOpened"
                    @click="openPropLibraryDialog"
                    size="default"
                  >
                    从道具库选择
                  </el-button>
                </div>
              </div>

              <div class="prop-image-list">
                <div
                  v-for="prop in episodeProps"
                  :key="prop.id"
                  class="prop-item"
                >
                  <el-card shadow="hover" class="fixed-card prop-card">
                    <div class="card-header">
                      <div class="header-left">
                        <h4>{{ prop.name }}</h4>
                        <el-tag v-if="prop.type" size="small">{{ prop.type }}</el-tag>
                      </div>
                      <el-button
                        type="warning"
                        size="small"
                        :icon="Close"
                        circle
                        @click="removePropFromEpisode(prop.id)"
                        title="从章节移除"
                      />
                    </div>

                    <div class="card-image-container">
                      <ImagePreview
                        :image-url="prop.image_url || getFirstRefImage(prop) || ''"
                        :alt="prop.name || '道具'"
                        :size="120"
                        dialog-width="900px"
                      >
                        <template #details>
                          <div v-if="prop.type" class="detail-section">
                            <div class="detail-label">类型</div>
                            <div class="detail-value">
                              <el-tag size="small">{{ prop.type }}</el-tag>
                            </div>
                          </div>
                          <CollapsibleText v-if="prop.description" label="描述" :text="prop.description" />
                        </template>
                      </ImagePreview>
                      <span v-if="!prop.image_url && getFirstRefImage(prop)" class="ref-badge">参考</span>
                    </div>
                  </el-card>
                </div>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 阶段 1: 剧本设计 -->
        <el-card v-show="currentStep === 1" shadow="never" class="stage-card">
          <div class="stage-body">
            <!-- 资源概览（可折叠） -->
            <el-collapse class="script-resource-collapse">
              <el-collapse-item>
                <template #title>
                  <div class="resource-collapse-title">
                    <span>已预设的资源</span>
                    <el-tag size="small" type="success" effect="plain">角色 {{ charactersCount }}</el-tag>
                    <el-tag size="small" type="warning" effect="plain">场景 {{ currentEpisode?.scenes?.length || 0 }}</el-tag>
                    <el-tag size="small" type="primary" effect="plain">道具 {{ episodeProps?.length || 0 }}</el-tag>
                  </div>
                </template>
                <div class="resource-tag-groups">
                  <div v-if="currentEpisode?.characters?.length" class="resource-tag-group">
                    <span class="resource-tag-label">角色</span>
                    <el-tag v-for="char in currentEpisode?.characters" :key="'c-' + char.id" size="small" type="info">
                      {{ char.name }}
                    </el-tag>
                  </div>
                  <div v-if="currentEpisode?.scenes?.length" class="resource-tag-group">
                    <span class="resource-tag-label">场景</span>
                    <el-tag v-for="scene in currentEpisode?.scenes" :key="'s-' + scene.id" size="small" type="warning">
                      {{ scene.name || scene.location || '未命名' }}
                    </el-tag>
                  </div>
                  <div v-if="episodeProps?.length" class="resource-tag-group">
                    <span class="resource-tag-label">道具</span>
                    <el-tag v-for="prop in episodeProps" :key="'p-' + prop.id" size="small" type="primary">
                      {{ prop.name }}
                    </el-tag>
                  </div>
                </div>
              </el-collapse-item>
            </el-collapse>

            <!-- 剧情输入 -->
            <div class="script-design-section">
              <div class="script-input-header">
                <h3>
                  <el-icon><EditPen /></el-icon>
                  剧情描述
                </h3>
                <el-button size="small" type="primary" text @click="saveDescription" :disabled="!scriptDraftInput.trim()">
                  <el-icon><Check /></el-icon>
                  保存
                </el-button>
              </div>

              <el-input
                v-model="scriptDraftInput"
                type="textarea"
                :rows="5"
                placeholder="请输入剧情构想，例如：&#10;姜小卷的普通周一。早上闹钟响赖床五分钟，换通勤装出门走梧桐路坐地铁上班..."
                style="margin-bottom: 12px;"
              />

              <!-- 工具栏：模型 + 时长 + 操作按钮 -->
              <div class="script-toolbar">
                <div class="script-toolbar-left">
                  <el-select v-model="selectedTextModel" placeholder="文本模型" size="default" style="width: 260px">
                    <el-option v-for="model in textModels" :key="model.modelName" :label="model.modelName" :value="model.modelName" />
                  </el-select>
                  <el-select v-model="episodeDuration" size="default" style="width: 110px">
                    <el-option v-for="d in durationOptions" :key="d" :label="`${d}秒`" :value="d" />
                  </el-select>
                </div>
                <div class="script-toolbar-right">
                  <el-button type="primary" @click="generateScript" :loading="generatingScript" :disabled="!scriptDraftInput.trim()">
                    <el-icon><MagicStick /></el-icon>
                    AI 生成剧本
                  </el-button>
                  <el-button @click="submitScriptDirectly" :disabled="!scriptDraftInput.trim()">
                    <el-icon><Check /></el-icon>
                    直接提交
                  </el-button>
                  <el-button
                    v-if="currentEpisode?.script_content"
                    type="warning"
                    plain
                    @click="showFeedbackDialog = true; scriptFeedback = ''"
                  >
                    <el-icon><ChatLineSquare /></el-icon>
                    修改意见
                  </el-button>
                </div>
              </div>

              <!-- 反馈重新生成对话框 -->
              <el-dialog v-model="showFeedbackDialog" title="修改意见" width="600px" :close-on-click-modal="false">
                <p style="color: #909399; font-size: 13px; margin-bottom: 12px;">
                  请描述对当前剧本不满意的地方，AI 将根据您的意见重新生成。
                </p>
                <el-input v-model="scriptFeedback" type="textarea" :rows="5" placeholder="例如：场景名太冗长，应该只写 location 值；承接上镜太笼统，需要写具体切入哪个场景..." />
                <template #footer>
                  <el-button @click="showFeedbackDialog = false">取消</el-button>
                  <el-button type="primary" :loading="generatingScript" :disabled="!scriptFeedback.trim()" @click="regenerateWithFeedback">重新生成</el-button>
                </template>
              </el-dialog>
            </div>

            <!-- 生成的剧本预览 -->
            <div v-if="currentEpisode?.script_content" class="script-preview-section">
              <div class="script-preview-header">
                <div class="script-preview-title">
                  <h3>
                    <el-icon><Document /></el-icon>
                    剧本预览
                  </h3>
                  <div class="script-stats">
                    <el-tag size="small" effect="plain">{{ scriptShotCount }} 个镜头</el-tag>
                    <el-tag size="small" effect="plain" type="success">{{ scriptTotalDuration }}秒</el-tag>
                    <el-tag size="small" effect="plain" type="info">{{ currentEpisode.script_content.length }} 字</el-tag>
                  </div>
                </div>
                <div class="script-preview-actions">
                  <el-button text type="primary" size="small" @click="scriptEditMode = !scriptEditMode">
                    <el-icon><EditPen v-if="!scriptEditMode" /><View v-else /></el-icon>
                    {{ scriptEditMode ? '预览' : '编辑' }}
                  </el-button>
                </div>
              </div>

              <div v-if="scriptEditMode" class="script-edit-area">
                <el-input v-model="currentEpisode.script_content" type="textarea" :rows="20" />
                <div style="display: flex; justify-content: flex-end; margin-top: 10px; gap: 8px;">
                  <el-button @click="scriptEditMode = false">取消</el-button>
                  <el-button type="primary" @click="saveScriptContent(); scriptEditMode = false">
                    <el-icon><Check /></el-icon>
                    保存剧本
                  </el-button>
                </div>
              </div>
              <div v-else class="script-preview-box">
                <pre style="white-space: pre-wrap; word-break: break-word; font-size: 13px; line-height: 1.8; margin: 0;">{{ currentEpisode.script_content }}</pre>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 阶段 2: 分镜生成 -->
        <el-card v-show="currentStep === 2" shadow="never" class="stage-card">
          <div class="stage-body">
            <!-- 剧本摘要 -->
            <el-alert v-if="currentEpisode?.script_content" type="success" :closable="false" style="margin-bottom: 16px;">
              <template #title>
                <span>剧本已就绪（{{ currentEpisode.script_content.length }} 字）</span>
              </template>
            </el-alert>
            <el-alert v-else type="warning" :closable="false" style="margin-bottom: 16px;">
              <template #title>
                <span>请先在「剧本设计」步骤中生成剧本</span>
              </template>
            </el-alert>

            <!-- 分镜配置 -->
            <div class="model-selector-bar">
              <div class="model-selector-item">
                <span class="model-label">文本模型：</span>
                <el-select
                  v-model="selectedTextModel"
                  placeholder="选择文本生成模型"
                  size="default"
                  style="width: 280px"
                >
                  <el-option
                    v-for="model in textModels"
                    :key="model.modelName"
                    :label="model.modelName"
                    :value="model.modelName"
                  />
                </el-select>
              </div>
              <div class="model-selector-item" style="margin-left: 24px;">
                <span class="model-label">镜头数量：</span>
                <el-tag type="info">由剧本自动决定</el-tag>
                <el-tooltip content="镜头数量已在「剧本设计」步骤中由 AI 根据时长自动确定" placement="top">
                  <el-icon style="margin-left: 4px; color: #909399;"><QuestionFilled /></el-icon>
                </el-tooltip>
              </div>
            </div>

            <el-divider style="margin: 12px 0" />

            <!-- 分镜列表 -->
            <div
              v-if="
                currentEpisode?.storyboards &&
                currentEpisode.storyboards.length > 0
              "
              class="shots-list"
            >
              <div class="shots-header">
                <h3>{{ $t("workflow.shotList") }}（共 {{ currentEpisode.storyboards.length }} 个镜头）</h3>
              </div>

              <el-table
                :data="currentEpisode.storyboards"
                border
                stripe
                style="margin-top: 16px"
              >
                <el-table-column
                  type="index"
                  :label="$t('storyboard.table.number')"
                  width="60"
                />
                <el-table-column
                  :label="$t('storyboard.table.title')"
                  width="120"
                  show-overflow-tooltip
                >
                  <template #default="{ row }">
                    {{ row.title || "-" }}
                  </template>
                </el-table-column>
                <el-table-column
                  :label="$t('storyboard.table.location')"
                  width="150"
                >
                  <template #default="{ row }">
                    <el-popover
                      placement="right"
                      :width="300"
                      trigger="hover"
                      :content="row.location || '-'"
                    >
                      <template #reference>
                        <span class="overflow-tooltip">{{ row.location || "-" }}</span>
                      </template>
                    </el-popover>
                  </template>
                </el-table-column>
                <el-table-column
                  :label="$t('storyboard.table.character')"
                  width="100"
                >
                  <template #default="{ row }">
                    <span v-if="row.characters && row.characters.length > 0">
                      {{ row.characters.map((c: any) => c.name || c).join(", ") }}
                    </span>
                    <span v-else>-</span>
                  </template>
                </el-table-column>
                <el-table-column
                  label="道具"
                  width="100"
                >
                  <template #default="{ row }">
                    <span v-if="row.props && row.props.length > 0">
                      {{ row.props.map((p: any) => p.name || p).join(", ") }}
                    </span>
                    <span v-else>-</span>
                  </template>
                </el-table-column>
                <el-table-column label="镜头描述">
                  <template #default="{ row }">
                    <el-popover
                      placement="right"
                      :width="400"
                      trigger="hover"
                    >
                      <template #default>
                        <div v-if="row.action"><strong>动作：</strong>{{ row.action }}</div>
                        <div v-if="row.dialogue" style="margin-top: 4px"><strong>对白：</strong>{{ row.dialogue }}</div>
                        <div v-if="row.result" style="margin-top: 4px"><strong>结果：</strong>{{ row.result }}</div>
                        <div v-if="row.atmosphere" style="margin-top: 4px"><strong>氛围：</strong>{{ row.atmosphere }}</div>
                      </template>
                      <template #reference>
                        <span class="overflow-tooltip">{{ row.action || "-" }}</span>
                      </template>
                    </el-popover>
                  </template>
                </el-table-column>
                <el-table-column
                  :label="$t('common.actions')"
                  width="120"
                  fixed="right"
                >
                  <template #default="{ row, $index }">
                    <el-button
                      type="primary"
                      size="small"
                      @click="editShot(row, $index)"
                      :icon="Edit"
                    >
                      {{ $t("common.edit") }}
                    </el-button>
                    <el-button
                      type="danger"
                      size="small"
                      @click="deleteShot(row.id)"
                      :icon="Delete"
                    />
                  </template>
                </el-table-column>
              </el-table>
            </div>

            <!-- 未拆分时显示 -->
            <div v-else class="empty-shots">
              <el-empty description="点击下方按钮，AI 将根据剧本自动拆分分镜">
                <el-button
                  type="primary"
                  size="large"
                  @click="generateShots"
                  :loading="generatingShots"
                  :disabled="!currentEpisode?.script_content"
                  :icon="MagicStick"
                >
                  {{ generatingShots ? 'AI 拆分中...' : 'AI 自动拆分分镜' }}
                </el-button>
              </el-empty>
            </div>
          </div>
        </el-card>
      </div>

      <div class="actions-container">
        <!-- Step 0: 资源定义 -->
        <div class="action-buttons" v-show="currentStep === 0">
          <el-button
            type="success"
            size="large"
            @click="nextStep"
            :disabled="charactersCount === 0 && (currentEpisode?.scenes?.length || 0) === 0 && (episodeProps?.length || 0) === 0"
          >
            下一步：剧本设计
            <el-icon><ArrowRight /></el-icon>
          </el-button>
          <div v-if="!hasExtractedData" style="margin-top: 8px">
            <el-alert type="warning" :closable="false" style="display: inline-block">
              <template #title>
                <span style="font-size: 12px">建议先添加角色、场景或道具后再进入下一步</span>
              </template>
            </el-alert>
          </div>
        </div>

        <!-- Step 1: 剧本设计 -->
        <div class="action-buttons" v-show="currentStep === 1">
          <el-button size="large" @click="prevStep">
            <el-icon><ArrowLeft /></el-icon>
            上一步
          </el-button>
          <el-button
            type="success"
            size="large"
            @click="nextStep"
            :disabled="!currentEpisode?.script_content"
          >
            下一步：分镜生成
            <el-icon><ArrowRight /></el-icon>
          </el-button>
          <div v-if="!currentEpisode?.script_content" style="margin-top: 8px">
            <el-alert type="warning" :closable="false" style="display: inline-block">
              <template #title>
                <span style="font-size: 12px">请先生成剧本后再进入分镜生成</span>
              </template>
            </el-alert>
          </div>
        </div>

        <!-- Step 2: 分镜生成 -->
        <div class="action-buttons" v-show="currentStep === 2">
          <el-button size="large" @click="prevStep">
            <el-icon><ArrowLeft /></el-icon>
            上一步
          </el-button>
          <el-button
            v-if="currentEpisode?.storyboards?.length > 0"
            size="large"
            @click="regenerateShots"
            :icon="MagicStick"
          >
            重新生成分镜
          </el-button>
          <el-button
            type="success"
            size="large"
            @click="goToProfessionalUI"
            :disabled="!currentEpisode?.storyboards?.length"
          >
            {{ $t("workflow.enterProfessional") }}
            <el-icon><ArrowRight /></el-icon>
          </el-button>
        </div>
      </div>
    </div>

    <div class="components-box">
      <!-- AI 标准化改写进度弹窗 -->
      <el-dialog
        v-model="rewriteProgressDialog.visible"
        title="AI 标准化改写"
        width="480px"
        :close-on-click-modal="false"
        :show-close="!rewriteProgressDialog.running"
        :close-on-press-escape="!rewriteProgressDialog.running"
      >
        <div style="padding: 16px 0;">
          <div v-if="rewriteProgressDialog.running" style="text-align: center;">
            <el-icon :size="40" color="#409EFF" style="margin-bottom: 16px;" class="is-loading">
              <Loading />
            </el-icon>
            <p style="font-size: 15px; font-weight: 500; margin-bottom: 16px; color: #303133;">
              AI 正在标准化改写剧本，请稍候...
            </p>
            <el-progress
              :percentage="rewriteProgressDialog.progress"
              :stroke-width="18"
              :text-inside="true"
              style="margin-bottom: 12px;"
            />
            <p style="font-size: 13px; color: #909399;">{{ rewriteProgressDialog.progressMsg }}</p>
            <p style="font-size: 12px; color: #c0c4cc; margin-top: 8px;">改写完成后将自动关闭此窗口</p>
          </div>
          <div v-else-if="rewriteProgressDialog.success" style="text-align: center;">
            <el-icon :size="48" color="#67c23a" style="margin-bottom: 12px;">
              <CircleCheckFilled />
            </el-icon>
            <p style="font-size: 16px; font-weight: 500; color: #303133;">标准化改写完成</p>
            <p style="font-size: 13px; color: #909399; margin-top: 8px;">剧本已更新为标准格式</p>
          </div>
          <div v-else-if="rewriteProgressDialog.error" style="text-align: center;">
            <el-icon :size="48" color="#f56c6c" style="margin-bottom: 12px;">
              <CircleCloseFilled />
            </el-icon>
            <p style="font-size: 16px; font-weight: 500; color: #303133;">改写失败</p>
            <p style="font-size: 13px; color: #f56c6c; margin-top: 8px;">{{ rewriteProgressDialog.errorMsg }}</p>
          </div>
        </div>
        <template #footer>
          <el-button v-if="!rewriteProgressDialog.running" @click="rewriteProgressDialog.visible = false">
            关闭
          </el-button>
        </template>
      </el-dialog>

      <!-- 拆分分镜进度弹窗 -->
      <el-dialog
        v-model="splitProgressDialog.visible"
        title="拆分分镜"
        width="480px"
        :close-on-click-modal="false"
        :show-close="!splitProgressDialog.running"
        :close-on-press-escape="!splitProgressDialog.running"
      >
        <div style="padding: 16px 0;">
          <div v-if="splitProgressDialog.running" style="text-align: center;">
            <el-icon :size="40" color="#409EFF" style="margin-bottom: 16px;" class="is-loading">
              <Loading />
            </el-icon>
            <p style="font-size: 15px; font-weight: 500; margin-bottom: 16px; color: #303133;">
              正在拆分分镜，请稍候...
            </p>
            <el-progress
              :percentage="taskProgress"
              :stroke-width="18"
              :text-inside="true"
              style="margin-bottom: 12px;"
            />
            <p style="font-size: 13px; color: #909399;">{{ taskMessage }}</p>
          </div>
          <div v-else-if="splitProgressDialog.success" style="text-align: center;">
            <el-icon :size="48" color="#67c23a" style="margin-bottom: 12px;">
              <CircleCheckFilled />
            </el-icon>
            <p style="font-size: 16px; font-weight: 500; color: #303133;">分镜拆分完成</p>
            <p style="font-size: 13px; color: #909399; margin-top: 8px;">即将跳转到编辑器...</p>
          </div>
          <div v-else-if="splitProgressDialog.error" style="text-align: center;">
            <el-icon :size="48" color="#f56c6c" style="margin-bottom: 12px;">
              <CircleCloseFilled />
            </el-icon>
            <p style="font-size: 16px; font-weight: 500; color: #303133;">拆分失败</p>
            <p style="font-size: 13px; color: #f56c6c; margin-top: 8px;">{{ splitProgressDialog.errorMsg }}</p>
          </div>
        </div>
        <template #footer>
          <el-button v-if="!splitProgressDialog.running" @click="splitProgressDialog.visible = false">
            关闭
          </el-button>
        </template>
      </el-dialog>

      <!-- 镜头编辑对话框 -->
      <el-dialog
        v-model="shotEditDialogVisible"
        :title="$t('workflow.editShot')"
        width="800px"
        :close-on-click-modal="false"
      >
        <el-form v-if="editingShot" label-width="100px" size="default">
          <el-form-item :label="$t('workflow.shotTitle')">
            <el-input
              v-model="editingShot.title"
              :placeholder="$t('workflow.shotTitlePlaceholder')"
            />
          </el-form-item>

          <el-row :gutter="16">
            <el-col :span="8">
              <el-form-item :label="$t('workflow.shotType')">
                <el-select
                  v-model="editingShot.shot_type"
                  :placeholder="$t('workflow.selectShotType')"
                >
                  <el-option label="远景" value="远景" />
                  <el-option label="全景" value="全景" />
                  <el-option label="中景" value="中景" />
                  <el-option label="近景" value="近景" />
                  <el-option label="特写" value="特写" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item :label="$t('workflow.cameraAngle')">
                <el-select
                  v-model="editingShot.angle"
                  :placeholder="$t('workflow.selectAngle')"
                >
                  <el-option label="平视" value="平视" />
                  <el-option label="俯视" value="俯视" />
                  <el-option label="仰视" value="仰视" />
                  <el-option label="高机位" value="高机位" />
                  <el-option label="低机位" value="低机位" />
                  <el-option label="过肩视角" value="过肩视角" />
                  <el-option label="主观视角" value="主观视角" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item :label="$t('workflow.cameraMovement')">
                <el-select
                  v-model="editingShot.movement"
                  :placeholder="$t('workflow.selectMovement')"
                >
                  <el-option label="固定镜头" value="固定镜头" />
                  <el-option label="推" value="推" />
                  <el-option label="拉" value="拉" />
                  <el-option label="摇" value="摇" />
                  <el-option label="移" value="移" />
                  <el-option label="跟" value="跟" />
                  <el-option label="升" value="升" />
                  <el-option label="降" value="降" />
                  <el-option label="甩" value="甩" />
                  <el-option label="环绕" value="环绕" />
                  <el-option label="旋转" value="旋转" />
                  <el-option label="变焦" value="变焦" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item :label="$t('workflow.location')">
                <el-input
                  v-model="editingShot.location"
                  :placeholder="$t('workflow.locationPlaceholder')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('workflow.time')">
                <el-input
                  v-model="editingShot.time"
                  :placeholder="$t('workflow.timeSetting')"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item :label="$t('workflow.actionDescription')">
            <el-input
              v-model="editingShot.action"
              type="textarea"
              :rows="3"
              :placeholder="$t('workflow.detailedAction')"
            />
          </el-form-item>

          <el-form-item :label="$t('workflow.dialogue')">
            <el-input
              v-model="editingShot.dialogue"
              type="textarea"
              :rows="2"
              :placeholder="$t('workflow.characterDialogue')"
            />
          </el-form-item>

          <el-form-item :label="$t('workflow.result')">
            <el-input
              v-model="editingShot.result"
              type="textarea"
              :rows="2"
              :placeholder="$t('workflow.actionResult')"
            />
          </el-form-item>

          <el-form-item :label="$t('workflow.atmosphere')">
            <el-input
              v-model="editingShot.atmosphere"
              type="textarea"
              :rows="2"
              :placeholder="$t('workflow.atmosphereDescription')"
            />
          </el-form-item>

          <el-form-item :label="$t('workflow.imagePrompt')">
            <el-input
              v-model="editingShot.image_prompt"
              type="textarea"
              :rows="3"
              :placeholder="$t('workflow.imagePromptPlaceholder')"
            />
          </el-form-item>

          <el-form-item :label="$t('workflow.videoPrompt')">
            <el-input
              v-model="editingShot.video_prompt"
              type="textarea"
              :rows="3"
              :placeholder="$t('workflow.videoPromptPlaceholder')"
            />
          </el-form-item>

          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item :label="$t('workflow.bgmHint')">
                <el-input
                  v-model="editingShot.bgm_prompt"
                  :placeholder="$t('workflow.bgmAtmosphere')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('workflow.soundEffect')">
                <el-input
                  v-model="editingShot.sound_effect"
                  :placeholder="$t('workflow.soundEffectDescription')"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item :label="$t('workflow.durationSeconds')">
            <el-input-number
              v-model="editingShot.duration"
              :min="1"
              :max="60"
            />
          </el-form-item>
        </el-form>

        <template #footer>
          <el-button @click="shotEditDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button
            type="primary"
            @click="saveShotEdit"
            :loading="savingShot"
            >{{ $t("common.save") }}</el-button
          >
        </template>
      </el-dialog>

      <!-- 提示词编辑对话框 -->
      <el-dialog
        v-model="promptDialogVisible"
        :title="$t('workflow.editPrompt')"
        width="900px"
      >
        <el-form label-position="top">
          <el-form-item :label="$t('common.name')">
            <el-input v-model="currentEditItem.name" />
          </el-form-item>
          <el-form-item label="生成尺寸">
            <el-tag v-if="currentEditType === 'scene'" type="info" effect="plain">1:1 正方形 · 2048×2048（四视图网格）</el-tag>
            <el-tag v-else type="info" effect="plain">4:3 横向 · 2304×1728（三视图）</el-tag>
          </el-form-item>
          <el-form-item :label="$t('workflow.imagePrompt')">
            <div style="display: flex; gap: 8px; margin-bottom: 8px;">
              <el-button
                type="primary"
                size="small"
                :loading="polishingPrompt"
                @click="polishPromptWithAI"
              >
                <el-icon><MagicStick /></el-icon>
                AI润色
              </el-button>
            </div>
            <el-input
              v-model="editPrompt"
              type="textarea"
              :rows="6"
              :placeholder="$t('workflow.imagePromptPlaceholder')"
            />
          </el-form-item>
          
          <el-form-item label="参考图片">
            <el-upload
              class="avatar-uploader"
              :action="uploadAction"
              :headers="uploadHeaders"
              :show-file-list="false"
              :on-success="handleReferenceImageUploadSuccess"
              :before-upload="beforeAvatarUpload"
            >
              <el-icon class="avatar-uploader-icon"><Plus /></el-icon>
            </el-upload>
          </el-form-item>
          
          <el-form-item v-if="editReferenceImages.length > 0">
            <div style="background: #f5f7fa; border-radius: 8px; padding: 12px; border: 1px solid #ebeef5; max-width: 100%; overflow: hidden;">
              <div
                v-for="(ref, idx) in editReferenceImages"
                :key="idx"
                style="display: flex; align-items: center; gap: 12px; padding: 8px 0; border-bottom: 1px solid #ebeef5; min-width: 0;"
              >
                <el-image
                  :src="getImageUrl({ image_url: ref.path, local_path: ref.path })"
                  fit="cover"
                  style="width: 80px; height: 80px; border-radius: 4px; border: 1px solid #dcdfe6; flex-shrink: 0;"
                />
                <div style="flex: 1; min-width: 0; overflow: hidden;">
                  <div style="font-size: 14px; color: #303133; margin-bottom: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">{{ ref.name }}</div>
                  <div style="color: #909399; font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">{{ ref.path }}</div>
                </div>
                <el-button
                  type="danger"
                  size="small"
                  link
                  @click="removeReferenceImage(idx)"
                  style="flex-shrink: 0;"
                >
                  删除
                </el-button>
              </div>
            </div>
          </el-form-item>
          
          <el-form-item>
            <el-text type="info" size="small">
              {{ currentEditType === 'character' ? '提示词将用于AI生成人物图片，可以手动修改以获得更好的效果。' : '提示词将用于AI生成场景图片，可以手动修改以获得更好的效果。' }}
            </el-text>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="promptDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button type="primary" @click="savePrompt">{{
            $t("common.save")
          }}</el-button>
          <el-button type="success" @click="generateImageWithSize" :loading="generatingEditItem">
            保存并生成
          </el-button>
        </template>
      </el-dialog>

      <!-- 角色库选择对话框 -->
      <el-dialog
        v-model="libraryDialogVisible"
        :title="getLibraryDialogTitle()"
        width="800px"
      >
        <div class="library-grid">
          <div
            v-for="item in libraryItems"
            :key="item.id"
            class="library-item"
            @click="selectLibraryItem(item)"
          >
            <el-image :src="getImageUrl(item)" fit="contain" lazy loading="lazy" />
            <div class="library-item-name">{{ item._displayName || item.name || item.location || '未命名' }}</div>
          </div>
        </div>
        <div v-if="libraryItems.length === 0" class="empty-library">
          <el-empty :description="$t('workflow.emptyLibrary')" />
        </div>
      </el-dialog>

      <!-- 图片上传对话框 -->
      <el-dialog
        v-model="uploadDialogVisible"
        :title="$t('tooltip.uploadImage')"
        width="500px"
      >
        <el-upload
          class="upload-area"
          drag
          :action="uploadAction"
          :headers="uploadHeaders"
          :on-success="handleUploadSuccess"
          :on-error="handleUploadError"
          :show-file-list="false"
          accept="image/jpeg,image/png,image/jpg"
        >
          <el-icon class="el-icon--upload"><Upload /></el-icon>
          <div class="el-upload__text">
            {{ $t("workflow.dragFilesHere")
            }}<em>{{ $t("workflow.clickToUpload") }}</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">
              {{ $t("workflow.uploadFormatTip") }}
            </div>
          </template>
        </el-upload>
      </el-dialog>

      <!-- 添加场景对话框 -->
      <el-dialog
        v-model="addSceneDialogVisible"
        :title="$t('workflow.addScene')"
        width="600px"
      >
        <el-form :model="newScene" label-width="100px">
          <el-form-item :label="$t('workflow.sceneImage')">
            <el-upload
              class="avatar-uploader"
              :action="`/api/v1/upload/image`"
              :show-file-list="false"
              :on-success="handleSceneImageSuccess"
              :before-upload="beforeAvatarUpload"
            >
              <img
                v-if="hasImage(newScene)"
                :src="getImageUrl(newScene)"
                class="avatar"
                style="width: 160px; height: 90px; object-fit: cover"
              />
              <el-icon
                v-else
                class="avatar-uploader-icon"
                style="
                  border: 1px dashed #d9d9d9;
                  border-radius: 6px;
                  cursor: pointer;
                  position: relative;
                  overflow: hidden;
                  width: 160px;
                  height: 90px;
                  font-size: 28px;
                  color: #8c939d;
                  text-align: center;
                  line-height: 90px;
                "
                ><Plus
              /></el-icon>
            </el-upload>
          </el-form-item>
          <el-form-item :label="$t('workflow.sceneName')">
            <el-input
              v-model="newScene.location"
              :placeholder="$t('workflow.sceneNamePlaceholder')"
            />
          </el-form-item>
          <el-form-item :label="$t('workflow.time')">
            <el-input
              v-model="newScene.time"
              :placeholder="$t('workflow.timePlaceholder')"
            />
          </el-form-item>
          <el-form-item :label="$t('workflow.sceneDescription')">
            <el-input
              v-model="newScene.prompt"
              type="textarea"
              :rows="4"
              :placeholder="$t('workflow.sceneDescriptionPlaceholder')"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="addSceneDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button type="primary" @click="saveScene">{{
            $t("common.confirm")
          }}</el-button>
        </template>
      </el-dialog>

      <!-- 添加角色对话框 -->
      <el-dialog
        v-model="addCharacterDialogVisible"
        title="添加角色"
        width="600px"
      >
        <el-form :model="newCharacter" label-width="100px">
          <el-form-item label="角色名称">
            <el-input v-model="newCharacter.name" placeholder="请输入角色名称" />
          </el-form-item>
          <el-form-item label="角色类型">
            <el-select v-model="newCharacter.role" placeholder="请选择角色类型" style="width: 100%">
              <el-option label="主角" value="主角" />
              <el-option label="配角" value="配角" />
              <el-option label="龙套" value="龙套" />
            </el-select>
          </el-form-item>
          <el-form-item label="外观描述">
            <el-input
              v-model="newCharacter.appearance"
              type="textarea"
              :rows="4"
              placeholder="请输入角色的外观描述"
            />
          </el-form-item>
          <el-form-item label="性格描述">
            <el-input
              v-model="newCharacter.personality"
              type="textarea"
              :rows="3"
              placeholder="请输入角色的性格描述"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="addCharacterDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button type="primary" @click="saveCharacter">
            {{ $t("common.confirm") }}
          </el-button>
        </template>
      </el-dialog>

      <!-- 添加道具对话框 -->
      <el-dialog
        v-model="addPropDialogVisible"
        title="添加道具"
        width="600px"
      >
        <el-form :model="newProp" label-width="100px">
          <el-form-item label="道具名称">
            <el-input v-model="newProp.name" placeholder="请输入道具名称" />
          </el-form-item>
          <el-form-item label="道具类型">
            <el-input v-model="newProp.type" placeholder="请输入道具类型（如：武器、工具、饰品等）" />
          </el-form-item>
          <el-form-item label="道具描述">
            <el-input
              v-model="newProp.description"
              type="textarea"
              :rows="4"
              placeholder="请输入道具的描述"
            />
          </el-form-item>
          <el-form-item label="提示词">
            <el-input
              v-model="newProp.prompt"
              type="textarea"
              :rows="3"
              placeholder="请输入用于生成道具图片的提示词"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="addPropDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button type="primary" @click="saveProp">
            {{ $t("common.confirm") }}
          </el-button>
        </template>
      </el-dialog>

      <!-- 从剧本提取场景对话框 -->
      <el-dialog
        v-model="extractScenesDialogVisible"
        :title="$t('workflow.extractSceneDialogTitle')"
        width="500px"
      >
        <el-alert type="info" :closable="false" style="margin-bottom: 16px">
          {{ $t("workflow.extractSceneDialogTip") }}
        </el-alert>
        <template #footer>
          <el-button @click="extractScenesDialogVisible = false">
            {{ $t("common.cancel") }}
          </el-button>
          <el-button
            type="primary"
            @click="handleExtractScenes"
            :loading="extractingScenes"
          >
            {{ $t("workflow.startExtract") }}
          </el-button>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  User,
  Location,
  Picture,
  MagicStick,
  ArrowRight,
  ArrowLeft,
  Place,
  Film,
  Edit,
  More,
  Upload,
  Delete,
  FolderAdd,
  FolderOpened,
  Setting,
  Loading,
  WarningFilled,
  CircleCheckFilled,
  CircleCloseFilled,
  Document,
  Plus,
  Check,
  Box,
  Close,
  EditPen,
  RefreshRight,
  ChatLineSquare,
  QuestionFilled,
  View,
} from "@element-plus/icons-vue";
import { dramaAPI } from "@/api/drama";
import { generationAPI } from "@/api/generation";
import { characterLibraryAPI } from "@/api/character-library";
import { sceneLibraryAPI } from "@/api/scene-library";
import { propLibraryAPI } from "@/api/prop-library";
import { characterAPI, sceneAPI } from "@/api/resource";
import { aiAPI } from "@/api/ai";
import type { AIServiceConfig } from "@/types/ai";
import { imageAPI } from "@/api/image";
import { propAPI } from "@/api/prop";
import type { Prop } from "@/types/prop";
import type { Drama } from "@/types/drama";
import { AppHeader } from "@/components/common";
import ImagePreview from "@/components/common/ImagePreview.vue";
import CollapsibleText from "@/components/common/CollapsibleText.vue";
import { getImageUrl, hasImage } from "@/utils/image";

const getFirstRefImage = (item: any): string => {
  if (!item?.reference_images || !Array.isArray(item.reference_images) || item.reference_images.length === 0) return '';
  const ref = item.reference_images[0];
  const path = typeof ref === 'string' ? ref : ref?.path;
  if (!path) return '';
  if (path.startsWith('http') || path.startsWith('/')) return path;
  return `/static/${path}`;
};

const route = useRoute();
const router = useRouter();
const { t: $t } = useI18n();
const dramaId = route.params.id as string;
const episodeNumber = parseInt(route.params.episodeNumber as string);

const drama = ref<Drama>();

// 生成 localStorage key
const getStepStorageKey = () =>
  `episode_workflow_step_${dramaId}_${episodeNumber}`;

// 每次进入页面从第一步开始，用户可通过步骤条手动切换
const currentStep = ref(0);
const scriptContent = ref("");
const generatingScript = ref(false);
const generatingShots = ref(false);

// ======= 剧本设计（Step 1）=======
const scriptDraftInput = ref("");
const scriptEditMode = ref(false);
const shotCount = ref(12);
const episodeDuration = ref(90);

const scriptShotCount = computed(() => {
  const content = currentEpisode.value?.script_content || '';
  const matches = content.match(/S\d+\s*·/g);
  return matches ? matches.length : 0;
});

const scriptTotalDuration = computed(() => {
  const content = currentEpisode.value?.script_content || '';
  const matches = content.match(/（(\d+)s）/g);
  if (!matches) return 0;
  return matches.reduce((sum: number, m: string) => {
    const d = m.match(/(\d+)/);
    return sum + (d ? parseInt(d[1]) : 0);
  }, 0);
});
const durationOptions = [30, 60, 90, 120, 150, 180, 210, 240, 270, 300];
const scriptFeedback = ref("");
const showFeedbackDialog = ref(false);
const extractingCharactersAndBackgrounds = ref(false);
const batchGeneratingCharacters = ref(false);
const batchGeneratingScenes = ref(false);
const generatingCharacterImages = ref<Record<number, boolean>>({});
const generatingSceneImages = ref<Record<string, boolean>>({});
const generatingPropImages = ref<Record<number, boolean>>({});
const generatingEditItem = ref(false);
const polishingPrompt = ref(false);

// 选择状态
const selectedCharacterIds = ref<number[]>([]);
const selectedSceneIds = ref<number[]>([]);
const selectAllCharacters = ref(false);
const selectAllScenes = ref(false);

// 对话框状态
const promptDialogVisible = ref(false);
const libraryDialogVisible = ref(false);
const uploadDialogVisible = ref(false);
const addSceneDialogVisible = ref(false);
const addCharacterDialogVisible = ref(false);
const addPropDialogVisible = ref(false);
const extractScenesDialogVisible = ref(false);
const currentEditItem = ref<any>({ name: "" });
const currentEditType = ref<"character" | "scene" | "prop">("character");
const editPrompt = ref("");
const originalEditPrompt = ref(""); // 保存原始提示词（不含比例要求）
const editOrientation = ref<"landscape" | "portrait">("portrait"); // 默认竖屏
const editReferenceImages = ref<{ name: string; path: string }[]>([]); // 参考图片
const libraryItems = ref<any[]>([]);
const currentUploadTarget = ref<any>(null);

const newCharacter = ref<any>({
  name: "",
  role: "",
  appearance: "",
  personality: "",
});

// 根据横竖屏生成带比例要求的提示词
const getEpisodePromptWithOrientation = (basePrompt: string, orientation: string) => {
  const orientationText = orientation === "landscape" ? "（16:9横屏）" : "（9:16竖屏）";
  const cleanPrompt = basePrompt.replace(/（\d+:\d+[横竖]屏）/g, "").replace(/，\s*$/g, "").replace(/\s*$/g, "");
  return cleanPrompt + orientationText;
};

// 监听 editOrientation 变化，更新提示词中的比例要求
watch(
  () => editOrientation.value,
  (newOrientation) => {
    // 更新提示词中的比例要求，但保留用户编辑的内容
    const newOrientationText = newOrientation === "landscape" ? "（16:9横屏）" : "（9:16竖屏）";
    // 只替换横竖屏标记，不重新构建整个提示词
    editPrompt.value = editPrompt.value.replace(/（\d+:\d+[横竖]屏）/g, "") + newOrientationText;
  }
);

// 移除 editReferenceImages 的监听，让用户自己编辑提示词
// 避免覆盖用户的编辑内容

// 添加场景相关
const newScene = ref<any>({
  location: "",
  time: "",
  prompt: "",
  image_url: "",
  local_path: "",
});

// 添加道具相关
const newProp = ref<any>({
  name: "",
  type: "",
  description: "",
  prompt: "",
  image_url: "",
});

const extractingScenes = ref(false);
const uploadAction = computed(() => "/api/v1/upload/image");
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${localStorage.getItem("token")}`,
}));

// AI模型配置
interface ModelOption {
  modelName: string;
  configName: string;
  configId: number;
  priority: number;
  price?: number;
}

const textModels = ref<ModelOption[]>([]);
const imageModels = ref<ModelOption[]>([]);
const selectedTextModel = ref<string>("");
const selectedImageModel = ref<string>("");

const selectedTextToImageModel = ref<string>("");
const selectedImageToImageModel = ref<string>("");

const isEditingScript = ref(false);

const hasScript = computed(() => {
  const currentEp = currentEpisode.value;
  return (
    currentEp && currentEp.script_content && currentEp.script_content.length > 0
  );
});

const currentEpisode = computed(() => {
  if (!drama.value?.episodes) return null;
  return drama.value.episodes.find((ep) => ep.episode_number === episodeNumber);
});

const episodeProps = computed(() => {
  return currentEpisode.value?.props || [];
});

const hasCharacters = computed(() => {
  return (
    currentEpisode.value?.characters &&
    currentEpisode.value.characters.length > 0
  );
});

const charactersCount = computed(() => {
  return currentEpisode.value?.characters?.length || 0;
});

const hasExtractedData = computed(() => {
  const hasScenes =
    currentEpisode.value?.scenes && currentEpisode.value.scenes.length > 0;
  // 只要有角色或场景，就认为已经提取过数据
  return hasCharacters.value || hasScenes;
});

// ======= 剧本内容质量检测 =======
interface ScriptQualityItem {
  label: string;
  passed: boolean;
  detail: string;
}

const scriptQuality = computed(() => {
  const script = currentEpisode.value?.script_content || "";
  if (!script) return { items: [] as ScriptQualityItem[], score: 0, sufficient: false };

  const items: ScriptQualityItem[] = [];

  // 1. 长度检测（最低 100 字）
  const len = script.length;
  items.push({
    label: "内容长度",
    passed: len >= 100,
    detail: len >= 100 ? `${len} 字（充足）` : `${len} 字（建议至少 100 字）`,
  });

  // === 基于 AI 标准化改写的剧本格式解析 ===
  // 格式：剧本标题/类型/时长 → 人物（性格+外观） → 场景 → [开场] → 对话+舞台指示 → 剧终
  const lines = script.split("\n").map((l: string) => l.trim()).filter(Boolean);

  // 2. 剧本头部检测 —— 剧本标题：《xxx》
  const hasTitle = lines.some((l: string) => /^剧本标题[：:]/.test(l));
  const hasType = lines.some((l: string) => /^类型[：:]/.test(l));
  items.push({
    label: "剧本信息",
    passed: hasTitle,
    detail: hasTitle
      ? `包含剧本标题${hasType ? "和类型" : ""}信息`
      : "未检测到「剧本标题：」标记，建议先进行AI标准化改写",
  });

  // 3. 人物设定检测 —— "人物：" 开头区块 + "性格："/"外观：" 字段
  const hasCharBlock = lines.some((l: string) => /^人物[：:]/.test(l));
  const personalityLines = lines.filter((l: string) => /^性格[：:]/.test(l));
  const appearanceLines = lines.filter((l: string) => /^外观[：:]/.test(l));
  // 提取角色名：角色名（身份，年龄，...）格式，在"人物："之后
  const characterDescs: string[] = [];
  let inCharSection = false;
  for (const line of lines) {
    if (/^人物[：:]/.test(line)) { inCharSection = true; continue; }
    if (/^场景[：:]|^\[开场/.test(line)) { inCharSection = false; continue; }
    if (inCharSection) {
      // 匹配: 角色名（身份信息）
      const m = line.match(/^([\u4e00-\u9fa5A-Za-z]{1,8})[（(]/);
      if (m) characterDescs.push(m[1]);
    }
  }
  items.push({
    label: "人物设定",
    passed: characterDescs.length >= 1 && personalityLines.length >= 1,
    detail: characterDescs.length >= 1
      ? `检测到 ${characterDescs.length} 个角色（${characterDescs.slice(0, 5).join("、")}），性格描述 ${personalityLines.length} 条，外观描述 ${appearanceLines.length} 条`
      : "未检测到人物设定区块，建议AI标准化改写",
  });

  // 4. 场景设定检测 —— "场景：" 开头
  const sceneLines = lines.filter((l: string) => /^场景[：:]/.test(l));
  items.push({
    label: "场景描述",
    passed: sceneLines.length >= 1,
    detail: sceneLines.length >= 1
      ? `检测到 ${sceneLines.length} 个场景描述`
      : "未检测到「场景：」标记，建议AI标准化改写",
  });

  // 5. 对话检测 —— 角色名（动作）：台词  或  角色名：台词
  // 新格式：冒号后不需要引号
  const isDialogueLine = (l: string) =>
    /^[\u4e00-\u9fa5A-Za-z]{1,8}(?:[（(][^）)]*[）)])?[：:]\s*\S/.test(l) &&
    !/^剧本标题|^类型|^时长|^人物|^场景|^性格|^外观/.test(l);
  const dialogueLinesList = lines.filter(isDialogueLine);
  // 从对话行提取角色名
  const dialogueCharacters = new Set<string>();
  for (const line of dialogueLinesList) {
    const m = line.match(/^([\u4e00-\u9fa5A-Za-z]{1,8})(?:[（(]|[：:])/);
    if (m) dialogueCharacters.add(m[1]);
  }
  items.push({
    label: "角色对话",
    passed: dialogueLinesList.length >= 4,
    detail: dialogueLinesList.length >= 4
      ? `检测到 ${dialogueLinesList.length} 行对话，${dialogueCharacters.size} 个对话角色（${[...dialogueCharacters].slice(0, 5).join("、")}）`
      : `仅 ${dialogueLinesList.length} 行对话，建议AI标准化改写后至少包含4轮对话`,
  });

  // 6. 舞台指示/动作描写 —— （括号内容）独立行
  const stageDirections = lines.filter(
    (l: string) => /^[（(].+[）)]$/.test(l)
  );
  // 还统计 [开场：...] 行
  const hasOpening = lines.some((l: string) => /^\[开场/.test(l));
  const actionTotal = stageDirections.length + (hasOpening ? 1 : 0);
  items.push({
    label: "舞台指示",
    passed: actionTotal >= 2,
    detail: actionTotal >= 2
      ? `包含 ${stageDirections.length} 条舞台指示${hasOpening ? "，含开场描写" : ""}`
      : "缺少舞台指示（独立括号行），建议AI标准化改写",
  });

  const passedCount = items.filter((i) => i.passed).length;
  const score = Math.round((passedCount / items.length) * 100);

  return {
    items,
    score,
    sufficient: score >= 60, // 至少 4/6 通过才算足够
  };
});

// 检测剧本中是否包含对话（供其他功能使用）
const scriptHasDialogue = computed(() => {
  const qi = scriptQuality.value.items.find((i) => i.label === "角色对话");
  return qi ? qi.passed : false;
});

const generatingDialogue = ref(false);

const allImagesGenerated = computed(() => {
  // 如果没有提取任何数据，允许跳过（可能是空章节或用户想直接进入拆解分镜）
  if (!hasExtractedData.value) return true;

  const characters = currentEpisode.value?.characters || [];
  const scenes = currentEpisode.value?.scenes || [];

  // 如果角色和场景都为空，允许跳过
  if (characters.length === 0 && scenes.length === 0) return true;

  // 检查所有有数据的项是否都已生成图片
  const allCharsHaveImages =
    characters.length === 0 || characters.every((char) => char.image_url);
  const allScenesHaveImages =
    scenes.length === 0 || scenes.every((scene) => scene.image_url);

  return allCharsHaveImages && allScenesHaveImages;
});

const goBack = () => {
  // 使用 replace 避免在历史记录中留下当前页面
  router.replace(`/dramas/${dramaId}`);
};

// 加载AI模型配置
const loadAIConfigs = async () => {
  try {
    const [textList, imageList] = await Promise.all([
      aiAPI.list("text"),
      aiAPI.list("image"),
    ]);

    // 只使用激活的配置
    const activeTextList = textList.filter((c) => c.is_active);
    const activeImageList = imageList.filter((c) => c.is_active);

    // 展开模型列表并去重（保留优先级最高的）
    const allTextModels = activeTextList
      .flatMap((config) => {
        const models = Array.isArray(config.model)
          ? config.model
          : [config.model];
        return models.map((modelName) => ({
          modelName,
          configName: config.name,
          configId: config.id,
          priority: config.priority || 0,
        }));
      })
      .sort((a, b) => b.priority - a.priority);

    // 按模型名称去重，保留优先级最高的（已排序，第一个就是优先级最高的）
    const textModelMap = new Map<string, ModelOption>();
    allTextModels.forEach((model) => {
      if (!textModelMap.has(model.modelName)) {
        textModelMap.set(model.modelName, model);
      }
    });
    textModels.value = Array.from(textModelMap.values()).filter((m) =>
      m.modelName === "doubao-1-5-pro-32k-250115"
    );

    const allImageModels = activeImageList
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
          configId: config.id,
          priority: config.priority || 0,
          price: pricePerImage
        }));
      })
      .sort((a, b) => b.priority - a.priority);

    // 按模型名称去重，保留优先级最高的
    const imageModelMap = new Map<string, ModelOption>();
    allImageModels.forEach((model) => {
      if (!imageModelMap.has(model.modelName)) {
        imageModelMap.set(model.modelName, {
          modelName: model.modelName,
          configName: model.configName,
          configId: model.configId,
          priority: model.priority,
          price: model.price
        });
      }
    });
    // 显示所有图片模型
    imageModels.value = Array.from(imageModelMap.values());

    // 设置默认选择（优先选择 pro/plus/max/deepseek 等强模型）
    if (textModels.value.length > 0 && !selectedTextModel.value) {
      const proModel = textModels.value.find((m) => {
        const name = m.modelName.toLowerCase();
        return (
          (name.includes("pro") ||
            name.includes("plus") ||
            name.includes("max") ||
            name.includes("deepseek")) &&
          !name.includes("lite") &&
          !name.includes("mini") &&
          !name.includes("tiny")
        );
      });
      selectedTextModel.value = proModel
        ? proModel.modelName
        : textModels.value[0].modelName;
    }
    if (imageModels.value.length > 0 && !selectedImageModel.value) {
      selectedTextToImageModel.value = imageModels.value[0].modelName;
      selectedImageToImageModel.value = imageModels.value[0].modelName;
      selectedImageModel.value = selectedTextToImageModel.value;
    }

    // 验证已选择的模型是否还在可用列表中，如果不在则重置为默认值
    const availableTextModelNames = textModels.value.map((m) => m.modelName);
    const availableImageModelNames = imageModels.value.map((m) => m.modelName);

    if (
      selectedTextModel.value &&
      !availableTextModelNames.includes(selectedTextModel.value)
    ) {
      selectedTextModel.value =
        textModels.value.length > 0 ? textModels.value[0].modelName : "";
      // 更新 localStorage
      if (selectedTextModel.value) {
        localStorage.setItem(
          `ai_text_model_${dramaId}`,
          selectedTextModel.value,
        );
      }
    }

    if (
      selectedImageModel.value &&
      !availableImageModelNames.includes(selectedImageModel.value)
    ) {
      selectedImageModel.value =
        imageModels.value.length > 0
          ? imageModels.value[0].modelName
          : "";
      // 更新 localStorage
      if (selectedImageModel.value) {
        localStorage.setItem(
          `ai_image_model_${dramaId}`,
          selectedImageModel.value,
        );
      }
    }
  } catch (error: any) {
    console.error("加载AI配置失败:", error);
  }
};

const nextStep = () => {
  if (currentStep.value < 2) {
    currentStep.value++;
  }
};

const goToNextStep = () => {
  nextStep();
};

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--;
  }
};

const loadDramaData = async () => {
  try {
    const data = await dramaAPI.get(dramaId);
    drama.value = data;

    if (!hasScript.value) {
      scriptContent.value = "";
      currentStep.value = 0;
    }

    await checkAndStartPolling();
  } catch (error: any) {
    ElMessage.error(error.message || "加载项目数据失败");
  }
};

// 检查并启动轮询
const checkAndStartPolling = async () => {
  if (!currentEpisode.value) return;

  // 检查角色的生成状态
  for (const char of currentEpisode.value.characters || []) {
    if (
      char.image_generation_status === "pending" ||
      char.image_generation_status === "processing"
    ) {
      // 查找对应的image_generation记录
      try {
        const imageGenList = await imageAPI.listImages({
          drama_id: dramaId,
          status: char.image_generation_status as any,
        });

        // 找到这个角色的image_generation记录
        const charImageGen = imageGenList.items.find(
          (img) =>
            img.character_id === char.id &&
            (img.status === "pending" || img.status === "processing"),
        );

        if (charImageGen) {
          // 启动轮询
          generatingCharacterImages.value[char.id] = true;
          pollImageStatus(charImageGen.id, async () => {
            await loadDramaData();
            ElMessage.success(`${char.name}的图片生成完成！`);
          }).finally(() => {
            generatingCharacterImages.value[char.id] = false;
          });
        }
      } catch (error) {
        console.error("[轮询] 查询角色图片生成记录失败:", error);
      }
    }
  }

  // 检查场景的生成状态
  for (const scene of currentEpisode.value.scenes || []) {
    if (
      scene.image_generation_status === "pending" ||
      scene.image_generation_status === "processing"
    ) {
      // 查找对应的image_generation记录
      try {
        const imageGenList = await imageAPI.listImages({
          drama_id: dramaId,
          status: scene.image_generation_status as any,
        });

        // 找到这个场景的image_generation记录
        const sceneImageGen = imageGenList.items.find(
          (img) =>
            img.scene_id === scene.id &&
            (img.status === "pending" || img.status === "processing"),
        );

        if (sceneImageGen) {
          // 启动轮询
          generatingSceneImages.value[scene.id] = true;
          pollImageStatus(sceneImageGen.id, async () => {
            await loadDramaData();
            ElMessage.success(`${scene.location}的图片生成完成！`);
          }).finally(() => {
            generatingSceneImages.value[scene.id] = false;
          });
        }
      } catch (error) {
        console.error("[轮询] 查询场景图片生成记录失败:", error);
      }
    }
  }
};

const saveChapterScript = async () => {
  try {
    const existingEpisodes = drama.value?.episodes || [];

    // 查找当前章节
    const episodeIndex = existingEpisodes.findIndex(
      (ep) => ep.episode_number === episodeNumber,
    );

    let updatedEpisodes;
    if (episodeIndex >= 0) {
      // 更新已有章节
      updatedEpisodes = [...existingEpisodes];
      updatedEpisodes[episodeIndex] = {
        ...updatedEpisodes[episodeIndex],
        script_content: scriptContent.value,
      };
    } else {
      // 创建新章节
      const newEpisode = {
        episode_number: episodeNumber,
        title: `第${episodeNumber}集`,
        script_content: scriptContent.value,
      };
      updatedEpisodes = [...existingEpisodes, newEpisode];
    }

    await dramaAPI.saveEpisodes(dramaId, updatedEpisodes);
    ElMessage.success("章节保存成功！");
    await loadDramaData();
  } catch (error: any) {
    ElMessage.error(error.message || "保存失败");
  }
};

const editCurrentEpisodeScript = () => {
  scriptContent.value = currentEpisode.value?.script_content || "";
  isEditingScript.value = true;
};

// 标准化改写状态
const standardizingScript = ref(false);
const standardizeProgress = ref("");

// 通用的任务轮询函数
const pollTaskUntilDone = async (
  taskId: string,
  onProgress?: (msg: string, progress?: number) => void,
): Promise<{ success: boolean; error?: string }> => {
  const maxAttempts = 90;
  for (let i = 0; i < maxAttempts; i++) {
    await new Promise((r) => setTimeout(r, 2000));
    try {
      const task = await generationAPI.getTaskStatus(taskId);
      if (onProgress) {
        onProgress(task.message || `处理中...`, task.progress);
      }
      if (task.status === "completed") {
        return { success: true };
      } else if (task.status === "failed") {
        return { success: false, error: task.error || task.message || "未知错误" };
      }
    } catch (e) {
      console.error("Poll task error:", e);
    }
  }
  return { success: false, error: "任务超时" };
};

// 执行 AI 标准化改写（核心逻辑，可被多处调用）
const doScriptRewrite = async (): Promise<boolean> => {
  const episodeId = currentEpisode.value?.id;
  if (!episodeId) return false;

  // 打开进度弹窗
  rewriteProgressDialog.visible = true;
  rewriteProgressDialog.running = true;
  rewriteProgressDialog.success = false;
  rewriteProgressDialog.error = false;
  rewriteProgressDialog.errorMsg = "";
  rewriteProgressDialog.progressMsg = "正在调用 AI 标准化改写...";
  rewriteProgressDialog.progress = 0;

  standardizingScript.value = true;
  standardizeProgress.value = "正在调用 AI 标准化改写...";
  try {
    const res = await generationAPI.rewriteScript(episodeId, selectedTextModel.value || undefined);
    rewriteProgressDialog.progress = 10;
    const result = await pollTaskUntilDone(res.task_id, (msg, progress) => {
      standardizeProgress.value = msg;
      rewriteProgressDialog.progressMsg = msg;
      if (progress !== undefined) {
        rewriteProgressDialog.progress = progress;
      }
    });
    if (result.success) {
      await loadDramaData();
      // 显示成功状态
      rewriteProgressDialog.running = false;
      rewriteProgressDialog.success = true;
      rewriteProgressDialog.progress = 100;
      // 1.5 秒后自动关闭弹窗
      setTimeout(() => {
        rewriteProgressDialog.visible = false;
      }, 1500);
      return true;
    } else {
      rewriteProgressDialog.running = false;
      rewriteProgressDialog.error = true;
      rewriteProgressDialog.errorMsg = result.error || "改写失败";
      return false;
    }
  } catch (error: any) {
    rewriteProgressDialog.running = false;
    rewriteProgressDialog.error = true;
    rewriteProgressDialog.errorMsg = error.message || "发起改写失败";
    return false;
  } finally {
    standardizingScript.value = false;
    standardizeProgress.value = "";
  }
};

// ======= 剧本设计：构建分镜剧本提示词 =======
const buildScriptPrompt = (userIdea: string, feedback?: string) => {
  const characters = currentEpisode.value?.characters || [];
  const characterContext = characters.map((c: any) => {
    return `${c.name}（${c.appearance || '无外貌描述'}）`;
  }).join('\n');

  const scenes = currentEpisode.value?.scenes || [];
  const sceneContext = scenes.map((s: any) => {
    const loc = s.location || s.name || '未命名';
    const time = s.time || '';
    return `location="${loc}" time="${time}"`;
  }).join('\n');

  const props = episodeProps.value || [];
  const propContext = props.map((p: any) => p.name).join('、');

  const existingScript = currentEpisode.value?.script_content || '';

  let prompt = `你是一位专业的短片分镜剧本编剧。请根据用户的创意构想，生成一份**分镜就绪的三段式剧本**。

=== 目标时长 ===
${episodeDuration.value}秒

=== 用户的剧情构想 ===
${userIdea}
`;

  if (feedback && existingScript) {
    prompt += `
=== 当前剧本（需要修改） ===
${existingScript}

=== 用户的修改意见 ===
${feedback}

请根据以上修改意见，在当前剧本基础上调整，保留满意的部分，修改不满意的部分。
`;
  }

  prompt += `
=== 可用角色（外貌描述用于三段描述中替代名字） ===
${characterContext || '无预设角色'}

=== 可用场景（"场景："字段只写location部分，不要附带时间） ===
${sceneContext || '无预设场景'}

=== 可用道具 ===
${propContext || '无预设道具'}

=== 输出格式（严格遵守，不要添加任何标题、注释、markdown标记） ===

人物：
角色名-装扮（完整外貌描述）
（每个角色一行，直接使用上方提供的外貌信息）

场景：
场景location1（时间段）、场景location2（时间段）...

---

S01 · 镜头标题（时长s）
场景：{只写location值，不带时间} ｜ 造型：{主角角色名-装扮} ｜ 景别：{景别}
角色：{多角色时列出所有角色名-装扮，单角色可省略此行}
首帧：{静态画面描述，30-80字}
过程：{动作+台词，字数按时长控制}
尾帧：{相对首帧的变化，30-80字}
道具：{完整道具名，逗号分隔}

S02 · 镜头标题（时长s）
场景：... ｜ 造型：... ｜ 景别：...
承接上镜：{同场景 or 新场景}（同场景时描述相对上一镜尾帧的镜头变化，如"机位从近景拉远至中景，人物已站起"；新场景时写"切入新场景"即可）
首帧：{同场景时以上一镜尾帧为基础描述当前画面；新场景时为全新独立画面描述，30-80字}
过程：...
尾帧：...
道具：...

（S01无需"承接上镜"，S02起每个镜头必须有。编号用两位数 S01、S02...）

=== 关键规则（违反任何一条都不合格） ===
1. ⚠️ 首帧/过程/尾帧中绝对禁止出现角色名字，必须用穿着特征指代（如"穿米色针织开衫的短卷发女孩"）
2. ⚠️ 用户构想中的台词/独白必须全部保留，不得遗漏任何一句。台词格式：穿着特征+说话动作+「台词内容」
3. ⚠️ 尾帧必须描述相对首帧的变化，格式："穿XX的女孩从A姿态变为B姿态，表情从C变为D"，禁止写独立叙述
4. 过程描述字数与时长匹配：4-5秒→20-50字，6-8秒→60-120字，9-12秒→120-250字
5. 所有镜头时长之和 = ${episodeDuration.value}秒，每个镜头4-12秒
6. "场景："字段只写location值（如"家中卧室"），不要把时间写进去
7. 道具名使用上方提供的完整名称（含角色前缀如"姜小卷-手机"）
8. 同一场景内连续剧情应合并为一个镜头（≤12秒），不要碎片化拆分
9. 首帧必须是纯静态画面，禁止出现动态词汇（"正在""开始"等）
10. 不要输出任何额外的标题、说明、markdown标记（如**粗体**）、代码块，直接从"人物："开始输出
11. ⚠️ 镜头间连续性：S02起必须有"承接上镜"。判断逻辑：若本镜场景与上一镜相同→写"同场景"并描述镜头变化（机位/景别/角度的调整+角色状态延续）；若场景不同→写"新场景"+"切入新场景"，首帧为全新画面
12. ⚠️ 同场景连续镜头中，首帧里的角色状态（姿态、表情、位置）必须与上一镜尾帧一致，不能矛盾`;

  return prompt;
};

const saveDescription = async () => {
  if (!currentEpisode.value?.id || !scriptDraftInput.value.trim()) return;
  try {
    await dramaAPI.saveEpisodeDescription(
      currentEpisode.value.id.toString(),
      scriptDraftInput.value.trim()
    );
    ElMessage.success("剧情描述已保存");
  } catch (error: any) {
    ElMessage.error(error.message || "保存失败");
  }
};

const doGenerateScript = async (feedback?: string) => {
  const episodeId = currentEpisode.value?.id;
  if (!episodeId) {
    ElMessage.error("章节信息不存在");
    return;
  }

  generatingScript.value = true;

  try {
    // 先保存剧情描述
    if (scriptDraftInput.value.trim()) {
      await dramaAPI.saveEpisodeDescription(episodeId.toString(), scriptDraftInput.value.trim());
    }

    const prompt = buildScriptPrompt(scriptDraftInput.value, feedback);

    const response = await dramaAPI.polishPrompt({
      prompt,
      type: 'script',
      orientation: 'horizontal',
      style: 'realistic',
    });

    if (response?.polished_prompt) {
      await dramaAPI.saveEpisodeScript(episodeId.toString(), response.polished_prompt);
      // 直接更新本地数据，不依赖 loadDramaData 的异步刷新
      if (currentEpisode.value) {
        currentEpisode.value.script_content = response.polished_prompt;
      }
      await loadDramaData();
      ElMessage.success(feedback ? "剧本已根据修改意见重新生成！" : "分镜剧本生成成功！");
    } else {
      ElMessage.warning("AI 返回内容为空，请重试");
    }
  } catch (error: any) {
    console.error("Script generation error:", error);
    ElMessage.error(error.message || "剧本生成失败");
  } finally {
    generatingScript.value = false;
  }
};

const generateScript = async () => {
  if (!scriptDraftInput.value.trim()) {
    ElMessage.warning("请输入剧情描述");
    return;
  }
  await doGenerateScript();
};

const regenerateWithFeedback = async () => {
  if (!scriptFeedback.value.trim()) {
    ElMessage.warning("请输入修改意见");
    return;
  }
  showFeedbackDialog.value = false;
  await doGenerateScript(scriptFeedback.value);
};

const submitScriptDirectly = async () => {
  if (!currentEpisode.value?.id || !scriptDraftInput.value.trim()) return;
  try {
    currentEpisode.value.script_content = scriptDraftInput.value.trim();
    await dramaAPI.saveEpisodeScript(
      currentEpisode.value.id.toString(),
      currentEpisode.value.script_content
    );
    ElMessage.success("剧本已提交");
  } catch (error: any) {
    ElMessage.error(error.message || "提交失败");
  }
};

const saveScriptContent = async () => {
  if (!currentEpisode.value?.id) return;
  try {
    await dramaAPI.saveEpisodeScript(
      currentEpisode.value.id.toString(),
      currentEpisode.value.script_content
    );
    ElMessage.success("剧本保存成功");
  } catch (error: any) {
    ElMessage.error(error.message || "保存失败");
  }
};

// 检测当前文本内容质量分数（纯函数，不依赖 computed 缓存）
const checkScriptQualityScore = (text: string): number => {
  if (!text) return 0;
  let passed = 0;
  const total = 5;

  // 1. 长度
  if (text.length >= 100) passed++;

  // 2. 对话
  const dialoguePatterns = [
    /[""\u201c\u201d].{2,}[""\u201c\u201d]/,
    /[\u4e00-\u9fa5]{1,5}[：:]\s*[""\u201c\u201d]/,
    /[\u4e00-\u9fa5]{1,5}[：:]\s*[\u4e00-\u9fa5]{4,}/,
    /（独白）|（旁白）|\(独白\)|\(旁白\)/,
  ];
  if (dialoguePatterns.some((p) => p.test(text))) passed++;

  // 3. 场景
  const scenePatterns = [/【.{2,}】/, /\[.{2,}\]/, /场景|地点|室内|室外|屋内|屋外|房间|客厅|卧室|办公|街道|公园|医院|学校|餐厅|酒吧|居酒屋/];
  if (scenePatterns.some((p) => p.test(text))) passed++;

  // 4. 角色 >=2
  const namePattern = /[\u4e00-\u9fa5]{2,4}(?=[：:"]|说|道|笑|叹|问|答|喊|叫|回|转|看|站|坐|走)/g;
  const uniqueNames = new Set(text.match(namePattern) || []);
  if (uniqueNames.size >= 2) passed++;

  // 5. 动作
  const actionPatterns = [/走|跑|站|坐|躺|拿|放|推|拉|转身|回头|低头|抬头|看向|走向|靠近|离开/];
  if (actionPatterns.some((p) => p.test(text))) passed++;

  return Math.round((passed / total) * 100);
};

const saveEditedScript = async () => {
  try {
    const existingEpisodes = drama.value?.episodes || [];
    const episodeIndex = existingEpisodes.findIndex(
      (ep) => ep.episode_number === episodeNumber,
    );

    let updatedEpisodes;
    if (episodeIndex >= 0) {
      updatedEpisodes = [...existingEpisodes];
      updatedEpisodes[episodeIndex] = {
        ...updatedEpisodes[episodeIndex],
        script_content: scriptContent.value,
      };
    } else {
      const newEpisode = {
        episode_number: episodeNumber,
        title: `第${episodeNumber}集`,
        script_content: scriptContent.value,
      };
      updatedEpisodes = [...existingEpisodes, newEpisode];
    }

    await dramaAPI.saveEpisodes(dramaId, updatedEpisodes);
    isEditingScript.value = false;
    await loadDramaData();

    // === 自动质量检测 + 标准化改写 ===
    const score = checkScriptQualityScore(scriptContent.value);
    if (score < 60) {
      // 质量不足，询问用户是否自动标准化
      ElMessage.success("内容已保存！检测到剧本内容不够完整，正在自动标准化改写...");
      const success = await doScriptRewrite();
      if (success) {
        ElMessage.success("标准化改写完成！剧本已自动更新为包含对话和场景描写的完整格式。");
      }
    } else {
      ElMessage.success("章节内容已保存！内容质量检测通过。");
    }
  } catch (error: any) {
    ElMessage.error(error.message || "保存失败");
  }
};

// AI 手动改写剧本（按钮触发）
const generateDialogueScript = async () => {
  const episodeId = currentEpisode.value?.id;
  if (!episodeId) {
    ElMessage.warning("请先选择章节");
    return;
  }

  try {
    await ElMessageBox.confirm(
      "AI 将把当前内容标准化改写为包含完整对话、动作描写和场景描述的剧本。原始内容将被替换，确定继续？",
      "AI 标准化改写",
      { confirmButtonText: "确认改写", cancelButtonText: "取消", type: "warning" }
    );
  } catch {
    return;
  }

  generatingDialogue.value = true;
  const success = await doScriptRewrite();
  if (success) {
    ElMessage.success("标准化改写完成！剧本内容已更新。");
  }
  generatingDialogue.value = false;
};

const handleExtractCharactersAndBackgrounds = async () => {
  // 如果已经提取过，显示确认对话框
  if (hasExtractedData.value) {
    try {
      await ElMessageBox.confirm(
        $t("workflow.reExtractConfirmMessage"),
        $t("workflow.reExtractConfirmTitle"),
        {
          confirmButtonText: $t("common.confirm"),
          cancelButtonText: $t("common.cancel"),
          type: "warning",
          distinguishCancelAndClose: true,
        },
      );
    } catch {
      ElMessage.info($t("workflow.extractCancelled"));
      return;
    }
  }

  // 显示即将开始的提示
  if (hasExtractedData.value) {
    ElMessage.info($t("workflow.startReExtracting"));
  }

  await extractCharactersAndBackgrounds();
};

// 轮询检查图片生成状态
const pollImageStatus = async (
  imageGenId: number,
  onComplete: () => Promise<void>,
) => {
  const maxAttempts = 100; // 最多轮询100次
  const pollInterval = 6000; // 每6秒轮询一次

  for (let i = 0; i < maxAttempts; i++) {
    try {
      await new Promise((resolve) => setTimeout(resolve, pollInterval));

      const imageGen = await imageAPI.getImage(imageGenId);

      if (imageGen.status === "completed") {
        // 生成成功
        await onComplete();
        return;
      } else if (imageGen.status === "failed") {
        // 生成失败
        ElMessage.error(`图片生成失败: ${imageGen.error_msg || "未知错误"}`);
        return;
      }
      // 如果是pending或processing，继续轮询
    } catch (error: any) {
      console.error("[轮询] 检查图片状态失败:", error);
      // 继续轮询，不中断
    }
  }

  // 超时
  ElMessage.warning("图片生成超时，请稍后刷新页面查看结果");
};

const extractCharactersAndBackgrounds = async () => {
  if (!currentEpisode.value?.id) {
    ElMessage.error("章节信息不存在");
    return;
  }

  extractingCharactersAndBackgrounds.value = true;

  try {
    const episodeId = currentEpisode.value.id;

    // V3: 使用程序解析（同步，秒级完成，不走 AI）
    const result = await generationAPI.parseExtract(episodeId);

    const charCount = result.characters?.length || 0;
    const sceneCount = result.scenes?.length || 0;

    if (charCount === 0 && sceneCount === 0) {
      ElMessage.warning("未从剧本中解析到角色或场景，请确保剧本已经过 AI 标准化改写（包含「人物：」和「场景：」区块）");
    } else {
      ElMessage.success(`提取完成：${charCount} 个角色，${sceneCount} 个场景`);
    }

    await loadDramaData();
  } catch (error: any) {
    console.error("提取角色和场景失败:", error);
    const errorMsg = error.response?.data?.error?.message || error.message || "提取失败";
    ElMessage.error(errorMsg);
  } finally {
    extractingCharactersAndBackgrounds.value = false;
  }
};

// 轮询提取任务状态
const generateCharacterImage = async (characterId: number, customPrompt?: string, referenceImages?: string[]) => {
  generatingCharacterImages.value[characterId] = true;

  try {
    const model = selectedImageToImageModel.value || selectedTextToImageModel.value || undefined;
    const response = await characterLibraryAPI.generateCharacterImage(
      characterId.toString(),
      model,
      customPrompt,
      referenceImages
    );
    const imageGenId = response.image_generation?.id;

    if (imageGenId) {
      ElMessage.info("角色图片生成中，请稍候...");
      await pollImageStatus(imageGenId, async () => {
        await loadDramaData();
        ElMessage.success("角色图片生成完成！");
      });
    } else {
      ElMessage.success("角色图片生成已启动");
      await loadDramaData();
    }
  } catch (error: any) {
    ElMessage.error(error.message || "生成失败");
  } finally {
    generatingCharacterImages.value[characterId] = false;
  }
};

const toggleSelectAllCharacters = () => {
  if (selectAllCharacters.value) {
    selectedCharacterIds.value =
      currentEpisode.value?.characters?.map((char) => char.id) || [];
  } else {
    selectedCharacterIds.value = [];
  }
};

const toggleSelectAllScenes = () => {
  if (selectAllScenes.value) {
    selectedSceneIds.value =
      currentEpisode.value?.scenes?.map((scene) => scene.id) || [];
  } else {
    selectedSceneIds.value = [];
  }
};

const batchGenerateCharacterImages = async () => {
  if (selectedCharacterIds.value.length === 0) {
    ElMessage.warning("请先选择要生成的角色");
    return;
  }

  batchGeneratingCharacters.value = true;
  try {
    // 获取用户选择的图片生成模型
    const model = selectedImageModel.value || undefined;

    // 使用批量生成API
    await characterLibraryAPI.batchGenerateCharacterImages(
      selectedCharacterIds.value.map((id) => id.toString()),
      model
    );

    ElMessage.success($t("workflow.batchTaskSubmitted"));
    await loadDramaData();
  } catch (error: any) {
    ElMessage.error(error.message || $t("workflow.batchGenerateFailed"));
  } finally {
    batchGeneratingCharacters.value = false;
  }
};

const generateSceneImage = async (sceneId: string, customPrompt?: string, referenceImages?: string[]) => {
  generatingSceneImages.value[sceneId] = true;

  try {
    const model = selectedImageToImageModel.value || selectedTextToImageModel.value || undefined;
    const response = await dramaAPI.generateSceneImage({
      scene_id: parseInt(sceneId),
      model,
      prompt: customPrompt,
      reference_images: referenceImages
    });
    const imageGenId = response.image_generation?.id;

    if (imageGenId) {
      ElMessage.info($t("workflow.sceneImageGenerating"));
      await pollImageStatus(imageGenId, async () => {
        await loadDramaData();
        ElMessage.success($t("workflow.sceneImageComplete"));
      });
    } else {
      ElMessage.success($t("workflow.sceneImageStarted"));
      await loadDramaData();
    }
  } catch (error: any) {
    ElMessage.error(error.message || "生成失败");
  } finally {
    generatingSceneImages.value[sceneId] = false;
  }
};

const batchGenerateSceneImages = async () => {
  if (selectedSceneIds.value.length === 0) {
    ElMessage.warning("请先选择要生成的场景");
    return;
  }

  batchGeneratingScenes.value = true;
  try {
    const promises = selectedSceneIds.value.map((sceneId) =>
      generateSceneImage(sceneId.toString()),
    );
    const results = await Promise.allSettled(promises);

    const successCount = results.filter((r) => r.status === "fulfilled").length;
    const failCount = results.filter((r) => r.status === "rejected").length;

    if (failCount === 0) {
      ElMessage.success(
        $t("workflow.batchCompleteSuccess", { count: successCount }),
      );
    } else {
      ElMessage.warning(
        $t("workflow.batchCompletePartial", {
          success: successCount,
          fail: failCount,
        }),
      );
    }
  } catch (error: any) {
    ElMessage.error(error.message || $t("workflow.batchGenerateFailed"));
  } finally {
    batchGeneratingScenes.value = false;
  }
};

const taskProgress = ref(0);
const taskMessage = ref("");
let pollTimer: any = null;

// AI 改写进度弹窗状态
const rewriteProgressDialog = reactive({
  visible: false,
  running: false,
  success: false,
  error: false,
  errorMsg: "",
  progressMsg: "正在调用 AI 标准化改写...",
  progress: 0,
});

// 拆分进度弹窗状态
const splitProgressDialog = reactive({
  visible: false,
  running: false,
  success: false,
  error: false,
  errorMsg: "",
});

const generateShots = async () => {
  if (!currentEpisode.value?.id) {
    ElMessage.error("章节信息不存在");
    return;
  }

  // 清空当前的分镜列表
  if (currentEpisode.value.storyboards) {
    currentEpisode.value.storyboards = [];
  }

  // 打开进度弹窗
  splitProgressDialog.visible = true;
  splitProgressDialog.running = true;
  splitProgressDialog.success = false;
  splitProgressDialog.error = false;
  splitProgressDialog.errorMsg = "";

  generatingShots.value = true;
  taskProgress.value = 0;
  taskMessage.value = "初始化任务...";

  try {
    const episodeId = currentEpisode.value.id.toString();

    // 创建异步任务
    const response = await generationAPI.generateStoryboard(
      episodeId,
      selectedTextModel.value,
      shotCount.value,
    );

    taskMessage.value = response.message || "任务已创建";

    // 开始轮询任务状态
    await pollTaskStatus(response.task_id);
  } catch (error: any) {
    ElMessage.error(error.message || "拆分失败");
    generatingShots.value = false;
    splitProgressDialog.running = false;
    splitProgressDialog.error = true;
    splitProgressDialog.errorMsg = error.message || "拆分失败";
  }
};

const pollTaskStatus = async (taskId: string) => {
  const checkStatus = async () => {
    try {
      const task = await generationAPI.getTaskStatus(taskId);

      taskProgress.value = task.progress;
      taskMessage.value = task.message || `处理中... ${task.progress}%`;

      if (task.status === "completed") {
        // 任务完成
        if (pollTimer) {
          clearInterval(pollTimer);
          pollTimer = null;
        }
        generatingShots.value = false;

        // 更新弹窗：显示成功
        splitProgressDialog.running = false;
        splitProgressDialog.success = true;
        taskProgress.value = 100;

        // 刷新数据
        await loadDramaData();

        ElMessage.success($t("workflow.splitSuccess"));

        // 1.5秒后自动关闭弹窗并跳转
        setTimeout(() => {
          splitProgressDialog.visible = false;
          router.push({
            name: "ProfessionalEditor",
            params: {
              dramaId: dramaId,
              episodeNumber: episodeNumber,
            },
          });
        }, 1500);
      } else if (task.status === "failed") {
        // 任务失败
        if (pollTimer) {
          clearInterval(pollTimer);
          pollTimer = null;
        }
        generatingShots.value = false;
        // 更新弹窗：显示失败
        splitProgressDialog.running = false;
        splitProgressDialog.error = true;
        splitProgressDialog.errorMsg = task.error || "分镜拆分失败";
        ElMessage.error(task.error || "分镜拆分失败");
      }
      // 否则继续轮询
    } catch (error: any) {
      if (pollTimer) {
        clearInterval(pollTimer);
        pollTimer = null;
      }
      generatingShots.value = false;
      // 更新弹窗：显示失败
      splitProgressDialog.running = false;
      splitProgressDialog.error = true;
      splitProgressDialog.errorMsg = "查询任务状态失败: " + error.message;
      ElMessage.error("查询任务状态失败: " + error.message);
    }
  };

  // 立即检查一次
  await checkStatus();

  // 每2秒轮询一次
  pollTimer = setInterval(checkStatus, 2000);
};

const regenerateShots = async () => {
  await ElMessageBox.confirm($t("workflow.reSplitConfirm"), $t("common.tip"), {
    type: "warning",
  });

  await generateShots();
};

const shotEditDialogVisible = ref(false);
const editingShot = ref<any>(null);
const editingShotIndex = ref<number>(-1);
const savingShot = ref(false);

const editShot = (shot: any, index: number) => {
  editingShot.value = { ...shot };
  editingShotIndex.value = index;
  shotEditDialogVisible.value = true;
};

const saveShotEdit = async () => {
  if (!editingShot.value) return;

  try {
    savingShot.value = true;

    // 调用API更新镜头
    await dramaAPI.updateStoryboard(
      editingShot.value.id.toString(),
      editingShot.value,
    );

    // 更新本地数据
    if (currentEpisode.value?.storyboards) {
      currentEpisode.value.storyboards[editingShotIndex.value] = {
        ...editingShot.value,
      };
    }

    ElMessage.success("镜头修改成功");
    shotEditDialogVisible.value = false;
  } catch (error: any) {
    ElMessage.error("保存失败: " + (error.message || "未知错误"));
  } finally {
    savingShot.value = false;
  }
};

// 对话框相关方法
const openPromptDialog = (item: any, type: "character" | "scene") => {
  currentEditItem.value = item;
  if (type === "scene") {
    currentEditItem.value.name = item.name || item.location;
  }
  currentEditType.value = type;
  
  // 人物和场景都默认竖屏
  editOrientation.value = "portrait";
  
  // 加载参考图片
  editReferenceImages.value = [];
  if (item.reference_images && Array.isArray(item.reference_images)) {
    editReferenceImages.value = item.reference_images.map((ref: any) => {
      if (typeof ref === 'string') {
        const name = ref.split('/').pop() || '参考图片';
        return { name, path: ref };
      } else if (ref.name && ref.path) {
        return ref;
      }
      return null;
    }).filter(Boolean);
  }
  
  // 获取原始提示词
  let originalPrompt = "";
  if (type === "character") {
    originalPrompt = item.appearance || item.description || "";
  } else {
    originalPrompt = item.prompt || `${item.location}, ${item.time}`;
  }
  
  // 清理提示词：移除自动添加的内容（如果存在）
  let cleanedPrompt = originalPrompt;
  
  // 移除构图要求（如果存在）
  const compositionText = "，正面照，全身照，纯色背景，干净背景，简洁背景，站姿，清晰的人物轮廓，专业人物摄影";
  const compositionPattern = new RegExp(compositionText.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g');
  cleanedPrompt = cleanedPrompt.replace(compositionPattern, "");
  
  // 移除重复的构图要求（单个关键词）
  const compositionKeywords = ["正面照", "全身照", "纯色背景", "干净背景", "简洁背景", "站姿", "清晰的人物轮廓", "专业人物摄影"];
  compositionKeywords.forEach(keyword => {
    // 统计关键词出现次数
    const regex = new RegExp(keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g');
    const matches = cleanedPrompt.match(regex);
    if (matches && matches.length > 1) {
      // 只保留第一个
      let firstIndex = cleanedPrompt.indexOf(keyword);
      cleanedPrompt = cleanedPrompt.slice(0, firstIndex + keyword.length) + 
                     cleanedPrompt.slice(firstIndex + keyword.length).replace(regex, "");
    }
  });
  
  // 移除种族特征（如果存在）
  const ethnicityText = "中国面孔，亚洲人特征，";
  if (cleanedPrompt.startsWith(ethnicityText)) {
    cleanedPrompt = cleanedPrompt.slice(ethnicityText.length);
  }
  
  // 移除参考图片说明（如果存在）
  const refDescStart = cleanedPrompt.indexOf("\n\n输入的参考图片说明");
  if (refDescStart !== -1) {
    cleanedPrompt = cleanedPrompt.slice(0, refDescStart);
  }
  
  // 移除横竖屏标记（如果存在）
  cleanedPrompt = cleanedPrompt.replace(/（\d+:\d+[横竖]屏）/g, "");
  
  // 移除末尾的逗号和空格
  cleanedPrompt = cleanedPrompt.replace(/，\s*$/g, "").replace(/\s*$/g, "");
  
  // 保存清理后的原始提示词
  originalEditPrompt.value = cleanedPrompt;
  
  // 直接使用清理后的提示词，不添加任何预设内容
  editPrompt.value = cleanedPrompt;
  
  promptDialogVisible.value = true;
};

const savePrompt = async () => {
  try {
    let promptToSave = editPrompt.value.replace(/（\d+:\d+[横竖]屏）/g, "").replace(/，\s*$/g, "").replace(/\s*$/g, "");
    
    if (currentEditType.value === "character") {
      await characterLibraryAPI.updateCharacter(currentEditItem.value.id, {
        name: currentEditItem.value.name,
        appearance: promptToSave,
        reference_images: editReferenceImages.value,
      });
    } else if (currentEditType.value === "scene") {
      await dramaAPI.updateScenePrompt(
        currentEditItem.value.id.toString(),
        currentEditItem.value.name,
        promptToSave,
        editReferenceImages.value
      );
    } else if (currentEditType.value === "prop") {
      await propAPI.update(currentEditItem.value.id, {
        prompt: promptToSave,
        reference_images: editReferenceImages.value,
      });
    }
    await loadDramaData();
    promptDialogVisible.value = false;
    ElMessage.success("提示词保存成功");
  } catch (error: any) {
    ElMessage.error(error.message || "保存失败");
  }
};

const savePromptWithoutClose = async () => {
  try {
    // 保存时移除横竖屏标记，只保存用户实际编辑的内容
    let promptToSave = editPrompt.value.replace(/（\d+:\d+[横竖]屏）/g, "").replace(/，\s*$/g, "").replace(/\s*$/g, "");
    
    if (currentEditType.value === "character") {
      await characterLibraryAPI.updateCharacter(currentEditItem.value.id, {
        name: currentEditItem.value.name,
        appearance: promptToSave,
        reference_images: editReferenceImages.value,
      });
    } else {
      await dramaAPI.updateScenePrompt(
        currentEditItem.value.id.toString(),
        currentEditItem.value.name,
        promptToSave,
        editReferenceImages.value
      );
    }
    await loadDramaData();
  } catch (error: any) {
    ElMessage.error(error.message || "保存失败");
    throw error;
  }
};

const generateImageWithSize = async () => {
  generatingEditItem.value = true;
  
  try {
    await savePrompt();
    
    const referenceImagePaths = editReferenceImages.value.map((r) => r.path);
    
    if (currentEditType.value === "character") {
      await generateCharacterImage(
        currentEditItem.value.id,
        editPrompt.value,
        referenceImagePaths
      );
    } else {
      await generateSceneImage(
        currentEditItem.value.id.toString(),
        editPrompt.value,
        referenceImagePaths
      );
    }
    
    promptDialogVisible.value = false;
  } catch (error: any) {
    ElMessage.error(error.message || "生成失败");
  } finally {
    generatingEditItem.value = false;
  }
};

const polishPromptWithAI = async () => {
  if (!editPrompt.value.trim()) {
    ElMessage.warning("请先输入提示词");
    return;
  }
  
  if (!drama.value?.style) {
    ElMessage.warning("项目风格未设置");
    return;
  }
  
  polishingPrompt.value = true;
  try {
    // 清理当前的提示词，移除自动添加的内容
    let promptToPolish = editPrompt.value;
    
    // 移除横竖屏标记
    promptToPolish = promptToPolish.replace(/（\d+:\d+[横竖]屏）/g, "");
    
    // 移除参考图片说明
    const refDescStart = promptToPolish.indexOf("\n\n输入的参考图片说明");
    if (refDescStart !== -1) {
      promptToPolish = promptToPolish.slice(0, refDescStart);
    }
    
    // 移除末尾的逗号和空格
    promptToPolish = promptToPolish.replace(/，\s*$/g, "").replace(/\s*$/g, "");
    
    // 调用后端API进行AI润色，使用项目的style
    const response = await dramaAPI.polishPrompt({
      prompt: promptToPolish,
      type: currentEditType.value,
      orientation: editOrientation.value,
      style: drama.value.style,
      reference_images: editReferenceImages.value.map((r) => r.path)
    });
    
    // 使用AI润色后的提示词
    editPrompt.value = response.polished_prompt;
    
    // 自动保存到数据库（不关闭对话框）
    await savePromptWithoutClose();
    
    ElMessage.success("提示词润色并保存成功！");
  } catch (error: any) {
    ElMessage.error(error.message || "润色失败");
  } finally {
    polishingPrompt.value = false;
  }
};

const getStyleName = (style: string): string => {
  const styleNames: Record<string, string> = {
    realistic: '写实',
    comic: '漫画',
  };
  return styleNames[style] || style;
};

const uploadCharacterImage = (characterId: number) => {
  currentUploadTarget.value = { id: characterId, type: "character" };
  uploadDialogVisible.value = true;
};

const uploadSceneImage = (sceneId: string) => {
  currentUploadTarget.value = { id: sceneId, type: "scene" };
  uploadDialogVisible.value = true;
};

const selectFromLibrary = async (characterId: number) => {
  try {
    const result = await characterLibraryAPI.list({ page_size: 50 });
    libraryItems.value = result.items || [];
    currentUploadTarget.value = characterId;
    libraryDialogVisible.value = true;
  } catch (error: any) {
    ElMessage.error(error.message || $t("workflow.loadLibraryFailed"));
  }
};

const addToCharacterLibrary = async (character: any) => {
  if (!character.image_url) {
    ElMessage.warning($t("workflow.generateImageFirst"));
    return;
  }

  try {
    await ElMessageBox.confirm(
      $t("workflow.addToLibraryConfirm", { name: character.name }),
      $t("workflow.addToLibrary"),
      {
        confirmButtonText: $t("common.confirm"),
        cancelButtonText: $t("common.cancel"),
        type: "info",
      },
    );

    await characterLibraryAPI.addCharacterToLibrary(character.id.toString());
    ElMessage.success($t("workflow.addedToLibrary"));
  } catch (error: any) {
    if (error !== "cancel") {
      ElMessage.error(error.message || $t("workflow.addFailed"));
    }
  }
};

const selectLibraryItem = async (item: any) => {
  try {
    if (currentUploadTarget.value?.type === "character") {
      // 给现有角色应用库中的形象
      await characterLibraryAPI.applyFromLibrary(
        currentUploadTarget.value.id.toString(),
        item.id,
      );
      ElMessage.success("应用角色形象成功！");
      await loadDramaData();
      libraryDialogVisible.value = false;
    } else if (currentUploadTarget.value?.type === "fromLibrary") {
      // 将已有角色关联到当前章节（引用，不创建副本）
      const episodeId = currentEpisode.value?.id;
      if (!episodeId) {
        ElMessage.warning("当前章节信息不可用");
        return;
      }
      try {
        await dramaAPI.addCharacterToEpisode(episodeId, item.id);
        ElMessage.success("角色已关联到本章节！");
        await loadDramaData();
        libraryDialogVisible.value = false;
      } catch (createError: any) {
        ElMessage.error(createError.message || "关联角色失败");
      }
    } else if (currentUploadTarget.value?.type === "fromLibraryScene") {
      // 将已有场景关联到当前章节（引用，不创建副本）
      const episodeId = currentEpisode.value?.id;
      if (!episodeId) {
        ElMessage.warning("当前章节信息不可用");
        return;
      }
      try {
        await dramaAPI.addSceneToEpisode(episodeId, item.id);
        ElMessage.success("场景已关联到本章节！");
        await loadDramaData();
        libraryDialogVisible.value = false;
      } catch (createError: any) {
        ElMessage.error(createError.message || "关联场景失败");
      }
    } else if (currentUploadTarget.value?.type === "fromLibraryProp") {
      // 将已有道具关联到当前章节（引用，不创建副本）
      const episodeId = currentEpisode.value?.id;
      if (!episodeId) {
        ElMessage.warning("当前章节信息不可用");
        return;
      }
      try {
        await dramaAPI.addPropToEpisode(episodeId, item.id);
        ElMessage.success("道具已关联到本章节！");
        await loadDramaData();
        libraryDialogVisible.value = false;
      } catch (createError: any) {
        ElMessage.error(createError.message || "关联道具失败");
      }
    }
  } catch (error: any) {
    ElMessage.error(error.message || "操作失败");
  }
};

const handleUploadSuccess = async (response: any) => {
  try {
    const imageUrl = response.url || response.data?.url;
    const localPath = response.local_path || response.data?.local_path;

    if (!imageUrl && !localPath) {
      ElMessage.error("上传失败：未获取到图片地址");
      return;
    }

    if (currentUploadTarget.value?.type === "character") {
      await characterLibraryAPI.updateCharacter(
        currentUploadTarget.value.id.toString(),
        {
          image_url: imageUrl,
          local_path: localPath,
        },
      );
      ElMessage.success("上传成功！");
    } else if (currentUploadTarget.value?.type === "scene") {
      // 更新场景图片
      await dramaAPI.updateScene(currentUploadTarget.value.id.toString(), {
        image_url: imageUrl,
        local_path: localPath,
      });
      ElMessage.success($t("workflow.sceneImageUploadSuccess"));
    }

    await loadDramaData();
    uploadDialogVisible.value = false;
  } catch (error: any) {
    ElMessage.error(error.message || "上传失败");
  }
};

const handleUploadError = () => {
  ElMessage.error("上传失败，请重试");
};

const removeCharacterFromEpisode = async (characterId: number) => {
  const episodeId = currentEpisode.value?.id;
  if (!episodeId) {
    ElMessage.warning("当前章节信息不可用");
    return;
  }
  try {
    await ElMessageBox.confirm(
      "确定要将该角色从本章节移除吗？（角色本身不会被删除，可在角色管理中查看）",
      "移除确认",
      {
        type: "warning",
        confirmButtonText: "确定移除",
        cancelButtonText: "取消",
      },
    );
    await dramaAPI.removeCharacterFromEpisode(episodeId, characterId);
    ElMessage.success("角色已从本章节移除");
    await loadDramaData();
  } catch (error: any) {
    if (error !== "cancel") {
      ElMessage.error(error.message || "移除失败");
    }
  }
};

const removeSceneFromEpisode = async (sceneId: number) => {
  const episodeId = currentEpisode.value?.id;
  if (!episodeId) {
    ElMessage.warning("当前章节信息不可用");
    return;
  }
  try {
    await ElMessageBox.confirm(
      "确定要将该场景从本章节移除吗？（场景本身不会被删除，可在场景管理中查看）",
      "移除确认",
      {
        type: "warning",
        confirmButtonText: "确定移除",
        cancelButtonText: "取消",
      },
    );
    await dramaAPI.removeSceneFromEpisode(episodeId, sceneId);
    ElMessage.success("场景已从本章节移除");
    await loadDramaData();
  } catch (error: any) {
    if (error !== "cancel") {
      ElMessage.error(error.message || "移除失败");
    }
  }
};

const removePropFromEpisode = async (propId: number) => {
  const episodeId = currentEpisode.value?.id;
  if (!episodeId) {
    ElMessage.warning("当前章节信息不可用");
    return;
  }
  try {
    await ElMessageBox.confirm(
      "确定要将该道具从本章节移除吗？（道具本身不会被删除，可在道具管理中查看）",
      "移除确认",
      {
        type: "warning",
        confirmButtonText: "确定移除",
        cancelButtonText: "取消",
      },
    );
    await dramaAPI.removePropFromEpisode(episodeId, propId);
    ElMessage.success("道具已从本章节移除");
    await loadDramaData();
  } catch (error: any) {
    if (error !== "cancel") {
      ElMessage.error(error.message || "移除失败");
    }
  }
};

const generatePropImage = async (propId: number) => {
  try {
    generatingPropImages.value[propId] = true;
    const response = await propAPI.generateImage(propId);
    ElMessage.success("道具图片生成已启动");
    
    if (response.task_id) {
      await pollImageGenerationTask(response.task_id);
    }
    
    await loadDramaData();
  } catch (error: any) {
    ElMessage.error(error.message || "生成失败");
  } finally {
    generatingPropImages.value[propId] = false;
  }
};

const editPropPrompt = (prop: Prop) => {
  currentEditItem.value = prop;
  currentEditType.value = "prop";
  editPrompt.value = prop.prompt || "";
  editReferenceImages.value = prop.reference_images || [];
  promptDialogVisible.value = true;
};

const openAddPropDialog = () => {
  newProp.value = {
    name: "",
    type: "",
    description: "",
    prompt: "",
    image_url: "",
  };
  addPropDialogVisible.value = true;
};

const saveProp = async () => {
  try {
    if (!newProp.value.name) {
      ElMessage.warning("请输入道具名称");
      return;
    }

    const createdProp = await propAPI.create({
      drama_id: dramaId,
      name: newProp.value.name,
      type: newProp.value.type,
      description: newProp.value.description,
      prompt: newProp.value.prompt,
      image_url: newProp.value.image_url,
    });

    // 关联到当前章节
    const episodeId = currentEpisode.value?.id;
    if (episodeId && createdProp?.id) {
      await dramaAPI.addPropToEpisode(episodeId, createdProp.id);
    }

    ElMessage.success("道具添加成功");
    addPropDialogVisible.value = false;
    await loadDramaData();
  } catch (error: any) {
    ElMessage.error(error.message || "添加失败");
  }
};

// 打开道具库对话框（从当前项目的道具管理中选择）
const openPropLibraryDialog = async () => {
  try {
    const items = await propAPI.list(dramaId);
    libraryItems.value = items || [];
    currentUploadTarget.value = { type: "fromLibraryProp" };
    libraryDialogVisible.value = true;
  } catch (error: any) {
    ElMessage.error(error.message || "加载道具列表失败");
    libraryItems.value = [];
  }
};

// 获取库对话框标题
const getLibraryDialogTitle = () => {
  if (currentUploadTarget.value?.type === "fromLibraryScene") {
    return "从场景库选择";
  } else if (currentUploadTarget.value?.type === "fromLibraryProp") {
    return "从道具库选择";
  } else {
    return "从角色库选择";
  }
};

const goToProfessionalUI = () => {
  if (!currentEpisode.value?.id) {
    ElMessage.error("章节信息不存在");
    return;
  }

  router.push({
    name: "ProfessionalEditor",
    params: {
      dramaId: dramaId,
      episodeNumber: episodeNumber,
    },
  });
};

const goToCompose = () => {
  if (!currentEpisode.value?.id) {
    ElMessage.error("章节信息不存在");
    return;
  }

  router.push({
    name: "SceneComposition",
    params: {
      id: dramaId,
      episodeId: currentEpisode.value.id,
    },
  });
};

// 打开添加场景对话框
const openAddSceneDialog = () => {
  newScene.value = {
    location: "",
    time: "",
    prompt: "",
    image_url: "",
    local_path: "",
  };
  addSceneDialogVisible.value = true;
};

// 保存场景
const saveScene = async () => {
  if (!newScene.value.location) {
    ElMessage.warning($t("workflow.pleaseEnterSceneName"));
    return;
  }

  if (!currentEpisode.value?.id) {
    ElMessage.error($t("workflow.chapterInfoNotExist"));
    return;
  }

  try {
    // 创建场景，关联到当前章节
    await dramaAPI.createScene({
      drama_id: parseInt(dramaId),
      episode_id: parseInt(currentEpisode.value.id),
      location: newScene.value.location,
      time: newScene.value.time || "",
      prompt: newScene.value.prompt,
      image_url: newScene.value.image_url,
      local_path: newScene.value.local_path,
    });

    ElMessage.success($t("workflow.sceneAddSuccess"));
    addSceneDialogVisible.value = false;

    // 重新加载数据以更新场景列表
    await loadDramaData();
  } catch (error: any) {
    ElMessage.error(error.message || $t("workflow.sceneAddFailed"));
  }
};

// 打开场景库对话框（从当前项目的场景管理中选择）
const openSceneLibraryDialog = async () => {
  try {
    const items = await sceneAPI.list(dramaId);
    libraryItems.value = items || [];
    currentUploadTarget.value = { type: "fromLibraryScene" };
    libraryDialogVisible.value = true;
  } catch (error: any) {
    ElMessage.error(error.message || "加载场景列表失败");
    libraryItems.value = [];
  }
};

// 打开添加角色对话框
const openAddCharacterDialog = () => {
  newCharacter.value = {
    name: "",
    role: "",
    appearance: "",
    personality: "",
  };
  addCharacterDialogVisible.value = true;
};

// 打开角色库对话框（从当前项目的角色管理中选择）
// 优先展示装扮角色（children），无装扮的独立角色也展示
const openCharacterLibraryDialog = async () => {
  try {
    const allChars = await characterAPI.list(dramaId);
    const flatList: any[] = [];
    for (const char of (allChars || [])) {
      if (char.children && char.children.length > 0) {
        for (const child of char.children) {
          flatList.push({
            ...child,
            _parentName: char.name,
            _displayName: `${char.name} · ${child.outfit_name || child.name}`,
          });
        }
      } else {
        flatList.push({
          ...char,
          _displayName: char.name,
        });
      }
    }
    libraryItems.value = flatList;
    currentUploadTarget.value = { type: "fromLibrary" };
    libraryDialogVisible.value = true;
  } catch (error: any) {
    ElMessage.error(error.message || "加载角色列表失败");
    libraryItems.value = [];
  }
};

// 保存角色
const saveCharacter = async () => {
  if (!newCharacter.value.name) {
    ElMessage.warning("请输入角色名称");
    return;
  }

  if (!currentEpisode.value?.id) {
    ElMessage.error("章节信息不存在");
    return;
  }

  try {
    await dramaAPI.createCharacter({
      drama_id: parseInt(dramaId),
      episode_id: parseInt(currentEpisode.value.id),
      name: newCharacter.value.name,
      role: newCharacter.value.role,
      appearance: newCharacter.value.appearance,
      personality: newCharacter.value.personality,
    });

    ElMessage.success("角色添加成功");
    addCharacterDialogVisible.value = false;
    await loadDramaData();
  } catch (error: any) {
    ElMessage.error(error.message || "角色添加失败");
  }
};

// 处理场景图片上传成功
const handleSceneImageSuccess = (response: any) => {
  // 处理不同的响应结构
  const imageUrl = response.url || response.data?.url;
  const localPath = response.local_path || response.data?.local_path;

  if (imageUrl) {
    newScene.value.image_url = imageUrl;
  }
  if (localPath) {
    newScene.value.local_path = localPath;
  }

  if (imageUrl || localPath) {
    ElMessage.success($t("workflow.imageUploadSuccess"));
  } else {
    ElMessage.warning($t("workflow.imageUploadSuccessNoUrl"));
  }
};

// 图片上传前的校验
const beforeAvatarUpload = (file: File) => {
  const isImage = file.type.startsWith("image/");
  const isLt10M = file.size / 1024 / 1024 < 10;

  if (!isImage) {
    ElMessage.error("只能上传图片文件!");
    return false;
  }
  if (!isLt10M) {
    ElMessage.error("图片大小不能超过 10MB!");
    return false;
  }
  return true;
};

// 参考图上传成功处理
const handleReferenceImageUploadSuccess = (response: any) => {
  const imageUrl = response.url || response.data?.url;
  const localPath = response.local_path || response.data?.local_path;
  
  if (imageUrl || localPath) {
    const path = localPath || imageUrl;
    const name = `参考图${editReferenceImages.value.length + 1}`;
    editReferenceImages.value.push({ name, path });
    
    // 重新构建提示词，更新参考图说明
    if (originalEditPrompt.value) {
      const promptWithOrientation = getEpisodePromptWithOrientation(originalEditPrompt.value, editOrientation.value);
      editPrompt.value = buildPromptWithReferences(promptWithOrientation, editReferenceImages.value);
    }
    
    ElMessage.success("参考图上传成功");
  } else {
    ElMessage.warning("上传成功但未获取到图片路径");
  }
};

// 删除参考图
const removeReferenceImage = (index: number) => {
  editReferenceImages.value.splice(index, 1);
  
  // 重新编号剩余的参考图
  editReferenceImages.value.forEach((ref, idx) => {
    ref.name = `参考图${idx + 1}`;
  });
  
  // 重新构建提示词，更新参考图说明
  if (originalEditPrompt.value) {
    const promptWithOrientation = getEpisodePromptWithOrientation(originalEditPrompt.value, editOrientation.value);
    editPrompt.value = buildPromptWithReferences(promptWithOrientation, editReferenceImages.value);
  }
};

// 构建带参考图信息的提示词
const buildPromptWithReferences = (basePrompt: string, references: { name: string; path: string }[]) => {
  if (references.length === 0) {
    return basePrompt;
  }
  
  const refDescLines = references.map((img, idx) => {
    return `【参考图片${idx + 1}】${img.name}`;
  });
  
  return basePrompt + `\n\n输入的参考图片说明（按顺序对应输入的图片）：\n${refDescLines.join("\n")}\n请严格按照以上参考图片来生成图片，保持场景环境和角色外貌的一致性。`;
};

// 打开从剧本提取场景对话框
const openExtractSceneDialog = () => {
  extractScenesDialogVisible.value = true;
};

// 从剧本提取场景
const handleExtractScenes = async () => {
  if (!currentEpisode.value?.id) {
    ElMessage.error($t("workflow.chapterInfoNotExist"));
    return;
  }

  try {
    extractingScenes.value = true;
    await dramaAPI.extractBackgrounds(currentEpisode.value.id.toString());

    ElMessage.success($t("workflow.sceneExtractSubmitted"));
    extractScenesDialogVisible.value = false;

    // 自动刷新几次
    let checkCount = 0;
    const maxChecks = 5;
    const checkInterval = setInterval(async () => {
      checkCount++;
      await loadDramaData();

      if (checkCount >= maxChecks) {
        clearInterval(checkInterval);
      }
    }, 3000);
  } catch (error: any) {
    ElMessage.error(error.message || $t("workflow.sceneExtractFailed"));
  } finally {
    extractingScenes.value = false;
  }
};

// 监听步骤变化，保存到 localStorage
watch(currentStep, (newStep) => {
  localStorage.setItem(getStepStorageKey(), newStep.toString());
});

onMounted(() => {
  loadDramaData();
  loadAIConfigs();
});

watch(currentEpisode, (ep) => {
  if (ep?.description && !scriptDraftInput.value) {
    scriptDraftInput.value = ep.description;
  }
}, { immediate: true });
</script>

<style scoped lang="scss">
/* ========================================
   Page Layout / 页面布局 - 紧凑边距
   ======================================== */
.page-container {
  min-height: 100vh;
  background: var(--bg-primary);
  // padding: var(--space-2) var(--space-3);
  transition: background var(--transition-normal);
}

@media (min-width: 768px) {
  .page-container {
    // padding: var(--space-3) var(--space-4);
  }
}

@media (min-width: 1024px) {
  .page-container {
    // padding: var(--space-4) var(--space-5);
  }
}

.content-wrapper {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0 auto;
  width: 100%;
  height: 100vh;
  overflow: hidden;
}

.content-container {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.actions-container {
  min-height: 70px;
  background: var(--glass-bg-heavy);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));
  border-top: 1px solid var(--glass-border);
  flex-shrink: 0;
}

/* Header styles matching PageHeader component */
.page-header {
  margin-bottom: var(--space-3);
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--border-primary);
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex-shrink: 0;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 0.875rem;
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;

  &:hover {
    background: var(--bg-card-hover);
    color: var(--text-primary);
    border-color: var(--border-secondary);
  }
}

.nav-divider {
  width: 1px;
  height: 2rem;
  background: var(--border-primary);
}

.header-title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.025em;
  line-height: 1.2;
  white-space: nowrap;
}

.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.header-right {
  flex-shrink: 0;
}

.workflow-card {
  height: calc(100% - 24px);
  margin: var(--space-3);
  background: var(--glass-bg-heavy);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-card);
  border: 1px solid var(--glass-border);
  transition: all var(--transition-normal);

  &:hover {
    box-shadow: var(--shadow-card-hover);
  }

  :deep(.el-card__body) {
    padding: 0;
  }
}

.custom-steps {
  display: flex;
  align-items: center;
  gap: var(--space-3);

  .step-item {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-4);
    border-radius: var(--radius-full);
    cursor: pointer;
    user-select: none;
    background: var(--glass-bg);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border: 1px solid var(--glass-border);
    transition: all var(--transition-normal);

    &.active {
      background: var(--accent-light);
      border-color: var(--accent);

      .step-circle {
        background: var(--accent);
        color: var(--text-inverse);
      }
    }

    &.current {
      background: var(--accent);
      color: var(--text-inverse);
      border-color: var(--accent);
      box-shadow: 0 0 16px rgba(14, 165, 233, 0.3);

      .step-circle {
        background: var(--bg-card);
        color: var(--accent);
      }

      .step-text {
        color: var(--text-inverse);
      }
    }

    .step-circle {
      width: 28px;
      height: 28px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      background: var(--border-secondary);
      color: var(--text-secondary);
      font-weight: 600;
      transition: all var(--transition-normal);
    }

    .step-text {
      font-size: 14px;
      font-weight: 500;
      white-space: nowrap;
    }
  }

  .step-arrow {
    color: var(--border-secondary);
  }
}

.stage-card {
  margin: 12px;

  &.stage-card-fullscreen {
    .stage-body-fullscreen {
      min-height: calc(100vh - 200px);
    }
  }
}

.stage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .header-left {
    display: flex;
    align-items: center;
    gap: 16px;

    .header-info {
      h2 {
        margin: 0 0 4px 0;
        font-size: 20px;
      }

      p {
        margin: 0;
        color: var(--text-muted);
        font-size: 14px;
      }
    }
  }
}

.stage-body {
  background: transparent;
  padding: var(--space-5);
  animation: fadeIn 0.5s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.action-buttons {
  display: flex;
  gap: 12px;
  margin: 12px 0;
  flex-wrap: wrap;
  justify-content: center;
  align-items: center;
}

.action-buttons-inline {
  display: flex;
  gap: 12px;
}

.script-textarea {
  margin: 16px 0;

  &.script-textarea-fullscreen {
    :deep(textarea) {
      min-height: 500px;
      font-size: 14px;
      line-height: 1.8;
    }
  }
}

.image-gen-section {
  margin-bottom: 40px;

  &.character-section {
    .section-header {
      border-color: var(--accent);
      background: var(--glass-bg);
    }

    .section-title h3 .el-icon {
      color: var(--accent);
    }
  }

  &.scene-section {
    .section-header {
      border-color: var(--success);
      background: var(--glass-bg);
    }

    .section-title h3 .el-icon {
      color: var(--success);
    }
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-5);
    padding: var(--space-4) var(--space-5);
    border-radius: var(--radius-xl);
    border: 1px solid var(--glass-border);
    background: var(--glass-bg);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    box-shadow: var(--shadow-sm);
    transition: all var(--transition-normal);

    &:hover {
      box-shadow: var(--shadow-md);
    }

    .section-title {
      display: flex;
      align-items: center;
      gap: 16px;

      h3 {
        display: flex;
        align-items: center;
        gap: 8px;
        margin: 0;
        font-size: 17px;
        font-weight: 700;
        color: var(--text-primary);
        letter-spacing: 0.3px;

        .el-icon {
          font-size: 20px;
          filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.1));
        }
      }

      .el-alert {
        border-radius: 6px;
        font-weight: 500;
      }
    }

    .section-actions {
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }
}

.empty-shots {
  padding: 60px 0;
  text-align: center;
}

.extracted-title {
  margin-bottom: 8px;
  color: var(--text-secondary);
}

.secondary-text {
  color: var(--text-muted);
  margin-left: 4px;
}

.task-message {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-muted);
  text-align: center;
}

/* 标准化改写进度 */
.standardize-progress {
  margin: 12px 0;
}

/* 内容质量检测面板 */
.quality-check-panel {
  padding: var(--space-4);
  background: var(--glass-bg);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-radius: var(--radius-lg);
  border: 1px solid var(--glass-border);
}

.quality-header {
  margin-bottom: 12px;
}

.quality-items {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.quality-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  line-height: 1.6;
}

.quality-icon {
  flex-shrink: 0;
  width: 18px;
  text-align: center;
}

.quality-label {
  flex-shrink: 0;
  font-weight: 500;
  color: var(--text-primary);
}

.quality-detail {
  color: var(--text-secondary);
}

.quality-action {
  margin-top: 12px;
}

.model-selector-bar {
  display: flex;
  align-items: center;
  gap: var(--space-6);
  padding: var(--space-3) var(--space-4);
  background: var(--glass-bg);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-radius: var(--radius-lg);
  border: 1px solid var(--glass-border);

  .model-selector-item {
    display: flex;
    align-items: center;
    gap: 8px;

    .model-label {
      font-size: 14px;
      font-weight: 500;
      color: var(--text-primary);
      white-space: nowrap;
    }
  }
}

.model-tip {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-muted);
}

.fixed-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  border-radius: var(--radius-xl);
  overflow: hidden;
  border: 1px solid var(--glass-border);
  box-shadow: var(--shadow-card);
  transition: all var(--transition-normal);
  background: var(--glass-bg-heavy);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));

  &.character-card {
    border-color: rgba(14, 165, 233, 0.3);
    
    &:hover {
      border-color: var(--accent);
      box-shadow: var(--shadow-card-hover), 0 0 20px rgba(14, 165, 233, 0.15);
    }

    .card-header {
      border-bottom-color: rgba(14, 165, 233, 0.15);
    }
  }

  &.scene-card {
    border-color: rgba(16, 185, 129, 0.3);
    
    &:hover {
      border-color: var(--success);
      box-shadow: var(--shadow-card-hover), 0 0 20px rgba(16, 185, 129, 0.15);
    }

    .card-header {
      border-bottom-color: rgba(16, 185, 129, 0.15);
    }
  }

  &:hover {
    box-shadow: var(--shadow-card-hover);
    transform: translateY(-3px);
    border-color: var(--accent);
  }

  :deep(.el-card__body) {
    flex: 1;
    padding: 0;
    display: flex;
    flex-direction: column;
  }

  .card-header {
    padding: var(--space-4);
    background: var(--glass-bg);
    border-bottom: 1px solid var(--glass-border);
    display: flex;
    justify-content: space-between;
    align-items: center;

    .header-left {
      flex: 1;
      min-width: 0;

      h4 {
        margin: 0 0 6px 0;
        font-size: 15px;
        font-weight: 700;
        color: var(--text-primary);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        letter-spacing: 0.2px;
      }

      .el-tag {
        margin-top: 0;
        font-weight: 500;
      }
    }
  }

  .card-image-container {
    flex: 1;
    width: 100%;
    min-height: 220px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-secondary);
    position: relative;
    overflow: hidden;

    .image-preview-wrapper {
      width: 100%;
      height: 100%;
    }

    .ref-badge {
      position: absolute;
      top: 6px;
      left: 6px;
      background: rgba(0, 0, 0, 0.55);
      color: #fff;
      font-size: 11px;
      padding: 2px 8px;
      border-radius: 4px;
      z-index: 2;
      backdrop-filter: blur(4px);
    }
  }

  .card-actions {
    padding: var(--space-3);
    background: var(--glass-bg);
    border-top: 1px solid var(--glass-border);
    display: flex;
    justify-content: center;
    gap: 10px;

    .el-button {
      margin: 0;
      transition: all 0.2s ease;

      &:hover {
        transform: scale(1.1);
      }
    }
  }
}

.character-image-list,
.scene-image-list,
.prop-image-list {
  padding: 8px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
  margin-top: 20px;

  .character-item,
  .scene-item,
  .prop-item {
    min-height: 380px;
    animation: cardFadeIn 0.5s ease;
  }
}

@keyframes cardFadeIn {
  from {
    opacity: 0;
    transform: translateY(20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

// 角色库选择对话框
.library-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  max-height: 500px;
  overflow-y: auto;
  padding: 8px;

  .library-item {
    cursor: pointer;
    border: 2px solid transparent;
    border-radius: var(--radius-lg);
    overflow: hidden;
    transition: all var(--transition-normal);

    &:hover {
      border-color: var(--accent);
      transform: translateY(-2px);
      box-shadow: var(--shadow-lg);
    }

    .el-image {
      width: 100%;
      height: 150px;
    }

    .library-item-name {
      padding: 8px;
      text-align: center;
      font-size: 12px;
      background: var(--bg-secondary);
      color: var(--text-primary);
    }
  }
}

.empty-library {
  padding: 40px 0;
}

// 上传区域
.upload-area {
  :deep(.el-upload-dragger) {
    width: 100%;
    height: 200px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
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

/* ========================================
   Dark Mode / 深色模式
   ======================================== */
:deep(.el-card) {
  background: var(--glass-bg-heavy);
  border-color: var(--glass-border);
}

:deep(.el-card__header) {
  background: var(--glass-bg);
  border-color: var(--glass-border);
}

:deep(.el-table) {
  --el-table-bg-color: var(--bg-card);
  --el-table-header-bg-color: var(--bg-secondary);
  --el-table-tr-bg-color: var(--bg-card);
  --el-table-row-hover-bg-color: var(--bg-card-hover);
  --el-table-border-color: var(--border-primary);
  --el-table-text-color: var(--text-primary);
  background: var(--bg-card);
}

:deep(.el-table th.el-table__cell),
:deep(.el-table td.el-table__cell) {
  background: var(--bg-card);
  border-color: var(--border-primary);
}

:deep(
  .el-table--striped .el-table__body tr.el-table__row--striped td.el-table__cell
) {
  background: var(--bg-secondary);
}

:deep(.el-table__header-wrapper th) {
  background: var(--bg-secondary) !important;
  color: var(--text-secondary);
}

:deep(.el-dialog) {
  background: var(--glass-bg-heavy);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border: 1px solid var(--glass-border);
}

:deep(.el-dialog__header) {
  background: transparent;
}

:deep(.el-form-item__label) {
  color: var(--text-primary);
}

:deep(.el-input__wrapper) {
  background: var(--bg-secondary);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
}

:deep(.el-input__inner) {
  color: var(--text-primary);
}

:deep(.el-textarea__inner) {
  background: var(--bg-secondary);
  color: var(--text-primary);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
}

:deep(.el-select-dropdown) {
  background: var(--bg-elevated);
  border-color: var(--border-primary);
}

:deep(.el-upload-dragger) {
  background: var(--bg-secondary);
  border-color: var(--border-primary);
}

.character-info-preview {
  padding: 8px 12px;
  margin: 8px 0;
  background: var(--bg-secondary);
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.4;
}


.script-preview-box {
  background: var(--glass-bg);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  padding: var(--space-5) var(--space-6);
  max-height: 600px;
  overflow-y: auto;
}

.script-design-section h3 {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0;
}

.script-preview-section h3 {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0;
}

.script-resource-collapse {
  margin-bottom: 16px;
  border: none;

  :deep(.el-collapse-item__header) {
    background: transparent;
    border: none;
    height: 36px;
    line-height: 36px;
    font-size: 14px;
  }

  :deep(.el-collapse-item__wrap) {
    border: none;
    background: transparent;
  }

  :deep(.el-collapse-item__content) {
    padding-bottom: 8px;
  }
}

.resource-collapse-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 500;
  color: var(--text-secondary);
  font-size: 13px;
}

.resource-tag-groups {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.resource-tag-group {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;

  .resource-tag-label {
    font-size: 12px;
    color: var(--text-muted);
    width: 36px;
    flex-shrink: 0;
  }
}

.script-input-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.script-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  padding: 10px 14px;
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
}

.script-toolbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.script-toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.script-preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 16px 0 12px;
}

.script-preview-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.script-stats {
  display: flex;
  gap: 6px;
}

.script-preview-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>

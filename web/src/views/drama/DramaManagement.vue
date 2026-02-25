<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <!-- Page Header / 页面头部 -->
      <AppHeader :fixed="false" :show-logo="false">
        <template #left>
          <el-button text @click="$router.push('/')" class="back-btn">
            <el-icon><ArrowLeft /></el-icon>
            <span>{{ $t("common.back") }}</span>
          </el-button>
          <div class="page-title">
            <h1>{{ drama?.title || "" }}</h1>
            <span class="subtitle">{{
              drama?.description || $t("drama.management.overview")
            }}</span>
          </div>
        </template>
      </AppHeader>

      <!-- Tabs / 标签页 -->
      <div class="tabs-wrapper">
        <el-tabs v-model="activeTab" class="management-tabs">
          <!-- 项目概览 -->
          <el-tab-pane :label="$t('drama.management.overview')" name="overview">
            <div class="stats-grid">
              <StatCard
                :label="$t('drama.management.episodeStats')"
                :value="episodesCount"
                :icon="Document"
                icon-color="var(--accent)"
                icon-bg="var(--accent-light)"
                value-color="var(--accent)"
                :description="$t('drama.management.episodesCreated')"
              />
              <StatCard
                :label="$t('drama.management.characterStats')"
                :value="charactersCount"
                :icon="User"
                icon-color="var(--success)"
                icon-bg="var(--success-light)"
                value-color="var(--success)"
                :description="$t('drama.management.charactersCreated')"
              />
              <StatCard
                :label="$t('drama.management.sceneStats')"
                :value="scenesCount"
                :icon="Picture"
                icon-color="var(--warning)"
                icon-bg="var(--warning-light)"
                value-color="var(--warning)"
                :description="$t('drama.management.sceneLibraryCount')"
              />
              <StatCard
                :label="$t('drama.management.propStats')"
                :value="propsCount"
                :icon="Box"
                icon-color="var(--primary)"
                icon-bg="var(--primary-light)"
                value-color="var(--primary)"
                :description="$t('drama.management.propsCreated')"
              />
            </div>

            <!-- 引导卡片：无章节时显示 -->
            <el-alert
              v-if="episodesCount === 0"
              :title="$t('drama.management.startFirstEpisode')"
              type="info"
              :closable="false"
              style="margin-top: 20px"
            >
              <template #default>
                <p style="margin: 8px 0">
                  {{ $t("drama.management.noEpisodesYet") }}
                </p>
                <el-button
                  type="primary"
                  :icon="Plus"
                  @click="createNewEpisode"
                  style="margin-top: 8px"
                >
                  {{ $t("drama.management.createFirstEpisode") }}
                </el-button>
              </template>
            </el-alert>

            <el-card shadow="never" class="project-info-card">
              <template #header>
                <div class="card-header">
                  <h3 class="card-title">
                    {{ $t("drama.management.projectInfo") }}
                  </h3>
                  <el-tag :type="getStatusType(drama?.status)" size="small">{{
                    getStatusText(drama?.status)
                  }}</el-tag>
                </div>
              </template>
              <el-descriptions :column="2" border class="project-descriptions">
                <el-descriptions-item
                  :label="$t('drama.management.projectName')"
                >
                  <span class="info-value">{{ drama?.title }}</span>
                </el-descriptions-item>
                <el-descriptions-item :label="$t('common.createdAt')">
                  <span class="info-value">{{
                    formatDate(drama?.created_at)
                  }}</span>
                </el-descriptions-item>
                <el-descriptions-item
                  :label="$t('drama.management.projectDesc')"
                  :span="2"
                >
                  <span class="info-desc">{{
                    drama?.description || $t("drama.management.noDescription")
                  }}</span>
                </el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-tab-pane>

          <!-- 章节管理 -->
          <el-tab-pane :label="$t('drama.management.episodes')" name="episodes">
            <div class="tab-header">
              <h2>{{ $t("drama.management.episodeList") }}</h2>
              <el-button
                type="primary"
                :icon="Plus"
                @click="createNewEpisode"
                >{{ $t("drama.management.createNewEpisode") }}</el-button
              >
            </div>

            <!-- 空状态引导 -->
            <el-empty
              v-if="episodesCount === 0"
              :description="$t('drama.management.noEpisodes')"
              style="margin-top: 40px"
            >
              <template #image>
                <el-icon :size="80" class="empty-icon"><Document /></el-icon>
              </template>
              <el-button type="primary" :icon="Plus" @click="createNewEpisode">
                {{ $t("drama.management.createFirstEpisode") }}
              </el-button>
            </el-empty>

            <template v-else>
              <!-- 人物 PV -->
              <div v-if="pvEpisodes.length > 0" class="episode-section">
                <h3 class="episode-section-title">人物 PV（角色介绍短片）</h3>
                <el-table :data="pvEpisodes" border stripe>
                  <el-table-column prop="episode_number" label="编号" width="80" />
                  <el-table-column prop="title" :label="$t('drama.management.episodeName')" min-width="200" />
                  <el-table-column prop="description" label="简介" min-width="280" show-overflow-tooltip />
                  <el-table-column :label="$t('common.status')" width="120">
                    <template #default="{ row }">
                      <el-tag :type="getEpisodeStatusType(row)">{{ getEpisodeStatusText(row) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="Shots" width="100">
                    <template #default="{ row }">{{ row.storyboards?.length || 0 }}</template>
                  </el-table-column>
                  <el-table-column :label="$t('storyboard.table.operations')" width="280" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" type="info" @click="editEpisodeName(row)">{{ $t("common.edit") }}</el-button>
                      <el-button size="small" type="primary" @click="enterEpisodeWorkflow(row)">{{ $t("drama.management.goToEdit") }}</el-button>
                      <el-button size="small" type="danger" @click="deleteEpisode(row)">{{ $t("common.delete") }}</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </div>

              <!-- 正式剧集 -->
              <div v-if="mainEpisodes.length > 0" class="episode-section">
                <h3 class="episode-section-title">正式剧集</h3>
                <el-table :data="mainEpisodes" border stripe>
                  <el-table-column prop="episode_number" label="编号" width="80" />
                  <el-table-column prop="title" :label="$t('drama.management.episodeName')" min-width="200" />
                  <el-table-column prop="description" label="简介" min-width="280" show-overflow-tooltip />
                  <el-table-column :label="$t('common.status')" width="120">
                    <template #default="{ row }">
                      <el-tag :type="getEpisodeStatusType(row)">{{ getEpisodeStatusText(row) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="Shots" width="100">
                    <template #default="{ row }">{{ row.storyboards?.length || 0 }}</template>
                  </el-table-column>
                  <el-table-column :label="$t('common.createdAt')" width="180">
                    <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
                  </el-table-column>
                  <el-table-column :label="$t('storyboard.table.operations')" width="280" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" type="info" @click="editEpisodeName(row)">{{ $t("common.edit") }}</el-button>
                      <el-button size="small" type="primary" @click="enterEpisodeWorkflow(row)">{{ $t("drama.management.goToEdit") }}</el-button>
                      <el-button size="small" type="danger" @click="deleteEpisode(row)">{{ $t("common.delete") }}</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </template>
          </el-tab-pane>

          <!-- 角色管理 -->
          <el-tab-pane
            :label="$t('drama.management.characters')"
            name="characters"
          >
            <div class="tab-header">
              <h2>{{ $t("drama.management.characterList") }}</h2>
              <div style="display: flex; gap: 10px; align-items: center;">
                <el-input
                  v-model="characterFilter"
                  placeholder="搜索角色名称..."
                  :prefix-icon="Search"
                  clearable
                  style="width: 200px"
                  size="default"
                />
                <el-select v-model="characterRoleFilter" placeholder="角色类型" clearable style="width: 120px" size="default">
                  <el-option label="主角" value="main" />
                  <el-option label="配角" value="supporting" />
                  <el-option label="群演" value="minor" />
                </el-select>
                <el-button
                  :icon="Document"
                  @click="openExtractCharacterDialog"
                  disabled
                  >自动导入（待开发）</el-button
                >
                <el-button
                  type="primary"
                  :icon="Plus"
                  @click="openAddCharacterDialog"
                  >添加角色</el-button
                >
                <el-button
                  :icon="MagicStick"
                  @click="openAiGenerateCharacterDialog"
                  >AI生成</el-button
                >
              </div>
            </div>

            <div class="char-list" style="margin-top: 12px">
              <div
                v-for="character in filteredCharacters"
                :key="character.id"
                class="char-card"
              >
                <!-- Top bar: name + tags + actions -->
                <div class="char-topbar">
                  <div class="char-name-row">
                    <h4>{{ character.name }}</h4>
                    <el-tag :type="character.role === 'main' ? 'danger' : 'info'" size="small">
                      {{ character.role === "main" ? "主角" : character.role === "supporting" ? "配角" : "群演" }}
                    </el-tag>
                    <el-tag v-if="character.description" type="success" size="small">{{ character.description }}</el-tag>
                    <el-tag v-if="character.children?.length" type="warning" size="small">{{ character.children.length }}套造型</el-tag>
                  </div>
                  <div class="char-actions">
                    <el-button size="small" link @click="generateBaseImage(character)">
                      {{ character.image_url ? '重新生成' : '生成基础形象' }}
                    </el-button>
                    <el-button size="small" link @click="editCharacter(character)">编辑</el-button>
                    <el-button size="small" link @click="openAddOutfitDialog(character)">添加造型</el-button>
                    <el-button size="small" link type="danger" @click="deleteCharacter(character)">删除</el-button>
                  </div>
                </div>
                <!-- Image gallery: base image + outfit images side by side -->
                <div class="char-gallery">
                  <div class="char-gallery-item char-gallery-base" @click="character.image_url ? undefined : generateBaseImage(character)">
                    <ImagePreview
                      v-if="character.local_path || character.image_url"
                      :image-url="getImageUrl(character)"
                      :alt="character.name + ' 基础形象'"
                      :size="140"
                      dialog-width="900px"
                      detail-layout="bottom"
                    >
                      <template #details>
                        <CollapsibleText v-if="character.appearance" label="外貌" :text="character.appearance" />
                      </template>
                    </ImagePreview>
                    <div v-else class="char-base-empty">
                      <el-icon :size="24"><Plus /></el-icon>
                      <span>基础形象</span>
                      <span v-if="character.reference_images?.length" class="ref-count-hint">{{ character.reference_images.length }}张参考</span>
                    </div>
                    <span class="char-gallery-label">基础形象</span>
                  </div>
                  <div
                    v-for="outfit in (character.children || [])"
                    :key="outfit.id"
                    class="char-gallery-item"
                  >
                    <ImagePreview
                      v-if="outfit.local_path || outfit.image_url"
                      :image-url="getImageUrl(outfit)"
                      :alt="outfit.outfit_name || outfit.name"
                      :size="140"
                      dialog-width="900px"
                      detail-layout="bottom"
                    >
                      <template #details>
                        <CollapsibleText v-if="outfit.appearance" label="外貌" :text="outfit.appearance" />
                      </template>
                    </ImagePreview>
                    <el-avatar v-else :size="48" shape="square">{{ (outfit.outfit_name || outfit.name || '?')[0] }}</el-avatar>
                    <span class="char-gallery-label">{{ outfit.outfit_name || outfit.name }}</span>
                    <div class="char-gallery-actions">
                      <el-button size="small" link @click.stop="editCharacter(outfit)">编辑</el-button>
                      <el-button
                        size="small" link
                        :type="character.image_url ? '' : 'warning'"
                        @click.stop="generateOutfitImage(outfit, character)"
                      >生成</el-button>
                      <el-button size="small" link type="danger" @click.stop="deleteCharacter(outfit)">删除</el-button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <el-empty
              v-if="filteredCharacters.length === 0 && (drama?.characters?.length || 0) > 0"
              description="没有匹配的角色"
            />
            <el-empty
              v-else-if="!drama?.characters || drama.characters.length === 0"
              :description="$t('drama.management.noCharacters')"
            />
          </el-tab-pane>

          <!-- 场景库管理 -->
          <el-tab-pane :label="$t('drama.management.sceneList')" name="scenes">
            <div class="tab-header">
              <h2>{{ $t("drama.management.sceneList") }}</h2>
              <div style="display: flex; gap: 10px; align-items: center;">
                <el-input
                  v-model="sceneFilter"
                  placeholder="搜索场景..."
                  :prefix-icon="Search"
                  clearable
                  style="width: 200px"
                  size="default"
                />
                <el-button
                  :icon="Document"
                  @click="openExtractSceneDialog"
                  disabled
                  >自动导入（待开发）</el-button
                >
                <el-button
                  type="primary"
                  :icon="Plus"
                  @click="openAddSceneDialog"
                  >添加场景</el-button
                >
                <el-button
                  :icon="MagicStick"
                  @click="openAiGenerateSceneDialog"
                  >AI生成</el-button
                >
                <el-button
                  v-if="batchSelecting !== 'scene'"
                  type="warning"
                  :icon="MagicStick"
                  @click="batchSelecting = 'scene'; selectedSceneIds = [];"
                >批量生成图片</el-button>
                <template v-else>
                  <el-button
                    type="success"
                    :icon="MagicStick"
                    :loading="batchGenerating"
                    :disabled="selectedSceneIds.length === 0"
                    @click="batchGenerateForType('scene')"
                  >确认生成（{{ selectedSceneIds.length }}）</el-button>
                  <el-button @click="batchSelecting = null; selectedSceneIds = [];">取消</el-button>
                </template>
              </div>
            </div>

            <div class="scene-list" style="margin-top: 16px">
              <div v-for="scene in filteredScenes" :key="scene.id" class="scene-list-item">
                <el-checkbox
                  v-if="batchSelecting === 'scene'"
                  :model-value="selectedSceneIds.includes(scene.id)"
                  @change="(val: boolean) => { if (val) selectedSceneIds.push(scene.id); else selectedSceneIds.splice(selectedSceneIds.indexOf(scene.id), 1); }"
                  style="margin-right: 8px;"
                />
                <div class="scene-list-thumb">
                  <ImagePreview
                    v-if="scene.local_path || scene.image_url"
                    :image-url="getImageUrl(scene)"
                    :alt="scene.name || scene.location || '场景'"
                    :size="80"
                    dialog-width="900px"
                  >
                    <template #details>
                      <div v-if="scene.location" class="detail-section">
                        <div class="detail-label">位置</div>
                        <div class="detail-value">{{ scene.location }}</div>
                      </div>
                      <CollapsibleText v-if="scene.description" label="描述" :text="scene.description" />
                    </template>
                  </ImagePreview>
                  <el-avatar v-else :size="60" shape="square">
                    <el-icon :size="28"><Picture /></el-icon>
                  </el-avatar>
                </div>
                <div class="scene-list-info">
                  <div class="scene-list-header">
                    <h4>{{ scene.name || scene.location }}</h4>
                  </div>
                  <p class="scene-list-desc">{{ scene.description || scene.prompt }}</p>
                </div>
                <div class="scene-list-actions">
                  <el-button size="small" link @click="editScene(scene)">编辑</el-button>
                  <el-button size="small" link @click="generateSceneImage(scene)">生成图片</el-button>
                  <el-button size="small" link type="danger" @click="deleteScene(scene)">删除</el-button>
                </div>
              </div>
              <el-empty v-if="filteredScenes.length === 0" description="暂无场景" />
            </div>

          </el-tab-pane>

          <!-- 道具管理 -->
          <el-tab-pane :label="$t('drama.management.propList')" name="props">
            <div class="tab-header">
              <h2>{{ $t("drama.management.propList") }}</h2>
              <div style="display: flex; gap: 10px; align-items: center;">
                <el-input
                  v-model="propFilter"
                  placeholder="搜索道具..."
                  :prefix-icon="Search"
                  clearable
                  style="width: 200px"
                  size="default"
                />
                <el-button
                  :icon="Document"
                  @click="openExtractDialog"
                  disabled
                  >自动导入（待开发）</el-button
                >
                <el-button
                  type="primary"
                  :icon="Plus"
                  @click="openAddPropDialog"
                  >添加道具</el-button
                >
                <el-button
                  :icon="MagicStick"
                  @click="openAiGeneratePropDialog"
                  >AI生成</el-button
                >
                <el-button
                  v-if="batchSelecting !== 'prop'"
                  type="warning"
                  :icon="MagicStick"
                  @click="batchSelecting = 'prop'; selectedPropIds = [];"
                >批量生成图片</el-button>
                <template v-else>
                  <el-button
                    type="success"
                    :icon="MagicStick"
                    :loading="batchGenerating"
                    :disabled="selectedPropIds.length === 0"
                    @click="batchGenerateForType('prop')"
                  >确认生成（{{ selectedPropIds.length }}）</el-button>
                  <el-button @click="batchSelecting = null; selectedPropIds = [];">取消</el-button>
                </template>
              </div>
            </div>

            <div class="prop-list" style="margin-top: 16px">
              <div v-for="prop in filteredProps" :key="prop.id" class="prop-list-item">
                <el-checkbox
                  v-if="batchSelecting === 'prop'"
                  :model-value="selectedPropIds.includes(prop.id)"
                  @change="(val: boolean) => { if (val) selectedPropIds.push(prop.id); else selectedPropIds.splice(selectedPropIds.indexOf(prop.id), 1); }"
                  style="margin-right: 8px;"
                />
                <div class="prop-list-thumb">
                  <ImagePreview
                    v-if="prop.local_path || prop.image_url"
                    :image-url="getImageUrl(prop)"
                    :alt="prop.name || '道具'"
                    :size="80"
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
                  <el-avatar v-else :size="60" shape="square">
                    <el-icon :size="28"><Box /></el-icon>
                  </el-avatar>
                </div>
                <div class="prop-list-info">
                  <div class="prop-list-header">
                    <h4>{{ prop.name }}</h4>
                    <el-tag size="small" v-if="prop.type" type="info">{{ prop.type }}</el-tag>
                  </div>
                  <p class="prop-list-desc">{{ prop.description || prop.prompt }}</p>
                </div>
                <div class="prop-list-actions">
                  <el-button size="small" link @click="editProp(prop)">编辑</el-button>
                  <el-button size="small" link @click="generatePropImage(prop)">生成图片</el-button>
                  <el-button size="small" link type="danger" @click="deleteProp(prop)">删除</el-button>
                </div>
              </div>
              <el-empty v-if="filteredProps.length === 0" description="暂无道具" />
            </div>

            <el-empty
              v-if="filteredProps.length === 0 && (drama?.props?.length || 0) > 0"
              description="没有匹配的道具"
            />
            <el-empty
              v-else-if="!drama?.props || drama.props.length === 0"
              :description="$t('drama.management.noProps')"
            />
          </el-tab-pane>
        </el-tabs>
      </div>

      <!-- 添加/编辑角色对话框 -->
      <el-dialog
        v-model="addCharacterDialogVisible"
        :title="editingCharacter ? $t('character.edit') : (newCharacter.parent_id ? '添加造型' : $t('character.add'))"
        width="700px"
      >
        <el-form :model="newCharacter" label-width="100px">
          <el-form-item :label="$t('character.image')">
            <el-upload
              class="avatar-uploader"
              :action="`/api/v1/upload/image`"
              :headers="uploadHeaders"
              :show-file-list="false"
              :on-success="handleCharacterAvatarSuccess"
              :before-upload="beforeAvatarUpload"
            >
              <div
                v-if="hasImage(newCharacter)"
                class="avatar-wrapper"
                style="
                  width: 200px;
                  height: 120px;
                  position: relative;
                  overflow: hidden;
                  border-radius: 6px;
                "
              >
                <img
                  :src="getImageUrl(newCharacter)"
                  class="avatar"
                  style="width: 100%; height: 100%; object-fit: contain; background: #f5f5f5;"
                />
                <div
                  class="avatar-overlay"
                  style="
                    position: absolute;
                    top: 0;
                    left: 0;
                    right: 0;
                    bottom: 0;
                    background: rgba(0, 0, 0, 0.5);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    opacity: 0;
                    transition: opacity 0.3s;
                    cursor: pointer;
                  "
                >
                  <el-icon style="color: white; font-size: 24px;">
                    <Plus />
                  </el-icon>
                </div>
              </div>
              <div
                v-else
                class="avatar-uploader-icon"
                style="
                  border: 1px dashed #d9d9d9;
                  border-radius: 6px;
                  cursor: pointer;
                  position: relative;
                  overflow: hidden;
                  width: 200px;
                  height: 120px;
                  font-size: 28px;
                  color: #8c939d;
                  text-align: center;
                  line-height: 120px;
                "
              >
                <el-icon><Plus /></el-icon>
              </div>
            </el-upload>
          </el-form-item>
          <el-form-item :label="$t('character.name')">
            <el-input
              v-model="newCharacter.name"
              :placeholder="$t('character.name')"
            />
          </el-form-item>
          <el-form-item v-if="newCharacter.parent_id" label="造型名称">
            <el-input
              v-model="newCharacter.outfit_name"
              placeholder="如：居家装、通勤装、运动装"
            />
          </el-form-item>
          <el-form-item :label="$t('character.role')">
            <el-select
              v-model="newCharacter.role"
              :placeholder="$t('common.pleaseSelect')"
            >
              <el-option label="Main" value="main" />
              <el-option label="Supporting" value="supporting" />
              <el-option label="Minor" value="minor" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('character.appearance')">
            <el-input
              v-model="newCharacter.appearance"
              type="textarea"
              :rows="3"
              :placeholder="$t('character.appearance')"
            />
          </el-form-item>
          <el-form-item :label="$t('character.personality')">
            <el-input
              v-model="newCharacter.personality"
              type="textarea"
              :rows="3"
              :placeholder="$t('character.personality')"
            />
          </el-form-item>
          <el-form-item :label="$t('character.description')">
            <el-input
              v-model="newCharacter.description"
              type="textarea"
              :rows="3"
              :placeholder="$t('common.description')"
            />
          </el-form-item>
          <el-form-item label="参考图片">
            <el-upload
              v-model:file-list="characterReferenceImages"
              :action="`/api/v1/upload/image`"
              :headers="uploadHeaders"
              list-type="picture-card"
              :on-success="handleCharacterReferenceImageSuccess"
              :before-upload="beforeAvatarUpload"
              :limit="5"
              multiple
            >
              <el-icon><Plus /></el-icon>
              <template #file="{ file }">
                <div class="upload-file-card">
                  <img :src="file.url" class="el-upload-list__item-thumbnail" />
                  <span class="el-upload-list__item-actions">
                    <span class="el-upload-list__item-preview" @click="handlePictureCardPreview(file)">
                      <el-icon><ZoomIn /></el-icon>
                    </span>
                    <span class="el-upload-list__item-download" @click="downloadRefImage(file.url, file.name)">
                      <el-icon><Download /></el-icon>
                    </span>
                    <span class="el-upload-list__item-delete" @click="handleCharRefRemove(file)">
                      <el-icon><Delete /></el-icon>
                    </span>
                  </span>
                </div>
              </template>
            </el-upload>
            <div style="color: #999; font-size: 12px; margin-top: 5px;">
              可以上传最多5张参考图片
            </div>
          </el-form-item>
          <el-form-item label="生成尺寸">
            <el-tag type="info" effect="plain">4:3 横向 · 2304×1728</el-tag>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="addCharacterDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button type="primary" @click="saveCharacter">{{
            $t("common.confirm")
          }}</el-button>
        </template>
      </el-dialog>

      <!-- AI生成道具对话框 -->
      <el-dialog
        v-model="aiGeneratePropDialogVisible"
        title="AI生成道具信息"
        width="700px"
      >
        <div style="margin-bottom: 16px;">
          <h4>输入原始信息</h4>
          <el-input
            v-model="aiPropInput"
            :rows="6"
            type="textarea"
            placeholder="请输入道具相关信息，例如：道具名称、类型、用途、描述等..."
          />
        </div>
        
        <div style="margin-bottom: 16px;">
          <h4>AI处理结果</h4>
          <el-form label-width="100px">
            <el-form-item label="道具名称">
              <el-input
                v-model="aiPropName"
                placeholder="道具名称"
                readonly
              />
            </el-form-item>
            <el-form-item label="道具类型">
              <el-input
                v-model="aiPropType"
                placeholder="道具类型"
                readonly
              />
            </el-form-item>
            <el-form-item label="道具描述">
              <el-input
                v-model="aiPropDescription"
                type="textarea"
                :rows="6"
                placeholder="道具描述"
                readonly
              />
            </el-form-item>
            <el-form-item label="图片提示词">
              <el-input
                v-model="aiPropPrompt"
                type="textarea"
                :rows="4"
                placeholder="图片提示词"
                readonly
              />
            </el-form-item>
          </el-form>
        </div>
        
        <template #footer>
          <el-button @click="aiGeneratePropDialogVisible = false">
            {{ $t("common.cancel") }}
          </el-button>
          <el-button @click="aiProcessProps" type="primary" plain>
            AI处理
          </el-button>
          <el-button @click="saveAiGeneratedProps" type="primary">
            导入到表单
          </el-button>
        </template>
      </el-dialog>

      <!-- AI生成角色对话框 -->
      <el-dialog
        v-model="aiGenerateCharacterDialogVisible"
        title="AI生成角色信息"
        width="700px"
      >
        <div style="margin-bottom: 16px;">
          <h4>输入原始信息</h4>
          <el-input
            v-model="aiCharacterInput"
            :rows="6"
            type="textarea"
            placeholder="请输入角色相关信息，例如：角色名称、外貌、性格、背景等..."
          />
        </div>
        
        <div style="margin-bottom: 16px;">
          <h4>AI处理结果</h4>
          <el-form label-width="100px">
            <el-form-item label="角色名称">
              <el-input
                v-model="aiCharacterName"
                placeholder="角色名称"
                readonly
              />
            </el-form-item>
            <el-form-item label="角色身份">
              <el-input
                v-model="aiCharacterRole"
                placeholder="角色身份"
                readonly
              />
            </el-form-item>
            <el-form-item label="外貌特征">
              <el-input
                v-model="aiCharacterAppearance"
                type="textarea"
                :rows="3"
                placeholder="外貌特征"
                readonly
              />
            </el-form-item>
            <el-form-item label="性格特点">
              <el-input
                v-model="aiCharacterPersonality"
                type="textarea"
                :rows="3"
                placeholder="性格特点"
                readonly
              />
            </el-form-item>
            <el-form-item label="声音特色">
              <el-input
                v-model="aiCharacterVoiceStyle"
                placeholder="声音特色"
                readonly
              />
            </el-form-item>
            <el-form-item label="背景故事">
              <el-input
                v-model="aiCharacterBackground"
                type="textarea"
                :rows="3"
                placeholder="背景故事"
                readonly
              />
            </el-form-item>
          </el-form>
        </div>
        
        <template #footer>
          <el-button @click="aiGenerateCharacterDialogVisible = false">
            {{ $t("common.cancel") }}
          </el-button>
          <el-button @click="aiProcessCharacters" type="primary" plain>
            AI处理
          </el-button>
          <el-button @click="saveAiGeneratedCharacters" type="primary">
            导入到表单
          </el-button>
        </template>
      </el-dialog>

      <!-- AI生成场景对话框 -->
      <el-dialog
        v-model="aiGenerateSceneDialogVisible"
        title="AI生成场景信息"
        width="700px"
      >
        <div style="margin-bottom: 16px;">
          <h4>输入原始信息</h4>
          <el-input
            v-model="aiSceneInput"
            :rows="6"
            type="textarea"
            placeholder="请输入场景相关信息，例如：地点、时间、环境描述、氛围等..."
          />
        </div>
        
        <div style="margin-bottom: 16px;">
          <h4>AI处理结果</h4>
          <el-form label-width="100px">
            <el-form-item label="场景地点">
              <el-input
                v-model="aiSceneLocation"
                placeholder="场景地点"
                readonly
              />
            </el-form-item>
            <el-form-item label="时间">
              <el-input
                v-model="aiSceneTime"
                placeholder="时间"
                readonly
              />
            </el-form-item>
            <el-form-item label="场景描述">
              <el-input
                v-model="aiSceneDescription"
                type="textarea"
                :rows="3"
                placeholder="场景描述"
                readonly
              />
            </el-form-item>
            <el-form-item label="氛围">
              <el-input
                v-model="aiSceneAtmosphere"
                type="textarea"
                :rows="2"
                placeholder="氛围"
                readonly
              />
            </el-form-item>
            <el-form-item label="光线效果">
              <el-input
                v-model="aiSceneLighting"
                placeholder="光线效果"
                readonly
              />
            </el-form-item>
            <el-form-item label="声音效果">
              <el-input
                v-model="aiSceneSound"
                placeholder="声音效果"
                readonly
              />
            </el-form-item>
          </el-form>
        </div>
        
        <template #footer>
          <el-button @click="aiGenerateSceneDialogVisible = false">
            {{ $t("common.cancel") }}
          </el-button>
          <el-button @click="aiProcessScenes" type="primary" plain>
            AI处理
          </el-button>
          <el-button @click="saveAiGeneratedScenes" type="primary">
            导入到表单
          </el-button>
        </template>
      </el-dialog>

      <!-- 图片生成对话框 -->
      <el-dialog
        v-model="generateImageDialogVisible"
        title="生成图片"
        width="600px"
      >
        <el-form label-width="100px">
          <el-form-item label="生成提示词">
            <el-input
              v-model="imagePrompt"
              type="textarea"
              :rows="6"
              placeholder="用于生成图片的提示词"
            />
            <div style="display: flex; justify-content: space-between; align-items: center; margin-top: 5px;">
              <div style="color: #999; font-size: 12px;">
                可以编辑提示词来优化生成效果
              </div>
              <el-button 
                type="primary" 
                plain 
                size="small" 
                @click="aiGeneratePrompt"
                :loading="generatingPrompt"
              >
                AI智能生成提示词
              </el-button>
            </div>
          </el-form-item>
          
          <el-form-item v-if="baseReferenceImages.length > 0" label="基础形象">
            <div class="base-ref-preview">
              <div
                v-for="(file, idx) in baseReferenceImages"
                :key="idx"
                class="base-ref-item"
              >
                <el-image
                  :src="file.url"
                  fit="cover"
                  :preview-src-list="baseReferenceImages.map(f => f.url)"
                  :initial-index="idx"
                />
                <el-tag size="small" type="success" class="base-ref-tag">基础形象</el-tag>
                <el-button
                  class="ref-download-btn"
                  :icon="Download"
                  circle
                  size="small"
                  @click.stop="downloadRefImage(file.url, file.name)"
                />
              </div>
            </div>
            <div style="color: #999; font-size: 12px; margin-top: 4px;">
              生成造型时将自动使用基础形象作为参考
            </div>
          </el-form-item>

          <el-form-item :label="baseReferenceImages.length > 0 ? '装扮参考' : '参考图片'">
            <el-upload
              v-model:file-list="referenceImageList"
              :action="`/api/v1/upload/image`"
              :headers="uploadHeaders"
              list-type="picture-card"
              :on-success="handleReferenceImageSuccess"
              :before-upload="beforeAvatarUpload"
              :limit="5"
              multiple
            >
              <el-icon><Plus /></el-icon>
              <template #file="{ file }">
                <div class="upload-file-card">
                  <img :src="file.url" class="el-upload-list__item-thumbnail" />
                  <span class="el-upload-list__item-actions">
                    <span class="el-upload-list__item-preview" @click="handlePictureCardPreview(file)">
                      <el-icon><ZoomIn /></el-icon>
                    </span>
                    <span class="el-upload-list__item-download" @click="downloadRefImage(file.url, file.name)">
                      <el-icon><Download /></el-icon>
                    </span>
                    <span class="el-upload-list__item-delete" @click="handleUploadRemove(file)">
                      <el-icon><Delete /></el-icon>
                    </span>
                  </span>
                </div>
              </template>
            </el-upload>
            <div style="color: #999; font-size: 12px; margin-top: 5px;">
              可以上传最多5张{{ baseReferenceImages.length > 0 ? '装扮' : '' }}参考图片
            </div>
          </el-form-item>
          <el-form-item label="生成尺寸">
            <el-tag v-if="currentGenerateTarget?.type === 'scene'" type="info" effect="plain">21:9 超宽银幕 · 3360×1440</el-tag>
            <el-tag v-else-if="currentGenerateTarget?.type === 'prop'" type="info" effect="plain">4:3 横向 · 2304×1728（三视图）</el-tag>
            <el-tag v-else type="info" effect="plain">4:3 横向 · 2304×1728（三视图）</el-tag>
          </el-form-item>
        </el-form>
        
        <template #footer>
          <el-button @click="generateImageDialogVisible = false">
            取消
          </el-button>
          <el-button @click="saveImageConfig" type="success" plain>
            保存配置
          </el-button>
          <el-button @click="debugGenerateImage" title="仅展示请求参数，不调用 API">Debug</el-button>
          <el-button @click="confirmGenerateImage" type="primary" :loading="generatingImage">
            开始生成
          </el-button>
        </template>
      </el-dialog>

      <!-- 参考图片预览弹窗 -->
      <el-dialog v-model="previewDialogVisible" title="图片预览" width="auto" append-to-body>
        <img :src="previewImageUrl" style="max-width: 100%; max-height: 80vh; display: block; margin: 0 auto;" />
      </el-dialog>

      <!-- Debug 信息弹窗 -->
      <el-dialog
        v-model="debugDialogVisible"
        title="Debug - 图片生成请求参数"
        width="700px"
      >
        <div style="margin-bottom: 12px;">
          <p style="font-size: 13px; font-weight: 600; color: #409eff; margin-bottom: 8px;">📋 发送到后端的 API 请求：</p>
          <pre style="background: #1e1e1e; color: #d4d4d4; padding: 12px; border-radius: 8px; font-size: 12px; line-height: 1.5; white-space: pre-wrap; word-break: break-all; max-height: 500px; overflow-y: auto;">{{ debugCurlCommand }}</pre>
        </div>
        <template #footer>
          <el-button @click="copyDebugCommand" type="primary" plain>复制</el-button>
          <el-button @click="debugDialogVisible = false">关闭</el-button>
        </template>
      </el-dialog>

      <!-- 添加/编辑场景对话框 -->
      <el-dialog
        v-model="addSceneDialogVisible"
        :title="editingScene ? $t('common.edit') : $t('common.add')"
        width="700px"
      >
        <el-form :model="newScene" label-width="100px">
          <el-form-item :label="$t('common.image')">
            <el-upload
              class="avatar-uploader"
              :action="`/api/v1/upload/image`"
              :headers="uploadHeaders"
              :show-file-list="false"
              :on-success="handleSceneImageSuccess"
              :before-upload="beforeAvatarUpload"
            >
              <div
                v-if="hasImage(newScene)"
                class="avatar-wrapper"
                style="
                  width: 160px;
                  height: 90px;
                  position: relative;
                  overflow: hidden;
                  border-radius: 6px;
                "
              >
                <img
                  :src="getImageUrl(newScene)"
                  class="avatar"
                  style="width: 100%; height: 100%; object-fit: cover"
                />
                <div
                  class="avatar-overlay"
                  style="
                    position: absolute;
                    top: 0;
                    left: 0;
                    right: 0;
                    bottom: 0;
                    background: rgba(0, 0, 0, 0.5);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    opacity: 0;
                    transition: opacity 0.3s;
                    cursor: pointer;
                  "
                >
                  <el-icon style="color: white; font-size: 24px;">
                    <Plus />
                  </el-icon>
                </div>
              </div>
              <div
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
              >
                <el-icon><Plus /></el-icon>
              </div>
            </el-upload>
          </el-form-item>
          <el-form-item :label="$t('common.name')">
            <el-input
              v-model="newScene.location"
              :placeholder="$t('common.name')"
            />
          </el-form-item>
          <el-form-item label="时间">
            <el-input
              v-model="newScene.time"
              placeholder="请输入时间，例如：白天、夜晚、黄昏等"
            />
          </el-form-item>
          <el-form-item label="氛围">
            <el-input
              v-model="newScene.atmosphere"
              type="textarea"
              :rows="2"
              placeholder="请输入场景氛围，例如：紧张、温馨、神秘等"
            />
          </el-form-item>
          <el-form-item label="光线效果">
            <el-input
              v-model="newScene.lighting"
              placeholder="请输入光线效果，例如：柔和、强烈、昏暗等"
            />
          </el-form-item>
          <el-form-item label="声音效果">
            <el-input
              v-model="newScene.sound"
              placeholder="请输入声音效果，例如：安静、嘈杂、舒缓等"
            />
          </el-form-item>
          <el-form-item :label="$t('common.description')">
            <el-input
              v-model="newScene.prompt"
              type="textarea"
              :rows="4"
              :placeholder="$t('common.description')"
            />
          </el-form-item>
          <el-form-item label="参考图片">
            <el-upload
              v-model:file-list="sceneReferenceImages"
              :action="`/api/v1/upload/image`"
              :headers="uploadHeaders"
              list-type="picture-card"
              :on-success="handleSceneReferenceImageSuccess"
              :before-upload="beforeAvatarUpload"
              :limit="5"
              multiple
            >
              <el-icon><Plus /></el-icon>
              <template #file="{ file }">
                <div class="upload-file-card">
                  <img :src="file.url" class="el-upload-list__item-thumbnail" />
                  <span class="el-upload-list__item-actions">
                    <span class="el-upload-list__item-preview" @click="handlePictureCardPreview(file)">
                      <el-icon><ZoomIn /></el-icon>
                    </span>
                    <span class="el-upload-list__item-download" @click="downloadRefImage(file.url, file.name)">
                      <el-icon><Download /></el-icon>
                    </span>
                    <span class="el-upload-list__item-delete" @click="handleSceneRefRemove(file)">
                      <el-icon><Delete /></el-icon>
                    </span>
                  </span>
                </div>
              </template>
            </el-upload>
            <div style="color: #999; font-size: 12px; margin-top: 5px;">
              可以上传最多5张参考图片
            </div>
          </el-form-item>
          <el-form-item label="生成尺寸">
            <el-tag type="info" effect="plain">1:1 正方形 · 2048×2048（四视图网格）</el-tag>
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

      <!-- 添加/编辑道具对话框 -->
      <el-dialog
        v-model="addPropDialogVisible"
        :title="editingProp ? $t('common.edit') : $t('common.add')"
        width="700px"
      >
        <el-form :model="newProp" label-width="100px">
          <el-form-item :label="$t('common.image')">
            <el-upload
              class="avatar-uploader"
              :action="`/api/v1/upload/image`"
              :headers="uploadHeaders"
              :show-file-list="false"
              :on-success="handlePropImageSuccess"
              :before-upload="beforeAvatarUpload"
            >
              <div
                v-if="hasImage(newProp)"
                class="avatar-wrapper"
                style="
                  width: 100px;
                  height: 100px;
                  position: relative;
                  overflow: hidden;
                  border-radius: 6px;
                "
              >
                <img
                  :src="getImageUrl(newProp)"
                  class="avatar"
                  style="width: 100%; height: 100%; object-fit: cover"
                />
                <div
                  class="avatar-overlay"
                  style="
                    position: absolute;
                    top: 0;
                    left: 0;
                    right: 0;
                    bottom: 0;
                    background: rgba(0, 0, 0, 0.5);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    opacity: 0;
                    transition: opacity 0.3s;
                    cursor: pointer;
                  "
                >
                  <el-icon style="color: white; font-size: 24px;">
                    <Plus />
                  </el-icon>
                </div>
              </div>
              <div
                v-else
                class="avatar-uploader-icon"
                style="
                  border: 1px dashed #d9d9d9;
                  border-radius: 6px;
                  cursor: pointer;
                  position: relative;
                  overflow: hidden;
                  width: 100px;
                  height: 100px;
                  font-size: 28px;
                  color: #8c939d;
                  text-align: center;
                  line-height: 100px;
                "
              >
                <el-icon><Plus /></el-icon>
              </div>
            </el-upload>
          </el-form-item>
          <el-form-item :label="$t('prop.name')">
            <el-input v-model="newProp.name" :placeholder="$t('prop.name')" />
          </el-form-item>
          <el-form-item :label="$t('prop.type')">
            <el-input
              v-model="newProp.type"
              :placeholder="$t('prop.typePlaceholder')"
            />
          </el-form-item>
          <el-form-item :label="$t('prop.description')">
            <el-input
              v-model="newProp.description"
              type="textarea"
              :rows="3"
              :placeholder="$t('prop.description')"
            />
          </el-form-item>
          <el-form-item :label="$t('prop.prompt')">
            <el-input
              v-model="newProp.prompt"
              type="textarea"
              :rows="4"
              :placeholder="$t('prop.promptPlaceholder')"
            />
          </el-form-item>
          <el-form-item label="参考图片">
            <el-upload
              v-model:file-list="propReferenceImages"
              :action="`/api/v1/upload/image`"
              :headers="uploadHeaders"
              list-type="picture-card"
              :on-success="handlePropReferenceImageSuccess"
              :before-upload="beforeAvatarUpload"
              :limit="5"
              multiple
            >
              <el-icon><Plus /></el-icon>
              <template #file="{ file }">
                <div class="upload-file-card">
                  <img :src="file.url" class="el-upload-list__item-thumbnail" />
                  <span class="el-upload-list__item-actions">
                    <span class="el-upload-list__item-preview" @click="handlePictureCardPreview(file)">
                      <el-icon><ZoomIn /></el-icon>
                    </span>
                    <span class="el-upload-list__item-download" @click="downloadRefImage(file.url, file.name)">
                      <el-icon><Download /></el-icon>
                    </span>
                    <span class="el-upload-list__item-delete" @click="handlePropRefRemove(file)">
                      <el-icon><Delete /></el-icon>
                    </span>
                  </span>
                </div>
              </template>
            </el-upload>
            <div style="color: #999; font-size: 12px; margin-top: 5px;">
              可以上传最多5张参考图片
            </div>
          </el-form-item>
          <el-form-item label="生成尺寸">
            <el-tag type="info" effect="plain">4:3 横向 · 2304×1728（三视图）</el-tag>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="addPropDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button type="primary" @click="saveProp">{{
            $t("common.confirm")
          }}</el-button>
        </template>
      </el-dialog>

      <!-- 从剧本提取道具对话框 -->
      <el-dialog
        v-model="extractPropsDialogVisible"
        :title="$t('prop.extractTitle')"
        width="500px"
      >
        <el-form label-width="100px">
          <el-form-item :label="$t('prop.selectEpisode')">
            <el-select
              v-model="selectedExtractEpisodeId"
              :placeholder="$t('common.pleaseSelect')"
              style="width: 100%"
            >
              <el-option
                v-for="ep in sortedEpisodes"
                :key="ep.id"
                :label="ep.title"
                :value="ep.id"
              />
            </el-select>
          </el-form-item>
          <el-alert
            :title="$t('prop.extractTip')"
            type="info"
            :closable="false"
            show-icon
          />
        </el-form>
        <template #footer>
          <el-button @click="extractPropsDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button
            type="primary"
            @click="handleExtractProps"
            :disabled="!selectedExtractEpisodeId"
            >{{ $t("prop.startExtract") }}</el-button
          >
        </template>
      </el-dialog>

      <!-- 从剧本提取角色对话框 -->
      <el-dialog
        v-model="extractCharactersDialogVisible"
        :title="$t('character.extractTitle')"
        width="500px"
      >
        <el-form label-width="100px">
          <el-form-item :label="$t('character.selectEpisode')">
            <el-select
              v-model="selectedExtractEpisodeId"
              :placeholder="$t('common.pleaseSelect')"
              style="width: 100%"
            >
              <el-option
                v-for="ep in sortedEpisodes"
                :key="ep.id"
                :label="ep.title"
                :value="ep.id"
              />
            </el-select>
          </el-form-item>
          <el-alert
            :title="$t('character.extractTip')"
            type="info"
            :closable="false"
            show-icon
          />
        </el-form>
        <template #footer>
          <el-button @click="extractCharactersDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button
            type="primary"
            @click="handleExtractCharacters"
            :disabled="!selectedExtractEpisodeId"
            >{{ $t("character.startExtract") }}</el-button
          >
        </template>
      </el-dialog>

      <!-- 从剧本提取场景对话框 -->
      <el-dialog
        v-model="extractScenesDialogVisible"
        :title="$t('workflow.extractSceneDialogTitle')"
        width="500px"
      >
        <el-form label-width="100px">
          <el-form-item :label="$t('prop.selectEpisode')">
            <el-select
              v-model="selectedExtractEpisodeId"
              :placeholder="$t('common.pleaseSelect')"
              style="width: 100%"
            >
              <el-option
                v-for="ep in sortedEpisodes"
                :key="ep.id"
                :label="ep.title"
                :value="ep.id"
              />
            </el-select>
          </el-form-item>
          <el-alert
            :title="$t('workflow.extractSceneDialogTip')"
            type="info"
            :closable="false"
            show-icon
          />
        </el-form>
        <template #footer>
          <el-button @click="extractScenesDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button
            type="primary"
            @click="handleExtractScenes"
            :disabled="!selectedExtractEpisodeId"
            >{{ $t("storyboard.startExtract") }}</el-button
          >
        </template>
      </el-dialog>

      <!-- 编辑章节对话框 -->
      <el-dialog
        v-model="editEpisodeDialogVisible"
        title="编辑章节"
        width="500px"
      >
        <el-form label-width="100px">
          <el-form-item label="章节编号">
            <el-input v-model="editingEpisode.episode_number" disabled />
          </el-form-item>
          <el-form-item label="章节名称">
            <el-input v-model="editingEpisode.title" placeholder="请输入章节名称" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="editingEpisode.status" style="width: 100%">
              <el-option label="草稿" value="draft" />
              <el-option label="已创建" value="created" />
              <el-option label="生成中" value="generating" />
              <el-option label="已完成" value="completed" />
            </el-select>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="editEpisodeDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveEpisodeName">保存</el-button>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  ArrowLeft,
  Document,
  User,
  Picture,
  Plus,
  Box,
  MagicStick,
  Search,
  Download,
  ZoomIn,
  Delete,
} from "@element-plus/icons-vue";
import { dramaAPI } from "@/api/drama";
import { characterLibraryAPI } from "@/api/character-library";
import { propAPI } from "@/api/prop";
import { aiAPI } from "@/api/ai";
import type { Drama } from "@/types/drama";
import {
  AppHeader,
  StatCard,
  EmptyState,
  ImagePreview,
} from "@/components/common";
import CollapsibleText from "@/components/common/CollapsibleText.vue";
import { getImageUrl, hasImage } from "@/utils/image";

const router = useRouter();
const route = useRoute();

const uploadHeaders = computed(() => {
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
});

const drama = ref<Drama>();
const activeTab = ref((route.query.tab as string) || "overview");
const scenes = ref<any[]>([]);

let pollingTimer: any = null; // Add polling timer definition

const addCharacterDialogVisible = ref(false);
const addSceneDialogVisible = ref(false);
const addPropDialogVisible = ref(false);
const aiGenerateCharacterDialogVisible = ref(false);
const aiGeneratePropDialogVisible = ref(false);
const aiGenerateSceneDialogVisible = ref(false);
const extractPropsDialogVisible = ref(false);
const extractCharactersDialogVisible = ref(false);
const extractScenesDialogVisible = ref(false);
const editEpisodeDialogVisible = ref(false);

const editingCharacter = ref<any>(null);
const editingScene = ref<any>(null);
const editingProp = ref<any>(null);
const editingEpisode = ref<any>(null);
const selectedExtractEpisodeId = ref<number | null>(null);

// AI生成角色相关
const aiCharacterInput = ref('');
const aiCharacterOutput = ref('');
const aiCharacterName = ref('');
const aiCharacterRole = ref('');
const aiCharacterAppearance = ref('');
const aiCharacterPersonality = ref('');
const aiCharacterVoiceStyle = ref('');
const aiCharacterBackground = ref('');
const aiCharacterDescription = ref('');

// AI生成道具相关
const aiPropInput = ref('');
const aiPropOutput = ref('');
const aiPropName = ref('');
const aiPropType = ref('');
const aiPropDescription = ref('');
const aiPropPrompt = ref('');

// AI生成场景相关
const aiSceneInput = ref('');
const aiSceneOutput = ref('');
const aiSceneLocation = ref('');
const aiSceneTime = ref('');
const aiSceneDescription = ref('');
const aiSceneAtmosphere = ref('');
const aiSceneLighting = ref('');
const aiSceneSound = ref('');

const generateImageDialogVisible = ref(false);
const referenceImageList = ref<any[]>([]);
const baseReferenceImages = ref<any[]>([]);
const imageOrientation = ref('horizontal');
const imagePrompt = ref('');
const currentGenerateTarget = ref<any>(null);
const generatingImage = ref(false);
const generatingPrompt = ref(false);

const selectedSceneIds = ref<number[]>([]);
const selectedPropIds = ref<number[]>([]);
const batchSelecting = ref<'scene' | 'prop' | null>(null);
const batchGenerating = ref(false);
const batchProgress = ref({ current: 0, total: 0, currentName: '' });
const debugDialogVisible = ref(false);
const debugCurlCommand = ref('');

const newCharacter = ref({
  name: "",
  role: "supporting",
  appearance: "",
  personality: "",
  description: "",
  image_url: "",
  local_path: "",
  reference_images: [],
  image_orientation: "horizontal",
});

const characterReferenceImages = ref([]);

const propReferenceImages = ref<any[]>([]);

const newProp = ref({
  name: "",
  description: "",
  prompt: "",
  type: "",
  image_url: "",
  local_path: "",
  reference_images: [] as string[],
  image_orientation: "horizontal",
});

const newScene = ref({
  location: "",
  prompt: "",
  image_url: "",
  local_path: "",
  reference_images: [],
  image_orientation: "horizontal",
});

const sceneReferenceImages = ref([]);

const characterFilter = ref('');
const characterRoleFilter = ref('');
const sceneFilter = ref('');
const propFilter = ref('');

const episodesCount = computed(() => drama.value?.episodes?.length || 0);
const charactersCount = computed(() => drama.value?.characters?.length || 0);
const scenesCount = computed(() => scenes.value.length);
const propsCount = computed(() => drama.value?.props?.length || 0);

const sortedEpisodes = computed(() => {
  if (!drama.value?.episodes) return [];
  return [...drama.value.episodes].sort(
    (a, b) => a.episode_number - b.episode_number,
  );
});

const PV_EPISODE_THRESHOLD = 14;

const pvEpisodes = computed(() =>
  sortedEpisodes.value.filter((ep) => ep.episode_number <= PV_EPISODE_THRESHOLD),
);

const mainEpisodes = computed(() =>
  sortedEpisodes.value.filter((ep) => ep.episode_number > PV_EPISODE_THRESHOLD),
);

const filteredCharacters = computed(() => {
  if (!drama.value?.characters) return [];
  const keyword = characterFilter.value.trim().toLowerCase();
  const role = characterRoleFilter.value;
  return [...drama.value.characters]
    .filter(c => {
      const matchSelf = (!keyword || c.name?.toLowerCase().includes(keyword) || c.description?.toLowerCase().includes(keyword))
        && (!role || c.role === role);
      const matchChild = (c.children || []).some((child: any) =>
        (!keyword || child.name?.toLowerCase().includes(keyword) || child.outfit_name?.toLowerCase().includes(keyword) || child.description?.toLowerCase().includes(keyword))
        && (!role || child.role === role)
      );
      return matchSelf || matchChild;
    })
    .sort((a, b) => (a.name || '').localeCompare(b.name || '', 'zh-CN'));
});

const filteredScenes = computed(() => {
  const keyword = sceneFilter.value.trim().toLowerCase();
  return [...scenes.value]
    .filter(s => {
      if (!keyword) return true;
      return (s.name || '').toLowerCase().includes(keyword)
        || (s.location || '').toLowerCase().includes(keyword)
        || (s.description || '').toLowerCase().includes(keyword);
    })
    .sort((a, b) => (a.name || a.location || '').localeCompare(b.name || b.location || '', 'zh-CN'));
});

const filteredProps = computed(() => {
  if (!drama.value?.props) return [];
  const keyword = propFilter.value.trim().toLowerCase();
  return [...drama.value.props]
    .filter(p => {
      if (!keyword) return true;
      return (p.name || '').toLowerCase().includes(keyword)
        || (p.description || '').toLowerCase().includes(keyword);
    })
    .sort((a, b) => (a.name || '').localeCompare(b.name || '', 'zh-CN'));
});

// Helper for polling
const startPolling = (
  callback: () => Promise<void>,
  maxAttempts = 20,
  interval = 3000,
) => {
  if (pollingTimer) clearInterval(pollingTimer);

  let attempts = 0;
  pollingTimer = setInterval(async () => {
    attempts++;
    await callback();
    if (attempts >= maxAttempts) {
      if (pollingTimer) clearInterval(pollingTimer);
      pollingTimer = null;
    }
  }, interval);
};

// Clear timer on unmount
import { onUnmounted } from "vue";
onUnmounted(() => {
  if (pollingTimer) clearInterval(pollingTimer);
});

const loadDramaData = async () => {
  try {
    const data = await dramaAPI.get(route.params.id as string);
    drama.value = data;
    // 确保在获取drama数据后再加载场景
    await loadScenes();
  } catch (error: any) {
    console.error('Failed to load drama data:', error);
    ElMessage.error(error.message || "加载项目数据失败");
  }
};

const loadScenes = async () => {
  try {
    // 直接从API获取场景数据
    if (drama.value?.id) {
      const sceneData = await dramaAPI.getScenes(drama.value.id.toString());
      scenes.value = sceneData || [];
    } else {
      scenes.value = [];
    }
  } catch (error) {
    console.error('Failed to load scenes:', error);
    scenes.value = [];
  }
};

const getStatusType = (status?: string) => {
  const map: Record<string, any> = {
    draft: "info",
    in_progress: "warning",
    completed: "success",
  };
  return map[status || "draft"] || "info";
};

const getStatusText = (status?: string) => {
  const map: Record<string, string> = {
    draft: "草稿",
    in_progress: "制作中",
    completed: "已完成",
  };
  return map[status || "draft"] || "草稿";
};

const episodeStatusMap: Record<string, { text: string; type: string }> = {
  draft: { text: "草稿", type: "info" },
  created: { text: "已创建", type: "warning" },
  generating: { text: "生成中", type: "" },
  completed: { text: "已完成", type: "success" },
};

const getEpisodeStatusType = (episode: any) => {
  return episodeStatusMap[episode.status]?.type ?? "info";
};

const getEpisodeStatusText = (episode: any) => {
  return episodeStatusMap[episode.status]?.text ?? "草稿";
};

const formatDate = (date?: string) => {
  if (!date) return "-";
  return new Date(date).toLocaleString("zh-CN");
};

const createNewEpisode = () => {
  const nextEpisodeNumber = episodesCount.value + 1;
  router.push({
    name: "EpisodeWorkflowNew",
    params: {
      id: route.params.id,
      episodeNumber: nextEpisodeNumber,
    },
  });
};

const enterEpisodeWorkflow = (episode: any) => {
  router.push({
    name: "EpisodeWorkflowNew",
    params: {
      id: route.params.id,
      episodeNumber: episode.episode_number,
    },
  });
};

const deleteEpisode = async (episode: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除第${episode.episode_number}章吗？此操作将同时删除该章节的所有相关数据（角色、场景、分镜等）。`,
      "删除确认",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      },
    );

    // 过滤掉要删除的章节
    const existingEpisodes = drama.value?.episodes || [];
    const updatedEpisodes = existingEpisodes
      .filter((ep) => ep.episode_number !== episode.episode_number)
      .map((ep, index) => ({
        id: ep.id,
        episode_number: index + 1,
        title: ep.title,
        script_content: ep.script_content,
        description: ep.description,
        duration: ep.duration,
        status: ep.status,
      }));

    // 保存更新后的章节列表
    await dramaAPI.saveEpisodes(drama.value!.id, updatedEpisodes);

    ElMessage.success(`第${episode.episode_number}章删除成功`);
    await loadDramaData();
  } catch (error: any) {
    if (error !== "cancel") {
      ElMessage.error(error.message || "删除失败");
    }
  }
};

const editEpisodeName = (episode: any) => {
  editingEpisode.value = { ...episode };
  editEpisodeDialogVisible.value = true;
};

const saveEpisodeName = async () => {
  if (!editingEpisode.value || !drama.value) return;

  try {
    await dramaAPI.updateEpisodeTitle(
      drama.value!.id,
      editingEpisode.value.episode_number,
      editingEpisode.value.title,
      editingEpisode.value.status
    );
    ElMessage.success("章节信息修改成功");
    editEpisodeDialogVisible.value = false;
    await loadDramaData();
  } catch (error: any) {
    console.error('saveEpisodeName error', error);
    ElMessage.error(error.message || "修改失败");
  }
};

const openAddCharacterDialog = () => {
  editingCharacter.value = null;
  newCharacter.value = {
    name: "",
    role: "supporting",
    appearance: "",
    personality: "",
    description: "",
    image_url: "",
    local_path: "",
    reference_images: [],
    image_orientation: "horizontal",
    parent_id: undefined,
    outfit_name: "",
  };
  characterReferenceImages.value = [];
  addCharacterDialogVisible.value = true;
};

const openAddOutfitDialog = (parent: any) => {
  editingCharacter.value = null;
  newCharacter.value = {
    name: parent.name + "-",
    role: parent.role || "supporting",
    appearance: "",
    personality: "",
    description: "",
    image_url: "",
    local_path: "",
    reference_images: [],
    image_orientation: "horizontal",
    parent_id: parent.id,
    outfit_name: "",
  };
  characterReferenceImages.value = [];
  addCharacterDialogVisible.value = true;
};

// 打开AI生成角色对话框
const openAiGenerateCharacterDialog = () => {
  aiCharacterInput.value = '';
  aiCharacterOutput.value = '';
  aiGenerateCharacterDialogVisible.value = true;
};

// 使用AI处理角色信息
// 使用AI处理角色信息
const aiProcessCharacters = async () => {
  if (!aiCharacterInput.value.trim()) {
    ElMessage.warning("请输入要处理的文本");
    return;
  }

  try {
    // 构建专业的角色信息提取提示词
    const prompt = `请分析以下文本并提取其中的角色信息，按照以下标准格式返回：
【角色名称】角色的正式名称
【角色身份】角色在故事中的身份或职业
【外貌特征】角色的外观描述，包括身高、体型、发型、面部特征等
【性格特点】角色的性格、行为习惯、个性等
【声音特色】角色的声音特征（如果有）
【背景故事】角色的背景信息、经历等

处理规则：
1. 如果原文本中已经包含上述标签，请提取对应的内容
2. 如果原文本中没有某个标签，但能根据文本内容合理推断，请生成相应内容
3. 如果某个标签信息不合理、无法推断或原文本中没有相关信息，请留空
4. 请严格按照以上格式返回，不要添加额外说明

${aiCharacterInput.value}`;
    
    // 调用现有API进行文本处理 - 使用polish-prompt端点
    const response = await dramaAPI.polishPrompt({
      prompt: prompt,
      type: 'refine',
      orientation: 'horizontal',
      style: drama.value?.style || 'realistic'
    });
    
    aiCharacterOutput.value = response.polished_prompt;
    ElMessage.success("AI处理完成！");
    
    // 解析AI输出并填充到各个文本框
    const aiLines = response.polished_prompt.split('\n').map(line => line.trim()).filter(line => line);
    
    let aiName = '';
    let aiRole = '';
    let aiAppearance = '';
    let aiPersonality = '';
    let aiVoiceStyle = '';
    let aiBackground = '';
    
    for (const line of aiLines) {
      if (line.includes('【角色名称】')) {
        aiName = line.replace('【角色名称】', '').trim();
      } else if (line.includes('【角色身份】')) {
        aiRole = line.replace('【角色身份】', '').trim();
      } else if (line.includes('【外貌特征】')) {
        aiAppearance = line.replace('【外貌特征】', '').trim();
      } else if (line.includes('【性格特点】')) {
        aiPersonality = line.replace('【性格特点】', '').trim();
      } else if (line.includes('【声音特色】')) {
        aiVoiceStyle = line.replace('【声音特色】', '').trim();
      } else if (line.includes('【背景故事】')) {
        aiBackground = line.replace('【背景故事】', '').trim();
      }
    }
    
    aiCharacterName.value = aiName || '';
    aiCharacterRole.value = aiRole || '';
    aiCharacterAppearance.value = aiAppearance || '';
    aiCharacterPersonality.value = aiPersonality || '';
    aiCharacterVoiceStyle.value = aiVoiceStyle || '';
    aiCharacterBackground.value = aiBackground || '';
  } catch (error: any) {
    // 如果AI调用失败，回退到本地格式化
    const lines = aiCharacterInput.value.split('\n')
      .map(line => line.trim())
      .filter(line => line);
    
    // 专业的角色信息提取和格式化
    let name = '';
    let identity = '';
    let appearance = '';
    let personality = '';
    let voiceStyle = '';
    let background = '';
    let description = '';
    
    for (const line of lines) {
      if (/(角色)?名称|名字|姓名|称呼/.test(line)) {
        name = line.replace(/(角色)?名称|名字|姓名|称呼|[：:]/g, '').trim();
      } else if (/身份|职业|职位|地位/.test(line)) {
        identity = line.replace(/身份|职业|职位|地位|[：:]/g, '').trim();
      } else if (/外貌|外观|形象|长相|容貌|身材/.test(line)) {
        appearance += (appearance ? '；' : '') + line.replace(/外貌|外观|形象|长相|容貌|身材|[：:]/g, '').trim();
      } else if (/性格|个性|脾气|特质|行为/.test(line)) {
        personality += (personality ? '；' : '') + line.replace(/性格|个性|脾气|特质|行为|[：:]/g, '').trim();
      } else if (/声音|音色|语调|说话/.test(line)) {
        voiceStyle += (voiceStyle ? '；' : '') + line.replace(/声音|音色|语调|说话|[：:]/g, '').trim();
      } else if (/背景|经历|出身|历史|过去|家族/.test(line)) {
        background += (background ? '；' : '') + line.replace(/背景|经历|出身|历史|过去|家族|[：:]/g, '').trim();
      } else if (/描述|介绍|概要|简介/.test(line)) {
        description += (description ? '；' : '') + line.replace(/描述|介绍|概要|简介|[：:]/g, '').trim();
      } else if (!name && !identity && !appearance && !personality && !voiceStyle && !background && !description) {
        // 如果还没有提取到任何信息，将第一行作为名称
        name = line;
      }
    }
    
    let formattedText = '';
    if (name) formattedText += `【角色名称】${name}\n`;
    if (identity) formattedText += `【角色身份】${identity}\n`;
    if (appearance) formattedText += `【外貌特征】${appearance}\n`;
    if (personality) formattedText += `【性格特点】${personality}\n`;
    if (voiceStyle) formattedText += `【声音特色】${voiceStyle}\n`;
    if (background) formattedText += `【背景故事】${background}\n`;
    if (description) formattedText += `【角色描述】${description}\n`;
    
    if (!formattedText) {
      // 如果没有识别到关键词，则将原文本按行输出
      formattedText = lines.join('\n');
    }
    
    aiCharacterOutput.value = formattedText.trim();
    ElMessage.warning("AI处理失败，使用本地格式化结果");
    
    // 解析本地格式化结果并填充到各个文本框
    const formattedLines = formattedText.trim().split('\n').map(line => line.trim()).filter(line => line);
    
    let localName = '';
    let localRole = '';
    let localAppearance = '';
    let localPersonality = '';
    let localVoiceStyle = '';
    let localBackground = '';
    
    for (const line of formattedLines) {
      if (line.includes('【角色名称】')) {
        localName = line.replace('【角色名称】', '').trim();
      } else if (line.includes('【角色身份】')) {
        localRole = line.replace('【角色身份】', '').trim();
      } else if (line.includes('【外貌特征】')) {
        localAppearance = line.replace('【外貌特征】', '').trim();
      } else if (line.includes('【性格特点】')) {
        localPersonality = line.replace('【性格特点】', '').trim();
      } else if (line.includes('【声音特色】')) {
        localVoiceStyle = line.replace('【声音特色】', '').trim();
      } else if (line.includes('【背景故事】')) {
        localBackground = line.replace('【背景故事】', '').trim();
      }
    }
    
    aiCharacterName.value = localName || '';
    aiCharacterRole.value = localRole || '';
    aiCharacterAppearance.value = localAppearance || '';
    aiCharacterPersonality.value = localPersonality || '';
    aiCharacterVoiceStyle.value = localVoiceStyle || '';
    aiCharacterBackground.value = localBackground || '';
  }
};

// 保存AI生成的角色信息到表单
const saveAiGeneratedCharacters = () => {
  if (!aiCharacterName.value.trim()) {
    ElMessage.warning("没有可保存的内容");
    return;
  }

  editingCharacter.value = null;
  newCharacter.value = {
    name: aiCharacterName.value,
    role: aiCharacterRole.value || "supporting",
    appearance: aiCharacterAppearance.value,
    personality: aiCharacterPersonality.value,
    description: aiCharacterBackground.value,
    image_url: "",
    local_path: "",
    reference_images: [],
    image_orientation: "horizontal",
  };

  // 关闭AI对话框，打开添加角色对话框
  aiGenerateCharacterDialogVisible.value = false;
  addCharacterDialogVisible.value = true;
};

const handleCharacterAvatarSuccess = (response: any) => {
  if (response && response.data && response.data.url) {
    newCharacter.value.image_url = response.data.url;
    newCharacter.value.local_path = response.data.local_path || "";
  }
};

const handleCharacterReferenceImageSuccess = (response: any, file: any) => {
  if (response && response.data && response.data.url) {
    file.url = response.data.url;
    newCharacter.value.reference_images = characterReferenceImages.value
      .map(f => f.url)
      .filter(url => url);
  }
};

const handleSceneImageSuccess = (response: any) => {
  if (response && response.data && response.data.url) {
    newScene.value.image_url = response.data.url;
    newScene.value.local_path = response.data.local_path || "";
  }
};

const handleSceneReferenceImageSuccess = (response: any, file: any) => {
  if (response && response.data && response.data.url) {
    file.url = response.data.url;
    newScene.value.reference_images = sceneReferenceImages.value
      .map(f => f.url)
      .filter(url => url);
  }
};

const beforeAvatarUpload = (file: any) => {
  const isImage = file.type.startsWith("image/");
  const isLt10M = file.size / 1024 / 1024 < 10;

  if (!isImage) {
    ElMessage.error("只能上传图片文件!");
  }
  if (!isLt10M) {
    ElMessage.error("图片大小不能超过 10MB!");
  }
  return isImage && isLt10M;
};

const generateCharacterImage = async (character: any) => {
  currentGenerateTarget.value = { type: 'character', data: character };
  imagePrompt.value = character.prompt || character.appearance || character.description || '';
  imageOrientation.value = character.image_orientation || 'horizontal';
  baseReferenceImages.value = [];
  referenceImageList.value = toRefFileList(character.reference_images);
  generateImageDialogVisible.value = true;
};

const generateBaseImage = async (character: any) => {
  const basePrompt = `保留人物的长相和五官特征，纯白色背景，人物只穿白色贴身运动套装（运动背心+运动短裤），赤脚站立，从左到右并排展示正面、3/4侧面、背面三个全身站立视角。`;
  const fullPrompt = character.appearance
    ? `${character.appearance}\n\n${basePrompt}`
    : basePrompt;

  currentGenerateTarget.value = { type: 'character', mode: 'base_image', data: character };
  imagePrompt.value = fullPrompt;
  imageOrientation.value = character.image_orientation || 'horizontal';
  baseReferenceImages.value = [];

  referenceImageList.value = toRefFileList(character.reference_images || []);
  generateImageDialogVisible.value = true;
};

const generateOutfitImage = async (outfit: any, parent: any) => {
  const parentBaseImage = parent?.image_url || parent?.local_path;
  if (!parentBaseImage) {
    try {
      await ElMessageBox.confirm(
        '该角色尚未生成基础形象，建议先生成基础形象再生成造型图片。\n是否先去生成基础形象？',
        '缺少基础形象',
        { confirmButtonText: '去生成基础形象', cancelButtonText: '仍然继续', type: 'warning' }
      );
      generateBaseImage(parent);
      return;
    } catch {
      // user chose "仍然继续"
    }
  }

  const outfitDesc = outfit.description || outfit.appearance || outfit.outfit_name || '';
  const basePrompt = `保留人物的长相和五官特征，穿着以下服装造型：${outfitDesc}`;

  currentGenerateTarget.value = { type: 'character', mode: 'outfit', data: outfit, parent };
  imagePrompt.value = basePrompt;
  imageOrientation.value = outfit.image_orientation || parent?.image_orientation || 'horizontal';

  baseReferenceImages.value = parentBaseImage ? toRefFileList([parentBaseImage]) : [];
  referenceImageList.value = toRefFileList(outfit.reference_images || []);
  generateImageDialogVisible.value = true;
};

const toDisplayUrl = (url: string): string => {
  if (!url) return '';
  if (url.startsWith('http') || url.startsWith('/')) return url;
  return `/static/${url}`;
};

const toRawUrl = (url: string): string => {
  if (!url) return '';
  if (url.startsWith('/static/')) return url.slice('/static/'.length);
  return url;
};

const toRefFileList = (urls: string[]) =>
  (urls || []).map((url: string) => ({
    name: url.split('/').pop(),
    url: toDisplayUrl(url),
  }));

const downloadRefImage = async (url: string, filename?: string) => {
  if (!url) return;
  try {
    const response = await fetch(url);
    const blob = await response.blob();
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = filename || url.split('/').pop() || 'image.jpg';
    link.click();
    URL.revokeObjectURL(link.href);
  } catch {
    window.open(url, '_blank');
  }
};

const previewImageUrl = ref('');
const previewDialogVisible = ref(false);

const handlePictureCardPreview = (file: any) => {
  previewImageUrl.value = file.url;
  previewDialogVisible.value = true;
};

const handleUploadRemove = (file: any) => {
  const idx = referenceImageList.value.findIndex((f: any) => f.uid === file.uid || f.url === file.url);
  if (idx > -1) referenceImageList.value.splice(idx, 1);
};

const syncCharRefToModel = () => {
  newCharacter.value.reference_images = characterReferenceImages.value
    .map((f: any) => toRawUrl(f.url))
    .filter((url: string) => url);
};

const handleCharRefRemove = (file: any) => {
  const idx = characterReferenceImages.value.findIndex((f: any) => f.uid === file.uid || f.url === file.url);
  if (idx > -1) characterReferenceImages.value.splice(idx, 1);
  syncCharRefToModel();
};

const syncSceneRefToModel = () => {
  newScene.value.reference_images = sceneReferenceImages.value
    .map((f: any) => toRawUrl(f.url))
    .filter((url: string) => url);
};

const handleSceneRefRemove = (file: any) => {
  const idx = sceneReferenceImages.value.findIndex((f: any) => f.uid === file.uid || f.url === file.url);
  if (idx > -1) sceneReferenceImages.value.splice(idx, 1);
  syncSceneRefToModel();
};

const syncPropRefToModel = () => {
  newProp.value.reference_images = propReferenceImages.value
    .map((f: any) => toRawUrl(f.url))
    .filter((url: string) => url);
};

const handlePropRefRemove = (file: any) => {
  const idx = propReferenceImages.value.findIndex((f: any) => f.uid === file.uid || f.url === file.url);
  if (idx > -1) propReferenceImages.value.splice(idx, 1);
  syncPropRefToModel();
};

const getAllReferenceImages = () => {
  const baseUrls = baseReferenceImages.value
    .map(file => toRawUrl(file.url))
    .filter(url => url);
  const outfitUrls = referenceImageList.value
    .map(file => toRawUrl(file.url))
    .filter(url => url);
  return [...baseUrls, ...outfitUrls];
};

const handleReferenceImageSuccess = (response: any, file: any) => {
  if (response && response.data && response.data.url) {
    file.url = response.data.url;
  }
};

const aiGeneratePrompt = async () => {
  if (!currentGenerateTarget.value) return;
  
  generatingPrompt.value = true;
  
  try {
    const referenceImages = getAllReferenceImages();
    
    const sizeInfo = imageOrientation.value === 'horizontal' ? '16:9横屏比例' : '9:16竖屏比例';
    const target = currentGenerateTarget.value;
    
    const styleKey = drama.value?.style || 'realistic';
    const styleNameMap: Record<string, string> = {
      realistic: '超写实摄影风格',
      comic: '漫画风格',
    };
    const currentStyleName = styleNameMap[styleKey] || styleKey + '风格';

    let baseDescription = '';
    if (target.type === 'character') {
      baseDescription = target.data.appearance || target.data.description || target.data.prompt || '';
    } else if (target.type === 'scene') {
      baseDescription = target.data.description || target.data.prompt || '';
    } else if (target.type === 'prop') {
      baseDescription = target.data.description || target.data.prompt || '';
    }
    
    let prompt = '';
    if (target.type === 'character' && target.mode === 'base_image') {
      prompt = `请根据以下角色信息生成「角色基础形象三视图设定图」的图片提示词。
提示词用于豆包Seedream模型，用简洁精准的自然语言描述，不超过300字。

【重要】这是角色的基础形象三视图，用于后续换装的基准。必须只穿白色贴身运动套装（运动背心+运动短裤），不穿任何外套、裙子、长裤、鞋子等额外服饰，赤脚站立。

画面风格：${currentStyleName}
角色描述：${baseDescription}
${referenceImages.length > 0 ? `参考图片：已上传${referenceImages.length}张参考图片，请严格参考图中人物的长相五官特征` : ''}

格式要求：
1. 开头写明「${currentStyleName}」
2. 4:3横向构图，纯白色背景，无场景无装饰无文字
3. 一张图内从左到右并排展示正面、3/4侧面、背面三个全身站立视角
4. 每个视角均为头顶至脚底完整全身像，双手自然垂放身侧

描述要求：
5. 从头到脚依次描述：发型发色→面部五官→体型身材→肤色
6. 服装只写「穿着白色贴身运动背心和运动短裤，赤脚」，不要描述任何其他服饰、配饰、鞋子
7. 严格保留参考图中人物的长相和五官特征，这是最重要的
8. 用具体的颜色、形状词汇描述五官，避免抽象形容
9. 只输出纯视觉描述，不写光影、剧情、分辨率
10. 直接输出提示词，不要解释`;
    } else if (target.type === 'character' && target.mode === 'outfit') {
      const outfitName = target.data.outfit_name || target.data.name || '';
      const parentAppearance = target.parent?.appearance || '';
      const baseImgCount = baseReferenceImages.value.length;
      const outfitImgCount = referenceImageList.value.filter(f => f.url).length;
      let refDesc = '';
      if (baseImgCount > 0 || outfitImgCount > 0) {
        refDesc += '参考图片说明：\n';
        if (baseImgCount > 0) {
          refDesc += `- 参考图片1~${baseImgCount} 为「角色基础形象」（纯白背景运动套装三视图），必须严格保留其中人物的长相五官特征\n`;
        }
        if (outfitImgCount > 0) {
          const startIdx = baseImgCount + 1;
          refDesc += `- 参考图片${startIdx}~${startIdx + outfitImgCount - 1} 为「装扮参考」，参考其中的服装造型样式\n`;
        }
      }
      prompt = `请根据以下信息生成「角色造型三视图设定图」的图片提示词。
提示词用于豆包Seedream模型，用简洁精准的自然语言描述，不超过300字。

【重要】这是基于角色基础形象的换装三视图。必须保留基础形象中人物的长相五官，只改变服装造型。

画面风格：${currentStyleName}
造型名称：${outfitName}
造型描述：${baseDescription}
角色基础外貌：${parentAppearance}
${refDesc}

格式要求：
1. 开头写明「${currentStyleName}」
2. 4:3横向构图，纯白色背景，无场景无装饰无文字
3. 一张图内从左到右并排展示正面、3/4侧面、背面三个全身站立视角
4. 每个视角均为头顶至脚底完整全身像，双手自然垂放身侧
5. 三视角外貌服饰配色统一

描述要求：
6. 先简要描述人物长相特征（发型发色、五官、体型），确保与基础形象一致
7. 重点详细描述造型服装：从上到下依次写上身服装→腰部→下身服装→腿部→鞋子
8. 用具体的颜色、材质、形状词汇描述服装细节
9. 不要写手持道具、手部动作、表情等，只描述人物外貌和服装
10. 只输出纯视觉描述，不写光影、剧情、分辨率
11. 直接输出提示词，不要解释`;
    } else if (target.type === 'character') {
      prompt = `请根据以下角色信息生成「角色三视图设定图」的图片提示词。
提示词用于豆包Seedream模型，用简洁精准的自然语言描述，不超过300字。

画面风格：${currentStyleName}
角色描述：${baseDescription}
${referenceImages.length > 0 ? `参考图片：已上传${referenceImages.length}张参考图片，请参考其风格` : ''}

格式要求：
1. 开头写明「${currentStyleName}」
2. 4:3横向构图，纯白背景，无场景无装饰无文字
3. 一张图内从左到右并排展示正面、3/4侧面、背面三个全身站立视角
4. 写明「身材修长高挑，头身比1:7.5，腿部修长占身高一半以上」
5. 每个视角均为头顶至鞋底完整全身像，三视角外貌服饰配色统一

描述要求：
6. 从头到脚依次描述：发型发色→面部五官→颈部配饰→上身服装→腰部→下身服装→腿部→鞋子
7. 用具体的颜色、材质、形状词汇，避免抽象形容（如"气质优雅"换成具体的姿态描述）
8. 双手自然垂放身侧，不持道具
9. 只输出纯视觉描述，不写光影、剧情、分辨率
10. 直接输出提示词，不要解释`;
    } else if (target.type === 'prop') {
      prompt = `请根据以下道具信息生成「道具三视图设定图」的图片提示词。
提示词用于豆包Seedream模型，用简洁精准的自然语言描述，不超过300字。

画面风格：${currentStyleName}
道具描述：${baseDescription}
${referenceImages.length > 0 ? `参考图片：已上传${referenceImages.length}张参考图片，请参考其风格` : ''}

格式要求：
1. 开头写明「${currentStyleName}」
2. 4:3横向构图，纯白背景，无场景无装饰无文字
3. 一张图内从左到右并排展示正面、3/4侧面、背面（或特写细节）三个视角
4. 各视角的外观、材质、配色完全一致

描述要求：
5. 详细描述道具的整体造型、材质质感、颜色配色、关键细节特征
6. 用具体的颜色、材质、形状词汇，避免抽象形容
7. 只输出纯视觉描述，不写光影、剧情、分辨率
8. 直接输出提示词，不要解释`;
    } else {
      const sceneName = target.data.name || target.data.location || '';
      prompt = `请根据以下场景信息生成一张「21:9超宽银幕电影级空镜头」的图片生成提示词。
提示词用于豆包Seedream模型，用简洁连贯的自然语言描述。提示词不超过300个汉字。

画面风格：${currentStyleName}
场景名称：${sceneName}
场景位置：${target.data.location || ''}
场景描述：${baseDescription}
${referenceImages.length > 0 ? `参考图片：已上传${referenceImages.length}张参考图片，请参考其风格和构图` : ''}

核心要求：
1. 提示词开头必须明确写出「${currentStyleName}」，确保风格贯穿整个描述
2. 21:9超宽银幕构图，电影级场景空镜头，无人物
3. 采用中远景机位，具有电影感的纵深透视与层次分明的前中远景构图
4. 详细描述画面内容，包括空间结构、物件摆放、材质颜色、光影氛围
5. ${referenceImages.length > 0 ? '参考上传图片的风格和构图，保持一致性' : '自行设计合理的构图'}
6. 重要：只输出纯视觉描述，去掉所有剧情叙事（如人物关系、故事背景等），图片模型不理解剧情
7. 直接输出提示词，不要解释，不要出现分辨率数值`;
    }

    const response = await dramaAPI.polishPrompt({
      prompt: prompt,
      type: target.type,
      orientation: imageOrientation.value,
      style: drama.value?.style || 'realistic',
      reference_images: referenceImages.length > 0 ? referenceImages : undefined
    });
    
    let polished = response.polished_prompt || '';

    if (target.type === 'character' && target.mode === 'outfit') {
      const baseUrls = baseReferenceImages.value.map(f => toRawUrl(f.url)).filter(u => u);
      const outfitUrls = referenceImageList.value.map(f => toRawUrl(f.url)).filter(u => u);
      if (baseUrls.length > 0 || outfitUrls.length > 0) {
        const cutIdx = polished.indexOf('\n\n输入的参考图片说明');
        const promptBody = cutIdx !== -1 ? polished.substring(0, cutIdx) : polished;

        let labeledRef = '\n\n输入的参考图片说明（按顺序对应输入的图片）：\n';
        let idx = 1;
        for (const url of baseUrls) {
          labeledRef += `【参考图片${idx}】${url}（角色基础形象，请严格保留长相五官）\n`;
          idx++;
        }
        for (const url of outfitUrls) {
          labeledRef += `【参考图片${idx}】${url}（装扮参考）\n`;
          idx++;
        }
        labeledRef += '请严格保留参考图片中基础形象的长相五官，仅改变服装造型。';
        polished = promptBody + labeledRef;
      }
    } else if (target.type === 'character' && target.mode === 'base_image') {
      const cutIdx = polished.indexOf('\n\n输入的参考图片说明');
      if (cutIdx !== -1) {
        const promptBody = polished.substring(0, cutIdx);
        let labeledRef = '\n\n输入的参考图片说明（按顺序对应输入的图片）：\n';
        referenceImages.forEach((url: string, i: number) => {
          labeledRef += `【参考图片${i + 1}】${url}（外貌参考，请严格保留长相五官特征）\n`;
        });
        labeledRef += '请严格按照以上参考图片中人物的长相五官来生成基础形象。';
        polished = promptBody + labeledRef;
      }
    }

    imagePrompt.value = polished;
    ElMessage.success("AI提示词生成成功！");
  } catch (error: any) {
    ElMessage.error(error.message || "生成提示词失败");
  } finally {
    generatingPrompt.value = false;
  }
};

const confirmGenerateImage = async () => {
  await saveImageConfig();
  
  if (!currentGenerateTarget.value) return;
  
  generatingImage.value = true;
  
  try {
    const referenceImages = getAllReferenceImages();
    
    const target = currentGenerateTarget.value;
    
    if (target.type === 'character') {
      const character = target.data;
      await characterLibraryAPI.generateCharacterImage(
        character.id,
        undefined,
        imagePrompt.value || undefined,
        referenceImages.length > 0 ? referenceImages : undefined,
        drama.value?.style || 'realistic'
      );
    } else if (target.type === 'scene') {
      const scene = target.data;
      await dramaAPI.generateSceneImage({
        scene_id: scene.id,
        prompt: imagePrompt.value || undefined,
        reference_images: referenceImages.length > 0 ? referenceImages : undefined
      });
    } else if (target.type === 'prop') {
      const prop = target.data;
      await propAPI.generateImage(
        prop.id,
        imagePrompt.value || undefined,
        referenceImages.length > 0 ? referenceImages : undefined
      );
    }
    
    ElMessage.success("图片生成任务已提交");
    generateImageDialogVisible.value = false;
    startPolling(loadDramaData);
  } catch (error: any) {
    ElMessage.error(error.message || "生成失败");
  } finally {
    generatingImage.value = false;
  }
};

const batchGenerateForType = async (type: 'scene' | 'prop') => {
  const ids = type === 'scene' ? selectedSceneIds.value : selectedPropIds.value;
  if (ids.length === 0) {
    ElMessage.warning(`请先勾选要生成的${type === 'scene' ? '场景' : '道具'}`);
    return;
  }

  const items = type === 'scene'
    ? filteredScenes.value.filter((s: any) => ids.includes(s.id))
    : filteredProps.value.filter((p: any) => ids.includes(p.id));

  try {
    await ElMessageBox.confirm(
      `将为 ${items.length} 个${type === 'scene' ? '场景' : '道具'}并发执行：AI生成提示词 → 保存 → 开始生成图片。继续？`,
      '批量生成图片',
      { confirmButtonText: '开始', cancelButtonText: '取消', type: 'info' }
    );
  } catch { return; }

  batchGenerating.value = true;
  batchProgress.value = { current: 0, total: items.length, currentName: '' };

  const styleKey = drama.value?.style || 'realistic';
  const styleNameMap: Record<string, string> = { realistic: '超写实摄影风格', comic: '漫画风格' };
  const currentStyleName = styleNameMap[styleKey] || styleKey + '风格';

  const generateOne = async (item: any) => {
    const itemName = type === 'scene' ? (item.name || item.location) : item.name;
    try {
      const baseDescription = item.description || item.prompt || '';
      let aiPrompt = '';
      if (type === 'scene') {
        const sceneName = item.name || item.location || '';
        aiPrompt = `请根据以下场景信息生成一张「21:9超宽银幕电影级空镜头」的图片生成提示词。
提示词用于豆包Seedream模型，用简洁连贯的自然语言描述。提示词不超过300个汉字。

画面风格：${currentStyleName}
场景名称：${sceneName}
场景位置：${item.location || ''}
场景描述：${baseDescription}

核心要求：
1. 提示词开头必须明确写出「${currentStyleName}」，确保风格贯穿整个描述
2. 21:9超宽银幕构图，电影级场景空镜头，无人物
3. 采用中远景机位，具有电影感的纵深透视与层次分明的前中远景构图
4. 详细描述画面内容，包括空间结构、物件摆放、材质颜色、光影氛围
5. 重要：只输出纯视觉描述，去掉所有剧情叙事（如人物关系、故事背景等），图片模型不理解剧情
6. 直接输出提示词，不要解释，不要出现分辨率数值`;
      } else {
        aiPrompt = `请根据以下道具信息生成「道具三视图设定图」的图片提示词。
提示词用于豆包Seedream模型，用简洁精准的自然语言描述，不超过300字。

画面风格：${currentStyleName}
道具描述：${baseDescription}

格式要求：
1. 开头写明「${currentStyleName}」
2. 4:3横向构图，纯白背景，无场景无装饰无文字
3. 一张图内从左到右并排展示正面、3/4侧面、背面（或特写细节）三个视角
4. 各视角的外观、材质、配色完全一致

描述要求：
5. 详细描述道具的整体造型、材质质感、颜色配色、关键细节特征
6. 用具体的颜色、材质、形状词汇，避免抽象形容
7. 只输出纯视觉描述，不写光影、剧情、分辨率
8. 直接输出提示词，不要解释`;
      }

      const polishResponse = await dramaAPI.polishPrompt({
        prompt: aiPrompt,
        type: type,
        orientation: 'horizontal',
        style: styleKey,
      });
      const polishedPrompt = polishResponse.polished_prompt || '';

      if (type === 'scene') {
        await dramaAPI.updateScenePrompt(item.id.toString(), '', polishedPrompt);
        await dramaAPI.generateSceneImage({ scene_id: item.id, prompt: polishedPrompt });
      } else {
        await propAPI.update(item.id, { prompt: polishedPrompt });
        await propAPI.generateImage(item.id, polishedPrompt);
      }

      return { name: itemName, success: true };
    } catch (error: any) {
      return { name: itemName, success: false, error: error.message || '未知错误' };
    }
  };

  const results = await Promise.allSettled(items.map((item: any) => generateOne(item)));

  let successCount = 0;
  let failMessages: string[] = [];
  for (const r of results) {
    if (r.status === 'fulfilled' && r.value.success) {
      successCount++;
    } else {
      const val = r.status === 'fulfilled' ? r.value : { name: '?', error: String(r.reason) };
      failMessages.push(`${val.name}: ${val.error}`);
    }
  }

  batchGenerating.value = false;
  batchSelecting.value = null;
  selectedSceneIds.value = [];
  selectedPropIds.value = [];

  if (failMessages.length === 0) {
    ElMessage.success(`${successCount} 个${type === 'scene' ? '场景' : '道具'}全部提交成功！`);
  } else {
    ElMessage.warning(`成功 ${successCount} 个，失败 ${failMessages.length} 个：${failMessages.join('；')}`);
  }

  startPolling(loadDramaData);
};

const debugGenerateImage = () => {
  if (!currentGenerateTarget.value) return;

  const referenceImages = getAllReferenceImages();

  const target = currentGenerateTarget.value;
  const apiBase = window.location.origin + '/api/v1';
  let endpoint = '';
  let body: any = {};

  if (target.type === 'character') {
    const character = target.data;
    endpoint = `${apiBase}/characters/${character.id}/generate-image`;
    body = {
      prompt: imagePrompt.value || undefined,
      style: drama.value?.style || 'realistic',
      reference_images: referenceImages.length > 0 ? referenceImages : undefined,
    };
  } else if (target.type === 'scene') {
    const scene = target.data;
    endpoint = `${apiBase}/scenes/generate-image`;
    body = {
      scene_id: scene.id,
      prompt: imagePrompt.value || undefined,
      reference_images: referenceImages.length > 0 ? referenceImages : undefined,
    };
  } else if (target.type === 'prop') {
    const prop = target.data;
    endpoint = `${apiBase}/props/${prop.id}/generate`;
    body = {
      prompt: imagePrompt.value || undefined,
      reference_images: referenceImages.length > 0 ? referenceImages : undefined,
    };
  }

  const curlBody = JSON.stringify(body, null, 2);
  const curlCmd = `curl -X POST '${endpoint}' \\\n  -H 'Content-Type: application/json' \\\n  -H 'Authorization: Bearer <YOUR_TOKEN>' \\\n  -d '${curlBody.replace(/'/g, "'\\''")}'`;

  debugCurlCommand.value = curlCmd;
  debugDialogVisible.value = true;
};

const copyDebugCommand = async () => {
  try {
    await navigator.clipboard.writeText(debugCurlCommand.value);
    ElMessage.success("已复制到剪贴板");
  } catch {
    const textarea = document.createElement('textarea');
    textarea.value = debugCurlCommand.value;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand('copy');
    document.body.removeChild(textarea);
    ElMessage.success("已复制到剪贴板");
  }
};

const saveImageConfig = async () => {
  if (!currentGenerateTarget.value) return;
  
  try {
    const target = currentGenerateTarget.value;
    const isOutfitMode = target.type === 'character' && target.mode === 'outfit';
    const referenceImages = isOutfitMode
      ? referenceImageList.value.map(file => toRawUrl(file.url)).filter(url => url)
      : getAllReferenceImages();
    
    if (target.type === 'character') {
      const character = target.data;
      await characterLibraryAPI.updateCharacter(character.id, {
        prompt: imagePrompt.value,
        reference_images: referenceImages,
        image_orientation: imageOrientation.value
      });
    } else if (target.type === 'scene') {
      const scene = target.data;
      await dramaAPI.updateScenePrompt(
        scene.id.toString(),
        '',
        imagePrompt.value || '',
        referenceImages.length > 0 ? referenceImages : undefined,
        imageOrientation.value
      );
    } else if (target.type === 'prop') {
      const prop = target.data;
      await propAPI.update(prop.id, {
        prompt: imagePrompt.value,
        reference_images: referenceImages.length > 0 ? referenceImages : undefined,
        image_orientation: imageOrientation.value
      });
    }
    
    await loadDramaData();
    ElMessage.success("配置保存成功！");
  } catch (error: any) {
    ElMessage.error(error.message || "保存失败");
  }
};

const openExtractCharacterDialog = () => {
  extractCharactersDialogVisible.value = true;
  if (sortedEpisodes.value.length > 0 && !selectedExtractEpisodeId.value) {
    selectedExtractEpisodeId.value = sortedEpisodes.value[0].id;
  }
};

const handleExtractCharacters = async () => {
  if (!selectedExtractEpisodeId.value) return;

  try {
    const res = await characterLibraryAPI.extractFromEpisode(
      selectedExtractEpisodeId.value,
    );
    extractCharactersDialogVisible.value = false;

    // 自动刷新几次
    let checkCount = 0;
    const checkInterval = setInterval(() => {
      loadDramaData();
      checkCount++;
      if (checkCount > 10) clearInterval(checkInterval);
    }, 5000);
  } catch (error: any) {
    ElMessage.error(error.message || "提取失败");
  }
};

const generateSceneImage = async (scene: any) => {
  currentGenerateTarget.value = { type: 'scene', data: scene };
  imagePrompt.value = scene.prompt || scene.description || '';
  imageOrientation.value = scene.image_orientation || 'horizontal';
  baseReferenceImages.value = [];
  referenceImageList.value = toRefFileList(scene.reference_images);
  generateImageDialogVisible.value = true;
};

const openExtractSceneDialog = () => {
  extractScenesDialogVisible.value = true;
  if (sortedEpisodes.value.length > 0 && !selectedExtractEpisodeId.value) {
    selectedExtractEpisodeId.value = sortedEpisodes.value[0].id;
  }
};

const handleExtractScenes = async () => {
  if (!selectedExtractEpisodeId.value) return;

  try {
    const res = await dramaAPI.extractBackgrounds(
      selectedExtractEpisodeId.value.toString(),
    );
    extractScenesDialogVisible.value = false;

    // 自动刷新几次
    let checkCount = 0;
    const checkInterval = setInterval(() => {
      loadScenes();
      checkCount++;
      if (checkCount > 10) clearInterval(checkInterval);
    }, 5000);
  } catch (error: any) {
    ElMessage.error(error.message || "提取失败");
  }
};

const saveCharacter = async () => {
  if (!newCharacter.value.name.trim()) {
    ElMessage.warning("请输入角色名称");
    return;
  }

  syncCharRefToModel();

  try {
    if (editingCharacter.value) {
      // Edit existing character using dedicated update endpoint
      await dramaAPI.updateCharacter(editingCharacter.value.id, {
        name: newCharacter.value.name,
        role: newCharacter.value.role,
        appearance: newCharacter.value.appearance,
        personality: newCharacter.value.personality,
        description: newCharacter.value.description,
        image_url: newCharacter.value.image_url,
        local_path: newCharacter.value.local_path,
        reference_images: newCharacter.value.reference_images,
        image_orientation: newCharacter.value.image_orientation,
      });
      ElMessage.success("角色更新成功");
    } else if (newCharacter.value.parent_id) {
      // Add outfit variant via createCharacter
      await dramaAPI.createCharacter({
        drama_id: drama.value!.id,
        parent_id: newCharacter.value.parent_id,
        name: newCharacter.value.name,
        outfit_name: newCharacter.value.outfit_name,
        role: newCharacter.value.role,
        appearance: newCharacter.value.appearance,
        personality: newCharacter.value.personality,
        description: newCharacter.value.description,
        image_url: newCharacter.value.image_url,
        local_path: newCharacter.value.local_path,
        reference_images: newCharacter.value.reference_images,
        image_orientation: newCharacter.value.image_orientation,
      });
      ElMessage.success("造型添加成功");
    } else {
      // Add new character
      const allCharacters = [
        ...(drama.value?.characters || []).map((c) => ({
          name: c.name,
          role: c.role,
          appearance: c.appearance,
          personality: c.personality,
          description: c.description,
          image_url: c.image_url,
          local_path: c.local_path,
          reference_images: c.reference_images,
          image_orientation: c.image_orientation,
        })),
        newCharacter.value,
      ];

      await dramaAPI.saveCharacters(drama.value!.id, allCharacters);
      ElMessage.success("角色添加成功");
    }

    addCharacterDialogVisible.value = false;
    await loadDramaData();
  } catch (error: any) {
    ElMessage.error(error.message || "操作失败");
  }
};

const editCharacter = (character: any) => {
  editingCharacter.value = character;
  
  let referenceImages: string[] = [];
  if (character.reference_images) {
    if (typeof character.reference_images === 'string') {
      referenceImages = [character.reference_images];
    } else if (Array.isArray(character.reference_images)) {
      referenceImages = character.reference_images.map((item: any) => {
        if (typeof item === 'string') {
          return item;
        } else if (item && item.url) {
          return item.url;
        }
        return '';
      }).filter((url: string) => url);
    }
  }
  
  newCharacter.value = {
    name: character.name,
    role: character.role || "supporting",
    appearance: character.appearance || "",
    personality: character.personality || "",
    description: character.description || "",
    image_url: character.image_url || "",
    local_path: character.local_path || "",
    reference_images: referenceImages,
    image_orientation: character.image_orientation || "horizontal",
  };
  
  characterReferenceImages.value = referenceImages.map((url: string) => ({
    url: toDisplayUrl(url),
    name: url.split('/').pop()
  }));
  
  addCharacterDialogVisible.value = true;
};

const deleteCharacter = async (character: any) => {
  if (!character.id) {
    ElMessage.error("角色ID不存在，无法删除");
    return;
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除角色"${character.name}"吗？此操作不可恢复。`,
      "删除确认",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      },
    );

    await characterLibraryAPI.deleteCharacter(character.id);
    ElMessage.success("角色已删除");
    await loadDramaData();
  } catch (error: any) {
    if (error !== "cancel") {
      console.error("删除角色失败:", error);
      ElMessage.error(error.message || "删除失败");
    }
  }
};

const openAddSceneDialog = () => {
  editingScene.value = null;
  newScene.value = {
    location: "",
    time: "",
    atmosphere: "",
    lighting: "",
    sound: "",
    prompt: "",
    image_url: "",
    local_path: "",
    reference_images: [],
    image_orientation: "horizontal",
  };
  sceneReferenceImages.value = [];
  addSceneDialogVisible.value = true;
};

// 打开AI生成场景对话框
const openAiGenerateSceneDialog = () => {
  aiSceneInput.value = '';
  aiSceneOutput.value = '';
  aiGenerateSceneDialogVisible.value = true;
};

// 使用AI处理场景信息
const aiProcessScenes = async () => {
  if (!aiSceneInput.value.trim()) {
    ElMessage.warning("请输入要处理的文本");
    return;
  }

  try {
    // 构建专业的场景信息提取提示词
    const prompt = `请分析以下文本并提取其中的场景信息，按照以下标准格式返回：
【场景地点】场景发生的地理位置或场所
【时间】场景发生的时间（时段、季节等）
【场景描述】场景的详细环境描述
【氛围】场景的整体氛围和情绪感受
【光线效果】场景中的光线条件和照明效果
【声音效果】场景中的声音环境和音效

处理规则：
1. 如果原文本中已经包含上述标签，请提取对应的内容
2. 如果原文本中没有某个标签，但能根据文本内容合理推断，请生成相应内容
3. 如果某个标签信息不合理、无法推断或原文本中没有相关信息，请留空
4. 请严格按照以上格式返回，不要添加额外说明

${aiSceneInput.value}`;
    
    // 调用现有API进行文本处理 - 使用polish-prompt端点
    const response = await dramaAPI.polishPrompt({
      prompt: prompt,
      type: 'refine',
      orientation: 'horizontal',
      style: drama.value?.style || 'realistic'
    });
    
    aiSceneOutput.value = response.polished_prompt;
    ElMessage.success("AI处理完成！");
    
    // 解析AI输出并填充到各个文本框
    const aiLines = response.polished_prompt.split('\n').map(line => line.trim()).filter(line => line);
    
    let aiLocation = '';
    let aiTime = '';
    let aiDescription = '';
    let aiAtmosphere = '';
    let aiLighting = '';
    let aiSound = '';
    
    for (const line of aiLines) {
      if (line.includes('【场景地点】')) {
        aiLocation = line.replace('【场景地点】', '').trim();
      } else if (line.includes('【时间】')) {
        aiTime = line.replace('【时间】', '').trim();
      } else if (line.includes('【场景描述】')) {
        aiDescription = line.replace('【场景描述】', '').trim();
      } else if (line.includes('【氛围】')) {
        aiAtmosphere = line.replace('【氛围】', '').trim();
      } else if (line.includes('【光线效果】')) {
        aiLighting = line.replace('【光线效果】', '').trim();
      } else if (line.includes('【声音效果】')) {
        aiSound = line.replace('【声音效果】', '').trim();
      }
    }
    
    aiSceneLocation.value = aiLocation || '';
    aiSceneTime.value = aiTime || '';
    aiSceneDescription.value = aiDescription || '';
    aiSceneAtmosphere.value = aiAtmosphere || '';
    aiSceneLighting.value = aiLighting || '';
    aiSceneSound.value = aiSound || '';
  } catch (error: any) {
    // 如果AI调用失败，回退到本地格式化
    const lines = aiSceneInput.value.split('\n')
      .map(line => line.trim())
      .filter(line => line);
    
    let location = '';
    let time = '';
    let description = '';
    let atmosphere = '';
    let lighting = '';
    let sound = '';
    
    for (const line of lines) {
      if (/地点|位置|场所|场景|环境/.test(line)) {
        location = line.replace(/地点|位置|场所|场景|环境|[：:]/g, '').trim();
      } else if (/时间|时段|早晨|上午|中午|下午|傍晚|夜晚|深夜|季节|月份/.test(line)) {
        time = line.replace(/时间|[：:]/g, '').trim();
      } else if (/描述|介绍|概要|概况/.test(line)) {
        description += (description ? '；' : '') + line.replace(/描述|介绍|概要|概况|[：:]/g, '').trim();
      } else if (/氛围|气氛|感觉|情绪|氛围感/.test(line)) {
        atmosphere += (atmosphere ? '；' : '') + line.replace(/氛围|气氛|感觉|情绪|氛围感|[：:]/g, '').trim();
      } else if (/光线|照明|灯光|亮光|阴暗/.test(line)) {
        lighting += (lighting ? '；' : '') + line.replace(/光线|照明|灯光|亮光|阴暗|[：:]/g, '').trim();
      } else if (/声音|音效|背景音|音响/.test(line)) {
        sound += (sound ? '；' : '') + line.replace(/声音|音效|背景音|音响|[：:]/g, '').trim();
      } else if (!location && !time && !description && !atmosphere && !lighting && !sound) {
        // 如果还没有提取到任何信息，将第一行作为地点
        location = line;
      }
    }
    
    let formattedText = '';
    if (location) formattedText += `【场景地点】${location}\n`;
    if (time) formattedText += `【时间】${time}\n`;
    if (description) formattedText += `【场景描述】${description}\n`;
    if (atmosphere) formattedText += `【氛围】${atmosphere}\n`;
    if (lighting) formattedText += `【光线效果】${lighting}\n`;
    if (sound) formattedText += `【声音效果】${sound}\n`;
    
    if (!formattedText) {
      // 如果没有识别到关键词，则将原文本按行输出
      formattedText = lines.join('\n');
    }
    
    aiSceneOutput.value = formattedText.trim();
    ElMessage.warning("AI处理失败，使用本地格式化结果");
    
    // 解析本地格式化结果并填充到各个文本框
    const formattedLines = formattedText.trim().split('\n').map(line => line.trim()).filter(line => line);
    
    let localLocation = '';
    let localTime = '';
    let localDescription = '';
    let localAtmosphere = '';
    let localLighting = '';
    let localSound = '';
    
    for (const line of formattedLines) {
      if (line.includes('【场景地点】')) {
        localLocation = line.replace('【场景地点】', '').trim();
      } else if (line.includes('【时间】')) {
        localTime = line.replace('【时间】', '').trim();
      } else if (line.includes('【场景描述】')) {
        localDescription = line.replace('【场景描述】', '').trim();
      } else if (line.includes('【氛围】')) {
        localAtmosphere = line.replace('【氛围】', '').trim();
      } else if (line.includes('【光线效果】')) {
        localLighting = line.replace('【光线效果】', '').trim();
      } else if (line.includes('【声音效果】')) {
        localSound = line.replace('【声音效果】', '').trim();
      }
    }
    
    aiSceneLocation.value = localLocation || '';
    aiSceneTime.value = localTime || '';
    aiSceneDescription.value = localDescription || '';
    aiSceneAtmosphere.value = localAtmosphere || '';
    aiSceneLighting.value = localLighting || '';
    aiSceneSound.value = localSound || '';
  }
};

// 保存AI生成的场景信息到表单
const saveAiGeneratedScenes = () => {
  if (!aiSceneLocation.value.trim()) {
    ElMessage.warning("没有可保存的内容");
    return;
  }

  editingScene.value = null;
  newScene.value = {
    location: aiSceneLocation.value || '新场景',
    time: aiSceneTime.value || '',
    atmosphere: aiSceneAtmosphere.value || '',
    lighting: aiSceneLighting.value || '',
    sound: aiSceneSound.value || '',
    prompt: aiSceneDescription.value || '',
    image_url: "",
    local_path: "",
    reference_images: [],
    image_orientation: "horizontal",
  };

  // 关闭AI对话框，打开添加场景对话框
  aiGenerateSceneDialogVisible.value = false;
  addSceneDialogVisible.value = true;
};

const saveScene = async () => {
  if (!newScene.value.location.trim()) {
    ElMessage.warning("请输入场景名称");
    return;
  }

  syncSceneRefToModel();

  // 合并氛围、光线效果、声音效果到prompt中
  let combinedPrompt = newScene.value.prompt || '';
  if (newScene.value.atmosphere) {
    combinedPrompt += `\n氛围：${newScene.value.atmosphere}`;
  }
  if (newScene.value.lighting) {
    combinedPrompt += `\n光线效果：${newScene.value.lighting}`;
  }
  if (newScene.value.sound) {
    combinedPrompt += `\n声音效果：${newScene.value.sound}`;
  }

  try {
    if (editingScene.value) {
      // Update existing scene
      await dramaAPI.updateScene(editingScene.value.id, {
        location: newScene.value.location,
        time: newScene.value.time,
        description: combinedPrompt,
        image_url: newScene.value.image_url,
        local_path: newScene.value.local_path,
        reference_images: newScene.value.reference_images,
        image_orientation: newScene.value.image_orientation,
      });
    } else {
      // Create new scene
      await dramaAPI.createScene({
        drama_id: drama.value!.id,
        location: newScene.value.location,
        time: newScene.value.time,
        prompt: combinedPrompt,
        description: combinedPrompt,
        image_url: newScene.value.image_url,
        local_path: newScene.value.local_path,
        reference_images: newScene.value.reference_images,
        image_orientation: newScene.value.image_orientation,
      });
    }

    ElMessage.success(editingScene.value ? "场景更新成功" : "场景添加成功");
    addSceneDialogVisible.value = false;
    await loadDramaData();
  } catch (error: any) {
    ElMessage.error(error.message || "操作失败");
  }
};

const editScene = (scene: any) => {
  editingScene.value = scene;
  
  let referenceImages: string[] = [];
  if (scene.reference_images) {
    if (typeof scene.reference_images === 'string') {
      referenceImages = [scene.reference_images];
    } else if (Array.isArray(scene.reference_images)) {
      referenceImages = scene.reference_images.map((item: any) => {
        if (typeof item === 'string') {
          return item;
        } else if (item && item.url) {
          return item.url;
        }
        return '';
      }).filter((url: string) => url);
    }
  }
  
  newScene.value = {
    location: scene.location || scene.name || "",
    time: scene.time || "",
    atmosphere: scene.atmosphere || "",
    lighting: scene.lighting || "",
    sound: scene.sound || "",
    prompt: scene.prompt || scene.description || "",
    image_url: scene.image_url || "",
    local_path: scene.local_path || "",
    reference_images: referenceImages,
    image_orientation: scene.image_orientation || "horizontal",
  };
  
  sceneReferenceImages.value = referenceImages.map((url: string) => ({
    url: toDisplayUrl(url),
    name: url.split('/').pop()
  }));
  
  addSceneDialogVisible.value = true;
};

const deleteScene = async (scene: any) => {
  if (!scene.id) {
    ElMessage.error("场景ID不存在，无法删除");
    return;
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除场景"${scene.name || scene.location}"吗？此操作不可恢复。`,
      "删除确认",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      },
    );

    await dramaAPI.deleteScene(scene.id.toString());
    ElMessage.success("场景已删除");
    await loadScenes();
  } catch (error: any) {
    if (error !== "cancel") {
      console.error("删除场景失败:", error);
      ElMessage.error(error.message || "删除失败");
    }
  }
};

const openAddPropDialog = () => {
  editingProp.value = null;
  newProp.value = {
    name: "",
    description: "",
    prompt: "",
    type: "",
    image_url: "",
    local_path: "",
    reference_images: [],
    image_orientation: "horizontal",
  };
  propReferenceImages.value = [];
  addPropDialogVisible.value = true;
};

// 打开AI生成道具对话框
const openAiGeneratePropDialog = () => {
  aiPropInput.value = '';
  aiPropOutput.value = '';
  aiGeneratePropDialogVisible.value = true;
};

// 使用AI处理道具信息
const aiProcessProps = async () => {
  if (!aiPropInput.value.trim()) {
    ElMessage.warning("请输入要处理的文本");
    return;
  }

  try {
    // 构建专业的道具信息提取提示词
    const prompt = `请分析以下文本并提取其中的道具信息，按照以下标准格式返回：

【道具名称】道具的正式名称
【道具类型】道具的分类或类型，如：武器、日常用品、监控设备、控制设备等
【道具描述】道具的详细描述，按照以下结构组织：
  - 风格特点：道具的整体风格和设计理念
  - 外观描述：道具的外观、形状、样式等详细描述
  - 颜色：道具的主要颜色和光线效果
  - 材质：道具的制作材料（如果是全息投影等特殊材质请注明）
  - 尺寸规格：道具的大小、重量等物理规格，包括不同状态下的变化
  - 功能用途：道具的作用和使用方法，包括特殊功能
  - 整体设计特点：道具的设计特色和视觉效果
【图片提示词】用于生成道具图片的详细提示词，包含风格、画质、光影、细节等所有视觉元素

处理规则：
1. 如果原文本中已经包含上述标签，请提取对应的内容
2. 如果原文本中没有某个标签，但能根据文本内容合理推断，请生成相应内容
3. 如果某个标签信息不合理、无法推断或原文本中没有相关信息，请留空
4. 图片提示词应基于道具描述生成，包含完整的视觉元素描述
5. 请严格按照以上格式返回，不要添加额外说明

${aiPropInput.value}`;
    
    // 调用现有API进行文本处理 - 使用polish-prompt端点
    const response = await dramaAPI.polishPrompt({
      prompt: prompt,
      type: 'refine',
      orientation: 'horizontal',
      style: drama.value?.style || 'realistic'
    });
    
    aiPropOutput.value = response.polished_prompt;
    ElMessage.success("AI处理完成！");
    
    // 解析AI输出并填充到各个文本框
    const aiLines = response.polished_prompt.split('\n').map(line => line.trim()).filter(line => line);
    
    let aiName = '';
    let aiType = '';
    let aiDescription = '';
    let aiPrompt = '';
    
    let currentSection = '';
    let descriptionParts: string[] = [];
    
    for (const line of aiLines) {
      if (line.includes('【道具名称】')) {
        aiName = line.replace('【道具名称】', '').trim();
      } else if (line.includes('【道具类型】')) {
        aiType = line.replace('【道具类型】', '').trim();
      } else if (line.includes('【道具描述】')) {
        currentSection = 'description';
      } else if (line.includes('【图片提示词】')) {
        currentSection = 'prompt';
        aiPrompt = line.replace('【图片提示词】', '').trim();
      } else if (currentSection === 'description') {
        if (line.startsWith('-') || line.startsWith('•')) {
          descriptionParts.push(line);
        } else if (line.trim()) {
          descriptionParts.push(line);
        }
      } else if (currentSection === 'prompt' && line.trim()) {
        aiPrompt += ' ' + line.trim();
      }
    }
    
    aiPropName.value = aiName || '';
    aiPropType.value = aiType || '';
    aiPropDescription.value = descriptionParts.join('\n') || '';
    aiPropPrompt.value = aiPrompt || '';
  } catch (error: any) {
    // 如果AI调用失败，回退到本地格式化
    const lines = aiPropInput.value.split('\n')
      .map(line => line.trim())
      .filter(line => line);
    
    let name = '';
    let type = '';
    let style = '';
    let appearance = '';
    let color = '';
    let material = '';
    let size = '';
    let function_desc = '';
    let design_features = '';
    let image_prompt = '';
    
    for (const line of lines) {
      if (/道具|物品|东西|器具/.test(line)) {
        name = line.replace(/道具|物品|东西|器具|[：:]/g, '').trim();
      } else if (/类型|种类|类别|分类|品类/.test(line)) {
        type = line.replace(/类型|种类|类别|分类|品类|[：:]/g, '').trim();
      } else if (/风格|风格特点|设计理念/.test(line)) {
        style += (style ? '；' : '') + line.replace(/风格|风格特点|设计理念|[：:]/g, '').trim();
      } else if (/外观|外观描述|样子|外形/.test(line)) {
        appearance += (appearance ? '；' : '') + line.replace(/外观|外观描述|样子|外形|[：:]/g, '').trim();
      } else if (/颜色|色彩|色调|色/.test(line)) {
        color += (color ? '；' : '') + line.replace(/颜色|色彩|色调|色|[：:]/g, '').trim();
      } else if (/材质|材料|质地|面料|原料/.test(line)) {
        material += (material ? '；' : '') + line.replace(/材质|材料|质地|面料|原料|[：:]/g, '').trim();
      } else if (/大小|尺寸|规格|体积|长宽高/.test(line)) {
        size += (size ? '；' : '') + line.replace(/大小|尺寸|规格|体积|长宽高|[：:]/g, '').trim();
      } else if (/功能|作用|用途|使用|干什么/.test(line)) {
        function_desc += (function_desc ? '；' : '') + line.replace(/功能|作用|用途|使用|干什么|[：:]/g, '').trim();
      } else if (/设计特点|设计特色|视觉效果/.test(line)) {
        design_features += (design_features ? '；' : '') + line.replace(/设计特点|设计特色|视觉效果|[：:]/g, '').trim();
      } else if (/图片提示词|prompt|提示词/.test(line)) {
        image_prompt += (image_prompt ? '；' : '') + line.replace(/图片提示词|prompt|提示词|[：:]/g, '').trim();
      } else if (!name && !type && !style && !appearance && !color && !material && !size && !function_desc && !design_features && !image_prompt) {
        name = line;
      }
    }
    
    let formattedText = '';
    if (name) formattedText += `【道具名称】${name}\n`;
    if (type) formattedText += `【道具类型】${type}\n`;
    if (style || appearance || color || material || size || function_desc || design_features) {
      formattedText += `【道具描述】\n`;
      if (style) formattedText += `  - 风格特点：${style}\n`;
      if (appearance) formattedText += `  - 外观描述：${appearance}\n`;
      if (color) formattedText += `  - 颜色：${color}\n`;
      if (material) formattedText += `  - 材质：${material}\n`;
      if (size) formattedText += `  - 尺寸规格：${size}\n`;
      if (function_desc) formattedText += `  - 功能用途：${function_desc}\n`;
      if (design_features) formattedText += `  - 整体设计特点：${design_features}\n`;
    }
    if (image_prompt) formattedText += `【图片提示词】${image_prompt}\n`;
    
    if (!formattedText) {
      // 如果没有识别到关键词，则将原文本按行输出
      formattedText = lines.join('\n');
    }
    
    aiPropOutput.value = formattedText.trim();
    ElMessage.warning("AI处理失败，使用本地格式化结果");
    
    // 解析本地格式化结果并填充到各个文本框
    const formattedLines = formattedText.trim().split('\n').map(line => line.trim()).filter(line => line);
    
    let localName = '';
    let localType = '';
    let localDescription = '';
    let localPrompt = '';
    
    let currentSection = '';
    let descriptionParts: string[] = [];
    
    for (const line of formattedLines) {
      if (line.includes('【道具名称】')) {
        localName = line.replace('【道具名称】', '').trim();
      } else if (line.includes('【道具类型】')) {
        localType = line.replace('【道具类型】', '').trim();
      } else if (line.includes('【道具描述】')) {
        currentSection = 'description';
      } else if (line.includes('【图片提示词】')) {
        currentSection = 'prompt';
        localPrompt = line.replace('【图片提示词】', '').trim();
      } else if (currentSection === 'description') {
        if (line.startsWith('-') || line.startsWith('•')) {
          descriptionParts.push(line);
        } else if (line.trim()) {
          descriptionParts.push(line);
        }
      } else if (currentSection === 'prompt' && line.trim()) {
        localPrompt += ' ' + line.trim();
      }
    }
    
    aiPropName.value = localName || '';
    aiPropType.value = localType || '';
    aiPropDescription.value = descriptionParts.join('\n') || '';
    aiPropPrompt.value = localPrompt || '';
  }
};

// 保存AI生成的道具信息到表单
const saveAiGeneratedProps = () => {
  if (!aiPropName.value.trim()) {
    ElMessage.warning("没有可保存的内容");
    return;
  }

  editingProp.value = null;
  newProp.value = {
    name: aiPropName.value || '新道具',
    type: aiPropType.value || '',
    description: aiPropDescription.value || '',
    prompt: aiPropPrompt.value || '',
    image_url: "",
    local_path: "",
    reference_images: [],
    image_orientation: "horizontal",
  };

  // 关闭AI对话框，打开添加道具对话框
  aiGeneratePropDialogVisible.value = false;
  addPropDialogVisible.value = true;
};

const saveProp = async () => {
  if (!newProp.value.name.trim()) {
    ElMessage.warning("请输入道具名称");
    return;
  }

  syncPropRefToModel();

  try {
    const propData = {
      drama_id: drama.value!.id,
      name: newProp.value.name,
      description: newProp.value.description,
      prompt: newProp.value.prompt,
      type: newProp.value.type,
      image_url: newProp.value.image_url,
      local_path: newProp.value.local_path,
      reference_images: newProp.value.reference_images,
      image_orientation: newProp.value.image_orientation,
    };

    if (editingProp.value) {
      await propAPI.update(editingProp.value.id, propData);
      ElMessage.success("道具更新成功");
    } else {
      await propAPI.create(propData as any);
      ElMessage.success("道具添加成功");
    }

    addPropDialogVisible.value = false;
    await loadDramaData();
  } catch (error: any) {
    ElMessage.error(error.message || "操作失败");
  }
};

const editProp = (prop: any) => {
  editingProp.value = prop;
  
  let referenceImages: string[] = [];
  if (prop.reference_images) {
    if (typeof prop.reference_images === 'string') {
      referenceImages = [prop.reference_images];
    } else if (Array.isArray(prop.reference_images)) {
      referenceImages = prop.reference_images.map((item: any) => {
        if (typeof item === 'string') {
          return item;
        } else if (item && item.url) {
          return item.url;
        }
        return '';
      }).filter((url: string) => url);
    }
  }
  
  newProp.value = {
    name: prop.name,
    description: prop.description || "",
    prompt: prop.prompt || "",
    type: prop.type || "",
    image_url: prop.image_url || "",
    local_path: prop.local_path || "",
    reference_images: referenceImages,
    image_orientation: prop.image_orientation || "horizontal",
  };
  
  propReferenceImages.value = referenceImages.map((url: string) => ({
    url: toDisplayUrl(url),
    name: url.split('/').pop()
  }));
  
  addPropDialogVisible.value = true;
};

const deleteProp = async (prop: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除道具"${prop.name}"吗？此操作不可恢复。`,
      "删除确认",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      },
    );

    await propAPI.delete(prop.id);
    ElMessage.success("道具已删除");
    await loadDramaData();
  } catch (error: any) {
    if (error !== "cancel") {
      ElMessage.error(error.message || "删除失败");
    }
  }
};

const generatePropImage = async (prop: any) => {
  currentGenerateTarget.value = { type: 'prop', data: prop };
  imagePrompt.value = prop.prompt || prop.description || '';
  imageOrientation.value = prop.image_orientation || 'horizontal';
  baseReferenceImages.value = [];
  referenceImageList.value = toRefFileList(prop.reference_images);
  generateImageDialogVisible.value = true;
};

const handlePropImageSuccess = (response: any) => {
  if (response && response.data && response.data.url) {
    newProp.value.image_url = response.data.url;
    newProp.value.local_path = response.data.local_path || "";
  }
};

const handlePropReferenceImageSuccess = (response: any, file: any) => {
  if (response && response.data && response.data.url) {
    file.url = response.data.url;
    newProp.value.reference_images = propReferenceImages.value
      .map(f => f.url)
      .filter(url => url);
  }
};

const openExtractDialog = () => {
  extractPropsDialogVisible.value = true;
  if (sortedEpisodes.value.length > 0 && !selectedExtractEpisodeId.value) {
    selectedExtractEpisodeId.value = sortedEpisodes.value[0].id;
  }
};

const handleExtractProps = async () => {
  if (!selectedExtractEpisodeId.value) return;

  try {
    const res = await propAPI.extractFromScript(selectedExtractEpisodeId.value);
    extractPropsDialogVisible.value = false;

    // 自动刷新几次
    let checkCount = 0;
    const checkInterval = setInterval(() => {
      loadDramaData();
      checkCount++;
      if (checkCount > 10) clearInterval(checkInterval);
    }, 5000);
  } catch (error: any) {
    ElMessage.error(error.message || t("common.failed"));
  }
};

onMounted(async () => {
  await loadDramaData();

  // 如果有query参数指定tab，切换到对应tab
  if (route.query.tab) {
    activeTab.value = route.query.tab as string;
  }
});
</script>

<style scoped>
/* ========================================
   Page Layout / 页面布局 - 紧凑边距
   ======================================== */
.page-container {
  min-height: 100vh;
  background: var(--bg-primary);
  /* padding: var(--space-2) var(--space-3); */
  transition: background var(--transition-normal);
}

@media (min-width: 768px) {
  .page-container {
    /* padding: var(--space-3) var(--space-4); */
  }
}

@media (min-width: 1024px) {
  .page-container {
    /* padding: var(--space-4) var(--space-5); */
  }
}

.content-wrapper {
  margin: 0 auto;
  width: 100%;
}

/* ========================================
   Stats Grid / 统计网格 - 紧凑间距
   ======================================== */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  gap: var(--space-2);
  margin-bottom: var(--space-3);
}

@media (min-width: 640px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--space-3);
  }
}

@media (min-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

/* ========================================
   Tabs Wrapper / 标签页容器 - 紧凑内边距
   ======================================== */
.tabs-wrapper {
  background: var(--glass-bg-heavy);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-xl);
  padding: var(--space-3);
  box-shadow: var(--shadow-card);
}

@media (min-width: 768px) {
  .tabs-wrapper {
    padding: var(--space-4);
  }
}

/* ========================================
   Tab Header / 标签页头部
   ======================================== */
.tab-header {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
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
  letter-spacing: -0.01em;
}

/* ========================================
   Episode Sections / 章节分组
   ======================================== */
.episode-section {
  margin-top: 24px;
}
.episode-section:first-child {
  margin-top: 16px;
}
.episode-section-title {
  margin: 0 0 12px 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  padding-left: 10px;
  border-left: 3px solid var(--el-color-primary);
}

/* ========================================
   Character & Scene Cards / 角色场景卡片
   ======================================== */
.character-card,
.scene-card,
.scene-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.scene-list-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-bg-heavy);
  transition: all var(--transition-normal);
}

.scene-list-item:hover {
  border-color: var(--accent);
  box-shadow: var(--shadow-card-hover);
}

.scene-list-thumb {
  width: 100px;
  height: 68px;
  flex-shrink: 0;
  border-radius: 8px;
  overflow: hidden;
  background: var(--bg-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.scene-list-thumb :deep(img) {
  object-fit: cover;
}

.scene-list-info {
  flex: 1;
  min-width: 0;
}

.scene-list-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.scene-list-header h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.scene-list-desc {
  margin: 0;
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.scene-list-actions {
  flex-shrink: 0;
  display: flex;
  gap: 4px;
}

.prop-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.prop-list-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-bg-heavy);
  transition: all var(--transition-normal);
}

.prop-list-item:hover {
  border-color: var(--accent);
  box-shadow: var(--shadow-card-hover);
}

.prop-list-thumb {
  width: 80px;
  height: 60px;
  flex-shrink: 0;
  border-radius: 8px;
  overflow: hidden;
  background: var(--bg-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.prop-list-thumb :deep(img) {
  object-fit: contain;
}

.prop-list-info {
  flex: 1;
  min-width: 0;
}

.prop-list-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.prop-list-header h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.prop-list-desc {
  margin: 0;
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.prop-list-actions {
  flex-shrink: 0;
  display: flex;
  gap: 4px;
}

.prop-card {
  margin-bottom: var(--space-4);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-xl);
  overflow: hidden;
  transition: all var(--transition-normal);
  background: var(--glass-bg-heavy);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));
}

.character-card:hover,
.scene-card:hover,
.prop-card:hover {
  border-color: var(--accent);
  box-shadow: var(--shadow-card-hover), 0 0 16px rgba(14, 165, 233, 0.1);
  transform: translateY(-2px);
}

.character-card :deep(.el-card__body),
.scene-card :deep(.el-card__body),
.prop-card :deep(.el-card__body) {
  padding: 0;
}

.character-preview,
.scene-preview,
.prop-preview {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 120px;
  background: linear-gradient(135deg, var(--accent) 0%, #06b6d4 100%);
  overflow: hidden;
}

.character-preview img,
.scene-preview img,
.prop-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-normal);
}

.character-card:hover .character-preview img,
.scene-card:hover .scene-preview img,
.prop-card:hover .prop-preview img {
  transform: scale(1.05);
}

.scene-placeholder {
  color: rgba(255, 255, 255, 0.7);
}

/* ========================================
   Character list layout
   ======================================== */
.char-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
}

.char-card {
  border: 2px solid #d0d5dd;
  border-radius: 12px;
  background: #fff;
  overflow: hidden;
  transition: all var(--transition-normal);
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
}

.char-card:hover {
  border-color: var(--accent);
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
}

.char-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  border-bottom: 1px solid #eee;
}

.char-gallery {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 10px 12px;
}

.char-gallery-item {
  width: 200px;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px;
  border-radius: 8px;
  position: relative;
  cursor: pointer;
  transition: background 0.2s;
}

.char-gallery-item:hover {
  background: #f0f4f8;
}

.char-gallery-item :deep(.image-preview-wrapper) {
  width: 184px;
  height: 124px;
}

.char-gallery-item :deep(.thumbnail-container) {
  border-radius: 6px;
}

.char-gallery-item :deep(img) {
  object-fit: contain;
}

.char-gallery-base :deep(.image-preview-wrapper) {
  border: 2px solid #67c23a;
  border-radius: 8px;
}

.char-base-empty {
  width: 184px;
  height: 124px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  color: #bbb;
  font-size: 11px;
  border: 2px dashed #ddd;
  border-radius: 6px;
  transition: all 0.2s;
}

.char-gallery-base:hover .char-base-empty {
  border-color: var(--accent);
  color: var(--accent);
}

.char-gallery-label {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  font-weight: 500;
  color: var(--text-secondary);
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 184px;
}

.char-gallery-actions {
  display: flex;
  gap: 0;
  margin-top: 2px;
  opacity: 0;
  transition: opacity 0.2s;
}

.char-gallery-item:hover .char-gallery-actions {
  opacity: 1;
}

.ref-count-hint {
  font-size: 10px;
  color: #67c23a;
  margin-top: 2px;
}

.base-ref-preview {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.base-ref-item {
  position: relative;
  width: 148px;
  height: 148px;
  border-radius: 8px;
  overflow: hidden;
  border: 2px solid #67c23a;
}

.base-ref-item .el-image {
  width: 100%;
  height: 100%;
}

.base-ref-tag {
  position: absolute;
  bottom: 4px;
  left: 50%;
  transform: translateX(-50%);
  pointer-events: none;
}

.ref-download-btn {
  position: absolute;
  top: 4px;
  right: 4px;
  opacity: 0;
  transition: opacity 0.2s;
  z-index: 2;
  background: rgba(255,255,255,0.85) !important;
}

.base-ref-item:hover .ref-download-btn {
  opacity: 1;
}

.upload-file-card {
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
}

.upload-file-card img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.upload-file-card .el-upload-list__item-actions {
  position: absolute;
  width: 100%;
  height: 100%;
  left: 0;
  top: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  background: rgba(0,0,0,0.5);
  opacity: 0;
  transition: opacity 0.2s;
  cursor: default;
}

.upload-file-card:hover .el-upload-list__item-actions {
  opacity: 1;
}

.upload-file-card .el-upload-list__item-actions span {
  color: #fff;
  font-size: 18px;
  cursor: pointer;
}

.upload-file-card .el-upload-list__item-actions span:hover {
  color: #409eff;
}

.char-name-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  flex-wrap: wrap;
}

.char-name-row .el-tag {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 1;
}

.char-name-row h4 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
}

.char-actions {
  flex-shrink: 0;
  display: flex;
  gap: 4px;
}

.outfit-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 14px;
  border-bottom: 1px solid var(--glass-border);
  transition: background var(--transition-fast);
  background: var(--bg-secondary, #fafafa);
}

.outfit-item:last-child {
  border-bottom: none;
}

.outfit-item:hover {
  background: var(--glass-bg);
}

.outfit-thumb {
  width: 66px;
  height: 44px;
  flex-shrink: 0;
  border-radius: 6px;
  overflow: hidden;
  background: var(--bg-secondary);
}

.outfit-thumb img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
}

.outfit-info {
  flex: 1;
  min-width: 0;
}

.outfit-name {
  font-weight: 600;
  font-size: 13px;
  color: var(--text-primary);
  line-height: 1.3;
}

.outfit-desc {
  font-size: 11px;
  color: var(--text-secondary);
  margin: 1px 0 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.outfit-actions {
  display: flex;
  gap: 2px;
  flex-shrink: 0;
}

.outfit-actions .el-button {
  font-size: 12px;
  padding: 2px 6px;
}

.character-info,
.scene-info,
.prop-info {
  text-align: center;
  padding: var(--space-4);
}

.character-name {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: var(--space-2);
}

.character-info h4,
.scene-info h4,
.prop-info h4 {
  /* margin: 0 0 var(--space-2) 0; */
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.desc {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: var(--space-2) 0;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
}

.character-actions,
.scene-actions,
.prop-actions {
  display: flex;
  gap: var(--space-2);
  justify-content: center;
  padding: 0 var(--space-4) var(--space-4);
}

.empty-icon {
  color: var(--accent);
}

/* ========================================
   Dark Mode / 深色模式
   ======================================== */
.dark .tabs-wrapper {
  background: var(--glass-bg-heavy);
}

.dark :deep(.el-card) {
  background: var(--glass-bg-heavy);
  border-color: var(--glass-border);
}

.dark :deep(.el-card__header) {
  background: var(--bg-secondary);
  border-color: var(--glass-border);
}

.dark :deep(.el-table) {
  background: var(--bg-card);
  --el-table-bg-color: var(--bg-card);
  --el-table-tr-bg-color: var(--bg-card);
  --el-table-header-bg-color: var(--bg-secondary);
  --el-fill-color-lighter: var(--bg-secondary);
}

.dark :deep(.el-table th),
.dark :deep(.el-table tr) {
  background: var(--bg-card);
}

.dark :deep(.el-table td),
.dark :deep(.el-table th) {
  border-color: var(--border-primary);
}

.dark :deep(.el-table--striped .el-table__body tr.el-table__row--striped td) {
  background: var(--bg-secondary);
}

.dark :deep(.el-table__body tr:hover > td) {
  background: var(--bg-card-hover) !important;
}

.dark :deep(.el-descriptions) {
  background: var(--bg-card);
}

.dark :deep(.el-descriptions__label) {
  background: var(--bg-secondary);
  color: var(--text-secondary);
  border-color: var(--border-primary);
}

.dark :deep(.el-descriptions__content) {
  background: var(--bg-card);
  color: var(--text-primary);
  border-color: var(--border-primary);
}

.dark :deep(.el-descriptions__cell) {
  border-color: var(--border-primary);
}

/* ========================================
   Project Info Card / 项目信息卡片
   ======================================== */
.project-info-card {
  margin-top: var(--space-5);
  border-radius: var(--radius-lg);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.project-descriptions {
  width: 100%;
}

:deep(.project-descriptions .el-descriptions__label) {
  width: 120px;
  font-weight: 500;
  color: var(--text-secondary);
}

:deep(.project-descriptions .el-descriptions__content) {
  min-width: 150px;
}

.info-value {
  font-weight: 500;
  color: var(--text-primary);
}

.info-desc {
  color: var(--text-secondary);
  line-height: 1.6;
}

.dark :deep(.el-dialog) {
  background: var(--glass-bg-heavy);
}

.dark :deep(.el-dialog__header) {
  background: transparent;
}

.dark :deep(.el-form-item__label) {
  color: var(--text-primary);
}

.dark :deep(.el-input__wrapper) {
  background: var(--bg-secondary);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
}

.dark :deep(.el-input__inner) {
  color: var(--text-primary);
}

.dark :deep(.el-textarea__inner) {
  background: var(--bg-secondary);
  color: var(--text-primary);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
}

.avatar-wrapper:hover .avatar-overlay {
  opacity: 1;
}
</style>

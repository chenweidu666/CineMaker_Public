# Changelog

## [v1.1.0] - 2026-02-20

### 用户认证与团队协作（阶段一完成）

#### 新增

- **用户系统**：注册、登录、JWT 认证（access_token + refresh_token，bcrypt 密码加密）
- **团队管理**：创建团队、邀请成员、移除成员、团队信息编辑
- **数据隔离**：`dramas`、`characters`、`ai_service_configs`、`assets` 等核心表加 `team_id`，查询自动按团队过滤
- **认证中间件**：Gin 中间件从 Authorization Header 解析 JWT，注入 userID/teamID/role 到 Context
- **前端路由守卫**：未登录自动跳转 `/login`，Token 过期自动尝试刷新
- **登录/注册页**：`Login.vue`、`Register.vue`，含表单验证
- **团队管理页**：`TeamManagement.vue`，成员列表、邀请、移除
- **AppHeader 用户菜单**：右上角显示用户名，下拉菜单含团队管理和退出
- **首次部署自动初始化**：`migrateDefaultTeam` 创建默认团队和管理员，已有数据自动关联 `team_id`

#### 改进

- **参考图系统**：场景/道具支持参考图导入；生成图片弹窗正确显示参考图预览（添加 `/static/` 前缀）
- **图片懒加载**：全局 `el-image` 添加 `lazy` 属性，角色/场景/道具/分镜列表加载性能提升
- **图片生成 Debug 模式**：生成图片弹窗新增 Debug 按钮，展示完整的 API 请求 curl 命令
- **Drama 更新 API 修复**：`UpdateDrama` 返回更新后的最新数据，修复风格修改后 API 响应返回旧值的问题

#### 涉及新增文件

- 后端：`api/handlers/auth_handler.go`、`team_handler.go`、`api/middlewares/auth.go`、`application/services/auth_service.go`、`domain/models/user.go`、`pkg/auth/context.go`、`pkg/auth/scope.go`
- 前端：`views/auth/Login.vue`、`Register.vue`、`views/team/TeamManagement.vue`、`stores/user.ts`、`api/auth.ts`、`api/team.ts`

---

## [v1.0.0] - 2026-02-18

### 首个正式版本

CineMaker AI 短剧制作平台 v1.0.0，核心功能包括：

- **项目管理**：短剧创建/编辑/删除，11种视觉风格
- **角色管理**：手动创建/AI从剧本提取/角色库/AI生成形象图/批量生成
- **场景管理**：手动创建/AI提取/AI生成场景图/提示词润色
- **道具管理**：手动创建/AI从剧本提取/AI生成图片
- **剧本设计**：编写剧本/AI重写(对白/动作)/AI提取角色场景
- **分镜生成**：AI从剧本自动拆分分镜/镜头属性编辑/帧提示词生成
- **图片生成**：单张/批量/角色/场景/道具/分镜/首帧/尾帧/宫格图
- **视频生成**：单镜头/从图片生成/批量/链式生成（支持 runway/pika/doubao/openai）
- **视频合成**：FFmpeg 合并分镜片段为完整集视频，20+种转场效果
- **简单剪辑**：时间线拖拽/片段裁剪/转场设置/音频提取
- **AI配置**：文本/图片/视频多供应商配置、测试连接
- **系统监控**：仪表盘/任务统计/API调用/资源使用
- **Docker 部署**：多阶段构建 Dockerfile + docker-compose，一键启动

---

## [v1.0.1] - 2026-02-18

### 清理

#### 移除

- **ScriptEdit 页面**：删除 `web/src/views/script/ScriptEdit.vue` 及路由 `/episodes/:id/edit`
  - 原因：占位页面，功能已在 EpisodeWorkflow 步骤2（剧本设计）中实现
- **StoryboardEdit 页面**：删除 `web/src/views/storyboard/StoryboardEdit.vue` 及路由 `/episodes/:id/storyboard`
  - 原因：占位页面，功能已在 ProfessionalEditor 中实现
- **语言切换功能**：
  - 删除后端 `api/handlers/settings.go`（GetLanguage/UpdateLanguage）
  - 删除后端路由 `/api/v1/settings/language`
  - 删除前端 `web/src/api/settings.ts`（settingsAPI）
  - 删除前端 `web/src/views/settings/SystemSettings.vue`（语言切换 UI）
  - 清理 `web/src/locales/index.ts` 中无用的 setLanguage/getCurrentLanguage
  - 清理 `web/src/locales/zh-CN.ts` 中语言切换相关的 i18n key
  - 清理 `DramaManagement.spec.ts` 中的过期组件引用
  - 原因：系统固定使用中文，语言切换功能已禁用且无英文翻译文件
- **工作流"专业制作"步骤指示器**：从 EpisodeWorkflow 步骤条中移除第4步，保留3步工作流（资源定义→剧本设计→分镜生成）
  - 原因：专业制作有独立页面和入口，不属于工作流步骤

#### 变更

- **start.sh**：重写为基于 docker compose 的管理脚本，适配 NAS 部署
  - 支持命令：start/stop/restart/rebuild/status/logs/update

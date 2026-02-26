<p align="center">
  <img src="https://github.com/chenweidu666.png" alt="CineMaker Logo" width="140" />
</p>

<p align="center">
  <h1>🎬 CineMaker</h1>
  <strong>用 AI 一键生成短剧</strong> · 从剧本到成片的全流程自动化制作
</p>

<p align="center">
  <a href="https://furalike.cn/cinemaker/"><strong>在线体验 Demo</strong></a> ·
  <a href="https://furalike.cn/blog/posts/cinemaker-workflow/"><strong>工作流程详解</strong></a> ·
  <a href="https://www.xiaohongshu.com/discovery/item/6999f9390000000015031bfa?source=webshare&xhsshare=pc_web&xsec_token=AB1YeBV8FFISOWlMopLLgorOWiW83maPlho8zWFMxMoZU=&xsec_source=pc_share"><strong>查看成片效果</strong></a>
</p>

<p align="center">
  <img src="./docs/guides/images/11-professional-editor.png" alt="CineMaker 专业编辑器效果展示" width="860" />
</p>

> 基于 [火宝短剧（huobao-drama）](https://github.com/chatfire-AI/huobao-drama) 二次开发

---

# 1. 为什么选择 CineMaker

**传统短剧制作**：编剧 → 分镜 → 拍摄 → 剪辑 → 后期，耗时数周，成本数万元

**CineMaker**：输入剧本 → AI 生成 → 视频成片，**10 分钟完成一集短剧**

---

# 2. 相对 huobao-drama 的核心优化与改进

## 2.1 核心功能增强

| 优化项 | 说明 |
|--------|------|
| **角色多造型系统** | 同一角色支持多套外貌造型，分镜级别自动匹配参考图，解决角色一致性问题 |
| **两阶段分镜拆分** | 先整体方案再逐镜头细化，三段描述（首帧/中间过程/尾帧）分别服务图片和视频模型 |
| **参考图驱动生成** | 角色三视图、场景四宫格自动注入生成流程，确保画面连贯性和角色一致性 |
| **图片微调编辑器** | 快捷指令（修手部/调表情/改光影等）+ 自然语言局部重绘，所见即所得 |
| **专业视频编辑器** | 时间线拖拽、20+ 转场效果、片段裁剪、键盘快捷键，媲美剪映的专业体验 |
| **团队协作支持** | JWT 认证、多团队数据隔离、成员管理，支持多人协同创作 |
| **AI 日志与可观测性** | 完整记录 System/User Prompt、模型参数、响应结果与耗时，便于定位问题和持续优化 |

## 2.2 提示词工程优化

| 优化项 | 说明 |
|--------|------|
| **角色三视图生成** | 优化提示词模板，确保角色正面、侧面、背面的一致性 |
| **风格映射表** | 支持 11 种视觉风格，自动映射到提示词，确保风格统一 |
| **appearance 描述规范** | 统一角色描述格式，提高 AI 生成质量 |
| **AI 智能生成提示词** | 优化提示词生成逻辑，提高生成准确性和效率 |

## 2.3 用户体验改进

| 优化项 | 说明 |
|--------|------|
| **本地存储默认** | 默认使用本地存储，无需配置云服务，开箱即用 |
| **Docker 部署优化** | 支持国内镜像加速，5 分钟快速部署 |
| **系统监控仪表盘** | 实时监控 AI 服务状态和任务进度 |
| **图片懒加载** | 优化图片加载性能，提升用户体验 |

---

# 3. 核心特性

| 特性 | 说明 |
|------|------|
| 🎭 **AI 全流程驱动** | 从剧本创作到视频成片，全流程 AI 自动化，无需手动拍摄 |
| 🎨 **角色多造型系统** | 同一角色支持多套外貌造型，分镜级别自动匹配参考图 |
| 📐 **两阶段分镜拆分** | 先整体方案再逐镜头细化，确保画面连贯性和一致性 |
| 🎯 **参考图驱动生成** | 角色三视图、场景四宫格自动注入生成流程，保证角色一致性 |
| 🖼️ **图片微调编辑器** | 快捷指令修手部/调表情 + 自然语言局部重绘，所见即所得 |
| 🎬 **专业视频编辑器** | 时间线拖拽、20+ 转场效果、键盘快捷键，媲美剪映的专业体验 |
| 👥 **团队协作支持** | JWT 认证、多团队数据隔离、成员管理，支持多人协同创作 |
| 💾 **开箱即用** | 默认本地存储，无需配置云服务，5 分钟快速部署 |

---

# 4. 成品展示

## 4.1 短剧成片
- **[姜小卷的周一：普通又认真的打工人日常](https://www.xiaohongshu.com/discovery/item/6999f9390000000015031bfa?source=webshare&xhsshare=pc_web&xsec_token=AB1YeBV8FFISOWlMopLLgorOWiW83maPlho8zWFMxMoZU=&xsec_source=pc_share)**（小红书）

## 4.2 PV 角色短片
- **[叶澜｜把心事都写进旋律里的吟游诗人](docs/剧本/AI女性Vlog/设计/PV/叶澜｜把心事都写进旋律里的吟游诗人.mp4)**
- **[夏紫萱｜穿军装驾机甲的二次元少女](docs/剧本/AI女性Vlog/设计/PV/夏紫萱｜穿军装驾机甲的二次元少女.mp4)**
- **[梅暗香｜穿裙子也能一脚踢翻你](docs/剧本/AI女性Vlog/设计/PV/梅暗香｜穿裙子也能一脚踢翻你.mp4)**
- **[温晚｜在音乐里自由落体的慵懒舞者](docs/剧本/AI女性Vlog/设计/PV/温晚｜在音乐里自由落体的慵懒舞者.mp4)**
- **[陈樱｜在吧台后面看尽人间故事](docs/剧本/AI女性Vlog/设计/PV/陈樱｜在吧台后面看尽人间故事.mp4)**
- **[顾秋雅｜用鼻子写诗的人](docs/剧本/AI女性Vlog/设计/PV/顾秋雅｜用鼻子写诗的人.mp4)**

---

# 5. 技术栈

| 层级 | 技术 |
|------|------|
| **后端** | Go 1.25, Gin, GORM, SQLite |
| **前端** | Vue 3, TypeScript, Vite, Element Plus |
| **AI 服务** | 火山引擎（Seedream 图片 / Seedance 视频）、OpenAI 兼容 API |
| **存储** | 本地文件系统（默认） / 腾讯云 COS（可选） |
| **部署** | Docker, Docker Compose |

---

# 6. 快速开始

## 6.1 Docker 部署（推荐）

### 6.1.1 启动与访问

```bash
git clone <repo-url> && cd cinemaker
docker compose up -d
docker compose logs -f
```

**启动后访问**：http://localhost:5678

### 6.1.2 国内加速（可选）

<details>
<summary>📦 国内加速（点击展开）</summary>

```bash
DOCKER_REGISTRY=registry.cn-hangzhou.aliyuncs.com/library/ \
NPM_REGISTRY=https://registry.npmmirror.com \
GO_PROXY=https://goproxy.cn,direct \
ALPINE_MIRROR=mirrors.aliyun.com \
docker compose up -d --build
```

</details>

## 6.2 本地开发

### 6.2.1 开发命令

```bash
./start.sh dev        # Docker 热加载模式（推荐）
./start.sh logs       # 查看后端日志
./start.sh stop       # 停止服务
```

- 后端：http://localhost:5678（Go + air 热加载）
- 前端：http://localhost:3012（Vite HMR）

---

# 7. 文档指南

| 文档 | 说明 |
|------|------|
| 📊 [产品工作流](docs/guides/1_产品工作流.md) | **大纲**：短剧管理首页 → 项目概览 → 角色管理 → 分镜生成 → 图片生成 → 视频合成。以 AI 女性 Vlog《姜小卷的周一》为例，图文讲解从资源准备到成片的全流程 |
| 🏗️ [技术架构](docs/guides/2_技术架构.md) | **大纲**：系统架构总览（技术栈、架构图）→ 核心模块实现（AI 集成、数据模型、提示词工程）→ 部署方案（Docker、本地开发）。面向开发者的技术方案详解 |
| ✍️ [提示词指南](docs/guides/3_提示词指南.md) | **大纲**：Seedream 4.0 助力 Seedance 生视频最佳实践（多图融合、首帧/首尾帧/多参考图生视频）→ Seedance 1.5 Pro 提示词参数与技巧 → SeedEdit 3.0 图片编辑指南 → **CineMaker 提示词工程（核心）**：角色三视图生成、风格映射表、appearance 描述规范、AI 智能生成提示词 |
| 🔑 [火山引擎 API 申请指南](docs/guides/4_火山引擎API申请指南.md) | **大纲**：注册火山引擎账号 → 申请 API 密钥（火山方舟）→ 开通模型（Doubao、Seedream、Seededit、Seedance）→ 在 CineMaker 中配置 → 验证配置 |

---

# 8. 已实现功能

| 模块 | 功能亮点 |
|------|---------|
| **项目管理** | 短剧创建/编辑/删除，11 种视觉风格，多集章节管理 |
| **角色管理** | 手动创建 / AI 提取 / 多造型 / 三视图自动生成 |
| **场景管理** | 四宫格设定图自动生成，参考图自动匹配 |
| **剧本设计** | AI 生成剧本、AI 辅助重写、自动提取角色/场景/道具 |
| **分镜生成** | 两阶段拆分、三段描述、按时长约束对话量、景别/运镜/转场编辑 |
| **图片生成** | 首帧/尾帧/关键帧/分镜板、参考图自动注入、AI 提示词生成 |
| **图片微调** | 快捷指令（修手部/调表情/改光影等）、自然语言局部重绘 |
| **视频生成** | 单图/首尾帧/多图参考/纯文本模式、AI 生成视频提示词 |
| **视频合成** | FFmpeg 合并、20+ 转场效果、时间线拖拽、片段裁剪 |

| **团队协作** | JWT 认证、多团队数据隔离、成员管理 |
| **系统功能** | AI 服务多供应商配置、系统监控仪表盘、Debug 模式 |

---

# 9. 费用参考

以一集 13 个镜头、1 分 30 秒的短剧为例：

| 操作 | 单次费用 | 总计 |
|------|---------|------|
| 生成剧本 / 分镜拆分 | < 0.1 元 | ~1 元 |
| 生成图片（角色/场景/首帧/尾帧） | 0.3 - 0.5 元 | ~5 元 |
| 图片微调 | 0.2 - 0.3 元 | ~2 元 |
| 生成视频（5-10 秒） | 1 - 2 元 | ~10 元 |
| **一集完整短剧**（13 镜头） | - | **约 20 元** |

> 💡 火山引擎新用户赠送 **50 万 tokens** 免费额度，足够体验多次！

---

# 10. 待开发功能

| 功能 | 说明 |
|------|------|
| 团队级 API Key 管理 | 每个团队独立管理 AI 服务密钥 |
| 独立语音合成（TTS） | 按角色分别合成语音，自动导入时间线 |
| 开发/部署环境分离 | 同一服务器上独立的开发和生产环境 |

---

# 11. 目录结构

```
├── api/                  # HTTP 路由和处理器
│   ├── handlers/         # 请求处理器（23 个）
│   ├── routes/           # 路由注册
│   └── middlewares/      # 中间件（JWT 认证、CORS、限流）
├── application/          # 业务逻辑服务层
│   └── services/         # 业务服务（33 个）
├── domain/               # 领域模型
│   ├── models/           # GORM 数据库模型（14 个）
│   └── errors.go         # 领域错误定义
├── infrastructure/       # 基础设施层
│   ├── database/         # GORM + SQLite 连接
│   ├── storage/          # 存储实现（Local / COS）
│   └── external/ffmpeg/  # FFmpeg 封装
├── pkg/                  # 共享包
│   ├── ai/               # LLM 文本生成客户端
│   ├── asr/              # 语音相关客户端
│   ├── auth/             # 认证上下文
│   ├── config/           # 配置加载
│   ├── image/            # 图片生成客户端
│   └── video/            # 视频生成客户端
├── configs/              # 配置文件
├── migrations/           # 数据库迁移 SQL
├── docs/                 # 文档
├── web/                  # Vue 3 前端
└── main.go               # 入口
```

---

# 12. 联系方式

- **作者邮箱**：s514351508@gmail.com
- **GitHub**：[@chenweidu666](https://github.com/chenweidu666)

---

# 13. 开源说明（Public 版）

- 默认仅使用本地存储路径（`./data/storage`），不依赖腾讯云 COS
- 运行前请复制 `.env.example` 为 `.env` 并按需填写 AI 服务配置
- 仓库不包含真实业务数据库与敏感凭据
- 欢迎提交 Issue 和 PR，共同完善 CineMaker！

---

# 14. 致谢

- 基于 [火宝短剧（huobao-drama）](https://github.com/chatfire-AI/huobao-drama) 二次开发
- 使用 [火山引擎](https://www.volcengine.com/) 提供的 AI 服务

---

<p align="center">
  <strong>用 AI 重新定义短剧制作</strong> 🎬
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go" alt="Go Version" />
  <img src="https://img.shields.io/badge/Vue-3-4FC08D?logo=vuedotjs" alt="Vue Version" />
  <img src="https://img.shields.io/badge/License-MIT-007BFF" alt="License" />
</p>

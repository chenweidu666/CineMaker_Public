
<p align="center">
  <h1>🎬 CineMaker</h1>
  <h3>下一代 AI 驱动短剧创作平台</h3>
  <p><strong>从单句创意到完整成片，仅需一次点击 · 全流程智能化生产</strong></p>
</p>

<p align="center">
  <a href="https://furalike.cn/cinemaker/"><strong>立即体验 Demo</strong></a>　·　
  <a href="https://furalike.cn/blog/posts/cinemaker-workflow/"><strong>完整工作流解析</strong></a>　·　
  <a href="https://www.xiaohongshu.com/discovery/item/6999f9390000000015031bfa"><strong>真实成片案例</strong></a>
</p>

<p align="center">
  <img src="./docs/guides/images/11-professional-editor.png" alt="CineMaker Professional Editing Interface" width="860" style="border-radius: 12px; box-shadow: 0 8px 32px rgba(0,0,0,0.15);" />
  <br><small>专业级时间线编辑器 · 像素级精准控制</small>
</p>

> 基于 [火宝短剧](https://github.com/chatfire-AI/huobao-drama) 的深度重构与商业级升级

## 重新定义短剧创作效率

**传统路径**：数周周期 · 数万成本 · 多工种协作  
**CineMaker**：10分钟内生成完整一集 · 平均成本仅约 ¥20 · 单人即可掌控电影级品质

## 相较原始 huobao-drama 的代际跃升

> 不只看文字，直接看功能截图（每项都对应实际页面）。

### 1）角色一致性：多造型 + 三视图参考
<p>
  <img src="./docs/guides/images/03-character-management.png" alt="角色多造型与三视图管理" width="860" />
</p>

### 2）分镜智能拆分：两阶段 + 首中尾三段式
<p>
  <img src="./docs/guides/images/10-storyboard-generation.png" alt="分镜生成与三段式描述" width="860" />
</p>

### 3）视觉控制力：21:9 场景设定图强制注入
<p>
  <img src="./docs/guides/images/05-scene-management.png" alt="场景库与21:9场景设定图控制" width="860" />
</p>

### 4）创作迭代效率：图片微调编辑器
<p>
  <img src="./docs/guides/images/13-image-fine-tuning.png" alt="图片微调与局部重绘" width="860" />
</p>

### 5）后期生产力：专业时间线编辑器
<p>
  <img src="./docs/guides/images/11-professional-editor.png" alt="专业视频编辑器与时间线" width="860" />
</p>

### 6）可观测性：全链路 AI 日志
<p>
  <img src="./docs/guides/images/06-ai-logs.png" alt="AI 调用日志与可观测性" width="860" />
</p>

## 旗舰特性一览

- 🎭 **端到端 AI 驱动** —— 剧本创作 → 分镜 → 视觉生成 → 视频合成全链路自动化
- 🎨 **工业级角色一致性引擎** —— 多造型 + 三视图 + 动态参考图匹配
- 📐 **两阶段分镜智能拆分** —— 全局叙事规划 + 逐镜头三段精描
- 🖼️ **参考图闭环控制** —— 角色三视图 / 21:9 场景设定图 / 关键帧全流程强制一致
- ✂️ **电影级后期工具** —— 专业时间线 + 海量转场 + 键盘加速工作流
- 👥 **企业级协作架构** —— 多团队 / 权限隔离 / 成员管理
- ⚡ **极致易部署** —— Docker 一键启动，默认本地存储，开箱即用

## 代表性成品

- **真实短剧**：[姜小卷的周一 —— 打工人真实写照](https://www.xiaohongshu.com/discovery/item/6999f9390000000015031bfa)
- **本地样片（仓库内）**：[第一集演示视频（压缩版）](docs/guides/images/episode-1_compressed.mov)
- **角色 PV 短片**（高品质示范）：
  - [叶澜 · 吟游诗人](docs/剧本/AI女性Vlog/设计/PV/叶澜｜把心事都写进旋律里的吟游诗人.mp4)
  - [夏紫萱 · 机甲战姬](docs/剧本/AI女性Vlog/设计/PV/夏紫萱｜穿军装驾机甲的二次元少女.mp4)
  - [梅暗香 · 优雅暴力美学](docs/剧本/AI女性Vlog/设计/PV/梅暗香｜穿裙子也能一脚踢翻你.mp4)
  - [梅暗香（版本2）](docs/剧本/AI女性Vlog/设计/PV/梅暗香｜穿裙子也能一脚踢翻你-1.mp4)
  - [温晚 · 慵懒自由舞者](docs/剧本/AI女性Vlog/设计/PV/温晚｜在音乐里自由落体的慵懒舞者.mp4)
  - [陈樱 · 深夜故事聆听者](docs/剧本/AI女性Vlog/设计/PV/陈樱｜在吧台后面看尽人间故事.mp4)
  - [顾秋雅 · 嗅觉诗人](docs/剧本/AI女性Vlog/设计/PV/顾秋雅｜用鼻子写诗的人.mp4)

## 现代技术架构

| 层级       | 技术选型                              |
|------------|---------------------------------------|
| Backend    | Go 1.25 · Gin · GORM · SQLite         |
| Frontend   | Vue 3 + TypeScript · Vite · Element Plus |
| AI Engine  | 火山引擎 Seedream / Seedance / SeedEdit + OpenAI 兼容接口 |
| Storage    | Local FS（默认） / 腾讯云 COS（可选） |
| Deployment | Docker Compose · 国内镜像加速支持     |

## 快速启动（推荐 Docker）

```bash
git clone https://github.com/chenweidu666/cinemaker && cd cinemaker
docker compose up -d
# 国内加速启动（可选）
DOCKER_REGISTRY=registry.cn-hangzhou.aliyuncs.com/library/ \
NPM_REGISTRY=https://registry.npmmirror.com \
GO_PROXY=https://goproxy.cn,direct \
docker compose up -d --build
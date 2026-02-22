# CineMaker Demo

> AI 短剧生成系统的交互式演示页面

### 成品效果

**[姜小卷的周一，普通又认真的打工人日常 - 小红书](https://www.xiaohongshu.com/discovery/item/6999f9390000000015031bfa?source=webshare&xhsshare=pc_web&xsec_token=AB1YeBV8FFISOWlMopLLgorOWiW83maPlho8zWFMxMoZU=&xsec_source=pc_share)**

以上视频完全由 CineMaker 系统生成（AI 绘图 + AI 视频 + 自动合成），点击可查看最终发布效果。

---

**在线体验编辑器：[https://furalike.cn/cinemaker/](https://furalike.cn/cinemaker/)**

## 说明

这是 [CineMaker](https://github.com/chenweidu666/CineMaker) 的**功能展示 Demo**，用于演示系统的核心工作流程。

**注意：** 此 Demo 为纯前端静态页面，**不包含数据库和 AI 生成功能**。所有数据均为硬编码的 mock 数据，"生成图片"和"生成视频"按钮仅模拟交互效果，不会调用任何 AI API。

## 展示内容

Demo 使用了第一集「周一又是元气满满的一天」的真实生产数据，包括：

- **项目管理** — 项目列表、角色管理（3 套造型）、场景管理（10 个场景）
- **分镜编辑器** — 13 个分镜的三栏布局编辑器
  - 左侧：分镜列表，缩略图 + 标题快速切换
  - 中间：首帧/尾帧图片预览 + 视频懒加载播放（点击播放按钮才加载视频，节省流量）
  - 右侧：镜头属性（场景、角色、三段描述）、镜头画面（首帧/尾帧分页 + 参考图）、视频生成（首尾帧对比 + 参数配置）
- **AI 生成模拟** — 图片生成进度条、视频生成进度动画

所有图片和视频资源托管在腾讯云 COS。

## 技术栈

- Vue 3 + Vite
- Element Plus + Tailwind CSS
- Vue Router（Hash 模式）
- 纯静态部署，无后端依赖

## 本地运行

```bash
npm install
npm run dev
```

## 构建

```bash
npm run build
```

构建产物在 `dist/` 目录，可直接部署到任何静态文件服务器。

## 完整项目

完整的 CineMaker 系统（包含 Go 后端、数据库、AI API 集成）见：[CineMaker](https://github.com/chenweidu666/CineMaker)（私有仓库）

## License

MIT

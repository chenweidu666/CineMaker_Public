# Contributing

感谢你对 CineMaker 的关注与贡献。

## 开发环境

- Backend: Go 1.25+
- Frontend: Node 22+, npm
- 默认使用本地存储路径：`./data/storage`

## 本地启动

```bash
# 后端
go run main.go

# 前端
cd web
npm install
npm run dev
```

## 提交规范

- 提交信息建议使用：`feat: ...` / `fix: ...` / `docs: ...` / `refactor: ...`
- 一次提交只做一类改动，避免混杂
- 不要提交密钥、Token、数据库文件、真实业务素材

## Pull Request 要求

1. 描述改动背景与目标
2. 列出主要改动点
3. 提供测试说明（本地如何验证）
4. 如涉及 UI，附截图或录屏

## 安全与隐私

- 请勿提交 `.env`、`*.db`、`cookies*.txt`、私有 URL 和凭据
- 发现安全问题请私下联系维护者，不要公开披露漏洞细节

## 联系方式

有问题或建议可联系：s514351508@gmail.com

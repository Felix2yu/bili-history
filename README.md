# BiliHistory

Bilibili 观看历史记录分析工具，合并自以下两个项目：

- **前端 (Vue 3)**：[BiliHistoryFrontend](https://github.com/LifeArchiveProject/BiliHistoryFrontend)
- **后端 (Python FastAPI)**：[BilibiliHistoryFetcher](https://github.com/LifeArchiveProject/BilibiliHistoryFetcher)

## 目录结构

```
bili-history/
├── frontend/     # Vue 3 + Vite + Tauri 前端
├── backend/      # Python FastAPI 后端
├── LICENSE       # MIT 许可证
├── NOTICE        # 原始项目归属声明
└── README.md
```

## 快速开始

### 后端

```bash
cd backend
pip install -r requirements.txt
python main.py
```

服务默认运行在 `http://localhost:8899`，API 文档见 `http://localhost:8899/docs`。

### 前端

```bash
cd frontend
npm install
npm run dev
```

开发服务器运行在 `http://localhost:5173`。

### Docker 部署

```bash
docker-compose up -d
```

## 许可证

本项目采用 [MIT 许可证](LICENSE)，详见 [NOTICE](NOTICE) 文件中的原始项目归属声明。

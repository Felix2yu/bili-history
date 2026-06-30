# BiliHistory

Bilibili 观看历史记录分析工具，合并自以下两个项目：

- **前端 (Nuxt 3 + Vue 3)**：[BiliHistoryFrontend](https://github.com/LifeArchiveProject/BiliHistoryFrontend)
- **后端 (Python FastAPI)**：[BilibiliHistoryFetcher](https://github.com/LifeArchiveProject/BilibiliHistoryFetcher)

## 目录结构

```
bili-history/
├── frontend/     # Nuxt 3 + Vue 3 前端（支持 SSR）
├── backend/      # Python FastAPI 后端
├── config/       # 配置文件目录（Docker 挂载）
├── docker-compose.yml
├── LICENSE       # MIT 许可证
├── NOTICE        # 原始项目归属声明
└── README.md
```

## 快速开始

### Docker 部署（推荐）

```bash
docker-compose up -d
```

- 前端：`http://localhost:3000`
- 后端 API：`http://localhost:8899`
- API 文档：`http://localhost:8899/docs`

### 后端

```bash
cd backend
pip install -r requirements.txt
python main.py
```

### 前端

```bash
cd frontend
npm install
npm run dev
```

## 架构说明

### SSR（服务端渲染）

前端使用 Nuxt 3 的 SSR 能力，在服务端预取初始数据，提升首屏加载速度：

- **HistoryContent**：登录状态 + 历史记录列表
- **Favorites**：收藏夹列表
- **WatchLater**：稍后再看列表
- **Downloads**：下载列表
- **Images**：图片状态
- **MyLikes**：点赞列表
- **SchedulerTasks**：计划任务列表
- **Search**：搜索结果

### 非 SSR 组件

以下组件因技术限制（DOM 依赖、浏览器 API）保持客户端渲染：

- **Analytics 页面**（20+ 个）：ECharts 图表可视化
- **VideoPlayer**：ArtPlayer 视频播放器 + 弹幕系统
- **PWA**：Service Worker 离线缓存
- **客户端状态管理**：Pinia stores

### 图片懒加载

稍后再看等列表页面使用 IntersectionObserver 实现真正的图片懒加载，仅在图片进入视口 200px 范围内才开始加载。

## 许可证

本项目采用 [MIT 许可证](LICENSE)，详见 [NOTICE](NOTICE) 文件中的原始项目归属声明。

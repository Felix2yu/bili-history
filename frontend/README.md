<div align="center">
  <img src="./public/logo.png" alt="Logo">
</div>

# BiliHistory Frontend

基于 Nuxt 3 + Vue 3 的 B 站历史记录分析前端，支持 SSR 服务端渲染。

## 技术栈

- **框架**：Nuxt 3 (Vue 3)
- **UI**：Tailwind CSS + Vant
- **图表**：ECharts
- **视频播放**：ArtPlayer + 弹幕插件
- **状态管理**：Pinia
- **PWA**：@vite-pwa/nuxt

## 快速开始

```bash
npm install
npm run dev
```

开发服务器运行在 `http://localhost:3000`。

## SSR 架构

### 服务端预取的数据

以下组件使用 `useAsyncData` 在服务端预取初始数据：

| 组件 | 预取内容 |
|------|----------|
| HistoryContent | 登录状态、历史记录列表、分类 |
| Favorites | 收藏夹列表 |
| WatchLater | 稍后再看列表 |
| Downloads | 下载列表 |
| Images | 图片下载状态 |
| MyLikes | 点赞列表 |
| SchedulerTasks | 计划任务列表 |
| Search | 搜索结果 |

### 客户端渲染的组件

- Analytics 页面（ECharts 图表）
- VideoPlayer（ArtPlayer + 弹幕）
- PWA Service Worker
- Pinia 状态管理

## 图片优化

- **懒加载**：使用 IntersectionObserver，视口 200px 范围内才加载
- **HTTPS 升级**：HTTP 图片 URL 自动升级为 HTTPS（避免混合内容警告）

## 部署

### Docker

```bash
docker build -t bili-history-frontend .
docker run -p 3000:3000 bili-history-frontend
```

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `NUXT_BACKEND_URL` | 后端 API 地址 | `http://localhost:8899` |
| `NUXT_PUBLIC_DEFAULT_BACKEND_URL` | 前端代理路径 | `/api` |

## 项目结构

```
frontend/
├── pages/              # Nuxt 页面路由
├── components/         # Vue 组件
│   ├── page/          # 页面级组件
│   ├── analytics/     # 分析页面
│   ├── scheduler/     # 计划任务组件
│   └── layout/        # 布局组件
├── composables/        # Vue 组合式函数
├── layouts/           # 页面布局
├── plugins/           # Nuxt 插件
├── server/            # Nuxt 服务端
│   └── api/           # API 代理
├── stores/            # Pinia 状态
├── utils/             # 工具函数
├── public/            # 静态资源
└── nuxt.config.ts     # Nuxt 配置
```

# 拾帧集 (BiliHistory)

Bilibili 观看历史记录管理与分析工具

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](#docker-部署)
[![Release](https://img.shields.io/github/v/release/Felix2yu/bili-history?label=Release)](https://github.com/Felix2yu/bili-history/releases)

## 功能特性

### 数据概览
- **年度概览**：核心统计卡片、常看分区/UP主排行、设备分布、标题热词、时长偏好、周内分布、最长观看视频、观看热力图
- **月度报告**：按月查看观看记录，包含 12 个分析维度（重刷、完播率、时长偏好、深夜占比、收藏率等）
- **周度报告**：按周查看观看记录与统计（ISO 8601 标准，周一起始）

### 年度分析
- 时间分析（24 小时分布、最活跃时段、深夜观看）
- 月度趋势、连续观看记录
- 视频完成率分析、UP主完成率分析
- 标签分析、视频时长分析
- 点赞 / 收藏 / 稍后再看分析

### 历史记录管理
- 每天一页 + 日期导航，日历控件快速切换
- 网格视图 / 列表视图切换
- 图片懒加载（IntersectionObserver）
- 高级筛选：分区、UP主、完播状态

### 收藏夹 / 稍后再看 / 点赞
- 在线收藏夹内容查看（本地优先 + 在线回退）
- 子收藏夹（season/collection）内容浏览
- 视频点赞切换

### 视频播放与下载
- ArtPlayer 播放器 + 弹幕系统
- bili-dl 视频下载集成（支持批量下载、UP主视频下载）

### 计划任务与通知
- 定时抓取历史记录，链式任务编排
- 每日报告推送（Shoutrrr 通知）
- 连续失败自动暂停 + 告警

### MCP 服务
- MCP Streamable HTTP 协议，AI 客户端可读取本地历史记录
- Bearer Token 认证

### 其他
- PWA 支持（离线缓存）
- 深色模式（跟随系统 / 浅色 / 深色）
- 字体大小设置（小 / 中 / 大三档）
- 数据导出 Excel、SQLite 数据库下载
- Docker 部署支持

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Nuxt 3 (SSG 静态化) + Vue 3 + Vant 4 + Tailwind CSS + ECharts |
| 后端 | Go + Gin + SQLite |
| 部署 | Docker 单容器 (前后端一体) + docker-compose |

## 快速开始

### Docker 部署（推荐）

```bash
docker compose up -d
```

访问：`http://localhost:8899`（前后端统一端口，API 走 `/api/*` 前缀）

### 本地开发

#### 后端

```bash
cd backend
go run ./cmd/main.go
```

#### 前端

```bash
cd frontend
pnpm install
pnpm dev
```

> 开发模式下前端默认通过代理将 `/api` 请求转发到后端 `localhost:8899`

## 目录结构

```
bili-history/
├── frontend/          # Nuxt 3 (SSG) + Vue 3 前端
│   ├── components/    # Vue 组件
│   ├── pages/         # Nuxt 路由页面
│   ├── utils/         # 工具函数
│   └── stores/        # Pinia 状态管理
├── backend/           # Go Gin 后端
│   ├── cmd/           # 入口
│   ├── database/      # 数据库操作
│   ├── routers/       # API 路由
│   ├── services/      # 业务逻辑
│   ├── scheduler/     # 定时任务
│   └── web/           # 前端静态资源嵌入 (Go embed)
├── Dockerfile         # 单容器多阶段构建
├── docker-compose.yml
├── LICENSE            # MIT 许可证
├── NOTICE             # 原始项目归属声明
└── CHANGELOG.md       # 变更日志
```

## 架构说明

### 单容器架构（前后端一体）

项目采用单容器部署，前端通过 Nuxt SSG 静态化后由 Go 后端直接提供服务：

```
浏览器 → Go Gin (单端口 8899)
           ├── /api/*       → 后端业务 API
           ├── /mcp         → MCP Streamable HTTP 服务
           ├── /health      → 健康检查
           └── /*           → 前端静态页面 (SPA, Go embed)
```

**核心设计：**
- **前端 SSG 静态化**：Nuxt 3 `ssr: false`，构建产出纯静态 HTML/CSS/JS
- **Go embed 嵌入**：前端产物通过 `//go:embed` 嵌入 Go 二进制，无需额外文件服务
- **单端口统一入口**：页面访问与 API 调用同源，无需配置 CORS 或 API 服务器地址
- **SPA 路由回退**：所有非 API/MCP 路径均返回 `index.html`，由 Vue Router 处理前端路由

### 图片懒加载

使用 IntersectionObserver 实现真正的图片懒加载，仅在图片进入视口 200px 范围内才开始加载。

## 许可证

本项目采用 [MIT 许可证](LICENSE)，详见 [NOTICE](NOTICE) 文件中的原始项目归属声明。

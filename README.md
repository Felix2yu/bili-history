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
- 每日报告推送（Apprise 通知）
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
├── cmd/bili-history/    # 入口 main.go + version.go
├── internal/            # Go 私有业务包
│   ├── backup/          # `bili-history backup` 多库快照导出
│   ├── biliapi/         # B 站 API 封装
│   ├── config/          # 配置加载
│   ├── database/        # 数据库操作
│   ├── models/          # 数据模型
│   ├── routers/         # API 路由
│   ├── scheduler/       # 定时任务
│   ├── services/        # 业务逻辑
│   ├── utils/           # 工具函数
│   └── web/             # 前端产物嵌入 (go:embed)
├── web/                 # Nuxt 3 (SSG) + Vue 3 前端
│   ├── components/      # Vue 组件
│   ├── pages/           # Nuxt 路由页面
│   ├── utils/           # 工具函数
│   └── stores/          # Pinia 状态管理
├── Dockerfile           # CI 运行时镜像（二进制预编译拼装）
├── docker-compose.yml
├── LICENSE              # MIT 许可证
├── NOTICE               # 原始项目归属声明
└── CHANGELOG.md         # 变更日志
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

## 备份与恢复

### 这个应用有六个数据库

拾帧集不是一个库一个应用。数据按功能分落在多个 SQLite 文件里：

| 库 | 位置（容器内） | 内容 |
|----|----------------|------|
| `bilibili_history.db` | `/app/output/` | 观看历史（按年分表）、分区分类 |
| `bilibili_dynamic.db` | `/app/output/database/` | 动态与动态主 |
| `bilibili_likes.db` | `/app/output/database/` | 点赞 |
| `bilibili_watchlater.db` | `/app/output/database/` | 稍后再看 |
| `bilibili_favorites.db` | `/app/output/database/` | 收藏夹与收藏内容 |
| `scheduler.db` | `/app/output/database/` | 计划任务与执行历史 |

主库的文件名来自配置 `db_file`（可改），其余几个是代码里写死的。备份时按实际目录**运行时发现**，
所以改过 `db_file` 或以后新增库都不需要改备份配置。

单个 `.db` 装不下这个应用，因此 `backup` 子命令的**默认产物是 zip**：
把上面每个库各做一份一致性快照，打进同一个归档，包内平铺、条目名就是裸文件名。

### 备份

```bash
docker exec <容器名> /app/bili-history backup > bili-history-$(date +%F).zip
```

只要一个 `.db`（**有损**：只有 `db_file` 那一个库，动态 / 点赞 / 稍后再看 / 收藏 / 调度器都不在包里），
可以显式指定格式，日常备份不建议用它：

```bash
docker exec <容器名> /app/bili-history backup --format db > bilibili_history-$(date +%F).db
```

实现要点：

- 每个库用 `VACUUM INTO` 由 SQLite 自己在事务里导出，**不热拷贝** `.db` / `-wal` / `-shm`
  —— 活跃写入下热拷贝会周期性产出损坏副本。
- 导出进程只读地打开既有的库：不建目录、不跑迁移、不写种子；
  库文件不存在就报错，**绝不产出一份空数据库**。
- 主库缺席 → 退出码非零、stdout 一个字节都没有；
  可选库缺席 → stderr 上告警并跳过，归档照样有效（一个部署可能压根没用过收藏夹）。
- 日志一律走 stderr，stdout 只承载产物字节，所以 `>` 重定向出来的就是干净的包。
- 与 `docker-db-auto-backup` 的对接见 `docker-compose.yml`：
  `BACKUP_SQLITE_CMD=/app/bili-history backup` + `BACKUP_SQLITE_EXT=zip`
  （命令写绝对路径：镜像把二进制放在 `/app`，而 `/app` 不在容器 PATH 上）。

### 一致性范围：每个库各自一致，跨库不是同一个时间点

SQLite 的 `VACUUM INTO` 保证的是**单个库内**的事务级自洽。六个库是六个互不相干的功能库
（历史、动态、点赞…），不是一个关系型整体，因此导出的是「每个库各自某一瞬间的一致快照」，
而不是「全应用同一个时间点的一致快照」—— 后者用这种方式做不到，对这份数据也不需要。
跨库对账（例如「收藏列表与历史记录在同一时刻的状态」）本来就没有事务边界，恢复后以各自为准即可。

### 恢复

包内是平铺的裸文件名，放回磁盘时按上表分两处：主库进 `output/`，其余进 `output/database/`。
主库那个条目叫什么取决于你的 `db_file`（默认 `bilibili_history.db`），下面的例子用默认名。

```bash
# 1. 停应用（避免恢复时主进程还在写同一批库）
docker compose stop bili-history

# 2. 先把旧数据留底，别覆盖掉还能救的东西
mv ./output ./output.bak-$(date +%F)

# 3. 放回数据目录：包内是平铺的，所以要按上面的表分两处落
ZIP=~/bili-history-YYYY-MM-DD.zip
TMP=$(mktemp -d)
unzip -o "$ZIP" -d "$TMP"
mkdir -p ./output/database
mv "$TMP/bilibili_history.db" ./output/                    # 主库 → output/
for f in bilibili_dynamic.db bilibili_likes.db bilibili_watchlater.db \
         bilibili_favorites.db scheduler.db; do
  [ -f "$TMP/$f" ] && mv "$TMP/$f" ./output/database/      # 其余 → output/database/
done
rmdir "$TMP"

# 4. 属主交给应用（PUID/PGID 默认 1000），别让 root 属主的文件挡住应用写自己的库
sudo chown -R 1000:1000 ./output

# 5. 起应用
docker compose up -d bili-history
```

包里只会是你实际有的库——某个可选库当初没被备份（比如这个部署没用过收藏夹），
上面那条 `mv` 会因为 `[ -f ]` 跳过，恢复照样成立。

`VACUUM INTO` 的产物是 delete 日志模式，应用启动会按自身逻辑切回 WAL，旁生的 `-wal`/`-shm`
不需要人工准备。

### 备份里**没有**什么

- **`config.yaml` 不在备份内**（刻意如此）。它挂在 `/app/config`，里面有 B 站 Cookie
  （`SESSDATA` / `bili_jct` / `DedeUserID`）与通知渠道 URL。它小、由宿主机管理、且带机密，
  放进自动备份的归档里等于把会话令牌抄送一份到备份目录。要备份配置请自己按需拷贝，
  并注意它的权限。
- **日志**（`output/logs/`）不备份——那是排障现场，不是数据。
- **图片与下载物**（`output/images/`、`output/downloads/`、`output/dynamic/<mid>/images/`）
  当前**不**备份，请注意这里有一处真实的取舍：
  - `output/images/` 是历史封面 / 头像的本地缓存，库里存的是远端 URL，丢了能重新拉；
  - `output/downloads/` 是 bili-dl 主动下载的视频，体积大且可按 BVID 重新下载；
  - **但 `output/dynamic/<mid>/images/` 不完全是缓存**：动态抓取成功后，库里的
    `cover` / `media_locals` 会被改写成指向本地的相对路径，只有 `raw_json` 里还留着原始远端地址。
    B 站原帖一旦删除，这些图就取不回来了。

  它们都不进包，原因是备份工具把 stdout 当产物管道（`docker exec … > 文件`），
  塞进数 GB 媒体不适合这个通道，而工具的 full 模式本就是「一个容器一个文件」。
  **如果你把动态配图当作必须保的内容，请单独同步 `output/dynamic/`（以及 `output/images/`、
  `output/downloads/`），例如用通用的目录备份而不是本应用的 backup 子命令。**
  要我把它们也纳入 zip 形态（像 mujian / qiansi 把 uploads 打进包那样）也可以说一声。

## 许可证

本项目采用 [MIT 许可证](LICENSE)，详见 [NOTICE](NOTICE) 文件中的原始项目归属声明。

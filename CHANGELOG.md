# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/lang/zh-CN/).

## [Unreleased]

### 新增

- 下载支持按 DASH 画质/音质 id 精确选流（含 dolby 与 flac 无损音轨），可选编码优先级
- 分 P 下载：单 P 指定或整稿全量展开
- 合集与番剧：合集按 `ugc_season` 真实展开，番剧支持 `ep`/`ss` 链接与整季下载
- 附加内容开关：封面（沿用原有开关）、弹幕、字幕，默认只下封面
- 下载可取消：断开 SSE 即中止底层请求与 ffmpeg 子进程；并发下载请求被明确拒绝
- 字节级下载进度（百分比、实时速度、当前资源类型）替代原先的纯文本阶段提示
- 限速选项（透传 bili-dl 的 `-rate` 语义）
- 收藏夹整夹下载接口 `/download/download_favorites`：前端早已在调用但后端没有注册，
  现在补上，本地收藏库优先、未同步时在线拉取
- 弹幕读取接口 `/download/stream_danmaku`：同样是前端一直在调却没有实现，
  现在按视频成品路径解析 `标题_BV.danmaku.ass`（也可按 BV 号检索），
  路径收敛在下载目录内；播放器不再用前端正则猜 sidecar 名与容器扩展名

### 变更

- 下载全流程交给 bili-dl 的新接口：不再由本仓库自行请求 playurl（此前缺少 WBI 签名）、
  自行选流、自行二次转封装；容器由实际编码决定（AV1→MKV、HEVC→MOV、其余→MP4）
- 产物命名改为 `下载目录/UP主/标题_BV号[ - 01 - 分P名].扩展名`，库列表、已下载判定、
  删除均按 BV 号关联；历史扁平 `[UP主]标题` 命名继续兼容解析
- `check_video_download` 与 `delete_downloaded_video` 的参数由 CID 改为 BV 号——
  按 CID 匹配文件名的判定此前从未命中过任何文件
- 下载事件流改为 JSON 事件（`log` / `progress` / `done`），`done` 携带成功/失败/取消状态

### 修复

- `stream_video` 会把 downloads 之外的任意本机路径读出来，现收敛在下载目录内并按容器返回 MIME
- 删除「同时删除整个目录」时可能连整个 UP 主目录一起删掉，现只删属于该 BV 的文件、目录清空后才移除
- 画质文案按长边取档，1080p 被显示成 2K、720p 被显示成 1080p，现按短边（B 站语义）计算

## [1.0.0] - 2026-07-26

首个正式发布版本。Bilibili 观看历史记录管理与分析工具。

### 架构

- **单容器部署**：前后端合并为单容器单端口（8899），前端 Nuxt SSG 静态化后通过 Go embed 嵌入后端二进制
- 所有后端 API 路由统一 `/api` 前缀，MCP 服务保持 `/mcp` 路径
- 3 阶段多架构 Docker 构建（前端 bun → Go 编译 embed → Alpine 运行）
- 镜像：`ghcr.io/felix2yu/bili-history`

### 数据概览

- 年度概览：核心统计卡片、常看分区/UP 主排行、设备分布、标题热词、时长偏好、周内分布、观看热力图
- 月度报告：12 个分析维度（重刷、完播率、时长偏好、深夜占比、收藏率、弃看率、黄金时段集中度等）
- 周度报告：ISO 8601 标准（周一起始）

### 年度分析

- 时间分析（24 小时分布、最活跃时段、深夜观看）
- 月度趋势、连续观看记录
- 视频/UP 主完成率分析、标签分析、视频时长分析
- 点赞 / 收藏 / 稍后再看分析

### 历史记录管理

- 每天一页 + 日期导航，日历控件快速切换
- 网格视图 / 列表视图切换
- 图片懒加载（IntersectionObserver）
- 高级筛选：分区、UP 主、完播状态

### 收藏夹 / 稍后再看 / 点赞

- 在线收藏夹内容查看（本地优先 + 在线回退）
- 子收藏夹（season/collection）内容浏览
- 视频点赞切换

### 视频播放与下载

- ArtPlayer 播放器 + 弹幕系统
- bili-dl 视频下载（支持批量下载、UP 主视频下载）
- 动态下载

### 计划任务与通知

- 定时抓取历史记录，链式任务编排
- 每日报告推送（Apprise 通知）
- 连续失败自动暂停 + 告警

### MCP 服务

- MCP Streamable HTTP 协议，AI 客户端可读取本地历史记录
- Bearer Token 认证

### 设置

- 深色模式（跟随系统 / 浅色 / 深色）
- 主题色切换（8 种预设）
- 字体大小设置（小 / 中 / 大三档，全局 rem 缩放）
- 侧边栏显示控制
- 本地图片源（离线访问）
- 数据完整性校验
- 同步已删除记录 / 同步删除 B 站历史记录
- MCP URL 自动使用当前访问地址，兼容反向代理

### 数据管理

- 数据导出 Excel（按年/月/日期范围）
- SQLite 数据库下载
- 数据库重置

### 其他

- PWA 支持（Service Worker 离线缓存）
- Docker 单容器部署（docker-compose）
- Docker 容器支持 PUID/PGID 非 root 运行
- 响应式布局（移动端 + 桌面端）

### 修复

- 修复年度观看时长统计：使用实际观看进度替代视频总时长累加
- 修复 Tailwind fontSize 和 CSS 硬编码 px 值不跟随字体大小设置缩放
- 修复设置页面打开时错误显示"同步已删除记录"通知
- 修复常看 UP 主数据为空（SQL 查询 GROUP BY 条件修正）
- 修复周数计算错误（改为 ISO 8601 周一起始）
- 修复翻页时重复显示相同内容的竞态问题
- 修复首次加载时闪现"暂无历史记录"
- 修复 RewatchStats / CompletionStats 类型冲突
- 修复柱状图高度计算问题
- 修复日期范围解析时区问题
- 修复高级筛选分区类型判断
- 修复 Docker 链式任务依赖顺序
- 修复每日报告空内容问题
- 修复封面图 HTTPS 升级避免不安全内容警告
- 全面修复 B 站 CDN 头像防盗链问题

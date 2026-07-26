# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/lang/zh-CN/).

## [1.1.0] - 2026-07-26

### Changed

#### 架构简化：单容器前后端一体化
- 前端从 Nuxt SSR 改为 **SSG 静态化** (`ssr: false`)，构建产出纯静态 SPA
- 前后端合并为**单容器单端口**部署，统一通过 `8899` 端口访问
- 前端静态资源通过 Go `embed` 嵌入后端二进制，无需额外文件服务
- 所有后端 API 路由统一加 `/api` 前缀，MCP 服务保持 `/mcp` 路径
- 移除前端独立 Dockerfile、Caddyfile、server API 代理等双容器相关文件
- CI 工作流从双镜像构建合并为单镜像构建
- 镜像名统一为 `ghcr.io/felix2yu/bili-history`

#### 部署配置
- `docker-compose.yml` 从 2 个服务（frontend + backend）简化为 1 个服务
- 数据卷映射路径保持不变 (`./backend/config:/app/config`、`./output:/app/output`)，与旧架构完全兼容
- 不再需要 `NUXT_*` 系列环境变量

#### 设置页面
- 移除"API 服务器地址"配置项（单容器架构下前端与后端同源，无需配置）
- MCP URL 显示改为自动使用当前页面的 origin

### Added

- 新增 `backend/web/` 目录存放 Go embed 相关代码与前端构建产物
- 新增 `Dockerfile`（根目录）：3 阶段多架构构建（前端 bun 构建 → Go 编译 embed → Alpine 运行）

### Removed

- 删除前端 `server/api/` 目录（API 代理层）
- 删除前端 `server/middleware/` 目录
- 删除前端 `plugins/*.server.ts`（SSR 专用插件）
- 删除前端独立 `Dockerfile`、`entrypoint.sh`、`deploy/Caddyfile`
- 删除前后端独立 CI 工作流文件

## [1.0.0] - 2026-07-12

### Added

#### 数据概览（原"周报月报" + "年度总结"合并）
- 数据概览页面整合周报月报与年度总结，统一入口
- 进入时弹窗选择概览类型（年度 / 月度 / 周度），点击即生效
- 年度概览新增：核心统计卡片、常看分区排行、常看UP主、设备分布、标题热词、时长偏好（短/中/长）、周内分布、最长观看视频
- 后端新增设备分布、标题热词提取、时长偏好、最长观看视频、周内分布等数据分析
- 周度计算采用 ISO 8601 标准（周一开始，周日结束）

#### 周报月报
- 新增周报/月报功能，按周/月维度展示观看记录及统计
- 报告页新增 12 个分析维度：重刷视频、完播率分布、时长偏好、深夜占比、收藏率、弃看率、黄金时段集中度、UP主多样性、活跃时段、标题热词、周内分布、最长观看
- 每日观看分布柱状图 + 24 小时时段分布柱状图
- 常看分区排行 + 常看UP主 + 设备分布

#### 年度总结（Analytics）
- 年度观看数据总览 + 观看热力图
- 时间分析：24 小时观看分布、最活跃时段、深夜观看统计
- 时间分布：周度分布 + 季节性趋势
- 月度观看趋势折线图
- 连续观看记录（最长连续 + 当前连续）
- 最爱重温视频排行
- 视频完成率分析（完成率分布 + 时长分布完成率）
- UP主完成率分析（最喜欢的UP主 + 观看最多的UP主）
- 标签分析（标签分布 + 标签完成率）
- 视频时长分析
- 点赞 / 收藏 / 稍后再看分析

#### 历史记录
- 每天一页 + 日期导航，替换传统分页
- 日历控件支持点击年月快速切换
- 网格视图 / 列表视图切换
- 图片 IntersectionObserver 懒加载
- 高级筛选：分区、UP主、完播状态

#### 收藏夹
- 支持在线收藏夹内容查看（本地优先 + 在线回退）
- 支持收藏夹子收藏夹（season/collection）内容浏览
- 收藏夹封面自动补全

#### 稍后再看 / 点赞
- 稍后再看列表管理
- 点赞列表管理 + 视频点赞切换

#### 视频功能
- ArtPlayer 视频播放器 + 弹幕系统
- 视频下载（bili-dl 集成，支持批量下载、UP主视频下载）
- 动态下载

#### 计划任务
- 定时抓取历史记录
- 链式任务编排（抓取 → 导入 → 分析 → 报告）
- 每日报告推送（Shoutrrr 通知集成）
- 连续抓取失败自动暂停 + 告警通知

#### MCP 服务
- MCP Streamable HTTP 协议支持
- AI 客户端可通过 MCP 读取本地历史记录
- Bearer Token 认证

#### 设置
- 深色模式（跟随系统 / 浅色 / 深色）
- 侧边栏显示控制
- 本地图片源（离线访问）
- 数据完整性校验
- 同步已删除记录
- 同步删除 B 站历史记录
- API 服务器地址配置

#### 数据管理
- 数据导出 Excel（按年/月/日期范围）
- SQLite 数据库下载
- 数据库重置

#### 其他
- PWA 支持（Service Worker 离线缓存）
- Docker 部署支持（docker-compose）
- Docker 容器支持 PUID/PGID 非 root 运行
- SSR 服务端渲染（历史记录、收藏夹、稍后再看等页面）
- 响应式布局（移动端 + 桌面端）

### Fixed
- 修复设置页面打开时错误显示"同步已删除记录"通知
- 修复常看UP主数据为空（SQL 查询 GROUP BY 条件修正）
- 修复周数计算错误（改为 ISO 8601 周一起始）
- 修复翻页时重复显示相同内容的竞态问题
- 修复首次加载时闪现"暂无历史记录"
- 修复年度总结页面 SSR 500 错误
- 修复 RewatchStats 类型冲突
- 修复 CompletionStats 类型冲突
- 修复柱状图高度计算问题
- 修复日期范围解析时区问题
- 修复高级筛选分区类型判断
- 修复 Docker 链式任务依赖顺序
- 修复每日报告空内容问题
- 修复封面图 HTTPS 升级避免不安全内容警告

### Changed
- "周报月报"改名为"数据概览"，合并"年度总结"入口
- "周概览"改名为"周度"
- 移除年度概览中收藏/点赞/稍后再看分析的平均播放量和平均弹幕数
- 热力图改为 HTML 实现（移除 PNG 生成）
- 报告页视频列表改为分页加载（初始 30 个，点击加载更多）
- 图片懒加载从 `loading="lazy"` 升级为 IntersectionObserver 实现
- 项目包管理器从 npm 切换为 bun
- Docker 后端镜像改为本地构建

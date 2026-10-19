# syntax=docker/dockerfile:1.7

# ===== 仅拼装的运行时镜像（CI 专用） =====
# 二进制由 CI 预编译：SQLite 已换为纯 Go 驱动（modernc.org/sqlite），
# CGO_ENABLED=0 出完全静态二进制，可交叉编译、与运行环境的 libc 无关。
# 这里没有任何 Node/Go 编译步骤，前端产物已通过 go:embed 嵌入二进制。
#
# 基镜 ghcr.io/felix2yu/base:alpine-3.24 提供 tzdata、ca-certificates、su-exec
# 与 PUID/PGID 降权入口；本仓库有两个绑定挂载的数据目录 /app/config 与 /app/output。
FROM ghcr.io/felix2yu/base:alpine-3.24

# ffmpeg：视频转码，基镜不含，按需追加（运行期就要用，不能用 --virtual 装完即删）
RUN --mount=type=cache,target=/var/cache/apk,sharing=locked \
    apk add --no-cache ffmpeg

ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    TZ=Asia/Shanghai \
    SESSDATA="" \
    BILI_JCT="" \
    DEDE_USER_ID="" \
    DEDE_USER_ID_CKMD5="" \
    SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8899 \
    APP_DATA_DIR="/app/config /app/output"

WORKDIR /app

# 本仓库的容器布局：二进制在 /app，数据仍在 /app/config 与 /app/output
# （两处绑定挂载的多目录应用，保留现状）。
# 外部备份工具（docker-db-auto-backup）经 `docker exec <容器> sh -c "/app/bili-history backup"`
# 采集快照：`/app` 不在容器 PATH 上，所以 `sh -c` 里必须写绝对路径，
# 工具按镜像名推导出的裸名（bili-history backup）在本镜像里解析不到。
# 工作目录仍是 /app，路径解析逻辑不受影响。
COPY --chmod=755 bin/bili-history /app/bili-history
COPY internal/config /app/config

RUN mkdir -p /app/output

EXPOSE 8899

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8899/health >/dev/null || exit 1

# 原入口用 tini + su-exec；基镜的入口直接 exec 到 PID 1，Go 的 os/exec 会回收子进程，
# 不再需要额外的 init。
CMD ["/app/bili-history"]

# syntax=docker/dockerfile:1.7

# ===== 仅拼装的运行时镜像（CI 专用） =====
# 二进制由 workflow 的 build job 在原生 runner 上用 musl 静态预编译
# （amd64 / arm64 矩阵），经 artifact 下载到 ./bin 后直接 COPY 进镜像。
# 这里没有任何 Node/Go 编译步骤，前端产物已通过 go:embed 嵌入二进制。

FROM alpine:3.24

ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8
ENV TZ=Asia/Shanghai

ENV SESSDATA=""
ENV BILI_JCT=""
ENV DEDE_USER_ID=""
ENV DEDE_USER_ID_CKMD5=""
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8899

RUN --mount=type=cache,target=/var/cache/apk,sharing=locked \
    apk add --no-cache \
      tzdata \
      ca-certificates \
      tini \
      ffmpeg \
      shadow \
      su-exec \
      curl

WORKDIR /app

COPY bin/bili-history /app/bili-history
COPY backend/config /app/config
COPY entrypoint.sh /entrypoint.sh

RUN mkdir -p /app/output && \
    chmod +x /entrypoint.sh /app/bili-history

EXPOSE 8899

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD curl -fsS http://localhost:8899/health || exit 1

ENTRYPOINT ["/sbin/tini", "--", "/entrypoint.sh"]

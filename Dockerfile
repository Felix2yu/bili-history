# syntax=docker/dockerfile:1.7

# ===== 阶段 1: 构建前端 =====
FROM node:22-slim AS frontend-builder
WORKDIR /app/frontend
ENV CI=true
RUN corepack enable && corepack prepare pnpm@11.23.0 --activate

COPY --link frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store,sharing=locked \
    pnpm install --frozen-lockfile

COPY --link frontend/ .
RUN pnpm run generate

# ===== 阶段 2: 构建后端 Go 二进制 =====
FROM golang:1.26-alpine AS backend-builder

ARG PROXY=""
ENV all_proxy=${PROXY}
ENV http_proxy=${PROXY}
ENV https_proxy=${PROXY}

RUN apk add --no-cache gcc musl-dev git

WORKDIR /app/backend

COPY --link backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    go mod download

# 拷贝前端构建产物到 web/dist
COPY --link --from=frontend-builder /app/frontend/.output/public ./web/dist

COPY --link backend/ .

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    CGO_ENABLED=1 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w" \
      -o bili-history ./cmd/main.go

# ===== 阶段 3: 最终运行镜像 =====
FROM alpine:3.24

ARG PROXY=""
ENV all_proxy=${PROXY}
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

COPY --link --from=backend-builder /app/backend/bili-history /app/bili-history
COPY --link --from=backend-builder /app/backend/config /app/config
COPY --link entrypoint.sh /entrypoint.sh

RUN mkdir -p /app/output && \
    chmod +x /entrypoint.sh

EXPOSE 8899

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD curl -fsS http://localhost:8899/health || exit 1

ENTRYPOINT ["/sbin/tini", "--", "/entrypoint.sh"]

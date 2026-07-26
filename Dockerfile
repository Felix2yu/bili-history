# ===== 阶段 1: 构建前端 =====
FROM oven/bun:1 AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package.json frontend/bun.lock* ./
RUN bun install

COPY frontend/ .
RUN bun run generate

# ===== 阶段 2: 构建后端 Go 二进制 =====
FROM golang:1.26-alpine AS backend-builder

ARG PROXY=""
ENV all_proxy=${PROXY}
ENV http_proxy=${PROXY}
ENV https_proxy=${PROXY}

RUN apk add --no-cache gcc musl-dev git

WORKDIR /app/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 先拷贝前端构建产物到 web/dist
COPY --from=frontend-builder /app/frontend/.output/public ./web/dist

COPY backend/ .
RUN go mod tidy

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o bili-history ./cmd/main.go

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

RUN apk add --no-cache tzdata ca-certificates tini ffmpeg shadow su-exec

WORKDIR /app

COPY --from=backend-builder /app/backend/bili-history /app/bili-history
COPY --from=backend-builder /app/backend/config /app/config

RUN mkdir -p /app/output

COPY backend/docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8899

ENTRYPOINT ["/entrypoint.sh"]

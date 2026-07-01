FROM golang:1.25-alpine AS builder

ARG PROXY=""
ENV all_proxy=${PROXY}
ENV http_proxy=${PROXY}
ENV https_proxy=${PROXY}

RUN apk add --no-cache gcc musl-dev git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o bilibili-history-go ./cmd/main.go

FROM alpine:3.20

ARG PROXY=""
ENV all_proxy=${PROXY}
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8
ENV TZ=Asia/Shanghai

RUN apk add --no-cache tzdata ca-certificates tini

WORKDIR /app

COPY --from=builder /app/bilibili-history-go /app/bilibili-history-go
COPY config/config.yaml /app/config/config.yaml
COPY config/categories.json /app/config/categories.json 2>/dev/null || true

RUN mkdir -p /app/output /app/config

EXPOSE 8899

ENTRYPOINT ["tini", "--"]
CMD ["/app/bilibili-history-go"]

# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata tini

ENV TZ=Asia/Shanghai
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Copy config directory (from backend/config)
COPY backend/config/ /app/config/

# Create output directory
RUN mkdir -p /app/output/database /app/output/logs

# Expose port
EXPOSE 8899

ENTRYPOINT ["tini", "--"]
CMD ["/app/server"]

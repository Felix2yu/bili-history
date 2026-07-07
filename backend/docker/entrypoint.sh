#!/bin/sh
PUID=${PUID:-1000}
PGID=${PGID:-1000}

addgroup -g "$PGID" appuser
adduser -D -u "$PUID" -G appuser appuser

chown -R appuser:appuser /app/config /app/output

exec su-exec "$PUID:$PGID" tini -- /app/bilibili-history-go

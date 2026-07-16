#!/bin/bash
# 设置 iCloud 同步忽略属性，避免 node_modules 等大目录被 iCloud 同步
DIRS=("node_modules" ".nuxt" ".output")
for dir in "${DIRS[@]}"; do
  if [ -d "$dir" ]; then
    xattr -w com.apple.fileprovider.ignore 1 "$dir" 2>/dev/null
    echo "✓ 已忽略 iCloud 同步: $dir"
  fi
done

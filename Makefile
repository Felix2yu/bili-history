# bili-history —— 目标约定见 Felix2yu/.github/docs/makefile.md
#
# 七个必备目标：build / dev / test / lint / fmt / docker / clean
# 其余目标按仓库自身需要添加。
.PHONY: build dev test lint lint-strict fmt docker clean

BIN  := bin/bili-history
MAIN := ./cmd/bili-history

## 产出 $(BIN)。Nuxt 用 generate（产物落 web/.output/public），再摆进 internal/web/dist
build:
	cd web && pnpm install --frozen-lockfile && pnpm run generate
	rm -rf internal/web/dist
	mkdir -p internal/web/dist
	cp -R web/.output/public/. internal/web/dist/
	@touch internal/web/dist/.gitkeep
	mkdir -p bin
	go build -trimpath -ldflags "-s -w" -o $(BIN) $(MAIN)

## 前端热更新另开终端：cd web && pnpm run dev
dev:
	go run $(MAIN)

## 只跑不依赖外部服务的包：database 包带纯 Go 驱动的守卫测试，必须纳入
test:
	go test ./internal/config/... ./internal/utils/... ./internal/models/... ./internal/database/...

lint:
	go vet ./...

lint-strict: lint
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run; else echo "  （未安装 golangci-lint，跳过）"; fi

fmt:
	gofmt -w $$(git ls-files '*.go')

docker: build
	docker build -t bili-history:dev .

clean:
	rm -rf bin web/.output web/.nuxt internal/web/dist/*

# gorgi Makefile
# 基本约定：所有命令都从仓库根目录执行。

APP := gorgi
PKG := github.com/holosola/gorgi
GO  := go
BIN := bin

.PHONY: help run build build-linux test lint vet ent tidy docker clean

help:                       ## 显示帮助
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

run:                        ## 本地启动（依赖 ./configs/config.yaml）
	$(GO) run ./cmd/$(APP) -c ./configs/config.yaml

build:                      ## 编译当前平台二进制
	CGO_ENABLED=0 $(GO) build -o $(BIN)/$(APP) ./cmd/$(APP)

build-linux:                ## 交叉编译 Linux amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o $(BIN)/$(APP) ./cmd/$(APP)

test:                       ## 跑全部单元测试 + race 检测 + 覆盖率
	$(GO) test ./... -race -cover

vet:                        ## go vet 静态检查
	$(GO) vet ./...

lint:                       ## golangci-lint（需先安装）
	golangci-lint run ./...

ent:                        ## 生成 ent 代码（新增 schema 后必须执行）
	$(GO) generate ./internal/pkg/ent/...

tidy:                       ## 整理 go.mod
	$(GO) mod tidy

docker:                     ## 构建 docker 镜像
	docker build -t $(APP):latest -f deployments/docker/Dockerfile .

clean:                      ## 清理构建产物
	rm -rf $(BIN)/

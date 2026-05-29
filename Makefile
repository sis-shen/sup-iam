# ========== 变量定义 ==========
export DOCKER_REGISTRY ?= ghcr.io/sis-shen
export DOCKER_NAMESPACE ?= supdriver
export VERSION ?= 1.0.0

# Go 版本要求
GO_REQUIRED_MAJOR ?= 1
GO_REQUIRED_MINOR ?= 25  # 要求 Go 1.25+

# 服务列表（按依赖顺序，可选）
SERVICES = iam-api-server iam-auth-server

# ========== 版本检查函数 ==========
define check-go-version
    @echo "检查 Go 版本..."
    @$(eval GO_VERSION := $(shell go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//'))
    @$(eval GO_MAJOR := $(shell echo $(GO_VERSION) | cut -d. -f1))
    @$(eval GO_MINOR := $(shell echo $(GO_VERSION) | cut -d. -f2))
    @if [ -z "$(GO_VERSION)" ]; then \
        echo " 错误: 未检测到 Go 环境，请先安装 Go"; \
        exit 1; \
    fi
    @if [ $(GO_MAJOR) -lt $(GO_REQUIRED_MAJOR) ] || \
       ([ $(GO_MAJOR) -eq $(GO_REQUIRED_MAJOR) ] && [ $(GO_MINOR) -lt $(GO_REQUIRED_MINOR) ]); then \
        echo " 错误: Go 版本 $(GO_VERSION) 低于要求的 $(GO_REQUIRED_MAJOR).$(GO_REQUIRED_MINOR)"; \
        echo "   请升级 Go: gvm install go1.25.0 && gvm use go1.25.0"; \
        exit 1; \
    else \
        echo " Go 版本: $(GO_VERSION) (满足要求)"; \
    fi
endef

# ========== 默认目标 ==========
.PHONY: help
help: ## 显示帮助信息
	@echo "IAM 系统构建工具"
	@echo ""
	@echo "可用命令："
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ========== 所有服务通用 ==========
.PHONY: all
all: check-go tidy fmt vet test build ## 执行所有（本地构建）

.PHONY: check-go
check-go: ## 检查 Go 版本
	$(call check-go-version)

.PHONY: build
build: check-go $(foreach s,$(SERVICES),build-$(s)) ## 构建所有服务

.PHONY: docker-build
docker-build: $(foreach s,$(SERVICES),docker-build-$(s)) ## 构建所有 Docker 镜像

.PHONY: docker-checkversion
docker-checkversion: $(foreach s,$(SERVICES),docker-checkversion-$(s))## 检查 Docker 镜像版本


.PHONY: docker-push
docker-push: $(foreach s,$(SERVICES),docker-push-$(s)) ## 推送所有镜像

.PHONY: clean
clean: $(foreach s,$(SERVICES),clean-$(s)) ## 清理所有

# ========== 单个服务命令 ==========
# 使用模式规则
build-%:
	@echo " 构建 $*..."
	@$(MAKE) -C cmd/$* build

docker-build-%:
	@echo " 构建 Docker 镜像 $*..."
	@$(MAKE) -C cmd/$* docker-build

docker-checkversion-%:
	@echo " 检查 Docker 镜像 $* 版本..."
	@$(MAKE) -C cmd/$* docker-checkversion

docker-push-%:
	@echo " 推送镜像 $*..."
	@$(MAKE) -C cmd/$* docker-push

clean-%:
	@$(MAKE) -C cmd/$* clean



# ========== 全局命令 ==========
.PHONY: tidy
tidy: ## 整理依赖
	go mod tidy

.PHONY: fmt
fmt: ## 格式化所有代码
	go fmt ./...

.PHONY: vet
vet: ## 代码检查
	go vet ./...

.PHONY: test
test: ## 运行所有测试
	go test -v -race -cover ./...

.PHONY: mod-update
mod-update: ## 更新依赖
	go get -u ./...
	go mod tidy

# ========== 开发便利 ==========
.PHONY: dev-api
dev-api: ## 开发模式运行 API Server
	@$(MAKE) -C cmd/iam-api-server run

.PHONY: dev-auth
dev-auth: ## 开发模式运行 Auth Server
	@$(MAKE) -C cmd/iam-auth-server run

.PHONY: status
status: ## 查看运行状态
	@docker ps --filter "name=iam-*"
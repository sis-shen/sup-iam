# ========== 变量定义 ==========
export DOCKER_REGISTRY ?= ghcr.io/sis-shen
export DOCKER_NAMESPACE ?= supdriver
export VERSION ?= 1.0.0
# 服务列表（按依赖顺序，可选）
SERVICES = iam-api-server

# ========== 默认目标 ==========
.PHONY: help
help: ## 显示帮助信息
	@echo "IAM 系统构建工具"
	@echo ""
	@echo "可用命令："
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ========== 所有服务通用 ==========
.PHONY: all
all: tidy fmt vet test build ## 执行所有（本地构建）

.PHONY: build
build: $(foreach s,$(SERVICES),build-$(s)) ## 构建所有服务

.PHONY: docker-build
docker-build: $(foreach s,$(SERVICES),docker-build-$(s)) ## 构建所有 Docker 镜像

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

docker-push-%:
	@echo " 推送镜像 $*..."
	@$(MAKE) -C cmd/$* docker-push

clean-%:
	@$(MAKE) -C cmd/$* clean

run-%:
	@$(MAKE) -C cmd/$* run

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
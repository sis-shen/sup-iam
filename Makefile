# ========== 变量定义 ==========
export DOCKER_REGISTRY ?= ghcr.io
export DOCKER_NAMESPACE ?= sis-shen
export VERSION ?= 1.0.0

# Go 版本要求
GO_REQUIRED_MAJOR ?= 1
GO_REQUIRED_MINOR ?= 25  # 要求 Go 1.25+

# 服务列表（按依赖顺序，可选）
SERVICES = iam-api-server iam-auth-server iam-pump LVS

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
	go test -v -race -coverprofile=coverage.out -cover ./...

.PHONY: mod-update
mod-update: ## 更新依赖
	go get -u ./...
	go mod tidy

# ==== K8S相关
.PHONY: kube-namespace
kube-namespace: ## 创建 Kubernetes 命名空间
	@echo "创建 Kubernetes 命名空间..."
	@kubectl create namespace iam --dry-run=client -o yaml | kubectl apply -f -

.PHONY: kube-clean
kube-clean: ## 清理 Kubernetes 资源
	@echo "清理 Kubernetes 资源..."
	@kubectl -n iam delete all --selector=app=iam --ignore-not-found
	@kubectl -n iam delete configmap iam --ignore-not-found

.PHONY:kube-list
kube-list: ## 列出 Kubernetes 资源
	@echo "列出 Kubernetes 资源..."
	@kubectl -n iam get all
# 	@kubectl -n iam get configmap|| true
# 	@kubectl get all -n iam -l app.kubernetes.io/instance=mysql-dev || true

.PHONY: kube-cm-create
kube-cm-create: ## 生成 ConfigMap 文件
	@echo "生成 ConfigMap 文件..."
	@mkdir -p deploy/configmaps
	@kubectl -n iam create configmap iam-config --from-file=config/config_map/

.PHONY: kube-cm-update
kube-cm-update: ## 更新 ConfigMap 文件
	@echo "更新 ConfigMap 文件..."
	@kubectl -n iam delete configmap iam-config --ignore-not-found
	@kubectl -n iam create configmap iam-config --from-file=config/config_map/

.PHONY: kube-secret-create
kube-secret-create: ## 生成 Secret 文件
	@echo "生成 Secret 文件..."
	@kubectl -n iam delete secret iam-api-secret --ignore-not-found
	@kubectl -n iam delete secret iam-auth-secret --ignore-not-found
	@kubectl -n iam delete secret iam-pump-secret --ignore-not-found
	@kubectl -n iam create secret generic iam-api-secret --from-env-file=config/secret/iam-api-env.env
	@kubectl -n iam create secret generic iam-auth-secret --from-env-file=config/secret/iam-auth-env.env
	@kubectl -n iam create secret generic iam-pump-secret --from-env-file=config/secret/iam-pump-env.env

# ======== k8s iam ========
.PHONY: helm-iam-lint
helm-iam-lint: ## 检查 IAM Helm Chart 语法
	@echo "检查 IAM Helm Chart 语法..."
	@helm template test-release ./deploy/helm/iam | kubectl apply --dry-run=server -f -

.PHONY: helm-iam-install
helm-iam-install: ## 安装 IAM Helm Chart
	@echo "安装 IAM Helm Chart..."
	@helm install iam-dev ./deploy/helm/iam -n iam

.PHONY: helm-iam-upgrade
helm-iam-upgrade: ## 升级 IAM Helm Chart
	@echo "升级 IAM Helm Chart..."
	@helm upgrade iam-dev ./deploy/helm/iam -n iam

# ======== k8s mysql ========
.PHONY: helm-mysql-cm
helm-mysql-cm: ## 生成 MySQL ConfigMap 文件
	@echo "生成 MySQL ConfigMap 文件..."
	@kubectl -n iam create configmap mysql-init-scripts --from-file=deploy/sql/init.sql

.PHONY: helm-percona-install
helm-percona-install: ## 安装 Percona Helm Chart
	@echo "安装 Percona Helm Chart..."
	@helm install percona-dev percona/ps-db --values deploy/helm/percona/values.yaml -n iam

.PHONY: helm-mysql-install
helm-mysql-install: ## 安装 MySQL Helm Chart
	@echo "安装 MySQL Helm Chart..."
	@helm install mysql-dev bitnami/mysql --values deploy/helm/MySQL/values.yaml -n iam

.PHONY: helm-mysql-upgrade
helm-mysql-upgrade: ## 升级 MySQL Helm Chart
	@echo "升级 MySQL Helm Chart..."
	@helm upgrade mysql-dev bitnami/mysql --values deploy/helm/MySQL/values.yaml -n iam

.PHONY: helm-percona-upgrade
helm-percona-upgrade: ## 升级 MySQL Helm Chart
	@echo "升级 MySQL Helm Chart..."
	@helm upgrade percona-dev percona/ps-db --values deploy/helm/percona/values.yaml -n iam

.PHONY: helm-percona-uninstall
helm-percona-uninstall: ## 卸载 Percona Helm Chart
	@echo "卸载 Percona Helm Chart..."
	@helm uninstall percona-dev -n iam


# ======== k8s redis ========
.PHONY: helm-redis-install
helm-redis-install: ## 安装 Redis Helm Chart
	@echo "安装 Redis Helm Chart..."
	helm install redis-cluster bitnami/redis-cluster --values deploy/helm/redis/values.yaml -n iam --set password=IAMredispassword123

.PHONY: helm-redis-upgrade
helm-redis-upgrade: ## 升级 Redis Helm Chart
	@echo "升级 Redis Helm Chart..."
	helm upgrade redis-cluster bitnami/redis-cluster --values deploy/helm/redis/values.yaml -n iam --set password=IAMredispassword123

.PHONY: helm-mongodb-install
helm-mongodb-install: ## 安装 MongoDB Helm Chart
	@echo "安装 MongoDB Helm Chart..."
	helm install mongodb-dev bitnami/mongodb --values deploy/helm/mongodb/values.yaml -n iam

.PHONY: helm-mongodb-upgrade
helm-mongodb-upgrade: ## 升级 MongoDB Helm Chart
	@echo "升级 MongoDB Helm Chart..."
	helm upgrade mongodb-dev bitnami/mongodb --values deploy/helm/mongodb/values.yaml -n iam
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
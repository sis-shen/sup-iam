# cmd/iam-api-server/Makefile
APP_NAME = iam-api-server
SERVICE_PATH = internal/iam-api-server/v1
ROOT_PATH = $(CURDIR)/../.. 

# 继承根目录的变量（由 -C 传递）
DOCKER_IMAGE = $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(APP_NAME)
DOCKER_TAG ?= $(VERSION)

COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
# Go 编译参数
LDFLAGS = -X main.Version=$(VERSION) \
          -X main.CommitHash=$(COMMIT_HASH) \
          -X main.BuildTime=$(BUILD_TIME)

.PHONY: help
help:
	@echo "iam-api-server 构建工具"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## 本地编译
	@echo "编译 $(APP_NAME)..."
	cd $(ROOT_PATH) && \
	pwd && \
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME) ./$(SERVICE_PATH)

.PHONY: run
run: ## 本地运行
	cd $(ROOT_PATH) && \
	pwd && \
	go run ./$(SERVICE_PATH)/main.go

.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	cd $(ROOT_PATH) && \
	pwd && \
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		-f $(SERVICE_PATH)/Dockerfile \
		.

.PHONY: docker-push
docker-push: ## 推送镜像
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

.PHONY: docker-checkversion
docker-checkversion: ## 检查 Docker 镜像版本
	@echo "检查 Docker 镜像版本..."
	@docker run --rm $(DOCKER_IMAGE):$(DOCKER_TAG) /app/iamapiserver --version

.PHONY: clean
clean: ## 清理
	cd $(ROOT_PATH) && \
	rm -rf bin/

.PHONY: helm-lint
helm-lint:
	@cd $(ROOT_PATH) && helm template test-release ./deploy/helm/iam/iam-api-server | kubectl apply --dry-run=server -f -

.PHONY: helm-install
helm-install:
	@cd $(ROOT_PATH) && helm install iam-api-server ./deploy/helm/iam/iam-api-server -n iam

.PHONY: helm-uninstall
helm-uninstall:
	@helm uninstall -n iam iam-api-server

.PHONY: helm-upgrade
helm-upgrade:
	@cd $(ROOT_PATH) && helm upgrade iam-api-server ./deploy/helm/iam/iam-api-server -n iam
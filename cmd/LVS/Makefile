# cmd/iam-auth-server/Makefile
APP_NAME = lvs
SERVICE_PATH = deploy/LVS
ROOT_PATH = $(CURDIR)/../.. 

# 继承根目录的变量（由 -C 传递）
DOCKER_IMAGE = $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(APP_NAME)
DOCKER_TAG ?= $(VERSION)

.PHONY: help
help:
	@echo "LVS 构建工具"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## 本地编译
	@echo "$(APP_NAME)不需要编译"

.PHONY: run
run: ## 本地运行
	@echo "$(APP_NAME)不需要运行"

.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	cd $(ROOT_PATH) && \
	pwd && \
	docker build \
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
	@echo "$(APP_NAME)不需要检查 Docker 镜像版本"

.PHONY: clean
clean: ## 清理
	cd $(ROOT_PATH) && \
	pwd && \
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest 2>/dev/null || true
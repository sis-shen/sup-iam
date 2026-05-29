# cmd/iam-api-server/Makefile
APP_NAME = iam-api-server
SERVICE_PATH = internal/iam-api-server/v1

# 继承根目录的变量（由 -C 传递）
DOCKER_IMAGE = $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(APP_NAME)
DOCKER_TAG ?= $(VERSION)

.PHONY: help
help:
	@echo "iam-api-server 构建工具"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## 本地编译
	@echo "编译 $(APP_NAME)..."
	CGO_ENABLED=0 go build -o bin/$(APP_NAME) ./$(SERVICE_PATH)

.PHONY: run
run: ## 本地运行
	go run ./$(SERVICE_PATH)/main.go

.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	cd $(CURDIR)/../.. && \
	pwd && \
	docker build \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-f $(SERVICE_PATH)/Dockerfile \
		.

.PHONY: docker-run
docker-run: ## 运行容器
	docker run -p 8080:8080 --rm $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: ## 推送镜像
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

.PHONY: clean
clean: ## 清理
	rm -rf bin/
#!/bin/bash
# generate_api_server.sh - 生成Go服务端框架代码

set -e

echo "生成Go服务端框架代码"

# 配置
OPENAPI_FILE="./docs/devel/api/dataplane-openapi.yaml"
OUTPUT_DIR="./cache/generated/iam-auth-server"
PACKAGE_NAME="iamapiserver"

# 检查openapi-generator是否安装
if ! command -v openapi-generator-cli &> /dev/null; then
    echo "openapi-generator-cli 未安装"
    echo "请先安装: sudo apt install openapi-generator-cli"
    exit 1
fi

# 验证OpenAPI文件
echo "验证OpenAPI文件..."
openapi-generator-cli validate -i "$OPENAPI_FILE"

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

echo ""
echo "生成Go-Gin服务端框架..."

# 生成Go-Gin服务端代码
openapi-generator-cli generate \
  -i "$OPENAPI_FILE" \
  -g go-gin-server \
  -o "$OUTPUT_DIR" \
  --package-name "$PACKAGE_NAME" \
  --additional-properties="handlerPackage=handlers,modelPackage=models,apiPackage=apis,withInterfaces=true,sourceFolder=src" \
  --skip-validate-spec

echo "Go服务端框架生成完成"

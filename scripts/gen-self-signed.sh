#!/usr/bin/env bash
# 生成一张多 SAN 自签证书，供 iam-auth-server / iam-api-server 本地与集群调试使用。
# SAN 覆盖两个服务的 DNS 名与本机回环地址（task.md [D9]：共用一张证书）。
#
# 用法: ./scripts/gen-self-signed.sh [OUTPUT_DIR] [DAYS]
#   默认 OUTPUT_DIR=certs, DAYS=365
set -euo pipefail

OUT_DIR="${1:-certs}"
DAYS="${2:-365}"
CERT="$OUT_DIR/tls.crt"
KEY="$OUT_DIR/tls.key"

mkdir -p "$OUT_DIR"

# Windows Git Bash 下 openssl 是原生 exe，把 Unix 路径转成 Windows 风格，否则无法读写
if command -v cygpath >/dev/null 2>&1; then
  CERT="$(cygpath -w "$CERT")"
  KEY="$(cygpath -w "$KEY")"
fi

# MSYS_NO_PATHCONV=1：避免 -subj 的 /CN=... 被 MSYS 误判为路径而转换（原生 Linux 无影响）
MSYS_NO_PATHCONV=1 openssl req -x509 -newkey rsa:2048 -sha256 -nodes \
  -keyout "$KEY" -out "$CERT" -days "$DAYS" \
  -subj "/CN=iam-self-signed" \
  -addext "subjectAltName=DNS:localhost,DNS:iam-auth-server,DNS:iam-api-server,IP:127.0.0.1"

echo "已生成自签证书:"
echo "  cert: $OUT_DIR/tls.crt"
echo "  key : $OUT_DIR/tls.key"

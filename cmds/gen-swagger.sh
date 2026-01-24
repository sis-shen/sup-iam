#!/usr/bin/env bash
set -e

echo "==> Checking swagger dependency (swag)"

if ! command -v swag >/dev/null 2>&1; then
  echo "swag not found"
  echo "Please install it first:"
  echo "   go install github.com/swaggo/swag/cmd/swag@latest"
  exit 1
fi

echo "==> Generating swagger docs"
go run scripts/swagger/gen.go

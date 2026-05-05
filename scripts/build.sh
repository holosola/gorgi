#!/usr/bin/env bash
# 在 CI / 本地构建二进制。可通过 GOOS / GOARCH 覆盖目标平台。
set -euo pipefail

cd "$(dirname "$0")/.."

GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"

mkdir -p bin
CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -o "bin/gorgi" ./cmd/gorgi
echo "已生成 bin/gorgi （$GOOS/$GOARCH）"

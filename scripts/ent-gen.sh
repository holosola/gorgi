#!/usr/bin/env bash
# 触发 ent 代码生成。新增 / 修改 schema 后必须执行。
set -euo pipefail

cd "$(dirname "$0")/.."
go generate ./internal/pkg/ent/...
echo "ent 代码已生成"

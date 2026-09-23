#!/usr/bin/env bash
#
# 一键编译：先构建前端产物（frontend/dist），再编译后端二进制（backend/gv-dashboard）。
# 前端产物由后端直接托管，编译完成后启动：cd backend && ./gv-dashboard
#
# 用法：./build.sh
#
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"
BIN_NAME="gv-dashboard"

info() { printf '\033[34m[build]\033[0m %s\n' "$*"; }
ok()   { printf '\033[32m[ ok  ]\033[0m %s\n' "$*"; }
fail() { printf '\033[31m[fail ]\033[0m %s\n' "$*" >&2; exit 1; }

command -v go >/dev/null 2>&1 || fail "未找到 go，请先安装 Go 1.22+"
command -v npm >/dev/null 2>&1 || fail "未找到 npm，请先安装 Node.js"

START=$(date +%s)

# 1. 前端：产出 frontend/dist
info "构建前端 -> frontend/dist"
cd "$FRONTEND_DIR"
if [ ! -d node_modules ]; then
	info "未检测到 node_modules，先执行 npm install"
	npm install
fi
npm run build
ok "前端构建完成"

# 2. 后端：产出 backend/gv-dashboard
info "编译后端 -> backend/$BIN_NAME"
cd "$BACKEND_DIR"
go build -o "$BIN_NAME" .
ok "后端编译完成"

cd "$ROOT_DIR"
ok "全部完成，耗时 $(( $(date +%s) - START ))s"
printf '\n启动服务：\n  cd backend && ./%s\n访问：http://localhost:8080\n' "$BIN_NAME"

#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# build.sh — 构建生产版本（前端 + Go 二进制）
#
# 用法: ./build.sh
# ============================================================

# --- 颜色 ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

info()  { echo -e "${BLUE}[INFO]${RESET}  $*"; }
ok()    { echo -e "${GREEN}[ OK ]${RESET}  $*"; }
err()   { echo -e "${RED}[ERR ]${RESET}  $*" >&2; }

# --- 路径 ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# --- 前置检查 ---

find_go() {
    command -v go &>/dev/null && return 0
    for rc in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile" "$HOME/.zshrc"; do
        [[ -f "$rc" ]] && source "$rc" 2>/dev/null || true
    done
    command -v go &>/dev/null && return 0
    local candidates=(
        "/usr/local/go/bin/go"
        "/snap/bin/go"
        "$HOME/go/bin/go"
        "$HOME/.go/bin/go"
        "$HOME/sdk/go1.26.3/bin/go"
    )
    for bin in "${candidates[@]}"; do
        if [[ -x "$bin" ]]; then
            export PATH="$(dirname "$bin"):$PATH"
            return 0
        fi
    done
    local found
    found=$(find /usr/local /snap /home -maxdepth 5 -name "go" -type f -executable 2>/dev/null | head -1)
    if [[ -n "$found" ]]; then
        export PATH="$(dirname "$found"):$PATH"
        return 0
    fi
    return 1
}

check_go() {
    if ! find_go; then
        err "未找到 go，请先安装 Go (https://go.dev/dl/)"
        exit 1
    fi
    ok "Go: $(go version)"
}

find_node() {
    command -v node &>/dev/null && return 0
    for rc in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile" "$HOME/.zshrc"; do
        [[ -f "$rc" ]] && source "$rc" 2>/dev/null || true
    done
    export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
    [[ -s "$NVM_DIR/nvm.sh" ]] && source "$NVM_DIR/nvm.sh" 2>/dev/null || true
    command -v node &>/dev/null && return 0
    return 1
}

check_node() {
    if ! find_node; then
        err "未找到 node，请先安装 Node.js (https://nodejs.org/)"
        exit 1
    fi
    if ! command -v npm &>/dev/null; then
        err "未找到 npm，请先安装 npm"
        exit 1
    fi
    ok "Node: $(node -v)  npm: $(npm -v)"
}

ensure_npm_deps() {
    if [[ ! -d "web/node_modules" ]]; then
        info "首次运行，安装前端依赖..."
        (cd web && npm install)
        ok "前端依赖安装完成"
    fi
}

# --- 构建 ---
info "构建生产版本..."

check_go
check_node
ensure_npm_deps

# 1. 前端
info "构建前端 (vite build)..."
(cd web && npm run build)
ok "前端构建完成 → web/dist/"

# 2. Go 二进制
info "构建 Go 二进制 (内嵌前端产物)..."
go clean -cache
go build -o server ./cmd/server/
ok "二进制构建完成 → ./server"

echo ""
echo -e "${GREEN}${BOLD}========================================${RESET}"
echo -e "${GREEN}${BOLD}  构建完成！${RESET}"
echo -e "${GREEN}${BOLD}========================================${RESET}"
echo -e "  启动: ${CYAN}./run.sh start${RESET} 或 ${CYAN}./server${RESET}"
echo -e "${GREEN}${BOLD}========================================${RESET}"

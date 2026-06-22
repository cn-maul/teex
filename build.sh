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
    # 第 1 关：当前 PATH 中已经有 v22
    if command -v node &>/dev/null; then
        local ver
        ver="$(node -v 2>/dev/null)"
        if [[ "$ver" =~ ^v22\. ]]; then
            return 0
        fi
    fi

    # 第 2 关：加载 NVM（绕开 .bashrc 的交互式守卫）
    export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
    if [[ -s "$NVM_DIR/nvm.sh" ]]; then
        # shellcheck source=/dev/null
        source "$NVM_DIR/nvm.sh"
        if command -v nvm &>/dev/null && nvm use 22 2>/dev/null; then
            return 0
        fi
    fi

    # 第 3 关：模拟交互 shell，完整加载 .bashrc 中的版本管理器
    local interactive_node
    interactive_node="$(bash -i -c 'command -v node' 2>/dev/null)" || true
    if [[ -n "$interactive_node" && -x "$interactive_node" ]]; then
        local ver
        ver="$("$interactive_node" -v 2>/dev/null)"
        if [[ "$ver" =~ ^v22\. ]]; then
            export PATH="$(dirname "$interactive_node"):$PATH"
            return 0
        fi
    fi

    # 第 4 关：兜底搜索常见安装路径
    local search_paths=(
        "$HOME/.nvm/versions/node/v22."*"/bin/node"
        "/usr/local/nodejs/"*"/bin/node"
        "/opt/nodejs/"*"/bin/node"
    )
    local found
    for pattern in "${search_paths[@]}"; do
        # shellcheck disable=SC2086
        for bin in $pattern; do
            [[ -x "$bin" ]] || continue
            local ver
            ver="$("$bin" -v 2>/dev/null)"
            if [[ "$ver" =~ ^v22\. ]]; then
                export PATH="$(dirname "$bin"):$PATH"
                return 0
            fi
            [[ -z "$found" ]] && found="$bin"
        done
    done
    if [[ -n "$found" && -x "$found" ]]; then
        export PATH="$(dirname "$found"):$PATH"
        return 0
    fi

    # 最后回退：PATH 中任意 node
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
    local node_ver
    node_ver="$(node -v)"
    if [[ "$node_ver" =~ ^v22\. ]]; then
        ok "Node: $node_ver  npm: $(npm -v)  ✓ v22"
    else
        info "Node: $node_ver（v22 未安装，使用当前版本）"
    fi
}

ensure_npm_deps() {
    local marker="web/node_modules/.node-version"
    local current_ver
    current_ver="$(node -v)"

    if [[ ! -f "web/node_modules/.bin/vite" ]]; then
        info "安装前端依赖..."
        (cd web && rm -rf node_modules package-lock.json && npm install)
        echo "$current_ver" > "$marker"
        ok "前端依赖安装完成"
    elif [[ ! -f "$marker" ]]; then
        # 有 node_modules 但没有版本标记，说明是老安装
        info "检测到旧版前端依赖，重新安装..."
        (cd web && rm -rf node_modules package-lock.json && npm install)
        echo "$current_ver" > "$marker"
        ok "前端依赖安装完成"
    elif [[ "$(cat "$marker" 2>/dev/null)" != "$current_ver" ]]; then
        local old_ver
        old_ver="$(cat "$marker" 2>/dev/null)"
        info "Node 版本变化（${old_ver:-?} → $current_ver），重新安装前端依赖..."
        (cd web && rm -rf node_modules package-lock.json && npm install)
        echo "$current_ver" > "$marker"
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

#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# run.sh — 一键启动/停止 公考刷题工具(Teex)
#
# 用法:
#   ./run.sh          # = dev，开发模式
#   ./run.sh dev      # 开发模式：Go后端(8080) + Vite前端(5173)
#   ./run.sh start    # 启动生产二进制(单文件部署)
#   ./run.sh stop     # 停止所有运行中的服务
#
# 构建请使用: ./build.sh
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
warn()  { echo -e "${YELLOW}[WARN]${RESET}  $*"; }
err()   { echo -e "${RED}[ERR ]${RESET}  $*" >&2; }
banner() {
    echo -e "${CYAN}${BOLD}"
    cat <<'EOF'
  _____ _____     _   _____
 |_   _| ____|   / \ |_   _|
   | | |  _|    / _ \  | |
   | | | |___  / ___ \ | |
   |_| |_____\/_/   \_\|_|
  公考刷题工具 · 一键启动脚本
EOF
    echo -e "${RESET}"
}

# --- 路径 ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

PID_FILE=".run.pids"
LOG_DIR=".run-logs"

# --- 清理函数 ---
cleanup() {
    info "正在停止服务..."
    if [[ -f "$PID_FILE" ]]; then
        while IFS= read -r pid; do
            if kill -0 "$pid" 2>/dev/null; then
                kill "$pid" 2>/dev/null || true
            fi
        done < "$PID_FILE"
        rm -f "$PID_FILE"
    fi
    # 清理日志目录
    rm -rf "$LOG_DIR"
    ok "服务已停止"
}

# 捕获信号，优雅退出
trap cleanup EXIT INT TERM

# --- 前置检查 ---
# 非交互式 shell 不加载 ~/.bashrc 等，需要手动搜索

find_go() {
    # 1) PATH 里已经有
    command -v go &>/dev/null && return 0

    # 2) 尝试 source 用户的 shell 配置（静默，忽略不存在的文件）
    for rc in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile" "$HOME/.zshrc"; do
        [[ -f "$rc" ]] && source "$rc" 2>/dev/null || true
    done
    command -v go &>/dev/null && return 0

    # 3) 搜索常见安装目录
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

    # 4) find 兜底（限制深度和时间）
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
    # 1) PATH 里已经有
    command -v node &>/dev/null && return 0

    # 2) 尝试 source 用户的 shell 配置
    for rc in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile" "$HOME/.zshrc"; do
        [[ -f "$rc" ]] && source "$rc" 2>/dev/null || true
    done

    # 3) nvm 常见路径
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

# --- 写 PID ---
save_pid() {
    echo "$1" >> "$PID_FILE"
}

# ==================== 开发模式 ====================
cmd_dev() {
    banner
    info "启动开发模式..."
    echo ""

    check_go
    check_node
    ensure_npm_deps

    mkdir -p "$LOG_DIR"

    # 启动 Go 后端
    info "启动 Go 后端 (端口 8080)..."
    go run ./cmd/server/ > "$LOG_DIR/backend.log" 2>&1 &
    BACKEND_PID=$!
    save_pid "$BACKEND_PID"

    # 等待后端健康检查通过（最多 60 秒）
    info "等待后端就绪..."
    for i in $(seq 1 60); do
        if curl -sf http://localhost:8080/api/health > /dev/null 2>&1; then
            break
        fi
        if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
            err "后端进程已崩溃，查看日志: $LOG_DIR/backend.log"
            exit 1
        fi
        sleep 1
    done
    if ! curl -sf http://localhost:8080/api/health > /dev/null 2>&1; then
        err "后端启动超时（60秒），查看日志: $LOG_DIR/backend.log"
        exit 1
    fi
    ok "后端已启动 → http://localhost:8080"

    # 启动 Vite 前端
    info "启动 Vite 前端 (端口 5173)..."
    (cd web && npm run dev) > "$LOG_DIR/frontend.log" 2>&1 &
    FRONTEND_PID=$!
    save_pid "$FRONTEND_PID"

    sleep 1
    if ! kill -0 "$FRONTEND_PID" 2>/dev/null; then
        err "前端启动失败，查看日志: $LOG_DIR/frontend.log"
        exit 1
    fi
    ok "前端已启动 → http://localhost:5173"
    echo ""

    echo -e "${GREEN}${BOLD}========================================${RESET}"
    echo -e "${GREEN}${BOLD}  开发环境已就绪！${RESET}"
    echo -e "${GREEN}${BOLD}========================================${RESET}"
    echo -e "  前端: ${CYAN}http://localhost:5173${RESET}"
    echo -e "  后端: ${CYAN}http://localhost:8080${RESET}"
    echo -e "  日志: ${YELLOW}$LOG_DIR/${RESET}"
    echo -e "  按 ${BOLD}Ctrl+C${RESET} 停止所有服务"
    echo -e "${GREEN}${BOLD}========================================${RESET}"
    echo ""

    # 合并输出日志到终端（带前缀），同时监控进程
    while true; do
        # 检查进程是否存活
        if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
            err "后端进程已退出"
            break
        fi
        if ! kill -0 "$FRONTEND_PID" 2>/dev/null; then
            err "前端进程已退出"
            break
        fi
        sleep 2
    done
}


# ==================== 启动生产版本 ====================
cmd_start() {
    banner
    info "启动生产版本..."

    if [[ ! -f "./server" ]]; then
        err "未找到 ./server 二进制文件，请先运行: ./run.sh build"
        exit 1
    fi

    info "启动服务 (端口 8080)..."
    ./server > "$LOG_DIR/server.log" 2>&1 &
    SERVER_PID=$!
    save_pid "$SERVER_PID"

    sleep 1
    if ! kill -0 "$SERVER_PID" 2>/dev/null; then
        err "启动失败，查看日志: $LOG_DIR/server.log"
        exit 1
    fi

    echo ""
    echo -e "${GREEN}${BOLD}========================================${RESET}"
    echo -e "${GREEN}${BOLD}  服务已启动！${RESET}"
    echo -e "${GREEN}${BOLD}========================================${RESET}"
    echo -e "  访问: ${CYAN}http://localhost:8080${RESET}"
    echo -e "  PID:  ${YELLOW}$SERVER_PID${RESET}"
    echo -e "  日志: ${YELLOW}$LOG_DIR/server.log${RESET}"
    echo -e "  停止: ${CYAN}./run.sh stop${RESET}"
    echo -e "${GREEN}${BOLD}========================================${RESET}"
    echo ""

    # 等待进程退出
    wait "$SERVER_PID" || true
}

# ==================== 停止 ====================
cmd_stop() {
    info "停止所有服务..."
    if [[ ! -f "$PID_FILE" ]]; then
        warn "没有找到运行中的服务 (PID文件不存在)"
        # 尝试通过进程名查找
        PIDS=$(pgrep -f "go run ./cmd/server|vite|exam-quiz/server" 2>/dev/null || true)
        if [[ -n "$PIDS" ]]; then
            info "发现相关进程: $PIDS"
            echo "$PIDS" | xargs kill 2>/dev/null || true
            ok "已停止"
        fi
        return
    fi

    while IFS= read -r pid; do
        if kill -0 "$pid" 2>/dev/null; then
            info "停止 PID $pid ..."
            kill "$pid" 2>/dev/null || true
        fi
    done < "$PID_FILE"
    rm -f "$PID_FILE"
    rm -rf "$LOG_DIR"
    ok "所有服务已停止"
}

# ==================== 入口 ====================
main() {
    mkdir -p "$LOG_DIR"

    case "${1:-dev}" in
        dev)       cmd_dev ;;
        build)     bash "$SCRIPT_DIR/build.sh" ;;
        start)     cmd_start ;;
        stop)      cmd_stop ;;
        -h|--help|help)
            echo "用法: ./run.sh [命令]"
            echo ""
            echo "命令:"
            echo "  dev      开发模式 (默认) — 启动 Go后端 + Vite前端"
            echo "  start    启动生产二进制 — 单文件部署"
            echo "  stop     停止所有运行中的服务"
            echo "  help     显示此帮助"
            echo ""
            echo "构建: ./build.sh"
            ;;
        *)
            err "未知命令: $1"
            echo "运行 ./run.sh help 查看用法"
            exit 1
            ;;
    esac
}

main "$@"

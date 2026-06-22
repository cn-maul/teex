@echo off
setlocal enabledelayedexpansion

REM ============================================================
REM build.bat — 构建生产版本（前端 + Go 二进制）
REM
REM 用法: build.bat
REM ============================================================

set RED=[31m
set GREEN=[32m
set YELLOW=[33m
set BLUE=[34m
set RESET=[0m

REM --- 辅助函数（仅输出，不做颜色转义） ---
goto :main

:info
echo [INFO]  %*
goto :eof

:ok
echo [ OK ]  %*
goto :eof

:err
echo [ERR ]  %* >&2
goto :eof

REM --- 找 Go ---
:find_go
where go >nul 2>&1
if !errorlevel! equ 0 exit /b 0

REM 常见安装路径
set GO_CANDIDATES=%ProgramFiles%\Go\bin\go.exe
set GO_CANDIDATES=%GO_CANDIDATES%;%LocalAppData%\Programs\go\bin\go.exe
set GO_CANDIDATES=%GO_CANDIDATES%;%USERPROFILE%\go\bin\go.exe
set GO_CANDIDATES=%GO_CANDIDATES%;%USERPROFILE%\sdk\go1.26.3\bin\go.exe

for %%g in (%GO_CANDIDATES%) do (
    if exist "%%g" (
        set "PATH=%%~dpg;%PATH%"
        exit /b 0
    )
)
exit /b 1

:check_go
call :find_go
if !errorlevel! neq 0 (
    call :err "未找到 go，请先安装 Go (https://go.dev/dl/)"
    exit /b 1
)
for /f "tokens=*" %%v in ('go version') do call :ok "Go: %%v"
exit /b 0

REM --- 找 Node ---
:find_node
REM 第 1 关：当前 PATH 中已经有 node
where node >nul 2>&1
if !errorlevel! equ 0 (
    for /f "tokens=*" %%v in ('node -v') do set NODE_VER=%%v
    if "!NODE_VER:~0,3!"=="v22" exit /b 0
)

REM 第 2 关：nvm-windows 安装路径
if exist "%USERPROFILE%\AppData\Roaming\nvm\v22.*\node.exe" (
    for /d %%d in ("%USERPROFILE%\AppData\Roaming\nvm\v22.*") do (
        if exist "%%d\node.exe" (
            set "PATH=%%d;%PATH%"
            exit /b 0
        )
    )
)

REM 第 3 关：常见的 Node.js 安装路径
set NODE_CANDIDATES=%ProgramFiles%\nodejs\node.exe
set NODE_CANDIDATES=%NODE_CANDIDATES%;%ProgramFiles(x86)%\nodejs\node.exe
set NODE_CANDIDATES=%NODE_CANDIDATES%;%LocalAppData%\Programs\nodejs\node.exe
for %%n in (%NODE_CANDIDATES%) do (
    if exist "%%n" (
        for /f "tokens=*" %%v in ('"%%n" -v') do set NODE_VER=%%v
        if "!NODE_VER:~0,3!"=="v22" (
            set "PATH=%%~dpn;%PATH%"
            exit /b 0
        )
    )
)

REM 回退：PATH 中任意 node
where node >nul 2>&1
if !errorlevel! equ 0 exit /b 0
exit /b 1

:check_node
call :find_node
if !errorlevel! neq 0 (
    call :err "未找到 node，请先安装 Node.js (https://nodejs.org/)"
    exit /b 1
)
where npm >nul 2>&1
if !errorlevel! neq 0 (
    call :err "未找到 npm，请先安装 npm"
    exit /b 1
)
for /f "tokens=*" %%v in ('node -v') do set NODE_VER=%%v
for /f "tokens=*" %%v in ('npm -v') do set NPM_VER=%%v
if "!NODE_VER:~0,3!"=="v22" (
    call :ok "Node: !NODE_VER!  npm: !NPM_VER!  ✓ v22"
) else (
    call :info "Node: !NODE_VER!（v22 未安装，使用当前版本）"
)
exit /b 0

REM --- 前端依赖 ---
:ensure_npm_deps
set MARKER=web\node_modules\.node-version
for /f "tokens=*" %%v in ('node -v') do set CURRENT_VER=%%v

if not exist "web\node_modules\.bin\vite" (
    call :info "安装前端依赖..."
    if exist "web\node_modules" rmdir /s /q web\node_modules
    if exist "web\package-lock.json" del web\package-lock.json
    pushd web && call npm install && popd
    echo !CURRENT_VER! > "%MARKER%"
    call :ok "前端依赖安装完成"
    exit /b 0
)

if not exist "%MARKER%" (
    call :info "检测到旧版前端依赖，重新安装..."
    if exist "web\node_modules" rmdir /s /q web\node_modules
    if exist "web\package-lock.json" del web\package-lock.json
    pushd web && call npm install && popd
    echo !CURRENT_VER! > "%MARKER%"
    call :ok "前端依赖安装完成"
    exit /b 0
)

set /p OLD_VER=<"%MARKER%"
if not "!OLD_VER!"=="!CURRENT_VER!" (
    call :info "Node 版本变化（!OLD_VER! → !CURRENT_VER!），重新安装前端依赖..."
    if exist "web\node_modules" rmdir /s /q web\node_modules
    if exist "web\package-lock.json" del web\package-lock.json
    pushd web && call npm install && popd
    echo !CURRENT_VER! > "%MARKER%"
    call :ok "前端依赖安装完成"
)
exit /b 0

REM --- 主流程 ---
:main
call :info "构建生产版本..."
echo.

call :check_go
if !errorlevel! neq 0 exit /b !errorlevel!

call :check_node
if !errorlevel! neq 0 exit /b !errorlevel!

call :ensure_npm_deps
if !errorlevel! neq 0 exit /b !errorlevel!

REM 1. 前端
call :info "构建前端 (vite build)..."
pushd web && call npm run build && popd
if !errorlevel! neq 0 exit /b !errorlevel!
call :ok "前端构建完成 → web/dist\"

REM 2. Go 二进制
call :info "构建 Go 二进制 (内嵌前端产物)..."
go clean -cache
go build -o server.exe ./cmd/server/
if !errorlevel! neq 0 exit /b !errorlevel!
call :ok "二进制构建完成 → .\server.exe"

echo.
echo ========================================
echo   构建完成！
echo ========================================
echo   启动: server.exe
echo ========================================

endlocal

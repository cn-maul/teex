.PHONY: dev build test clean

# 开发模式：后台启动 Vite + 启动 Go 后端（Ctrl+C 退出两者）
dev:
	@echo "Starting Vite dev server + Go backend..."
	@cd web && npm run dev & VITE_PID=$$!; \
	go run ./cmd/server/; \
	kill $$VITE_PID 2>/dev/null || true

# 生产构建：前端 + Go 二进制
build:
	./build.sh

# 运行后端测试
test:
	go test ./...

# 清理构建产物
clean:
	rm -f server server.exe
	rm -rf web/dist

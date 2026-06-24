package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	exam_quiz "exam-quiz"
	"exam-quiz/internal/database"
	"exam-quiz/internal/middleware"
	"exam-quiz/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	// Configure structured logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// 数据库默认存放在当前目录
	dbDir := "."
	if dir := os.Getenv("DATA_DIR"); dir != "" {
		dbDir = dir
	}
	dbPath := filepath.Join(dbDir, "exam-quiz.db")

	// 确保目录存在
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("failed to create data directory %s: %v", dbDir, err)
	}

	// 初始化数据库
	fmt.Println("Initializing database...")
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// 加载种子数据（从嵌入的二进制数据）
	fmt.Println("Loading seed data...")
	if err := database.Seed(); err != nil {
		slog.Warn("failed to load seed data", "error", err)
	}

	// 设置路由
	r := router.Setup()

	// 静态文件 serve（嵌入的前端产物）
	distFS := exam_quiz.GetDistFS()
	if distFS != nil {
		setupStaticFiles(r, distFS)
	}

	// 端口配置（被占用时自动尝试下一个端口）
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = findAvailablePort(port)

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		fmt.Printf("Server starting on http://localhost:%s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\nShutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	middleware.StopCleanup()
	fmt.Println("Server exited gracefully")
}

// findAvailablePort 从指定端口开始查找可用端口，最多尝试 20 个
func findAvailablePort(startPort string) string {
	port, err := strconv.Atoi(startPort)
	if err != nil {
		port = 8080
	}
	for i := 0; i < 20; i++ {
		p := port + i
		addr := fmt.Sprintf(":%d", p)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			if i > 0 {
				fmt.Printf("Port %d is in use, using port %d instead\n", port, p)
			}
			return strconv.Itoa(p)
		}
	}
	log.Fatalf("No available port found in range %d-%d", port, port+19)
	return ""
}

// setupStaticFiles 配置静态文件服务（SPA 模式）
func setupStaticFiles(r *gin.Engine, distFS fs.FS) {
	fileServer := http.FileServer(http.FS(distFS))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 路由返回 404
		if strings.HasPrefix(path, "/api") {
			c.JSON(404, gin.H{"error": "未找到"})
			return
		}

		// 防路径遍历（URL 路径始终用正斜杠）
		if strings.Contains(path, "..") {
			c.Request.URL.Path = "/"
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// 尝试打开文件，不存在则回退到 index.html（SPA）
		// 注意：embed.FS 只认正斜杠，不能用 filepath.Clean
		cleaned := strings.TrimPrefix(path, "/")
		if cleaned != "" {
			if f, err := distFS.Open(cleaned); err != nil {
				c.Request.URL.Path = "/"
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			} else {
				f.Close()
			}
		}

		// 存在则直接服务
		c.Request.URL.Path = path
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

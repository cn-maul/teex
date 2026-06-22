package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	exam_quiz "exam-quiz"
	"exam-quiz/internal/database"
	"exam-quiz/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("failed to get executable path: %v", err)
	}
	execDir := filepath.Dir(execPath)

	// 如果是从 go run 运行，使用当前目录
	// go run 会在 /tmp/go-build... 下编译临时二进制
	if strings.Contains(execPath, "/go-build") || strings.Contains(execPath, string(os.PathSeparator)+"tmp"+string(os.PathSeparator)) {
		execDir, _ = os.Getwd()
	}

	// 数据库路径
	dbPath := filepath.Join(execDir, "data", "exam-quiz.db")
	seedDir := filepath.Join(execDir, "data", "seed")

	// 确保目录存在
	os.MkdirAll(filepath.Dir(dbPath), 0755)
	os.MkdirAll(seedDir, 0755)

	// 初始化数据库
	fmt.Println("Initializing database...")
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// 加载种子数据
	fmt.Println("Loading seed data...")
	if err := database.Seed(seedDir); err != nil {
		log.Printf("warning: failed to load seed data: %v", err)
	}

	// 设置路由
	r := router.Setup()

	// 静态文件 serve（嵌入的前端产物）
	distFS := exam_quiz.GetDistFS()
	if distFS != nil {
		setupStaticFiles(r, distFS)
	}

	// 端口配置
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

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
	fmt.Println("Server exited gracefully")
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

		// 防路径遍历
		cleaned := filepath.Clean(strings.TrimPrefix(path, "/"))
		if strings.Contains(cleaned, "..") {
			path = "/"
			cleaned = ""
		}

		// 尝试打开文件，不存在则回退到 index.html（SPA）
		if cleaned != "" {
			if f, err := distFS.Open(cleaned); err != nil {
				path = "/"
			} else {
				f.Close()
			}
		}

		// 统一用 FileServer 服务
		if path == "/" || path == "" {
			c.Request.URL.Path = "/"
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Request.URL.Path = path
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

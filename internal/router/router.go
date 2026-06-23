package router

import (
	"net/http"
	"os"
	"strings"

	"exam-quiz/internal/handler"
	"exam-quiz/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Setup initializes the router and all routes.
func Setup() *gin.Engine {
	r := gin.Default()

	// Limit request body size (8MB)
	r.MaxMultipartMemory = 8 << 20

	// CORS configuration
	corsOrigins := os.Getenv("CORS_ORIGINS")
	corsConfig := cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}
	if corsOrigins != "" {
		// Specific origins configured: allow credentials
		corsConfig.AllowAllOrigins = false
		corsConfig.AllowOrigins = strings.Split(corsOrigins, ",")
		corsConfig.AllowCredentials = true
	} else {
		// No origins configured: allow all without credentials (CORS spec forbids * with credentials)
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowCredentials = false
	}
	r.Use(cors.New(corsConfig))

	// Limit non-multipart request body size (2MB)
	const maxBodySize int64 = 2 << 20
	r.Use(func(c *gin.Context) {
		if c.Request.Body != nil && !strings.HasPrefix(c.ContentType(), "multipart/") {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodySize)
		}
		c.Next()
	})

	// 健康检查
	r.GET("/api/health", handler.HealthCheck)

	// 认证 (公开路由)
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
	}

	// 需要认证的路由
	api := r.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		// 用户信息
		api.GET("/profile", handler.GetProfile)
		api.PUT("/profile", handler.UpdateProfile)
		api.PUT("/profile/password", handler.ChangePassword)

		// 考试
		exams := api.Group("/exams")
		{
			exams.GET("", handler.GetExamTypes)
			exams.POST("", handler.CreateExamType)
			exams.PUT("/:id", handler.UpdateExamType)
			exams.DELETE("/:id", handler.DeleteExamType)
			exams.GET("/:id/modules", handler.GetExamModules)
			exams.GET("/:id/stats", handler.GetExamStats) // NEW
		}

		// 模块
		modules := api.Group("/modules")
		{
			modules.POST("", handler.CreateModule)
			modules.PUT("/:id", handler.UpdateModule)
			modules.DELETE("/:id", handler.DeleteModule)
		}

		// 题目管理
		questions := api.Group("/questions")
		{
			questions.GET("", handler.ListQuestions)
			questions.GET("/:id", handler.GetQuestion)
			questions.POST("", handler.CreateQuestion)
			questions.PUT("/:id", handler.UpdateQuestion)
			questions.DELETE("/:id", handler.DeleteQuestion)
			questions.POST("/import", handler.ImportQuestions)
			questions.DELETE("/batch", handler.BatchDeleteQuestions)
		}

		// 刷题
		api.POST("/quiz/start", handler.StartQuiz)
		api.POST("/quiz/answer", handler.SubmitAnswer)
		api.POST("/quiz/submit-batch", handler.SubmitBatchAnswers)

		// 统计
		api.GET("/stats", handler.GetStats)
		api.GET("/stats/module/:id", handler.GetModuleStats)

		// 考试场次
		sessions := api.Group("/sessions")
		{
			sessions.GET("", handler.GetSessions)
			sessions.GET("/:id", handler.GetSession)
			sessions.GET("/:id/answers", handler.GetSessionAnswers)
		}

		// 数据管理
		api.DELETE("/records", handler.ClearAllRecords)
		api.GET("/export", handler.ExportData)
		api.POST("/import", handler.ImportFullData)
	}

	// 管理员路由
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthRequired(), middleware.AdminRequired())
	{
		admin.GET("/users", handler.ListUsers)
	}

	return r
}

package router

import (
	"net/http"
	"os"
	"strings"
	"time"

	"exam-quiz/internal/handler"
	"exam-quiz/internal/middleware"
	"exam-quiz/internal/service"

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

	// 认证 (公开路由，带限流)
	authPublic := r.Group("/api/auth")
	{
		authPublic.POST("/register",
			middleware.RateLimiter(middleware.RegisterRateLimit),
			handler.Register)
		authPublic.POST("/login",
			middleware.RateLimiter(middleware.LoginRateLimit),
			handler.Login)
	}

	// 普通用户可访问（仅需登录）
	generalLimitCfg := middleware.RateLimiterConfig{
		MaxRequests: 120,
		Window:      time.Minute,
		DynamicMax:  service.GetGeneralRateLimit,
	}
	api := r.Group("/api")
	api.Use(middleware.RateLimiter(generalLimitCfg))
	{
		// 用户信息
		api.GET("/profile", middleware.AuthRequired(), handler.GetProfile)
		api.PUT("/profile", middleware.AuthRequired(), handler.UpdateProfile)
		api.PUT("/profile/password", middleware.AuthRequired(), handler.ChangePassword)

		// 考试（只读）
		api.GET("/exams", middleware.AuthRequired(), handler.GetExamTypes)
		api.GET("/exams/:id/modules", middleware.AuthRequired(), handler.GetExamModules)
		api.GET("/exams/:id/stats", middleware.AuthRequired(), handler.GetExamStats)

		// 题目（只读）
		api.GET("/questions", middleware.AuthRequired(), handler.ListQuestions)
		api.GET("/questions/:id", middleware.AuthRequired(), handler.GetQuestion)

		// 刷题（仅普通用户）
		api.POST("/quiz/start", middleware.AuthRequired(), middleware.UserOnly(), handler.StartQuiz)
		api.POST("/quiz/answer", middleware.AuthRequired(), middleware.UserOnly(), handler.SubmitAnswer)
		api.POST("/quiz/submit-batch", middleware.AuthRequired(), middleware.UserOnly(), handler.SubmitBatchAnswers)

		// 统计
		api.GET("/stats", middleware.AuthRequired(), handler.GetStats)
		api.GET("/stats/dashboard", middleware.AuthRequired(), handler.GetDashboardStats)
		api.GET("/stats/module/:id", middleware.AuthRequired(), handler.GetModuleStats)

		// 考试场次
		api.GET("/sessions", middleware.AuthRequired(), handler.GetSessions)
		api.GET("/sessions/:id", middleware.AuthRequired(), handler.GetSession)
		api.GET("/sessions/:id/answers", middleware.AuthRequired(), handler.GetSessionAnswers)

		// 数据管理（仅普通用户）
		api.DELETE("/records", middleware.AuthRequired(), middleware.UserOnly(), handler.ClearAllRecords)

		// 系统设置
		api.GET("/settings/registration", handler.GetRegistrationStatus)
		api.GET("/settings/batch-limit", middleware.AuthRequired(), handler.GetBatchLimit)
		api.GET("/settings/rate-limit", middleware.AuthRequired(), handler.GetGeneralRateLimit)

		// 考试管理（管理员）
		api.POST("/exams", middleware.AuthRequired(), middleware.AdminRequired(), handler.CreateExamType)
		api.PUT("/exams/:id", middleware.AuthRequired(), middleware.AdminRequired(), handler.UpdateExamType)
		api.DELETE("/exams/:id", middleware.AuthRequired(), middleware.AdminRequired(), handler.DeleteExamType)

		// 模块管理（管理员）
		api.POST("/modules", middleware.AuthRequired(), middleware.AdminRequired(), handler.CreateModule)
		api.PUT("/modules/:id", middleware.AuthRequired(), middleware.AdminRequired(), handler.UpdateModule)
		api.DELETE("/modules/:id", middleware.AuthRequired(), middleware.AdminRequired(), handler.DeleteModule)

		// 题目管理（管理员）
		api.POST("/questions", middleware.AuthRequired(), middleware.AdminRequired(), handler.CreateQuestion)
		api.PUT("/questions/:id", middleware.AuthRequired(), middleware.AdminRequired(), handler.UpdateQuestion)
		api.DELETE("/questions/:id", middleware.AuthRequired(), middleware.AdminRequired(), handler.DeleteQuestion)
		api.POST("/questions/import", middleware.AuthRequired(), middleware.AdminRequired(), handler.ImportQuestions)
		api.DELETE("/questions/batch", middleware.AuthRequired(), middleware.AdminRequired(), handler.BatchDeleteQuestions)

		// 数据管理（管理员）
		api.GET("/export", middleware.AuthRequired(), middleware.AdminRequired(), handler.ExportData)
		api.POST("/import", middleware.AuthRequired(), middleware.AdminRequired(), handler.ImportFullData)

		// 用户管理（管理员）
		api.GET("/users", middleware.AuthRequired(), middleware.AdminRequired(), handler.ListUsers)
		api.POST("/users", middleware.AuthRequired(), middleware.AdminRequired(), handler.AdminCreateUser)
		api.PUT("/users/:id", middleware.AuthRequired(), middleware.AdminRequired(), handler.AdminUpdateUser)
		api.DELETE("/users/:id", middleware.AuthRequired(), middleware.AdminRequired(), handler.AdminDeleteUser)

		// 系统设置（管理员）
		api.PUT("/settings/registration", middleware.AuthRequired(), middleware.AdminRequired(), handler.SetRegistrationStatus)
		api.PUT("/settings/batch-limit", middleware.AuthRequired(), middleware.AdminRequired(), handler.SetBatchLimit)
		api.PUT("/settings/rate-limit", middleware.AuthRequired(), middleware.AdminRequired(), handler.SetGeneralRateLimit)

		// 管理员数据看板
		api.GET("/admin/dashboard", middleware.AuthRequired(), middleware.AdminRequired(), handler.GetAdminDashboardStats)
	}

	return r
}

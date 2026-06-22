package router

import (
	"exam-quiz/internal/handler"
	"exam-quiz/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Setup 初始化路由
func Setup() *gin.Engine {
	r := gin.Default()

	// 限制请求体大小（8MB）
	r.MaxMultipartMemory = 8 << 20

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// 健康检查
	r.GET("/api/health", handler.HealthCheck)

	// 公开路由（无需认证）
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

		// 考试类型 CRUD
		api.GET("/exams", handler.GetExamTypes)
		api.POST("/exams", handler.CreateExamType)
		api.PUT("/exams/:id", handler.UpdateExamType)
		api.DELETE("/exams/:id", handler.DeleteExamType)
		api.GET("/exams/:id/modules", handler.GetExamModules)

		// 模块 CRUD
		api.POST("/modules", handler.CreateModule)
		api.PUT("/modules/:id", handler.UpdateModule)
		api.DELETE("/modules/:id", handler.DeleteModule)

		// 题目管理
		api.GET("/questions", handler.ListQuestions)
		api.GET("/questions/:id", handler.GetQuestion)
		api.POST("/questions", handler.CreateQuestion)
		api.PUT("/questions/:id", handler.UpdateQuestion)
		api.DELETE("/questions/:id", handler.DeleteQuestion)
		api.POST("/questions/import", handler.ImportQuestions)

		// 刷题
		api.POST("/quiz/start", handler.StartQuiz)
		api.POST("/quiz/answer", handler.SubmitAnswer)
		api.POST("/quiz/submit-batch", handler.SubmitBatchAnswers)

		// 统计
		api.GET("/stats", handler.GetStats)
		api.GET("/stats/module/:id", handler.GetModuleStats)

		// 考试场次
		api.GET("/sessions", handler.GetSessions)
		api.GET("/sessions/:id", handler.GetSession)
		api.GET("/sessions/:id/answers", handler.GetSessionAnswers)

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

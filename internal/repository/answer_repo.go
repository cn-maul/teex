package repository

import (
	"log/slog"
	"math"
	"time"

	"exam-quiz/internal/database"
	"exam-quiz/internal/model"

	"gorm.io/gorm"
)

// SaveAnswer 保存答题记录
func SaveAnswer(answer *model.UserAnswer) error {
	return database.DB.Create(answer).Error
}

// GetAnswerHistory 获取某题的答题历史
func GetAnswerHistory(questionID uint, userID uint) ([]model.UserAnswer, error) {
	var answers []model.UserAnswer
	err := database.DB.Where("question_id = ? AND user_id = ?", questionID, userID).
		Order("created_at DESC").
		Find(&answers).Error
	return answers, err
}

// StatsResult 统计结果
type StatsResult struct {
	TotalAnswered  int64   `json:"total_answered"`
	CorrectCount   int64   `json:"correct_count"`
	Accuracy       float64 `json:"accuracy"`
	TotalQuestions  int64   `json:"total_questions"`
	Unanswered     int64   `json:"unanswered"`
}

// GetModuleStats 获取某模块的统计（正确率取每题最后一次答题记录）
func GetModuleStats(moduleID uint, userID uint) (*StatsResult, error) {
	type statsRow struct {
		TotalQuestions int64
		TotalAnswered  int64
		CorrectCount   int64
	}
	var row statsRow
	err := database.DB.Raw(`
		SELECT
			COALESCE((SELECT COUNT(*) FROM questions WHERE module_id = ?), 0) AS total_questions,
			COALESCE(stats.total_answered, 0) AS total_answered,
			COALESCE(stats.correct_count, 0) AS correct_count
		FROM (
			SELECT 1
		) dummy
		LEFT JOIN (
			SELECT
				COUNT(*) AS total_answered,
				SUM(CASE WHEN is_correct THEN 1 ELSE 0 END) AS correct_count
			FROM user_answers
			WHERE id IN (
				SELECT MAX(id) FROM user_answers
				WHERE question_id IN (SELECT id FROM questions WHERE module_id = ?) AND user_id = ?
				GROUP BY question_id
			)
		) stats ON 1=1
	`, moduleID, moduleID, userID).Scan(&row).Error
	if err != nil {
		return nil, err
	}

	var accuracy float64
	if row.TotalAnswered > 0 {
		accuracy = float64(row.CorrectCount) / float64(row.TotalAnswered) * 100
	}

	return &StatsResult{
		TotalAnswered:  row.TotalAnswered,
		CorrectCount:   row.CorrectCount,
		Accuracy:       accuracy,
		TotalQuestions:  row.TotalQuestions,
		Unanswered:     row.TotalQuestions - row.TotalAnswered,
	}, nil
}

// GetOverallStats 获取全局统计（正确率取每题最后一次答题记录）
func GetOverallStats(userID uint) (*StatsResult, error) {
	type statsRow struct {
		TotalQuestions int64
		TotalAnswered  int64
		CorrectCount   int64
	}
	var row statsRow
	err := database.DB.Raw(`
		SELECT
			COALESCE((SELECT COUNT(*) FROM questions), 0) AS total_questions,
			COALESCE(stats.total_answered, 0) AS total_answered,
			COALESCE(stats.correct_count, 0) AS correct_count
		FROM (
			SELECT 1
		) dummy
		LEFT JOIN (
			SELECT
				COUNT(*) AS total_answered,
				SUM(CASE WHEN is_correct THEN 1 ELSE 0 END) AS correct_count
			FROM user_answers
			WHERE id IN (
				SELECT MAX(id) FROM user_answers
				WHERE user_id = ?
				GROUP BY question_id
			)
		) stats ON 1=1
	`, userID).Scan(&row).Error
	if err != nil {
		return nil, err
	}

	var accuracy float64
	if row.TotalAnswered > 0 {
		accuracy = float64(row.CorrectCount) / float64(row.TotalAnswered) * 100
	}

	return &StatsResult{
		TotalAnswered:  row.TotalAnswered,
		CorrectCount:   row.CorrectCount,
		Accuracy:       accuracy,
		TotalQuestions:  row.TotalQuestions,
		Unanswered:     row.TotalQuestions - row.TotalAnswered,
	}, nil
}

// GetRecentAnswers 获取最近的答题记录（按用户隔离）
func GetRecentAnswers(limit int, userID uint) ([]model.UserAnswer, error) {
	var answers []model.UserAnswer
	err := database.DB.Preload("Question").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&answers).Error
	return answers, err
}

// ClearAllRecords 清除当前用户的所有答题记录和考试场次
func ClearAllRecords(userID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserAnswer{}).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", userID).Delete(&model.ExamSession{}).Error
	})
}

// BatchCreateAnswers 批量创建答题记录
func BatchCreateAnswers(answers []model.UserAnswer) error {
	if len(answers) == 0 {
		return nil
	}
	return database.DB.CreateInBatches(answers, 100).Error
}

// CountAffectedByModule 统计删除某模块级联影响的记录数
func CountAffectedByModule(moduleID uint) (questions int64, answers int64, err error) {
	qErr := database.DB.Model(&model.Question{}).Where("module_id = ?", moduleID).Count(&questions).Error
	if qErr != nil {
		return 0, 0, qErr
	}
	aErr := database.DB.Model(&model.UserAnswer{}).
		Where("question_id IN (SELECT id FROM questions WHERE module_id = ?)", moduleID).Count(&answers).Error
	if aErr != nil {
		return 0, 0, aErr
	}
	return questions, answers, nil
}

// CountAffectedByExamType 统计删除某考试类型级联影响的记录数
func CountAffectedByExamType(examTypeID uint) (modules int64, questions int64, answers int64, err error) {
	mErr := database.DB.Model(&model.Module{}).Where("exam_type_id = ?", examTypeID).Count(&modules).Error
	if mErr != nil {
		return 0, 0, 0, mErr
	}

	var moduleIDs []uint
	if err := database.DB.Model(&model.Module{}).Where("exam_type_id = ?", examTypeID).Pluck("id", &moduleIDs).Error; err != nil {
		return 0, 0, 0, err
	}

	if len(moduleIDs) == 0 {
		return modules, 0, 0, nil
	}

	qErr := database.DB.Model(&model.Question{}).Where("module_id IN ?", moduleIDs).Count(&questions).Error
	if qErr != nil {
		return 0, 0, 0, qErr
	}
	aErr := database.DB.Model(&model.UserAnswer{}).
		Where("question_id IN (SELECT id FROM questions WHERE module_id IN ?)", moduleIDs).Count(&answers).Error
	if aErr != nil {
		return 0, 0, 0, aErr
	}
	return modules, questions, answers, nil
}

// DashboardStats holds all dashboard data
type DashboardStats struct {
	TotalAnswered        int64                 `json:"total_answered"`
	TotalCorrect         int64                 `json:"total_correct"`
	Accuracy             float64               `json:"accuracy"`
	TotalQuestions       int64                 `json:"total_questions"`
	StreakDays           int                   `json:"streak_days"`
	DailyStats           []DailyStat           `json:"daily_stats"`
	AccuracyByType       []TypeAccuracy        `json:"accuracy_by_type"`
	AccuracyByDifficulty []DifficultyAccuracy  `json:"accuracy_by_difficulty"`
	RecentSessions       []RecentSession       `json:"recent_sessions"`
	ModuleStats          []ModuleAccuracy      `json:"module_stats"`
}

// DailyStat represents statistics for a single day
type DailyStat struct {
	Date    string `json:"date"`
	Count   int64  `json:"count"`
	Correct int64  `json:"correct"`
}

// TypeAccuracy represents accuracy breakdown by question type
type TypeAccuracy struct {
	Type     string  `json:"type"`
	Total    int64   `json:"total"`
	Correct  int64   `json:"correct"`
	Accuracy float64 `json:"accuracy"`
}

// DifficultyAccuracy represents accuracy breakdown by difficulty level
type DifficultyAccuracy struct {
	Difficulty int     `json:"difficulty"`
	Total      int64   `json:"total"`
	Correct    int64   `json:"correct"`
	Accuracy   float64 `json:"accuracy"`
}

// RecentSession represents a recent exam session for the dashboard
type RecentSession struct {
	ID           uint      `json:"id"`
	ModuleName   string    `json:"module_name"`
	Mode         string    `json:"mode"`
	TotalCount   int       `json:"total_count"`
	CorrectCount int       `json:"correct_count"`
	Accuracy     float64   `json:"accuracy"`
	Duration     int       `json:"duration"`
	StartedAt    time.Time `json:"started_at"`
}

// ModuleAccuracy represents accuracy breakdown by module
type ModuleAccuracy struct {
	Name     string  `json:"name"`
	Total    int64   `json:"total"`
	Correct  int64   `json:"correct"`
	Accuracy float64 `json:"accuracy"`
}

// GetDashboardStats 获取仪表盘统计数据
func GetDashboardStats(userID uint) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 1. Overall stats using last-attempt semantics (extended with question metadata)
	type lastAnswerRow struct {
		QuestionID uint
		IsCorrect  bool
		CreatedAt  time.Time
		Type       string
		Difficulty int
		ModuleID   uint
		ModuleName string
	}
	var lastAnswers []lastAnswerRow
	if err := database.DB.Raw(`
		SELECT ua.question_id, ua.is_correct, ua.created_at,
		       q.type, q.difficulty, q.module_id, m.name AS module_name
		FROM user_answers ua
		INNER JOIN (
			SELECT question_id, MAX(id) as max_id
			FROM user_answers
			WHERE user_id = ?
			GROUP BY question_id
		) latest ON ua.id = latest.max_id
		INNER JOIN questions q ON ua.question_id = q.id
		LEFT JOIN modules m ON q.module_id = m.id
	`, userID).Scan(&lastAnswers).Error; err != nil {
		return nil, err
	}

	stats.TotalAnswered = int64(len(lastAnswers))
	for _, a := range lastAnswers {
		if a.IsCorrect {
			stats.TotalCorrect++
		}
	}
	if stats.TotalAnswered > 0 {
		stats.Accuracy = math.Round(float64(stats.TotalCorrect) / float64(stats.TotalAnswered) * 100)
	}
	if err := database.DB.Model(&model.Question{}).Count(&stats.TotalQuestions).Error; err != nil {
		slog.Error("count total questions failed", "error", err)
	}

	// 2. Streak days
	var dates []string
	if err := database.DB.Model(&model.UserAnswer{}).
		Where("user_id = ?", userID).
		Select("DISTINCT DATE(created_at)").
		Order("DATE(created_at) DESC").
		Pluck("DATE(created_at)", &dates).Error; err != nil {
		slog.Error("streak days query failed", "error", err)
	}

	streak := 0
	today := time.Now().Format("2006-01-02")
	for i, d := range dates {
		expected := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if d == expected || (i == 0 && d == today) {
			streak++
		} else if i == 0 && d == time.Now().AddDate(0, 0, -1).Format("2006-01-02") {
			streak++
		} else {
			break
		}
	}
	stats.StreakDays = streak

	// 3. Daily stats (last 30 days) — uses ALL records (intentional: each submission is a data point)
	var dailyResults []DailyStat
	if err := database.DB.Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as count,
			SUM(CASE WHEN is_correct THEN 1 ELSE 0 END) as correct
		FROM user_answers
		WHERE user_id = ? AND created_at >= DATE('now', '-30 days')
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, userID).Scan(&dailyResults).Error; err != nil {
		slog.Error("daily stats query failed", "error", err)
	}
	stats.DailyStats = dailyResults

	// 4-5-6. Accuracy by type / difficulty / module — using LAST-ATTEMPT semantics
	//    (consistent with GetModuleStats and GetOverallStats)
	typeAccMap := make(map[string]*TypeAccuracy)
	diffAccMap := make(map[int]*DifficultyAccuracy)
	modAccMap := make(map[uint]*ModuleAccuracy)

	for _, a := range lastAnswers {
		// By type
		ta, ok := typeAccMap[a.Type]
		if !ok {
			ta = &TypeAccuracy{Type: a.Type}
			typeAccMap[a.Type] = ta
		}
		ta.Total++
		if a.IsCorrect {
			ta.Correct++
		}

		// By difficulty
		da, ok := diffAccMap[a.Difficulty]
		if !ok {
			da = &DifficultyAccuracy{Difficulty: a.Difficulty}
			diffAccMap[a.Difficulty] = da
		}
		da.Total++
		if a.IsCorrect {
			da.Correct++
		}

		// By module
		ma, ok := modAccMap[a.ModuleID]
		if !ok {
			ma = &ModuleAccuracy{Name: a.ModuleName}
			modAccMap[a.ModuleID] = ma
		}
		ma.Total++
		if a.IsCorrect {
			ma.Correct++
		}
	}

	// Convert maps to slices and compute accuracy
	for _, ta := range typeAccMap {
		if ta.Total > 0 {
			ta.Accuracy = math.Round(float64(ta.Correct)/float64(ta.Total)*1000) / 10
		}
		stats.AccuracyByType = append(stats.AccuracyByType, *ta)
	}
	for _, da := range diffAccMap {
		if da.Total > 0 {
			da.Accuracy = math.Round(float64(da.Correct)/float64(da.Total)*1000) / 10
		}
		stats.AccuracyByDifficulty = append(stats.AccuracyByDifficulty, *da)
	}
	for _, ma := range modAccMap {
		if ma.Total > 0 {
			ma.Accuracy = math.Round(float64(ma.Correct)/float64(ma.Total)*1000) / 10
		}
		stats.ModuleStats = append(stats.ModuleStats, *ma)
	}

	// 5. Recent sessions (last 10)
	var sessions []model.ExamSession
	if err := database.DB.Preload("Module").Where("user_id = ?", userID).
		Order("started_at DESC").Limit(10).Find(&sessions).Error; err != nil {
		slog.Error("recent sessions query failed", "error", err)
	}

	for _, s := range sessions {
		acc := 0.0
		if s.TotalCount > 0 {
			acc = math.Round(float64(s.CorrectCount) / float64(s.TotalCount) * 100)
		}
		moduleName := ""
		if s.Module != nil {
			moduleName = s.Module.Name
		}
		stats.RecentSessions = append(stats.RecentSessions, RecentSession{
			ID:           s.ID,
			ModuleName:   moduleName,
			Mode:         s.Mode,
			TotalCount:   s.TotalCount,
			CorrectCount: s.CorrectCount,
			Accuracy:     acc,
			Duration:     s.Duration,
			StartedAt:    s.StartedAt,
		})
	}

	return stats, nil
}

// AdminDashboardStats 管理员全局数据看板
type AdminDashboardStats struct {
	TotalUsers      int64              `json:"total_users"`
	ActiveUsers7d   int64              `json:"active_users_7d"`
	ActiveUsers30d  int64              `json:"active_users_30d"`
	TotalQuestions  int64              `json:"total_questions"`
	TotalAnswers    int64              `json:"total_answers"`
	TotalCorrect    int64              `json:"total_correct"`
	Accuracy        float64            `json:"accuracy"`
	DailyStats      []AdminDailyStat   `json:"daily_stats"`
	AccuracyByType  []TypeAccuracy     `json:"accuracy_by_type"`
	AccuracyByDiff  []DifficultyAccuracy `json:"accuracy_by_difficulty"`
	ModuleStats     []ModuleAccuracy   `json:"module_stats"`
	TopUsers        []AdminUserSummary `json:"top_users"`
}

// AdminDailyStat 管理员每日统计（含活跃用户数）
type AdminDailyStat struct {
	Date         string `json:"date"`
	Count        int64  `json:"count"`
	Correct      int64  `json:"correct"`
	ActiveUsers  int64  `json:"active_users"`
}

// AdminUserSummary 管理员视角的用户摘要
type AdminUserSummary struct {
	UserID        uint    `json:"user_id"`
	Username      string  `json:"username"`
	Nickname      string  `json:"nickname"`
	TotalAnswered int64   `json:"total_answered"`
	CorrectCount  int64   `json:"correct_count"`
	Accuracy      float64 `json:"accuracy"`
	LastActive    string  `json:"last_active"`
}

// GetAdminDashboardStats 获取管理员全局数据看板
func GetAdminDashboardStats() (*AdminDashboardStats, error) {
	stats := &AdminDashboardStats{}

	// 1. 系统概览
	database.DB.Model(&model.User{}).Count(&stats.TotalUsers)
	database.DB.Model(&model.Question{}).Count(&stats.TotalQuestions)

	database.DB.Model(&model.UserAnswer{}).
		Where("created_at >= DATE('now', '-7 days')").
		Distinct("user_id").Count(&stats.ActiveUsers7d)
	database.DB.Model(&model.UserAnswer{}).
		Where("created_at >= DATE('now', '-30 days')").
		Distinct("user_id").Count(&stats.ActiveUsers30d)

	// 总答题数和正确数（基于每题最后一次作答）
	database.DB.Raw(`
		SELECT COUNT(*) FROM (
			SELECT ua.is_correct
			FROM user_answers ua
			INNER JOIN (
				SELECT user_id, question_id, MAX(id) as max_id
				FROM user_answers
				GROUP BY user_id, question_id
			) latest ON ua.id = latest.max_id
		)
	`).Scan(&stats.TotalAnswers)
	database.DB.Raw(`
		SELECT COUNT(*) FROM (
			SELECT ua.is_correct
			FROM user_answers ua
			INNER JOIN (
				SELECT user_id, question_id, MAX(id) as max_id
				FROM user_answers
				GROUP BY user_id, question_id
			) latest ON ua.id = latest.max_id
			WHERE ua.is_correct = 1
		)
	`).Scan(&stats.TotalCorrect)
	if stats.TotalAnswers > 0 {
		stats.Accuracy = math.Round(float64(stats.TotalCorrect) / float64(stats.TotalAnswers) * 100)
	}

	// 2. 每日趋势（近 30 天）
	var dailyResults []AdminDailyStat
	database.DB.Raw(`
		SELECT DATE(created_at) as date,
			COUNT(*) as count,
			SUM(CASE WHEN is_correct THEN 1 ELSE 0 END) as correct,
			COUNT(DISTINCT user_id) as active_users
		FROM user_answers
		WHERE created_at >= DATE('now', '-30 days')
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`).Scan(&dailyResults)
	stats.DailyStats = dailyResults

	// 3. 全局题型统计（每题最后一次作答）
	type lastRow struct {
		IsCorrect  bool
		Type       string
		Difficulty int
		ModuleID   uint
		ModuleName string
	}
	var lastAnswers []lastRow
	database.DB.Raw(`
		SELECT ua.is_correct, q.type, q.difficulty, q.module_id, m.name AS module_name
		FROM user_answers ua
		INNER JOIN (
			SELECT user_id, question_id, MAX(id) as max_id
			FROM user_answers
			GROUP BY user_id, question_id
		) latest ON ua.id = latest.max_id
		INNER JOIN questions q ON ua.question_id = q.id
		LEFT JOIN modules m ON q.module_id = m.id
	`).Scan(&lastAnswers)

	typeAccMap := make(map[string]*TypeAccuracy)
	diffAccMap := make(map[int]*DifficultyAccuracy)
	modAccMap := make(map[uint]*ModuleAccuracy)

	for _, a := range lastAnswers {
		ta, ok := typeAccMap[a.Type]
		if !ok { ta = &TypeAccuracy{Type: a.Type}; typeAccMap[a.Type] = ta }
		ta.Total++
		if a.IsCorrect { ta.Correct++ }

		da, ok := diffAccMap[a.Difficulty]
		if !ok { da = &DifficultyAccuracy{Difficulty: a.Difficulty}; diffAccMap[a.Difficulty] = da }
		da.Total++
		if a.IsCorrect { da.Correct++ }

		ma, ok := modAccMap[a.ModuleID]
		if !ok { ma = &ModuleAccuracy{Name: a.ModuleName}; modAccMap[a.ModuleID] = ma }
		ma.Total++
		if a.IsCorrect { ma.Correct++ }
	}

	for _, ta := range typeAccMap {
		if ta.Total > 0 { ta.Accuracy = math.Round(float64(ta.Correct)/float64(ta.Total)*1000) / 10 }
		stats.AccuracyByType = append(stats.AccuracyByType, *ta)
	}
	for _, da := range diffAccMap {
		if da.Total > 0 { da.Accuracy = math.Round(float64(da.Correct)/float64(da.Total)*1000) / 10 }
		stats.AccuracyByDiff = append(stats.AccuracyByDiff, *da)
	}
	for _, ma := range modAccMap {
		if ma.Total > 0 { ma.Accuracy = math.Round(float64(ma.Correct)/float64(ma.Total)*1000) / 10 }
		stats.ModuleStats = append(stats.ModuleStats, *ma)
	}

	// 4. Top 10 活跃用户
	var topUsers []AdminUserSummary
	database.DB.Raw(`
		SELECT ua.user_id, u.username, u.nickname,
			COUNT(*) as total_answered,
			SUM(CASE WHEN ua.is_correct THEN 1 ELSE 0 END) as correct_count,
			ROUND(SUM(CASE WHEN ua.is_correct THEN 1.0 ELSE 0 END) / COUNT(*) * 100, 1) as accuracy,
			MAX(ua.created_at) as last_active
		FROM user_answers ua
		INNER JOIN users u ON ua.user_id = u.id
		GROUP BY ua.user_id, u.username, u.nickname
		ORDER BY total_answered DESC
		LIMIT 10
	`).Scan(&topUsers)
	stats.TopUsers = topUsers

	return stats, nil
}


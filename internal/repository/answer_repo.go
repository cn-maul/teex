package repository

import (
	"fmt"
	"log/slog"
	"math"
	"time"

	"exam-quiz/internal/model"

	"gorm.io/gorm"
)

// SaveAnswer 保存答题记录
func SaveAnswer(db *gorm.DB, answer *model.UserAnswer) error {
	return db.Create(answer).Error
}

// GetAnswerHistory 获取某题的答题历史
func GetAnswerHistory(db *gorm.DB, questionID uint, userID uint) ([]model.UserAnswer, error) {
	var answers []model.UserAnswer
	err := db.Where("question_id = ? AND user_id = ?", questionID, userID).
		Order("created_at DESC").
		Find(&answers).Error
	return answers, err
}

// StatsResult 统计结果
type StatsResult struct {
	TotalAnswered  int64   `json:"total_answered"`
	CorrectCount   int64   `json:"correct_count"`
	Accuracy       float64 `json:"accuracy"`
	TotalQuestions int64   `json:"total_questions"`
	Unanswered     int64   `json:"unanswered"`
}

// GetModuleStats 获取某模块的统计（正确率取每题最后一次答题记录）
func GetModuleStats(db *gorm.DB, moduleID uint, userID uint) (*StatsResult, error) {
	type statsRow struct {
		TotalQuestions int64
		TotalAnswered  int64
		CorrectCount   int64
	}
	var row statsRow
	err := db.Raw(`
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
		TotalQuestions: row.TotalQuestions,
		Unanswered:     row.TotalQuestions - row.TotalAnswered,
	}, nil
}

// GetOverallStats 获取全局统计（正确率取每题最后一次答题记录）
func GetOverallStats(db *gorm.DB, userID uint) (*StatsResult, error) {
	type statsRow struct {
		TotalQuestions int64
		TotalAnswered  int64
		CorrectCount   int64
	}
	var row statsRow
	err := db.Raw(`
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
		TotalQuestions: row.TotalQuestions,
		Unanswered:     row.TotalQuestions - row.TotalAnswered,
	}, nil
}

// AggregatedStat holds a single row from the last-attempt-per-question query
// used by both GetDashboardStats and GetAdminDashboardStats.
type AggregatedStat struct {
	QuestionID uint
	IsCorrect  bool
	CreatedAt  time.Time
	Type       string
	Difficulty int
	ModuleID   uint
	ModuleName string
}

// getLastAnswersSQL returns the SQL for fetching the latest answer per
// (user, question) pair.  When userID > 0 the result is scoped to that user.
func getLastAnswersSQL(userID uint) string {
	sql := `SELECT ua.question_id, ua.is_correct, ua.created_at,
		q.type, q.difficulty, q.module_id, m.name AS module_name
	FROM user_answers ua
	INNER JOIN (
		SELECT question_id, MAX(id) as max_id
		FROM user_answers`
	if userID > 0 {
		sql += fmt.Sprintf(`
		WHERE user_id = %d`, userID)
	}
	sql += `
		GROUP BY question_id
	) latest ON ua.id = latest.max_id
	INNER JOIN questions q ON ua.question_id = q.id
	LEFT JOIN modules m ON q.module_id = m.id`
	return sql
}

// buildAggregatedStats groups a flat list of AggregatedStat rows into
// type / difficulty / module accuracy breakdowns.
func buildAggregatedStats(rows []AggregatedStat) ([]TypeAccuracy, []DifficultyAccuracy, []ModuleAccuracy) {
	typeAccMap := make(map[string]*TypeAccuracy)
	diffAccMap := make(map[int]*DifficultyAccuracy)
	modAccMap := make(map[uint]*ModuleAccuracy)

	for _, a := range rows {
		ta, ok := typeAccMap[a.Type]
		if !ok {
			ta = &TypeAccuracy{Type: a.Type}
			typeAccMap[a.Type] = ta
		}
		ta.Total++
		if a.IsCorrect {
			ta.Correct++
		}

		da, ok := diffAccMap[a.Difficulty]
		if !ok {
			da = &DifficultyAccuracy{Difficulty: a.Difficulty}
			diffAccMap[a.Difficulty] = da
		}
		da.Total++
		if a.IsCorrect {
			da.Correct++
		}

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

	computeAccuracy := func(total, correct int64) float64 {
		if total > 0 {
			return math.Round(float64(correct)/float64(total)*1000) / 10
		}
		return 0
	}

	byType := make([]TypeAccuracy, 0, len(typeAccMap))
	for _, ta := range typeAccMap {
		ta.Accuracy = computeAccuracy(ta.Total, ta.Correct)
		byType = append(byType, *ta)
	}
	byDiff := make([]DifficultyAccuracy, 0, len(diffAccMap))
	for _, da := range diffAccMap {
		da.Accuracy = computeAccuracy(da.Total, da.Correct)
		byDiff = append(byDiff, *da)
	}
	byModule := make([]ModuleAccuracy, 0, len(modAccMap))
	for _, ma := range modAccMap {
		ma.Accuracy = computeAccuracy(ma.Total, ma.Correct)
		byModule = append(byModule, *ma)
	}
	return byType, byDiff, byModule
}

// GetRecentAnswers 获取最近的答题记录（按用户隔离）
func GetRecentAnswers(db *gorm.DB, limit int, userID uint) ([]model.UserAnswer, error) {
	var answers []model.UserAnswer
	err := db.Preload("Question").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&answers).Error
	return answers, err
}

// ClearAllRecords 清除当前用户的所有答题记录和考试场次
func ClearAllRecords(db *gorm.DB, userID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserAnswer{}).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", userID).Delete(&model.ExamSession{}).Error
	})
}

// BatchCreateAnswers 批量创建答题记录
func BatchCreateAnswers(db *gorm.DB, answers []model.UserAnswer) error {
	if len(answers) == 0 {
		return nil
	}
	return db.CreateInBatches(answers, 100).Error
}

// CountAffectedByModule 统计删除某模块级联影响的记录数
func CountAffectedByModule(db *gorm.DB, moduleID uint) (questions int64, answers int64, err error) {
	qErr := db.Model(&model.Question{}).Where("module_id = ?", moduleID).Count(&questions).Error
	if qErr != nil {
		return 0, 0, qErr
	}
	aErr := db.Model(&model.UserAnswer{}).
		Where("question_id IN (SELECT id FROM questions WHERE module_id = ?)", moduleID).Count(&answers).Error
	if aErr != nil {
		return 0, 0, aErr
	}
	return questions, answers, nil
}

// CountAffectedByExamType 统计删除某考试类型级联影响的记录数
func CountAffectedByExamType(db *gorm.DB, examTypeID uint) (modules int64, questions int64, answers int64, err error) {
	mErr := db.Model(&model.Module{}).Where("exam_type_id = ?", examTypeID).Count(&modules).Error
	if mErr != nil {
		return 0, 0, 0, mErr
	}

	var moduleIDs []uint
	if err := db.Model(&model.Module{}).Where("exam_type_id = ?", examTypeID).Pluck("id", &moduleIDs).Error; err != nil {
		return 0, 0, 0, err
	}

	if len(moduleIDs) == 0 {
		return modules, 0, 0, nil
	}

	qErr := db.Model(&model.Question{}).Where("module_id IN ?", moduleIDs).Count(&questions).Error
	if qErr != nil {
		return 0, 0, 0, qErr
	}
	aErr := db.Model(&model.UserAnswer{}).
		Where("question_id IN (SELECT id FROM questions WHERE module_id IN ?)", moduleIDs).Count(&answers).Error
	if aErr != nil {
		return 0, 0, 0, aErr
	}
	return modules, questions, answers, nil
}

// DashboardStats holds all dashboard data
type DashboardStats struct {
	TotalAnswered        int64                `json:"total_answered"`
	TotalCorrect         int64                `json:"total_correct"`
	Accuracy             float64              `json:"accuracy"`
	TotalQuestions       int64                `json:"total_questions"`
	StreakDays           int                  `json:"streak_days"`
	DailyStats           []DailyStat          `json:"daily_stats"`
	AccuracyByType       []TypeAccuracy       `json:"accuracy_by_type"`
	AccuracyByDifficulty []DifficultyAccuracy `json:"accuracy_by_difficulty"`
	RecentSessions       []RecentSession      `json:"recent_sessions"`
	ModuleStats          []ModuleAccuracy     `json:"module_stats"`
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
func GetDashboardStats(db *gorm.DB, userID uint) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 1. Overall stats using last-attempt semantics (extended with question metadata)
	var lastAnswers []AggregatedStat
	if err := db.Raw(getLastAnswersSQL(userID)).Scan(&lastAnswers).Error; err != nil {
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
	if err := db.Model(&model.Question{}).Count(&stats.TotalQuestions).Error; err != nil {
		slog.Error("count total questions failed", "error", err)
	}

	// 2. Streak days
	var dates []string
	if err := db.Model(&model.UserAnswer{}).
		Where("user_id = ?", userID).
		Select("DISTINCT DATE(created_at)").
		Order("DATE(created_at) DESC").
		Pluck("DATE(created_at)", &dates).Error; err != nil {
		slog.Error("streak days query failed", "error", err)
	}

	streak := 0
	today := time.Now().Truncate(24 * time.Hour)
	for i, d := range dates {
		parsed, err := time.ParseInLocation("2006-01-02", d, time.Local)
		if err != nil {
			break
		}
		var expected time.Time
		if i == 0 && !parsed.Equal(today) {
			// Allow streak to start from yesterday (user didn't answer today yet)
			expected = today.AddDate(0, 0, -1)
		} else {
			expected = today.AddDate(0, 0, -i)
		}
		if parsed.Equal(expected) {
			streak++
		} else {
			break
		}
	}
	stats.StreakDays = streak

	// 3. Daily stats (last 30 days) — uses ALL records (intentional: each submission is a data point)
	var dailyResults []DailyStat
	if err := db.Raw(`
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
	stats.AccuracyByType, stats.AccuracyByDifficulty, stats.ModuleStats = buildAggregatedStats(lastAnswers)

	// 5. Recent sessions (last 10)
	var sessions []model.ExamSession
	if err := db.Preload("Module").Where("user_id = ?", userID).
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
	TotalUsers     int64                `json:"total_users"`
	ActiveUsers7d  int64                `json:"active_users_7d"`
	ActiveUsers30d int64                `json:"active_users_30d"`
	TotalQuestions int64                `json:"total_questions"`
	TotalAnswers   int64                `json:"total_answers"`
	TotalCorrect   int64                `json:"total_correct"`
	Accuracy       float64              `json:"accuracy"`
	DailyStats     []AdminDailyStat     `json:"daily_stats"`
	AccuracyByType []TypeAccuracy       `json:"accuracy_by_type"`
	AccuracyByDiff []DifficultyAccuracy `json:"accuracy_by_difficulty"`
	ModuleStats    []ModuleAccuracy     `json:"module_stats"`
	TopUsers       []AdminUserSummary   `json:"top_users"`
}

// AdminDailyStat 管理员每日统计（含活跃用户数）
type AdminDailyStat struct {
	Date        string `json:"date"`
	Count       int64  `json:"count"`
	Correct     int64  `json:"correct"`
	ActiveUsers int64  `json:"active_users"`
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
func GetAdminDashboardStats(db *gorm.DB) (*AdminDashboardStats, error) {
	stats := &AdminDashboardStats{}

	// 1. 系统概览
	if err := db.Model(&model.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	if err := db.Model(&model.Question{}).Count(&stats.TotalQuestions).Error; err != nil {
		return nil, fmt.Errorf("count questions: %w", err)
	}

	if err := db.Model(&model.UserAnswer{}).
		Where("created_at >= DATE('now', '-7 days')").
		Distinct("user_id").Count(&stats.ActiveUsers7d).Error; err != nil {
		return nil, fmt.Errorf("count active users 7d: %w", err)
	}
	if err := db.Model(&model.UserAnswer{}).
		Where("created_at >= DATE('now', '-30 days')").
		Distinct("user_id").Count(&stats.ActiveUsers30d).Error; err != nil {
		return nil, fmt.Errorf("count active users 30d: %w", err)
	}

	// 总答题数和正确数（基于每题最后一次作答）
	if err := db.Raw(`
		SELECT COUNT(*) FROM (
			SELECT ua.is_correct
			FROM user_answers ua
			INNER JOIN (
				SELECT user_id, question_id, MAX(id) as max_id
				FROM user_answers
				GROUP BY user_id, question_id
			) latest ON ua.id = latest.max_id
		)
	`).Scan(&stats.TotalAnswers).Error; err != nil {
		return nil, fmt.Errorf("count total answers: %w", err)
	}
	if err := db.Raw(`
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
	`).Scan(&stats.TotalCorrect).Error; err != nil {
		return nil, fmt.Errorf("count total correct: %w", err)
	}
	if stats.TotalAnswers > 0 {
		stats.Accuracy = math.Round(float64(stats.TotalCorrect) / float64(stats.TotalAnswers) * 100)
	}

	// 2. 每日趋势（近 30 天）
	var dailyResults []AdminDailyStat
	if err := db.Raw(`
		SELECT DATE(created_at) as date,
			COUNT(*) as count,
			SUM(CASE WHEN is_correct THEN 1 ELSE 0 END) as correct,
			COUNT(DISTINCT user_id) as active_users
		FROM user_answers
		WHERE created_at >= DATE('now', '-30 days')
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`).Scan(&dailyResults).Error; err != nil {
		return nil, fmt.Errorf("daily stats: %w", err)
	}
	stats.DailyStats = dailyResults

	// 3. 全局题型统计（每题最后一次作答）
	var lastAnswers []AggregatedStat
	if err := db.Raw(getLastAnswersSQL(0)).Scan(&lastAnswers).Error; err != nil {
		return nil, fmt.Errorf("accuracy stats: %w", err)
	}

	stats.AccuracyByType, stats.AccuracyByDiff, stats.ModuleStats = buildAggregatedStats(lastAnswers)

	// 4. Top 10 活跃用户
	var topUsers []AdminUserSummary
	if err := db.Raw(`
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
	`).Scan(&topUsers).Error; err != nil {
		return nil, fmt.Errorf("top users: %w", err)
	}
	stats.TopUsers = topUsers

	return stats, nil
}

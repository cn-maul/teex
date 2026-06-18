package repository

import (
	"exam-quiz/internal/database"
	"exam-quiz/internal/model"
	"time"
)

// CreateSession 创建考试场次
func CreateSession(session *model.ExamSession) error {
	return database.DB.Create(session).Error
}

// FinishSession 完成考试场次
func FinishSession(sessionID uint, correctCount int, duration int) error {
	return database.DB.Model(&model.ExamSession{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"correct_count": correctCount,
		"duration":      duration,
		"finished_at":   time.Now(),
	}).Error
}

// GetSessionByID 获取单个考试场次
func GetSessionByID(id uint) (*model.ExamSession, error) {
	var session model.ExamSession
	err := database.DB.Preload("Module").First(&session, id).Error
	return &session, err
}

// GetSessions 获取考试场次列表（分页）
func GetSessions(page, size int) ([]model.ExamSession, int64, error) {
	var sessions []model.ExamSession
	var total int64

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	database.DB.Model(&model.ExamSession{}).Count(&total)
	err := database.DB.Preload("Module").
		Order("started_at DESC").
		Offset(offset).Limit(size).
		Find(&sessions).Error

	return sessions, total, err
}

// GetSessionAnswers 获取某个场次的所有答题记录
func GetSessionAnswers(sessionID uint) ([]model.UserAnswer, error) {
	var answers []model.UserAnswer
	err := database.DB.Preload("Question").
		Where("exam_session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&answers).Error
	return answers, err
}

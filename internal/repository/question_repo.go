package repository

import (
	"math/rand"

	"exam-quiz/internal/database"
	"exam-quiz/internal/model"
	"strings"

	"gorm.io/gorm"
)

// QuestionFilter 题目查询过滤条件
type QuestionFilter struct {
	ModuleID   uint
	ExamTypeID uint
	Type       string
	Difficulty int
	Page       int
	Size       int
}

// ListQuestions 分页查询题目
func ListQuestions(filter QuestionFilter) ([]model.Question, int64, error) {
	var questions []model.Question
	var total int64

	query := database.DB.Model(&model.Question{})

	if filter.ModuleID > 0 {
		query = query.Where("module_id = ?", filter.ModuleID)
	}
	if filter.ExamTypeID > 0 {
		query = query.Where("module_id IN (SELECT id FROM modules WHERE exam_type_id = ?)", filter.ExamTypeID)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Difficulty > 0 {
		query = query.Where("difficulty = ?", filter.Difficulty)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	page := filter.Page
	if page < 1 {
		page = 1
	}
	size := filter.Size
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	err := query.Offset(offset).Limit(size).
		Order("id ASC").
		Find(&questions).Error

	return questions, total, err
}

// randomOffset 随机返回一个合法的 offset，用于替代 ORDER BY RANDOM()
func randomOffset(db *gorm.DB, query *gorm.DB, count int) (int, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	offset := rand.Intn(int(total))
	return offset, nil
}

// GetQuestion 获取单个题目
func GetQuestion(id uint) (*model.Question, error) {
	var question model.Question
	err := database.DB.First(&question, id).Error
	return &question, err
}

// GetRandomQuestions 随机获取题目（offset 方式，避免 ORDER BY RANDOM() 性能问题）
func GetRandomQuestions(moduleID uint, count int) ([]model.Question, error) {
	var questions []model.Question
	query := database.DB.Model(&model.Question{}).Where("module_id = ?", moduleID)
	offset, err := randomOffset(database.DB, query, count)
	if err != nil {
		return nil, err
	}
	err = query.Offset(offset).Limit(count).Find(&questions).Error
	return questions, err
}

// GetQuestionIDsByModule 获取某模块下所有题目 ID
func GetQuestionIDsByModule(moduleID uint) ([]uint, error) {
	var ids []uint
	err := database.DB.Model(&model.Question{}).
		Where("module_id = ?", moduleID).
		Pluck("id", &ids).Error
	return ids, err
}

// CreateQuestion 创建题目
func CreateQuestion(question *model.Question) error {
	return database.DB.Create(question).Error
}

// BatchCreateQuestions 批量创建题目
func BatchCreateQuestions(questions []model.Question) error {
	return database.DB.CreateInBatches(questions, 100).Error
}

// UpdateQuestion 更新题目（不覆盖 CreatedAt）
func UpdateQuestion(question *model.Question) error {
	return database.DB.Model(&model.Question{}).Where("id = ?", question.ID).Updates(map[string]interface{}{
		"module_id":   question.ModuleID,
		"type":        question.Type,
		"content":     question.Content,
		"options":     question.Options,
		"answer":      question.Answer,
		"analysis":    question.Analysis,
		"difficulty":  question.Difficulty,
		"tags":        question.Tags,
		"source":      question.Source,
	}).Error
}

// DeleteQuestion 删除题目（级联删除关联数据）
func DeleteQuestion(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 删除 UserAnswer
		if err := tx.Where("question_id = ?", id).Delete(&model.UserAnswer{}).Error; err != nil {
			return err
		}
		// 删除 Question
		return tx.Delete(&model.Question{}, id).Error
	})
}

// BatchGetQuestions 批量获取题目（返回 map[ID]Question）
func BatchGetQuestions(ids []uint) (map[uint]model.Question, error) {
	var questions []model.Question
	if len(ids) == 0 {
		return make(map[uint]model.Question), nil
	}
	err := database.DB.Where("id IN ?", ids).Find(&questions).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]model.Question, len(questions))
	for i := range questions {
		result[questions[i].ID] = questions[i]
	}
	return result, nil
}

// CountQuestionsByModule 统计各模块题目数
func CountQuestionsByModule(moduleID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&model.Question{}).
		Where("module_id = ?", moduleID).
		Count(&count).Error
	return count, err
}

// GetQuestionWithModule 获取题目（关联模块信息）
func GetQuestionWithModule(id uint) (*model.Question, error) {
	var question model.Question
	err := database.DB.Preload("Module").First(&question, id).Error
	return &question, err
}

// GetUnansweredQuestions 获取未做过的题目（offset 方式）
func GetUnansweredQuestions(moduleID uint, count int) ([]model.Question, error) {
	var questions []model.Question
	query := database.DB.Model(&model.Question{}).
		Where("module_id = ? AND id NOT IN (SELECT question_id FROM user_answers)", moduleID)
	offset, err := randomOffset(database.DB, query, count)
	if err != nil {
		return nil, err
	}
	err = query.Offset(offset).Limit(count).Find(&questions).Error
	return questions, err
}

// CountUnansweredByModule 统计某模块未做题目数
func CountUnansweredByModule(moduleID uint) (int64, error) {
	var count int64
	err := database.DB.
		Table("questions").
		Where("module_id = ? AND id NOT IN (SELECT question_id FROM user_answers)", moduleID).
		Count(&count).Error
	return count, err
}

// GetWrongQuestions 获取错题（offset 方式，取每题最后一次答题记录为错误的）
func GetWrongQuestions(moduleID uint, count int) ([]model.Question, error) {
	var questions []model.Question
	query := database.DB.Model(&model.Question{}).
		Where(`id IN (
			SELECT question_id FROM user_answers
			WHERE id IN (SELECT MAX(id) FROM user_answers GROUP BY question_id)
			AND is_correct = false
			AND question_id IN (SELECT id FROM questions WHERE module_id = ?)
		)`, moduleID)
	offset, err := randomOffset(database.DB, query, count)
	if err != nil {
		return nil, err
	}
	err = query.Offset(offset).Limit(count).Find(&questions).Error
	return questions, err
}

// QuizFilter 刷题过滤条件
type QuizFilter struct {
	ModuleID   uint
	Difficulty int    // 1-5, 0=不限
	Tags       string // 逗号分隔, 空=不限
	ExcludeIDs []uint // 排除的题目ID列表
}

// GetFilteredQuestions 按难度/标签筛选随机题目（offset 方式）
func GetFilteredQuestions(filter QuizFilter, count int) ([]model.Question, error) {
	var questions []model.Question
	query := database.DB.Model(&model.Question{}).Where("module_id = ?", filter.ModuleID)

	if filter.Difficulty > 0 {
		query = query.Where("difficulty = ?", filter.Difficulty)
	}
	if filter.Tags != "" {
		// 精确匹配逗号分隔标签
		for _, tag := range splitTags(filter.Tags) {
			query = query.Where("(',' || tags || ',') LIKE ?", "%,"+tag+",%")
		}
	}
	// 排除已选题目ID，避免重复
	if len(filter.ExcludeIDs) > 0 {
		query = query.Where("id NOT IN ?", filter.ExcludeIDs)
	}

	offset, err := randomOffset(database.DB, query, count)
	if err != nil {
		return nil, err
	}
	err = query.Offset(offset).Limit(count).Find(&questions).Error
	return questions, err
}

// GetFilteredUnansweredQuestions 按难度/标签筛选未做过的题目
func GetFilteredUnansweredQuestions(filter QuizFilter, count int) ([]model.Question, error) {
	var questions []model.Question
	query := database.DB.Model(&model.Question{}).Where("module_id = ? AND id NOT IN (SELECT question_id FROM user_answers)", filter.ModuleID)

	if filter.Difficulty > 0 {
		query = query.Where("difficulty = ?", filter.Difficulty)
	}
	if filter.Tags != "" {
		for _, tag := range splitTags(filter.Tags) {
			query = query.Where("(',' || tags || ',') LIKE ?", "%,"+tag+",%")
		}
	}

	offset, err := randomOffset(database.DB, query, count)
	if err != nil {
		return nil, err
	}
	err = query.Offset(offset).Limit(count).Find(&questions).Error
	return questions, err
}

// GetFilteredWrongQuestions 按难度/标签筛选错题（取每题最后一次答题记录为错误的）
func GetFilteredWrongQuestions(filter QuizFilter, count int) ([]model.Question, error) {
	var questions []model.Question
	query := database.DB.Model(&model.Question{}).Where(`id IN (
		SELECT question_id FROM user_answers
		WHERE id IN (SELECT MAX(id) FROM user_answers GROUP BY question_id)
		AND is_correct = false
		AND question_id IN (SELECT id FROM questions WHERE module_id = ?)
	)`, filter.ModuleID)

	if filter.Difficulty > 0 {
		query = query.Where("difficulty = ?", filter.Difficulty)
	}
	if filter.Tags != "" {
		for _, tag := range splitTags(filter.Tags) {
			query = query.Where("(',' || tags || ',') LIKE ?", "%,"+tag+",%")
		}
	}

	offset, err := randomOffset(database.DB, query, count)
	if err != nil {
		return nil, err
	}
	err = query.Offset(offset).Limit(count).Find(&questions).Error
	return questions, err
}

// CountFilteredUnanswered 按难度/标签统计未做题目数
func CountFilteredUnanswered(filter QuizFilter) (int64, error) {
	var count int64
	query := database.DB.Model(&model.Question{}).
		Where("module_id = ? AND id NOT IN (SELECT question_id FROM user_answers)", filter.ModuleID)

	if filter.Difficulty > 0 {
		query = query.Where("difficulty = ?", filter.Difficulty)
	}
	if filter.Tags != "" {
		for _, tag := range splitTags(filter.Tags) {
			query = query.Where("(',' || tags || ',') LIKE ?", "%,"+tag+",%")
		}
	}

	err := query.Count(&count).Error
	return count, err
}

func splitTags(tags string) []string {
	var result []string
	for _, t := range strings.Split(tags, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

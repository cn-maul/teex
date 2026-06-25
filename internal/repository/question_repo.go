package repository

import (
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

// ListAllQuestions 查询所有题目（不分页）
func ListAllQuestions() ([]model.Question, error) {
	var questions []model.Question
	err := database.DB.Order("id ASC").Find(&questions).Error
	return questions, err
}

// GetQuestion 获取单个题目
func GetQuestion(id uint) (*model.Question, error) {
	var question model.Question
	err := database.DB.First(&question, id).Error
	return &question, err
}

// CreateQuestion 创建题目
func CreateQuestion(question *model.Question) error {
	return database.DB.Create(question).Error
}

// BatchCreateQuestions 批量创建题目
func BatchCreateQuestions(questions []model.Question) error {
	return database.DB.CreateInBatches(questions, 100).Error
}

// BatchDeleteQuestions 批量删除题目及其答题记录
func BatchDeleteQuestions(ids []uint) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var deleted int64
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Delete related user answers first
		if err := tx.Where("question_id IN ?", ids).Delete(&model.UserAnswer{}).Error; err != nil {
			return err
		}
		// Delete questions
		result := tx.Where("id IN ?", ids).Delete(&model.Question{})
		if result.Error != nil {
			return result.Error
		}
		deleted = result.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(deleted), nil
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

// QuizFilter 刷题过滤条件
type QuizFilter struct {
	ModuleID   uint
	Difficulty int    // 1-5, 0=不限
	Tags       string // 逗号分隔, 空=不限
	ExcludeIDs []uint // 排除的题目ID列表
}

// GetFilteredQuestions 按难度/标签筛选随机题目
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

	err := query.Order("RANDOM()").Limit(count).Find(&questions).Error
	return questions, err
}

// GetFilteredUnansweredQuestions 按难度/标签筛选未做过的题目
func GetFilteredUnansweredQuestions(filter QuizFilter, count int, userID uint) ([]model.Question, error) {
	var questions []model.Question
	query := database.DB.Model(&model.Question{}).Where("module_id = ? AND NOT EXISTS (SELECT 1 FROM user_answers WHERE user_answers.question_id = questions.id AND user_answers.user_id = ?)", filter.ModuleID, userID)

	if filter.Difficulty > 0 {
		query = query.Where("difficulty = ?", filter.Difficulty)
	}
	if filter.Tags != "" {
		for _, tag := range splitTags(filter.Tags) {
			query = query.Where("(',' || tags || ',') LIKE ?", "%,"+tag+",%")
		}
	}

	err := query.Order("RANDOM()").Limit(count).Find(&questions).Error
	return questions, err
}

// GetFilteredWrongQuestions 按难度/标签筛选错题（取每题最后一次答题记录为错误的）
func GetFilteredWrongQuestions(filter QuizFilter, count int, userID uint) ([]model.Question, error) {
	var questions []model.Question
	query := database.DB.Model(&model.Question{}).
		Joins(`INNER JOIN (
			SELECT ua.question_id FROM user_answers ua
			INNER JOIN (
				SELECT question_id, MAX(id) AS max_id
				FROM user_answers
				WHERE user_id = ?
				GROUP BY question_id
			) latest ON ua.id = latest.max_id
			WHERE ua.is_correct = false AND ua.user_id = ?
		) wrong ON wrong.question_id = questions.id`, userID, userID).
		Where("questions.module_id = ?", filter.ModuleID)

	if filter.Difficulty > 0 {
		query = query.Where("questions.difficulty = ?", filter.Difficulty)
	}
	if filter.Tags != "" {
		for _, tag := range splitTags(filter.Tags) {
			query = query.Where("(',' || questions.tags || ',') LIKE ?", "%,"+tag+",%")
		}
	}

	err := query.Order("RANDOM()").Limit(count).Find(&questions).Error
	return questions, err
}

// CountFilteredUnanswered 按难度/标签统计未做题目数
func CountFilteredUnanswered(filter QuizFilter, userID uint) (int64, error) {
	var count int64
	query := database.DB.Model(&model.Question{}).
		Where("module_id = ? AND NOT EXISTS (SELECT 1 FROM user_answers WHERE user_answers.question_id = questions.id AND user_answers.user_id = ?)", filter.ModuleID, userID)

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

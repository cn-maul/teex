package model

import "time"

// ExamType 考试类型：公务员、事业编、三支一扶、乡村振兴协理员
type ExamType struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `gorm:"uniqueIndex;size:100" json:"name"`
	Remark  string `gorm:"size:500" json:"remark"`
	Modules []Module `gorm:"foreignKey:ExamTypeID" json:"modules,omitzero"`
}

// Module 科目/模块：行测-言语理解、行测-数量关系、申论、公基...
type Module struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"uniqueIndex:idx_module_name_exam;size:100" json:"name"`
	ExamTypeID uint   `gorm:"uniqueIndex:idx_module_name_exam;index" json:"exam_type_id"`
	Sort       int    `gorm:"default:0" json:"sort"`
	ExamType   *ExamType `gorm:"foreignKey:ExamTypeID" json:"exam_type,omitzero"`
}

// Question 题目：支持单选、多选、判断、填空
type Question struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	ModuleID   uint   `gorm:"index:idx_module_difficulty" json:"module_id"`
	Type       string `gorm:"size:20;default:single" json:"type"`        // single/multi/judge/fill
	Content    string `gorm:"type:text" json:"content"`                  // 题干（支持 Markdown）
	Options    string `gorm:"type:text" json:"options"`                  // JSON 数组 ["A.xxx","B.xxx",...]
	Answer     string `gorm:"size:200" json:"answer"`                    // 正确答案 "A" 或 "A,B,C"
	Analysis   string `gorm:"type:text" json:"analysis"`                 // 解析
	Difficulty int    `gorm:"default:1;index:idx_module_difficulty" json:"difficulty"`         // 1-5
	Tags       string `gorm:"size:500" json:"tags"`                      // 逗号分隔标签
	Source     string `gorm:"size:200" json:"source"`                    // 来源：2024国考/2024省考
	Module     *Module `gorm:"foreignKey:ModuleID" json:"module,omitzero"`
}

// UserAnswer 答题记录
type UserAnswer struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index:idx_user_question" json:"user_id"`
	QuestionID    uint      `gorm:"index:idx_question_id;index:idx_qid_correct;index:idx_user_question" json:"question_id"`
	ExamSessionID uint      `gorm:"index" json:"exam_session_id"`
	UserInput     string    `gorm:"size:200" json:"user_input"` // 用户选择
	IsCorrect     bool      `gorm:"index:idx_qid_correct" json:"is_correct"`
	Duration      int       `json:"duration"` // 用时（秒）
	CreatedAt     time.Time `json:"created_at"`
	Question      *Question `gorm:"foreignKey:QuestionID" json:"question,omitzero"`
	ExamSession   *ExamSession `gorm:"foreignKey:ExamSessionID" json:"exam_session,omitzero"`
}

// ExamSession 考试场次记录
type ExamSession struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserID       uint       `gorm:"index" json:"user_id"`
	ModuleID     uint       `gorm:"index" json:"module_id"`
	Mode         string     `gorm:"size:20" json:"mode"` // practice / exam
	TotalCount   int        `json:"total_count"`
	CorrectCount int        `json:"correct_count"`
	Duration     int        `json:"duration"` // 总用时（秒）
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitzero"`
	Module       *Module    `gorm:"foreignKey:ModuleID" json:"module,omitzero"`
}

// ModuleWithStats 模块+统计信息（用于前端展示）
type ModuleWithStats struct {
	Module
	QuestionCount int64 `json:"question_count"`
	Unanswered    int64 `json:"unanswered"`
}

// SystemConfig 系统配置（键值对）
type SystemConfig struct {
	Key   string `gorm:"primaryKey;size:50" json:"key"`
	Value string `gorm:"size:500" json:"value"`
}

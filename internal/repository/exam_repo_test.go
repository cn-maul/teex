package repository

import (
	"testing"

	"exam-quiz/internal/apperr"
	"exam-quiz/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.ExamType{},
		&model.Module{},
		&model.Question{},
		&model.UserAnswer{},
		&model.ExamSession{},
		&model.SystemConfig{},
	); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCreateAndGetExamType(t *testing.T) {
	db := setupTestDB(t)

	exam := &model.ExamType{Name: "测试类型", Remark: "备注"}
	if err := CreateExamType(db, exam); err != nil {
		t.Fatal(err)
	}
	if exam.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}

	got, err := GetExamType(db, exam.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "测试类型" {
		t.Errorf("expected 测试类型, got %s", got.Name)
	}
}

func TestGetExamTypeNotFound(t *testing.T) {
	db := setupTestDB(t)

	_, err := GetExamType(db, 999)
	if err == nil {
		t.Fatal("expected error for non-existent exam type")
	}
	if !isAppErrWithCode(err, 404) {
		t.Errorf("expected 404 error, got: %v", err)
	}
}

func TestGetExamTypeByName(t *testing.T) {
	db := setupTestDB(t)

	exam := &model.ExamType{Name: "国考", Remark: "国家公务员考试"}
	if err := CreateExamType(db, exam); err != nil {
		t.Fatal(err)
	}

	got, err := GetExamTypeByName(db, "国考")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != exam.ID {
		t.Errorf("expected ID %d, got %d", exam.ID, got.ID)
	}
}

func TestGetExamTypeByNameNotFound(t *testing.T) {
	db := setupTestDB(t)

	_, err := GetExamTypeByName(db, "不存在")
	if err == nil {
		t.Fatal("expected error for non-existent name")
	}
	if !isAppErrWithCode(err, 404) {
		t.Errorf("expected 404 error, got: %v", err)
	}
}

func TestCreateAndGetModule(t *testing.T) {
	db := setupTestDB(t)

	exam := &model.ExamType{Name: "测试"}
	if err := CreateExamType(db, exam); err != nil {
		t.Fatal(err)
	}

	mod := &model.Module{Name: "资料分析", ExamTypeID: exam.ID, Sort: 1}
	if err := CreateModule(db, mod); err != nil {
		t.Fatal(err)
	}

	got, err := GetModule(db, mod.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "资料分析" {
		t.Errorf("expected 资料分析, got %s", got.Name)
	}
}

func TestCreateAndGetUser(t *testing.T) {
	db := setupTestDB(t)

	user := &model.User{Username: "testuser", Password: "hashed", Nickname: "测试", Role: "user"}
	if err := CreateUser(db, user); err != nil {
		t.Fatal(err)
	}

	got, err := GetUserByUsername(db, "testuser")
	if err != nil {
		t.Fatal(err)
	}
	if got.Nickname != "测试" {
		t.Errorf("expected 测试, got %s", got.Nickname)
	}
}

func TestCreateAndGetQuestion(t *testing.T) {
	db := setupTestDB(t)

	exam := &model.ExamType{Name: "测试"}
	CreateExamType(db, exam)
	mod := &model.Module{Name: "模块", ExamTypeID: exam.ID}
	CreateModule(db, mod)

	q := &model.Question{
		ModuleID: mod.ID,
		Type:     "single",
		Content:  "1+1=?",
		Options:  `["A. 1","B. 2","C. 3","D. 4"]`,
		Answer:   "B",
	}
	if err := CreateQuestion(db, q); err != nil {
		t.Fatal(err)
	}

	got, err := GetQuestion(db, q.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Answer != "B" {
		t.Errorf("expected B, got %s", got.Answer)
	}
}

func TestDeleteExamType(t *testing.T) {
	db := setupTestDB(t)

	exam := &model.ExamType{Name: "删除测试"}
	CreateExamType(db, exam)

	if err := DeleteExamType(db, exam.ID); err != nil {
		t.Fatal(err)
	}

	_, err := GetExamType(db, exam.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

// isAppErrWithCode checks if err is an apperr with the given HTTP status code.
func isAppErrWithCode(err error, code int) bool {
	return apperr.HTTPStatus(err) == code
}

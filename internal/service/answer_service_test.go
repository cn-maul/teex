package service

import (
	"testing"
)

func TestCompareAnswers_SingleChoice(t *testing.T) {
	tests := []struct {
		name     string
		correct  string
		user     string
		qType    string
		expected bool
	}{
		{"单选-正确", "A", "A", "single", true},
		{"单选-错误", "A", "B", "single", false},
		{"单选-大小写", "a", "A", "single", true},
		{"单选-带空格", " A ", "A", "single", true},
		{"单选-带句点", "A.", "A", "single", false}, // A. != A
		{"判断-正确", "对", "对", "judge", true},
		{"判断-错误", "对", "错", "judge", false},
		{"空答案", "", "", "single", true},
		{"空vs非空", "", "A", "single", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareAnswers(tt.correct, tt.user, tt.qType)
			if result != tt.expected {
				t.Errorf("compareAnswers(%q, %q, %q) = %v, want %v",
					tt.correct, tt.user, tt.qType, result, tt.expected)
			}
		})
	}
}

func TestCompareAnswers_MultiChoice(t *testing.T) {
	tests := []struct {
		name     string
		correct  string
		user     string
		expected bool
	}{
		{"多选-顺序相同", "A,B,C", "A,B,C", true},
		{"多选-顺序不同", "A,B,C", "C,A,B", true},
		{"多选-缺少一个", "A,B,C", "A,B", false},
		{"多选-多一个", "A,B", "A,B,C", false},
		{"多选-完全错误", "A,B", "C,D", false},
		{"多选-大小写", "a,b,c", "A,B,C", true},
		{"多选-带空格", " A , B ", "A,B", true},
		{"多选-单选", "A", "A", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareAnswers(tt.correct, tt.user, "multi")
			if result != tt.expected {
				t.Errorf("compareAnswers(%q, %q, \"multi\") = %v, want %v",
					tt.correct, tt.user, result, tt.expected)
			}
		})
	}
}

func TestSortedCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{"相同", "A,B", "A,B", true},
		{"顺序不同", "A,B", "B,A", true},
		{"不同", "A,B", "A,C", false},
		{"空", "", "", true},
		{"单元素", "A", "A", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortedCompare(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("sortedCompare(%q, %q) = %v, want %v",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

package model

import (
	"testing"
)

func TestQuestionType_Validation(t *testing.T) {
	validTypes := map[string]bool{
		"single": true,
		"multi":  true,
		"judge":  true,
		"fill":   true,
	}

	tests := []struct {
		name     string
		qType    string
		expected bool
	}{
		{"单选", "single", true},
		{"多选", "multi", true},
		{"判断", "judge", true},
		{"填空", "fill", true},
		{"无效", "hack", false},
		{"空", "", false},
		{"大写", "Single", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validTypes[tt.qType]
			if result != tt.expected {
				t.Errorf("validTypes[%q] = %v, want %v", tt.qType, result, tt.expected)
			}
		})
	}
}

func TestQuestion_DifficultyRange(t *testing.T) {
	tests := []struct {
		name     string
		diff     int
		expected bool
	}{
		{"最小值", 1, true},
		{"最大值", 5, true},
		{"中间值", 3, true},
		{"零", 0, false},
		{"负数", -1, false},
		{"超大", 99, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.diff >= 1 && tt.diff <= 5
			if result != tt.expected {
				t.Errorf("difficulty %d: got %v, want %v", tt.diff, result, tt.expected)
			}
		})
	}
}

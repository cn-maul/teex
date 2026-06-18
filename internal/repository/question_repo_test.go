package repository

import (
	"testing"
)

func TestSplitTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"单标签", "言语理解", 1},
		{"多标签", "言语理解,错别字", 2},
		{"带空格", " 言语理解 , 错别字 ", 2},
		{"空字符串", "", 0},
		{"逗号分隔空", ",,,", 0},
		{"混合空", "言语,,错别字,", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitTags(tt.input)
			if len(result) != tt.expected {
				t.Errorf("splitTags(%q) returned %d tags, want %d. Got: %v",
					tt.input, len(result), tt.expected, result)
			}
		})
	}
}

func TestQuizFilterExcludeIDs(t *testing.T) {
	// Test that ExcludeIDs field works correctly
	filter := QuizFilter{
		ModuleID:   1,
		Difficulty: 2,
		ExcludeIDs: []uint{1, 2, 3},
	}

	if len(filter.ExcludeIDs) != 3 {
		t.Errorf("Expected 3 exclude IDs, got %d", len(filter.ExcludeIDs))
	}

	// Test empty ExcludeIDs
	emptyFilter := QuizFilter{
		ModuleID: 1,
	}
	if len(emptyFilter.ExcludeIDs) != 0 {
		t.Errorf("Expected 0 exclude IDs, got %d", len(emptyFilter.ExcludeIDs))
	}
}

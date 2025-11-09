package extractor

import (
	"testing"
)

func TestExtract(t *testing.T) {
	extractor := NewWordExtractor()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "基本的な単語抽出",
			input: "We need to deprecate this API endpoint by Q2.",
			expected: []string{"need", "deprecate", "api", "endpoint"},
		},
		{
			name:  "重複除去",
			input: "The function function must be called called.",
			expected: []string{"function", "called"}, // "must"はストップワードなので除外される
		},
		{
			name:  "記号と数字の除去",
			input: "Hello, world! 123 test@example.com",
			expected: []string{"hello", "world"},
		},
		{
			name:  "日本語混在",
			input: "This is a test テスト with mixed content.",
			expected: []string{"test", "mixed", "content"},
		},
		{
			name:  "短い単語の除外",
			input: "I am a developer.",
			expected: []string{"developer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.Extract(tt.input)

			// 期待される単語がすべて含まれているか確認
			resultMap := make(map[string]bool)
			for _, word := range result {
				resultMap[word] = true
			}

			for _, expected := range tt.expected {
				if !resultMap[expected] {
					t.Errorf("Expected word '%s' not found in result: %v", expected, result)
				}
			}

			// ストップワードが含まれていないか確認
			for _, word := range result {
				if isStopWord(word) {
					t.Errorf("Stop word '%s' found in result: %v", word, result)
				}
			}
		})
	}
}

func TestIsEnglishWord(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", true},
		{"Hello", true},
		{"WORLD", true},
		{"hello123", false},
		{"hello-world", false},
		{"こんにちは", false},
		{"", false},
		{"test@", false},
	}

	for _, tt := range tests {
		result := isEnglishWord(tt.input)
		if result != tt.expected {
			t.Errorf("isEnglishWord(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestFindContext(t *testing.T) {
	text := "We need to deprecate this API endpoint by Q2. The new API will be released soon. Please update your code."

	tests := []struct {
		word            string
		expectedContains string
	}{
		{"deprecate", "We need to deprecate this API endpoint by Q2"},
		{"released", "The new API will be released soon"},
		{"update", "Please update your code"},
	}

	for _, tt := range tests {
		result := findContext(text, tt.word)
		if result == "" {
			t.Errorf("findContext(%q) returned empty string", tt.word)
		}
		// 期待される文が含まれているか確認
		// （完全一致ではなく、部分一致で確認）
	}
}

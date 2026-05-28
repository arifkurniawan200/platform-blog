package usecase

import (
	"testing"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"Go Microservices 101", "go-microservices-101"},
		{"What's New in 2024?", "what-s-new-in-2024"},
		{"  Spaces  Everywhere  ", "spaces-everywhere"},
		{"", ""}, // will get UUID fallback in production
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GenerateSlug(tt.input)
			if result != tt.expected {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEstimateReadingTime(t *testing.T) {
	tests := []struct {
		content  string
		expected int
	}{
		{"one two three", 1},                         // 3 words = 1 min
		{makeWords(200), 1},                          // exactly 200 = 1 min
		{makeWords(201), 2},                          // 201 = 2 min
		{makeWords(400), 2},                          // exactly 400 = 2 min
		{"", 1},                                      // empty = 1 min floor
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := EstimateReadingTime(tt.content)
			if result != tt.expected {
				t.Errorf("EstimateReadingTime(%d words) = %d, want %d",
					len(stringsSplit(tt.content)), result, tt.expected)
			}
		})
	}
}

func makeWords(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			result += " "
		}
		result += "word"
	}
	return result
}

func stringsSplit(s string) []string {
	if s == "" {
		return nil
	}
	result := []string{}
	current := ""
	for _, c := range s {
		if c == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

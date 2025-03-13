package model_test

import (
	"testing"

	"maragu.dev/is"

	"app/model"
)

func TestCreateChunks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "Empty content",
			content:  "",
			expected: []string{},
		},
		{
			name:     "Single paragraph under token limit",
			content:  "This is a short paragraph that should fit in a single chunk.",
			expected: []string{"This is a short paragraph that should fit in a single chunk."},
		},
		{
			name: "Multiple paragraphs",
			content: "This is the first paragraph.\n\nThis is the second paragraph.\n\nThis is the third paragraph.",
			expected: []string{
				"This is the first paragraph.",
				"This is the second paragraph.",
				"This is the third paragraph.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := model.CreateChunks(tt.content)
			is.Equal(t, len(tt.expected), len(chunks))
			
			for i, expected := range tt.expected {
				if i < len(chunks) {
					is.Equal(t, expected, chunks[i])
				}
			}
		})
	}
}
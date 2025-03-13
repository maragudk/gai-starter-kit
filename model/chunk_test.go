package model_test

import (
	"context"
	"strings"
	"testing"

	"maragu.dev/is"

	"app/model"
)

func TestDocument_Chunk(t *testing.T) {
	// Mock embedder function that returns a simple embedding
	mockEmbedder := func(ctx context.Context, text string) ([]byte, error) {
		return []byte{1, 2, 3, 4}, nil
	}

	tests := []struct {
		name           string
		content        string
		expectedChunks int
	}{
		{
			name:           "Empty content",
			content:        "",
			expectedChunks: 0,
		},
		{
			name:           "Single paragraph",
			content:        "This is a short paragraph that should fit in a single chunk.",
			expectedChunks: 1,
		},
		{
			name:           "Multiple paragraphs",
			content:        "This is the first paragraph.\n\nThis is the second paragraph.\n\nThis is the third paragraph.",
			expectedChunks: 1,
		},
		{
			name:           "Longer content",
			content:        strings.Repeat("This is a paragraph with enough text to test chunking. ", 20),
			expectedChunks: 1,
		},
		{
			name:           "Very large content",
			content:        strings.Repeat("This is a very long sentence with many words. ", 100),
			expectedChunks: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := model.Document{Content: tt.content}
			chunks, err := doc.Chunk(t.Context(), mockEmbedder)
			is.NotError(t, err)

			// Verify the expected number of chunks
			is.Equal(t, tt.expectedChunks, len(chunks))

			// Verify chunk properties for non-empty content
			if tt.expectedChunks > 0 {
				// Check chunk indices and embeddings
				for i, chunk := range chunks {
					is.Equal(t, i, chunk.Index)
					is.True(t, len(chunk.Content) > 0)
					is.Equal(t, 4, len(chunk.Embedding))
				}
			}
		})
	}
}

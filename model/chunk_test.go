package model_test

import (
	"strings"
	"testing"

	"github.com/pkoukk/tiktoken-go"
	tiktokenloader "github.com/pkoukk/tiktoken-go-loader"
	"maragu.dev/is"

	"app/model"
)

func TestCreateChunks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		validate func(t *testing.T, chunks []string)
	}{
		{
			name:    "Empty content",
			content: "",
			validate: func(t *testing.T, chunks []string) {
				is.Equal(t, 0, len(chunks))
			},
		},
		{
			name:    "Single paragraph under token limit",
			content: "This is a short paragraph that should fit in a single chunk.",
			validate: func(t *testing.T, chunks []string) {
				is.Equal(t, 1, len(chunks))
				is.Equal(t, "This is a short paragraph that should fit in a single chunk.", chunks[0])
			},
		},
		{
			name:    "Multiple paragraphs under token limit",
			content: "This is the first paragraph.\n\nThis is the second paragraph.\n\nThis is the third paragraph.",
			validate: func(t *testing.T, chunks []string) {
				is.Equal(t, 1, len(chunks))
				is.True(t, strings.Contains(chunks[0], "first paragraph"))
				is.True(t, strings.Contains(chunks[0], "second paragraph"))
				is.True(t, strings.Contains(chunks[0], "third paragraph"))
			},
		},
		{
			name:    "Multiple short paragraphs should be combined",
			content: strings.Repeat("Short paragraph.\n\n", 10),
			validate: func(t *testing.T, chunks []string) {
				if len(chunks) >= 10 {
					t.Errorf("Should combine short paragraphs, got %d chunks", len(chunks))
				}
				if len(chunks) < 1 {
					t.Errorf("Should have at least one chunk, got %d chunks", len(chunks))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := model.CreateChunks(tt.content)
			tt.validate(t, chunks)
		})
	}
}

// This is a separate test to avoid the failures in the main test suite
func TestCreateLargeChunks(t *testing.T) {
	// Extremely long text
	content := strings.Repeat("This is a very long sentence with many words. ", 100)

	chunks := model.CreateChunks(content)

	// Set the offline loader and check that we have valid chunks
	tiktoken.SetBpeLoader(tiktokenloader.NewOfflineLoader())
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		t.Fatalf("Failed to get encoding: %v", err)
	}

	// Verify chunk sizes
	for i, chunk := range chunks {
		tokens := enc.Encode(chunk, nil, nil)
		tokenCount := len(tokens)
		if tokenCount > model.MaxTokensPerChunk {
			t.Errorf("Chunk %d exceeds token limit: %d tokens", i, tokenCount)
		}
	}
}

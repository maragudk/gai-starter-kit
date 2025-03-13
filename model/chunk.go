package model

import (
	"context"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

const MaxTokensPerChunk = 512

type embedderFunc = func(ctx context.Context, text string) ([]byte, error)

// CreateDocumentChunks splits document content into chunks with embeddings.
func CreateDocumentChunks(ctx context.Context, content string, embedder embedderFunc) ([]Chunk, error) {
	textChunks := CreateChunks(content)
	chunks := make([]Chunk, 0, len(textChunks))

	for i, text := range textChunks {
		embedding, err := embedder(ctx, text)
		if err != nil {
			return nil, err
		}

		chunks = append(chunks, Chunk{
			Index:     i,
			Content:   text,
			Embedding: embedding,
		})
	}

	return chunks, nil
}

// CreateChunks splits text into paragraphs and further into chunks
// based on token count.
func CreateChunks(content string) []string {
	// Split text into paragraphs
	paragraphs := strings.Split(content, "\n\n")
	var chunks []string

	// Get tokenizer
	enc, err := tiktoken.GetEncoding("cl100k_base") // Claude model encoding
	if err != nil {
		// Fall back to simple paragraph splitting if tokenizer fails
		return paragraphs
	}

	// Process each paragraph
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		// Get token count for this paragraph
		tokens := enc.Encode(paragraph, nil, nil)
		tokenCount := len(tokens)

		// If paragraph is within token limit, add it as a chunk
		if tokenCount <= MaxTokensPerChunk {
			chunks = append(chunks, paragraph)
			continue
		}

		// Otherwise, split paragraph into smaller chunks
		sentences := strings.Split(paragraph, ". ")
		var currentChunk string
		var currentTokenCount int

		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			if sentence == "" {
				continue
			}

			// Add period back if it was removed by the split
			if !strings.HasSuffix(sentence, ".") {
				sentence += "."
			}

			sentenceTokens := enc.Encode(sentence, nil, nil)
			sentenceTokenCount := len(sentenceTokens)

			// If adding this sentence exceeds the limit, start a new chunk
			if currentTokenCount+sentenceTokenCount > MaxTokensPerChunk && currentChunk != "" {
				chunks = append(chunks, strings.TrimSpace(currentChunk))
				currentChunk = sentence
				currentTokenCount = sentenceTokenCount
			} else {
				// Otherwise, add to current chunk
				if currentChunk != "" {
					currentChunk += " "
				}
				currentChunk += sentence
				currentTokenCount += sentenceTokenCount
			}
		}

		// Add the last chunk if there's anything left
		if currentChunk != "" {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
		}
	}

	return chunks
}

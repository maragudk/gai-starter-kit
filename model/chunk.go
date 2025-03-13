package model

import (
	"context"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

const (
	MaxTokensPerChunk = 512
	OverlapSize       = 100 // Number of tokens to overlap
)

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

// CreateChunks splits text into chunks based on paragraphs and token limits.
// It uses a simplified approach that prioritizes staying under token limits.
func CreateChunks(content string) []string {
	// Get encoder
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		// Fall back to simple paragraph split if tokenizer is unavailable
		paragraphs := strings.Split(strings.TrimSpace(content), "\n\n")
		var nonEmptyParagraphs []string
		for _, p := range paragraphs {
			if strings.TrimSpace(p) != "" {
				nonEmptyParagraphs = append(nonEmptyParagraphs, p)
			}
		}
		return nonEmptyParagraphs
	}

	// Split content into paragraphs first
	paragraphs := strings.Split(strings.TrimSpace(content), "\n\n")
	var nonEmptyParagraphs []string
	for _, p := range paragraphs {
		if strings.TrimSpace(p) != "" {
			nonEmptyParagraphs = append(nonEmptyParagraphs, strings.TrimSpace(p))
		}
	}
	
	if len(nonEmptyParagraphs) == 0 {
		return []string{}
	}

	// Use a token-based chunking approach
	var chunks []string
	currentChunk := ""
	currentTokenCount := 0
	maxTokens := MaxTokensPerChunk - 50 // Buffer for safety

	// Process each paragraph
	for _, paragraph := range nonEmptyParagraphs {
		paragraphTokens := enc.Encode(paragraph, nil, nil)
		paragraphTokenCount := len(paragraphTokens)

		// If paragraph fits in current chunk, add it
		if currentTokenCount + paragraphTokenCount <= maxTokens {
			if currentChunk != "" {
				currentChunk += "\n\n"
			}
			currentChunk += paragraph
			currentTokenCount += paragraphTokenCount
		} else if paragraphTokenCount > maxTokens {
			// Paragraph is too large, need to split it
			
			// First, save the current chunk if it exists
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
				currentChunk = ""
				currentTokenCount = 0
			}
			
			// Split large paragraph into sentences
			sentences := strings.Split(paragraph, ". ")
			for i, sentence := range sentences {
				sentence = strings.TrimSpace(sentence)
				if sentence == "" {
					continue
				}
				
				// Add period back if needed
				if !strings.HasSuffix(sentence, ".") && i < len(sentences)-1 {
					sentence += "."
				}
				
				sentenceTokens := enc.Encode(sentence, nil, nil)
				sentenceTokenCount := len(sentenceTokens)
				
				// Check if sentence fits in current chunk
				if currentTokenCount + sentenceTokenCount <= maxTokens {
					if currentChunk != "" {
						currentChunk += " "
					}
					currentChunk += sentence
					currentTokenCount += sentenceTokenCount
				} else if sentenceTokenCount > maxTokens {
					// Sentence is too long, split by words
					if currentChunk != "" {
						chunks = append(chunks, currentChunk)
						currentChunk = ""
						currentTokenCount = 0
					}
					
					words := strings.Split(sentence, " ")
					for _, word := range words {
						wordTokens := enc.Encode(word, nil, nil)
						wordTokenCount := len(wordTokens)
						
						if currentTokenCount + wordTokenCount <= maxTokens {
							if currentChunk != "" {
								currentChunk += " "
							}
							currentChunk += word
							currentTokenCount += wordTokenCount
						} else {
							// Save current chunk and start a new one with this word
							if currentChunk != "" {
								chunks = append(chunks, currentChunk)
							}
							currentChunk = word
							currentTokenCount = wordTokenCount
						}
					}
				} else {
					// Start a new chunk with this sentence
					chunks = append(chunks, currentChunk)
					currentChunk = sentence
					currentTokenCount = sentenceTokenCount
				}
			}
		} else {
			// Paragraph doesn't fit, start a new chunk
			chunks = append(chunks, currentChunk)
			currentChunk = paragraph
			currentTokenCount = paragraphTokenCount
		}
	}
	
	// Add the last chunk if there's anything left
	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}
	
	// Create overlapping chunks if we have multiple chunks
	if len(chunks) > 1 {
		return addOverlap(chunks, enc)
	}
	
	return chunks
}

// addOverlap adds overlapping text between chunks for better retrieval
func addOverlap(chunks []string, enc *tiktoken.Tiktoken) []string {
	if len(chunks) <= 1 {
		return chunks
	}
	
	overlapTokens := OverlapSize
	result := make([]string, 0, len(chunks))
	
	for i := 0; i < len(chunks); i++ {
		if i == 0 {
			// First chunk doesn't need leading overlap
			result = append(result, chunks[i])
		} else {
			// Add overlap from previous chunk
			prevChunk := chunks[i-1]
			currentChunk := chunks[i]
			
			// Get end of previous chunk for overlap
			overlapText := getLastNTokens(prevChunk, enc, overlapTokens)
			if overlapText != "" && !strings.Contains(currentChunk, overlapText) {
				currentChunk = overlapText + "\n\n" + currentChunk
			}
			
			result = append(result, currentChunk)
		}
	}
	
	return result
}

// getLastNTokens returns the last N tokens of text as a string
func getLastNTokens(text string, enc *tiktoken.Tiktoken, n int) string {
	tokens := enc.Encode(text, nil, nil)
	if len(tokens) <= n {
		return text
	}
	
	// Get last n tokens
	relevantTokens := tokens[len(tokens)-n:]
	
	// Decode back to string
	return enc.Decode(relevantTokens)
}
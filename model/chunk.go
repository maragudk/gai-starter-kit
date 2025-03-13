package model

import (
	"context"

	"maragu.dev/gai"
)

type embedderFunc = func(ctx context.Context, text string) ([]byte, error)

// Chunk splits document content into chunks with embeddings.
// It uses a fixed size chunker with the specified size and overlap.
func (d Document) Chunk(ctx context.Context, embedder embedderFunc) ([]Chunk, error) {
	tokenizer := &gai.NaiveWordTokenizer{}
	chunker := gai.NewFixedSizeChunker(gai.NewFixedSizeChunkerOptions{
		Tokenizer: tokenizer,
		Size:      256,
		Overlap:   0.2,
	})

	textChunks := chunker.Chunk(ctx, d.Content)
	return createChunksWithEmbeddings(ctx, textChunks, embedder)
}

func createChunksWithEmbeddings(ctx context.Context, textChunks []string, embedder embedderFunc) ([]Chunk, error) {
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

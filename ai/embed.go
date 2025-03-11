package ai

import (
	"context"

	"maragu.dev/gai"
)

// Embed a string to a binary vector.
// See https://huggingface.co/blog/embedding-quantization for details on binary embeddings.
func (c *Client) Embed(ctx context.Context, req gai.EmbedRequest) (gai.EmbedResponse[int], error) {
	res, err := c.embedder.Embed(ctx, req)
	if err != nil {
		return gai.EmbedResponse[int]{}, err
	}

	var embedding []int
	for _, f := range res.Embedding {
		b := 0
		if f > 0 {
			b = 1
		}
		embedding = append(embedding, b)
	}

	return gai.EmbedResponse[int]{Embedding: embedding}, nil
}

var _ gai.Embedder[int] = (*Client)(nil)

package ai

import (
	"context"
	"strings"

	"maragu.dev/gai"

	"app/model"
)

// Embed to a binary vector.
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

// EmbedString as a convenience wrapper around [Client.Embed].
func (c *Client) EmbedString(ctx context.Context, s string) ([]byte, error) {
	res, err := c.Embed(ctx, gai.EmbedRequest{
		Input: strings.NewReader(s),
	})
	if err != nil {
		return nil, err
	}

	return model.QuantizeEmbedding(res.Embedding), nil
}

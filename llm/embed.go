package llm

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
	"maragu.dev/errors"
)

// Embed a string to a binary vector.
// See https://huggingface.co/blog/embedding-quantization for details on binary embeddings.
func (c *Client) Embed(ctx context.Context, text string) ([]int, error) {
	res, err := c.oc.Client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input:          openai.F[openai.EmbeddingNewParamsInputUnion](shared.UnionString(text)),
		EncodingFormat: openai.F(openai.EmbeddingNewParamsEncodingFormatFloat),
		Dimensions:     openai.F(int64(1024)),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error embedding")
	}
	if len(res.Data) == 0 {
		return nil, errors.New("no embedding returned")
	}

	var embedding []int
	for _, f := range res.Data[0].Embedding {
		b := 0
		if f > 0 {
			b = 1
		}
		embedding = append(embedding, b)
	}

	return embedding, nil
}

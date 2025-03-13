package ai

import (
	"bytes"
	"context"
	"encoding/binary"
	"strings"

	"maragu.dev/gai"
)

// Embed to a binary vector.
// See https://huggingface.co/blog/embedding-quantization for details on binary embeddings.
func (c *Client) Embed(ctx context.Context, req gai.EmbedRequest) (gai.EmbedResponse[float32], error) {
	res, err := c.embedder.Embed(ctx, req)
	if err != nil {
		return gai.EmbedResponse[float32]{}, err
	}

	var embedding []float32
	for _, v := range res.Embedding {
		embedding = append(embedding, float32(v))
	}

	return gai.EmbedResponse[float32]{Embedding: embedding}, nil
}

var _ gai.Embedder[float32] = (*Client)(nil)

// EmbedString is a convenience wrapper around [Client.Embed] and [sqlitevec.SerializeEmbedding].
func (c *Client) EmbedString(ctx context.Context, s string) ([]byte, error) {
	res, err := c.Embed(ctx, gai.EmbedRequest{
		Input: strings.NewReader(s),
	})
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, res.Embedding)
	return buf.Bytes(), nil
}

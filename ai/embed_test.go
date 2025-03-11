package ai_test

import (
	"strings"
	"testing"

	"maragu.dev/gai"
	"maragu.dev/is"

	"app/aitest"
)

func TestClient_Embed(t *testing.T) {
	t.Run("can return a binary embedding for a string", func(t *testing.T) {
		c := aitest.NewClient(t)

		res, err := c.Embed(t.Context(), gai.EmbedRequest{Input: strings.NewReader("I'd like to have this text embedded, please.")})
		is.NotError(t, err)
		is.Equal(t, 1024, len(res.Embedding))
		for _, v := range res.Embedding {
			is.True(t, v == 0 || v == 1)
		}
	})
}

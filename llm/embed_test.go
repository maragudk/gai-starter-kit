package llm_test

import (
	"testing"

	"app/llmtest"

	"maragu.dev/is"
)

func TestClient_Embed(t *testing.T) {
	t.Run("can return a binary embedding for a string", func(t *testing.T) {
		c := llmtest.NewEmbedder(t)

		e, err := c.Embed(t.Context(), "I'd like to have this text embedded, please.")
		is.NotError(t, err)
		is.Equal(t, len(e), 1024)
		for _, v := range e {
			is.True(t, v == 0 || v == 1)
		}
	})
}

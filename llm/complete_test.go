package llm_test

import (
	"app/llmtest"
	"testing"

	"maragu.dev/gai"
	"maragu.dev/is"
)

func TestClient_Complete(t *testing.T) {
	t.Run("can complete based on the given messages", func(t *testing.T) {
		c := llmtest.NewCompleter(t)

		messages := []gai.Message{
			gai.NewUserTextMessage("Hello!"),
		}

		res, err := c.Complete(t.Context(), messages)
		is.NotError(t, err)
		is.Equal(t, "Hi", res)
	})
}

package ai_test

import (
	"testing"

	"maragu.dev/gai"
	"maragu.dev/is"

	"app/aitest"
)

func TestClient_ChatComplete(t *testing.T) {
	t.Run("can chat complete based on the given messages", func(t *testing.T) {
		c := aitest.NewClient(t)

		messages := []gai.Message{
			gai.NewUserTextMessage("Hello!"),
		}

		res, err := c.ChatComplete(t.Context(), gai.ChatCompleteRequest{
			Messages:    messages,
			Temperature: gai.Ptr(gai.Temperature(0)),
		})
		is.NotError(t, err)

		var output string
		for part, err := range res.Parts() {
			is.NotError(t, err)
			output += part.Text()
		}

		is.Equal(t, "Hello! How can I assist you today?", output)
	})
}

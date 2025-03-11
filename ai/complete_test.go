package ai_test

import (
	"testing"

	"maragu.dev/gai"
	"maragu.dev/gai/eval"
	"maragu.dev/is"

	"app/aitest"
)

func TestEvalClient_ChatComplete(t *testing.T) {
	eval.Run(t, "can chat complete based on the given messages", func(t *testing.T, e *eval.E) {
		c := aitest.NewClient(t)

		res, err := c.ChatComplete(t.Context(), gai.ChatCompleteRequest{
			Messages: []gai.Message{
				gai.NewUserTextMessage("Hello!"),
			},
			Temperature: gai.Ptr(gai.Temperature(0)),
		})
		is.NotError(t, err)

		var output string
		for part, err := range res.Parts() {
			is.NotError(t, err)
			output += part.Text()
		}

		sample := eval.Sample{
			Input:    "Hello!",
			Output:   output,
			Expected: "Hello! How can I assist you today?",
		}

		result := e.Score(sample, eval.LexicalSimilarityScorer(eval.LevenshteinDistance))

		e.Log(sample, result)
	})
}

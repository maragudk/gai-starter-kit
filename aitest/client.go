// Package aitest provides test helpers for the ai package.
package aitest

import (
	"log/slog"
	"testing"

	"app/ai"
)

func NewClient(t *testing.T) *ai.Client {
	t.Helper()

	return ai.NewClient(ai.NewClientOptions{
		ChatCompleterBaseURL: "http://localhost:8081/v1",
		EmbedderBaseURL:      "http://localhost:8082/v1",
		Log:                  slog.New(slog.NewTextHandler(&testWriter{t: t}, nil)),
	})
}

type testWriter struct {
	t *testing.T
}

func (t *testWriter) Write(p []byte) (n int, err error) {
	t.t.Log(string(p))
	return len(p), nil
}

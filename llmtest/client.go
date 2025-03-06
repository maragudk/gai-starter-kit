// Package llmtest provides test helpers for the llm package.
package llmtest

import (
	"log/slog"
	"testing"

	"app/llm"
)

func NewClient(t *testing.T) *llm.Client {
	t.Helper()

	return llm.NewClient(llm.NewClientOptions{
		BaseURL: "http://localhost:8082/v1",
		Log:     slog.New(slog.NewTextHandler(&testWriter{t: t}, nil)),
	})
}

type testWriter struct {
	t *testing.T
}

func (t *testWriter) Write(p []byte) (n int, err error) {
	t.t.Log(string(p))
	return len(p), nil
}

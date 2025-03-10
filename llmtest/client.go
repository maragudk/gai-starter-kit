// Package llmtest provides test helpers for the llm package.
package llmtest

import (
	"log/slog"
	"testing"

	"maragu.dev/env"

	"app/llm"
)

func NewCompleter(t *testing.T) *llm.Client {
	t.Helper()

	_ = env.Load("../.env.test.local")

	return llm.NewClient(llm.NewClientOptions{
		BaseURL: "http://localhost:8081/v1",
		Log:     slog.New(slog.NewTextHandler(&testWriter{t: t}, nil)),
		Key:     env.GetStringOrDefault("OPENAI_KEY", ""),
	})
}

func NewEmbedder(t *testing.T) *llm.Client {
	t.Helper()

	_ = env.Load("../.env.test.local")

	return llm.NewClient(llm.NewClientOptions{
		BaseURL: "http://localhost:8082/v1",
		Log:     slog.New(slog.NewTextHandler(&testWriter{t: t}, nil)),
		Key:     env.GetStringOrDefault("OPENAI_KEY", ""),
	})
}

type testWriter struct {
	t *testing.T
}

func (t *testWriter) Write(p []byte) (n int, err error) {
	t.t.Log(string(p))
	return len(p), nil
}

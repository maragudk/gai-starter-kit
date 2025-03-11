package ai

import (
	"log/slog"

	"maragu.dev/gai"
	openai "maragu.dev/gai-openai"
)

// Client wraps both a [gai.ChatCompleter] and a [gai.Embedder].
type Client struct {
	chatCompleter gai.ChatCompleter
	embedder      gai.Embedder[float64]
	log           *slog.Logger
}

type NewClientOptions struct {
	ChatCompleterBaseURL string
	EmbedderBaseURL      string
	Log                  *slog.Logger
}

func NewClient(opts NewClientOptions) *Client {
	if opts.Log == nil {
		opts.Log = slog.New(slog.DiscardHandler)
	}

	c := openai.NewClient(openai.NewClientOptions{
		BaseURL: opts.ChatCompleterBaseURL,
		Log:     opts.Log,
	})

	cc := c.NewChatCompleter(openai.NewChatCompleterOptions{
		Model: openai.ChatCompleteModel("llama3"),
	})

	c = openai.NewClient(openai.NewClientOptions{
		BaseURL: opts.EmbedderBaseURL,
		Log:     opts.Log,
	})

	e := c.NewEmbedder(openai.NewEmbedderOptions{
		Dimensions: 1024,
		Model:      openai.EmbedModel("mxbai-embed-large-v1-f16"),
	})

	return &Client{
		chatCompleter: cc,
		embedder:      e,
		log:           opts.Log,
	}
}

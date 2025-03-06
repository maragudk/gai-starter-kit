package llm

import (
	"log/slog"

	"maragu.dev/gai"
)

type Client struct {
	client *gai.OpenAIClient
	log    *slog.Logger
}

type NewClientOptions struct {
	BaseURL string
	Key     string
	Log     *slog.Logger
}

func NewClient(opts NewClientOptions) *Client {
	if opts.Log == nil {
		opts.Log = slog.New(slog.DiscardHandler)
	}

	oc := gai.NewOpenAIClient(gai.NewOpenAIClientOptions{
		BaseURL: opts.BaseURL,
		Key:     opts.Key,
		Log:     opts.Log,
	})

	return &Client{
		client: oc,
		log:    opts.Log,
	}
}

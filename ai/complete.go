package ai

import (
	"context"

	"maragu.dev/gai"
)

func (c *Client) ChatComplete(ctx context.Context, req gai.ChatCompleteRequest) (gai.ChatCompleteResponse, error) {
	return c.chatCompleter.ChatComplete(ctx, req)
}

var _ gai.ChatCompleter = (*Client)(nil)

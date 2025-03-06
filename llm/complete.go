package llm

import (
	"context"

	"github.com/openai/openai-go"
	"maragu.dev/gai"
)

func (c *Client) Complete(ctx context.Context, messages []gai.Message) (string, error) {
	res, err := c.client.Client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("Only say 'Hi', nothing more."),
			openai.UserMessage("Hi"),
		}),
		Model:       openai.F(openai.ChatModelGPT4oMini),
		Temperature: openai.F(0.0),
	})
	if err != nil {
		return "", err
	}
	return res.Choices[0].Message.Content, nil
}

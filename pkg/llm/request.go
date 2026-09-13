package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

var ErrEmptyResponse = errors.New("empty response")

type RequestOptions struct {
	model    Model
	messages []Message
}

type RequestOption func(*RequestOptions)

func WithModel(model Model) RequestOption {
	return func(opts *RequestOptions) {
		opts.model = model
	}
}

func WithSystemMessage(content string) RequestOption {
	return WithMessage(Message{
		Role:    RoleSystem,
		Content: content,
	})
}

func WithUserMessage(content string) RequestOption {
	return WithMessage(Message{
		Role:    RoleUser,
		Content: content,
	})
}

func WithMessage(msg Message) RequestOption {
	return func(opts *RequestOptions) {
		opts.messages = append(opts.messages, msg)
	}
}

func (c *Client) _Request(ctx context.Context, message string, opts ...RequestOption) (io.Reader, error) {
	opts = slices.Concat(opts, []RequestOption{WithUserMessage(message)})

	o := &RequestOptions{
		model: c.cfg.Model,
	}

	for _, opt := range opts {
		opt(o)
	}

	return c.request(ctx, Request{
		Model:    o.model,
		Messages: o.messages,
	})
}

func (c *Client) Request(ctx context.Context, req Request) (io.Reader, error) {
	return c.request(ctx, req)
}

func (c *Client) request(ctx context.Context, req Request) (io.Reader, error) {
	if req.Model == "" {
		req.Model = c.cfg.Model
	}

	res, err := c.sdk.Chat.Completions.New(ctx, req.ChatCompletionNewParams())
	if err != nil {
		return nil, err
	}

	if len(res.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices", ErrEmptyResponse)
	}

	content := res.Choices[0].Message.Content
	if content == "" {
		return nil, fmt.Errorf("%w: no content", ErrEmptyResponse)
	}

	return strings.NewReader(content), nil
}

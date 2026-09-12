package llm

import (
	"encoding/json"

	"github.com/openai/openai-go/v3"
)

type Model string

const (
	DeepSeekFlash Model = "deepseek-flash"
	DeepSeekV4Pro Model = "deepseek-v4-pro"
)

type Role string

const (
	RoleSystem Role = "system"
	RoleUser   Role = "user"
)

type Message struct {
	Role    Role
	Content string
}

func UserMessage(content string) Message {
	return Message{
		Role:    RoleUser,
		Content: content,
	}
}

func UserJSONMessage(content any) Message {
	data, _ := json.Marshal(content)

	return Message{
		Role:    RoleUser,
		Content: string(data),
	}
}

func SystemMessage(content string) Message {
	return Message{
		Role:    RoleSystem,
		Content: content,
	}
}

func SystemFile(content string) Message {
	return Message{
		Role:    RoleSystem,
		Content: content,
	}
}

type Request struct {
	Model    Model
	Messages []Message
}

func (r Request) ChatCompletionNewParams() openai.ChatCompletionNewParams {
	res := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(r.Model),
		Messages: make([]openai.ChatCompletionMessageParamUnion, len(r.Messages)),
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
		},
	}

	for i, msg := range r.Messages {
		switch msg.Role {
		case RoleSystem:
			res.Messages[i] = openai.SystemMessage(msg.Content)
		case RoleUser:
			res.Messages[i] = openai.UserMessage(msg.Content)
		default:
			res.Messages[i] = openai.UserMessage(msg.Content)
		}
	}

	return res
}

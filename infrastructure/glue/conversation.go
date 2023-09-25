package glue

import (
	"github.com/jborkows/openai/model"
	"github.com/jborkows/openai/openai"
)

var repository model.ConversationRepository = model.NewInMemoryConversationRepository()

func NewConversation() *model.Conversation {
	return repository.CreateConversation(new(openAI3_5Connection))
}
func GetConversation(id string) (*model.Conversation, error) {
	return repository.GetConversation(id)
}

type openAI3_5Connection struct {
}

func (c *openAI3_5Connection) Send(history []model.Message, question string) (*model.Sender[model.Message], error) {
	if len(history) == 0 {
		history = []model.Message{
			openai.SystemMessage("You are a helpful assistant."),
		}
	}
	output, err := openai.Dialog(openai.DialogInput{
		History:  history,
		Question: question,
		Model:    "gpt-3.5-turbo",
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (c *openAI3_5Connection) ID() string {
	return "ChatGpt 3.5"
}

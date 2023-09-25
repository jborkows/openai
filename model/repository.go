package model

import (
	"fmt"
)

type ConversationRepository interface {
	Save(conversation *Conversation) (error error)
	CreateConversation(ConversationConnection ConversationConnection) *Conversation
	GetConversation(id string) (*Conversation, error)
}

type InMemoryConversationRepository struct {
	conversations *SafeMap[string, *Conversation]
}

func NewInMemoryConversationRepository() *InMemoryConversationRepository {
	return &InMemoryConversationRepository{
		conversations: NewSafeMap[string, *Conversation](),
	}
}

var counter = AtomicCounter{counter: 0}

func (self *InMemoryConversationRepository) CreateConversation(ConversationConnection ConversationConnection) *Conversation {
	ID := fmt.Sprintf("%d", counter.increment())

	Conversation := Conversation{
		ID:                     ID,
		Messages:               []Message{},
		ConversationConnection: ConversationConnection,
	}

	self.conversations.Put(ID, &Conversation)
	return &Conversation
}

func (repository *InMemoryConversationRepository) Save(conversation *Conversation) (error error) {
	repository.conversations.Put(conversation.ID, conversation)
	return
}

func (repository *InMemoryConversationRepository) GetConversation(id string) (*Conversation, error) {
	conversation, ok := repository.conversations.Get(id)
	if !ok {
		return nil, fmt.Errorf("Conversation %s not found", id)

	}
	return conversation, nil
}

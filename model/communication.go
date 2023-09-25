package model

import (
	"fmt"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ConversationConnection interface {
	Send(history []Message, question string) (*Sender[Message], error)
	ID() string
}

type Conversation struct {
	ID                     string    `json:"id"`
	Messages               []Message `json:"messages"`
	Listeners              []*Receiver[Message]
	ConversationConnection ConversationConnection
}

func (c *Conversation) AddListener(listener *Receiver[Message]) {
	c.Listeners = append(c.Listeners, listener)
}
func (c *Conversation) RemoveListener(listener *Receiver[Message]) {
	for index, value := range c.Listeners {
		if value == listener {
			c.Listeners = append(c.Listeners[:index], c.Listeners[index+1:]...)
		}
	}
}

func (c *Conversation) WaitForListeners() {
	for {
		select {
		case <-time.After(200 * time.Millisecond):
			if len(c.Listeners) > 0 {
				fmt.Printf("Found %d listeners\n", len(c.Listeners))
				return
			}
		}
	}
}

func (c *Conversation) Send(question string) error {
	output, error := c.ConversationConnection.Send(c.Messages, question)
	if error != nil {
		fmt.Printf("Error sending message: %s", error)
		return error
	}
	buffer := ""
	go func() {
		c.WaitForListeners()
		for output.Channel != nil {
			select {
			case value, ok := <-output.Channel:
				if !ok {
					fmt.Printf("Channel closed!")
					output.Channel = nil
					c.Messages = append(c.Messages, Message{
						Role:    value.Role,
						Content: buffer,
					})

					fmt.Printf("Messages after appending: %v\n", c.Messages)
					buffer = ""
				} else {
					fmt.Printf("Sending message to listeners: %v\n", value)
					for _, listener := range c.Listeners {
						listener.Channel <- value
					}

					buffer = buffer + value.Content
				}
			}
		}
	}()

	return nil
}

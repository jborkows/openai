package openai

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/jborkows/openai/model"
)

type Delta struct {
	Content string `json:"content"`
}

type Choice struct {
	Index        int    `json:"index"`
	Delta        Delta  `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type Event struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type ChatRequest struct {
	Model    string    `json:"model"`
	Stream   bool      `json:"stream"`
	Messages []Message `json:"messages"`
}

func SystemMessage(message string) model.Message {
	return model.Message{
		Role:    "system",
		Content: message,
	}
}

func UserMessage(message string) model.Message {
	return model.Message{
		Role:    "user",
		Content: message,
	}
}

type DialogInput struct {
	History  []model.Message
	Question string
	Model    string
}

func Dialog(input DialogInput) (*model.Sender[model.Message], error) {
	messages := model.MapOver(
		append(input.History, UserMessage(input.Question)),
		func(message model.Message) Message {
			return Message{
				Role:    message.Role,
				Content: message.Content,
			}
		})
	request := ChatRequest{
		Model:    input.Model,
		Stream:   true,
		Messages: messages,
	}
	// request to byte array
	marshalled, err := json.Marshal(request)
	if err != nil {
		log.Printf("Error marshalling request: %s", err)
		return nil, err
	}

	resp, error := callOpenAI(url("chat/completions"), marshalled, &CallOpenAIConfig{LogResponseImmediatelly: false})
	if error != nil {
		log.Printf("Error calling OpenAI %s", error)
		return nil, error
	}

	receiver := make(chan model.Message)
	go func() {
		defer resp.Body.Close()
		defer close(receiver)

		finished := false
		// Copy SSE events from the other service to the client
		for !finished {
			select {
			default:
				// Read an SSE event from the other service
				// log.Printf("Read SSE event")
				eventBytes := make([]byte, 1024*1024*10) // Adjust the buffer size as needed
				n, err := resp.Body.Read(eventBytes)
				// log.Printf("Read content: '%s'", eventBytes[:n])
				// log.Printf("Read SSE event - finished")
				//Remember SSE events are splitted by /n/n
				concatanatedEvents := ""
				for _, event := range strings.Split(string(eventBytes[:n]), "\n\n") {
					if event == "" {
						continue
					}
					fmt.Printf("event = '%s'", event)
					prefix := "data: "
					if strings.HasPrefix(event, prefix) {
						event = strings.TrimPrefix(event, prefix)
						// Process the result
					}

					if event == "[DONE]" {
						log.Printf("Done")
						finished = true
						break
					}

					var chatCompletion Event
					if err := json.Unmarshal([]byte(event), &chatCompletion); err != nil {
						log.Printf("Error unmarshalling event: %s", err)
						return
					}
					if len(chatCompletion.Choices) == 0 {
						log.Printf("No choices in event")
						return
					}

					if chatCompletion.Choices[0].FinishReason == "stop" {
						log.Printf("Stop reason")
						finished = true
					}
					answer := chatCompletion.Choices[0].Delta.Content

					if err != nil {
						log.Printf("Error reading event: %s", err)
						return
					}
					concatanatedEvents += answer
				}

				log.Printf("Sum up Answer: '%s'", concatanatedEvents)
				receiver <- SystemMessage(concatanatedEvents)
			}
		}
	}()
	return &model.Sender[model.Message]{Channel: receiver}, nil
}
func Question(message string, output chan<- string) {
	// response := response_wrapper(output)
	response := output

	log.Printf("OpenAIHandler stream called with question %s", message)
	payload := []byte(`{
		"model": "gpt-3.5-turbo",
		"stream": true,
		"messages": [
			{
				"role": "system",
				"content": "You are a helpful assistant."
			},
			{
				"role": "user",
				"content": "` + message + `"
			}
		]
	}`)

	resp, error := callOpenAI(url("chat/completions"), payload, &CallOpenAIConfig{LogResponseImmediatelly: false})
	if error != nil {
		log.Printf("Error calling OpenAI %s", error)
		response <- "Error calling OpenAI"
	}
	defer resp.Body.Close()

	finished := false
	// Copy SSE events from the other service to the client
	for !finished {
		select {
		default:
			// Read an SSE event from the other service
			log.Printf("Read SSE event")
			eventBytes := make([]byte, 1024*1024*10) // Adjust the buffer size as needed
			n, err := resp.Body.Read(eventBytes)
			log.Printf("Read content: '%s'", eventBytes[:n])
			log.Printf("Read SSE event - finished")
			//TODO rember SSE events are splitted by /n/n
			concatanatedEvents := ""
			for _, event := range strings.Split(string(eventBytes[:n]), "\n\n") {
				if event == "" {
					continue
				}
				fmt.Printf("event = '%s'", event)
				prefix := "data: "
				if strings.HasPrefix(event, prefix) {
					event = strings.TrimPrefix(event, prefix)
					// Process the result
				}

				if event == "[DONE]" {
					log.Printf("Done")
					finished = true
					break
				}

				var chatCompletion Event
				if err := json.Unmarshal([]byte(event), &chatCompletion); err != nil {
					log.Printf("Error unmarshalling event: %s", err)
					return
				}
				if len(chatCompletion.Choices) == 0 {
					log.Printf("No choices in event")
					return
				}

				if chatCompletion.Choices[0].FinishReason == "stop" {
					log.Printf("Stop reason")
					finished = true
				}
				answer := chatCompletion.Choices[0].Delta.Content

				// log.Println("############################")

				if err != nil {
					log.Printf("Error reading event: %s", err)
					return
				}
				// log.Printf("Answer: '%s'", answer)
				concatanatedEvents += answer
			}

			log.Printf("Sum up Answer: '%s'", concatanatedEvents)
			response <- concatanatedEvents
		}
	}
}

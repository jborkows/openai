package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

func callOpenAI2(url string, payload []byte) (*http.Response, error) {

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{
		// Transport: &LoggingTransport{Transport: http.DefaultTransport},
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	return resp, nil
}
func Question(message string, response chan<- string) {
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

	resp, error := callOpenAI2(url("chat/completions"), payload)
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

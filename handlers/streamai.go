package handlers

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

func OpenAIStreamHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("OpenAIHandler stream called with question %s", question)
	url := "https://api.openai.com/v1/chat/completions"

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Enable CORS for testing
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
				"content": "` + question + `"
			}
		]
	}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		http.Error(w, "Error connecting to the other service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy SSE events from the other service to the client
	for {
		select {
		case <-r.Context().Done():
			// Client disconnected, stop copying events
			fmt.Println("Client disconnected")
			return
		default:
			// Read an SSE event from the other service
			eventBytes := make([]byte, 4048) // Adjust the buffer size as needed
			n, err := resp.Body.Read(eventBytes)
			//TODO rember SSE events are splitted by /n/n
			log.Printf("Read content: %s", eventBytes[:n])
			log.Println("############################")

			if err != nil {
				fmt.Println("Error reading event from the other service:", err)
				return
			}

			text := string(eventBytes[:n])
			prefix := "data: "
			if strings.HasPrefix(text, prefix) {
				text = strings.TrimPrefix(text, prefix)
				// Process the result
			}
			eventName := "streamai"
			var chatCompletion Event
			if err := json.Unmarshal([]byte(text), &chatCompletion); err != nil {
				fmt.Println("Error unmarshalling event:", err)

				// Write the data as an SSE event
				//In this version new line charactrs are not allowed in data

				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventName, "Cannot unmarshal event")
				return
			}
			if len(chatCompletion.Choices) == 0 {
				fmt.Println("No choices in event")
				return
			}
			answer := chatCompletion.Choices[0].Delta.Content
			data := fmt.Sprintf("<div>%s</div>", answer)
			// Write the data as an SSE event

			log.Printf("Sent event %s with data %s", eventName, data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventName, data)

			// Flush the response writer to ensure the event is sent immediately
			w.(http.Flusher).Flush()
			if chatCompletion.Choices[0].FinishReason != "" {
				return
			} else {
				log.Printf("Sent event %s with data %s", "answer", data)
			}
		}
	}
}

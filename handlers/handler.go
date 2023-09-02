package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/jborkows/openai/metrics"
)

type PageVersion struct {
	Version string
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	version := time.Now().Format("20060102150405")
	// write variable version with current timestamp value as string
	pageVersion := PageVersion{Version: version}

	t, _ := template.ParseFiles("templates/index.html")

	aFailure := t.Execute(w, pageVersion)
	if aFailure != nil {
		metrics.IncrementHTTPRequestsTotal("500", "GET")
	}

	metrics.IncrementHTTPRequestsTotal("200", "GET")

}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	// Handle the dynamic update (you can customize this part)
	w.Write([]byte("This is the dynamic content!"))
	metrics.IncrementHTTPRequestsTotal("200", "POST")
}
func ClearExampleSSE(w http.ResponseWriter, r *http.Request) {
	// Handle the dynamic update (you can customize this part)
	w.Write([]byte(""))
	metrics.IncrementHTTPRequestsTotal("200", "POST")
}

func ExampleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Enable CORS for testing

	flusher, _ := w.(http.Flusher)
	log.Printf("ExampleSSE called")
	counter := 0
	for {
		// Generate some data to send to the client
		data := fmt.Sprintf("<div>Message from server at %s</div>", time.Now().Format(time.RFC1123))
		eventName := "exampleMessage"
		// Write the data as an SSE event
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventName, data)
		log.Printf("Sent event %s with data %s", eventName, data)
		flusher.Flush()
		counter = counter + 1
		if counter > 6 {
			return
		}
		time.Sleep(2 * time.Second) // Wait for a while before sending the next event
	}
}

// lame...
var question = "What is the meaning of life?"

func ReadQuestionHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	question = r.Form.Get("question")

	t, _ := template.ParseFiles("templates/openai.html")

	aFailure := t.Execute(w, nil)
	if aFailure != nil {
		metrics.IncrementHTTPRequestsTotal("500", "GET")
	}

	metrics.IncrementHTTPRequestsTotal("200", "GET")
}

var apiKey string

func init() {
	apiKey = os.Getenv("OPEN_API_KEY")
	if apiKey == "" {
		log.Fatal("OPEN_API_KEY not set")
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type XChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatCompletion struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []XChoice `json:"choices"`
	Usage   Usage     `json:"usage"`
}

func OpenAIHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("OpenAIHandler called with question %s", question)
	url := "https://api.openai.com/v1/chat/completions"

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Enable CORS for testing
	payload := []byte(`{
		"model": "gpt-3.5-turbo",
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
			eventBytes := make([]byte, 2048) // Adjust the buffer size as needed
			n, err := resp.Body.Read(eventBytes)
			log.Printf("Read content: %s", eventBytes[:n])
			if err != nil {
				fmt.Println("Error reading event from the other service:", err)
				return
			}

			var chatCompletion ChatCompletion
			text := string(eventBytes[:n])
			eventName := "aimessage"
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
			answer := chatCompletion.Choices[0].Message.Content
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

package handlers

import (
	"fmt"
	"log"
	"net/http"
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

func ExampleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Enable CORS for testing

	flusher, _ := w.(http.Flusher)

	for {
		// Generate some data to send to the client
		data := fmt.Sprintf("<div>Message from server at %s</div>", time.Now().Format(time.RFC1123))
		eventName := "exampleMessage"
		// Write the data as an SSE event
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventName, data)
		log.Printf("Sent event %s with data %s", eventName, data)
		flusher.Flush()

		time.Sleep(2 * time.Second) // Wait for a while before sending the next event
	}
}

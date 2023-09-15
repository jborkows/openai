package pages

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jborkows/openai/metrics"
	"github.com/jborkows/openai/openai"
)

var communication = make(chan string, 5000)

func Chats(rw http.ResponseWriter, req *http.Request) {

	log.Printf("HAHAHA %s %s", req.Method, req.URL.Path)
	if req.Method == "GET" {
		if strings.Contains(req.URL.Path, "sse") {
			sse(rw, communication)
		} else {

			RenderSite(rw, req, &ContentDefinition{templateFile: "chats.html"})
		}
	} else if req.Method == "POST" {
		req.ParseForm()
		message := req.Form.Get("message")
		log.Printf("Message: %s", message)
		fmt.Fprint(rw, "<input type=\"text\" name=\"message\" /><button type=\"submit\">Send</button>")
		close(communication)
		communication = make(chan string, 5000)
		openai.Question(message, communication)
	} else {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		metrics.ReportFailure(req)
	}
}

func sse(response http.ResponseWriter, ch <-chan string) {

	response.Header().Set("Content-Type", "text/event-stream")
	response.Header().Set("Cache-Control", "no-cache")
	response.Header().Set("Connection", "keep-alive")
	response.Header().Set("Access-Control-Allow-Origin", "*") // Enable CORS for testing
	//execute code every 1 second
	ticker := time.NewTicker(1 * time.Second)
	counter := 0
	for {
		select {
		case <-ticker.C:
			//send message to client
			counter++
			// fmt.Fprintf(response, "data: %d\n\n", counter)
			// response.(http.Flusher).Flush()
			//TODO replace communication with channel or object which contains channel
		case value, _ := <-communication:
			// case value, ok := <-ch:
			// 	if !ok {
			// 		log.Println("Channel closed!")
			// 		ch = nil
			// 		continue
			// 	}
			log.Printf("Sending message: %s", value)
			value = strings.ReplaceAll(value, "\r\n", "<br>")
			value = strings.ReplaceAll(value, "\n", "<br>")
			value = strings.ReplaceAll(value, " ", "&nbsp;")
			// value = strings.ReplaceAll(value, "\t", "&nbsp;")
			fmt.Fprintf(response, "data: %s\n\n", value)
			response.(http.Flusher).Flush()
		}
	}

}

package pages

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jborkows/openai/infrastructure/glue"
	"github.com/jborkows/openai/metrics"
	"github.com/jborkows/openai/model"
)

// TODO: make it thread safe
var communication = make(chan string, 5000)

func Chats(rw http.ResponseWriter, req *http.Request) {

	log.Printf("HAHAHA %s %s", req.Method, req.URL.Path)
	if req.URL.Path == "/chats" {
		root(req, rw)
	} else if strings.HasPrefix(req.URL.Path, "/chats/") {
		chat(req, rw)
	} else {
		http.Error(rw, "Not found", http.StatusNotFound)
		metrics.ReportFailure(req)
	}
}

func useConversation(chatID string, req *http.Request, rw http.ResponseWriter, f func(conversation *model.Conversation)) {
	conversation, error := glue.GetConversation(chatID)
	if error != nil {
		log.Printf("Error: %s", error)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		metrics.ReportFailure(req)
		return
	}
	if conversation == nil {
		http.Error(rw, "Not found", http.StatusNotFound)
		metrics.ReportFailure(req)
	} else {
		f(conversation)
	}
}

func chat(req *http.Request, rw http.ResponseWriter) {

	if strings.Contains(req.URL.Path, "sse") {
		chatID := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/chats/"), "/sse")
		useConversation(chatID, req, rw, func(conversation *model.Conversation) {
			sse(rw, conversation)
		})
	} else {
		chatID := strings.TrimPrefix(req.URL.Path, "/chats/")
		useConversation(chatID, req, rw, func(conversation *model.Conversation) {
			RenderSite(rw, req, &ContentDefinition{templateFile: "chat.html", templateData: conversation})
			metrics.ReportSuccess(req)
		})
	}
}

type HXLocationRedirect struct {
	Path   string `json:"path"`
	Target string `json:"target"`
}

func root(req *http.Request, rw http.ResponseWriter) {
	if req.Method == "GET" {
		RenderSite(rw, req, &ContentDefinition{templateFile: "chats.html"})
		metrics.ReportSuccess(req)
	} else if req.Method == "POST" {
		req.ParseForm()
		message := req.Form.Get("message")
		log.Printf("Message: %s", message)
		conversation := glue.NewConversation()

		if req.Header.Get("HX-Request") == "true" {
			location := HXLocationRedirect{
				Path:   fmt.Sprintf("/chats/%s", conversation.ID),
				Target: "main",
			}
			locationJson, error := json.Marshal(location)
			if error != nil {
				log.Printf("Error: %s", error)
				http.Error(rw, "Internal server error", http.StatusInternalServerError)
				metrics.ReportFailure(req)
				return
			}
			rw.Header().Set("HX-Location", string(locationJson))
			rw.WriteHeader(http.StatusCreated)

			err := conversation.Send(message)
			if err != nil {
				log.Printf("Error: %s", err)
				http.Error(rw, "Internal server error", http.StatusInternalServerError)
				metrics.ReportFailure(req)
			}
			// } else {
			// 	http.Redirect(rw, req, fmt.Sprintf("/chats/%s", conversation.ID), http.StatusFound)
		} else {
			http.ResponseWriter(rw).WriteHeader(http.StatusCreated)

			http.Redirect(rw, req, fmt.Sprintf("/chats/%s", conversation.ID), http.StatusFound)
		}
		// }
	} else {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		metrics.ReportFailure(req)
	}
}

func sse(response http.ResponseWriter, conversation *model.Conversation) {

	response.Header().Set("Content-Type", "text/event-stream")
	response.Header().Set("Cache-Control", "no-cache")
	response.Header().Set("Connection", "keep-alive")
	response.Header().Set("Access-Control-Allow-Origin", "*") // Enable CORS for testing
	receiver := make(chan model.Message)
	wrapper := model.Receiver[model.Message]{Channel: receiver}
	conversation.AddListener(&wrapper)
	defer conversation.RemoveListener(&wrapper)
	defer close(receiver)
	for {
		select {
		case message, ok := <-receiver:
			if !ok {
				return
			}
			value := message.Content
			log.Printf("[Page] Sending message: %v", value)
			value = strings.ReplaceAll(value, "\r\n", "<br>")
			value = strings.ReplaceAll(value, "\n", "<br>")
			value = strings.ReplaceAll(value, " ", "&nbsp;")
			value = strings.ReplaceAll(value, "\t", "&nbsp;&nbsp;&nbsp;&nbsp;")
			fmt.Fprintf(response, "data: %s\n\n", value)
			response.(http.Flusher).Flush()
		}
	}

}

package infrastructure

import (
	"net/http"

	"github.com/jborkows/openai/infrastructure/pages"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRouting() {

	// http.HandleFunc("/", handlers.IndexHandler)
	// http.HandleFunc("/update", handlers.UpdateHandler)
	// http.HandleFunc("/exampleSSE", handlers.ExampleSSE)
	// http.HandleFunc("/clear", handlers.ClearExampleSSE)
	// http.HandleFunc("/question", handlers.ReadQuestionHandler)
	// http.HandleFunc("/ai", handlers.OpenAIHandler)
	// http.HandleFunc("/aistream", handlers.OpenAIStreamHandler)

	http.HandleFunc("/", pages.Home)
	http.HandleFunc("/images", pages.Image)
	http.HandleFunc("/chats", pages.Chats)
	http.HandleFunc("/chats/", pages.Chats)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Set up Prometheus metrics

	http.Handle("/metrics", promhttp.Handler())
}

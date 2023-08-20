// main.go
package main

import (
	"net/http"

	"github.com/jborkows/openai/handlers"

	_ "github.com/jborkows/openai/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Set up the routes
	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/update", handlers.UpdateHandler)
	http.HandleFunc("/example", handlers.ExampleSSE)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Set up Prometheus metrics

	http.Handle("/metrics", promhttp.Handler())

	// Start the server
	http.ListenAndServe(":8080", nil)
}

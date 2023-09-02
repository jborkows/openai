// main.go
package main

import (
	"net/http"

	"github.com/jborkows/openai/infrastructure"
	_ "github.com/jborkows/openai/metrics"
)

func main() {

	infrastructure.RegisterRouting()

	// Start the server
	http.ListenAndServe(":8080", nil)
}

package handlers

import (
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

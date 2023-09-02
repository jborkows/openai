package pages

import (
	"encoding/base64"
	"net/http"
	"text/template"

	htmlTemplate "html/template"

	"github.com/jborkows/openai/metrics"
	"github.com/jborkows/openai/openai"
)

func Image(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Method == "GET" {
		RenderSite(httpResponseWriter, httpRequest, &ContentDefinition{templateFile: "image.html"})
	} else if httpRequest.Method == "POST" {
		generateImage(httpResponseWriter, httpRequest)
	} else {
		http.Error(httpResponseWriter, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func generateImage(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {

	err := httpRequest.ParseForm()
	if err != nil {
		http.Error(httpResponseWriter, "Error parsing form data", http.StatusBadRequest)
		metrics.ReportFailure(httpRequest)
		return
	}
	prompt := httpRequest.Form.Get("prompt")

	data, error := openai.GenerateImage(prompt)
	if error != nil {
		http.Error(httpResponseWriter, error.Error(), http.StatusInternalServerError)
		metrics.ReportFailure(httpRequest)
		return
	}
	pageTemplate := template.Must(template.ParseFiles("templates/partials/generated_image.html"))

	// Prepare data for the template
	pageData := struct {
		ImageSource htmlTemplate.HTML
		Prompt      string
	}{
		ImageSource: htmlTemplate.HTML(base64.StdEncoding.EncodeToString(data)),
		Prompt:      prompt,
	}

	// Execute the template and send it as an HTTP response
	if err := pageTemplate.Execute(httpResponseWriter, pageData); err != nil {
		http.Error(httpResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	} else {
		metrics.ReportSuccess(httpRequest)
	}
}

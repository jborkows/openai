package pages

import (
	"bytes"
	"html/template"
	"io"
	"net/http"

	"github.com/jborkows/openai/metrics"
)

type ContentDefinition struct {
	templateFile string
	templateData interface{}
}

func RenderSite(response http.ResponseWriter, request *http.Request, content *ContentDefinition) {

	if request.Header.Get("myPartial") == "true" {
		fragment(content, request, response)
	} else {
		fullPage(content, request, response)
	}

	// fullPage(content, request, response)
}

func fullPage(content *ContentDefinition, request *http.Request, response http.ResponseWriter) {
	data := struct {
		Content template.HTML
		Version string
	}{Version: version()}
	Content, err := renderTemplate(content)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		metrics.ReportFailure(request)
		return
	}
	data.Content = Content
	layout := template.Must(template.ParseFiles("templates/layout.html"))

	if err := layout.Execute(response, data); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		metrics.ReportFailure(request)
		return
	} else {
		metrics.ReportSuccess(request)
	}
}

func fragment(content *ContentDefinition, request *http.Request, response http.ResponseWriter) {
	if err := returnTemplate(content, response); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		metrics.ReportFailure(request)
		return
	} else {
		metrics.ReportSuccess(request)
	}
}

func renderTemplate(definition *ContentDefinition) (template.HTML, error) {
	var contentBuffer bytes.Buffer
	if err := useTemplate(definition, &contentBuffer); err != nil {
		return template.HTML(""), err
	} else {
		return template.HTML(contentBuffer.String()), nil
	}
}

func useTemplate(definition *ContentDefinition, writer io.Writer) error {
	contentTemplate := template.Must(template.ParseFiles("templates/" + definition.templateFile))
	if err := contentTemplate.Execute(writer, definition.templateData); err != nil {
		return err
	}
	return nil
}
func returnTemplate(definition *ContentDefinition, response http.ResponseWriter) error {
	return useTemplate(definition, response)
}

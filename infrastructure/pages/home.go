package pages

import "net/http"

func Home(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
	RenderSite(httpResponseWriter, httpRequest, &ContentDefinition{templateFile: "home.html"})
}

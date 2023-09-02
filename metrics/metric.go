package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"code", "method"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
}

func IncrementHTTPRequestsTotal(code, method string) {
	httpRequestsTotal.WithLabelValues(code, method).Inc()
}

func ReportSuccess(request *http.Request) {
	httpRequestsTotal.WithLabelValues("200", request.Method).Inc()
}

func ReportFailure(request *http.Request) {
	httpRequestsTotal.WithLabelValues("500", request.Method).Inc()
}

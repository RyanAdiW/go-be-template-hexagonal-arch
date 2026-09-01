package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type HTTPMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func NewHTTPMetrics(registry prometheus.Registerer) *HTTPMetrics {
	m := &HTTPMetrics{requests: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "http_requests_total", Help: "HTTP requests."}, []string{"method", "path", "status"}), duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "HTTP request duration."}, []string{"method", "path"})}
	registry.MustRegister(m.requests, m.duration)
	return m
}
func (m *HTTPMetrics) Observe(method, path string, status int, duration time.Duration) {
	if m == nil {
		return
	}
	m.requests.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
	m.duration.WithLabelValues(method, path).Observe(duration.Seconds())
}

package sis

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// latencyBuckets cover the range the service actually works in: from tens of
// microseconds (in-memory, warm pool) up to the DB timeout.
var latencyBuckets = []float64{
	.0001, .00025, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5,
}

// metrics holds the service's Prometheus collectors in a dedicated registry.
type metrics struct {
	registry *prometheus.Registry

	requests        *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	storeDuration   *prometheus.HistogramVec
	legacyGets      prometheus.Counter
}

// collectorProvider is implemented by stores that expose their own metrics
// (e.g. connection pool statistics).
type collectorProvider interface {
	Collector() prometheus.Collector
}

func newMetrics(store Store) *metrics {
	m := &metrics{
		registry: prometheus.NewRegistry(),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "sis_requests_total",
			Help: "Subscriber API requests by method and response status.",
		}, []string{"method", "code"}),
		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "sis_request_duration_seconds",
			Help:    "Subscriber API request duration, including validation and the store call.",
			Buckets: latencyBuckets,
		}, []string{"method", "code"}),
		storeDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "sis_store_duration_seconds",
			Help:    "Duration of store calls (for PostgreSQL: pool acquire + query).",
			Buckets: latencyBuckets,
		}, []string{"op", "result"}),
		legacyGets: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "subscribers_total",
			Help: `Deprecated: use sis_requests_total{method="GET"}.`,
		}),
	}

	m.registry.MustRegister(
		m.requests,
		m.requestDuration,
		m.storeDuration,
		m.legacyGets,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	if p, ok := store.(collectorProvider); ok {
		m.registry.MustRegister(p.Collector())
	}

	return m
}

func (m *metrics) observeRequest(method string, code int, d time.Duration) {
	codeStr := strconv.Itoa(code)
	m.requests.WithLabelValues(method, codeStr).Inc()
	m.requestDuration.WithLabelValues(method, codeStr).Observe(d.Seconds())
	if method == "GET" {
		m.legacyGets.Inc()
	}
}

// observeStore records a store call; result is "ok", "not_found", "timeout" or "error".
func (m *metrics) observeStore(op, result string, d time.Duration) {
	m.storeDuration.WithLabelValues(op, result).Observe(d.Seconds())
}

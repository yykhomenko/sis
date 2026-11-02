package sis

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// metrics keeps request counters and renders them in the Prometheus text format.
type metrics struct {
	mu       sync.RWMutex
	requests map[requestKey]*atomic.Uint64
}

type requestKey struct {
	method string
	code   int
}

func newMetrics() *metrics {
	return &metrics{requests: make(map[requestKey]*atomic.Uint64)}
}

func (m *metrics) inc(method string, code int) {
	key := requestKey{method: method, code: code}

	m.mu.RLock()
	counter, ok := m.requests[key]
	m.mu.RUnlock()

	if !ok {
		m.mu.Lock()
		if counter, ok = m.requests[key]; !ok {
			counter = &atomic.Uint64{}
			m.requests[key] = counter
		}
		m.mu.Unlock()
	}

	counter.Add(1)
}

func (m *metrics) render() string {
	m.mu.RLock()
	keys := make([]requestKey, 0, len(m.requests))
	for k := range m.requests {
		keys = append(keys, k)
	}
	m.mu.RUnlock()

	sort.Slice(keys, func(i, j int) bool {
		if keys[i].method != keys[j].method {
			return keys[i].method < keys[j].method
		}
		return keys[i].code < keys[j].code
	})

	var b strings.Builder
	var gets uint64

	b.WriteString("# HELP sis_requests_total Subscriber API requests by method and response status.\n")
	b.WriteString("# TYPE sis_requests_total counter\n")
	m.mu.RLock()
	for _, k := range keys {
		v := m.requests[k].Load()
		if k.method == "GET" {
			gets += v
		}
		fmt.Fprintf(&b, "sis_requests_total{method=%q,code=\"%d\"} %d\n", k.method, k.code, v)
	}
	m.mu.RUnlock()

	// Kept for existing dashboards: number of GET /subscribers requests.
	b.WriteString("# HELP subscribers_total Deprecated: use sis_requests_total{method=\"GET\"}.\n")
	b.WriteString("# TYPE subscribers_total counter\n")
	fmt.Fprintf(&b, "subscribers_total %d\n", gets)

	return b.String()
}

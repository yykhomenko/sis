package sis

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector exposes pgxpool statistics as Prometheus metrics.
func (s *pgStore) Collector() prometheus.Collector {
	return newPoolCollector(s.pool)
}

type poolCollector struct {
	pool *pgxpool.Pool

	acquiredConns        *prometheus.Desc
	idleConns            *prometheus.Desc
	constructingConns    *prometheus.Desc
	totalConns           *prometheus.Desc
	maxConns             *prometheus.Desc
	acquires             *prometheus.Desc
	acquireDuration      *prometheus.Desc
	emptyAcquires        *prometheus.Desc
	emptyAcquireWaitTime *prometheus.Desc
	canceledAcquires     *prometheus.Desc
	newConns             *prometheus.Desc
}

func newPoolCollector(pool *pgxpool.Pool) *poolCollector {
	desc := func(name, help string) *prometheus.Desc {
		return prometheus.NewDesc("sis_db_pool_"+name, help, nil, nil)
	}
	return &poolCollector{
		pool:                 pool,
		acquiredConns:        desc("acquired_conns", "Connections currently in use."),
		idleConns:            desc("idle_conns", "Idle connections in the pool."),
		constructingConns:    desc("constructing_conns", "Connections being established."),
		totalConns:           desc("total_conns", "All connections in the pool."),
		maxConns:             desc("max_conns", "Maximum pool size (SIS_DB_MAX_CONNS)."),
		acquires:             desc("acquires_total", "Successful connection acquires."),
		acquireDuration:      desc("acquire_duration_seconds_total", "Total time spent acquiring connections."),
		emptyAcquires:        desc("empty_acquires_total", "Acquires that had to wait because no connection was idle."),
		emptyAcquireWaitTime: desc("empty_acquire_wait_seconds_total", "Total time spent waiting in empty acquires."),
		canceledAcquires:     desc("canceled_acquires_total", "Acquires cancelled by context (e.g. SIS_DB_TIMEOUT)."),
		newConns:             desc("new_conns_total", "Connections opened since start."),
	}
}

func (c *poolCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *poolCollector) Collect(ch chan<- prometheus.Metric) {
	s := c.pool.Stat()
	gauge := func(d *prometheus.Desc, v float64) {
		ch <- prometheus.MustNewConstMetric(d, prometheus.GaugeValue, v)
	}
	counter := func(d *prometheus.Desc, v float64) {
		ch <- prometheus.MustNewConstMetric(d, prometheus.CounterValue, v)
	}

	gauge(c.acquiredConns, float64(s.AcquiredConns()))
	gauge(c.idleConns, float64(s.IdleConns()))
	gauge(c.constructingConns, float64(s.ConstructingConns()))
	gauge(c.totalConns, float64(s.TotalConns()))
	gauge(c.maxConns, float64(s.MaxConns()))
	counter(c.acquires, float64(s.AcquireCount()))
	counter(c.acquireDuration, s.AcquireDuration().Seconds())
	counter(c.emptyAcquires, float64(s.EmptyAcquireCount()))
	counter(c.emptyAcquireWaitTime, s.EmptyAcquireWaitTime().Seconds())
	counter(c.canceledAcquires, float64(s.CanceledAcquireCount()))
	counter(c.newConns, float64(s.NewConnsCount()))
}

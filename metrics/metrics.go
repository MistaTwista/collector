package metrics

import (
	"collector/config"

	"github.com/prometheus/client_golang/prometheus"
)

type Metricable interface {
	Set(float64)
}

type Adder interface {
	Add(float64)
}

type Setter interface {
	Set(float64)
}

type Counter struct {
	val float64
	prev float64
	collector Adder
}

func NewCounter(a Adder) *Counter {
	return &Counter{
		collector: a,
	}
}

func (c *Counter) Value() float64 {
	return c.val
}

func (c *Counter) Set(n float64) {
	diff := 0.0
	if n < c.prev {
		diff = n
	}
	if n == c.prev {
		return
	}
	if n > c.prev {
		diff = n - c.prev
	}

	c.val += diff
	c.collector.Add(diff)
	c.prev = n
}

type Gauge struct {
	val float64
	collector Setter
}

func NewGauge(s Setter) *Gauge {
	return &Gauge{
		collector: s,
	}
}

func (g *Gauge) Value() float64 {
	return g.val
}

func (g *Gauge) Set(n float64) {
	g.collector.Set(n)
}

func NewMetrics(job config.Job) map[config.MetricName]Metricable {
	metricMap := make(map[config.MetricName]Metricable)

	for _, t := range job.Tasks {
		var m Metricable
		switch t.Type {
		case config.Counter:
			col := prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: job.Namespace,
				Subsystem: job.Subsystem,
				Name:      string(t.Name),
				Help:      t.Description,
			})
			prometheus.MustRegister(col)

			m = NewCounter(col)
		case config.Gauge:
			col := prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: job.Namespace,
				Subsystem: job.Subsystem,
				Name:      string(t.Name),
				Help:      t.Description,
			})
			prometheus.MustRegister(col)

			m = NewGauge(col)
		default:
			panic("Use of bad task type")
		}

		metricMap[t.Name] = m
	}

	return metricMap
}



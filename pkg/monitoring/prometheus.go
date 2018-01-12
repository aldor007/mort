package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

type PrometheusReporter struct {
	countersVec   map[string]*prometheus.CounterVec
	counters      map[string]prometheus.Counter
	gaugesVec     map[string]prometheus.GaugeVec
	gauges        map[string]prometheus.Gauge
	histograms    map[string]prometheus.Histogram
	histogramsVec map[string]prometheus.HistogramVec
}

func NewPrometheusReporter() *PrometheusReporter {
	p := PrometheusReporter{}
	p.countersVec = make(map[string]*prometheus.CounterVec)
	p.counters = make(map[string]prometheus.Counter)
	p.gaugesVec = make(map[string]prometheus.GaugeVec)
	p.gauges = make(map[string]prometheus.Gauge)
	p.histograms = make(map[string]prometheus.Histogram)
	p.histogramsVec = make(map[string]prometheus.HistogramVec)
	return &p
}

// metric - status_codes;sc:200
func (p *PrometheusReporter) Inc(metric string) {
	parts := strings.Split(metric, ";")
	name := parts[0]
	if len(parts) == 1 {
		p.counters[name].Inc()
	} else {
		c, ok := p.countersVec[name]
		if ok {
			c.With(getLabels(parts[1])).Inc()
		}

	}
}

func (p *PrometheusReporter) Counter(metric string, val float64) {
	parts := strings.Split(metric, ";")
	name := parts[0]
	if len(parts) == 1 {
		p.counters[name].Add(val)
	} else {
		c, ok := p.countersVec[name]
		if ok {
			c.With(getLabels(parts[1])).Add(val)
		}

	}
}

func (p *PrometheusReporter) Gauge(metric string, val float64) {
	parts := strings.Split(metric, ";")
	name := parts[0]
	if len(parts) == 1 {
		p.gauges[name].Add(val)
	} else {
		c, ok := p.gaugesVec[name]
		if ok {
			c.With(getLabels(parts[1])).Add(val)
		}
	}
}

func (p *PrometheusReporter) Histogram(metric string, val float64) {
	parts := strings.Split(metric, ";")
	name := parts[0]
	if len(parts) == 1 {
		p.histograms[name].Observe(val)
	} else {
		c, ok := p.histogramsVec[name]
		if ok {
			c.With(getLabels(parts[1])).Observe(val)
		}
	}
}

func (p *PrometheusReporter) RegisterCounter(name string, c prometheus.Counter) {
	prometheus.MustRegister(c)
	p.counters[name] = c

}

func (p *PrometheusReporter) RegisterCounterVec(name string, c *prometheus.CounterVec) {
	prometheus.MustRegister(c)
	p.countersVec[name] = c
}

func (p *PrometheusReporter) RegisterGauge(name string, c prometheus.Gauge) {
	prometheus.MustRegister(c)
	p.gauges[name] = c

}

func (p *PrometheusReporter) RegisterGaugeVec(name string, c prometheus.GaugeVec) {
	prometheus.MustRegister(c)
	p.gaugesVec[name] = c
}

func (p *PrometheusReporter) RegisterHistogram(name string, c prometheus.Histogram) {
	prometheus.MustRegister(c)
	p.histograms[name] = c
}

func (p *PrometheusReporter) RegisterHistogramVec(name string, c prometheus.HistogramVec) {
	prometheus.MustRegister(c)
	p.histogramsVec[name] = c
}

func getLabels(label string) prometheus.Labels {
	parts := strings.Split(label, ":")
	partsLen := len(parts)
	labels := make(prometheus.Labels)
	for i, p := range parts {
		if i+1 == partsLen {
			break
		}

		labels[p] = parts[i+1]
	}

	return labels
}

package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"time"
)

// PrometheusReporter is a reporter that allow you to "send" metrics to
// prometheus
// to allow use all of prometheus metric has format:
// metric_name;label:value,label1:value2
type PrometheusReporter struct {
	countersVec   map[string]*prometheus.CounterVec
	counters      map[string]prometheus.Counter
	gaugesVec     map[string]prometheus.GaugeVec
	gauges        map[string]prometheus.Gauge
	histograms    map[string]prometheus.Histogram
	histogramsVec map[string]*prometheus.HistogramVec
}

// NewPrometheusReporter create instance of reporter that allow you to report stats to prometheus
func NewPrometheusReporter() *PrometheusReporter {
	p := PrometheusReporter{}
	p.countersVec = make(map[string]*prometheus.CounterVec)
	p.counters = make(map[string]prometheus.Counter)
	p.gaugesVec = make(map[string]prometheus.GaugeVec)
	p.gauges = make(map[string]prometheus.Gauge)
	p.histograms = make(map[string]prometheus.Histogram)
	p.histogramsVec = make(map[string]*prometheus.HistogramVec)
	return &p
}

// Inc increment value of metric
// metric schema: metric_name;label:value,label1:value2
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

// Counter increment value of metric by val
// metric schema: metric_name;label:value,label1:value2
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

// Gauge increment value of metric by val
// metric schema: metric_name;label:value,label1:value2
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

// Histogram observer value of metric histogram
// metric schema: metric_name;label:value,label1:value2
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

// Timer allows you to measure time. It returns Timer object on which you have to call Done to end measurement
func (p *PrometheusReporter) Timer(metric string) Timer {
	t := Timer{time.Now(), func(start time.Time) {
		timeDiff := time.Since(start)
		p.Histogram(metric, float64(timeDiff.Nanoseconds()/1000))
	}}

	return t
}

// RegisterCounter register counter in prometheus default registry
// and add to map counter
func (p *PrometheusReporter) RegisterCounter(name string, c prometheus.Counter) {
	prometheus.MustRegister(c)
	p.counters[name] = c

}

// RegisterCounterVec register counterVec in prometheus default registry
// and add to map counter
func (p *PrometheusReporter) RegisterCounterVec(name string, c *prometheus.CounterVec) {
	prometheus.MustRegister(c)
	p.countersVec[name] = c
}

// RegisterGauge register gauge in prometheus default registry
// and add to map counter
func (p *PrometheusReporter) RegisterGauge(name string, c prometheus.Gauge) {
	prometheus.MustRegister(c)
	p.gauges[name] = c

}

// RegisterGaugeVec register gaugeVec in prometheus default registry
// and add to map counter
func (p *PrometheusReporter) RegisterGaugeVec(name string, c prometheus.GaugeVec) {
	prometheus.MustRegister(c)
	p.gaugesVec[name] = c
}

// RegisterHistogram register histogram in prometheus default registry
// and add to map counter
func (p *PrometheusReporter) RegisterHistogram(name string, c prometheus.Histogram) {
	prometheus.MustRegister(c)
	p.histograms[name] = c
}

// RegisterHistogramVec register histogramVec in prometheus default registry
// and add to map counter
func (p *PrometheusReporter) RegisterHistogramVec(name string, c *prometheus.HistogramVec) {
	prometheus.MustRegister(c)
	p.histogramsVec[name] = c
}

func getLabels(label string) prometheus.Labels {
	parts := strings.Split(label, ",")
	labels := make(prometheus.Labels)
	for _, p := range parts {
		v := strings.Split(p, ":")
		labels[v[0]] = v[1]
	}

	return labels
}

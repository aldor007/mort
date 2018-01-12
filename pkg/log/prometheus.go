package log


import (
	"github.com/prometheus/client_golang/prometheus"
"strings"
)

type PrometheusReporter struct {
	registry prometheus.Registry
	countersVec map[string]prometheus.CounterVec
	counters  map[string]prometheus.Counter
	gaugesVec map[string]prometheus.GaugeVec
	gauges map[string]prometheus.Gauge
	histograms map[string] prometheus.Histogram
}

func (p *PrometheusReporter) RegisterCounter(name string, c prometheus.Counter)  error {
	err := p.registry.Register(c)
	if err != nil {
		return err
	}

	p.counters[name] = c

	return nil

}

func (p *PrometheusReporter) RegisterCounterVec(name string, c prometheus.CounterVec)  error {
	err := p.registry.Register(c)
	if err != nil {
		return err
	}
	p.countersVec[name] = c

	return nil
}

// metric - status_codes;sc:200
func (p *PrometheusReporter) Inc(metric string)  {
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

func (p *PrometheusReporter) Gauage(metric string, val float64)  {
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

func getLabels(label string) prometheus.Labels {
	parts := strings.Split(label, ":")
	partsLen := len(parts)
	labels := make(prometheus.Labels)
	for i, p := range parts {
		if i + 1 == partsLen {
			break
		}

		labels[p] = parts[i + 1]
	}

	return labels
}
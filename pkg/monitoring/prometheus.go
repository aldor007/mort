package monitoring


import (
	"github.com/prometheus/client_golang/prometheus"
"strings"
)

type PrometheusReporter struct {
	countersVec map[string]prometheus.CounterVec
	counters  map[string]prometheus.Counter
	gaugesVec map[string]prometheus.GaugeVec
	gauges map[string]prometheus.Gauge
	histograms map[string] prometheus.Histogram
	histogramsVec map[string] prometheus.HistogramVec
}

func NewPrometheusReporter()  *PrometheusReporter {
	p := PrometheusReporter{}
	p.countersVec = make(map[string]prometheus.CounterVec)
	p.counters = make(map[string]prometheus.Counter)
	p.gaugesVec = make(map[string]prometheus.GaugeVec)
	p.gauges = make(map[string]prometheus.Gauge)
	p.histograms = make(map[string]prometheus.Histogram)
	p.histogramsVec = make(map[string]prometheus.HistogramVec)
	return &p
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

func(p *PrometheusReporter) Histogram(metric string, val float64) {
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

func (p *PrometheusReporter) RegisterCounter(name string, c prometheus.Counter)  error {
	err := prometheus.Register(c)
	if err != nil {
		return err
	}

	p.counters[name] = c

	return nil

}

func (p *PrometheusReporter) RegisterCounterVec(name string, c prometheus.CounterVec)  error {
	err := prometheus.Register(c)
	if err != nil {
		return err
	}
	p.countersVec[name] = c

	return nil
}

func (p *PrometheusReporter) RegisterGauge(name string, c prometheus.Gauge)  error {
	err := prometheus.Register(c)
	if err != nil {
		return err
	}

	p.gauges[name] = c

	return nil

}

func (p *PrometheusReporter) RegisterGaugeVec(name string, c prometheus.GaugeVec)  error {
	err := prometheus.Register(c)
	if err != nil {
		return err
	}
	p.gaugesVec[name] = c

	return nil
}

func (p *PrometheusReporter) RegisterHistogram(name string, c prometheus.Histogram)  error {
	err := prometheus.Register(c)
	if err != nil {
		return err
	}

	p.histograms[name] = c

	return nil

}

func (p *PrometheusReporter) RegisterHistogramVec(name string, c prometheus.HistogramVec)  error {
	err := prometheus.Register(c)
	if err != nil {
		return err
	}
	p.histogramsVec[name] = c

	return nil
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

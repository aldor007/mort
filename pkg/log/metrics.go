package log


type Reporter interface {
	Counter(label string, val float64)
	Inc(label string)
	Histogram(label string, val float64)

}

type NopReporter struct {

}

func (n NopReporter) Counter(_ string, _ float64)  {

}

func (n NopReporter) Inc(_ string)  {

}

func (n NopReporter) Histogram(_ string, _ float64)  {

}

type PrometheusReporter struct {
	registry promethues.Registry
	counters map[string]prometheus.Counter
	histograms map[string] prometheus.Histogram
}

func (p *PrometheusReporter) RegisterCounter(c *prometheus.Counter)  {

}

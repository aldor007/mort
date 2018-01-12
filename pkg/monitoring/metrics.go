package monitoring

import (
)


type Reporter interface {
	Counter(label string, val float64)
	Inc(label string)
	Histogram(label string, val float64)
	Gauge(label string, val float64)

}

type NopReporter struct {

}

func (n NopReporter) Counter(_ string, _ float64)  {

}

func (n NopReporter) Inc(_ string)  {

}

func (n NopReporter) Histogram(_ string, _ float64)  {

}

func (n NopReporter) Gauge(_ string, _ float64)  {

}

var reporter Reporter = &NopReporter{}

func Report() Reporter {
	return reporter
}

func RegisterReporter(r Reporter)  {
	reporter = r
}
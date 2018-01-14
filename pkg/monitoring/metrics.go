package monitoring

import "time"

// Reporter is a interface for gather information and send
// external monitoring tool
type Reporter interface {
	Counter(label string, val float64)
	Inc(label string)
	Histogram(label string, val float64)
	Gauge(label string, val float64)
	Timer(label string) Timer
}

type Timer struct {
	start time.Time
	done  func(start time.Time)
}

func (t Timer) Done() {
	t.done(t.start)
}

// NopReporter is reporter that does nothing
type NopReporter struct {
}

// Counter does nothing
func (n NopReporter) Counter(_ string, _ float64) {

}

// Inc does nothing
func (n NopReporter) Inc(_ string) {

}

// Histogram does nothing
func (n NopReporter) Histogram(_ string, _ float64) {

}

// Gauge does nothing
func (n NopReporter) Gauge(_ string, _ float64) {

}

// Timer returns Timer object that measure time between its creation and calling Done Function
func (n NopReporter) Timer(_ string) Timer {
	t := Timer{}
	t.done = func(_ time.Time) {

	}
	return t
}

// reporter instance for use as singleton
var reporter Reporter = &NopReporter{}

// Report is a function that returns reporter instance
// It allows you to report stats to external monitoring
func Report() Reporter {
	return reporter
}

// RegisterReporter change current used reporter with provider
// default reporter is NopReporter that do nothing
func RegisterReporter(r Reporter) {
	reporter = r
}

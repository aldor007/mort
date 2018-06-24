package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"testing"

	"github.com/stretchr/testify/assert"
	"time"
)

func TestPrometheusReporter_Counter(t *testing.T) {
	p := NewPrometheusReporter()

	p.RegisterCounter("test", prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mort_cache_ratio",
		Help: "mort cache ratio",
	}))

	p.Counter("test", 2)

	result := dto.Metric{}
	p.counters["test"].Write(&result)

	assert.Equal(t, *result.Counter.Value, 2.)
}

func TestPrometheusReporter_Gauge(t *testing.T) {
	p := NewPrometheusReporter()

	p.RegisterGauge("test1", prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mort_cache_ratio1",
		Help: "mort cache ratio",
	}))

	p.Gauge("test1", 5)

	result := dto.Metric{}
	p.gauges["test1"].Write(&result)

	assert.Equal(t, *result.Gauge.Value, 5.)
}

func TestPrometheusReporter_Histogram(t *testing.T) {
	p := NewPrometheusReporter()

	p.RegisterHistogram("test2", prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "mort_cache_ratio2",
		Help: "mort cache ratio",
	}))

	p.Histogram("test2", 55)

	result := dto.Metric{}
	p.histograms["test2"].Write(&result)

	assert.Equal(t, *result.Histogram.SampleSum, 55.)
}

func TestPrometheusReporter_Inc(t *testing.T) {
	p := NewPrometheusReporter()

	p.RegisterCounter("test3", prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mort_cache_ratio3",
		Help: "mort cache ratio",
	}))

	p.Inc("test3")

	result := dto.Metric{}
	p.counters["test3"].Write(&result)

	assert.Equal(t, *result.Counter.Value, 1.)
}

func TestPrometheusReporter_Timer(t *testing.T) {
	p := NewPrometheusReporter()

	p.RegisterHistogram("test-2", prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "mort_cache_ratio_2",
		Help: "mort cache ratio",
	}))

	tr := p.Timer("test-2")
	time.Sleep(time.Millisecond * 100)
	tr.Done()

	result := dto.Metric{}
	p.histograms["test-2"].Write(&result)

	assert.InEpsilon(t, *result.Histogram.SampleSum, 100, 5)
}

func TestPrometheusReporter_CounterVec(t *testing.T) {
	p := NewPrometheusReporter()

	p.RegisterCounterVec("testvec", prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mort_cache_ratio_vec",
		Help: "mort cache ratio",
	},
		[]string{"label"},
	))

	p.Counter("testvec;label:elo", 2)
	p.Inc("testvec;label:elo")

	result := dto.Metric{}
	p.countersVec["testvec"].With(prometheus.Labels{"label": "elo"}).Write(&result)

	assert.Equal(t, *result.Counter.Value, 3.)
}

func TestPrometheusReporter_GaugeVec(t *testing.T) {
	p := NewPrometheusReporter()

	p.RegisterGaugeVec("test1vec", prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mort_cache_ratio1_vec",
		Help: "mort cache ratio",
	},
		[]string{"label"},
	))

	p.Gauge("test1vec;label:hej", 5)

	result := dto.Metric{}
	p.gaugesVec["test1vec"].With(prometheus.Labels{"label": "hej"}).Write(&result)

	assert.Equal(t, *result.Gauge.Value, 5.)
}

func TestPrometheusReporter_TimerVec(t *testing.T) {
	p := NewPrometheusReporter()

	p.RegisterHistogramVec("test-2-hist-vec", prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "mort_cache_ratio_2_vec_hist",
		Help: "mort cache ratio",
		Buckets: []float64{10, 50, 100, 200, 300, 400, 500, 1000, 2000, 3000, 4000, 5000, 6000, 10000, 30000, 60000, 70000, 80000},
		},
		[]string{"label"},
	))

	tr := p.Timer("test-2-hist-vec;label:elo")
	time.Sleep(time.Millisecond * 100)
	tr.Done()

	result := dto.Metric{}
	p.histogramsVec["test-2-hist-vec"].With(prometheus.Labels{"label": "elo"}).Write(&result)

	assert.InEpsilon(t, *result.Histogram.SampleSum, 100, 5)
}

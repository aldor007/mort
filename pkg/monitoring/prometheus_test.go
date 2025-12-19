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
		Name:    "mort_cache_ratio_2_vec_hist",
		Help:    "mort cache ratio",
		Buckets: []float64{10, 50, 100, 200, 300, 400, 500, 1000, 2000, 3000, 4000, 5000, 6000, 10000, 30000, 60000, 70000, 80000},
	},
		[]string{"label"},
	))

	tr := p.Timer("test-2-hist-vec;label:elo")
	time.Sleep(time.Millisecond * 100)
	tr.Done()

	result := dto.Metric{}
	p.histogramsVec["test-2-hist-vec"].With(prometheus.Labels{"label": "elo"}).(prometheus.Metric).Write(&result)

	assert.InEpsilon(t, *result.Histogram.SampleSum, 100, 5)
}

// TestPrometheusReporter_UnregisteredCounter verifies that using an unregistered counter doesn't panic
func TestPrometheusReporter_UnregisteredCounter(t *testing.T) {
	t.Parallel()
	p := NewPrometheusReporter()

	// This should not panic, just log a warning
	assert.NotPanics(t, func() {
		p.Inc("unregistered_counter")
	}, "Inc on unregistered counter should not panic")

	assert.NotPanics(t, func() {
		p.Counter("unregistered_counter", 5.0)
	}, "Counter on unregistered counter should not panic")
}

// TestPrometheusReporter_UnregisteredCounterVec verifies that using an unregistered counter vec doesn't panic
func TestPrometheusReporter_UnregisteredCounterVec(t *testing.T) {
	t.Parallel()
	p := NewPrometheusReporter()

	// This should not panic, just log a warning
	assert.NotPanics(t, func() {
		p.Inc("unregistered_counter_vec;label:value")
	}, "Inc on unregistered counter vec should not panic")

	assert.NotPanics(t, func() {
		p.Counter("unregistered_counter_vec;label:value", 5.0)
	}, "Counter on unregistered counter vec should not panic")
}

// TestPrometheusReporter_UnregisteredGauge verifies that using an unregistered gauge doesn't panic
func TestPrometheusReporter_UnregisteredGauge(t *testing.T) {
	t.Parallel()
	p := NewPrometheusReporter()

	// This should not panic, just log a warning
	assert.NotPanics(t, func() {
		p.Gauge("unregistered_gauge", 10.0)
	}, "Gauge on unregistered gauge should not panic")
}

// TestPrometheusReporter_UnregisteredGaugeVec verifies that using an unregistered gauge vec doesn't panic
func TestPrometheusReporter_UnregisteredGaugeVec(t *testing.T) {
	t.Parallel()
	p := NewPrometheusReporter()

	// This should not panic, just log a warning
	assert.NotPanics(t, func() {
		p.Gauge("unregistered_gauge_vec;label:value", 10.0)
	}, "Gauge on unregistered gauge vec should not panic")
}

// TestPrometheusReporter_UnregisteredHistogram verifies that using an unregistered histogram doesn't panic
func TestPrometheusReporter_UnregisteredHistogram(t *testing.T) {
	t.Parallel()
	p := NewPrometheusReporter()

	// This should not panic, just log a warning
	assert.NotPanics(t, func() {
		p.Histogram("unregistered_histogram", 100.0)
	}, "Histogram on unregistered histogram should not panic")
}

// TestPrometheusReporter_UnregisteredHistogramVec verifies that using an unregistered histogram vec doesn't panic
func TestPrometheusReporter_UnregisteredHistogramVec(t *testing.T) {
	t.Parallel()
	p := NewPrometheusReporter()

	// This should not panic, just log a warning
	assert.NotPanics(t, func() {
		p.Histogram("unregistered_histogram_vec;label:value", 100.0)
	}, "Histogram on unregistered histogram vec should not panic")
}

// TestPrometheusReporter_UnregisteredTimer verifies that using an unregistered timer doesn't panic
func TestPrometheusReporter_UnregisteredTimer(t *testing.T) {
	t.Parallel()
	p := NewPrometheusReporter()

	// This should not panic, just log a warning
	assert.NotPanics(t, func() {
		tr := p.Timer("unregistered_timer")
		time.Sleep(time.Millisecond * 10)
		tr.Done()
	}, "Timer on unregistered histogram should not panic")
}

// TestPrometheusReporter_MixedRegisteredAndUnregistered verifies resilience with mixed scenarios
func TestPrometheusReporter_MixedRegisteredAndUnregistered(t *testing.T) {
	t.Parallel()
	p := NewPrometheusReporter()

	// Register one counter
	p.RegisterCounter("registered_counter", prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_registered_counter",
		Help: "test registered counter",
	}))

	// Use registered counter - should work
	assert.NotPanics(t, func() {
		p.Inc("registered_counter")
		p.Counter("registered_counter", 5.0)
	}, "Registered counter should work")

	// Use unregistered counter - should not panic
	assert.NotPanics(t, func() {
		p.Inc("unregistered_counter")
		p.Counter("unregistered_counter", 5.0)
	}, "Unregistered counter should not panic")

	// Verify registered counter has correct value
	result := dto.Metric{}
	p.counters["registered_counter"].Write(&result)
	assert.Equal(t, 6.0, *result.Counter.Value, "Registered counter should have value 6")
}

// TestPrometheusReporter_RealWorldScenario simulates the vips_cleanup_count scenario
func TestPrometheusReporter_RealWorldScenario(t *testing.T) {
	t.Parallel()
	p := NewPrometheusReporter()

	// Simulate forgetting to register the metric (like the vips_cleanup_count bug)
	// This should not panic
	assert.NotPanics(t, func() {
		for i := 0; i < 10; i++ {
			p.Inc("vips_cleanup_count")
		}
	}, "Should handle unregistered metric gracefully")

	// Now register it
	p.RegisterCounter("vips_cleanup_count", prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mort_vips_cleanup_count",
		Help: "mort count of vips cache cleanups",
	}))

	// Now it should work and accumulate
	assert.NotPanics(t, func() {
		p.Inc("vips_cleanup_count")
		p.Inc("vips_cleanup_count")
	}, "Should work after registration")

	// Verify the counter has the correct value (only the 2 after registration)
	result := dto.Metric{}
	p.counters["vips_cleanup_count"].Write(&result)
	assert.Equal(t, 2.0, *result.Counter.Value, "Should only count increments after registration")
}

// TestPrometheusReporter_GlacierMetrics verifies GLACIER metrics registration and usage
func TestPrometheusReporter_GlacierMetrics(t *testing.T) {
	t.Parallel()
	p := NewPrometheusReporter()

	// Register GLACIER error detected counter
	p.RegisterCounter("glacier_error_detected", prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mort_glacier_error_detected",
		Help: "mort count of GLACIER/DEEP_ARCHIVE errors detected",
	}))

	// Register GLACIER restore initiated counter
	p.RegisterCounter("glacier_restore_initiated", prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mort_glacier_restore_initiated",
		Help: "mort count of GLACIER restore requests initiated",
	}))

	// Increment glacier_error_detected
	p.Inc("glacier_error_detected")
	p.Inc("glacier_error_detected")

	// Increment glacier_restore_initiated
	p.Inc("glacier_restore_initiated")

	// Verify glacier_error_detected value
	result1 := dto.Metric{}
	p.counters["glacier_error_detected"].Write(&result1)
	assert.Equal(t, 2.0, *result1.Counter.Value, "glacier_error_detected should be 2")

	// Verify glacier_restore_initiated value
	result2 := dto.Metric{}
	p.counters["glacier_restore_initiated"].Write(&result2)
	assert.Equal(t, 1.0, *result2.Counter.Value, "glacier_restore_initiated should be 1")
}

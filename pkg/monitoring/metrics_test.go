package monitoring

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNopReporter(t *testing.T) {

	nop := NopReporter{}

	assert.NotPanics(t, func() {
		nop.Counter("a", 3.)
		nop.Gauge("b", 4.)
		nop.Histogram("c", 1.0)
		nop.Inc("d")
		t := nop.Timer("e")
		t.Done()
	})
}

func TestRegisterReporter(t *testing.T) {
	nop := NopReporter{}

	RegisterReporter(nop)
	nop2 := Report()

	assert.Equal(t, nop, nop2)
}

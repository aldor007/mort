package monitoring

import (
	"go.uber.org/zap"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	l := zap.NewNop()
	RegisterLogger(l)
	assert.Equal(t, l, Log())

	s := l.Sugar()
	assert.Equal(t, s, Logs())
}
